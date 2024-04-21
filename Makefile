#
# Makefile
#
# Scott Higgins 9/5/18
#

# check if golang exists and give message telling user to
# install if missing
ifeq (, $(shell which go))
$(info "No 'golang' in PATH, please install golang to build")
$(info "  linux      : sudo apt-get install golang")
$(info "  macosx     : brew install golang")
$(info "  download   : https://golang.org/download/")
$(error "unable to complete build")
endif


.DEFAULT_GOAL := standard

BINARY_NAME ?= main
DOCKER_NAME ?= service-prov
TEST_DOCKER_NAME ?= service-prov-test

# Get the GitLab version
RELEASE := $(shell git describe --long | cut -d'-' -f1)
CHANGENUM := $(shell git describe --long | cut -d'-' -f2)
FORMATNUM := $(shell printf "%05d\n" $(CHANGENUM))

AIRBIN := air
ifeq ($(shell uname -s),Linux)
   AIRBIN = air-linux
endif


standard:
	$(info Building binary with name: $(BINARY_NAME))
	@go build -o $(BINARY_NAME) .

watch:
	$(info Using ./$(AIRBIN) utility to watch and live-reload)
	@./$(AIRBIN)

linux:
	$(info Building binary with name: $(BINARY_NAME)-linux)
	@GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux .

rpm:
	$(info Building binary with name: $(BINARY_NAME)-linux)
	@GOOS=linux GOARCH=amd64 go build -ldflags=-linkmode=external -o $(BINARY_NAME)-linux .

linuxpackage: linux
	$(info Creating tar file that can be used with linux distros)
	@tar -cvzf svc-prov-linux.tar.gz $(BINARY_NAME)-linux certificates config RELEASE documentation public

tests:
	$(info Running golang unit tests)
	@go test ./lib
	@go test ./controllers
	@go test ./common

gocommon:
	$(info Synchronizing go-common code)
	@git fetch go-common
	@git subtree pull -P common/ --squash go-common master

docs:
	$(info Building Swagger documentation)
	@swagger-cli bundle -t yaml -o ./public/doc/service-prov.yml ./documentation/service-prov.yml

docker:
	$(info Building docker image with name: $(DOCKER_NAME))
	@docker build --build-arg gitlab_version=$(RELEASE).$(FORMATNUM) -t $(DOCKER_NAME) -f Dockerfile .

dockertest:
	$(info Building docker test image with name: $(TEST_DOCKER_NAME))
	@docker build -t $(TEST_DOCKER_NAME) -f Dockerfile.test .

clean:
	@rm -f $(BINARY_NAME)
	@rm -f $(BINARY_NAME)-linux
	@rm -f svc-prov-linux.tar.gz
