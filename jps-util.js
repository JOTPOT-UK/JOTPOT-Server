//Returns a promise that resolves with the contents of the given requests body.
function getData(req) {
	return new Promise((resolve, reject)=>{
		//If the request body has a specified length
		if (typeof req.headers["content-length"] !== "undefined") {
			//Parse and reject if not number
			let dLength = parseInt(req.headers["content-length"], 10) ;
			if (isNaN(dLength)) {
				reject(new Error("Content-Length header must be a number")) ;
				return ;
			}
			//Create the buffer
			let data = Buffer.alloc(dLength) ;
			let currentPos = 0 ;
			let errorSent = false ;
			req.on("data", d=>{
				//If we have more data than we can take, reject.
				if (currentPos + d.length > data.length) {
					//But only reject once
					if (errorSent) {
						return ;
					}
					errorSent = true ;
					reject(new Error("Request body was longer than the Content-Length header.")) ;
					return ;
				}
				//Write the recieved data to the output and increment the curret position by the amount of data copied.
				currentPos += d.copy(data, currentPos) ;
			}) ;
			//Resolve when the request has ended.
			req.on("end", ()=>{
				resolve(data) ;
			}) ;
			return ;
		}
		//No idea how much data, so we have to concatinate all the Buffers :(
		let data = Buffer.alloc(0) ;
		req.on("data", d=>{
			data = Buffer.concat([data, d], data.length + d.length) ;
		}) ;
		//Resolve when we are done
		req.on("end", ()=>{
			resolve(data) ;
		}) ;
	}) ;
}

module.exports = {
	getData
} ;
