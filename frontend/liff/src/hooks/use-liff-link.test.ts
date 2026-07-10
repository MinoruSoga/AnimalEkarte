import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, expect, it, vi, beforeEach, afterEach } from 'vitest';

// リンク導線のスモークテスト（FE-refactor.md R-F4: liffは規模が小さく健全なため smoke 1本）。
// R-F21: useLiff は frontend/src/shared-liff/use-liff.ts に統合され、内部の LIFF_MOCK は
// VITE_LIFF_MOCK 環境変数をモジュール評価時に直接判定する。そのため vi.mock('../lib/liff-config')
// では shared-liff 側まで届かない — vi.stubEnv + vi.resetModules + 動的 import で
// モジュール評価前に環境変数を確定させ、実 @line/liff SDK / 実 API 呼び出しに触れず
// useLiffLink の状態遷移ロジック（idToken/clinicId/linkToken の検証 → リンク処理）を検証する。
describe('useLiffLink（R-F4 smoke: LINEアカウント連携導線）', () => {
  beforeEach(() => {
    vi.resetModules();
    vi.stubEnv('VITE_LIFF_MOCK', 'true');
  });

  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it('clinicId または linkToken が空のとき、無効なURLエラーになる', async () => {
    const { useLiffLink } = await import('./use-liff-link');
    const { result } = renderHook(() => useLiffLink('', ''));

    await waitFor(() => {
      expect(result.current.status).toBe('error');
    });
    expect(result.current.errorMessage).toBe('無効なURLです。QRコードを再度読み取ってください');
  });

  it('clinicId/linkToken が揃っているとき、linking を経て success になる', async () => {
    vi.useFakeTimers();
    try {
      const { useLiffLink } = await import('./use-liff-link');
      const { result } = renderHook(() => useLiffLink('1', 'link-token-abc'));

      expect(result.current.status).toBe('linking');

      await act(async () => {
        await vi.advanceTimersByTimeAsync(800);
      });

      expect(result.current.status).toBe('success');
    } finally {
      vi.useRealTimers();
    }
  });
});
