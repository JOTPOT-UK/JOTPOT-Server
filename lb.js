let config = JSON.parse(require("fs").readFileSync("lbconfig.json").toString()) ;

for (let doing in config.servers) {
	
	config.servers[doing].load = 0 ;
	if (config.servers[doing].capacity > 999999) {
		
		console.warn("Having a server capacity of over 999999 can make the load balencer less efficient. It is not recomended to do this.") ;
		
	}
	config.servers[doing].capacity = 1 / config.servers[doing].capacity ;
	
}

let http = require("http") ;

function sendToServer(req,resp,server,st) {
	
	config.servers[server].load += config.servers[server].capacity ;
	req.headers["jp-source"] = req.headers["jp-source"] || req.connection.remoteAddress || req.socket.remoteAddress || req.connection.socket.remoteAddress ;
	
	let f = http.request({
		
		protocol:"http:",
		host:config.servers[server].host,
		port:config.servers[server].port,
		method:req.method,
		path:req.url,
		headers:req.headers
		
	}) ;
	f.on("error",err=>{
		
		console.warn("Error") ;
		console.warn(err) ;
		config.servers[server].load -= config.servers[server].capacity ;
		resp.writeHead(503,{"Content-Type":"text/plain"}) ;
		resp.end("JOTPOT Server: 503 Service Unavailable.\r\nThe server was unable to reach your destination.") ;
		
	}) ;
	f.on("response",r=>{
		
		resp.writeHead(r.statusCode,r.headers) ;
		r.on("data",d=>resp.write(d)) ;
		r.on("end",_=>{
			
			resp.end() ;
			config.servers[server].load -= config.servers[server].capacity ;
			
		}) ;
		
	}) ;
	req.on("data",d=>f.write(d)) ;
	req.on("end",_=>f.end()) ;
	
}

let reqListener ;

if (config.fastmode) {
	
	reqListener = (req,resp) => {
		
		let domain = "default" ;
		let server = config.domains[domain][0] ;
		if(config.servers[server].load > 0.000001){for (let doing = 1 ; doing < config.domains[domain].length ; doing++) {
			
			if (config.servers[server].load > config.servers[config.domains[domain][doing]].load) {
				
				server = config.domains[domain][doing] ;
				if (config.servers[server].load < 0.000001) {
					
					break ;
					
				}
				
			}
			
		}}
		sendToServer(req,resp,server) ;
		
	} ;
	
}

else {

	reqListener = (req,resp) => {
		
		let domain = req.headers.host || "default" ;
		if (typeof config.domains[domain] === "undefined") {
			
			domain = "default" ;
			
		}
		let server = config.domains[domain][0] ;
		if(config.servers[server].load > 0.000001){for (let doing = 1 ; doing < config.domains[domain].length ; doing++) {
			
			if (config.servers[server].load > config.servers[config.domains[domain][doing]].load) {
				
				server = config.domains[domain][doing] ;
				if (config.servers[server].load < 0.000001) {
					
					break ;
					
				}
				
			}
			
		}}
		sendToServer(req,resp,server) ;
		
	} ;

}

for (let doing in config.listen) {
	
	http.createServer(reqListener).listen(config.listen[doing].port,config.listen[doing].host) ;
	
}