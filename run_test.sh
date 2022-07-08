#!/bin/bash
cd tests && go test  -coverprofile=c.out -coverpkg=../... $@
