#!/bin/bash

mkdir -p neo_client
cp configs/client/config.yml neo_client/config.yml
mkdir -p neo_client/exploits
touch neo_client/exploits/.keep

mkdir -p neo_server/data
cp configs/server/config.yml neo_server/config.yml
touch neo_server/data/.keep


goreleaser --rm-dist

rm -rf client/
rm -rf server/

