//This file tests handlers, and the templating

if (!server.isMaster) {
	//Add time, href and method variable to all allowed requests.
	server.handle("allowedrequest", (req, resp) => {
		resp.vars.time = req.time.getTime().toString(10) ;
		resp.vars.method = req.method.toUpperCase() ;
		resp.vars.href = req.url.href ;
	}) ;
	//Add user_ip, user_ip_remote variables to default/index.html
	server.handlePage("default/index.html", (req, resp) => { // eslint-disable-line consistent-return
		req.HandledPOST = true ;
		resp.addVars = 1 ;
		resp.vars["user_ip"] = req.ip ;
		resp.vars["user_ip_remote"] = req.remoteAddress ;
		//Add the POST Data to the end if we need to
		resp.vars.footer = "" ;
		if (req.method === "POST") {
			return server.getData(req).then(d=>{
				resp.vars.footer = `<h3>POST Data:</h3>${server.escapeHTML(d.toString())}` ;
			}) ;
		}
	}) ;
}
