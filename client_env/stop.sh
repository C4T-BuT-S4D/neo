#!/bin/bash -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

source "${DIR}/vars.sh"

echo "Removing container ${NEO_CONTAINER_NAME}"
docker rm -f "${NEO_CONTAINER_NAME}"
