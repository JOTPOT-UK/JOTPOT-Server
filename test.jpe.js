//We cannot set pages in master, we have to set them in our workers.
if (!server.isMaster) {

	server.pages["localhost/testpage"] = ["cache","OK"] ;
	server.handle("fullrequest",_=>{
		
		console.info("Done") ;
		server.pages["localhost/testpage"] = ["cache","YAY!!!"] ;
		
	}) ;

}