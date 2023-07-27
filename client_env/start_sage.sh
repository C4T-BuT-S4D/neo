#!/bin/bash -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

IMAGE_NAME="neo_env_sage" /bin/bash -e "${DIR}/start_impl.sh" "$@"
