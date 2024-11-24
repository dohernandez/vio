#!/usr/bin/env bash

# This script exports environmental variables from .env file and runs a command
# Usage: env-run.sh some command

# Check if ENV_PATH is not set or empty, and set a default value .env
[ -z "$ENV_PATH" ] && ENV_PATH=".env"

# Export variables from the .env file
set -o allexport
# shellcheck disable=SC1091
. "$ENV_PATH"
set +o allexport

# Run the provided command
"$@"