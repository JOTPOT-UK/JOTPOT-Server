/*
	
	JOTPOT Server
	Version 25E
	
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


//Node modules
let fs = require("fs") ;
let vm = require("vm") ;
let events = require("events") ;
let fakefs = require("./ExtensionFileSystem.js") ;

let handles = new Object() ;
let handleOver = new Object() ;

module.exports.handles = handles ;

module.exports.generateServerObject = _ => {return {};} ;
module.exports.generateLimitedServerObject = _ => {return {};} ;
module.exports.generateMasterServerObject = _ => {return {};} ;
module.exports.generateLimitedMasterServerObject = _ => {return {};} ;

function handle(evt,func,allowOverwrite=true) {
	
	//New handles allowed?
	if (typeof handleOver[evt] === "undefined") {
		
		handleOver[evt] = allowOverwrite ;
		
	}
	
	else if (handleOver[evt] === false) {
		
		return false ;
		
	}
	
	
	//If there are no handles yet for the event.
	if (typeof handles[evt] === "undefined") {
		
		//Set this as the only handle
		handles[evt] = [func] ;
		
	}
	
	else {
		
		//Or add this as a handle
		handles[evt].push(func) ;
		
	}
	
	//YAY Handle set!!!
	return true ;
	
}

module.exports.doEvt = function (evt,...args) {
	
	return new Promise((resolve,reject)=>{
	
	//If the event has no handles.
	if (typeof handles[evt] === "undefined") {
		
		//Carry on
		resolve(false) ;
		
	}
	
	//Will resolve false if nothing returns true.
	let rv = false ;
	let totalPromises = 0 ;
	
	//Function to handle resolving promises
	let handleResolved = (d) => {
		
		totalPromises-- ;
		//If it is handled
		if (d) {
			
			//Then we need to resolve true, but not yet.
			rv = true ;
			
		}
		
		//If there aren't any pending promesses left, then resolve.
		if (totalPromises === 0) {
			
			//resolve what we need to resolve.
			resolve(rv) ;
			
		}
		
	}
	
	//Go through all the handles for the event.
	for (let doing in handles[evt]) {
		
		//Call the handle
		let val = handles[evt][doing](...args) ;
		
		//If it is falsey, then we cant check because it will error when we try to get .then.
		if (!val) {
			
			continue ;
			
		}
		
		//Is it a promise?
		if (typeof val.then === "function") {
			
			//One more pending promise.
			totalPromises++ ;
			
			//Set up the handleResolved function when it resolves.
			val.then(handleResolved) ;
			
		}
		
		//If it returns true
		else if (val) {
			
			//Then we should resolve true... But later
			rv = true ;
			
		}
		
	}
	
	//If there aren't any pending promesses, then resolve.
	if (totalPromises === 0) {
		
		//resolve what we need to resolve.
		resolve(rv) ;
		
	}
	
	}) ;
	
}

//Set up an event emmitor for recieving variable stuff.
let varEvt = new events() ;
process.on("message",m=>{
	
	//Got a var
	if (m[0] === "gv") {
		
		varEvt.emit("got " + m[1],m[2]) ;
		
	}
	
	
	//Function called.
	if (m[0] === "cf") {
		
		//Send of the event.
		varEvt.emit("called " + m[1],m[2]) ;
		
	}
	
	//Set a var
	else if (m[0] === "sv") {
		
		varEvt.emit("set " + m[1]) ;
		
	}
	
	//Var no exist
	else if (m[0] === "fv") {
		
		varEvt.emit("got " + m[1],void(0)) ;
		
	}
	
}) ;

//Function to load an extention. File is path to extention, and serverObj is the object that will be the base of the server variable.
module.exports.loadExt = (file,lock=null) => {
	
	//If the extention doesn't exist...
	if (!fs.existsSync(file)) {
		
		//We don't load with the error "Not found"
		return {"loaded":false,"error":"Not found"} ;
		
	}
	
	//Ah, but if it isn't a file:
	if (!fs.statSync(file).isFile()) {
		
		//We don't load with the error "Not file"
		return {"loaded":false,"error":"Not file"} ;
		
	}
	
	let serverObj ;
	let wasThereALock = true ;
	//If there is no lock
	if (lock === null) {
		
		//Create a new one with no limits.
		lock = new module.exports.lock() ;
		
		serverObj = module.exports.generateServerObject(lock.hosts, lock.fs) ;
		
		//This extention is not limited
		serverObj.limited = false ;
		
		wasThereALock = false ;
		
	}
	
	//LOCKED!!!
	else {
		
		serverObj = module.exports.generateLimitedServerObject(lock.hosts, lock.fs) ;
		
		//Now we are locked
		serverObj.limited = true ;
		
	}
	
	//Just, dissapointing...
	serverObj.isMaster = false ;
	
	//Add the handle function to the serverObj
	serverObj.handle = (evts,func) => {
		
		if (typeof evts === "string") {
			
			let evt = evts ;
			
			//If it is only allowed to handle cirton hosts.
			if (lock.mode === 1) {
				
				//And it is trying to handle requests.
				if (evt === "request" || evt === "fullrequest" || evt === "allowedrequest") {
					
					//Go through all the hosts it can access
					for (let doing in lock.hosts) {
						
						//Handle the event, but only for the host.
						handle(`${lock.hosts[doing]}/${evt}`,func) ;
						
					}
					return ;
					
				}
				
				else if (evt.indexOf("/") !== -1) {
					
					if (evt.lastIndexOf("/request") === evt.length - 8 || evt.lastIndexOf("/fullrequest") === evt.length - 12 || evt.lastIndexOf("/allowedrequest") === evt.length - 15) {
						
						let isViolation = true ;
						//Go through all the hosts it can access
						for (let doing in lock.hosts) {
							
							//If the event is for this host
							if (evt === `${lock.hosts[doing]}/${evt.split("/")[evt.split("/").length-1]}`) {
								
								//Not a violation, we can stop checking now.
								isViolation = false ;
								break ;
								
							}
							
						}
						
						//If it is a violation
						if (isViolation) {
							
							console.warn("Not allowed to handle event!") ;
							return ;
							
						}
						
					}
					
				}
				
			}
			
			//We can just handle the event now.
			handle(evt,func) ;
			return ;
			
		}
		
		for (let doing in evts) {
			
			let evt = evts[doing] ;
			
			//If it is only allowed to handle cirton hosts.
			if (lock.mode === 1) {
				
				//And it is trying to handle requests.
				if (evt === "request" || evt === "fullrequest" || evt === "allowedrequest") {
					
					//Go through all the hosts it can access
					for (let doing in lock.hosts) {
						
						//Handle the event, but only for the host.
						handle(`${lock.hosts[doing]}/${evt}`,func) ;
						
					}
					continue ;
					
				}
				
				else if (evt.indexOf("/") !== -1) {
					
					if (evt.lastIndexOf("/request") === evt.length - 8 || evt.lastIndexOf("/fullrequest") === evt.length - 12 || evt.lastIndexOf("/allowedrequest") === evt.length - 15) {
						
						let isViolation = true ;
						//Go through all the hosts it can access
						for (let doing in lock.hosts) {
							
							//If the event is for this host
							if (evt === `${lock.hosts[doing]}/${evt.split("/")[evt.split("/").length-1]}`) {
								
								//Not a violation, we can stop checking now.
								isViolation = false ;
								break ;
								
							}
							
						}
						
						//If it is a violation
						if (isViolation) {
							
							console.warn("Not allowed to handle event!") ;
							continue ;
							
						}
						
					}
					
				}
				
			}
			
			//We can just handle the event now.
			handle(evt,func) ;
			continue ;
			
		}
		
	} ;
	
	//Function for worker to call a master function.
	serverObj.callFunc = (toCall,...args) => {
		
		return new Promise(resolve=>{
			
			//OK, when it is called.
			varEvt.once("called " + toCall + (lock.vars===null?"":"---lock"+lock.vars) ,d=>{
				
				//Lets reaolve.
				resolve(d) ;
				
			});
			
			//Send of to call it.
			process.send(["cf",toCall,lock.vars,args]) ;
			
		}) ;
		
	} ;
	
	//Functiom to set a global variable. 1st arg is var name to get, seond is wether or not to mo
	serverObj.getGlobal = (varTG,moding=false) => {
		
		return new Promise(resolve=>{
			
			//When we get the variable.
			varEvt.once("got " + varTG + (lock.vars===null?"":"---lock"+lock.vars) ,d=>{
				
				//If we are modding, resolve with the contents and a callack
				if (moding) {
					
					resolve([d,val=>{
						
						return new Promise(resolve=>{
							
							//When it is set, resolve
							varEvt.once("set " + varTG + (lock.vars===null?"":"---lock"+lock.vars) ,_=>resolve()) ;
							//Tell the master to set it.
							process.send(["sv",varTG,val,lock.vars,true]) ;
							
						}) ;
						
					}]) ;
					
				}
				
				else {
					
					//Resolve with the contents only
					resolve(d) ;
					
				}
				
			});
			
			//Send a request to the master process to get the variable.
			process.send(["gv",varTG,lock.vars,moding]) ;
			
		}) ;
		
	} ;
	
	//Alias for server.getGlobal but sets moding to true
	serverObj.modGlobal = v => serverObj.getGlobal(v,true) ;
	
	//Function to set a variable. 1st arg is var name, second is value to set it to.
	serverObj.setGlobal = (varTS,val) => {
		
		return new Promise(resolve=>{
			
			//When it is set, resolve
			varEvt.once("set " + varTS + (lock.vars===null?"":"---lock"+lock.vars) ,_=>resolve()) ;
			//Tell the master to set it.
			process.send(["sv",varTS,val,lock.vars,false]) ;
			
		}) ;
		
	} ;
	
	//Only add if not locked extention.
	if (!wasThereALock) {
		
		//Function for extention to load another extention.
		serverObj.loadExt = (ePath,eLock=null) => {
			
			//Return their server object.
			return module.exports.loadExt(ePath,eLock) ;
			
		}
		
		//Create a clone of the lock class to add to the server object.
		serverObj.lock = class extends module.exports.lock {} ;
		
	}
	
	//Wrap the source
	let source = `(function(require,server,console,setTimeout,setInterval,setImmediate,fs){\r\n${fs.readFileSync(file).toString()}\r\n});` ;
	
	//To catch so ther server doesn't go down if the extention fails to load.
	try {
		
		//Run it with the required arguments
		vm.runInNewContext(source,{},{
			
			filename: file
			
		})(require,serverObj,console,setTimeout,setInterval,setImmediate,lock.fs) ;
		
		//So, it loaded...
		return {"loaded":true,"serverObj":serverObj} ;
		
	}
	catch(err) {
		
		//Somthing went wrong:
		console.warn("Error loading extention:") ;
		console.warn(err) ;
		return {"loaded":false,"error":err} ;
		
	}
	
}

//Function to load an extention in master mode. Similar to loadExt. vars arg is vars object for seting and getting globals.
module.exports.loadMasterExt = (file,lock=null,vars,funcs) => {
	
	//If the extention doesn't exist...
	if (!fs.existsSync(file)) {
		
		//We don't load with the error "Not found"
		return {"loaded":false,"error":"Not found"} ;
		
	}
	
	//Ah, but if it isn't a file:
	if (!fs.statSync(file).isFile()) {
		
		//We don't load with the error "Not file"
		return {"loaded":false,"error":"Not file"} ;
		
	}
	
	let serverObj ;
	let wasThereALock = true ;
	//If there is no lock
	if (lock === null) {
		
		//Create a new one with no limits.
		lock = new module.exports.lock() ;
		
		serverObj = module.exports.generateMasterServerObject(lock.hosts, lock.fs) ;
		
		//This extention is not limited
		serverObj.limited = false ;
		
		wasThereALock = false ;
		
	}
	
	//LOCKED!!!
	else {
		
		serverObj = module.exports.generateLimitedMasterServerObject(lock.hosts, lock.fs) ;
		
		//Now we are locked
		serverObj.limited = true ;
		
	}
	
	//This is the MASTER!!!
	serverObj.isMaster = true ;
	
	//Add the handle function to the serverObj
	serverObj.handle = handle ;
	
	//Function for master to set function that cal be called by workers.
	serverObj.setFunc = (toSet,func) => {
		
		return new Promise(resolve=>{
			
			//If there is no lock...
			if (lock.vars === null) {
				
				//Then just set the function.
				funcs[toSet] = func ;
				
			}
			
			//But if there is a lock.
			else {
				
				//And there are no functions for this lock/
				if (typeof funcs[lock.vars] === "undefined") {
					
					//Create an object to put them in.
					funcs[lock.vars] = new Object() ;
					
				}
				
				//And set the function.
				funcs[lock.vars][toSet] = func ;
				
			}
			
			//All done, resolve instantly.
			resolve() ;
			
		}) ;
		
	} ;
	
	//Function to set a global variable. 1 arg is var name to get.
	serverObj.getGlobal = varTG => {
		
		return new Promise(resolve=>{
			
			//Resolve instantly, because it is master, there are no async events to do :(
			if (lock.vars === null) {
				
				//No lock so get var fromm root variables.
				resolve(vars[varTG]) ;
				
			}
			
			else {
				
				//Does the lock accualy have any variables.
				if (typeof vars[lock.vars] === "undefined") {
					
					//Resolve undefined because the var cant exist...
					resolve(void(undefined)) ;
					
				}
				
				else {
					
					//Get var under the lock object.
					resolve(vars[lock.vars][varTG]) ;
					
				}
				
			}
			
		}) ;
		
	} ;
	
	//Function to set a variable. 1st arg is var name, second is value to set it to.
	serverObj.setGlobal = (varTS,val) => {
		
		return new Promise(resolve=>{
			
			//set the value and resolve instantly, because it is master, there are no async events to do :(
			if (lock.vars === null) {
				
				//No lock
				vars[varTS] = val ;
				
			}
			
			else {
				
				//With a lock.
				
				if (typeof vars[lock.vars] === "undefined") {
					
					vars[lock.vars] = new Object() ;
					
				}
				
				vars[lock.vars][varTS] = val ;
				
			}
			
			resolve() ;
			
		}) ;
		
	} ;
	
	//Only add if not locked extention
	if (!wasThereALock) {
		
		//Functino for extention to load another extention.
		serverObj.loadExt = (ePath,eLock=null) => {
			
			//Return their server object.
			return module.exports.loadMasterExt(ePath,eLock,vars) ;
			
		}
		
		//Create a clone of the lock class to add to the server object.
		serverObj.lock = class extends module.exports.lock {} ;
		
	}
	
	//Wrap the source
	let source = `(function(require,server,console,setTimeout,setInterval,setImmediate,fs){\r\n${fs.readFileSync(file).toString()}\r\n});` ;
	
	//To catch so ther server doesn't go down if the extention fails to load.
	try {
		
		//Run it with the required arguments
		vm.runInNewContext(source,{},{
			
			filename: file + "fun"
			
		})(require,serverObj,console,setTimeout,setInterval,setImmediate,lock.fs) ;
		
		//So, it loaded...
		return {"loaded":true,"serverObj":serverObj} ;
		
	}
	catch(err) {
		
		//Somthing went wrong:
		console.info(err) ;
		return {"loaded":false,"error":err} ;
		
	}
	
}

module.exports.lock = class {
	
	constructor(vars=null,mode=0,hosts=[]) {
		
		if (typeof vars !== "string" && vars !== null) {
			
			throw new Error("First argument must be a string or null") ;
			
		}
		
		if (typeof mode !== "number") {
			
			throw new Error("2nd argument (mode) must be a number") ;
			
		}
		
		else if (Math.round(mode) !== mode) {
			
			throw new Error("2nd argument (mode) must be an integer.") ;
			
		}
		
		else if (mode < 0 || mode > 2) {
			
			throw new Error("2nd argument (mode) must be 0, 1 or 2") ;
			
		}
		
		if (mode === 2) {
			
			console.info("Lock mode 2 not yet supported, assuming mode 0.") ;
			console.warn("Lock mode 2 not yet supported, assuming mode 0.") ;
			mode = 0 ;
			
		}
		
		this.vars = vars ;
		this.mode = mode ;
		if (mode > 0) {
			
			let hostValid = true ;
			
			if (typeof hosts !== "object") {
				
				hostValid = false ;
				
			}
			
			else if (typeof hosts[0] === "undefined") {
				
				hostValid = false ;
				
			}
			
			if (!hostValid) {
				
				throw new Error("3rd argument must be an array containing at least one host, unless mode is 0") ;
				
			}
			
			this.hosts = hosts ;
			
			if (hosts.length === 0) {
				
				this.fs = require("fs") ;
				
			}
			
			else {
				
				this.fs = new fakefs(hosts[0]) ;
				this.fs = this.fs.public ;
				
			}
			
		}
		
	}
	
}
