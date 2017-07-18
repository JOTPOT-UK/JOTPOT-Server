/*
	
	JOTPOT Server
	Version 25F
	
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

//Console is now YAY!!!
console.log = console.warn = (...args) => {
	
	process.send(["log",args.join(" ")]) ;
	
}

//Modules
let http = require("http") ;
let https = require("https") ;
let fs = require("fs") ;
let path = require("path") ;
let proc = require("./accounts.js") ;
let externals = require("./externals.js") ;
let URL = require("./url-object.js") ;
let {Transform,Readable} = require("stream") ;
let cluster ;

//Load the config
let config ;

const defaultConfig = {
	
	"otherProcesses": [],
	
	"dataPort": 500,
	
	"httpServers": [
		
		{
			
			"port": 80
			
		}
		
	],
	"httpsServers": [],
	
	"redirectToHttps": [],
	"canBeHttp": [],
	
	"hostRedirects":{},
	"hostAlias":{},
	"pageAlias":{},
	
	"addVarsByDefault": false,
	"doVarsForIfNotByDefault": [],
	
	"cache": [],
	
	"errorTemplate": "errorTemp.jpt",
	
	"defaultHost": "default:0",
	"useDefaultHostIfHostDoesNotExist": true,
	
	"behindLoadBalencer": false
	
} ;

function loadConfig() {

	if (fs.existsSync("config.json")) {
		
		config = fs.readFileSync("config.json").toString() ;
		
		try {
			
			config = JSON.parse(config) ;
			
		}
		
		catch(err) {
			
			console.warn("Error parsing config.json!") ;
			console.info("Error parsing config.json!") ;
			console.warn(err) ;
			console.info(err) ;
			console.warn("Exiting") ;
			console.info("Exiting") ;
			process.exit() ;
			
		}
		
	}

	else {
		
		console.warn("Config file does not exist, using default config.") ;
		config = new Object() ;
		Object.assign(config, defaultConfig) ;
		return ;
		
	}
	
	for (let doing in defaultConfig) {
		
		if (typeof config[doing] === "undefined") {
			
			config[doing] = defaultConfig[doing] ;
			
		}
		
		else if (typeof config[doing] !== typeof defaultConfig[doing]) {
			
			throw new Error(`The ${doing} property in config.json must be of type ${typeof defaultConfig[doing]}.`) ;
			
		}
		
	}
	
}
loadConfig() ;

let availHosts = [] ;
if (config.useDefaultHostIfHostDoesNotExist) {
	
	availHosts = fs.readdirSync("./sites") ;
	
}

//Vars to add to files
let vars = new Object ;
vars.Global = new Object() ;

//Set up cache
let pages = new Object() ;

//Setup accounts
let allAccountSystems = new Array() ;

//Vars config
let doVarsFor = ["error_page"].concat(config.doVarsForIfNotByDefault) ;
for (let doing in doVarsFor) {
	
	doVarsFor[doing] = path.normalize(doVarsFor[doing]) ;
	doVarsFor[doing] = path.join(process.cwd(),"sites",doVarsFor[doing]) ;
	
}
let dontDoVarsFor = [] ;

//Set up current ID
let currentID = 0 ;

//Get buffer of string for inserting vars.
let startOfVar = Buffer.from("$::") ;
let endOfVar = Buffer.from("::$") ;

//Error file
let errorFile ;
if (!fs.existsSync(config.errorTemplate)) {
	
	console.warn("Error template file does not exist, using the default.") ;
	errorFile = `<html>
<head>
<title>
Error $:::error_code:::$ - $:::error_type:::$
</title>
</head>
<body>
	<b>Erm, this isn't meant to happen...</b>
	<h2>$:::error_code:::$: $:::error_type:::$</h2>
	<hr>
	$:::error_message:::$
	<br><br>
	For more infomation please see 
	<a href="https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/$:::error_code:::$">this MDN page</a> 
	or 
	<a href="https://httpstatuses.com/$:::error_code:::$">httpstatuses.com/$:::error_code:::$</a>
	<hr>
</body>
</html>` ;
	
}

else {
	
	errorFile = fs.readFileSync(config.errorTemplate).toString() ;
	 
}

let errorCodes = new Object() ;
errorCodes[403] = "Sorry, however you are not permitted to access this file." ;
errorCodes[404] = "The page you are looking for may have been removed or moved to a new location!" ;
errorCodes[500] = "An unknown error occured." ;

//Pipe for adding vars
//When creating, give 1st argument as the path of the path of the file that will be piped.
class addVars extends Transform {
	
	constructor (path,extraVars={},ignoreList=false,arg) {
		
		if (typeof path === "undefined") {
			
			throw "DUDE!!! WHAT IS THE PATH YOU ARE PIPING TO ME??? (Top tip, first argument needs to be the path.)" ;
			return false ;
			
		}
		super(arg) ;
		this.lastData = "" ;
		this.path = path ;
		this.extraVars = extraVars ;
		this.ignoreList = ignoreList ;
		
	}
	
	_transform (data,encoding,callback) {
		
		try {
			
			let doPush = true ;
			
			do {
				
				if (data.indexOf(startOfVar) !== -1) {
					
					if (data.indexOf(endOfVar) !== -1) {
						
						let dataString = this.lastData + data.toString() ;
						
						let varsKeys = Object.keys(this.extraVars) ;
						for (let doing in varsKeys) {
							
							let toReplace = `\\$\\:\\:\\:${varsKeys[doing]}\\:\\:\\:\\$` ;
							let replaceWith = String(this.extraVars[varsKeys[doing]]) ;
							dataString = dataString.replace(new RegExp(toReplace,"g"),replaceWith) ;
							
						}
						
						varsKeys = Object.keys(vars) ;
						for (let doingScope in varsKeys) {
							
							let innerVars = Object.keys(vars[varsKeys[doingScope]]) ;
							for (let doing in innerVars) {
								
								let toReplace = `\\$\\:\\:\\:${innerVars[doing]}\\:\\:\\:\\$` ;
								let replaceWith = String(vars[varsKeys[doingScope]][innerVars[doing]]) ;
								dataString = dataString.replace(new RegExp(toReplace,"g"),replaceWith) ;
								
							}
							
						}
						
						data = dataString ;
						
						break ;
						
					}
					
					else {
						
						this.lastData = dataString ;
						doPush = false ;
						break ;
						
					}
					
				}
				
				else {
					
					break ;
					
				}
				
			} while(1)
			
			if (doPush) {
				this.push(data) ;
			}
			
			callback() ;
			
		}
		
		catch(err) {
			
			coughtError(err,resp) ;
			
		}
		
	}
	
}

//Sorting out mime types.
let mimes = JSON.parse(fs.readFileSync("./mimes.dat").toString()) ;
function getMimeType(file) {
	
	try {
		
		let ext = path.extname(file) ;
		
		//Give type HTML if there is no extention.
		if (ext === "") {
			
			return "text/html" ;
			
		}
		
		ext = ext.substring(1,ext.length) ;
		
		if (typeof mimes[ext] !== "undefined") {
			
			return mimes[ext] ;
			
		}
		
		else {
			
			return "text/plain" ;
			
		}
		
	}
	
	catch(err) {
		
		coughtError(err,resp) ;
		
	}
	
}

//Proxy to other local server
function forwardToOtherServer(req,resp,port) {
	
	try {
		
		let ended = false ;
		let headersToSend = req.headers ;
		headersToSend["x-forwarded-for"] = req.headers['x-forwarded-for'] || req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress ;
		let forward = http.request({
			
			protocol:"http:",
			host:"localhost",
			port:port,
			method:req.method,
			path:req.url
			
		}) ;
		forward.on("response",iResp=>{
			
			resp.writeHead(iResp.statusCode,iResp.headers) ;
			iResp.pipe(resp) ;
			iResp.on("end",_=>{if(ended){resp.end();ended=true;}}) ;
			
		}) ;
		req.on("end",_=>{if(!ended){forward.socket.end();ended=true;}}) ;
		req.pipe(forward) ;
		
	}
	
	catch(err) {
		
		coughtError(err,resp) ;
		
	}
	
}

function wrapURL(req) {
	
	let url = new URL(req, config.defaultHost || config.defaultDomain) ;
	Object.defineProperty(req, "url", {
		enumerable: true,
		configurable: false,
		get: _=>url,
		set: v=>url.value=v
	}) ;
	
}

//Pipes the file through the transform pipe into the main pipe.
//Calls the callback with the first argument as a boolean - true if succeded, false if not.
function getFile(file,callWithStats,pipeTo,callback) {
	
	fs.stat(file,(err,stats) => {
		
		if (err) {
			
			callback(false,err) ;
			return false ;
			
		}
		
		if (stats.isFile()) {
			
			if (callWithStats(stats)) {
				
				//Pipe file to the pipe.
				fs.createReadStream(file,{
					
					flags: 'r',
					autoClose: true
					
				}).pipe(pipeTo) ;
				
				callback(true,"Erm") ;
				
			}
			
			else {
				
				callback(false, "Rejected because of stats...") ;
				
			}
			
		}
		
		else {
			
			callback(false,"TEST") ;
			
		}
		
	}) ;
	
}

//Sends the file specified to the pipe as the second argument - goes through the getFile & thus vars pipe.
function sendFile(file,resp,customVars,rID="") {
	
	try {
		
		//Look in the sites dir.
		let start = path.join(process.cwd(), "sites") ;
		file = path.join(start,URL.toDir(file)) ;
		
		//If we aren't in the sites dir, then throw.
		if (file.indexOf(start) !== 0) {
			
			throw new Error("Server has left the serving directory!") ;
			
		}
		
		//Make a pipe to send it to.
		let mainPipe = "none" ;
		let doingTransform = resp.pipeThrough.length - 1 ;
		let lengthknown = false ;
		
		//If we need to add vars.
		if (config.addVarsByDefault || doVarsFor.indexOf(file) !== -1) {
			
			//Vars go first.
			mainPipe = new addVars(file,customVars) ;
			
			//Add the resp.pipeThrough
			while (doingTransform > -1) {
				
				mainPipe.pipe(resp.pipeThrough[doingTransform]) ;
				doingTransform-- ;
				
			}
			
			//Then to the client.
			mainPipe.pipe(resp) ;
			
		}
		
		//No vars, but still pipe.
		else if (doingTransform > -1) {
			
			//Start at last thing
			mainPipe = resp.pipeThrough[doingTransform] ;
			doingTransform-- ;
			
			//Now do the rest
			while (doingTransform > -1) {
				
				mainPipe.pipe(resp.pipeThrough[doingTransform]) ;
				doingTransform-- ;
				
			}
			
			//And guess what... The client!
			mainPipe.pipe(resp) ;
			
		}
		
		else {
			
			//No pipes at all, so only to client.
			mainPipe = resp ;
			
			//We can now get the length from the stats
			lengthknown = true ;
			
		}
		
		return new Promise((resolve,reject) => {
				
			getFile(file,stats => {
				
				let mime = resp.forceDownload?"application/octet-stream":getMimeType(file) ;
				console.log(`${rID}\t200 OK.   ${file} (${mime}) loaded from disk.`) ;
				if (lengthknown) {
					resp.setHeader("Content-Length", stats.size) ;
				}
				resp.writeHead(200,{
					
					"Content-Type": mime,
					
					//Added because google does it :)
					"status": 200
					
				}) ;
				
				return true ;
				
			},mainPipe,(done,err) => {
				
				resolve([done,err]) ;
				
			});
			
		}) ;
		
	}
	
	catch(err) {
		
		//coughtError(err,resp,rID) ;
		//Return an instantly rejecting promise
		return new Promise((resolve, reject) => reject(err)) ;
		
	}
	
}

function sendCache(file,cache,resp,customVars,status=200,rID="") {
	
	try {
		
		//Look in the sites dir.
		file = path.join(process.cwd(),"sites",file) ;
		
		
		//Make a pipe to send it to.
		let mainPipe = "none" ;
		resp.pipeThrough = [] ;
		let doingTransform = resp.pipeThrough.length - 1 ;
		let lengthknown = false ;
		
		//If we need to add vars.
		if (config.addVarsByDefault || doVarsFor.indexOf(file) !== -1) {
			
			//Vars go first.
			mainPipe = new addVars(file,customVars) ;
			
			//Add the resp.pipeThrough
			while (doingTransform > -1) {
				
				mainPipe.pipe(resp.pipeThrough[doingTransform]) ;
				doingTransform-- ;
				
			}
			
			//Then to the client.
			mainPipe.pipe(resp) ;
			
		}
		
		//No vars, but still pipe.
		else if (doingTransform > -1) {
			
			//Start at last thing
			mainPipe = resp.pipeThrough[doingTransform] ;
			doingTransform-- ;
			
			//Now do the rest
			while (doingTransform > -1) {
				
				mainPipe.pipe(resp.pipeThrough[doingTransform]) ;
				doingTransform-- ;
				
			}
			
			//And guess what... The client!
			mainPipe.pipe(resp) ;
			
		}
		
		else {
			
			//No pipes at all, so only to client.
			lengthknown = true ;
			mainPipe = resp ;
			
		}
		
		//Get the mime type.
		let mime = getMimeType(file) ;
		console.log(`${rID}\t${status} ${http.STATUS_CODES[status]}.   ${file} (${mime}) loaded from cache.`) ;
		if (lengthknown) {
			resp.setHeader("Content-Length", cache.length) ;
		}
		resp.writeHead(status,{
			
			"Content-Type": mime,
			
			//Added because google does it :)
			"status": status
			
		}) ;
		
		//Write the cached data & end.
		mainPipe.write(cache) ;
		mainPipe.end() ;
		
	}
	
	catch(err) {
		
		coughtError(err,resp) ;
		
	}
	
}

function sendError(code,message,resp,rID="") {
	
	sendCache("error_page",errorFile,resp,{error_code:code,error_type:http.STATUS_CODES[code],error_message:message},code,rID) ;
	return ;
	
}

function coughtError(err,resp,rID="") {
	
	let isUnknown = false ;
	console.warn("---------------") ;
	if (err && err.stack) {
		console.warn("!!! Error in main request handler:") ;
		console.warn("\t" + err.stack.replace(/\n/g,"\n\t")) ;
	} else if (err) {
		console.warn("!!! Error in main request handler:") ;
		console.warn("\t" + err.replace(/\n/g,"\n\t")) ;
	} else {
		console.warn("!!! Error in main request handler, details unknown. Stack unavailable.") ;
		isUnknown = true ;
	}
	console.warn("---------------") ;
	sendError(500,`A${isUnknown?"n unknown":" known "} error occured.${isUnknown?"":" I just don't want to tell you what went wrong. Nothing personal, honestly! It's not like I don't strust you."}.`,resp,rID) ;
	
}

//Function to handle http requests.
function handleRequest(req,resp) {
	
	try {
		
		//Get time stuff.
		let timeRecieved = process.hrtime() ;
		let requestTime = new Date() ;
		
		//Set server header
		resp.setHeader("Server","JOTPOT Server") ;
		
		//Get IP, follow jp-source headers if we are behind a load balancer
		let user_ip, user_ip_remote ;
		if (config.behindLoadBalancer) {
			
			user_ip = (req.headers['x-forwarded-for'] || req.headers["jp-source-ip"] || req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress).replace(/::ffff:/g,"") ;
			user_ip_remote = (req.headers["jp-source-ip"] || req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress).replace(/::ffff:/g,"") ;
			req.overHttps = req.secure = req.headers["jp-source-secure"] === "https" ;
			
		} else {
			
			user_ip = (req.headers['x-forwarded-for'] || req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress).replace(/::ffff:/g,"") ;
			user_ip_remote = (req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress).replace(/::ffff:/g,"") ;
			
		}
		
		//Create URL object
		//req.url = new URL(req, config.defaultHost || config.defaultDomain) ;
		wrapURL(req) ;
		
		//Add stuff to resp object.
		resp.vars = {"user_ip":user_ip,"user_ip_remote":user_ip_remote,"utctime":requestTime.toUTCString(),"time":requestTime.getTime(),"host":req.host,"purl":JSON.stringify(req.purl)} ;
		resp.pipeThrough = new Array() ;
		req.ip = user_ip ;
		req.remoteAddress = user_ip_remote ;
		resp.forceDownload = false ;
		
		//Do request handle.
		let cont = true ;
		let gotOtherPromise = false ;
		externals.doEvt("request",req,resp).then(d=>{
			
			if (d) {
				
				cont = false ;
				
			}
			
			if (gotOtherPromise && cont) {
				
				handleRequestPart2(req,resp,timeRecieved,requestTime,user_ip,user_ip_remote) ;
				
			}
			
			gotOtherPromise = true ;
			
		})
		
		externals.doEvt(`${req.host}/request`,req,resp).then(d=>{
			
			if (d) {
				
				cont = false ;
				
			}
			
			if (gotOtherPromise && cont) {
				
				handleRequestPart2(req,resp,timeRecieved,requestTime,user_ip,user_ip_remote) ;
				
			}
			
			gotOtherPromise = true ;
			
		}) ;
		
	}
	
	catch(err) {
		
		coughtError(err,resp) ;
		console.log("Error trace: Error handling incoming data.") ;
		
	}
	
}

function handleRequestPart2(req,resp,timeRecieved,requestTime,user_ip,user_ip_remote) {
	
	//Secure URL. Remove '..' to prevent it from going to a parent directory.
	req.url.pathname = req.url.pathname.replace(/\.\./g,"") ;
	
	//Should we redirect to https.
	if (req.overHttps === false && config.redirectToHttps.indexOf(req.host) !== -1 && config.canBeHttp.indexOf(req.url.value) === -1) {
		
		console.log(`${req.jpid}\tfrom ${user_ip_remote}(${user_ip}) for ${req.url.value} being handled by thread ${cluster.worker.id}.`) ;
		console.log(`${req.jpid}\t302 Found.   Redirecting to ${req.url.href}.`) ;
		
		req.url.protocol = "https:" ;
		
		resp.writeHead(301,{"Content-Type":"text/plain","location":req.url.href,"status":301}) ;
		resp.write("Redirecting you to our secure site...") ;
		resp.end() ;
		
		let timeTaken = process.hrtime(timeRecieved) ;
		console.log(`${req.jpid}\tRequest took ${timeTaken[0] * 1000 + timeTaken[1] * 10e-6}ms to handle.`) ;
		
		return ;
		
	}
	
	//Is the host an alias
	while (typeof config.hostAlias[req.host] !== "undefined") {
		
		req.host = config.hostAlias[req.host] ;
		
	}
	
	//If we are set to goto a default host, check if the host doesn't exist, if so, we are now default :)
	if (config.useDefaultHostIfHostDoesNotExist) {
		
		if (availHosts.indexOf(URL.toDir(req.url.host)) === -1) {
			
			req.url.host = config.defaultHost || "default:0" ;
			
		}
		
	}
	
	//Should we redirect to another host.
	if (typeof config.hostRedirects[req.url.host] !== "undefined") {
		
		console.log(`${req.jpid}\tfrom ${user_ip_remote}(${user_ip}) for ${req.url} being handled by thread ${cluster.worker.id}.`) ;
		
		//Set new host
		req.url.host = config.hostRedirects[req.url.host] ;
		
		//Set correct protocol
		let isRedirectHttps = config.redirectToHttps.indexOf(config.hostRedirects[req.host]) !== -1 && config.canBeHttp.indexOf(req.url) === -1 ;
		req.url.protocol = isRedirectHttps?"https:":"http:" ;
		
		//And send response
		console.log(`${req.jpid}\t302 Found.   Redirecting to ${req.url.href}.`) ;
		resp.writeHead(301,{"Content-Type":"text/plain","location":req.url.href, "status":301}) ;
		resp.write("Redirecting you to " + req.url.href + "...") ;
		resp.end() ;
		
		let timeTaken = process.hrtime(timeRecieved) ;
		console.log(`${rID}\tRequest took ${timeTaken[0] * 1000 + timeTaken[1] * 10e-6}ms to handle.`) ;
		
		return ;
		
	}
	
	//Page alias?
	if (typeof config.pageAlias[req.url.value] !== "undefined") {
		
		req.url.fullvalue = config.pageAlias[req.url.value] ;
		
	}
	
	{
		let rID = `#${(currentID++).toString(16).toUpperCase()}` ;
		Object.defineProperty(req, "jpid", {
			configurable: false,
			enumerable: false,
			value: rID,
			writable: false
		}) ;
	}
	
	console.log(`${req.jpid}\tfrom ${user_ip_remote}(${user_ip}) for ${req.url.value} being handled by thread ${cluster.worker.id}.`) ;
	
	//Handle for full request.
	
	let cont = true ;
	let gotOtherPromise = false ;
	externals.doEvt("fullrequest",req,resp).then(d=>{
		
		if (d) {
			
			cont = false ;
			
		}
		
		if (gotOtherPromise && cont) {
			
			handleRequestPart3(req,resp,timeRecieved,requestTime,user_ip,user_ip_remote) ;
			
		}
		
		gotOtherPromise = true ;
		
	})
	
	externals.doEvt(`${req.host}/fullrequest`,req,resp).then(d=>{
		
		if (d) {
			
			cont = false ;
			
		}
		
		if (gotOtherPromise && cont) {
			
			handleRequestPart3(req,resp,timeRecieved,requestTime,user_ip,user_ip_remote) ;
			
		}
		
		gotOtherPromise = true ;
		
	}) ;
	
}

function handleRequestPart3(req,resp,timeRecieved,requestTime,user_ip,user_ip_remote) {
	
	//If there are no account systems, then dont bother checking if the user has permission.
	if (allAccountSystems.length === 0) {
		
		let cont = true ;
		let gotOtherPromise = false ;
		externals.doEvt("allowedrequest",req,resp).then(d=>{
			
			if (d) {
				
				cont = false ;
				
			}
			
			if (gotOtherPromise && cont) {
				
				allowedRequest(req.host,req,resp,user_ip,user_ip_remote,timeRecieved) ;
				
			}
			
			gotOtherPromise = true ;
			
		}) ;
		
		externals.doEvt(`${req.host}/allowedrequest`,req,resp).then(d=>{
			
			if (d) {
				
				cont = false ;
				
			}
			
			if (gotOtherPromise && cont) {
				
				allowedRequest(req.host,req,resp,user_ip,user_ip_remote,timeRecieved) ;
				
			}
			
			gotOtherPromise = true ;
			
		}) ;
		
		return true ;
		
	}
	
	
	//Check the user is allowed to load the page.
	let checkingSystem = 0 ;
	
	//Function to load next check.
	let nextCheck =_=>{
		
		//Ask account system what to do.
		allAccountSystems[checkingSystem].doAnything(req,resp).then(returned=>{
			
			let canAccess = returned[0] ;
			
			//No perms
			if (canAccess === false) {
				
				if (returned[1] === "redirect") {
					
					console.log(`\t302 Found.   Redirecting to ${returned[2]}.`) ;
					resp.writeHead(302,{"Content-Type":"text/plain",location:returned[2]}) ;
					resp.write("Redirecting you to the login page.") ;
					resp.end() ;
					return ;
					
				}
				
				else {
					
					console.log(`\t401 Unauthorized.   Account system ${checkingSystem} denide access.`) ;
					resp.writeHead(401,{"Content-Type":"text/plain"}) ;
					resp.write("Nope.") ;
					resp.end() ;
					return false ;
					
				}
				
			}
			
			//Access is OK from this account system.
			else if (canAccess === true) {
				
				//Next system
				checkingSystem++ ;
				if (checkingSystem >= allAccountSystems.length) {
					
					//The request is allowed...
					let cont = true ;
					let gotOtherPromise = false ;
					externals.doEvt("allowedrequest",req,resp).then(d=>{
						
						if (d) {
							
							cont = false ;
							
						}
						
						if (gotOtherPromise && cont) {
							
							allowedRequest(req.host,req,resp,user_ip,user_ip_remote,timeRecieved) ;
							
						}
						
						gotOtherPromise = true ;
						
					})
					
					externals.doEvt(`${req.host}/allowedrequest`,req,resp).then(d=>{
						
						if (d) {
							
							cont = false ;
							
						}
						
						if (gotOtherPromise && cont) {
							
							allowedRequest(req.host,req,resp,user_ip,user_ip_remote,timeRecieved) ;
							
						}
						
						gotOtherPromise = true ;
						
					}) ;
					
				}
				
				else {
					
					//Check the next.
					nextCheck() ;
					
				}
				
			}
			
			//Dunnow ;)
			else if (canAccess === null) {
				
				return ;
				
			}
			
			//Somthing has gone wrong, so play it safe & send a 401.
			else {
				
				console.log(`${req.jpid}\t401 Unauthorized.   Account system ${checkingSystem} denide access.`) ;
				resp.writeHead(401,{"Content-Type":"text/plain"}) ;
				resp.write("Nope.") ;
				resp.end() ;
				return false ;
				
			}
			
		}).catch(err=>{coughtError(err,resp,req.jpid);console.log("Error trace: Error recieving account data.");}) ;
		
	} ;
	//Check
	nextCheck() ;
	
	return ;
	
}

//Should be called when a request is allowed.
function allowedRequest(host,req,resp,user_ip,user_ip_remote,timeRecieved) {
	
	try {
		
		//Is cached or special page.
		if (typeof pages[req.url.fullvalue] === "object") {
			
			//Is just page alias.
			if (pages[req.url.fullvalue][0] === "file") {
				
				req.url.fullvalue = req.host + pages[req.url.fullvalue][1] ;
				
			}
			
			//Check that the page is still in the cache after a change.
			if (typeof pages[req.url.fullvalue] === "object") {
				
				//If it is a cached file send it.
				if (pages[req.url.fullvalue][0] === "cache") {
					
					sendCache(req.url.fullvalue,pages[req.url.fullvalue][1],resp,resp.vars) ;
					return true ;
					
				}
				
				//If it is a function, then execute the function & dont continute if it returns true.
				else if (pages[req.url.fullvalue][0] === "func") {
					
					if (req.url.fullvalue,pages[req.url.fullvalue][1](req,resp)) {
						
						return true ;
						
					}
					
				}
				
			}
			
		}
		
		let normPath = path.normalize(req.url.value) ;
		let rID = req.jpid ;
		
		//Try the URL the user entered.
		sendFile(normPath,resp,resp.vars,rID).then(done=>{
			
			//If the file failed to send.
			if (!done[0]) {
				
				//Try with .page extention.
				return sendFile(`${normPath}.page`,resp,resp.vars,rID);
				
			}
			
			else {
				
				return [true] ;
				
			}
			
		}).then(done=>{
			
			//If the file failed to send.
			if (!done[0]) {
				
				//Try the index.html.
				return sendFile(path.join(normPath,"/index.html"),resp,resp.vars,rID);
				
			}
			
			else {
				
				return [true] ;
				
			}
			
		}).then(done=>{
			
			//If it still hasn't worked.
			if (!done[0]) {
				
				//Get the error code.
				let code = 500 ;
				if (done[1].code === "EACCES") {
					
					code = 403 ;
					
				}
				else if (done[1].code === "ENOENT") {
					
					code = 404 ;
					
				}
				//And send the error.
				sendError(code,errorCodes[code],resp,rID) ;
				
			}
			
			let timeTaken = process.hrtime(timeRecieved) ;
			console.log(`${rID}\tRequest took ${timeTaken[0] * 1000 + timeTaken[1] * 10e-6}ms to handle.`) ;
			
		}).catch(err=>{
			
			coughtError(err,resp,rID) ;
			console.log("Error trace: Error in send promise.") ;
			
		}) ;
		
	}
	
	catch(err) {
		
		coughtError(err,resp,req.jpid) ;
		console.log("Error trace: Request allowed, issue processubg headers.") ;
		
	}
	
}

/*
let testAccounts = new proc("dev","test.db",["www.jotpot.co.uk/software","www.jotpot.co.uk/games"],"www.jotpot.co.uk/testlogin","www.jotpot.co.uk/testloginpage.html","www.jotpot.co.uk/testlogout","www.jotpot.co.uk/testlogoutpage.html","www.jotpot.co.uk/testreg","www.jotpot.co.uk/testregpage.html") ;
let realmsLogin = new proc("realms","test.db",["realms.jotpot.co.uk*"],"realms.jotpot.co.uk/rlogin","realms.jotpot.co.uk/login.html","realms.jotpot.co.uk/logout","realms.jotpot.co.uk/login","realms.jotpot.co.uk/rreg","realms.jotpot.co.uk/login") ;

allAccountSystems.push(testAccounts) ;
allAccountSystems.push(realmsLogin) ;
*/

