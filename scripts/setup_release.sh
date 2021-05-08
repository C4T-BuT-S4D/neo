#!/bin/bash -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

pushd "${DIR}/../"

# clean up
rm -rf dist neo_client neo_server

# setup client release
mkdir -p neo_client
cp configs/client/config.yml neo_client/config.yml
mkdir -p neo_client/exploits
touch neo_client/exploits/.keep
cp client_env/requirements.txt neo_client/
cp client_env/start.sh neo_client/
cp README.md neo_client/

# setup server release
mkdir -p neo_server/data
touch neo_server/data/.keep
cp configs/server/config.yml neo_server/config.yml
cp README.md neo_server/

popd
