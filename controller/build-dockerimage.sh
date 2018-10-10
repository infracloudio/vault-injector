#!/bin/bash
cd ./cmd/controller && go build && cd -
docker build -t vault-injector:dev .
