#!/bin/bash

# reformats my source
gofmt=/opt/gae/google-cloud-sdk/platform/google_appengine/goroot-1.8/bin/gofmt
cd `dirname $0`
cd ..
find ./ -name \*.go | xargs $gofmt -d
echo -n "Reformat source [y/N]?"
read ans
case "$ans" in
  y*)
    find ./ -name \*.go | xargs $gofmt -w
    echo "Reformatted"
    exit 0
    ;;
  *)
    echo "Canceled"
    exit 1
   ;; 
esac
