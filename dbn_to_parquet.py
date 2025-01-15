#!/usr/bin/env python3
# Copyright (c) 2025 Neomantra Corp

import sys
import databento as db

def main():
    if len(sys.argv) < 2:
        sys.stderr.write("Usage: dbn_to_parquet.py <filename>\n")
        sys.exit(1)
    for filename in sys.argv[1:]:
        sys.stderr.write("Processing " + filename + "\n")
        dbn_data = db.DBNStore.from_file(filename)
        pqfilename = filename + ".parquet"
        dbn_data.to_parquet(pqfilename)
        sys.stderr.write("Created " + pqfilename + "\n")

if __name__ == "__main__":
    main()
