export NEO_BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"

if [[ -z "${NEO_VERSION}" && -f "${NEO_BASE_DIR}/.version" ]]; then
  NEO_VERSION=$(xargs <"${NEO_BASE_DIR}/.version")
fi

if [[ -z "${NEO_VERSION}" ]]; then
  NEO_VERSION=latest
fi

export NEO_VERSION
export NEO_CONTAINER_NAME="neo-${VERSION}"
export NEO_IMAGE="ghcr.io/c4t-but-s4d/${IMAGE_NAME}:${NEO_VERSION}"
