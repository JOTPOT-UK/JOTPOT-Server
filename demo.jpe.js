server.handle("request",(req,resp)=>{
	
	console.info("Got it") ;
	console.info(resp.vars) ;
	if (req.url === "/Hello_World") {
		
		resp.writeHead(200) ;
		resp.end("YAY") ;
		return true ;
		
	}
	return false ;
	
}) ;

server.handle("fullrequest",(req,resp)=>{
	
	console.info("Got it - full") ;
	console.info(req.orig_url) ;
	console.info(req.url) ;
	console.info(JSON.stringify(resp.vars)) ;
	if (req.url === "/hello_world_yay") {
		
		resp.writeHead(200) ;
		resp.end("YAY 2.0") ;
		return true ;
		
	}
	return false ;
	
}) ;