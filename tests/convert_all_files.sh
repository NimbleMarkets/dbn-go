#!/bin/bash

# Get the directory where the script is located
# So very LLM...
SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# Iterate over directories
for df in "$SCRIPT_DIR"/data/*.dbn.zst; do
    echo "$df" 
    "$SCRIPT_DIR"/../bin/dbn-go-file metadata -v "$df" > /dev/null
    "$SCRIPT_DIR"/../bin/dbn-go-file json -v "$df" > /dev/null
    # "$SCRIPT_DIR"/../bin/dbn-go-file parquet -v "$df" > /dev/null
done
