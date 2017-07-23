if (!server.isMaster) {
	server.addCORSRule(2, "default:0", /.*/g, null, true, [], ["GET", "HEAD"], ["test"]) ;
}