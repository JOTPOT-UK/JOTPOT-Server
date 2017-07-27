let cp = require("child_process") ;
function loadUp() {
	
	let current = cp.fork("run") ;
	current.on("exit",loadUp) ;
	
}
loadUp() ;