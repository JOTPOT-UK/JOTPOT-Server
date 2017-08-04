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


process.title = "JOTPOT Server" ;

//Make the logs accually go to the logs.
let logs = new Array() ;
console.log = console.warn = (...args) => {
	
	logs.push(args.join(" ")) ;
	
} ;

global.requireJPS = mod => require(path.join(__dirname, mod)) ;

//Node Modules
let fs = require("fs") ;
const path = require("path") ;
let cluster = require("cluster") ;

//Check sites directory exists
if (!fs.existsSync("sites") || !fs.statSync("sites").isDirectory()) {
	console.warn("'sites' directory must exist!") ;
	console.info("'sites' directory must exist!") ;
	throw new Error("'sites' directory must exist!") ;
}

//JPS Modules
let externals = requireJPS("externals") ;
let parseFlags = requireJPS("flag-parser") ;

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
	
	"enableLearning": true
	
} ;

//Load the comfig and fill in any blanks. If it doesn't exist, set the config to the default config.
function loadConfig() {
	
	//If it exists, load it, parse it and fill in any blanks or throw if the types aren't correct
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
if (flags["-data"]) {
	if (flags["-data"].length > 1) {
		console.warn("Only 1 data server listener can be specified.") ;
		throw new Error("Only 1 data server listener can be specified.") ;
	}
	config.dataPort = flags["-data"][0] ;
}
config.dataPort = [config.dataPort] ;
if (typeof config.dataPort[0] === "string") {
	if (config.dataPort[0].split(":").length > 1) {
		let splitHost = config.dataPort[0].split(":") ;
		config.dataPort[0] = splitHost.pop() ;
		config.dataPort[1] = splitHost.join(":") ;
	}
}

//Get stuff ready for user systems
// eslint-disable-next-line
let UIDs = new Array() ;
let authedUsers = new Object() ;

//If this is the master process.
if (cluster.isMaster) {
	
	console.info("Loading JOTPOT Server.") ;
	console.log("Loading JOTPOT Server.") ;
	process.title = "JOTPOT Server 3" ;
	
	let vars = new Object() ;
	let modingVars = new Array() ;
	let waitingForAvail = new Object() ;
	
	//Add the function to the que for v to be available
	const whenAvail = (v,func) => {
		
		//If it isn't in the waiting list, call func imediatally
		// and return false because there is no que
		if (modingVars.indexOf(v) === -1) {
			
			func() ;
			return false ;
			
		}
		
		//If not, add it to the waiting list for this variable
		// and return true because there is a que
		
			
		//But if the waiting list doesn't ecist, create it
		if (typeof waitingForAvail[v] === "undefined") {
				
			waitingForAvail[v] = new Array() ;
				
		}
			
		waitingForAvail[v].push(func) ;
		return true ;
			
		
		
	} ;
	
	let funcs = new Object() ;
	
	//Load the extentions
	let currentDir = fs.readdirSync(process.cwd()) ;
	//Go through all the files in the cwd.
	for (let doing in currentDir) {
		
		//If it is an extention, load it.
		if (currentDir[doing].substr(currentDir[doing].length - 7,7) === ".jpe.js") {
			
			externals.generateMasterServerObject = () => {
				
				return {
					
					"config": config,
					"reloadConfig":()=>loadConfig()
					
				} ;
				
			} ;
			
			externals.loadMasterExt(currentDir[doing],null,vars,funcs) ;
			
		}
		
	}
	
	//Function to create a new worker.
	const newFork = () => {
		
		console.log("Creating a new worker.") ;
		let thisFork = cluster.fork() ;
		//eslint-disable-next-line consistent-return
		thisFork.on("message",(...args)=>{
			
			let toDo = args[0] ;
			
			//If the worker is logging somthing.
			if (toDo[0] === "log") {
				
				//Add it to the loggs.
				logs.push(toDo[1]) ;
				
			}
			
			//Call a function?
			else if (toDo[0] === "cf") {
				
				//If there is no lock
				if (toDo[2] === null) {
					
					//No exist?
					if (typeof funcs[toDo[1]] === "undefined") {
						
						//Failed.
						thisFork.send(["fc",toDo[1]]) ;
						
					}
					
					else {
						
						//Call and send result.
						thisFork.send(["cf",toDo[1],funcs[toDo[1]](...toDo[3])]) ;
						
					}
					
				}
				
				//With a lock
				else {
					
					//Lock doesn't exist.
					if (typeof funcs[toDo[2]] === "undefined") {
						
						//Failed.
						thisFork.send(["fc",toDo[1]+"---lock"+toDo[2]]) ;
						
					}
					
					//Func doesn't exist.
					else if (typeof funcs[toDo[2]][toDo[1]] === "undefined") {
						
						//Failed.
						thisFork.send(["fc",toDo[1]+"---lock"+toDo[2]]) ;
						
					}
					
					//YAY, lets gooooo
					else {
						
						//Call and send result.
						thisFork.send(["cf",toDo[1]+"---lock"+toDo[2],funcs[toDo[2]][toDo[1]](...toDo[3])]) ;
						
					}
					
				}
				
			}
			
			else if (toDo[0] === "gv") {
				
				//Run when var is available
				whenAvail(toDo[2]===null?toDo[1]:toDo[1]+"---lock"+toDo[2],()=>{
					
					//If there isn't a lock
					if (toDo[2] === null) {
						
						//If the var doesn't exist.
						if (typeof vars[toDo[1]] === "undefined") {
							
							//Dont send a value back.
							thisFork.send(["fv",toDo[1]]) ;
							
						}
						
						//The var exists
						else {
							
							//Send the value back from the root vars because there is no lock.
							thisFork.send(["gv",toDo[1],vars[toDo[1]]]) ;
							
						}
						
						//If we should wait, push the to waiting list
						if (toDo[3]) {
							
							modingVars.push(toDo[1]) ;
							
						}
						
					}
					
					//So there is a lock.
					else {
						
						//If there are no vars for that lock yet.
						if (typeof vars[toDo[2]] === "undefined") {
							
							//Set up an object for the vars.
							//vars[toDo[2]] = new Object() ;
							
							//Var cant exist, so dont send anything back.
							thisFork.send(["fv",toDo[1]+"---lock"+toDo[2]]) ;
							
						}
						
						//If the var doesn't exist.
						else if (typeof vars[toDo[2]][toDo[1]] === "undefined") {
							
							//Dont sand anything back.
							thisFork.send(["fv",toDo[1]+"---lock"+toDo[2]]) ;
							
						}
						
						//Var exists
						else {
							
							//Send it back.
							thisFork.send(["gv",toDo[1]+"---lock"+toDo[2],vars[toDo[2]][toDo[1]]]) ;
							
						}
						
						//If we should wait, push the to waiting list
						if (toDo[3]) {
							
							modingVars.push(toDo[1]+"---lock"+toDo[2]) ;
							
						}
						
					}
					
				}) ;
				
			}
			
			else if (toDo[0] === "sv") {
				
				//Function to accualy set the variable
				const go =()=> {
					
					//If there isn't a lock
					if (toDo[3] === null) {
						
						//Send the var
						vars[toDo[1]] = toDo[2] ;
						
						//Tell the worker it is set.
						thisFork.send(["sv",toDo[1]]) ;
						
					}
					
					//There is a lock
					else {
						
						//If there aren't any vars for this lock yet
						if (typeof vars[toDo[3]] === "undefined") {
							
							//Create the lock object
							vars[toDo[3]] = new Object() ;
							
						}
						
						//Set the var
						vars[toDo[3]][toDo[1]] = toDo[2] ;
						
						//Tell the worker it is set.
						thisFork.send(["sv",toDo[1]+"---lock"+toDo[3]]) ;
						
					}
					
				} ;
				
				const varName = toDo[3]===null?toDo[1]:toDo[1]+"---lock"+toDo[3] ;
				
				//If we don't need to unlock
				if (!toDo[4]) {
					
					//Run when var is available
					whenAvail(varName,go) ;
					
				}
				
				else {
					
					//Set the var first
					go() ;
					//Then remove it from the moding list
					modingVars.splice(modingVars.indexOf(varName),1) ;
					//For everything on the waiting list
					while (waitingForAvail.length) {
						
						//Call the function and remove it from the array
						waitingForAvail[varName].shift()() ;
						
						//If we are locked again, call no more
						if (modingVars.indexOf(varName) > -1) {
							
							break ;
							
						}
						
					}
					
				}
				
			}
			
			//If it is an account action.
			else if (toDo[0] === "proc") {
				
				//Update the workers users if needed.
				if (toDo[1] === "get") {
					
					thisFork.send(["proc-update",authedUsers]) ;
					
				}
				
				else if (toDo[1] === "authed") {
					
					if (typeof authedUsers[toDo[2]][toDo[3]] !== "undefined") {
						
						thisFork.send(["proc-authed",toDo[2],toDo[3],true,toDo[4]]) ;
						
					}
					
					else {
						
						thisFork.send(["proc-authed",toDo[2],toDo[3],false,toDo[4]]) ;
						
					}
					
				}
				
				else if (toDo[1] === "usn") {
					
					if (typeof authedUsers[toDo[2]][toDo[3]] !== "undefined") {
						
						console.log("Res true") ;
						thisFork.send(["proc-usn",toDo[2],toDo[3],true,authedUsers[toDo[2]][toDo[3]]]) ;
						
					}
					
					else {
						
						console.log("Res false") ;
						thisFork.send(["proc-usn",toDo[2],toDo[3],false]) ;
						
					}
					
				}
				
				//Add a new account system.
				else if (toDo[1] === "new") {
					
					//If the account system does not already exist, create it.
					if (typeof authedUsers[toDo[2]] === "undefined") {
						
						authedUsers[toDo[2]] = new Object() ;
						return true ;
						
					}
					
					return false ;
					
				}
				
				//Add a user as autherised to the account system.
				else if (toDo[1] === "add") {
					
					let addTheUser =()=> {
						
						try {
							
							//Set the UID property of the account system as their username.
							authedUsers[toDo[2]][toDo[3]] = toDo[4] ;
							console.log("OK, added") ;
							thisFork.send(["proc-added",toDo[3]]) ;
							return true ;
							
						}
						
						catch(err) {
							
							//Somthing gone wrong - posible not created yet.
							console.warn(`Error ${err} occured while adding a user.\nRetrying in half a second.`) ;
							setTimeout(addTheUser,500) ;
							return false ;
							
						}
						
					} ;
					
					return addTheUser() ;
					
				}
				
				//User logged out.
				else if (toDo[1] === "del") {
					
					delete authedUsers[toDo[2]][toDo[3]] ;
					thisFork.send(["proc-deled",toDo[3]]) ;
					
				}
				
			}
			
		}) ;
		//When a worker exits.
		thisFork.on("exit",()=>{
			
			console.info("A worker died!!!") ;
			console.log("A worker died!!!") ;
			//Create a replacment.
			newFork() ;
			
		}) ;
		
	} ;
	
	//Modules for master.
	let os = require("os") ;
	let net = require("net") ;
	
	//Create a worker for every core that is not being used by the master or another child process.
	for (let doing = 0 ; doing < Math.max(os.cpus().length, 1) ; doing++) {
		
		newFork() ;
		
	}
	
	//Create a server to get the loggs.
	
	if (!(typeof config.dataPort[0] === "number" && config.dataPort[0] < 0)) {
		
		net.createServer(s=>{
			
			if (typeof config.controlers === "object") {
				
				if (config.controlers.length > 0) {
					
					if (config.controlers.indexOf(s.remoteAddress) === -1) {
						
						s.write("No perms") ;
						return ;
						
					}
					
				}
				
			}
			s.on("data",d=>{
				
				d = d.toString() ;
				if (d === "getlogs") {
					
					s.write(logs.join("\n")) ;
					s.end() ;
					
				}
				else if (d === "reload" || d === "stop") {
					
					process.exit(0) ;
					
				}
				
			}) ;
			
		}).listen(...config.dataPort) ;
		
	}
	
	console.info("Master up, all threads spawned.") ;
	console.log("Master up, all threads spawned.") ;
	
	externals.doEvt("ready") ;
		
}
else {
	
	console.log("Worker " + cluster.worker.id + " loaded, starting up now...") ;
	console.info("Worker " + cluster.worker.id + " loaded, starting up now...") ;
	//Load the server module & init it.
	requireJPS("server").init(cluster) ;
	console.log(`Worker ${cluster.worker.id} running.`) ;
	console.info(`Worker ${cluster.worker.id} running.`) ;
	
}