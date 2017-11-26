const util = require("util") ;

class serverConsole {
	constructor(error, info, warn, log, serverlog) {
		this._error = error ;
		this.error = (...args) => this._error(util.format(...args)) ;
		this._info = info ;
		this.info = (...args) => this._info(util.format(...args)) ;
		this._warn = warn ;
		this.warn = (...args) => this._warn(util.format(...args)) ;
		this._log = log ;
		this.log = (...args) => this._log(util.format(...args)) ;
		this._serverlog = serverlog ;
		this.serverlog = (...args) => this._serverlog(args.join(' ')) ;
	}
}

module.exports = {
	serverConsole
} ;
