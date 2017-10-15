#!/bin/bash

cd `dirname $0`
source _scripts/func.sh

[ -n "$WEBMON_APP_ID" ] || {

	echo "Variable WEBMON_APP_ID undefined"
	echo "Set it to your PROJECT_ID in GAE and try again"
	exit 1
}
gen_app_yaml
set -e

# set project id (grr!!!)
set -x
gcloud config set project $WEBMON_APP_ID
set +x

# check number of version (max 15)
max_versions=2
deployed_versions=$(gcloud app versions list | egrep -v '^SERVICE' | wc -l)
dp_m1=$(( deployed_versions - 1 ))
if [ "$dp_m1" -gt 1 ]
then
	last_versions=$(gcloud app versions list |
                       egrep -v '^SERVICE' | head -$dp_m1 | awk '{printf("%s ",$2)}')
	last_v="${last_versions##* }"
	[ -n "$last_versions" ]
	echo "Too many deployed versions $deployed_versions - deleteing  $dp_m1"
	set -x
	gcloud app versions delete $last_versions
	set +x
fi
set -x
gcloud app deploy
gcloud app deploy cron.yaml
# please ignore "skipped"/"copied" messages before deployment
# they have nothing common with Upload filter...
#gcloud app deploy --verbosity=info
set +x
deployed_versions=$(gcloud app versions list | egrep -v '^SERVICE' | wc -l)
dp_m1=$(( deployed_versions - 1 ))
if [ "$dp_m1" -gt 0 ]
	then
	# delete (auto scaling cannot be stopped, grr!!!)
	last_versions=$(gcloud app versions list |
                       egrep -v '^SERVICE' | head -"$dp_m1" | awk '{printf("%s ",$2)}')
	set -x
	gcloud app versions delete $last_versions
	set +x
fi
gcloud app browse

exit 0
