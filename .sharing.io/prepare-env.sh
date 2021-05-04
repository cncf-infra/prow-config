#!/usr/bin/env bash

GIT_ROOT=$(git rev-parse --show-toplevel)

echo "===="
echo "Prow"
echo "===="
echo
echo "Using prow-config on Pair will fork itself into your GitHub user account's projects"
read -r -p "Press enter to continue, or C-c to cancel"

