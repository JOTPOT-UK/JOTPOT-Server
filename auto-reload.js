let cp = require("child_process") ;
function loadUp() {
	
	current = cp.fork("run") ;
	current.on("exit",loadUp) ;
	
}
loadUp() ;