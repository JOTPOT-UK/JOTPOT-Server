const net = require("net") ;
const fs = require("fs") ;
console.log("Sup") ;
/*let con = net.connect({
	fd: 3
}) ;
con.on("data", d=>console.log(d.toString())) ;
con.on("connect", _=>console.log("Connected")) ;
con.on("end", _=>console.log("Ended")) ;

setTimeout(_=>{
	con.write("Hello :)") ;
	console.log("Writen") ;
}, 5000) ;

console.log("LOL!!!") ;
*/
//fs.writeSync(3, "Hello World :)") ;
console.log("OK, writen?") ;
let con = net.createConnection({
	fd: 3
}, _=>{
	console.log("LOL") ;
	con.write("Oh yes!") ;
}) ;