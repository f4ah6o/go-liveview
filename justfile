default:
    @just --list

# Go
build:
    go build ./...

test:
    go test ./...

test-pbt:
    go test -v ./tests/properties/...

generate:
    templ generate

# JavaScript
js-build:
    cd js && npm run build

js-test:
    cd js && npm test

js-publish:
    cd js && npm publish --access public

# Development
dev:
    air

# All
ci: generate build test js-build js-test

# Sample Applications
run-counter:
    @echo "Starting Counter server on http://localhost:8080"
    @echo "Features: Increment/Decrement buttons with real-time updates"
    @go run examples/counter/cmd/main.go

run-chat:
    @echo "Starting Chat server on http://localhost:8080"
    @echo "Features: Real-time messaging with PubSub"
    @echo "Open multiple browsers to test multi-client sync"
    @go run examples/chat/cmd/main.go

run-form:
    @echo "Starting Form server on http://localhost:8080"
    @echo "Features: Form validation with phx-change and phx-submit"
    @go run examples/form/cmd/main.go

# Build all samples
build-samples:
    go build -o bin/counter examples/counter/cmd/main.go
    go build -o bin/chat examples/chat/cmd/main.go
    go build -o bin/form examples/form/cmd/main.go
    @echo "Samples built in ./bin/"
