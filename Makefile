install-easyjson:
	go get github.com/mailru/easyjson && go install github.com/mailru/easyjson/...@latest

generate:
	go generate ./...
