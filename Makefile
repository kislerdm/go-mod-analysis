.DEFAULT_GOAL := help

help: ## Prints help message.
	@ grep -h -E '^[a-zA-Z0-9_-].+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1m%-30s\033[0m %s\n", $$1, $$2}'


SERVICE :=
OS := darwin
ARCH := arm64

compile: ## Compiles the app. Parameters: SERVICE, OS, ARCH.
	@ test -d bin || mkdir bin
	@ cd app &&\
 		go mod tidy &&\
    		GOOS=$(OS) GOARCH=$(ARCH) go build\
    			-a -gcflags=all="-l -B -C" -ldflags="-w -s" -o ../bin/$(SERVICE)-$(OS)-$(ARCH) ./cmd/$(SERVICE)/*.go
