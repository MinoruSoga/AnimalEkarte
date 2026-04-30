import { describe, it, expect } from 'vitest';
import { readFileSync } from 'fs';
import { resolve } from 'path';

const sourceCode = readFileSync(
  resolve(__dirname, '../AccountingDetail.tsx'),
  'utf-8'
);

/**
 * AccountingDetail — Print Performance (#20)
 *
 * AccountingDocument は印刷時に即座に DOM へ挿入される必要がある。
 * lazy() + Suspense を使うと印刷が遅延するため、static import であることを検証する。
 * ソースコード検査による静的解析テスト。
 */
describe('AccountingDetail - Print Performance (#20)', () => {
  const sourceCode = readFileSync(
    resolve(__dirname, '../AccountingDetail.tsx'),
    'utf-8'
  );

  it('AccountingDocument が static import されている（lazy でない）', () => {
    // static import が存在する
    expect(sourceCode).toContain('import { AccountingDocument }');

    // lazy import が存在しない
    expect(sourceCode).not.toMatch(/lazy\s*\(\s*\(\)\s*=>\s*import\(.*AccountingDocument/);
  });

  it('AccountingDocument が Suspense でラップされていない', () => {
    // Suspense で AccountingDocument をラップする記述がない
    expect(sourceCode).not.toMatch(/<Suspense[\s\S]*?AccountingDocument/);
  });

  it('window.print() が実装されている', () => {
    // print 呼び出しがソース内に存在する
    expect(sourceCode).toMatch(/window\.print\(\)/);
  });

  it('AccountingDocument が JSX 内で使用されている', () => {
    // コンポーネントが JSX 内で呼び出されている
    expect(sourceCode).toMatch(/<AccountingDocument/);
  });
});

/**
 * AccountingDetail — ReadOnly banner
 *
 * `accounting:view` は持つが `accounting:edit` を持たないユーザーが
 * 既存会計詳細 (`/accounting/:id`) を開いた場合に「閲覧専用」バナーを表示する。
 *
 * 表示条件:
 *   - id（既存会計）が存在する AND canEdit === false
 *   - 新規作成モード（id なし）では表示しない
 */
describe('AccountingDetail — ReadOnly banner', () => {
  it('id && !canEdit の条件でバナーが描画される', () => {
    // 既存会計かつ編集権限なしの場合のみバナーを表示する条件式
    expect(sourceCode).toMatch(/\bid\b\s*&&\s*!canEdit/);
  });

  it('バナーに bgWarning50 デザイントークンを使用している', () => {
    expect(sourceCode).toMatch(/C\.bgWarning50/);
  });

  it('バナーに borderWarning20 デザイントークンを使用している', () => {
    expect(sourceCode).toMatch(/C\.borderWarning20/);
  });

  it('バナーに EyeOff アイコンを使用している', () => {
    expect(sourceCode).toContain('EyeOff');
  });

  it('バナーに role="status" が設定されているアクセシビリティ対応', () => {
    expect(sourceCode).toContain('role="status"');
  });

  it('新規作成モード（id なし）ではバナーを表示しない — !id && !canEdit の形でない', () => {
    // !id && !canEdit で新規も含めてしまうような条件でないことを確認
    expect(sourceCode).not.toMatch(/!id\s*&&\s*!canEdit/);
  });
});
