$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"
go build -trimpath -ldflags="-s -w" -o bootstrap .
~\Go\Bin\build-lambda-zip.exe -o lambda-handler.zip bootstrap