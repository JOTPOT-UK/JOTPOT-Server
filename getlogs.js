let net = require("net") ;
let s = net.connect({host:"localhost",port:500}) ;
s.on("data",d=>{
	
	console.log(d.toString()) ;
	s.end() ;
	
}) ;
s.write("getlogs") ;