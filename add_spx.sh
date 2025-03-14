#!/usr/bin/env bash
set -Eeuo pipefail

# Define the header to be added
HEADER="// Copyright (c) 2025-present unTill Software Development Group B. V. and Contributors
// SPDX-License-Identifier: Apache-2.0"

# Find source files (modify extensions as needed)
FILES=$(find . -type f \( -name "*.c" -o -name "*.cpp" -o -name "*.h" -o -name "*.hpp" -o -name "*.go" -o -name "*.js" -o -name "*.ts" -o -name "*.py" \))

# Process each file
for FILE in $FILES; do
    # Check if the file already contains the SPDX header
    if ! grep -q "SPDX-License-Identifier: Apache-2.0" "$FILE"; then
        # Add the header at the beginning of the file
        echo -e "$HEADER\n\n$(cat "$FILE")" > "$FILE"
        # Print the relative path of the modified file
        echo "Added SPDX to: $FILE"
    fi
done

echo "Processing complete."