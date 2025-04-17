#!/usr/bin/env bash

set -Eeuo pipefail  # Exit on error, treat unset vars as errors, fail on pipe errors

go-licenses check --allowed_licenses ISC,MIT,BSD-3-Clause,BSD-2-Clause,Apache-2.0 ../... 2> check_licenses.out || { echo "License check failed with exit code: $?"; cat check_licenses.out; exit 1; }