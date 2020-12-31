#!/bin/bash -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"
DIRNAME=$(basename "$DIR")

source "${DIR}/vars.env"

echo "Make sure you're logged in as ${ACCOUNT}"
pushd "${DIR}/../"
docker build -t "${IMAGE}" -f "${DIRNAME}/Dockerfile" .
popd

"${DIR}/test.sh"

if [[ $* == *'--push'* ]]; then
  docker push "${IMAGE}"
fi
