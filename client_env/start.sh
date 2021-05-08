#!/bin/bash -e

IMAGE=c4tbuts4d/neo_env:latest
CONTAINER_NAME=neo_env

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"
COMMAND="$*"

OUT=$(docker ps --filter "name=${CONTAINER_NAME}" --format "{{ .Names }}")

if [[ $OUT ]]; then
  # shellcheck disable=SC2068
  docker exec -it "${CONTAINER_NAME}" ${COMMAND[@]}
else
  docker run -it \
    --rm \
    --volume "${DIR}":/work \
    --security-opt seccomp=unconfined \
    --security-opt apparmor=unconfined \
    --cap-add=NET_ADMIN \
    --privileged \
    --name "${CONTAINER_NAME}" \
    --hostname "${CONTAINER_NAME}" \
    "${IMAGE}" \
    "${COMMAND}"
fi
