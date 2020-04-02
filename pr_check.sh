#!/bin/bash

set -exv

export GO111MODULE="on"
go test -v -race -covermode=atomic ./...
