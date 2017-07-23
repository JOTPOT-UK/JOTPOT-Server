if (!server.isMaster) {
	server.addCORSRule(2, "default:0", /.*/g, null, true, [], ["GET", "HEAD"], ["test"]) ;
	server.cache.createLink("default:0/page1", "default:0/page2") ;
	server.cache.createLink("default:0/page3", "default:0/page4") ;
	server.cache.createLink("default:0/page4", "default:0/page5") ;
	server.cache.createLink("default:0/page6", "default:0/page7") ;
	server.handle("allowedrequest", req => {
		if (req.url.pathname === "/page2") {
			req.url.pathname = "/page3" ;
		}
	}) ;
	server.handlePage("default:0/page5", req => {
		req.url.pathname = "/page6" ;
	}) ;
}