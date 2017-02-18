//If this extention is loaded in the master process.
if (server.isMaster) {
	
	console.info("Master extention loading!!!") ;
	
	//Set the global variable "set_in_master" to the value "Hello :)".
	server.setGlobal("set_in_master","Hello :)").then(_=>{
		
		//When it is set, get the variable and log it.
		server.getGlobal("set_in_master").then(d=>console.info(d)) ;
		
	}) ;
	
}

//Not master, so worker.
else {
	
	console.info("Worker extention loading!!!") ;
	
	//Get the global var "set_in_master" that we set in the master and log it.
	server.getGlobal("set_in_master").then(d=>console.info("Getting the global set in master resolved:",d)) ;
	
	//Request handle is called as soon as  a request is recieved & before anything important happens.
	server.handle("request",(req,resp)=>{
		
		console.info("Got it") ;
		
		if (req.url === "/Hello_World") {
			
			resp.writeHead(200) ;
			resp.end("YAY") ;
			//When true is returned, the server will stop handling this request, it is to say that you have handled it.
			return true ;
			
		}
		
		//Get the global variable "hello_world"
		server.getGlobal("hello_world").then(d=>{
			
			//Log the variable, it will be undefined the first time the page loads because it has not been set yet.
			console.info(d) ;
			//Now we set the variable to "test"
			server.setGlobal("hello_world","test") ;
			
		}) ;
		
		//We return false & thus the server carries on handling the request as it would.
		return false ;
		
	}) ;
	
	//This is called when a request has been processed, redirects, host alieses etc applied.
	server.handle("fullrequest",(req,resp)=>{
		
		console.info("Got it - full") ;
		
		//The orig_url property is the origional, unchanged URL.
		console.info(req.orig_url) ;
		
		//The URL now has the hostname at the start of it & is lower cased.
		console.info(req.url) ;
		
		if (req.url === "/hello_world_yay") {
			
			resp.writeHead(200) ;
			resp.end("YAY 2.0") ;
			//Return true so ther server no longer handles it.
			return true ;
			
		}
		
		//The server can carry on because we return false.
		return false ;
		
	}) ;
	
	server.handle("request",(req,resp)=>{
		
		//Multiple handles can be set, they will all get fired, even if 1 returns true, so remember that.
		console.info("This also gets triggered!") ;
		
		//A promise can also be returned, the resolve value is the equivilent of the return value.
		return new Promise(resolve=>resolve(false)) ;
		
	}) ;
	
}