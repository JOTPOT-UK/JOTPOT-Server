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
function addRule(protocols, host, pathYes, pathNo, allowAllOrigins, allowOrigins, allowMethods, allowHeaders=[], exposeHeaders=[], allowCredentials=false, priority=0, maxAge=0) {
	if (typeof protocols !== "number" || protocols > 2 || protocols < 0) {
		throw new Error("protocols must be a Number that is 0, 1 or 2") ;
	}
	if (typeof pathYes !== "object" || pathYes.constructor !== RegExp) {
		if (typeof pathYes === "string") {
			pathYes = new RegExp(pathYes, "g") ;
		} else {
			throw new Error("pathYes must be a RegExp or String") ;
		}
	}
	if (pathNo !== null && (typeof pathNo !== "object" || pathNo.constructor !== RegExp)) {
		if (typeof pathNo === "string") {
			pathNo = new RegExp(pathNo, "g") ;
		} else {
			throw new Error("pathNo must be a RegExp, String or null") ;
		}
	}
	if (typeof allowAllOrigins !== "boolean") {
		throw new Error("allowAllOrigins must be a Boolean") ;
	}
	if (typeof allowOrigins !== "object" || allowOrigins.constructor !== Array) {
		throw new Error("allowOrigins must be an Array") ;
	}
	if (typeof allowMethods !== "object" || allowMethods.constructor !== Array) {
		throw new Error("allowMethods must be an Array") ;
	}
	if (typeof allowHeaders !== "object" || allowHeaders.constructor !== Array) {
		throw new Error("allowHeaders must be an Array") ;
	}
	if (typeof exposeHeaders !== "object" || exposeHeaders.constructor !== Array) {
		throw new Error("exposeHeaders must be an Array") ;
	}
	if (typeof allowCredentials !== "boolean") {
		throw new Error("allowCredentials must be a Boolean") ;
	}
	if (typeof priority !== "number" || priority < -1 || priority > 1) {
		throw new Error("priority must be a Number that is -1, 0 or 1") ;
	}
	if (typeof maxAge !== "number") {
		throw new Error("maxAge must be a Number") ;
	}
	if (typeof host === "object" && host.constructor !== Array) {
		for (let doing in host) {
			addRule(protocols, host[doing], pathYes, pathNo, allowOrigins, allowMethods, allowHeaders, exposeHeaders, allowCredentials)
		}
		return ;
	} else if (typeof host !== "string") {
		throw new Error("host must be a String or Array of Strings.") ;
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
	if (req.method === "OPTIONS" && req.headers["access-control-request-method"] && req.headers["access-control-request-headers"]) {
		if (rule[7]) {
			resp.setHeader("Access-Control-Allow-Methods", rule[7]) ;
		}
		if (rule[8]) {
			resp.setHeader("Access-Control-Allow-Headers", rule[8]) ;
		}
		if (rule[9]) {
			resp.setHeader("Access-Control-Expose-Headers", rule[9]) ;
		}
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
	if (rule[3] !== null && req.url.path.match(rule[3])) {
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
			resp.setHeader("Vary", "Origin") ;
			allowed(rule, req, resp) ;
			return true ;
		}
		//Or check through all the expressions
		for (let exp of rule[6]) {
			if (req.headers.origin.match(exp)) {
				resp.setHeader("Access-Control-Allow-Origin", req.headers.origin) ;
				resp.setHeader("Vary", "Origin") ;
				allowed(rule, req, resp) ;
				return true ;
			}
		}
	}
	return false ;
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
