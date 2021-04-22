#!/bin/bash

APP_NAME="uhc-auth-proxy"  # name of app-sre "application" folder this component lives in
COMPONENT_NAME="uhc-auth-proxy"  # name of app-sre "resourceTemplate" in deploy.yaml for this component
IMAGE="quay.io/cloudservices/uhc-auth-proxy"  # image location on quay

IQE_PLUGINS="uhcproxy"  # name of the IQE plugin for this app.
IQE_MARKER_EXPRESSION="uhcproxy_smoke"  # This is the value passed to pytest -m
IQE_FILTER_EXPRESSION=""  # This is the value passed to pytest -k

# Install bonfire repo/initialize
CICD_URL=https://raw.githubusercontent.com/RedHatInsights/bonfire/master/cicd
curl -s $CICD_URL/bootstrap.sh > .cicd_bootstrap.sh && source .cicd_bootstrap.sh

# Build image for testing
source $CICD_ROOT/build.sh

# Run unit tests
source $APP_ROOT/unit_test.sh

# Checkout and deploy an eph env for testing
#source $CICD_ROOT/deploy_ephemeral_env.sh

# Run smoke tests when ready
# source $CICD_ROOT/smoke_test.sh
