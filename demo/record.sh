#!/bin/bash
# Record ttt demo using VHS
# Usage: ./record.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(dirname "${SCRIPT_DIR}")"
IMAGES_DIR="${REPO_DIR}/images"
TTT_DIR="${HOME}/.ttt"
TASKS_FILE="${TTT_DIR}/tasks.md"
BACKUP_FILE="${TTT_DIR}/tasks.md.backup"

# Check prerequisites
if ! command -v vhs &> /dev/null; then
    echo "Error: vhs is not installed"
    echo "Install with: brew install vhs"
    exit 1
fi

if ! command -v ttt &> /dev/null; then
    echo "Error: ttt is not installed"
    exit 1
fi

# Ensure ttt directory exists
mkdir -p "${TTT_DIR}"

# Backup existing tasks.md if it exists
if [[ -f "${TASKS_FILE}" ]]; then
    echo "Backing up ${TASKS_FILE}..."
    cp "${TASKS_FILE}" "${BACKUP_FILE}"
fi

# Copy sample tasks
echo "Setting up sample tasks..."
cp "${SCRIPT_DIR}/sample-tasks.md" "${TASKS_FILE}"

# Run VHS
echo "Recording demo..."
cd "${SCRIPT_DIR}"
vhs demo.tape

# Restore original tasks.md
if [[ -f "${BACKUP_FILE}" ]]; then
    echo "Restoring original tasks.md..."
    mv "${BACKUP_FILE}" "${TASKS_FILE}"
else
    echo "Removing sample tasks.md..."
    rm -f "${TASKS_FILE}"
fi

# Move GIF to images directory
echo "Moving demo.gif to images/..."
mkdir -p "${IMAGES_DIR}"
mv "${SCRIPT_DIR}/demo.gif" "${IMAGES_DIR}/demo.gif"

echo "Done! Demo saved to ${IMAGES_DIR}/demo.gif"
