.DEFAULT_GOAL := help

help: ## Prints help message.
	@ grep -h -E '^[a-zA-Z0-9_-].+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1m%-30s\033[0m %s\n", $$1, $$2}'


TYPE := pipeline
SERVICE :=
OS := darwin
ARCH := arm64

#test -f proto/$(SERVICE).proto && protoc --go_out=$(PWD) proto/$(SERVICE).proto &&\

compile: ## Compiles the app. Parameters: SERVICE, OS, ARCH.
	@ test -d bin || mkdir bin
	@ cd $(TYPE) &&\
 		go mod tidy &&\
    		GOOS=$(OS) GOARCH=$(ARCH) go build\
    			-a -gcflags=all="-l -B -C" -ldflags="-w -s" -o ../bin/$(SERVICE)-$(OS)-$(ARCH) ./$(SERVICE)/cmd/*.go

test: ## Runs unit tests.
	@ cd app &&\
		go mod tidy &&\
		cd $(SERVICE) &&\
		go test -v --parallel=8 .
