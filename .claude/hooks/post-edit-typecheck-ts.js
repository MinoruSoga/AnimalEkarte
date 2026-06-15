#!/usr/bin/env node
/**
 * post-edit-typecheck-ts.js
 * PostToolUse hook: TypeScript ファイル編集後に型チェックを実行
 *
 * - .ts / .tsx ファイルの編集時のみ実行
 * - Docker 経由で tsc --noEmit を実行
 * - 決定論的な型エラーは exit 2 でブロック
 * - タイムアウト等の非決定論的エラーは exit 0（警告のみ）
 * - 型エラーがある場合は stderr に出力してユーザーに通知
 */

const { execSync } = require('child_process');

let input = '';
process.stdin.on('data', (chunk) => { input += chunk; });

process.stdin.on('end', () => {
  try {
    const data = JSON.parse(input);
    const toolName = data.tool_name || '';
    const filePath = data.tool_input?.file_path || data.tool_input?.path || '';

    // .ts / .tsx ファイルの編集時のみ実行
    const isTypescriptFile = /\.(ts|tsx)$/.test(filePath);
    const isEditOrWrite = ['Edit', 'Write', 'MultiEdit'].includes(toolName);

    if (!isEditOrWrite || !isTypescriptFile) {
      process.exit(0);
    }

    // frontend/src/ 配下のファイルのみ対象
    if (!filePath.includes('frontend/src/') && !filePath.includes('frontend\\src\\')) {
      process.exit(0);
    }

    // Docker Compose が起動しているか確認（起動していない場合はスキップ）
    try {
      execSync('docker compose ps frontend --format json 2>/dev/null', {
        encoding: 'utf8',
        stdio: ['pipe', 'pipe', 'pipe'],
        timeout: 5000,
      });
    } catch {
      // Docker が起動していない場合はスキップ
      process.exit(0);
    }

    // TypeScript 型チェック実行（incremental build でキャッシュ再利用）
    const TSC_CMD =
      'docker compose exec -T frontend npx tsc --noEmit --incremental' +
      ' --tsBuildInfoFile /tmp/tsc-hook.tsbuildinfo 2>&1';
    try {
      execSync(TSC_CMD, {
        encoding: 'utf8',
        stdio: ['pipe', 'pipe', 'pipe'],
        timeout: 30000,
      });
      // 型エラーなし
      process.exit(0);
    } catch (err) {
      // タイムアウト（err.killed === true）は警告のみ、ブロックしない
      if (err.killed) {
        process.stderr.write(
          '\n⏱  TypeScript 型チェックが 30 秒でタイムアウトしました（スキップ）\n' +
          '手動確認: docker compose exec frontend npx tsc --noEmit\n\n'
        );
        process.exit(0);
      }
      const output = err.stdout || err.stderr || err.message || '';
      if (output.trim()) {
        // 型エラーあり: stderr に出力して exit 2 でブロック
        process.stderr.write(
          `\n🚫 TypeScript 型エラーが検出されました:\n${output}\n` +
          `修正方法: docker compose exec frontend npx tsc --noEmit\n\n`
        );
        process.exit(2);
      }
      // 出力なし（予期しない非ゼロ終了）: 警告のみ
      process.exit(0);
    }
  } catch {
    // JSON パースエラー等: 何もせず終了
    process.exit(0);
  }
});
