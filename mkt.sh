#!/bin/bash

rm -rf simrun/*

go run main.go > simrun/test.txt
cat simrun/test.txt | grep Market > simrun/mkt.txt

cp prices.html simrun/
awk '/Market Price for wheat:/ {printf "%s,", $5}' simrun/mkt.txt | sed 's/,$//' | awk '{print "const prices = [" $0 "];"}' > simrun/prices.js
sed -i '/const prices = \[\];/r simrun/prices.js' simrun/prices.html
sed -i '/const prices = \[\];/d' simrun/prices.html
