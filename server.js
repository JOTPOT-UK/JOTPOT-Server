/*
	
	JOTPOT Server 3.
	Copyright (c) Jacob O'Toole 2016-2017
	
*/

//Console is now YAY!!!
console.log = console.warn = (...args) => {
	
	process.send(["log",args.join(" ")]) ;
	
}/**/

//Node modules
let http = require("http") ;
let https = require("https") ;
let fs = require("fs") ;
let path = require("path") ;
let zlib = require("zlib") ;
let proc = require("./accounts.js") ;
let externals = require("./externals") ;
let {Transform,Readable} = require("stream") ;
let cluster ;

let config ;
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

else if (fs.existsSync("config.json")) {
	
	console.warn("Config file does not exist.") ;
	console.info("Config file does not exist.") ;
	console.warn("Exiting") ;
	console.info("Exiting") ;
	process.exit() ;
	
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
	doVarsFor[doing] = path.join("sites",doVarsFor[doing]) ;
	
}
let dontDoVarsFor = [] ;

//Get buffer of string for inserting vars.
let startOfVar = new Buffer("$::") ;
let endOfVar = new Buffer("::$") ;

//Error file
let errorFile = fs.readFileSync(config.errorTemplate).toString() ;
let errorCodes = new Object() ;
errorCodes[403] = "Sorry, however you are not permitted to access this file." ;
errorCodes[404] = "The page you are looking for may have been removed or moved to a new location!" ;

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
			
			/*if (doVarsDefault || doVarsFor.indexOf(this.path) !== -1 || this.ignoreList)*/ do {
				
				if (data.indexOf(startOfVar) !== -1) {
					
					if (data.indexOf(endOfVar) !== -1) {
						
						let dataString = this.lastData + data.toString() ;
						
						let varsKeys = Object.keys(vars) ;
						for (let doingScope in varsKeys) {
							
							let innerVars = Object.keys(vars[varsKeys[doingScope]]) ;
							for (let doing in innerVars) {
								
								let toReplace = `\\$\\:\\:\\:${innerVars[doing]}\\:\\:\\:\\$` ;
								let replaceWith = String(vars[varsKeys[doingScope]][innerVars[doing]]) ;
								dataString = dataString.replace(new RegExp(toReplace,"g"),replaceWith) ;
								
							}
							
						}
						
						varsKeys = Object.keys(this.extraVars) ;
						for (let doing in varsKeys) {
							
							let toReplace = `\\$\\:\\:\\:${varsKeys[doing]}\\:\\:\\:\\$` ;
							let replaceWith = String(this.extraVars[varsKeys[doing]]) ;
							//console.log(`Replacing ${toReplace} with ${replaceWith}.`) ;
							dataString = dataString.replace(new RegExp(toReplace,"g"),replaceWith) ;
							
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
		
		if (ext === "") {
			
			return "text/html" ;
			
		}
		
		/*else if (ext === ".gz" || ext === ".zz") {
			
			ext = path.extname(file.substring(0,file.length - 3)) ;
			
		}*/
		
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

//Proxy to realms
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

//Pipes the file through the transform pipe into the main pipe.
//Calls the callback with the first argument as a boolean - true if succeded, false if not.
function getFile(file,callWithStats,transformPipe,pipeTo,callback) {
	
	fs.stat(file,(err,stats) => {
		
		if (err) {
			
			callback(false,err) ;
			return false ;
			
		}
		
		if (stats.isFile()) {
			
			if (callWithStats(stats)) {
				
				console.log(`\tRequest took ${Date.now() - requestGotAt}ms to process.`) ;
				
				if (config.addVarsByDefault || doVarsFor.indexOf(file) !== -1) {
					
					fs.createReadStream(file,{
						
						flags: 'r',
						autoClose: true
						
					}).pipe(transformPipe).pipe(pipeTo) ;
					
				}
				
				else {
					
					fs.createReadStream(file,{
						
						flags: 'r',
						autoClose: true
						
					}).pipe(pipeTo) ;
					
				}
				
				callback(true,"Erm") ;
				
			}
			
		}
		
		else {
			
			callback(false,"TEST") ;
			
		}
		
	}) ;
	
}

//Sends the file specified to the pipe as the second argument - goes through the getFile & thus vars pipe.
function sendFile(file,resp,customVars) {
	
	try {
		
		file = path.join("./sites/",file) ;
		
		let transformPipe = "none" ;
		if (config.addVarsByDefault || doVarsFor.indexOf(file) !== -1) {
			
			transformPipe = new addVars(file,customVars) ;
			
		}
		
		return new Promise((resolve,reject) => {getFile(file,(stats) => {
			
			let mime = getMimeType(file) ;
			console.log(`\t200 OK.   ${file} (${mime}) loaded from disk.`) ;
			resp.writeHead(200,{
				
				"Content-Type": mime,
				"Server": "JOTPOT Server 3",
				
				//Removed because it is now sent via chunked encoding:
				/*"Content-Length": stats.size,
				"Accept-Ranges": "bytes",*/
				
				
				//Added because google does it :)
				"status": 200
				
			}) ;
			
			return true ;
			
		},transformPipe,resp,(done,err) => {
			
			resolve([done,err]) ;
			
		});}) ;
		
	}
	
	catch(err) {
		
		coughtError(err,resp) ;
		
	}
	
}

function sendCache(file,cache,resp,customVars,status=200) {
	
	try {
		
		file = path.join("./sites/",file) ;
		
		let transformPipe = "none" ;
		if (config.addVarsByDefault || doVarsFor.indexOf(file) !== -1) {
			
			transformPipe = new addVars(file,customVars) ;
			
		}
		
		let mime = getMimeType(file) ;
		console.log(`\t200 OK.   ${file} (${mime}) loaded from cache.`) ;
		resp.writeHead(status,{
			
			"Content-Type": mime,
			"Server": "JOTPOT Server 3",
			
			//Removed because it is now sent via chunked encoding:
			/*"Content-Length": stats.size,
			"Accept-Ranges": "bytes",*/
			
			
			//Added because google does it :)
			"status": status
			
		}) ;
		
		console.log(`\tRequest took ${Date.now() - requestGotAt}ms to process.`) ;
		
		if (config.addVarsByDefault || doVarsFor.indexOf(file) !== -1) {
			
			transformPipe.pipe(resp) ;
			
			transformPipe.write(cache) ;
			transformPipe.end() ;
			
		}
		
		else {
			
			resp.write(cache) ;
			resp.end() ;
			
		}
		
	}
	
	catch(err) {
		
		coughtError(err,resp) ;
		
	}
	
}

function sendError(code,message,resp) {
	
	sendCache("error_page",errorFile,resp,{error_code:code,error_type:http.STATUS_CODES[code],error_message:message},code) ;
	return ;
	
}

function coughtError(err,resp) {
	
	console.warn(`Error occured in main request handler: ${err}`) ;
	sendError(500,"An unknowen error occured.",resp) ;
	
}

let pages = {
	
	/*"/time?p" : ["file","/time_utc_plain"] ,
	"/time?now" : ["file","/time_now_plain"] ,*/
	"www.jotpot.co.uk/time?p": ["cache","$:::utctime:::$"],
	"www.jotpot.co.uk/time?now": ["cache","$:::time:::$"],
	"www.jotpot.co.uk/ip?p": ["cache","$:::user_ip:::$"],
	"www.jotpot.co.uk/": ["file","/index.html"],
	"www.jotpot.co.uk/new-troubleshooter-case": ["func",function (req,resp) {
		
		if (req.method.toLowerCase() !== "post") {
			
			sendError(405,"You need to make a post reqest to create a new case.",resp) ;
			
		}
		
		else {
			
			resp.writeHead(200,{"Content-Type":"text/plain"}) ;
			req.on("data",(pData) => {
				
				var caseNumber ;
				fs.readFile("./currentcase.dat",(err,fData) => {
					
					if (err) {
						
						sendError(500,"No idea, do you really think I can be bothered to make a debug for this?<br>If you really care, this is error 1:*** where I cant be bothered to give you the star bits because you probably dont need them.",resp) ;
						return false ;
						
					}
					
					caseNumber = parseInt(fData.toString()) + 1 ;
					
					fs.writeFile("./currentcase.dat",caseNumber.toString(),(err) => {
						
						if (err) {
							
							sendError(500,"No idea, do you really think I can be bothered to make a debug for this?<br>If you really care, this is error 2:*** where I cant be bothered to give you the star bits because you probably dont need them.",resp) ;
							return false ;
							
						}
						
						fs.writeFile("./currentcase.dat",caseNumber.toString(),(err) => {
							
							if (err) {
								
								sendError(500,"No idea, do you really think I can be bothered to make a debug for this?<br>If you really care, this is error 3:*** where I cant be bothered to give you the star bits because you probably dont need them.",resp) ;
								return false ;
								
							}
							
							fs.appendFile("./sites/www.jotpot.co.uk/tCases.txt",`\r\nCase ${caseNumber}:\r\n${pData.toString()}\r\n\r\n`,(err) => {
								
								if (err) {
									
									sendError(500,"No idea, do you really think I can be bothered to make a debug for this?<br>If you really care, this is error 4:*** where I cant be bothered to give you the star bits because you probably dont need them.",resp) ;
									return false ;
									
								}
								
								resp.writeHead(200,{"Content-Type":"text/plain","Case":String(caseNumber)}) ;
								resp.end("Your case has been saved succesfully.") ;
								return true ;
								
							}) ;
							
						}) ;
						
					}) ;
					
				}) ;
				
			}) ;
			return true ;
			
		}
		return true ;
		
	}]
	
} ;

//Function to handle http requests.
function handleRequest(req,resp) {
	
	try {
		
		let timeRecieved = Date.now() ;
		requestGotAt = timeRecieved ;
		let host = req.headers.host || config.defaultDomain ;
		let user_ip = req.headers['x-forwarded-for'] || req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress ;
		let user_ip_remote = req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress ;
		let requestTime = new Date(timeRecieved) ;
		resp.vars = {"user_ip":user_ip,"user_ip_remote":user_ip_remote,"utctime":requestTime.toUTCString(),"time":requestTime.getTime(),"host":host} ;
		if (externals.handles.request(req,resp)) {
			
			return true ;
			
		}
		
		req.orig_url = req.url ;
		req.url = req.url.toLowerCase().replace(/\.\./g,"") ;
		
		for (let doing in config.otherProcesses) {
			
			if (config.otherProcesses[doing].forwardUrls.indexOf(host + req.url) !== -1) {
				
				return forwardToOtherServer(req,resp,config.otherProcesses[doing].forwardPort) ;
				
			}
			
		}
		
		if (req.overHttps === false && config.redirectToHttps.indexOf(host) !== -1 && config.canBeHttp.indexOf(req.url) === -1) {
			
			console.log(`\n\nRequest from ${user_ip_remote}(${user_ip}) for ${host}${req.url} being handled by thread ${cluster.worker.id}.`) ;
			console.log(`\t302 Found.   Redirecting to https://${host}${req.url}.`) ;
			resp.writeHead(302,{"Content-Type":"text/plain","location":"https://" + host + req.url}) ;
			resp.write("Redirecting you to our secure site...") ;
			resp.end() ;
			return ;
			
		}
		
		if (typeof config.hostAlias[host] !== "undefined") {
			
			host = config.hostAlias[host] ;
			
		}
		
		if (typeof config.hostRedirects[host] !== "undefined") {
			
			console.log(`\n\nRequest from ${user_ip_remote}(${user_ip}) for ${host}${req.url} being handled by thread ${cluster.worker.id}.`) ;
			console.log(`\t302 Found.   Redirecting to ${config.hostRedirects[host]}.`) ;
			resp.writeHead(302,{"Content-Type":"text/plain","location":["http://","https://"][Number(req.overHttps)] + hostRedirects[host]}) ;
			resp.write("Redirecting you to " + ["http://","https://"][Number(req.overHttps)] + hostRedirects[host] + "...") ;
			resp.end() ;
			return ;
			
		}
		
		req.host = host ;
		
		if (externals.handles.fullrequest(req,resp)) {
			
			return true ;
			
		}
		
		req.url = host + req.url ;
		
		console.log(`\n\nRequest from ${user_ip_remote}(${user_ip}) for ${req.url} being handled by thread ${cluster.worker.id}.`) ;
		
		if (allAccountSystems.length === 0) {
			
			allowedRequest(host,req,resp,user_ip,user_ip_remote,timeRecieved) ;
			return true ;
			
		}
		
		let checkingSystem = 0 ;
		
		let nextCheck =_=>{
			
			allAccountSystems[checkingSystem].doAnything(req,resp).then(returned=>{
				
				let canAccess = returned[0] ;
				
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
				
				else if (canAccess === true) {
					
					checkingSystem++ ;
					if (checkingSystem >= allAccountSystems.length) {
						
						allowedRequest(host,req,resp,user_ip,user_ip_remote,timeRecieved) ;
						
					}
					
					else {
						
						nextCheck() ;
						
					}
					
				}
				
				else if (canAccess === null) {
					
					return ;
					
				}
				
				else {
					
					console.log(`\t401 Unauthorized.   Account system ${checkingSystem} denide access.`) ;
					resp.writeHead(401,{"Content-Type":"text/plain"}) ;
					resp.write("Nope.") ;
					resp.end() ;
					return false ;
					
				}
				
			}).catch(err=>{coughtError(err,resp);console.log("Error trace: Error recieving account data.");}) ;
			
		} ;
		nextCheck() ;
		
		return ;
		
	}
	
	catch(err) {
		
		coughtError(err,resp) ;
		console.log("Error trace: Error handling incoming data.") ;
		
	}
	
}

function allowedRequest(host,req,resp,user_ip,user_ip_remote) {
	
	try {
		
		//requestTime = new Date(requestTime) ;
		//let varsToSend = {"user_ip":user_ip,"user_ip_remote":user_ip_remote,"utctime":requestTime.toUTCString(),"time":requestTime.getTime(),"host":host} ;
		let varsToSend = resp.vars ;
		
		if (typeof pages[req.url] === "object") {
			
			if (pages[req.url][0] === "file") {
				
				req.url = host + pages[req.url][1] ;
				
			}
			
			if (pages[req.url][0] === "cache") {
				
				sendCache(req.url,pages[req.url][1],resp,varsToSend) ;
				return true ;
				
			}
			
			else if (pages[req.url][0] === "func") {
				
				if (!req.url,pages[req.url][1](req,resp)) {
					
					return true ;
					
				}
				
			}
			
		}
		
		sendFile(req.url,resp,varsToSend).then(done=>{
			
			if (!done[0]) {
				
				return sendFile(`${req.url}.page`,resp,varsToSend);
				
			}
			
			else {
				
				return [true] ;
				
			}
			
		}).then(done=>{
			
			if (!done[0]) {
				
				return sendFile(path.join(req.url,"/index.html"),resp,varsToSend);
				
			}
			
			else {
				
				return [true] ;
				
			}
			
		}).then(done=>{
			
			if (!done[0]) {
				
				let code = 500 ;
				if (done[1].code === "EACCES") {
					
					code = 403 ;
					
				}
				else if (done[1].code === "ENOENT") {
					
					code = 404 ;
					
				}
				return sendError(code,errorCodes[code],resp) ;
				
			}
			
			else {
				
				return [true] ;
				
			}
			
		}).catch(err=>{coughtError(err,resp);console.log("Error trace: Error in send promise.") ;}) ;
		
	}
	
	catch(err) {
		
		coughtError(err,resp) ;
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
	
	init:(clusterGiven) => {
		
		let currentDir = fs.readdirSync(process.cwd()) ;
		for (let doing in currentDir) {
			
			if (currentDir[doing].substr(currentDir[doing].length - 7,7) === ".jpe.js") {
				
				externals.loadExt(currentDir[doing],{
					
					"pages": pages,
					"vars": vars
					
				}) ;
				
			}
			
		}
		
		if (typeof externals.handles.request === "undefined") {
			
			externals.handles.request =_=>false ;
			
		}
		
		if (typeof externals.handles.fullrequest === "undefined") {
			
			externals.handles.fullrequest =_=>false ;
			
		}
		
		//Set up the HTTP servers
		for (let doing in config.httpServers) {
			
			http.createServer((req,resp) => {
				
				req.overHttps = false ;
				handleRequest(req,resp) ;
				
			}).listen(config.httpServers[doing].port) ;
			
		}
		
		//Set up the HTTPS servers
		for (let doing in config.httpsServers) {
			
			https.createServer({pfx:fs.readFileSync("server.pfx"),passphrase:"JpotjpotKEY1!",key:fs.readFileSync("key.key"),ca:fs.readFileSync("inter.cer")},(req,resp) => {
				
				req.overHttps = true ;
				handleRequest(req,resp) ;
				
			}).listen(config.httpsServers[doing].port) ;
			
		}
		
		//Get the cluster module of parent.
		cluster = clusterGiven ;
		vars.Global["thread_id"] = cluster.worker.id ;

	}

}

let requestGotAt ;