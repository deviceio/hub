$env:GOOS='windows';$env:GOARCH='amd64';go install std; go build -o $env:GOPATH\bin\deviceio-hub.windows.amd64.exe github.com/deviceio/hub 
$env:GOOS='windows';$env:GOARCH='386';go install std; go build -o $env:GOPATH\bin\deviceio-hub.windows.386.exe github.com/deviceio/hub 

$env:GOOS='linux';$env:GOARCH='amd64';go install std; go build -o $env:GOPATH\bin\deviceio-hub.linux.amd64 github.com/deviceio/hub
$env:GOOS='linux';$env:GOARCH='386';go install std; go build -o $env:GOPATH\bin\deviceio-hub.linux.386 github.com/deviceio/hub
$env:GOOS='linux';$env:GOARCH='ppc64';go install std; go build -o $env:GOPATH\bin\deviceio-hub.linux.ppc64 github.com/deviceio/hub
$env:GOOS='linux';$env:GOARCH='ppc64le';go install std; go build -o $env:GOPATH\bin\deviceio-hub.linux.ppc64le github.com/deviceio/hub
$env:GOOS='linux';$env:GOARCH='mips';go install std; go build -o $env:GOPATH\bin\deviceio-hub.linux.mips github.com/deviceio/hub
$env:GOOS='linux';$env:GOARCH='mipsle';go install std; go build -o $env:GOPATH\bin\deviceio-hub.linux.mipsle github.com/deviceio/hub
$env:GOOS='linux';$env:GOARCH='mips64';go install std; go build -o $env:GOPATH\bin\deviceio-hub.linux.mips64 github.com/deviceio/hub
$env:GOOS='linux';$env:GOARCH='mips64le';go install std; go build -o $env:GOPATH\bin\deviceio-hub.linux.mips64le github.com/deviceio/hub

$env:GOOS='darwin';$env:GOARCH='amd64';go install std; go build -o $env:GOPATH\bin\deviceio-hub.darwin.amd64 github.com/deviceio/hub
$env:GOOS='darwin';$env:GOARCH='386';go install std; go build -o $env:GOPATH\bin\deviceio-hub.darwin.386 github.com/deviceio/hub