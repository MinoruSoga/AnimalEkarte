#!/usr/bin/env node
/**
 * PreToolUse Hook: Warn on modification of an existing
 * backend/migrations/NNN_*.sql file.
 *
 * This project's migrations are additive-only once applied
 * (backend/migrations/CLAUDE.md: "適用済み migration / seed を編集すると
 * 既存 DB は checksum mismatch になる"). A lightweight hook cannot know
 * whether a specific file was actually applied to a specific database
 * (fresh checkout vs. local vs. STG), so it warns on any modification of
 * an existing migration file rather than guessing DB state. False
 * positives (editing a migration you just created and never applied) are
 * expected and harmless — this hook never blocks.
 *
 * Trigger conditions:
 *   - Edit / MultiEdit on a migrations/*.sql path (target must already exist)
 *   - Write on a migrations/*.sql path where the file already exists on
 *     disk (i.e. a full overwrite, not a brand-new migration file)
 *
 * Does NOT block (exit 0) — warning only.
 */
'use strict';

const fs = require('fs');

const MAX_STDIN = 1024 * 1024;
const MIGRATION_PATTERN = /backend\/migrations\/(\d+)_[^/]+\.sql$/;

let data = '';
process.stdin.setEncoding('utf8');

process.stdin.on('data', chunk => {
  if (data.length < MAX_STDIN) {
    data += chunk.substring(0, MAX_STDIN - data.length);
  }
});

process.stdin.on('end', () => {
  try {
    const input = JSON.parse(data);
    const toolName = String(input.tool_name || '');
    const filePath = String(input.tool_input?.file_path || '');
    const match = MIGRATION_PATTERN.exec(filePath);

    if (match) {
      const isEditLike = toolName === 'Edit' || toolName === 'MultiEdit';
      const isOverwrite = toolName === 'Write' && fs.existsSync(filePath);

      if (isEditLike || isOverwrite) {
        const num = match[1];
        process.stderr.write(
          `\n[Hook] ⚠ MIGRATION EDIT: ${num}_*.sql already exists on disk\n` +
          '[Hook] If this migration was already applied to any environment ' +
          '(local/STG), editing it will cause a checksum mismatch on next deploy.\n' +
          '[Hook] Convention: migrations are additive-only — add a new ' +
          'numbered file instead of editing this one (see backend/migrations/CLAUDE.md).\n' +
          '[Hook] Recovery if needed: docs/ops/deploy/LOCAL_DB_RESET.md (local). ' +
          'Shared STG recovery requires explicit approval and the PlanetScale runbook; ' +
          'the current workflow has no db_reset input.\n\n'
        );
      }
    }
  } catch {
    // Parse error — pass through
  }

  process.stdout.write(data);
  process.exit(0);
});
