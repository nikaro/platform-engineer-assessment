APP=plateform-engineer-assessment
PREFIX?=/usr/local
_INSTDIR=${DESTDIR}${PREFIX}
BINDIR?=${_INSTDIR}/bin

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

.PHONY: all
all: build

.PHONY: build
## build: Build for the current target
build:
	@echo "Building..."
	env CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -o build/${APP}-${GOOS}-${GOARCH} .

.PHONY: install
## install: Install the application
install:
	@echo "Installing..."
	install -Dm755 build/${APP}-${GOOS}-${GOARCH} ${BINDIR}/${APP}

.PHONY: uninstall
## uninstall: Uninstall the application
uninstall:
	@echo "Uninstalling..."
	rm -f ${BINDIR}/${APP}
	rmdir --ignore-fail-on-non-empty ${BINDIR}

.PHONY: lint
## lint: Run linters
lint:
	@echo "Linting..."
	go vet ./...
	go fix ./...
	staticcheck ./...

.PHONY: format
## format: Runs goimports on the project
format:
	@echo "Formatting..."
	go fmt ./...

.PHONY: test
## test: Runs go test
test:
	@echo "Testing..."
	go test ./...

.PHONY: run
## run: Runs go run
run:
	go run -race main.go

.PHONY: clean
## clean: Cleans the binary
clean:
	@echo "Cleaning..."
	rm -rf build/

.PHONY: docker/build
## docker/build: Builds the Docker image
docker/build:
	docker build -t ghcr.io/nikaro/plateform-engineer-assessment -f Dockerfile .

.PHONY: part4
## part4: Runs the part 4 of the assessment
part4:
	@echo "Running commands for part 4..."
	# works on linux and macos
	cat urls.txt | tr '[:upper:]' '[:lower:]' | sed -E 's|(https?://)?(.+\.)?([a-z0-9]+\.[a-z]+)(\.)?|\3|' | sort -u
	# works on linux only (GNU sed)
	cat urls.txt | sed -E 's|(https?://)?(.+\.)?([a-zA-Z0-9]+\.[a-zA-Z]+)(\.)?|\L\3|' | sort -u
	# works on linux only and macos
	cat urls.txt | awk '{ sub(/^https?:\/\//, ""); sub(/\.$$/, ""); print tolower($$0) }' | rev | cut -d'.' -f 1,2 | rev | sort -u
	# ^(?:https?:\/\/)?(?:\w+\.)*(\w+\.\w+)\.?$

.PHONY: help
## help: Print this help message
help:
	@echo -e "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
