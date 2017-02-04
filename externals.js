/*
	
	JOTPOT Server.
	Copyright (c) Jacob O'Toole 2016-2017
	
*/

//Version 1.0.1


let fs = require("fs") ;
let vm = require("vm") ;
let events = require("events") ;

let handles = new Object() ;
let handleOver = new Object() ;
module.exports.handles = handles ;
function handle(evt,func,allowOverwrite=true) {
	
	if (typeof handleOver[evt] === "undefined") {
		
		handleOver[evt] = allowOverwrite ;
		
	}
	
	else if (handleOver[evt] === false) {
		
		return false ;
		
	}
	
	handles[evt] = func ;
	return true ;
	
}

let varEvt = new events() ;
process.on("message",m=>{
	
	if (m[0] === "gv") {
		
		varEvt.emit("got " + m[1],m[2]) ;
		
	}
	
	else if (m[0] === "sv") {
		
		varEvt.emit("set " + m[1]) ;
		
	}
	
}) ;

module.exports.loadExt = (file,serverObj) => {
	
	if (!fs.existsSync(file)) {
		
		return {"loaded":false,"error":"Not found"} ;
		
	}
	
	if (!fs.statSync(file).isFile()) {
		
		return {"loaded":false,"error":"Not file"} ;
		
	}
	
	serverObj.handle = handle ;
	serverObj.getGlobal = varTG => {
		
		return new Promise(_=>{
			
			varEvt.once("got " + varTG,d=>{
				
				resolve(d) ;
				
			});
			process.send(["gv",varTG]) ;
			
		}) ;
		
	} ;
	
	serverObj.setGlobal = (varTS,val) => {
		
		return new Promise(_=>{
			
			varEvt.once("set " + varTS,_=>resolve()) ;
			process.send("sv",varTS,val) ;
			
		}) ;
		
	} ;
	
	let source = `(function(require,server,console,setTimeout,setInterval){${fs.readFileSync(file).toString()}});` ;
	try {
		
		vm.runInNewContext(source,{},{
			
			filename: file + "fun"
			
		})(require,serverObj,console,setTimeout,setInterval) ;
		return {"loaded":true,"serverObj":serverObj} ;
		
	}
	catch(err) {
		
		return {"loaded":false,"error":err} ;
		
	}
	
}