#!/bin/bash

# shell helper functions

gen_app_yaml () {
	# default Monitored URLs (
	url1="https://www.henryx.info"
	url2="http://henryx.atwebpages.com/"
	export MON_URLS="${MON_URLS:-$url1 $url2}"
	{
		echo "# DO NOT EDIT - Generated at `date`"
		cat app.yaml.template
		echo "env_variables:"
		for i in MON_URLS
		do
			eval val="\$$i"
			[ -n "$val" ] || {
				echo "Mandatory variable '$val' undefined" >&2
				exit 1
			}
			echo "    $i: '$val'"
		done
	} > app.yaml

}

# point GOPATH to over src/github.com/hpaluch/webmon-go
# NOTE: recent go versions (1.11) forbids relative path...
export GOPATH=$(realpath ../../../../)

