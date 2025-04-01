# Document

# Project Configuration

## Auto Build

Since it depends on another private project "endorphin", you need to configure a public key on GitHub first, and then set the following configurations to pull the code:

```shell
go get github.com/yottalabsai/endorphin
go mod tidy
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o k8s-maestro main.go

# for test
go test -v ./worker/test/service/...
```
