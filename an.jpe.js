if (!server.isMaster) {
	
	let currentReq = 0 ;
	server.handle("request",_=>{
		
		console.info(currentReq++) ;
		return false ;
		
	}) ;
	
}