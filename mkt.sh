#!/bin/bash

sedi() {
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "$@"
  else
    sed -i "$@"
  fi
}

mkdir -p simrun
rm -rf simrun/*

go run main.go > simrun/test.txt
cat simrun/test.txt | grep Market > simrun/mkt.txt

cp prices.html simrun/
awk '/Market Price for wheat:/ {printf "%s,", $5}' simrun/mkt.txt | sed 's/,$//' | awk '{print "const prices = [" $0 "];"}' > simrun/prices.js
sedi '/const prices = \[\];/r simrun/prices.js' simrun/prices.html
sedi '/const prices = \[\];/d' simrun/prices.html
