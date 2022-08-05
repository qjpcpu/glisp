#!/bin/bash
VERBOSE=0
COVERAGE=0
for var in "$@"; do
  case $var in
          "-v")
                  VERBOSE=1
                  ;;
          "-c")
                  COVERAGE=1
                  ;;
  esac
done

case "${VERBOSE}.${COVERAGE}" in
  "0.0")
      cd tests && go test  -coverprofile=c.out -coverpkg=../... | tee test.out
     ;;
  "0.1")
     cd tests && go test  -coverprofile=c.out -coverpkg=../...  | tee test.out && go tool cover -html=c.out
     ;;
  "1.0")
     cd tests && go test  -coverprofile=c.out -coverpkg=../...  -v | tee test.out
     ;;
  "1.1")
     cd tests && go test  -coverprofile=c.out -coverpkg=../...  -v | tee test.out && go tool cover -html=c.out
     ;;
esac

COVERAGE=`grep coverage test.out |grep -oE '[0-9]+[^%]*'`
curl -s "https://img.shields.io/badge/coverage-$COVERAGE-green" > codcov.svg
rm -f test.out
