/*
	
	JOTPOT Server.
	Copyright (c) Jacob O'Toole 2016-2017
	
*/

//Version 1.0.0


let fs = require("fs") ;
let vm = require("vm") ;

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

module.exports.loadExt = (file,serverObj) => {
	
	if (!fs.existsSync(file)) {
		
		return {"loaded":false,"error":"Not found"} ;
		
	}
	
	if (!fs.statSync(file).isFile()) {
		
		return {"loaded":false,"error":"Not file"} ;
		
	}
	
	serverObj.handle = handle ;
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