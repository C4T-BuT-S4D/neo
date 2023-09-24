#!/bin/bash -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

source "${DIR}/vars.sh"

echo "Using image: ${NEO_IMAGE}, container name: ${NEO_CONTAINER_NAME}, base dir ${NEO_BASE_DIR}"

OUT=$(docker ps --filter "name=^${NEO_CONTAINER_NAME}\$" --format "{{ .Names }}")

if [[ ! $OUT ]]; then
    echo "Starting a new container"
    docker run \
        --detach \
        --rm \
        --volume "${NEO_BASE_DIR}":/work \
        --security-opt seccomp=unconfined \
        --security-opt apparmor=unconfined \
        --cap-add=NET_ADMIN \
        --privileged \
        --network host \
        --name "${NEO_CONTAINER_NAME}" \
        --hostname "${NEO_CONTAINER_NAME}" \
        "${NEO_IMAGE}" \
        "/usr/local/bin/reaper"
else
    echo "Container already exists"
fi

# shellcheck disable=SC2068
docker exec -it "${NEO_CONTAINER_NAME}" "/entrypoint.sh" "$@"
