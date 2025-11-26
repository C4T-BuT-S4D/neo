#!/bin/bash -e

for file in $(find src/proto -name '*.ts' -type f); do
  (echo "// @ts-nocheck" && cat "${file}") > .kek && mv .kek "${file}"
done