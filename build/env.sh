#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"


# Set up the environment to use the workspace.
GOPATH="$workspace":"/Users/bing/"
export GOPATH


# Launch the arguments with the configured environment.
exec "$@"
