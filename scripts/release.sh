#!/bin/bash

mkdir -p client
cp configs/client/config.yml client/config.yml
mkdir -p client/exploits
touch client/exploits/.keep

mkdir -p server/data
cp configs/server/config.yml server/config.yml
touch server/data/.keep


goreleaser --rm-dist

rm -rf client/
rm -rf server/

