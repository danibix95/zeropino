
DEFAULT_BUILD_FLAGS = -ldflags="-w -s"
DEFAULT_TEST_FLAGS = -timeout 30s -failfast
ADVANCED_BUILD_FLAGS = -trimpath -buildmode=pie -mod=readonly -modcacherw

.PHONY: test clean

all: clean compile

compile: test
	@mkdir -p bin
	GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build ${DEFAULT_BUILD_FLAGS}  ./...

test:
	@go test ${DEFAULT_TEST_FLAGS} -race ./...

cover:
	@go test ${DEFAULT_TEST_FLAGS} -cover -coverprofile=coverage.out ./...

bench:
	@go test ${DEFAULT_TEST_FLAGS} -bench=. ./...

show-coverage: cover
	@go tool cover -html=coverage.out

clean:
	@rm -rf bin/
