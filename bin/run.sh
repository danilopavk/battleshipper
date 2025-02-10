#!bin/sh

echo "generating templates"
templ generate
echo "copying tailwind"
npx @tailwindcss/cli -i ./static/input.css -o ./static/output.css
echo "running"
go run ./...
