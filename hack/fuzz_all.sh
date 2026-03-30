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
fuzzFunc="${2:-}"   # optional: only run fuzz functions whose name contains this string
cleanFuzz=false     # set to true via --clean to remove saved fuzzer-found corpus before running
sequential=false    # set to true via --sequential to run fuzz tests one at a time

# Parse remaining arguments
for arg in "$@"; do
    case "$arg" in
        --clean)      cleanFuzz=true ;;
        --sequential) sequential=true ;;
    esac
done

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
    Function filter:  ${fuzzFunc:-(all)}
    Mode:             ${sequential:+sequential}${sequential:-parallel}
    Files:
$files
EOF

go clean --cache

if [[ "${cleanFuzz}" == "true" ]]; then
    echo "Cleaning fuzzer-found corpus (testdata/fuzz/) directories..."
    find "${repoRoot}/pkg" "${repoRoot}/internal" -type d -name fuzz -path '*/testdata/fuzz' \
        -exec rm -rf {} + 2>/dev/null || true
fi

fuzzFailed=false
for file in ${files}
do
    funcs="$(grep -oP 'func \K(Fuzz\w*)' "$file")"
    for func in ${funcs}
    do
        # Skip functions that don't match the optional filter
        if [[ -n "${fuzzFunc}" && "${func}" != *"${fuzzFunc}"* ]]; then
            continue
        fi
        if [[ "${sequential}" == "true" ]]; then
            echo "Fuzzing $func in $file"
            parentDir="$(dirname "$file")"
            set +e
            timeout "${fuzzTime}" go test "./$parentDir" -fuzz="$func" -run="$func" -fuzztime="${fuzzTime}" -v -parallel 4
            exit_code=$?
            set -e
            if [ $exit_code -ne 0 ] && [ $exit_code -ne 124 ]; then
                fuzzFailed=true
            fi
        else
            {
                echo "Fuzzing $func in $file"
                parentDir="$(dirname "$file")"
                set +e
                timeout "${fuzzTime}" go test "./$parentDir" -fuzz="$func" -run="$func" -fuzztime="${fuzzTime}" -v -parallel 4
                exit_code=$?
                set -e
                if [ $exit_code -ne 0 ] && [ $exit_code -ne 124 ]; then
                    exit 1
                fi
                exit 0
            } &
        fi
    done
done

if [[ "${sequential}" != "true" ]]; then
    for job in `jobs -p`
    do
        echo "Waiting for PID $job to finish"
        wait $job || fuzzFailed=true
    done
fi

if [[ "${fuzzFailed}" == "true" ]]; then
    echo "One or more fuzz tests failed or found a failure case. Check output above."
    exit 1
fi
