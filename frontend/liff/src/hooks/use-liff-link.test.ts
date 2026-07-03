import { renderHook, waitFor, act } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

// リンク導線のスモークテスト（FE-refactor.md R-F4: liffは規模が小さく健全なため smoke 1本）。
// LIFF_MOCK=true にすることで実 @line/liff SDK / 実 API 呼び出しに触れず、
// useLiffLink の状態遷移ロジック（idToken/clinicId/linkToken の検証 → リンク処理）を検証する。
vi.mock('../lib/liff-config', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../lib/liff-config')>();
  return { ...actual, LIFF_MOCK: true };
});

import { useLiffLink } from './use-liff-link';

describe('useLiffLink（R-F4 smoke: LINEアカウント連携導線）', () => {
  it('clinicId または linkToken が空のとき、無効なURLエラーになる', async () => {
    const { result } = renderHook(() => useLiffLink('', ''));

    await waitFor(() => {
      expect(result.current.status).toBe('error');
    });
    expect(result.current.errorMessage).toBe('無効なURLです。QRコードを再度読み取ってください');
  });

  it('clinicId/linkToken が揃っているとき、linking を経て success になる', async () => {
    vi.useFakeTimers();
    try {
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
