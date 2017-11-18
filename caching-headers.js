/*
	
	JOTPOT Server
	Version 27A-0
	
	Copyright (c) 2016-2017 Jacob O'Toole
	
	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:
	
	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.
	
	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE.
	
*/

"use strict";

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
