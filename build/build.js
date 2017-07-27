console.log("Welcome to the JOTPOT Server automated build tool.") ;
console.log("Please make sure that the command 'python' points to Python 2.7!") ;
console.log("Build will start in 5 seconds...") ;
setTimeout(()=>{

	const p = "J:\\JOTPOT-Server" ;

	const node_version = "6.11.1" ;
	const fs = require("fs") ;
	const path = require("path") ;
	const cp = require("child_process") ;
	fs.mkdirSync("build") ;
	process.chdir("build") ;
	function make(dir,...spawnArgs) {
		
		const startDir = process.cwd() ;
		fs.mkdirSync(dir) ;
		process.chdir(dir) ;
		cp.execSync(`bash -c "wget https://nodejs.org/dist/v${node_version}/node-v${node_version}.tar.gz"`) ;
		cp.execSync(`bash -c "tar -xzf node-v${node_version}.tar.gz"`) ;
		cp.execSync(`bash -c "rm node-v${node_version}.tar.gz"`) ;
		cp.execSync(`bash -c "mv ./node-v${node_version} ./build"`) ;
		process.chdir("build") ;
		fs.mkdirSync("./lib/internal/jps") ;
		fs.writeFileSync("./lib/internal/jps/run.js",sortPaths(fs.readFileSync(path.join(p,"run.js")))) ;
		//let mimes = JSON.parse(fs.readFileSync("./mimes.dat").toString()) ;
		fs.writeFileSync("./lib/internal/jps/server.js",sortPaths(fs.readFileSync(path.join(p,"server.js"))).replace(/let mimes = JSON\.parse\(fs\.readFileSync\(.*\)\.toString\(\)\) ;/g,`let mimes = ${fs.readFileSync(path.join(p,"mimes.dat"))} ;`)) ;
		fs.writeFileSync("./lib/internal/jps/externals.js",sortPaths(fs.readFileSync(path.join(p,"externals.js")))) ;
		fs.writeFileSync("./lib/internal/jps/accounts.js",sortPaths(fs.readFileSync(path.join(p,"accounts.js")))) ;
		fs.writeFileSync("./lib/internal/jps/ExtensionFileSystem.js",sortPaths(fs.readFileSync(path.join(p,"ExtensionFileSystem.js")))) ;
		fs.writeFileSync("./src/res/node.rc",fs.readFileSync("./src/res/node.rc").toString()
			.replace(/Node\.js: Server-side JavaScript/g,"JOTPOT Server")
			.replace(/VALUE "ProductName", ".*"/g,`VALUE "ProductName", "JOTPOT Server"`)
			.replace(/VALUE "CompanyName", ".*"/g,`VALUE "CompanyName", "JOTPOT"`)
			.replace(/VALUE "OriginalFilename", ".*"/g,`VALUE "OriginalFilename", "jps.exe"`)
			.replace(/VALUE "LegalCopyright", ".*"/g,`VALUE "LegalCopyright", "JOTPOT Server Copyright Jacob O'Toole. MIT Licence. Node.js Copyright Node.js contributors. MIT license."`)
			.replace(/Node\.js: Server-side JavaScript/g,"JOTPOT Server")
			.replace(/node\.ico/g,"jps.ico")) ;
		fs.writeFileSync("./src/res/jps.ico",fs.readFileSync(path.join(p,"jps.ico"))) ;
		fs.writeFileSync("./node.gyp",fs.readFileSync("./node.gyp").toString().replace(/ {6}'lib\/internal\/bootstrap_node\.js',/g,"      'lib/internal/bootstrap_node.js',\n      'lib/internal/jps/run.js',\n      'lib/internal/jps/server.js',\n      'lib/internal/jps/externals.js',\n      'lib/internal/jps/ExtensionFileSystem.js',\n      'lib/internal/jps/accounts.js',")) ;
		fs.unlinkSync("./src/res/node.ico") ;
		fs.writeFileSync("./lib/internal/bootstrap_node.js",fs.readFileSync("./lib/internal/bootstrap_node.js").toString().replace(/if \(NativeModule.exists\('_third_party_main'\)\) {/g,"if (true) {if (process.argv[1] && process.env.NODE_UNIQUE_ID) {const cluster=NativeModule.require('cluster');cluster._setupWorker();delete process.env.NODE_UNIQUE_ID;}process.nextTick(function(){NativeModule.require('internal/jps/run');})} else if (NativeModule.exists('_third_party_main')) {")) ;
		if (spawnArgs[0]){
			
			cp.spawnSync(spawnArgs[0],spawnArgs[1],{stdio:"inherit"}) ;
			
		}
		process.chdir(startDir) ;
		
	}

	const sortPaths = c => c.toString().replace(/\.\/accounts.js/gi,"internal/jps/accounts").replace(/\.\/externals.js/gi,"internal/jps/externals").replace(/\.\/server.js/gi,"internal/jps/server").replace(/\.\/run.js/gi,"internal/jps/run").replace(/\.\/ExtensionFileSystem.js/gi,"internal/jps/ExtensionFileSystem") ;

	make("source") ;
	make("win-x64",".\\vcbuild.bat",["x64"]) ;
	make("win-x86",".\\vcbuild.bat",["x86"]) ;
	make("linux-x64","bash",["-c","./configure --dest-cpu=x64 && make -j4"]) ;
	make("linux-x86","bash",["-c","./configure --dest-cpu=x86 && make -j4"]) ;
	make("linux-ia32","bash",["-c","./configure --dest-cpu=ia32 && make -j4"]) ;
	make("linux-arm","bash",["-c","./configure --dest-cpu=arm && make -j4"]) ;
	make("linux-arm64","bash",["-c","./configure --dest-cpu=arm64 && make -j4"]) ;

	try {
		
		fs.mkdirSync("../Release") ;

	}

	catch(err)  {
		
		console.warn("Error") ;
		console.warn(err) ;
		
	}

	try {
		
		fs.writeFileSync("../Release/jps-win-x64.exe",fs.readFileSync("./win-x64/build/Release/node.exe")) ;

	}

	catch(err)  {
		
		console.warn("Error") ;
		console.warn(err) ;
		
	}

	try {
		
		fs.writeFileSync("../Release/jps-win-x86.exe",fs.readFileSync("./win-x86/build/Release/node.exe")) ;

	}

	catch(err)  {
		
		console.warn("Error") ;
		console.warn(err) ;
		
	}

	try {
		
		fs.writeFileSync("../Release/jps-linux-x64",fs.readFileSync("./linux-x64/build/out/Release/node")) ;

	}

	catch(err)  {
		
		console.warn("Error") ;
		console.warn(err) ;
		
	}

	try {
		
		fs.writeFileSync("../Release/jps-linux-x86",fs.readFileSync("./linux-x86/build/out/Release/node")) ;

	}

	catch(err)  {
		
		console.warn("Error") ;
		console.warn(err) ;
		
	}

	try {
		
		fs.writeFileSync("../Release/jps-linux-ia32",fs.readFileSync("./linux-ia32/build/out/Release/node")) ;

	}

	catch(err)  {
		
		console.warn("Error") ;
		console.warn(err) ;
		
	}

	try {
		
		fs.writeFileSync("../Release/jps-linux-arm",fs.readFileSync("./linux-arm/build/out/Release/node")) ;

	}

	catch(err)  {
		
		console.warn("Error") ;
		console.warn(err) ;
		
	}

	try {
		
		fs.writeFileSync("../Release/jps-linux-arm64",fs.readFileSync("./linux-arm64/build/out/Release/node")) ;

	}

	catch(err)  {
		
		console.warn("Error") ;
		console.warn(err) ;
		
	}

},5000) ;
