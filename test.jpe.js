server.pages["localhost/testpage"] = ["cache","OK"] ;
server.handle("fullrequest",_=>{
	
	console.info("Done") ;
	server.pages["localhost/testpage"] = ["cache","YAY!!!"] ;
	
}) ;