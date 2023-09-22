#!/bin/bash

golangci-lint run --config=.golangci.yml --fix --new --issues-exit-code 0 ./...

go build -o /bin/sb .

/bin/sb


