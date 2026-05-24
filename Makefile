tools:
	go get -tool github.com/mailru/easyjson/easyjson@v0.9.2

generate:
	go generate ./...
