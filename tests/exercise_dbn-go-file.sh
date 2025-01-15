#!/bin/bash

set -e

DBN_GO_FILE=${DBN_GO_FILE:-./bin/dbn-go-file}

echo "$ dbn-go-file metadata ./tests/data/test_data.ohlcv-1s.v1.dbn"
"${DBN_GO_FILE}" metadata ./tests/data/test_data.ohlcv-1s.v1.dbn
echo

echo "$ dbn-go-file json ./tests/data/test_data.ohlcv-1s.v1.dbn"
"${DBN_GO_FILE}" json ./tests/data/test_data.ohlcv-1s.v1.dbn
echo

echo "$ dbn-go-file split -v -d tests/split ./tests/data/*.dbn"
"${DBN_GO_FILE}" split -v -d tests/split ./tests/data/*.dbn
echo


echo "$ dbn-go-file parquet tests/data/test_data.ohlcv-1s.dbn && parquet-reader tests/data/test_data.ohlcv-1s.dbn.parquet > go.parquet.txt"
"${DBN_GO_FILE}" parquet tests/data/test_data.ohlcv-1s.dbn && parquet-reader tests/data/test_data.ohlcv-1s.dbn.parquet > go.parquet.txt
echo

echo "$ ./dbn_to_parquet.py tests/data/test_data.ohlcv-1s.dbn  && parquet-reader tests/data/test_data.ohlcv-1s.dbn.parquet > py.parquet.txt"
./dbn_to_parquet.py tests/data/test_data.ohlcv-1s.dbn  && parquet-reader tests/data/test_data.ohlcv-1s.dbn.parquet > py.parquet.txt
echo

echo "$ diff go.parquet.txt py.parquet.txt"
diff go.parquet.txt py.parquet.txt