{
	
	for (let doing in config.cache) {
		
		pages[config.cache[doing]] = ["cache",fs.readFileSync(path.join("./sites",config.cache[doing]))]
		
	}

}


module.exports = {
	
	//Function to init the server.
	init:(clusterGiven) => {
		
		//Load the extentions
		let currentDir = fs.readdirSync(process.cwd()) ;
		//Go through all the files in the cwd.
		for (let doing in currentDir) {
			
			//If it is an extention, load it.
			if (currentDir[doing].substr(currentDir[doing].length - 7,7) === ".jpe.js") {
				
				externals.generateServerObject = _ => {
					
					return {
						
						"pages": pages,
						"vars": vars,
						"sendError":(...eArgs)=>sendError(...eArgs),
						"createAccountSystem":(args) => {
							
							let creatingAcc = new proc.proc(
								
								args.name,
								args.db,
								args.pages,
								args.exclude,
								args.loginURL,
								args.loginPage,
								args.logoutURL,
								args.logoutPage,
								args.regURL,
								args.regPage,
								args.loginRedirect,
								args.https
								
							) ;
							allAccountSystems.push(creatingAcc) ;
							return creatingAcc ;
							
						},
						"getUserID":(req,resp)=>{
							
							let userID = proc.getUserID(req) ;
							if (userID) {
								
								return userID ;
								
							}
							return proc.makeNewUID(req,resp) ;
							
						},
						"getMimeType": (...args)=>getMimeType(...args),
						"config": config,
						"getData": req=>{
							
							let data = new Array() ;
							req.on("data",d=>data.push(d)) ;
							return new Promise(resolve=>req.on("end",_=>resolve(Buffer.concat(data)))) ;
							
						},
						"reloadConfig": _=>loadConfig(),
						"multipartFormDataParser": require("./multipart-form-data-parser.js"),
						"sendFile": (file, resp, req={jpid:""}) => sendFile(file, resp, resp.vars, req.jpid),
						"sendCache": (file, cache, resp, status=200, req={jpid:""}) => sendCache(file, cache, resp, resp.vars, status, req.jpid)
						
					} ;
					
				} ;
				
				let currentLimitedAccountSystem = 0 ;
				externals.generateLimitedServerObject = (domains, fs) => {
					
					return {
						
						"sendError":(...eArgs)=>sendError(...eArgs),
						"createAccountSystem":(args) => {
							
							let toCheck = ["pages", "exclude","loginURL","loginPage","logoutURL","logoutPage","regURL","regPage","loginRedirect"] ;
							
							for (let checking of toCheck) {
								
								if (typeof args[checking] === "string") {
									
									let ok = false ;
									for (let domain in domains) {
										
										if (args[checking].indexOf(domains[domain]) === 0) {
											
											ok = true ;
											break ;
											
										}
										
									}
									if (!ok) {
										
										throw new Error("You cannot set up an account system using these domains...") ;
										
									}
									continue ;
									
								}
								
								for (let doing = 0 ; doing < args[checking].length ; doing++) {
									
									let ok = false ;
									for (let domain in domains) {
										
										if (args[checking][doing].indexOf(domains[domain]) === 0) {
											
											ok = true ;
											break ;
											
										}
										
									}
									if (!ok) {
										
										throw new Error("You cannot set up an account system using these domains...") ;
										
									}
									
								}
								
							}
							
							let creatingAcc = new proc.proc(
								
								`limitedsystem-${++currentLimitedAccountSystem}`,
								fs.realpathSync(args.db, {real:true}),
								args.pages,
								args.exclude,
								args.loginURL,
								args.loginPage,
								args.logoutURL,
								args.logoutPage,
								args.regURL,
								args.regPage,
								args.loginRedirect,
								args.https
								
							) ;
							allAccountSystems.push(creatingAcc) ;
							return creatingAcc ;
							
						},
						"getUserID":(req,resp)=>{
							
							let userID = proc.getUserID(req) ;
							if (userID) {
								
								return userID ;
								
							}
							return proc.makeNewUID(req,resp) ;
							
						},
						"getMimeType":(...args)=>getMimeType(...args),
						"getData": req=>{
							
							let data = new Array() ;
							req.on("data",d=>data.push(d)) ;
							return new Promise(resolve=>req.on("end",_=>resolve(Buffer.concat(data)))) ;
							
						},
						"multipartFormDataParser": require("./multipart-form-data-parser.js")//,
						//"sendFile": (file, resp, req={jpid:""}) => sendFile(path.normalize(file).replace(/\.\./g,""), resp, resp.vars, req.jpid)
						
					} ;
					
				} ;
				
				let currentLoad = externals.loadExt(currentDir[doing]) ;
				
				if (currentLoad.loaded) {
					
					pages = currentLoad.serverObj.pages ;
					vars = currentLoad.serverObj.vars ;
					
				}
				
			}
			
		}
		
		//Set up the HTTP servers
		for (let doing in config.httpServers) {
			
			let options = new Array() ;
			options[0] = config.httpServers[doing].port ;
			if (typeof config.httpServers[doing].host !== "undefined") {
				
				options[1] = config.httpServers[doing].host ;
				
			}
			
			http.createServer((req,resp) => {
				
				req.overHttps = req.secure = req.secureToServer = false ;
				req.port = options[0] ;
				handleRequest(req,resp) ;
				
			}).listen(...options) ;
			
		}
		
		//Set up the HTTPS servers
		for (let doing in config.httpsServers) {
			
			https.createServer({key:fs.readFileSync("privkey.pem"),ca:fs.readFileSync("fullchain.pem"),cert:fs.readFileSync("cert.pem")},(req,resp) => {
				
				req.overHttps = req.secure = req.secureToServer = true ;
				req.port = config.httpsServers[doing].port ;
				handleRequest(req,resp) ;
				
			}).listen(config.httpsServers[doing].port) ;
			
		}
		
		//Get the cluster module of parent.
		cluster = clusterGiven ;
		vars.Global["thread_id"] = cluster.worker.id ;
		
		externals.doEvt("ready") ;

	}

}
