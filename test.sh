#!/bin/sh
# Create directory for binaries
mkdir ./bin

# Setup broker
rm -f ./bin/broker
go build -o ./bin/broker ./cmd/broker

# Setup render
rm -f ./bin/render
go build -o ./bin/render ./cmd/render

# Run
PORT=3000 ./bin/broker