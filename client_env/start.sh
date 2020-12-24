#!/bin/bash

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"
ARGS="$*"

source "${DIR}/vars.env"
OUT=$(docker ps -a | grep "${CONTAINER_NAME}")

set -e

if [[ $OUT ]]; then
  docker exec -it "$NAME" "$ARGS"
else
  docker run -v "${DIR}":/work -w /work \
    --security-opt seccomp=unconfined \
    --security-opt apparmor=unconfined \
    --cap-add=NET_ADMIN \
    --privileged \
    --name "${CONTAINER_NAME}" \
    --hostname "${CONTAINER_NAME}" \
    -it --rm \
    "${IMAGE}" \
    "${ARGS}"
fi
