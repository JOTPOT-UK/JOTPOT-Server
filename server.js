/*
	
	JOTPOT Server
	Version 26A-0
	
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
	
} ;

global.requireJPS = mod => require(path.join(__dirname, mod)) ;
const loadAJSONFile = p => {
	if (fs.existsSync(p)) {
		return fs.readFileSync(p).toString() ;
	}
	return fs.readFileSync(path.join(__dirname, p)).toString() ;
} ;
const doesAJSONFileExist = p => (fs.existsSync(p) || fs.existsSync(path.join(__dirname, p))) ;

//Node Modules
let http = require("http") ;
let https = require("https") ;
let fs = require("fs") ;
let path = require("path") ;
let {Transform} = require("stream") ;
let cluster ;

//JPS Modules
let proc = requireJPS("accounts") ;
let externals = requireJPS("externals") ;
let CORS = requireJPS("cors") ;
let responseMaker = requireJPS("do-response") ;
let parseFlags = requireJPS("flag-parser") ;
let jpsUtil = requireJPS("jps-util") ;
let urlObject = requireJPS("url-object") ;
let {URL} = urlObject ;

//Load the config
let config ;

//Default configuration
const defaultConfig = {
	
	"dataPort": 500,
	"controlers":["::1","127.0.0.1","::ffff:127.0.0.1"],
	
	"httpServers": [
		{
			"port": 80
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

//Load the comfig and fill in any blanks. If it doesn't exist, set the config to the default config.
function loadConfig() {
	
	//If it exists, load it, parse it and fill in any blanks or throw if the types aren't correct
	if (doesAJSONFileExist("config.json")) {
		
		config = loadAJSONFile("config.json") ;
		
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
			process.exit(1) ;
			
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

	else {
		
		console.warn("Config file does not exist, using default config.") ;
		config = new Object() ;
		Object.assign(config, defaultConfig) ;
		
		
	}
	
}
loadConfig() ;

let flags = parseFlags() ;
if (flags["-http-port"]) {
	if (flags["-port"]) {
		flags["-port"] = flags["-port"].concat(flags["-http-port"]) ;
	} else {
		flags["-port"] = flags["-http-port"] ;
	}
}
if (flags["-port"]) {
	config.httpServers = [] ;
	let port ;
	for (let doing in flags["-port"]) {
		port = parseInt(flags["-port"][doing], 10) ;
		if (!isNaN(port)) {
			config.httpServers.push({
				"port": port
			}) ;
		}
	}
}

let availHosts = [] ;
if (config.useDefaultHostIfHostDoesNotExist || config.fallbackToNoPort) {
	
	availHosts = fs.readdirSync("./sites") ;
	
}

//Vars to add to files
let vars = new Object ;
vars.Global = new Object() ;

//Setup accounts
let allAccountSystems = new Array() ;

//Vars config
let doVarsFor = ["error_page"].concat(config.doVarsForIfNotByDefault) ;
for (let doing in doVarsFor) {
	
	doVarsFor[doing] = path.normalize(doVarsFor[doing]) ;
	doVarsFor[doing] = path.join(process.cwd(),"sites",doVarsFor[doing]) ;
	
}
//let dontDoVarsFor = [] ;

//Currently imprelemted methods
const defaultMethods = ["GET", "POST", "HEAD", "OPTIONS"] ;
let implementedMethods = {} ;

//Set up current ID
let currentID = 0 ;

//Get buffer of string for inserting vars.
let startOfVar = Buffer.from("$::") ;
let endOfVar = Buffer.from("::$") ;

//Error file
let errorFile ;
if (!doesAJSONFileExist(config.errorTemplate)) {
	
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
	
	errorFile = loadAJSONFile(config.errorTemplate) ;
	
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
			
			throw new Error("DUDE!!! WHAT IS THE PATH YOU ARE PIPING TO ME??? (Top tip, first argument needs to be the path.)") ;
			
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
			
			for (;;) {
				
				if (data.indexOf(startOfVar) !== -1) {
					
					if (data.indexOf(endOfVar) !== -1) {
						
						let dataString = this.lastData.toString() + data.toString() ;
						
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
						
						this.lastData = data ;
						doPush = false ;
						break ;
						
					}
					
				}
				
				else {
					
					break ;
					
				}
				
			}
			
			if (doPush) {
				this.push(data) ;
			}
			
			callback() ;
			return ;
			
		}
		
		catch(err) {
			
			console.warn("Error in pipe that adds vars") ;
			
		}
		
	}
	
}

if (!doesAJSONFileExist("mimes.json")) {
	throw new Error("mimes.json file not found in CWD in " + __dirname) ;
}

//Sorting out mime types.
let mimes = JSON.parse(loadAJSONFile("mimes.json")) ;
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
		
		
			
		return "text/plain" ;
			
		
		
	}
	
	catch(err) {
		
		console.warn(err) ;
		return "text/plain" ;
		
	}
	
}

//Add the URL property to a request object. Make it so that if it is set, it sets the value instead.
function wrapURL(req, secure) {
	const defaultProtocols = ["http:", "https:"] ;
	const secureProtocols = ["https:", "sftp:"] ;
	
	req.overHttps = secure ;
	let url = new URL(req, config.defaultHost || config.defaultDomain) ;
	Object.defineProperty(req, "url", {
		enumerable: true,
		configurable: false,
		get: ()=>url,
		set: v=>{url.value=v;}
	}) ;
	
	const setSecure = val => {
		if (defaultProtocols.indexOf(req.url.protocol) !== -1) {
			req.url.protocol = val?"https:":"http:" ;
		}
	} ;
	const getSecure = () => (secureProtocols.indexOf(req.url.protocol) !== -1) ;
	Object.defineProperty(req, "overHttps", {
		get: getSecure,
		set: setSecure,
		enumerable: true,
		configurable: false
	}) ;
	Object.defineProperty(req, "secure", {
		get: getSecure,
		set: setSecure,
		enumerable: true,
		configurable: false
	}) ;
	Object.defineProperty(req, "secureToServer", {
		value: secure,
		writable: false,
		enumerable: true,
		configurable: false
	}) ;
	
	Object.defineProperty(req, "uri", {
		get: ()=>req.url.path,
		set: val=>req.url.path=val,
		enumerable: true,
		configurable: false
	}) ;
}

//Pipes the file through the transform pipe into the main pipe.
//Calls the callback with the first argument as a boolean - true if succeded, false if not.
function getFile(file,callWithStats,pipeTo,callback,range=null) {
	
	fs.stat(file,(err,stats) => {
		
		if (err) {
			
			callback(false,err) ;
			return ;
			
		}
		
		if (stats.isFile()) {
			
			if (callWithStats(stats)) {
				
				let opts = {
					flags: "r",
					autoClose: true
				} ;
				
				if (range !== null) {
					if (isNaN(range[1])) {
						range[1] = stats.size - 1 ;
					}
					opts.start = range[0] ;
					opts.end = range[1] ;
					if (opts.end >= stats.size) {
						range[1] = stats.size - 1 ;
					}
				}
				
				//Pipe file to the pipe.
				fs.createReadStream(file, opts).pipe(pipeTo) ;
				
				callback(true, null) ;
				return ;
				
			}
			
			callback(false, "CBR") ;
			return ;
			
		}
		callback(false,"DIR") ;
		
	}) ;
	
}

//Sends the file specified to the pipe as the second argument - goes through the getFile & thus vars pipe.
function sendFile(file,resp,customVars,req) {
	
	try {
		
		//Look in the sites dir.
		let start = path.join(process.cwd(), "sites") ;
		file = path.join(start,URL.toDir(file)) ;
		
		//If we aren't in the sites dir, then throw.
		if (file.indexOf(start) !== 0) {
			
			throw new Error("Server has left the serving directory!") ;
			
		}
		
		//Make a pipe to send it to.
		let mainPipe ;
		let doingTransform = resp.pipeThrough.length - 1 ;
		let lengthknown = false ;
		let ranges = null ;
		let status = 200 ;
		
		//Only bother with the pipes if we are accualy sending the body
		if (resp.sendBody) {
			
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
			
		}
		
		else if (!(config.addVarsByDefault || doVarsFor.indexOf(file) !== -1) && doingTransform < 0) {
			
			lengthknown = true ;
			
		}
		
		if (lengthknown && typeof req.headers.range === "string") {
			let rangesArr = req.headers.range.split("=") ;
			
			//We can only use bytes as a range value
			if (rangesArr[0] !== "bytes") {
				sendError(416, `${rangesArr[0]} is not a valid range unit.`, resp, req.jpid) ;
				return new Promise((_, reject)=>reject()) ;
			}
			
			//Create array of all the range values
			rangesArr = rangesArr[1].split(", ") ;
			
			//If there is only 1 range
			if (rangesArr.length === 1) {
				//Split the start from the end
				rangesArr = rangesArr[0].split("-") ;
				//Only carry on if we have a start and end
				if (rangesArr.length === 2) {
					ranges = new Array(2) ;
					//If there is no value or the value isn't a number, set the value to the start or end. Otherwise, set the value
					if (rangesArr[0] === "") {
						ranges[0] = 0 ;
					} else {
						let toSetTo = parseInt(rangesArr[0], 10) ;
						if (isNaN(toSetTo)) {
							ranges[0] = 0 ;
						} else {
							ranges[0] = toSetTo ;
						}
					}
					if (rangesArr[1] === "") {
						rangesArr[1] = NaN ;
					} else {
						let toSetTo = parseInt(rangesArr[1], 10) ;
						if (isNaN(toSetTo)) {
							ranges[1] = NaN ;
						} else {
							ranges[1] = toSetTo ;
						}
					}
					//If it is the entire document, then don't bother with the ranges
					if (ranges[0] === 0 && isNaN(ranges[1])) {
						ranges = null ;
					} else if (ranges[0] > ranges[1]) {
						ranges = null ;
					} else {
						//Otherwise, set the status and relivent headers
						status = 206 ;
					}
				}
			} else if (rangesArr.length > 1) {
				return new Promise((resolve, reject)=>{
					fs.stat(file, (err, stats) => {
						if (err) {
							resolve(false, err) ;
							return ;
						}
						if (!stats.isFile()) {
							resolve(false, "DIR") ;
							return ;
						}
						let cont = true ;
						for (let doing in rangesArr) {
							rangesArr[doing] = rangesArr[doing].split("-") ;
							if (rangesArr[doing][0] === "") {
								rangesArr[doing][0] = 0 ;
							} else {
								let toStartAt = parseInt(rangesArr[doing][0], 10) ;
								if (isNaN(toStartAt)) {
									rangesArr[doing][0] = 0 ;
								} else {
									rangesArr[doing][0] = toStartAt ;
								}
							}
							if (rangesArr[doing][1] === "") {
								rangesArr[doing][1] = stats.size - 1 ;
							} else {
								let toEndAt = parseInt(rangesArr[doing][1], 10) ;
								if (isNaN(toEndAt)) {
									rangesArr[doing][1] = stats.size - 1 ;
								} else {
									rangesArr[doing][1] = toEndAt ;
								}
							}
							if (rangesArr[doing][0] === 0 && rangesArr[doing][1] === stats.size - 1) {
								cont = false ;
								break ;
							} else if (rangesArr[doing][0] > rangesArr[doing][1]) {
								cont = false ;
								break ;
							}
							//▀‗ð►╠Ä#/'╠╩♦♀0╦┐¶ýÄ↔A8─oeÀ╚Ä´Há*
							//«Þø[§
						}
						if (cont) {
							try {
								const boundary = config.multipartResponseBoundary || "58dca288fd8c0f00" ;
								const mime = resp.forceDownload?"application/octet-stream":getMimeType(file) ;
								const getBoundary = (start, end) => `\r\n--${boundary}\r\nContent-Type: ${mime}\r\nContent-Range: bytes ${start}-${end}/${stats.size}\r\n\r\n` ;
								let length = 0 ;
								for (let doing in rangesArr) {
									length += getBoundary(rangesArr[doing][0], rangesArr[doing][1]).length ;
									length += rangesArr[doing][1] - rangesArr[doing][0] + 1 ;
								}
								length += (`\r\n--${boundary}--\r\n`).length ;
								resp.writeHead(206, {
									"Accept-Ranges": "bytes",
									"Content-Type": `multipart/byteranges; boundary=${boundary}`,
									"Content-Length": length,
									"Status": 206
								}) ;
								if (!resp.sendBody) {
									resp.end() ;
									return ;
								}
								let doing = -1 ;
								const nextWrite = () => {
									doing++ ;
									if (doing >= rangesArr.length) {
										resp.write(`\r\n--${boundary}--\r\n`) ;
										resp.end() ;
										return ;
									}
									resp.write(getBoundary(rangesArr[doing][0], rangesArr[doing][1])) ;
									let reader = fs.createReadStream(file, {
										flags: "r",
										autoClose: true,
										start: rangesArr[doing][0],
										end: rangesArr[doing][1]
									}) ;
									reader.on("data", d=>resp.write(d)) ;
									reader.on("end", nextWrite) ;
								} ;
								nextWrite() ;
								resolve([true, null]) ;
							} catch (err) {
								reject(err) ;
							}
						}
					}) ;
				}) ;
			}
		}
		
		return new Promise(resolve => {
				
			getFile(file,stats => {
				
				const mime = resp.forceDownload?"application/octet-stream":getMimeType(file) ;
				console.log(`${req.jpid}\t${status} ${http.STATUS_CODES[status]}.   ${file} (${mime}) loaded from disk.`) ;
				if (status === 206) {
					resp.setHeader("Content-Range", `bytes ${ranges[0]}-${isNaN(ranges[1])?(stats.size-1):Math.min(ranges[1],stats.size-1)}/${stats.size}`) ;
					resp.setHeader("Content-Length", (isNaN(ranges[1])?(stats.size-1):Math.min(ranges[1],stats.size-1)) - ranges[0] + 1) ;
				} else if (lengthknown) {
					resp.setHeader("Content-Length", stats.size) ;
				}
				resp.writeHead(status,{
					
					"Content-Type": mime,
					"Accept-Ranges": lengthknown?"bytes":"none",
					
					//Added because google does it :)
					"Status": status
					
				}) ;
				
				if (!resp.sendBody) {
					
					resp.end() ;
					return false ;
					
				}
				
				return true ;
				
			},mainPipe,(done,err) => {
				
				//If we judt didn't need to send the body, then there isn't an error
				if (!resp.sendBody && !done && err === "CBR") {
					
					resolve([true,null]) ;
					
				}
				
				else {
					
					resolve([done,err]) ;
					
				}
				
			}, ranges);
			
		}) ;
		
	}
	
	catch(err) {
		
		//coughtError(err,resp,rID) ;
		//Return an instantly rejecting promise
		return new Promise((resolve, reject) => reject(err)) ;
		
	}
	
}

function sendCache(file,cache,resp,customVars,req,status=200) {
	
	try {
		
		//Look in the sites dir.
		file = path.join(process.cwd(),"sites",file) ;
		
		
		//Make a pipe to send it to.
		let mainPipe ;
		resp.pipeThrough = [] ;
		let doingTransform = resp.pipeThrough.length - 1 ;
		let lengthknown = false ;
		let ranges = null ;
		
		//Only bother with the pipes if we accualy have to send the requests
		if (resp.sendBody) {
			
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
			
		}
		
		else if (!(config.addVarsByDefault || doVarsFor.indexOf(file) !== -1) && doingTransform < 0) {
			lengthknown = true ;
			mainPipe = resp ;
		} else {
			mainPipe = resp ;
		}
		
		if (lengthknown && typeof req.headers.range === "string") {
			let rangesArr = req.headers.range.split("=") ;
			
			//We can only use bytes as a range value
			if (rangesArr[0] !== "bytes") {
				sendError(416, `${rangesArr[0]} is not a valid range unit.`, resp, req.jpid) ;
				return ;
			}
			
			//Create array of all the range values
			rangesArr = rangesArr[1].split(", ") ;
			
			//If there is only 1 range
			if (rangesArr.length === 1) {
				//Split the start from the end
				rangesArr = rangesArr[0].split("-") ;
				//Only carry on if we have a start and end
				if (rangesArr.length === 2) {
					ranges = new Array(2) ;
					//If there is no value or the value isn't a number, set the value to the start or end. Otherwise, set the value
					if (rangesArr[0] === "") {
						ranges[0] = 0 ;
					} else {
						let toSetTo = parseInt(rangesArr[0], 10) ;
						if (isNaN(toSetTo)) {
							ranges[0] = 0 ;
						} else {
							ranges[0] = toSetTo ;
						}
					}
					if (rangesArr[1] === "") {
						rangesArr[1] = cache.length - 1 ;
					} else {
						let toSetTo = parseInt(rangesArr[1], 10) ;
						if (isNaN(toSetTo)) {
							ranges[1] = cache.length - 1 ;
						} else {
							ranges[1] = toSetTo ;
						}
					}
					//If it is the entire document, then don't bother with the ranges
					if (ranges[0] === 0 && ranges[1] === cache.length - 1) {
						ranges = null ;
					} else if (ranges[0] > ranges[1]) {
						ranges = null ;
					} else {
						//Otherwise, set the status and relivent headers
						status = 206 ;
					}
				}
			} else if (rangesArr.length > 1) {
				let cont = true ;
				for (let doing in rangesArr) {
					rangesArr[doing] = rangesArr[doing].split("-") ;
					if (rangesArr[doing][0] === "") {
						rangesArr[doing][0] = 0 ;
					} else {
						let toStartAt = parseInt(rangesArr[doing][0], 10) ;
						if (isNaN(toStartAt)) {
							rangesArr[doing][0] = 0 ;
						} else {
							rangesArr[doing][0] = toStartAt ;
						}
					}
					if (rangesArr[doing][1] === "") {
						rangesArr[doing][1] = cache.length - 1 ;
					} else {
						let toEndAt = parseInt(rangesArr[doing][1], 10) ;
						if (isNaN(toEndAt)) {
							rangesArr[doing][1] = cache.length - 1 ;
						} else {
							rangesArr[doing][1] = toEndAt ;
						}
					}
					if (rangesArr[doing][0] === 0 && rangesArr[doing][1] === cache.length - 1) {
						cont = false ;
						break ;
					} else if (rangesArr[doing][0] > rangesArr[doing][1]) {
						cont = false ;
						break ;
					}
				}
				if (cont) {
					const boundary = config.multipartResponseBoundary || "58dca288fd8c0f00" ;
					const mime = resp.forceDownload?"application/octet-stream":getMimeType(file) ;
					const getBoundary = (start, end) => `\r\n--${boundary}\r\nContent-Type: ${mime}\r\nContent-Range: bytes ${start}-${end}/${cache.length}\r\n\r\n` ;
					let length = 0 ;
					for (let doing in rangesArr) {
						length += getBoundary(rangesArr[doing][0], rangesArr[doing][1]).length ;
						length += rangesArr[doing][1] - rangesArr[doing][0] + 1 ;
					}
					length += (`\r\n--${boundary}--\r\n`).length ;
					resp.writeHead(206, {
						"Accept-Ranges": "bytes",
						"Content-Type": `multipart/byteranges; boundary=${boundary}`,
						"Content-Length": length,
						"Status": 206
					}) ;
					if (!resp.sendBody) {
						resp.end() ;
						return ;
					}
					let doing = -1 ;
					const next = () => {
						doing++ ;
						if (doing >= rangesArr.length) {
							resp.write(`\r\n--${boundary}--\r\n`) ;
							resp.end() ;
							return ;
						}
						resp.write(getBoundary(rangesArr[doing][0], rangesArr[doing][1])) ;
						resp.write(cache.slice(rangesArr[doing][0], rangesArr[doing][1] + 1)) ;
						next() ;
					} ;
					next() ;
					return ;
				}
			}
		}
		
		//Get the mime type.
		const mime = resp.forceDownload?"application/octet-stream":getMimeType(file) ;
		console.log(`${req.jpid}\t${status} ${http.STATUS_CODES[status]}.   ${file} (${mime}) loaded from cache.`) ;
		
		if (status === 206) {
			resp.setHeader("Content-Range", `bytes ${ranges[0]}-${Math.min(ranges[1],cache.length-1)}/${cache.length}`) ;
			resp.setHeader("Content-Length", Math.min(ranges[1],cache.length-1) - ranges[0] + 1) ;
		} else if (lengthknown) {
			resp.setHeader("Content-Length", cache.length) ;
		}
		resp.writeHead(status,{
			
			"Content-Type": mime,
			"Accept-Ranges": lengthknown?"bytes":"none",
			
			//Added because google does it :)
			"Status": status
			
		}) ;
		
		//Write the cached data (if we need to) & end.
		if (resp.sendBody) {
			if (ranges === null) {
				mainPipe.write(cache) ;
			} else {
				mainPipe.write(cache.slice(ranges[0], ranges[1] + 1)) ;
			}
		}
		mainPipe.end() ;
		
	}
	
	catch(err) {
		
		coughtError(err,resp) ;
		
	}
	
}

function sendError(code,message,resp,rID="") {
	
	sendCache("error_page",errorFile,resp,{error_code:code,error_type:http.STATUS_CODES[code],error_message:message},{jpid:rID,headers:{}},code) ;
	
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
	sendError(500,`A${isUnknown?"n unknown":" known "} error occured.${isUnknown?"":" I just don't want to tell you what went wrong. Nothing personal, honestly! It's not like I don't trust you."}.`,resp,rID) ;
	
}

function make6d(str, pad="0") {
	while (str.length < 6) {
		str = pad + str ;
	}
	return str ;
}

//Function to handle http requests.
function handleRequest(req,resp,secure) {
	
	try {
		
		//Get time stuff.
		let timeRecieved = process.hrtime() ;
		let requestTime = new Date() ;
		
		//Set server header
		resp.setHeader("Server","JOTPOT Server") ;
		
		//Get IP, follow jp-source headers if we are behind a load balancer
		let user_ip, user_ip_remote ;
		if (config.behindLoadBalancer) {
			
			user_ip = (req.headers["x-forwarded-for"] || req.headers["jp-source-ip"] || req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress).replace(/::ffff:/g,"") ;
			user_ip_remote = (req.headers["jp-source-ip"] || req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress).replace(/::ffff:/g,"") ;
			secure = req.headers["jp-source-secure"] === "https" ;
			
		} else {
			
			user_ip = (req.headers["x-forwarded-for"] || req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress).replace(/::ffff:/g,"") ;
			user_ip_remote = (req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress).replace(/::ffff:/g,"") ;
			
		}
		
		//Add stuff to the req object
		req.ip = user_ip ;
		req.remoteAddress = user_ip_remote ;
		req.usePortInDirectory = true ;
		//Create URL object and secure, and overHttps and secureToServer
		wrapURL(req, secure) ;
		//orig_url object
		let orig_url = new Object() ;
		Object.assign(orig_url, req.url) ;
		Object.defineProperty(req, "orig_url", {
			enumerable: true,
			configurable: false,
			writable: false,
			value: orig_url
		}) ;
		
		//Add stuff to resp object.
		resp.vars = {"user_ip":user_ip,"user_ip_remote":user_ip_remote,"time":requestTime.getTime().toString(),"href":req.url.href,"method":req.method} ;
		resp.pipeThrough = new Array() ;
		resp.forceDownload = false ;
		resp.sendBody = true ;
		
		//Set default headers
		for (let doing in config.defaultHeaders) {
			resp.setHeader(doing, config.defaultHeaders[doing]) ;
		}
		
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
			
		}) ;
		
		externals.doEvt(`${req.url.host}/request`,req,resp).then(d=>{
			
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
	//And replace // with /, along with removing any trailing /
	do {
		req.url.pathname = req.url.pathname.replace(/\.\./g, "").replace(/\/\//g, "/") ;
	} while (req.url.pathname.indexOf("//") !== -1) ;
	while (req.url.pathname.length > 1 && req.url.pathname.indexOf("/") === req.url.pathname.length - 1) {
		req.url.pathname = req.url.pathname.substring(0, req.url.pathname.length - 1) ;
	}
	
	//Should we redirect to https.
	if (!req.overHttps && config.dontRedirect.indexOf(req.url.value) === -1) {
		
		if (config.mustRedirectToHttps.indexOf(req.url.host) !== -1) {
			
			console.log(`${req.jpid}\tfrom ${user_ip_remote}(${user_ip}) for ${req.url.value} being handled by thread ${cluster.worker.id}.`) ;
			
			req.url.protocol = "https:" ;
			if (req.url.port === 80) {
				req.url.port = 443 ;
			}
			
			console.log(`${req.jpid}\t302 Found.   Redirecting to ${req.url.location} because this page MUST be HTTPS.`) ;
			
			resp.writeHead(302, {"Content-Type": "text/plain", "location": req.url.location, "Status": 302}) ;
			resp.write("Redirecting you to our secure site...") ;
			resp.end() ;
			
			let timeTaken = process.hrtime(timeRecieved) ;
			console.log(`${req.jpid}\tRequest took ${timeTaken[0] * 1000 + timeTaken[1] * 10e-6}ms to handle.`) ;
			
			return ;
			
		} else if (config.redirectToHttps.indexOf(req.url.host) !== -1 && req.headers["upgrade-insecure-requests"] && req.headers["upgrade-insecure-requests"] === '1') {
			
			console.log(`${req.jpid}\tfrom ${user_ip_remote}(${user_ip}) for ${req.url.value} being handled by thread ${cluster.worker.id}.`) ;
			
			req.url.protocol = "https:" ;
			if (req.url.port === 80) {
				req.url.port = 443 ;
			}
			
			console.log(`${req.jpid}\t307 Moved Temporarily.   Redirecting to ${req.url.location} because the user requested an upgrade.`) ;
			
			resp.writeHead(307,{"Content-Type": "text/plain", "location": req.url.location, "Status": 307, "Vary": "Upgrade-Insecure-Requests"}) ;
			resp.write("Redirecting you to our secure site...") ;
			resp.end() ;
			
			let timeTaken = process.hrtime(timeRecieved) ;
			console.log(`${req.jpid}\tRequest took ${timeTaken[0] * 1000 + timeTaken[1] * 10e-6}ms to handle.`) ;
			
			return ;
			
		}
		
	}
	
	//Is the host an alias?
	while (typeof config.hostAlias[req.url.host] !== "undefined") {
		req.url.host = config.hostAlias[req.url.host] ;
	}
	//Is the hostname an alias?
	while (typeof config.hostnameAlias[req.url.hostname] !== "undefined") {
		req.url.hostname = config.hostnameAlias[req.url.hostname] ;
	}
	
	//If we might need to fallback and the host doesn't exist
	if (config.fallbackToNoPort && availHosts.indexOf(URL.toDir(req.url.host)) === -1) {
		
		//But if we ignore the port and it still doesn't, and we should fallback to default, then fallback to default.
		if (availHosts.indexOf(URL.toDir(req.url.hostname)) === -1 && config.useDefaultHostIfHostDoesNotExist) {
			req.url.host = config.defaultHost || "default:0" ;
			if (availHosts.indexOf(URL.toDir(req.url.host)) === -1) {
				req.usePortInDirectory = false ;
			}
		}
		
		//Otherwise, it does exist so lets not use the port
		else {
			
			req.usePortInDirectory = false ;
			
		}
		
	} else if (config.useDefaultHostIfHostDoesNotExist) {
		//If we are set to goto a default host, check if the host doesn't exist, if so, we are now default :)
		if (availHosts.indexOf(URL.toDir(req.url.host)) === -1) {
			req.url.host = config.defaultHost || "default:0" ;
			if (availHosts.indexOf(URL.toDir(req.url.host)) === -1) {
				req.usePortInDirectory = false ;
			}
		}
	}
	
	//Should we redirect to another host.
	if (typeof config.hostRedirects[req.url.host] !== "undefined" && config.dontRedirect.indexOf(req.url.value) === -1) {
		
		console.log(`${req.jpid}\tfrom ${user_ip_remote}(${user_ip}) for ${req.url} being handled by thread ${cluster.worker.id}.`) ;
		
		//Set new host
		req.url.host = config.hostRedirects[req.url.host] ;
		
		//Set correct protocol
		let isRedirectHttps = req.headers["upgrade-insecure-requests"] && req.headers["upgrade-insecure-requests"] === '1' && (config.redirectToHttps.indexOf(req.url.host) !== -1 || config.mustRedirectToHttps.indexOf(req.url.host) !== -1) ;
		req.url.protocol = (isRedirectHttps||req.secure)?"https:":"http:" ;
		if ((isRedirectHttps||req.secure) && req.url.port === 80) {
			req.url.port = 443 ;
		}
		
		//And send response
		console.log(`${req.jpid}\t302 Found.   Redirecting to ${req.url.location} because of a host redirect.`) ;
		resp.writeHead(302, {"Content-Type": "text/plain", "location": req.url.location, "Status": 302}) ;
		resp.write("Redirecting you to " + req.url.location + "...") ;
		resp.end() ;
		
		let timeTaken = process.hrtime(timeRecieved) ;
		console.log(`${req.jpid}\tRequest took ${timeTaken[0] * 1000 + timeTaken[1] * 10e-6}ms to handle.`) ;
		
		return ;
		
	}
	
	if (typeof config.hostnameRedirects[req.url.hostname] !== "undefined" && config.dontRedirect.indexOf(req.url.value) === -1) {
		
		console.log(`${req.jpid}\tfrom ${user_ip_remote}(${user_ip}) for ${req.url} being handled by thread ${cluster.worker.id}.`) ;
		
		//Set new host
		req.url.hostname = config.hostnameRedirects[req.url.hostname] ;
		
		//Set correct protocol
		let isRedirectHttps = req.headers["upgrade-insecure-requests"] && req.headers["upgrade-insecure-requests"] === '1' && (config.redirectToHttps.indexOf(req.url.host) !== -1 || config.mustRedirectToHttps.indexOf(req.url.host) !== -1) ;
		req.url.protocol = (isRedirectHttps||req.secure)?"https:":"http:" ;
		if ((isRedirectHttps||req.secure) && req.url.port === 80) {
			req.url.port = 443 ;
		}
		
		//And send response
		console.log(`${req.jpid}\t302 Found.   Redirecting to ${req.url.location} because of a hostname redirect.`) ;
		resp.writeHead(302, {"Content-Type": "text/plain", "location": req.url.location, "Status": 302}) ;
		resp.write("Redirecting you to " + req.url.location + "...") ;
		resp.end() ;
		
		let timeTaken = process.hrtime(timeRecieved) ;
		console.log(`${req.jpid}\tRequest took ${timeTaken[0] * 1000 + timeTaken[1] * 10e-6}ms to handle.`) ;
		
		return ;
		
	}
	
	responseMaker.doLinks(req) ;
	
	{
		let rID = `#${cluster.worker.id}-${make6d((currentID++).toString(16).toUpperCase())}` ;
		Object.defineProperty(req, "jpid", {
			configurable: false,
			enumerable: false,
			value: rID,
			writable: false
		}) ;
	}
	
	console.log(`${req.jpid}\tfrom ${user_ip_remote}(${user_ip}) for ${req.url.value} being handled by thread ${cluster.worker.id}.`) ;
	
	CORS.setHeaders(req, resp) ;
	
	
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
		
	}) ;
	
	externals.doEvt(`${req.url.host}/fullrequest`,req,resp).then(d=>{
		
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
				
				allowedRequest(req.url.host,req,resp,user_ip,user_ip_remote,timeRecieved,false) ;
				
			}
			
			gotOtherPromise = true ;
			
		}) ;
		
		externals.doEvt(`${req.url.host}/allowedrequest`,req,resp).then(d=>{
			
			if (d) {
				
				cont = false ;
				
			}
			
			if (gotOtherPromise && cont) {
				
				allowedRequest(req.url.host,req,resp,user_ip,user_ip_remote,timeRecieved,false) ;
				
			}
			
			gotOtherPromise = true ;
			
		}) ;
		
		return ;
		
	}
	
	
	//Check the user is allowed to load the page.
	let checkingSystem = 0 ;
	
	//Function to load next check.
	let nextCheck = () =>{
		
		//Ask account system what to do.
		allAccountSystems[checkingSystem].doAnything(req,resp).then(returned=>{
			
			let canAccess = returned[0] ;
			
			//No perms
			if (canAccess === false) {
				
				if (returned[1] === "redirect") {
					
					console.log(`\t302 Found.   Redirecting to ${returned[2]} because of account system.`) ;
					resp.writeHead(302,{"Content-Type":"text/plain",location:returned[2]}) ;
					resp.write("Redirecting you to the login page.") ;
					resp.end() ;
					
				}
				
				else {
					
					console.log(`\t401 Unauthorized.   Account system ${checkingSystem} denide access.`) ;
					resp.writeHead(401,{"Content-Type":"text/plain"}) ;
					resp.write("Nope.") ;
					resp.end() ;
					
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
							
							allowedRequest(req.url.host,req,resp,user_ip,user_ip_remote,timeRecieved,false) ;
							
						}
						
						gotOtherPromise = true ;
						
					}) ;
					
					externals.doEvt(`${req.url.host}/allowedrequest`,req,resp).then(d=>{
						
						if (d) {
							
							cont = false ;
							
						}
						
						if (gotOtherPromise && cont) {
							
							allowedRequest(req.url.host,req,resp,user_ip,user_ip_remote,timeRecieved,false) ;
							
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
				
				//nothing
				
			}
			
			//Somthing has gone wrong, so play it safe & send a 401.
			else {
				
				console.log(`${req.jpid}\t401 Unauthorized.   Account system ${checkingSystem} denide access.`) ;
				resp.writeHead(401,{"Content-Type":"text/plain"}) ;
				resp.write("Nope.") ;
				resp.end() ;
				
			}
			
		}).catch(err=>{coughtError(err,resp,req.jpid);console.log("Error trace: Error recieving account data.");}) ;
		
	} ;
	//Check
	nextCheck() ;
	
	
	
}

//Should be called when a request is allowed.
function allowedRequest(host,req,resp,user_ip,user_ip_remote,timeRecieved,postDone) {
	try {
		//Determine how to handle the method
		if (typeof implementedMethods[req.method] !== "undefined") {
			if (!postDone) {
				for (let doing in implementedMethods[req.method]) {
					if (implementedMethods[req.method][doing][0](req, resp)) {
						//If it is a custom method and it hasn't already run call it.
						let rv =  implementedMethods[req.method][doing][1](req, resp) ;
						//Promise? If it resolved false, then rerun
						if (typeof rv.then === "function") {
							rv.then(handled=>{
								if (!handled) {
									allowedRequest(host, req, resp, user_ip, user_ip_remote, timeRecieved, true) ;
								}
							}) ;
							return ;
						}
						//Dont run if we returned true
						else if (rv) {
							return ;
						}
					}
				}
			}
		} else if (req.method === "GET") {
			//Do nothing, jsut dont check the rest lol
		} else if (req.method === "HEAD") {
			//Carry on, but don't send the body of the request.
			resp.sendBody = false ;
		} else if (req.method === "POST") {
			//Only if we havn't already got the data
			if (!postDone) {
				//Collect the data (optimise if content-length is set)
				if (typeof req.headers["content-length"] !== "undefined") {
					let dLength = parseInt(req.headers["content-length"], 10) ;
					if (isNaN(dLength)) {
						sendError(400, "Content-Length header must be a number") ;
						return ;
					}
					let data = Buffer.alloc(dLength) ;
					let currentPos = 0 ;
					let errorSent = false ;
					req.on("data", d=>{
						if (currentPos + d.length > data.length) {
							if (errorSent) {
								return ;
							}
							errorSent = true ;
							sendError(400, "Request body was longer than the Content-Length header.") ;
							return ;
						}
						currentPos += d.copy(data, currentPos) ;
					}) ;
					req.on("end", ()=>{
						//Encode data in base64 and add it to resp.vars
						resp.vars.body = data.toString("base64") ;
						allowedRequest(host,req,resp,user_ip,user_ip_remote,timeRecieved,true) ;
					}) ;
					return ;
				} 
				let data = Buffer.alloc(0) ;
				req.on("data", d=>{
					data = Buffer.concat([data, d], data.length + d.length) ;
				}) ;
				req.on("end", ()=>{
					//Encode data in base64 and add it to resp.vars
					resp.vars.body = data.toString("base64") ;
					allowedRequest(host,req,resp,user_ip,user_ip_remote,timeRecieved,true) ;
				}) ;
				return ;
			}
		} else if (req.method === "OPTIONS") {
			//What custom methods support this URL?
			let suppMethods = [] ;
			for (let doing in implementedMethods) {
				for (let theOne in implementedMethods[doing]) {
					//If this handler supports it, push this method to the array and move on to the next method
					if (implementedMethods[doing][theOne][0](req)) {
						suppMethods.push(doing) ;
						break ;
					}
				}
			}
			//Send empty response with allow header as the sorted protocols
			resp.writeHead(200,{
				"Allow": defaultMethods.concat(suppMethods).sort().join(", "),
				"Content-Length": 0
			}) ;
			resp.end() ;
			return ;
		} else {
			//We cannot handle that protocol, so set the allow header (as in the OPTIONS method)
			let suppMethods = [] ;
			for (let doing in implementedMethods) {
				for (let theOne in implementedMethods[doing]) {
					if (implementedMethods[doing][theOne][0](req)) {
						suppMethods.push(doing) ;
						break ;
					}
				}
			}
			resp.setHeader("Allow", defaultMethods.concat(suppMethods).sort().join(", ")) ;
			//And send a 405
			sendError(405, "That method is not supported for this URL. Sorry :(") ;
			return ;
		}
		
		//Use responseMaker to generate the response, see do-response.js
		responseMaker.createResponse(req, resp, timeRecieved).then(hmmm=>{
			//Log if it was leared from
			if (hmmm[0]) {
				console.log(`${req.jpid}\tResponse was based on a previous response.`) ;
			} else {
				console.log(`${req.jpid}\tThe response has been learned from to improve handle time next time round.`) ;
			}
			//Log times
			let timeTaken = process.hrtime(timeRecieved) ;
			console.log(`${req.jpid}\tRequest took ${timeTaken[0] * 1000 + timeTaken[1] * 10e-6}ms to handle.`) ;
		}) ;
		return ;
	} catch (err) {
		coughtError(err,resp,req.jpid) ;
		console.log("Error trace: Request allowed, issue processubg headers.") ;
	}
}

responseMaker.sendCache = sendCache ;
responseMaker.sendFile = sendFile ;
responseMaker.sendError = sendError ;
responseMaker.enableLearning = config.enableLearning ;

for (let doing in config.cache) {
	try {
		responseMaker.cacheFileSync(config.cache[doing]) ;
	} catch (err) {
		console.info(`Failed to cache '${config.cache[doing]}'.`) ;
		console.warn(`Failed to cache '${config.cache[doing]}':`) ;
		console.warn(err) ;
	}
}

module.exports = {
	//Function to init the server.
	init:(clusterGiven) => {
		externals.generateServerObject = () => {
			return {
				
				//Server configuration
				"config": config,
				"reloadConfig": ()=>loadConfig(),
				
				//Recieving requests
				"getData": req=>jpsUtil.getData(req),
				"multipartFormDataParser": require("./multipart-form-data-parser.js"),
				
				//Sending responses
				"sendFile": (file, resp, req) => sendFile(file, resp, resp.vars, req),
				"sendCache": (file, cache, resp, req, status=200) => sendCache(file, cache, resp, resp.vars, req, status),
				"sendError":(code, message, resp, req={jpid:""})=>sendError(code, message, resp, req.jpid),
				"vars": vars,
				
				//Account handling
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
				
				//HTTP Stuff
				//Methods
				"implementMethod": (method, checker, handler) => {
					
					method = method.toUpperCase() ;
					
					//Check types
					if (typeof checker !== "function" || typeof handler !== "function" || typeof method !== "string") {
						
						throw new Error("Both the checker and the handler has to be functions, and the method must be a string.") ;
						
					}
					
					//If there are no current implementations of this method, create it as an empty array.
					if (typeof implementedMethods[method] !== "object") {
						
						implementedMethods[method] = new Array() ;
						
					}
					
					//Push the implementation
					implementedMethods[method].push([checker, handler]) ;
					
				},
				//CORS
				"addCORSRule": (...args) => CORS.addRule(...args),
				
				//Caching
				"cache": {
					"newCache": (url, cache, incSearch=false) => responseMaker.addCache(url, cache, incSearch),
					"cacheFile": url => responseMaker.cacheFile(url),
					"cacheFileAs": (url, file) => responseMaker.cacheFileAs(url, file),
					"isCache": (url, incSearch=false) => responseMaker.isCache(url, incSearch),
					"getCache": (url, incSearch=false) => responseMaker.getCache(url, incSearch),
					"removeCache": (url, incSearch=false) => responseMaker.removeCache(url, incSearch),
					"createLink": (from, to, incSearch=false) => responseMaker.createLink(from, to, incSearch),
					"isLink": (from, incSearch=false) => responseMaker.isLink(from, incSearch),
					"getLink": (from, incSearch=false) => responseMaker.getLink(from, incSearch),
					"removeLink": (from, incSearch=false) => responseMaker.removeLink(from, incSearch)
				},
				
				//Page handling
				"handlePage": (url, handler, incSearch=false) => responseMaker.handlePage(url, handler, incSearch),
				"isHandled": (url, incSearch=false) => responseMaker.isHandled(url, incSearch),
				"removePageHandler": (url, incSearch=false) => responseMaker.removePageHandler(url, incSearch),
				
				//Linking
				"isLearned": (url, checkLevel=0) => responseMaker.isLearned(url, checkLevel),
				"unlearn": (url, level=0) => responseMaker.unlearn(url, level),
				
				//Other stuff
				"getMimeType": (...args)=>getMimeType(...args),
				"createURL": opts=>{
					urlObject.defaultHost = config.defaultHost ;
					urlObject.createURL(opts) ;
				},
				"createURLFromString": url=>{
					urlObject.defaultHost = config.defaultHost ;
					urlObject.createURLFromString(url) ;
				}
			} ;
		} ;
		let currentLimitedAccountSystem = 0 ;
		externals.generateLimitedServerObject = (domains, fs) => {
			const checkFile = file => {
				const fp = path.join(process.cwd(), "sites", file) ;
				let ok = false ;
				for (let doing in domains) {
					if (fp.indexOf(path.join(process.cwd(), "sites", domains[doing])) === 0) {
						ok = true ;
						break ;
					}
				}
				return ok ;
			} ;
			const checkURLString = url => {
				let host = url.split("/").shift() ;
				return domains.indexOf(host) !== -1 ;
			} ;
			return {
				
				//Recieving requests
				"getData": req=>jpsUtil.getData(req),
				"multipartFormDataParser": require("./multipart-form-data-parser.js"),
				
				//Sending responses
				"sendFile": (file, resp, req) => {
					if (checkFile(file)) {
						return sendFile(file, resp, resp.vars, req) ;
					}
					throw new Error(`This limited extension doesn't have power over ${file}.`) ;
				},
				"sendCache": (file, cache, resp, req, status=200) => {
					if (checkFile(file)) {
						return sendCache(file, cache, resp, resp.vars, req, status) ;
					}
					throw new Error(`This limited extension doesn't have power over ${file}.`) ;
				},
				"sendError":(...eArgs)=>sendError(...eArgs),
				
				//Account handling
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
				
				//HTTP Stuff
				//CORS
				"addCORSRule": (...args) => {
					if (domains.indexOf(args[1]) === -1) {
						throw new Error(`This limited extension cannot set up CORS rules for the host: ${args[1]}.`) ;
					} else {
						CORS.addRule(...args) ;
					}
				},
				
				//Caching
				"cache": {
					"newCache": (url, cache, incSearch=false) => {
						if (checkURLString(url)) {
							return responseMaker.addCache(url, cache, incSearch) ;
						} 
						throw new Error(`Sorry, you cannot create a cache for the host '${url.split("/").shift()}'.`) ;
						
					},
					"cacheFile": (url) => {
						if (checkURLString(url)) {
							return responseMaker.cacheFile(url) ;
						} 
						throw new Error(`Sorry, you cannot create a cache for the host '${url.split("/").shift()}'.`) ;
						
					},
					"cacheFileAs": (url, file) => {
						if (checkURLString(url) && checkURLString(file)) {
							return responseMaker.cacheFileAs(url, file) ;
						} 
						throw new Error(`Sorry, you cannot create a cache for the host '${url.split("/").shift()}' or '${file.split("/").shift()}'.`) ;
						
					},
					"isCache": (url, incSearch=false) => {
						if (checkURLString(url)) {
							return responseMaker.isCache(url, incSearch) ;
						} 
						throw new Error(`Sorry, you cannot view a cache for the host '${url.split("/").shift()}'.`) ;
						
					},
					"getCache": (url, incSearch=false) => {
						if (checkURLString(url)) {
							return responseMaker.getCache(url, incSearch) ;
						} 
						throw new Error(`Sorry, you cannot view a cache for the host '${url.split("/").shift()}'.`) ;
						
					},
					"removeCache": (url, incSearch=false) => {
						if (checkURLString(url)) {
							return responseMaker.removeCache(url, incSearch) ;
						} 
						throw new Error(`Sorry, you cannot remove a cache for the host '${url.split("/").shift()}'.`) ;
						
					},
					"createLink": (from, to, incSearch=false) => {
						if (checkURLString(from)) {
							return responseMaker.createLink(from, to, incSearch) ;
						} 
						throw new Error(`Sorry, you cannot create a link for the host '${from.split("/").shift()}'.`) ;
						
					},
					"isLink": (from, incSearch=false) => {
						if (checkURLString(from)) {
							return responseMaker.isLink(from, incSearch) ;
						} 
						throw new Error(`Sorry, you cannot view a link for the host '${from.split("/").shift()}'.`) ;
						
					},
					"getLink": (from, incSearch=false) => {
						if (checkURLString(from)) {
							return responseMaker.getLink(from, incSearch) ;
						} 
						throw new Error(`Sorry, you cannot view a link for the host '${from.split("/").shift()}'.`) ;
						
					},
					"removeLink": (from, incSearch=false) => {
						if (checkURLString(from)) {
							return responseMaker.removeLink(from, incSearch) ;
						} 
						throw new Error(`Sorry, you cannot remove a link for the host '${from.split("/").shift()}'.`) ;
						
					}
				},
				
				//Page handling
				"handlePage": (url, handler, incSearch=false) => {
					if (checkURLString(url)) {
						return responseMaker.handlePage(url, handler, incSearch) ;
					} 
					throw new Error(`Sorry, you cannot handle a page for the host '${url.split("/").shift()}'.`) ;
					
				},
				"isHandled": (url, incSearch=false) => {
					if (checkURLString(url)) {
						return responseMaker.isHandled(url, incSearch) ;
					} 
					throw new Error(`Sorry, you cannot view a handler for the host '${url.split("/").shift()}'.`) ;
					
				},
				"removePageHandler": (url, incSearch=false) => {
					if (checkURLString(url)) {
						return responseMaker.removePageHandler(url, incSearch) ;
					} 
					throw new Error(`Sorry, you cannot remove a handler for the host '${url.split("/").shift()}'.`) ;
					
				},
				
				//Linking
				"isLearned": (url, checkLevel=0) => {
					if (checkURLString(url)) {
						return responseMaker.isLearned(url, checkLevel) ;
					} 
					throw new Error(`Sorry, you cannot view learning details for the host '${url.split("/").shift()}'.`) ;
					
				},
				"unlearn": (url, level=0) => {
					if (checkURLString(url)) {
						return responseMaker.isLearned(url, level) ;
					} 
					throw new Error(`Sorry, you cannot change learning details for the host '${url.split("/").shift()}'.`) ;
					
				},
				
				//Other stuff
				"getMimeType":(...args)=>getMimeType(...args),
				"createURL": opts=>{
					urlObject.defaultHost = config.defaultHost ;
					urlObject.createURL(opts) ;
				},
				"createURLFromString": url=>{
					urlObject.defaultHost = config.defaultHost ;
					urlObject.createURLFromString(url) ;
				}
				
			} ;
		} ;
		
		//Load the extentions
		let currentDir = fs.readdirSync(process.cwd()) ;
		//Go through all the files in the cwd.
		for (let doing in currentDir) {
			//If it is an extention, load it.
			if (currentDir[doing].substr(currentDir[doing].length - 7,7) === ".jpe.js") {
				let currentLoad = externals.loadExt(currentDir[doing]) ;
				if (currentLoad.loaded) {
					vars = currentLoad.serverObj.vars ;
				}
			}
		}
		for (let doing in config.CORS) {
			if (typeof config.CORS[doing] !== "object" || config.CORS[doing].constructor !== Array) {
				console.warn(`Not adding CORS rule ${doing} as it is not an Array`) ;
				continue ;
			}
			let thisOb = new Object() ;
			Object.assign(thisOb, config.CORS[doing]) ;
			for (let tRE in thisOb[5]) {
				if (thisOb[5][tRE].indexOf("*") === 0) {
					thisOb[5][tRE] = new RegExp(thisOb[5][tRE].substring(1, thisOb[5][tRE].length), "g") ;
				}
			}
			CORS.addRule(...thisOb) ;
		}
		//Set up the HTTP servers
		for (const doing in config.httpServers) {
			let options = new Array() ;
			options[0] = config.httpServers[doing].port ;
			if (typeof config.httpServers[doing].host !== "undefined") {
				options[1] = config.httpServers[doing].host ;
			}
			http.createServer((req,resp) => {
				req.port = options[0] ;
				handleRequest(req,resp,false) ;
			}).listen(...options) ;
		}
		//Set up the HTTPS servers
		for (const doing in config.httpsServers) {
			https.createServer({key:fs.readFileSync("privkey.pem"),ca:fs.readFileSync("fullchain.pem"),cert:fs.readFileSync("cert.pem")}, (req,resp) => {
				req.port = config.httpsServers[doing].port ;
				handleRequest(req,resp,true) ;
			}).listen(config.httpsServers[doing].port) ;
		}
		//Get the cluster module of parent.
		cluster = clusterGiven ;
		vars.Global["thread_id"] = cluster.worker.id ;
		//Ready event
		externals.doEvt("ready") ;
	}
} ;
