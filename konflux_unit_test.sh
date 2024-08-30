#!/bin/bash

set -exv

export GO111MODULE="on"
go test -v -race --coverprofile=coverage.txt --covermode=atomic ./...
result=$?

if [ $result -ne 0 ]; then
  echo '====================================='
  echo '====  ✖ ERROR: UNIT TEST FAILED  ===='
  echo '====================================='
  exit 1
fi
