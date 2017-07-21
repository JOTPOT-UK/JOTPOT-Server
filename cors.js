let rules1 = new Array() ;
let rules2 = new Array() ;
let rules3 = new Array() ;

let corsEnabled = false ;

//Adds a rule
//protocols: 0 for HTTP only, 1 for HTTPS only, 2 either protocol
//host: host that this applys to
//	If host is an array, it recursivaly calls itself with each host
//pathYes: RegExp that path must match
//pathNo: RegExp that path must NOT match
//	The path has to match pathYes and not match pathNo for this rule to be taken into account
//	If either pathNo is null, it will be ignored and 'nothing' will match it
//allowAllOrigins: Bool, if true, all origins will be allowed
//allowedOrigings: Array of origins that can access the resource
//	Can be string or regexp
//allowedMethods: array of allowed HTTP methods
function addRule(protocols, host, pathYes, pathNo, allowAllOrigins, allowOrigins, allowMethods, allowHeaders, exposeHeaders, allowCredentials=false, priority=0, maxAge=0) {
	if (typeof host === "object" && typeof host[0] !== "undefined") {
		for (let doing in host) {
			addRule(protocols, host[doing], pathYes, pathNo, allowOrigins, allowMethods, allowHeaders, exposeHeaders, allowCredentials)
		}
		return ;
	}
	let allowOriginsRegExp = new Array() ;
	for (let doing in allowOrigins) {
		if (allowOrigins[doing].constructor === RegExp) {
			allowOriginsRegExp.push(allowOrigins.splice(parseInt(doing), 1)) ;
		}
	}
	if (priority === 1) {
		rules1.push([protocols, host, pathYes, pathNo, allowAllOrigins, allowOrigins, allowOriginsRegExp, allowMethods.join(", "), allowHeaders.join(", "), exposeHeaders.join(", "), allowCredentials, maxAge]) ;
	} else if (priority === 0) {
		rules2.push([protocols, host, pathYes, pathNo, allowAllOrigins, allowOrigins, allowOriginsRegExp, allowMethods.join(", "), allowHeaders.join(", "), exposeHeaders.join(", "), allowCredentials, maxAge]) ;
	} else if (priority === -1) {
		rules3.push([protocols, host, pathYes, pathNo, allowAllOrigins, allowOrigins, allowOriginsRegExp, allowMethods.join(", "), allowHeaders.join(", "), exposeHeaders.join(", "), allowCredentials, maxAge]) ;
	}
	corsEnabled = true ;
}

//Add any relivent headers if we know that a specific rule includes the request, add the relivent response headers.
function allowed(rule, req, resp) {
	//Set Access-Control-Allow-Credentials if needed
	if (rule[10]) {
		resp.setHeader("Access-Control-Allow-Credentials", "true") ;
	}
	//If it is a preflight request, add the relivent headers
	if (req.method === "OPTIONS" && req.headers["Access-Control-Request-Method"] && req.headers["Access-Control-Request-Method"]) {
		resp.setHeader("Access-Control-Allow-Methods", rule[7]) ;
		resp.setHeader("Access-Control-Allow-Headers", rule[8]) ;
		resp.setHeader("Access-Control-Expose-Headers", rule[9]) ;
		resp.setHeader("Access-Control-Max-Age", String(rules[11])) ;
	}
}

function checkWith(rule, req, resp) {
	//Check protocol
	if (rule[0] < 2 && ((rule[0] === 0 && req.secure) || (rule[0] === 1 && !req.secure))) {
		return false ;
	}
	//Check host
	if (req.url.host !== rule[1]) {
		return false ;
	}
	//Check pathYes
	if (!req.url.path.match(rule[2])) {
		return false ;
	}
	//Check pathNo
	if (rule[3] !== null && req.url.match(rule[3])) {
		return false ;
	}
	//OK, so we now apply
	//If it allows anything, allow wildcard
	if (rule[4]) {
		resp.setHeader("Access-Control-Allow-Origin", '*') ;
		allowed(rule, req, resp) ;
		return true ;
	}
	//Otherwise, it might have a specific origin
	if (typeof req.headers.origin === "string") {
		//If it is a string value
		if (rule[5].indexOf(req.headers.origin) !== -1) {
			resp.setHeader("Access-Control-Allow-Origin", req.headers.origin) ;
			allowed(rule, req, resp) ;
			return true ;
		}
		//Or check through all the expressions
		for (let exp of rule[6]) {
			if (req.headers.origin.match(exp)) {
				resp.setHeader("Access-Control-Allow-Origin", req.headers.origin) ;
				allowed(rule, req, resp) ;
				return true ;
			}
		}
	}
}

//Function to set headers, goes through all the lists
function setHeaders(req, resp) {
	if (corsEnabled) {
		for (let rule of rules1) {
			if (checkWith(rule, req, resp)) {
				return ;
			}
		} for (let rule of rules2) {
			if (checkWith(rule, req, resp)) {
				return ;
			}
		} for (let rule of rules3) {
			if (checkWith(rule, req, resp)) {
				return ;
			}
		}
	}
}

module.exports = {
	addRule,
	setHeaders
} ;
