.PHONY: build cluster test

default: build

build:
	go build -o bin/swarm    github.com/pocethereum/pochain/cmd/swarm
	go build -o bin/geth     github.com/pocethereum/pochain/cmd/geth
	go build -o bin/bootnode github.com/pocethereum/pochain/cmd/bootnode

cluster: build
	scripts/boot-cluster.sh

test:
	go test -v github.com/pocethereum/pochain/swarm/...
