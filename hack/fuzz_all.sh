#!/bin/bash

# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

## Modified from Ethan Davidson
## https://stackoverflow.com/questions/71584005/
## how-to-run-multi-fuzz-test-cases-wirtten-in-one-source-file-with-go1-18

# clean all subprocesses on ctl-c

trap "trap - SIGTERM && kill -- -$$ || true" SIGINT SIGTERM

set -e

fuzzTime="${1:-1}"  # read from argument list or fallback to default - 1 minute

repoRoot="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ -z "${KUBEBUILDER_ASSETS:-}" ]]; then
    autoAssets="$(find "${repoRoot}/bin/k8s" -maxdepth 2 -type f -name etcd -printf '%h\n' 2>/dev/null | head -n 1)"
    if [[ -n "${autoAssets}" ]]; then
        KUBEBUILDER_ASSETS="${autoAssets}"
        export KUBEBUILDER_ASSETS
        echo "KUBEBUILDER_ASSETS was not set, using auto-detected assets at ${KUBEBUILDER_ASSETS}"
    else
        if [[ ! -x "${repoRoot}/bin/setup-envtest" ]]; then
            echo "KUBEBUILDER_ASSETS is not set and local envtest assets were not found under ${repoRoot}/bin/k8s"
            echo "setup-envtest binary was not found at ${repoRoot}/bin/setup-envtest"
            echo "Run 'make envtest' and then rerun fuzzing, or use 'make fuzz FUZZTIME=${fuzzTime}'"
            exit 1
        fi

        echo "KUBEBUILDER_ASSETS was not set and local envtest assets were not found. Downloading envtest assets..."
        KUBEBUILDER_ASSETS="$(${repoRoot}/bin/setup-envtest use 1.31.0 --bin-dir "${repoRoot}/bin" -p path)"
        export KUBEBUILDER_ASSETS
        if [[ -z "${KUBEBUILDER_ASSETS}" ]]; then
            echo "Failed to resolve envtest assets path via setup-envtest"
            exit 1
        fi
        echo "Using downloaded envtest assets at ${KUBEBUILDER_ASSETS}"
    fi
elif [[ "${KUBEBUILDER_ASSETS}" != /* ]]; then
    KUBEBUILDER_ASSETS="${repoRoot}/${KUBEBUILDER_ASSETS}"
    export KUBEBUILDER_ASSETS
fi

files=$(grep -r --include='**_test.go' --files-with-matches 'func Fuzz' pkg internal)

cat <<EOF
Starting fuzzing tests.
    One test timeout: $fuzzTime
    Files:
$files
EOF

go clean --cache

for file in ${files}
do
    funcs="$(grep -oP 'func \K(Fuzz\w*)' "$file")"
    for func in ${funcs}
    do
        {
            echo "Fuzzing $func in $file"
            parentDir="$(dirname "$file")"
            go test "./$parentDir" -fuzz="$func" -run="$func" -fuzztime="${fuzzTime}" -v -parallel 4
        } &
    done
done

for job in `jobs -p`
do
    echo "Waiting for PID $job to finish"
    wait $job
done
