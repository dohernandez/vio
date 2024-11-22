#!/usr/bin/env bash

# This script exports environmental variables from .env file and runs command
# Usage: env-run.sh some command

set -o allexport
# shellcheck disable=SC1091
. .env
set +o allexport

"$@"