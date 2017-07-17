if (server.isMaster) {

	console.info("Locked demo loaded as master, but limited is",server.limited) ;

}

else {
	
	console.info("Locked demo loaded as a worker, but limited is",server.limited) ;
	
	server.handle("request",_=>{
		
		console.info("Locked demo got the request!!!") ;
		
	}) ;
	
}

server.getGlobal("set_in_master").then(d=>{
	
	//Notice how even though this var is set, we cant access it.
	console.info("What the limited extention sees as \"set_in_master\":",d) ;
	
	//Lets set a var, you will see this accessed from the demo.jpe.js too.
	server.setGlobal("limited","YAY").then(_=>{
		
		//The global is set, so lets read it:
		server.getGlobal("limited").then(d=>{
			
			//So we can access this, but not the globals from unlimited extentions.
			console.info("But when we get out own var:",d) ;
			
		}) ;
		
	}) ;
	
}) ;
