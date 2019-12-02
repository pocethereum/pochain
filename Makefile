# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: poc poc-cross swarm evm all test clean
.PHONY: poc-linux poc-linux-386 poc-linux-amd64 poc-linux-mips64 poc-linux-mips64le
.PHONY: poc-linux-arm poc-linux-arm-5 poc-linux-arm-6 poc-linux-arm-7 poc-linux-arm64
.PHONY: poc-darwin poc-darwin-386 poc-darwin-amd64
.PHONY: poc-windows poc-windows-386 poc-windows-amd64

GOBIN = $(shell pwd)/build/bin
GO ?= latest

poc:
	build/env.sh go run build/ci.go install ./cmd/poc
	@echo "Done building."
	@echo "Run \"$(GOBIN)/poc\" to launch poc."

all:
	build/env.sh go run build/ci.go install

test: all
	build/env.sh go run build/ci.go test

lint: ## Run linters.
	build/env.sh go run build/ci.go lint

clean:
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

swarm:
	build/env.sh go run build/ci.go install ./cmd/swarm
	@echo "Done building."
	@echo "Run \"$(GOBIN)/swarm\" to launch swarm."

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

poc-cross: poc-linux poc-darwin poc-windows
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/poc-*

poc-linux: poc-linux-386 poc-linux-amd64 poc-linux-arm poc-linux-mips64 poc-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/poc-linux-*

poc-linux-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/poc
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/poc-linux-* | grep 386

poc-linux-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/poc
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/poc-linux-* | grep amd64

poc-linux-arm: poc-linux-arm-5 poc-linux-arm-6 poc-linux-arm-7 poc-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/poc-linux-* | grep arm

poc-linux-arm-5:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/poc
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/poc-linux-* | grep arm-5

poc-linux-arm-6:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/poc
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/poc-linux-* | grep arm-6

poc-linux-arm-7:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/poc
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/poc-linux-* | grep arm-7

poc-linux-arm64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/poc
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/poc-linux-* | grep arm64

poc-linux-mips:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/poc
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/poc-linux-* | grep mips

poc-linux-mipsle:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/poc
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/poc-linux-* | grep mipsle

poc-linux-mips64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/poc
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/poc-linux-* | grep mips64

poc-linux-mips64le:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/poc
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/poc-linux-* | grep mips64le

poc-darwin: poc-darwin-386 poc-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/poc-darwin-*

poc-darwin-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/poc
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/poc-darwin-* | grep 386

poc-darwin-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/poc
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/poc-darwin-* | grep amd64

poc-windows: poc-windows-386 poc-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/poc-windows-*

poc-windows-386:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/poc
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/poc-windows-* | grep 386

poc-windows-amd64:
	build/env.sh go run build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/poc
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/poc-windows-* | grep amd64
