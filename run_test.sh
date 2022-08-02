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
      cd tests && go test  -coverprofile=c.out -coverpkg=../...
     ;;
  "0.1")
     cd tests && go test  -coverprofile=c.out -coverpkg=../...  && go tool cover -html=c.out
     ;;
  "1.0")
     cd tests && go test  -coverprofile=c.out -coverpkg=../...  -v
     ;;
  "1.1")
     cd tests && go test  -coverprofile=c.out -coverpkg=../...  -v && go tool cover -html=c.out
     ;;
esac
