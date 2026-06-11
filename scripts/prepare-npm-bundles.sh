#!/usr/bin/env bash
set -euo pipefail

tag="${1:-${GITHUB_REF_NAME:-}}"
if [[ -z "${tag}" ]]; then
  version="$(node -p "require('./package.json').version")"
  tag="v${version}"
fi

repo="${AGNES_GITHUB_REPOSITORY:-Constantine1916/agnes-cli}"
bundle_dir="${AGNES_NPM_BUNDLES_DIR:-npm-bundles}"

rm -rf "${bundle_dir}"
mkdir -p "${bundle_dir}"

gh release download "${tag}" \
  --repo "${repo}" \
  --pattern "agnes_*" \
  --pattern "SHA256SUMS" \
  --dir "${bundle_dir}"

test -s "${bundle_dir}/SHA256SUMS"
find "${bundle_dir}" -maxdepth 1 -type f -print | sort
