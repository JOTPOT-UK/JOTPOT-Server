let net = require("net") ;
let s = net.connect({host:"192.168.1.11",port:500}) ;
s.on("data",d=>{
	
	console.log(d.toString()) ;
	s.end() ;
	
}) ;
s.write("getlogs") ;