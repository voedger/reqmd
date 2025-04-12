#!/usr/bin/env bash
# =============================================================================
# add_spx.sh - Add SPDX License Headers to Source Files
# =============================================================================
# Description:
#   This script automatically adds SPDX license headers to source code files
#   that don't already have them. It searches for common source file extensions
#   and prepends the Apache-2.0 license header to each file.
#
# Usage:
#   ./add_spx.sh             # Run in the current directory
#
# Notes:
#   - Only modifies files that don't already contain the SPDX identifier
#   - Currently supported extensions: .c, .cpp, .h, .hpp, .go, .js, .ts, .py
#   - Make sure to run this script from the ./scripts directory of your project
# =============================================================================

set -Eeuo pipefail  # Exit on error, treat unset vars as errors, fail on pipe errors

# Define the header to be added
# This will be prepended to all source files that don't already have it
HEADER="// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0"

# Find source files with supported extensions
# You can modify this list to add or remove extensions as needed
FILES=$(find .. -type f \( -name "*.c" -o -name "*.cpp" -o -name "*.h" -o -name "*.hpp" -o -name "*.go" -o -name "*.js" -o -name "*.ts" -o -name "*.py" \))

# Process each file found
for FILE in $FILES; do
    # Skip files that are in testdata folders (handle both Unix and Windows paths)
    if [[ "$FILE" == *"/testdata/"* || "$FILE" == *"\\testdata\\"* ]]; then
        continue
    fi
    
    # Check if the file already contains the SPDX header to avoid duplicate headers
    if ! grep -q "SPDX-License-Identifier: Apache-2.0" "$FILE"; then
        # Create a temporary file with the header
        TMPFILE=$(mktemp)
        echo -e "$HEADER" > "$TMPFILE"
        echo "" >> "$TMPFILE"
        # Append the original file preserving its line endings
        cat "$FILE" >> "$TMPFILE"
        # Replace the original file with the temporary one
        mv "$TMPFILE" "$FILE"
        # Print the relative path of the modified file for logging purposes
        echo "Added SPDX to: $FILE"
    fi
done

echo "Processing complete."

# Exit successfully
exit 0