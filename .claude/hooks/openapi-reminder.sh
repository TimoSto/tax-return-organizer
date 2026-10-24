#!/usr/bin/env bash
# PostToolUse hook: after an edit to inbound HTTP handler code, remind Claude to sync backend/api/openapi.yaml.
set -euo pipefail

file=$(jq -r '.tool_input.file_path // empty')

case "$file" in
  *_test.go) exit 0 ;;
  */backend/internal/adapters/inbound/httpadapter/*.go) ;;
  *) exit 0 ;;
esac

jq -n '{hookSpecificOutput: {hookEventName: "PostToolUse", additionalContext: "Inbound HTTP handler code changed. Before finishing, run the sync-openapi skill so backend/api/openapi.yaml matches the handlers."}}'
