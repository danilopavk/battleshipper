#!bin/sh

echo "generating templates"
templ generate
echo "Formatting"
go fmt ./...
echo "Formatting imports"
goimports -l -w .
echo "Running linters"
golangci-lint run ./...
echo "Runing vuln check"
govulncheck ./...
echo "Running tests"
go test ./... -v
echo "All done"
