/*
	
	JOTPOT Server
	Version 27A-0
	
	Copyright (c) 2016-2017 Jacob O'Toole
	
	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:
	
	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.
	
	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE.
	
*/

"use strict";

//Modules
const fs = require("fs") ;
const path = require("path") ;
const os = require("os") ;
const http = require("http") ;

//Default configuration
const defaultConfig = {
	
	"dataPort": 500,
	"controlers":["::1","127.0.0.1","::ffff:127.0.0.1"],
	
	"httpServers": [
		{
			"port": 80,
			"websocket": false
		}
	],
	"httpsServers": [],
	
	"redirectToHttps": [],
	"mustRedirectToHttps": [],
	"dontRedirect": [],
	
	"hostRedirects":{},
	"hostnameRedirects":{},
	"hostAlias":{},
	"hostnameAlias":{},
	"pageAlias":{},
	
	"addVarsByDefault": false,
	"doVarsForIfNotByDefault": [],
	
	"cache": [],
	
	"errorTemplate": "errorTemp.jpt",
	
	"defaultHost": "default:0",
	"useDefaultHostIfHostDoesNotExist": true,
	
	"behindLoadBalencer": false,
	"fallbackToNoPort": true,
	
	"defaultHeaders": {},
	
	"CORS":[],
	
	"enableLearning": true,
	
	"threads": 0
	
} ;

const loadConfigFile = p => {
	if (fs.existsSync(p)) {
		return fs.readFileSync(p).toString() ;
	}
	return fs.readFileSync(path.join(__dirname, p)).toString() ;
} ;
const doesConfigFileExist = p => (fs.existsSync(p) || fs.existsSync(path.join(__dirname, p))) ;

//Load the comfig and fill in any blanks. If it doesn't exist, set the config to the default config.
function loadConfig() {
	let config ;
	//If it exists, load it, parse it and fill in any blanks or throw if the types aren't correct
	if (doesConfigFileExist("config.json")) {
		config = loadConfigFile("config.json") ;
		try {
			config = JSON.parse(config) ;
		} catch (err) {
			console.warn("Error parsing config.json!") ;
			console.info("Error parsing config.json!") ;
			console.warn(err) ;
			console.info(err) ;
			console.warn("Exiting") ;
			console.info("Exiting") ;
			process.exit(1) ;
		}
		for (let doing in defaultConfig) {
			if (typeof config[doing] === "undefined") {
				config[doing] = defaultConfig[doing] ;
			} else if (typeof config[doing] !== typeof defaultConfig[doing]) {
				throw new Error(`The ${doing} property in config.json must be of type ${typeof defaultConfig[doing]}.`) ;
			}
		}
	} else {
		console.warn("Config file does not exist, using default config.") ;
		config = new Object() ;
		Object.assign(config, defaultConfig) ;
	}
	return config ;
}

//Returns a promise that resolves with the contents of the given requests body.
function getData(req) {
	return new Promise((resolve, reject)=>{
		//If the request body has a specified length
		if (typeof req.headers["content-length"] !== "undefined") {
			//Parse and reject if not number
			let dLength = parseInt(req.headers["content-length"], 10) ;
			if (isNaN(dLength)) {
				reject(new Error("Content-Length header must be a number")) ;
				return ;
			}
			//Create the buffer
			let data = Buffer.alloc(dLength) ;
			let currentPos = 0 ;
			let errorSent = false ;
			req.on("data", d=>{
				//If we have more data than we can take, reject.
				if (currentPos + d.length > data.length) {
					//But only reject once
					if (errorSent) {
						return ;
					}
					errorSent = true ;
					reject(new Error("Request body was longer than the Content-Length header.")) ;
					return ;
				}
				//Write the recieved data to the output and increment the curret position by the amount of data copied.
				currentPos += d.copy(data, currentPos) ;
			}) ;
			//Resolve when the request has ended.
			req.on("end", ()=>{
				resolve(data) ;
			}) ;
			return ;
		}
		//No idea how much data, so we have to concatinate all the Buffers :(
		let data = Buffer.alloc(0) ;
		req.on("data", d=>{
			data = Buffer.concat([data, d], data.length + d.length) ;
		}) ;
		//Resolve when we are done
		req.on("end", ()=>{
			resolve(data) ;
		}) ;
	}) ;
}

function coughtError(err, where="", resp, rID="", userMessage="") {
	let isUnknown = false ;
	console.warn("---------------") ;
	if (err && err.stack) {
		console.warn("!!! Error" + where + ":") ;
		console.warn("\t" + err.stack.replace(/\n/g,"\n\t")) ;
	} else if (err) {
		console.warn("!!! Error" + where + ":") ;
		console.warn("\t" + err.replace(/\n/g,"\n\t")) ;
	} else {
		console.warn("!!! Error" + where + ", details unknown. Stack unavailable.") ;
		isUnknown = true ;
	}
	console.warn("---------------") ;
	if (resp) {
		try {
			if (userMessage) {
				module.exports.sendError(500, userMessage, resp, rID) ;
			} else {
				module.exports.sendError(500, `A${isUnknown?"n unknown":" known "} error occured.${isUnknown?"":" I just don't want to tell you what went wrong. Nothing personal, honestly! It's not like I don't trust you."}.`, resp, rID) ;
			}
		} catch (err) {
			coughtError(err, " sending error response") ;
		}
	}
}

