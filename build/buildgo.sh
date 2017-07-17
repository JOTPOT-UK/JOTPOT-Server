if [ ! -d "GoRelease" ]
then
	mkdir GoRelease
fi
export GOOS="windows"
export GOARCH="amd64"
go build -o GoRelease/lb-win-x64.exe lb.go
export GOARCH="386"
go build -o GoRelease/lb-win-x86.exe lb.go
export GOOS="linux"
go build -o GoRelease/lb-linux-x86 lb.go
export GOARCH="amd64"
go build -o GoRelease/lb-linux-x64 lb.go
export GOARCH="arm"
go build -o GoRelease/lb-linux-arm lb.go
export GOARCH="arm64"
go build -o GoRelease/lb-linux-arm64 lb.go
