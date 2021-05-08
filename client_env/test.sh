#!/bin/bash

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"
source "${DIR}/vars.env"

NEED_COMMANDS=(curl wget dig nc file nslookup ifconfig python3 pip3)
NEED_PACKAGES=(pymongo pymysql psycopg2 redis z3 secrets checklib requests pwn numpy bs4 hashpumpy dnslib regex lxml gmpy2 sympy)

for cmd in "${NEED_COMMANDS[@]}"; do
  echo "Checking for command ${cmd}..."
  if docker run --entrypoint /bin/bash --rm "${IMAGE}" which "${cmd}" >/dev/null; then
    echo "ok"
  else
    echo "Command ${cmd} not found in image"
    exit 1
  fi
done

for pkg in "${NEED_PACKAGES[@]}"; do
  echo "Checking for package ${pkg}..."
  if docker run --entrypoint "/bin/bash" --rm "${IMAGE}" -c "python3 -c 'import ${pkg}'"; then
    echo "ok"
  else
    echo "Package ${pkg} not found in image"
    exit 1
  fi
done
