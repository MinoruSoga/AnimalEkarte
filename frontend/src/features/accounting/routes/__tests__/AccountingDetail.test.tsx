import { describe, it, expect } from 'vitest';
import { readFileSync } from 'fs';
import { resolve } from 'path';

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
