all: prepare clean get test-race build-race
	@echo "*** Done!"


prepare:
	@echo "*** Create bin & pkg dirs, if not exists..."
	@mkdir -p bin
	@mkdir -p pkg

get:
	@echo "*** Resolve dependencies..."
	@go get -v ./...

test:
	@echo "*** Run tests..."
	go test -v ./rest/...

test-race:
	@echo "*** Run tests with race condition..."
	@go test --race -v ./rest/...

test-cover:
	@go test -covermode=count -coverprofile=/tmp/coverage_rest.out ./rest/...

	@rm -f /tmp/movie_service_coverage.out
	@echo "mode: count" > /tmp/movie_service_coverage.out

	@cat /tmp/coverage_rest.out | tail -n +2  >> /tmp/movie_service_coverage.out
	@rm /tmp/coverage_rest.out

	@go tool cover -html=/tmp/movie_service_coverage.out

build:
	@echo "*** Build project..."
	@go build -v -o bin/movie-service main.go defaults.go

build-race:
	@echo "*** Build project with race condition..."
	@go build --race -v -o bin/movie-service-race main.go defaults.go

clean-bin:
	@echo "*** Clean up bin/ directory..."
	@rm -rf bin/*

clean-pkg:
	@echo "*** Clean up pkg/ directory..."
	@rm -rf pkg/*

clean: clean-bin clean-pkg
