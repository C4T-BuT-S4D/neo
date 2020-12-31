#!/bin/bash -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

pushd "${DIR}/../"

./scripts/setup_release.sh

if [[ $* == *"--dry-run"* ]]; then
  goreleaser --skip-validate --skip-publish
else
  goreleaser --rm-dist
fi

rm -rf neo_client/
rm -rf neo_server/

popd
