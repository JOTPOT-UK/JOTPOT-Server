//Adds Last-Modified header, returns true if the response should be fully served, so 200, or false if it should be a 304
function processCacheHeaders(req, resp, lastMod) {
	//Set header
	resp.setHeader("Last-Modified", lastMod.toGMTString()) ;
	//If we are conditional
	if (req.headers["if-modified-since"]) {
		//Get the time from the header and only carry on if we could parse it
		let imf = (new Date(req.headers["if-modified-since"])).getTime() ;
		if (!isNaN(imf)) {
			//If it is newer or the same age as the mod time of the file (rounded down to the nearest second), then we can send a 304 (and thus no body)
			if (imf >= 1000*Math.floor(lastMod.getTime()/1000)) {
				resp.sendBody = false ;
				return false ;
			}
		}
	}
	return true ;
}

module.exports = {
	processCacheHeaders
} ;
