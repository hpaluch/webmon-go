#!/bin/bash

cd `dirname $0`
source _scripts/func.sh
gen_app_yaml
set -ex
go get -u google.golang.org/appengine
dev_appserver.py app.yaml

