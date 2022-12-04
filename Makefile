dev:
	air

test:
	go test ./... -v

format:
	go fmt ./...

regenerate:
	go run -mod=mod github.com/99designs/gqlgen generate

schema_json:
	apollo schema:download --endpoint=http://localhost:8080/query schema.json

deploy:
	gcloud run deploy
