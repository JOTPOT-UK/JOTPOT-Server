server.handle("request",(req,resp)=>{
	
	console.info("Got it") ;
	if (req.url === "/Hello_World") {
		
		resp.writeHead(200) ;
		resp.end("YAY") ;
		return true ;
		
	}
	server.getGlobal("hello_world").then(d=>{
		
		console.info(d) ;
		server.setGlobal("hello_world","test") ;
		
	}) ;
	return false ;
	
}) ;

server.handle("fullrequest",(req,resp)=>{
	
	console.info("Got it - full") ;
	console.info(req.orig_url) ;
	console.info(req.url) ;
	if (req.url === "/hello_world_yay") {
		
		resp.writeHead(200) ;
		resp.end("YAY 2.0") ;
		return true ;
		
	}
	return false ;
	
}) ;