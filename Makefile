tools:
	go install tool

generate:
	go generate ./...

tests-cover:
	go test -cover ./...

tests:
	go test ./...

db:
	docker compose up -d
