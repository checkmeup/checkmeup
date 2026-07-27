#!/bin/bash
# PreToolUse guard for Edit|Write — apps/api/internal/db/ is sqlc-generated
# code (ADR-004: no ORM, sqlc generates the query layer from
# apps/api/queries/*.sql). Hand-editing it gets silently overwritten by the
# next `sqlc generate` and masks the real place to make the change.
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

if [[ "$FILE_PATH" == *"apps/api/internal/db/"* ]]; then
  echo "Blocked: $FILE_PATH is sqlc-generated (apps/api/internal/db/) — never hand-edit. Change apps/api/queries/*.sql and run 'sqlc generate' instead — see the new-migration skill." >&2
  exit 2
fi

exit 0
