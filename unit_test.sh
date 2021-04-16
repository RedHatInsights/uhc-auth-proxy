#!/bin/bash

set -exv

export GO111MODULE="on"
go test -v -race -covermode=atomic ./...
result=$?

if [ $result != 0 ]; then
    exit $result
else
    # If your unit tests store junit xml results, you should store them in a file matching format `artifacts/junit-*.xml`
    # If you have no junit file, use the below code to create a 'dummy' result file so Jenkins will not fail
    mkdir -p $WORKSPACE/artifacts
    cat << EOF > $WORKSPACE/artifacts/junit-dummy.xml
    <testsuite tests="1">
        <testcase classname="dummy" name="dummytest"/>
    </testsuite>
EOF
fi
