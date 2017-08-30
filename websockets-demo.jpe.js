//Only run if we are a worker
if (!server.isMaster) {
	const {Buffer} = require("buffer") ;
	server.handleWebSocket("localhost:8080/ws.example", req=>{
		//Accept the session (in production, you should always check origins etc.)
		req.accept((err, resp) => {
			if (err) {
				//Throw if there was an error
				throw err ;
			}
			req.on("text", text => {
				//When we get a text message, log it out.
				console.info("Text message:", text) ;
				//And send a message back, saying "Your message: $TEXT"
				//Write the first frame with the first part of the message
				resp.write("Your message: ") ;
				//Then write the end frame with the second part of the message
				resp.end(text) ;
				
				//This could also be done as
				// resp.end("Your message: "+text)
				//Which would be more efficient as it would be sent as a single frame.
				//However 2 frames have been sent for purposes of example.
			}) ;
			req.on("binary", buff => {
				//When we get a binary message, log it as hex.
				console.info("Binary message:", buff.toString("hex")) ;
			}) ;
			
			//Send a text message saying "Hello!"
			resp.end("Hello!") ;
			setTimeout(() => {
				//After 5 seconds, send a binary frame with the data 010203.
				resp.end(Buffer.from([1,2,3])) ;
			}, 5000) ;
		}) ;
		//Must return true to say we've handled it, if we don't, the server ends the socket for safety purposes.
		return true ;
	}) ;
}

/*
//Run in browser to test (with the server open on localhost:8080):

//Create WebSocket connection.
const socket = new WebSocket("ws://localhost:8080/ws.example") ;

socket.addEventListener("open", () => {
	//Send a message when the connection is made
	socket.send("Hello from the client!") ;
	socket.send(new Blob([123,12,1])) ;
	setTimeout(() => {
		//Send another message 3 seconds later.
		socket.send("3 seconds later.") ;
	}, 3000) ;
}) ;

//Add an event listener for a message
socket.addEventListener("message", event => {
	if (event.data.constructor.name === "Blob") {
		//If it is a binary message, Create an ArrayBuffer from the result
		let reader = new FileReader() ;
		reader.readAsArrayBuffer(event.data) ;
		reader.addEventListener("loadend", () => {
			let arr = new Uint8Array(reader.result) ;
			//Log the data
			console.log("Binary message from server:", arr.join(",")) ;
		}) ;
	} else {
		//If it's a text message, just log the text
		console.log("Text message from server:", event.data) ;
	}
}) ;

//Log that we are closed if the connection is closed.
socket.addEventListener("close", () => {
	console.log("Connection to server closed!") ;
}) ;

/**/