function getFilePath(req) {
	return path.normalize((req.url.usePortInDirectory?req.url.host:req.url.hostname).replace(":", ";") + req.url.pathname) ;
}

function escapeHTML(text) {
	return text
		.replace(/&/g, "&amp;")
		.replace(/</g, "&lt;")
		.replace(/>/g, "&gt;")
		.replace(/"/g, "&#34;")
		.replace(/'/g, "&#39;") ;
}

function getTime() {
	let time = process.hrtime() ;
	return time[0] * 10e3 + time[1] * 10e-6 ;
}

class time {
	constructor(total, fileOp) {
		this.total = total ;
		this.fileOp = fileOp ;
		this.process = total - fileOp ;
	}
}

class timer {
	constructor() {
		this.fileOp = false ;
		this.fileOpStart = 0 ;
		this.fileOpTime = 0 ;
		this.start = getTime() ;
	}
	startFileOp() {
		if (!this.fileOp) {
			this.fileOp = true ;
			this.fileOpStart = getTime() ;
		}
	}
	stopFileOp() {
		if (this.fileOp) {
			this.fileOpTime += getTime() - this.fileOpStart ;
			this.fileOp = false ;
		}
	}
	currentTime() {
		const ct = getTime() ;
		if (this.fileOp) {
			this.fileOpTime += ct - this.fileOpStart ;
			this.fileOpStart = ct ;
		}
		return new time(ct - this.start, this.fileOpTime) ;
	}
}

function getLogDate(d=new Date()) {
	return `${d.getUTCFullYear()}-${d.getUTCMonth()}-${d.getUTCDate()}` ;
}

function getLogTime(d=new Date()) {
	return `${d.getUTCHours()}-${d.getUTCMinutes()}-${d.getUTCSeconds()}.${d.getUTCMilliseconds()/1000}` ;
}

function startLog() {
	const d = new Date() ;
	return `#Version: 1.0${os.EOL}#Software: JOTPOT Server${os.EOL}#Start-Date: ${getLogDate(d)} ${getLogTime(d)}${os.EOL}#Fields: ${fields}` ;
}

function updateLog() {
	const d = new Date() ;
	return `#Date: ${getLogDate(d)} ${getLogTime(d)}` ;
}

function log(req, status, location, cached, learned) {
	console.serverlog(writeLog(req, status, location, cached, learned)) ;
}

const fields = [
	"r-id",
	"c-ip",
	"c-ip-forwarded",
	"c-id",
	"c-username",
	"recieved-date",
	"recieved-time",
	"date",
	"time",
	"time-taken",
	"time-taken-process",
	"time-taken-disk",
	"cs-method",
	"cs-uri",
	"cs-uri-stem",
	"cs-uri-query",
	"cs-url",
	"sc-url-served",
	"cs-requested",
	"sc-served",
	"cs-http-version",
	"cs(User-Agent)",
	"cs(Referer)",
	"cached",
	"learned",
	"sc-status",
	"sc-comment",
	"sc-location"
].join(" ") ;

function writeLog(req, status, location, cached, learned) {
	const d = new Date() ;
	const t = req.timer.currentTime() ;
	return [
		req.jpid, //r-id
		req.remoteAddress, //c-ip
		req.ip, //c-ip-forward
		req.user||'-', //c-id
		'-', //c-username
		getLogDate(req.time), //recieved-date
		getLogTime(req.time), //recieved-time
		getLogDate(d), //date
		getLogTime(d), //time
		t.total.toString(10), //time-taken
		t.process.toString(10), //time-taken-process
		t.fileOp.toString(10), //time-taken-disk
		req.method, //cs-method
		req.orig_url.path, //cs-uri
		req.orig_url.pathname, //cs-uri-stem
		req.orig_url.query, //cs-uri-query
		req.orig_url.href, //cs-url
		req.url.href, //sc-url-served
		req.orig_url.fullvalue, //cs-requested
		req.url.fullvalue, //cs-served
		req.httpVersion, //cs-http-version
		req.headers["user-agent"]||'-', //cs(User-Agent)
		req.headers["referer"]||'-', //cs(Referer)
		cached, //cached
		learned, //leaned
		status.toString(10), //sc-status
		http.STATUS_CODES[status], //sc-comment
		location //sc-location
	].join(' ') ;
}

module.exports = {
	coughtError,
	doesConfigFileExist,
	escapeHTML,
	getData,
	getFilePath,
	loadConfig,
	loadConfigFile,
	log,
	startLog,
	updateLog,
	sendError: ()=>{},
	timer
} ;
