#!/bin/bash
source goprj.preamble.shlib

rm -rf "$GOBIN" "$GOPATH"/{bin,pkg,tmp}
rm -rf .pipeline.demo
