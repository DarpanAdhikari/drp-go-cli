#!/usr/bin/env bash
# Cross-compile drp binaries for release (mirrors the GoReleaser matrix).
# Output goes to dist/ — same structure GoReleaser produces locally.
# Usage: ./scripts/build-release.sh [version]
set -euo pipefail

VERSION="${1:-dev}"
MODULE="github.com/DarpanAdhikari/drp-go-cli"
LDFLAGS="-s -w -X ${MODULE}/cmd.Version=${VERSION}"
MAIN="./drp"
DIST="$(cd "$(dirname "$0")/.." && pwd)/dist"

platforms=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
)

echo "Building drp ${VERSION} into ${DIST}/"
mkdir -p "${DIST}"

for platform in "${platforms[@]}"; do
  GOOS="${platform%/*}"
  GOARCH="${platform#*/}"
  binary="drp"
  [[ "${GOOS}" == "windows" ]] && binary="drp.exe"
  archive_name="drp_${VERSION}_${GOOS}_${GOARCH}"
  out_dir="${DIST}/${archive_name}"
  mkdir -p "${out_dir}"

  echo "  → ${GOOS}/${GOARCH}"
  CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" \
    go build -trimpath -ldflags "${LDFLAGS}" -o "${out_dir}/${binary}" "${MAIN}"

  # Copy supporting files
  for f in README.md LICENSE CHANGELOG.md; do
    [[ -f "${f}" ]] && cp "${f}" "${out_dir}/"
  done

  # Archive
  if [[ "${GOOS}" == "windows" ]]; then
    (cd "${DIST}" && zip -qr "${archive_name}.zip" "${archive_name}/")
  else
    (cd "${DIST}" && tar -czf "${archive_name}.tar.gz" "${archive_name}/")
  fi
  rm -rf "${out_dir}"
done

# Checksums
(cd "${DIST}" && sha256sum drp_*.tar.gz drp_*.zip 2>/dev/null > "drp_${VERSION}_checksums.txt" || true)

echo ""
echo "Done! Archives in ${DIST}/"
ls -lh "${DIST}/"
