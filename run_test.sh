#!/bin/bash
cd tests && go test  -coverprofile=c.out -coverpkg=../... $@ #&& go tool cover -html=c.out
