#!/bin/bash

set -xe

source utils.sh

sleep 120

# After 120s from the start, block height will be 24 or 25
# check height from nodes
for ((port=2821;port<=2823;port++)); do
    height=$(getBlockHeight ${port})
    if [ "$height" != "24" ] && [ "$height" != "25" ] ; then
        die "Expected height to be 24 or 25, not ${height}"
    fi
done