#!/bin/sh
set -x
go test ./maya_zcash -c .
./maya_zcash.test
