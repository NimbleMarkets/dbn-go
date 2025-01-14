#!/bin/bash

set -e

DBN_GO_FILE=${DBN_GO_FILE:-dbn-go-file}

echo "$ dbn-go-file metadata ./tests/data/test_data.ohlcv-1s.v1.dbn"
"${DBN_GO_FILE}" metadata ./tests/data/test_data.ohlcv-1s.v1.dbn
echo

echo "$ dbn-go-file json ./tests/data/test_data.ohlcv-1s.v1.dbn"
"${DBN_GO_FILE}" json ./tests/data/test_data.ohlcv-1s.v1.dbn
echo

echo "$ dbn-go-file split -v -d tests/split ./tests/data/*.dbn"
"${DBN_GO_FILE}" split -v -d tests/split ./tests/data/*.dbn
echo
