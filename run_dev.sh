#!/bin/bash

cd `dirname $0`
source _scripts/func.sh
gen_app_yaml
set -ex
dev_appserver.py app.yaml

