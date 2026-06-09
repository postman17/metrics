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

migrate:
	migrate -database "postgres://metrics_user:metrics_user_password@localhost:5432/metrics?sslmode=disable" -path ./migrations up
