.PHONY: \
    test \
    swagger \
    docs \
    run

swagger:
	swag init \
	-g cmd/server/main.go \
	-o docs \
	--parseInternal

docs:
	rm -rf public
	mkdir -p public/docs/api/
	cp site/index.html public/docs/api/
	cp docs/swagger.yaml public/docs/api/
	cp docs/swagger.json public/docs/api/

run:
	go run cmd/server/main.go

test:
	go test ./...

test-repository:
	go test ./internal/infrastructure/db

cover-record:
	go test -coverprofile=cover.out ./internal/domain/record
	go tool cover -func=cover.out

cover-usecase:
	go test -coverprofile=cover.out ./internal/usecase
	go tool cover -func=cover.out

cover-handler:
	go test -coverprofile=cover.out ./internal/handler
	go tool cover -func=cover.out

show-cover:
	go tool cover -html=cover.out
