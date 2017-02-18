/*
	
	JOTPOT Server.
	Copyright (c) Jacob O'Toole 2016-2017
	
*/

//Version 1.0.1


//Node modules
let fs = require("fs") ;
let vm = require("vm") ;
let events = require("events") ;

let handles = new Object() ;
let handleOver = new Object() ;

module.exports.handles = handles ;

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
module.exports.loadExt = (file,serverObj) => {
	
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
	
	//Just, dissapointing...
	serverObj.isMaster = false ;
	
	//Add the handle function to the serverObj
	serverObj.handle = handle ;
	
	//Functiom to set a global variable. 1 arg is var name to get.
	serverObj.getGlobal = varTG => {
		
		return new Promise(resolve=>{
			
			//When we get the variable.
			varEvt.once("got " + varTG,d=>{
				
				//Resolve with the contents
				resolve(d) ;
				
			});
			
			//Send a request to the master process to get the variable.
			process.send(["gv",varTG]) ;
			
		}) ;
		
	} ;
	
	//Function to set a variable. 1st arg is var name, second is value to set it to.
	serverObj.setGlobal = (varTS,val) => {
		
		return new Promise(resolve=>{
			
			//When it is set, resolve
			varEvt.once("set " + varTS,_=>resolve()) ;
			//Tell the master to set it.
			process.send(["sv",varTS,val]) ;
			
		}) ;
		
	} ;
	
	//Wrap the source
	let source = `(function(require,server,console,setTimeout,setInterval){${fs.readFileSync(file).toString()}});` ;
	
	//To catch so ther server doesn't go down if the extention fails to load.
	try {
		
		//Run it with the required arguments
		vm.runInNewContext(source,{},{
			
			filename: file + "fun"
			
		})(require,serverObj,console,setTimeout,setInterval) ;
		
		//So, it loaded...
		return {"loaded":true,"serverObj":serverObj} ;
		
	}
	catch(err) {
		
		//Somthing went wrong:
		return {"loaded":false,"error":err} ;
		
	}
	
}

//Function to load an extention in master mode. Similar to loadExt. vars arg is vars object for seting and getting globals.
module.exports.loadMasterExt = (file,serverObj,vars) => {
	
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
	
	//This is the MASTER!!!
	serverObj.isMaster = true ;
	
	//Add the handle function to the serverObj
	serverObj.handle = handle ;
	
	//Functiom to set a global variable. 1 arg is var name to get.
	serverObj.getGlobal = varTG => {
		
		return new Promise(resolve=>{
			
			//Resolve instantly, because it is master, there are no async events to do :(
			resolve(vars[varTG]) ;
			
		}) ;
		
	} ;
	
	//Function to set a variable. 1st arg is var name, second is value to set it to.
	serverObj.setGlobal = (varTS,val) => {
		
		//set the value and resolve instantly, because it is master, there are no async events to do :(
		vars[varTS] = val ;
		resolve() ;
		
	} ;
	
	//Wrap the source
	let source = `(function(require,server,console,setTimeout,setInterval){${fs.readFileSync(file).toString()}});` ;
	
	//To catch so ther server doesn't go down if the extention fails to load.
	try {
		
		//Run it with the required arguments
		vm.runInNewContext(source,{},{
			
			filename: file + "fun"
			
		})(require,serverObj,console,setTimeout,setInterval) ;
		
		//So, it loaded...
		return {"loaded":true,"serverObj":serverObj} ;
		
	}
	catch(err) {
		
		//Somthing went wrong:
		return {"loaded":false,"error":err} ;
		
	}
	
}