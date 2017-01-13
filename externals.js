let fs = require("fs") ;
let vm = require("vm") ;

let handles = new Object() ;
module.exports.handles = handles ;
function handle(evt,func) {
	
	handles[evt] = func ;
	
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
		return serverObj ;
		
	}
	catch(err) {
		
		return {"loaded":false,"error":err} ;
		
	}
	
}