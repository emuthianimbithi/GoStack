#!/bin/bash

# Usage: ./scripts/rename_module.sh <new_module_name>
# Example: ./scripts/rename_module.sh github.com/myuser/mynewservice

set -e

if [ -z "$1" ]; then
    echo "Usage: $0 <new_module_name>"
    exit 1
fi

NEW_MODULE=$1
OLD_MODULE=$(go list -m)

echo "Renaming module from '$OLD_MODULE' to '$NEW_MODULE'..."

# 1. Update go.mod
go mod edit -module "$NEW_MODULE"

# 2. Find all .go files and replace imports
# We use LC_ALL=C to avoid illegal byte sequence errors on some systems
find . -type f -name "*.go" -not -path "*/vendor/*" -exec sed -i '' "s|$OLD_MODULE|$NEW_MODULE|g" {} +

echo "Module renamed successfully!"
echo "Run 'go mod tidy' to refresh dependencies."
