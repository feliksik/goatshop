server:
	go build server.go

loadgoat:
	go build -o loadgoat cmd/main.go

test:
	go test api/api_test.go
