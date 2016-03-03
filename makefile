all: prepare clean get test-race build-race
	@echo "*** Done!"


prepare:
	@echo "*** Create bin & pkg dirs, if not exists..."
	@mkdir -p bin
	@mkdir -p pkg

get:
	@echo "*** Resolve dependencies..."
	@go get github.com/wadey/gocovmerge
	@go get -v ./...

test:
	@echo "*** Run tests..."
	go test -v ./...

test-race:
	@echo "*** Run tests with race condition..."
	@go test --race -v ./...

test-cover:
	go list -f '{{if gt (len .TestGoFiles) 0}}"go test -covermode count -coverprofile {{.Name}}.coverprofile -coverpkg ./... {{.ImportPath}}"{{end}}' ./... | xargs -I {} bash -c {}
	@gocovmerge `ls *.coverprofile` > /tmp/movie_service_coverage.out
	@go tool cover -html=/tmp/movie_service_coverage.out

build:
	@echo "*** Build project..."
	@go build -v -o bin/movie-service

build-race:
	@echo "*** Build project with race condition..."
	@go build --race -v -o bin/movie-service-race

clean-bin:
	@echo "*** Clean up bin/ directory..."
	@rm -rf bin/*

clean-pkg:
	@echo "*** Clean up pkg/ directory..."
	@rm -rf pkg/*

clean: clean-bin clean-pkg
