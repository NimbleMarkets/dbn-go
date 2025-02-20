#!/bin/bash

set -e

if [ -z "$DATABENTO_API_KEY" ]; then
    echo "DATABENTO_API_KEY must be set"
    exit 1
fi

DBN_GO_HIST=${DBN_GO_HIST:-dbn-go-hist}

DATASET=${DATASET:-DBEQ.BASIC}

SCHEMA=${SCHEMA:-ohlcv-1h}

SYMBOLS=${SYMBOLS:-MSFT}

START_DATE=${START_DATE:-"2024-03-01"}
END_DATE=${END_DATE:-"2024-03-05"}

echo "$ dbn-go-hist datasets"
"${DBN_GO_HIST}" datasets
echo

echo "$ dbn-go-hist publishers | head -5"
"${DBN_GO_HIST}" publishers | head -5
echo

echo "$ dbn-go-hist schemas -d ${DATASET}"
"${DBN_GO_HIST}" schemas -d "${DATASET}"
echo

echo "$ dbn-go-hist fields -s ${SCHEMA}"
"${DBN_GO_HIST}" fields -s "${SCHEMA}"
echo

echo "$ dbn-go-hist unit-prices -d ${DATASET}"
"${DBN_GO_HIST}" unit-prices -d "${DATASET}"
echo

echo "$ dbn-go-hist dataset-condition -d ${DATASET} -t ${START_DATE} -e ${END_DATE} ${SYMBOLS}"
"${DBN_GO_HIST}" dataset-condition -d "${DATASET}" -t ${START_DATE} -e "${END_DATE}" "${SYMBOLS}"
echo

echo "$ dbn-go-hist dataset-range -d ${DATASET}"
"${DBN_GO_HIST}" dataset-range -d "${DATASET}"
echo

echo "$ dbn-go-hist cost -d ${DATASET} -s ${SCHEMA} -t ${START_DATE} -e ${END_DATE} ${SYMBOLS}"
"${DBN_GO_HIST}" cost -d "${DATASET}" -s "${SCHEMA}" -t "${START_DATE}" -e "${END_DATE}" "${SYMBOLS}"
echo

echo "$ dbn-go-hist jobs"
"${DBN_GO_HIST}" jobs
echo
