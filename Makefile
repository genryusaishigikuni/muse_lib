build:
	@go build -o bin/muse_lib cmd/main.go

run: build
	@./bin/muse_lib

migration:
	@migrate create -ext sql -dir cmd/migrate/migrations $(filter-out $@,$(MAKECMDGOALS))

migrate-up:
	@go run cmd/migrate/main.go up

migrate-down:
	@go run cmd/migrate/main.go down


build_mock:
	@go build -o bin/mock_api mockApi/main.go

run_mock: build_mock
	@./bin/mock_api