#!/usr/bin/env bash

GIT_ROOT=$(git rev-parse --show-toplevel)

echo "===="
echo "Prow"
echo "===="
echo
echo "Please fork this repo to continue."
read -r -p "Press enter to continue, or C-c to cancel"

