import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { LiffSettings } from './types/models';

vi.mock('./api/liff-api', () => ({
  liffApi: {
    getSettings: vi.fn(),
    getProfile: vi.fn(),
  },
}));

vi.mock('@/shared-liff/use-liff', () => ({
  useLiff: vi.fn(),
}));

vi.mock('./lib/liff-config', () => ({
  getClinicId: vi.fn(),
}));

import { App } from './App';
import { liffApi } from './api/liff-api';
import { useLiff } from '@/shared-liff/use-liff';
import { getClinicId } from './lib/liff-config';

const SETTINGS: LiffSettings = {
  liff_id: 'liff-123',
  clinic_name: 'ノア動物病院',
  header_text: 'ノア動物病院',
  phone_number: '03-1234-5678',
  status: 'running',
  request_example: '',
  reservation_notice: '',
  cancel_notice: '',
  privacy_policy: '',
  show_no_staff_option: false,
  booking_window: 30,
};

function mockUseLiff(overrides: Partial<ReturnType<typeof useLiff>> = {}) {
  vi.mocked(useLiff).mockReturnValue({
    idToken: null,
    displayName: '',
    pictureUrl: null,
    isReady: false,
    initError: false,
    ...overrides,
  });
}

describe('App（FE5-20: ナビゲーション特性テスト）', () => {
  beforeEach(() => {
    vi.mocked(liffApi.getSettings).mockReset();
    vi.mocked(liffApi.getProfile).mockReset();
    vi.mocked(getClinicId).mockReset();
    vi.mocked(useLiff).mockReset();
  });

  it('App: clinicId がない場合エラーページを表示する', async () => {
    vi.mocked(getClinicId).mockReturnValue('');
    mockUseLiff();

    render(<App />);

    expect(await screen.findByText('クリニックIDが見つかりません')).toBeInTheDocument();
  });

  it('App: 設定取得失敗時にエラーページを表示する', async () => {
    vi.mocked(getClinicId).mockReturnValue('1');
    vi.mocked(liffApi.getSettings).mockRejectedValue(new Error('network error'));
    mockUseLiff();

    render(<App />);

    expect(await screen.findByText('設定の取得に失敗しました')).toBeInTheDocument();
  });

  it('App: メンテナンスフラグが立っていればメンテナンスページを表示する', async () => {
    vi.mocked(getClinicId).mockReturnValue('1');
    vi.mocked(liffApi.getSettings).mockResolvedValue({ ...SETTINGS, status: 'stopped' });
    mockUseLiff();

    render(<App />);

    expect(await screen.findByText('メンテナンス中')).toBeInTheDocument();
  });

  it('App: 初期表示は Top ページである', async () => {
    vi.mocked(getClinicId).mockReturnValue('1');
    vi.mocked(liffApi.getSettings).mockResolvedValue(SETTINGS);
    vi.mocked(liffApi.getProfile).mockResolvedValue({
      line_user_id: 'U1',
      display_name: 'テストユーザー',
      additional_fields: {},
    });
    mockUseLiff({ idToken: 'mock-token', isReady: true });

    render(<App />);

    expect(await screen.findByText('ノア動物病院')).toBeInTheDocument();
    expect(screen.getByText('新規予約')).toBeInTheDocument();
  });
});
