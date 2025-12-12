#!/bin/bash

go run main.go > simrun/test.txt
cat simrun/test.txt | grep Market > simrun/mkt.txt
