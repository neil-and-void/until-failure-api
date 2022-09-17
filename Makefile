dev:
	go run server.go

test:
	go test ./... -v

format:
	go fmt ./...

regenerate:
	go run -mod=mod github.com/99designs/gqlgen generate
