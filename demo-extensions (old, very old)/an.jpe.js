if (!server.isMaster) {
	
	let currentReq = 0 ;
	server.handle("request",()=>{
		
		console.info(currentReq++) ;
		return false ;
		
	}) ;
	
}