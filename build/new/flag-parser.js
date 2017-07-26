module.exports = _ => {
	let args = {} ;
	let inArg = false ;
	let argDone = false ;
	let cArg = new String() ;
	for (let arg of process.argv) {
		console.log(inArg, argDone, cArg) ;
		if (arg.indexOf("-") === 0) {
			if (inArg && !argDone) {
				args[cArg].push(true) ;
			}
			inArg = true ;
			argDone = false ;
			cArg = arg ;
			if (!args[arg]) {
				args[arg] = new Array() ;
			}
		} else if (inArg) {
			argDone = true ;
			args[cArg].push(arg) ;
		}
	}
	if (inArg && !argDone) {
		args[cArg].push(true) ;
	}
	return args ;
} ;
