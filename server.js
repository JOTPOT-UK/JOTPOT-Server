/*
	
	JOTPOT Server.
	Copyright (c) Jacob O'Toole 2016-2017
	
*/

//Version 1.0.1

//Console is now YAY!!!
console.log = console.warn = (...args) => {
	
	process.send(["log",args.join(" ")]) ;
	
}

//Modules
let http = require("http") ;
let https = require("https") ;
let fs = require("fs") ;
let path = require("path") ;
let zlib = require("zlib") ;
let proc = require("./accounts") ;
let externals = require("./externals") ;
let {Transform,Readable} = require("stream") ;
let cluster ;

//Load the config
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

else {
	
	console.warn("Config file does not exist.") ;
	console.info("Config file does not exist.") ;
	console.warn("Exiting") ;
	console.info("Exiting") ;
	process.exit() ;
	
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
	doVarsFor[doing] = path.join("sites",doVarsFor[doing]) ;
	
}
let dontDoVarsFor = [] ;

//Get buffer of string for inserting vars.
let startOfVar = new Buffer("$::") ;
let endOfVar = new Buffer("::$") ;

//Error file
if (!fs.existsSync(config.errorTemplate)) {
	
	console.warn("Error template file does not exist.") ;
	console.info("Error template file does not exist.") ;
	console.warn("Exiting") ;
	console.info("Exiting") ;
	process.exit() ;
	
}
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
			
			do {
				
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
function getFile(file,callWithStats,pipeTo,callback) {
	
	fs.stat(file,(err,stats) => {
		
		if (err) {
			
			callback(false,err) ;
			return false ;
			
		}
		
		if (stats.isFile()) {
			
			if (callWithStats(stats)) {
				
				console.log(`\tRequest took ${Date.now() - requestGotAt}ms to process.`) ;
				
				//Pipe file to the pipe.
				fs.createReadStream(file,{
					
					flags: 'r',
					autoClose: true
					
				}).pipe(pipeTo) ;
				
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
		
		//Look in the sites dir.
		file = path.join("./sites/",file) ;
		
		
		//Make a pipe to send it to.
		let mainPipe = "none" ;
		let doingTransform = resp.pipeThrough.length - 1 ;
		
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
			
		}
		
		return new Promise((resolve,reject) => {getFile(file,(stats) => {
			
			let mime = getMimeType(file) ;
			console.log(`\t200 OK.   ${file} (${mime}) loaded from disk.`) ;
			resp.writeHead(200,{
				
				"Content-Type": mime,
				"Server": "JOTPOT Server",
				
				//Added because google does it :)
				"status": 200
				
			}) ;
			
			return true ;
			
		},mainPipe,(done,err) => {
			
			resolve([done,err]) ;
			
		});}) ;
		
	}
	
	catch(err) {
		
		coughtError(err,resp) ;
		
	}
	
}

function sendCache(file,cache,resp,customVars,status=200) {
	
	try {
		
		//Look in the sites dir.
		file = path.join("./sites/",file) ;
		
		
		//Make a pipe to send it to.
		let mainPipe = "none" ;
		let doingTransform = resp.pipeThrough.length - 1 ;
		
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
			
		}
		
		//Get the mime type.
		let mime = getMimeType(file) ;
		console.log(`\t200 OK.   ${file} (${mime}) loaded from cache.`) ;
		resp.writeHead(status,{
			
			"Content-Type": mime,
			"Server": "JOTPOT Server",
			
			//Added because google does it :)
			"status": status
			
		}) ;
		
		console.log(`\tRequest took ${Date.now() - requestGotAt}ms to process.`) ;
		
		//Write the cached data & end.
		mainPipe.write(cache) ;
		mainPipe.end() ;
		
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

//Function to handle http requests.
function handleRequest(req,resp) {
	
	try {
		
		//Get time stuff.
		let timeRecieved = Date.now() ;
		requestGotAt = timeRecieved ;
		let requestTime = new Date(timeRecieved) ;
		
		//For vars
		let host = req.headers.host || config.defaultDomain ;
		req.host = host ;
		let user_ip = req.headers['x-forwarded-for'] || req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress ;
		let user_ip_remote = req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress ;
		
		//Add stuff to resp object.
		resp.vars = {"user_ip":user_ip,"user_ip_remote":user_ip_remote,"utctime":requestTime.toUTCString(),"time":requestTime.getTime(),"host":host} ;
		resp.pipeThrough = new Array() ;
		
		//Do request handle.
		//if (externals.handles.request(req,resp)) {
		if (externals.doEvt("request",req,resp)) {
			
			return true ;
			
		}
		
		//Secure URL.
		req.orig_url = req.url ;
		req.url = req.url.toLowerCase().replace(/\.\./g,"") ;
		
		//Check if we need to forward to another port.
		for (let doing in config.otherProcesses) {
			
			if (config.otherProcesses[doing].forwardUrls.indexOf(host + req.url) !== -1) {
				
				return forwardToOtherServer(req,resp,config.otherProcesses[doing].forwardPort) ;
				
			}
			
		}
		
		//Should we redirect to https.
		if (req.overHttps === false && config.redirectToHttps.indexOf(host) !== -1 && config.canBeHttp.indexOf(req.url) === -1) {
			
			console.log(`\n\nRequest from ${user_ip_remote}(${user_ip}) for ${host}${req.url} being handled by thread ${cluster.worker.id}.`) ;
			console.log(`\t302 Found.   Redirecting to https://${host}${req.url}.`) ;
			resp.writeHead(301,{"Content-Type":"text/plain","location":"https://" + host + req.url}) ;
			resp.write("Redirecting you to our secure site...") ;
			resp.end() ;
			return ;
			
		}
		
		//Is the host an alias
		if (typeof config.hostAlias[host] !== "undefined") {
			
			host = config.hostAlias[host] ;
			
		}
		
		//Should we redirect to another host.
		if (typeof config.hostRedirects[host] !== "undefined") {
			
			console.log(`\n\nRequest from ${user_ip_remote}(${user_ip}) for ${host}${req.url} being handled by thread ${cluster.worker.id}.`) ;
			console.log(`\t302 Found.   Redirecting to ${config.hostRedirects[host]}.`) ;
			resp.writeHead(301,{"Content-Type":"text/plain","location":["http://","https://"][Number(req.overHttps)] + config.hostRedirects[host] + req.url}) ;
			resp.write("Redirecting you to " + ["http://","https://"][Number(req.overHttps)] + config.hostRedirects[host] + req.url + "...") ;
			resp.end() ;
			return ;
			
		}
		
		//Add host to URL
		req.accualHost = req.host ;
		req.host = host ;
		req.path = req.url ;
		req.url = host + req.url ;
		
		//Page alias?
		if (typeof config.pageAlias[req.url] !== "undefined") {
			
			req.url = config.pageAlias[req.url] ;
			
		}
		
		//Handle for full request.
		//if (externals.handles.fullrequest(req,resp)) {
		if (externals.doEvt("fullrequest",req,resp)) {
			
			return true ;
			
		}
		
		console.log(`\n\nRequest from ${user_ip_remote}(${user_ip}) for ${req.url} being handled by thread ${cluster.worker.id}.`) ;
		
		//If there are no account systems, then dont bother checking if the user has permission.
		if (allAccountSystems.length === 0) {
			
			allowedRequest(host,req,resp,user_ip,user_ip_remote,timeRecieved) ;
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
						allowedRequest(host,req,resp,user_ip,user_ip_remote,timeRecieved) ;
						
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
					
					console.log(`\t401 Unauthorized.   Account system ${checkingSystem} denide access.`) ;
					resp.writeHead(401,{"Content-Type":"text/plain"}) ;
					resp.write("Nope.") ;
					resp.end() ;
					return false ;
					
				}
				
			}).catch(err=>{coughtError(err,resp);console.log("Error trace: Error recieving account data.");}) ;
			
		} ;
		//Check
		nextCheck() ;
		
		return ;
		
	}
	
	catch(err) {
		
		coughtError(err,resp) ;
		console.log("Error trace: Error handling incoming data.") ;
		
	}
	
}

//Should be called when a request is allowed.
function allowedRequest(host,req,resp,user_ip,user_ip_remote) {
	
	try {
		
		let varsToSend = resp.vars ;
		
		//Is cached or special page.
		if (typeof pages[req.url] === "object") {
			
			//Is just page alias.
			if (pages[req.url][0] === "file") {
				
				req.url = host + pages[req.url][1] ;
				
			}
			
			//Check that the page is still in the cache after a change.
			if (typeof pages[req.url] === "object") {
				
				//If it is a cached file send it.
				if (pages[req.url][0] === "cache") {
					
					sendCache(req.url,pages[req.url][1],resp,varsToSend) ;
					return true ;
					
				}
				
				//If it is a function, then execute the function & dont continute if it returns true.
				else if (pages[req.url][0] === "func") {
					
					if (req.url,pages[req.url][1](req,resp)) {
						
						return true ;
						
					}
					
				}
				
			}
			
		}
		
		//Try the URL the user entered.
		sendFile(req.url,resp,varsToSend).then(done=>{
			
			//If the file failed to send.
			if (!done[0]) {
				
				//Try with .page extention.
				return sendFile(`${req.url}.page`,resp,varsToSend);
				
			}
			
			else {
				
				return [true] ;
				
			}
			
		}).then(done=>{
			
			//If the file failed to send.
			if (!done[0]) {
				
				//Try the index.html.
				return sendFile(path.join(req.url,"/index.html"),resp,varsToSend);
				
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
	
	//Function to init the server.
	init:(clusterGiven) => {
		
		//Load the extentions
		let currentDir = fs.readdirSync(process.cwd()) ;
		//Go through all the files in the cwd.
		for (let doing in currentDir) {
			
			//If it is an extention, load it.
			if (currentDir[doing].substr(currentDir[doing].length - 7,7) === ".jpe.js") {
				
				let currentLoad = externals.loadExt(currentDir[doing],{
					
					"pages": pages,
					"vars": vars,
					"sendError":(...eArgs)=>sendError(...eArgs),
					"createAccountSystem":(args) => {
						
						let creatingAcc = new proc(
							
							args.name,
							args.db,
							args.pages,
							args.exclude,
							args.loginURL,
							args.loginPage,
							args.logoutURL,
							args.logoutPage,
							args.regURL,
							args.regPage
							
						) ;
						allAccountSystems.push(creatingAcc) ;
						
					}
					
				}) ;
				
				if (currentLoad.loaded) {
					
					pages = currentLoad.serverObj.pages ;
					vars = currentLoad.serverObj.vars ;
					
				}
				
			}
			
		}
		
		//If the correct handes are not defined, create them.
		/*if (typeof externals.handles.request === "undefined") {
			
			externals.handles.request =_=>false ;
			
		}
		
		if (typeof externals.handles.fullrequest === "undefined") {
			
			externals.handles.fullrequest =_=>false ;
			
		}*/
		
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