#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-${0}}")" && pwd)"

# Path to the bundled udocker CLI (as distributed in bin/udocker).
UDOCKER_BIN="${SCRIPT_DIR}/bin/udocker/udocker"

# Ensure a usable 'python' on PATH for udocker's shebang.
# On macOS 등에서는 'python3'만 있고 'python'이 없을 수 있으므로,
# 필요할 때만 로컬 shim 을 추가합니다.
if ! command -v python >/dev/null 2>&1; then
  if command -v python3 >/dev/null 2>&1; then
    SHIM_DIR="${SCRIPT_DIR}/.shim"
    mkdir -p "${SHIM_DIR}"
    if [ ! -x "${SHIM_DIR}/python" ]; then
      cat > "${SHIM_DIR}/python" <<'EOF'
#!/usr/bin/env bash
exec python3 "$@"
EOF
      chmod +x "${SHIM_DIR}/python"
    fi
    export PATH="${SHIM_DIR}:${PATH}"
  fi
fi

if [ ! -x "${UDOCKER_BIN}" ]; then
  echo "ERROR: udocker CLI not found or not executable at: ${UDOCKER_BIN}" >&2
  echo "       Make sure the repository was cloned correctly and that python3 is installed." >&2
  exit 1
fi

# Use a repository directory local to this project, so we don't pollute any global udocker state.
export UDOCKER_DIR="${SCRIPT_DIR}/.udocker"

# Default OCI image name; can be overridden via env vars.
IMAGE_NAME="${FPSB_IMAGE_NAME:-fpsb-jvm:latest}"

# Ensure the image is present; if not, print a clear hint and exit.
if ! "${UDOCKER_BIN}" images 2>/dev/null | awk '{print $1":"$2}' | grep -q "^${IMAGE_NAME}\$"; then
  echo "[run.sh] Image '${IMAGE_NAME}' not found in local udocker repo at ${UDOCKER_DIR}." >&2
  echo "[run.sh] To import it from a docker-save tarball, run (once):" >&2
  echo "         UDOCKER_DIR=\"${UDOCKER_DIR}\" bin/udocker/udocker load -i download/fpsb-jvm-latest.tar" >&2
  echo "[run.sh] 또는 udocker pull/create 등을 사용해 '${IMAGE_NAME}' 이미지를 준비해 주세요." >&2
  exit 1
fi

# Host plugin directory to expose to the container.
HOST_PLUGINS_DIR="${FPSB_PLUGINS_DIR:-${SCRIPT_DIR}/plugins}"
mkdir -p "${HOST_PLUGINS_DIR}"

CONTAINER_PLUGINS_DIR="/opt/app/plugins"

echo "[run.sh] Using host plugin directory: ${HOST_PLUGINS_DIR}"
echo "[run.sh] Running image '${IMAGE_NAME}' via udocker..."

exec "${UDOCKER_BIN}" run \
  --rm \
  -v "${HOST_PLUGINS_DIR}:${CONTAINER_PLUGINS_DIR}" \
  -e "FPSB_PLUGINS_DIR=${CONTAINER_PLUGINS_DIR}" \
  "${IMAGE_NAME}" "$@"
