process.title = "JOTPOT Server 3" ;

let logs = new Array() ;
console.log = console.warn = (...args) => {
	
	logs.push(args.join(" ")) ;
	
}/**/

let fs = require("fs") ;

let config ;
if (fs.existsSync("config.json")) {
	
	config = fs.readFileSync("config.json").toString() ;
	
	try {
		
		config = JSON.parse(config) ;
		
	}
	
	catch(err) {
		
		console.warn("Error parsing config.json!") ;
		console.info("Error parsing config.json!") ;
		console.warn(err) ;
		console.info(err) ;
		console.warn("Exiting") ;
		console.info("Exiting") ;
		process.exit() ;
		
	}
	
}

else if (fs.existsSync("config.json")) {
	
	console.warn("Config file does not exist.") ;
	console.info("Config file does not exist.") ;
	console.warn("Exiting") ;
	console.info("Exiting") ;
	process.exit() ;
	
}

let cluster = require("cluster") ;
//let net = require("net") ;
let UIDs = new Array() ;
let authedUsers = new Object() ;

if (cluster.isMaster) {
	
	console.info("Loading JOTPOT Server.") ;
	console.log("Loading JOTPOT Server.") ;
	process.title = "JOTPOT Server 3" ;
	function newFork() {
		
		console.log("Creating a new worker.") ;
		let thisFork = cluster.fork() ;
		thisFork.on("message",(...args)=>{
			
			let toDo = args[0] ;
			
			if (toDo[0] === "log") {
				
				logs.push(toDo[1]) ;
				
			}
			
			else if (toDo[0] === "proc") {
				
				if (toDo[1] === "get") {
					
					thisFork.send(["proc-update",authedUsers]) ;
					
				}
				
				else if (toDo[1] === "new") {
					
					if (typeof authedUsers[toDo[2]] === "undefined") {
						
						authedUsers[toDo[2]] = new Object() ;
						return true ;
						
					}
					
					else {
						
						return false ;
						
					}
					
				}
				
				else if (toDo[1] === "add") {
					
					//console.log(`Adding ${toDo[3]} as ${toDo[4]} to ${toDo[2]}`) ;
					
					let addTheUser =_=> {
						
						try {
							
							authedUsers[toDo[2]][toDo[3]] = toDo[4] ;
							return true ;
							
						}
						
						catch(err) {
							
							console.warn(`Error ${err} occured while adding a user.\nRetrying in half a second.`) ;
							setTimeout(addTheUser,500) ;
							return false ;
							
						}
						
					}
					
					return addTheUser() ;
					return false ;
					
				}
				
				else if (toDo[1] === "del") {
					
					delete authedUsers[toDo[2]][toDo[3]] ;
					
				}
				
			}
			
		}) ;
		thisFork.on("exit",(workerClosed)=>{
			
			console.info("A worker died!!!") ;
			console.log("A worker died!!!") ;
			newFork() ;
			
		}) ;
		
	}
	
	let os = require("os") ;
	let cp = require("child_process") ;
	let net = require("net") ;
	let workers = new Array() ;
	let otherProcesses = new Array() ;
	
	for (let doing = 0 ; doing < os.cpus().length - 1 - config.otherProcesses.length ; doing++) {
		
		newFork() ;
		
	}
	
	for (let doing in config.otherProcesses) {
		
		otherProcesses.push(cp.fork(config.otherProcesses[doing].filename)) ;
		
	}
	
	//setInterval(_=>console.log(Object.keys(cluster.workers)),5000) ;
	net.createServer(s=>{
	
	s.on("data",d=>{
			
			d = d.toString() ;
			if (d === "getlogs") {
				
				s.write(logs.join("\n")) ;
				
			}
			
		}) ;
		
	}).listen(config.dataPort || 500) ;
	console.info("Master up, all threads spawned.") ;
	console.log("Master up, all threads spawned.") ;
		
}
else {
	
	console.log("Worker " + cluster.worker.id + " loaded, starting up now...") ;
	console.info("Worker " + cluster.worker.id + " loaded, starting up now...") ;
	require("./server.js").init(cluster) ;
	console.log(`Worker ${cluster.worker.id} running.`) ;
	console.info(`Worker ${cluster.worker.id} running.`) ;
	
}