import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import type { LiffSettings, CustomerInfo } from './types/models';

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

// FE5-21: handleSlotTaken に到達するための中間ページのみスタブ化する
// （TopPage/ErrorPage/MaintenancePage は実物のまま。FE5-20 の4テストの
// アサーション対象と非破壊で両立させる）
vi.mock('./pages/CustomerInfoPage', () => ({
  CustomerInfoPage: ({ onNext }: { onNext: (info: CustomerInfo) => void }) => (
    <button onClick={() => onNext({ name: 'テスト', phone: '000', ownerName: '', pets: [] })}>
      next-step1
    </button>
  ),
}));

vi.mock('./pages/CourseSelectPage', () => ({
  CourseSelectPage: ({
    onSelect,
  }: {
    onSelect: (id: number, name: string, category?: 'general' | 'trimming') => void;
  }) => <button onClick={() => onSelect(10, '一般診察', 'general')}>next-step2</button>,
}));

vi.mock('./pages/StaffSelectPage', () => ({
  StaffSelectPage: ({ onSelect }: { onSelect: (id: number, name: string) => void }) => (
    <button onClick={() => onSelect(1, 'スタッフA')}>next-step3</button>
  ),
}));

vi.mock('./pages/DateSelectPage', () => ({
  DateSelectPage: ({
    onSelect,
    onNext,
  }: {
    onSelect: (date: string) => void;
    onNext: () => void;
  }) => (
    <button
      onClick={() => {
        onSelect('2026-08-01');
        onNext();
      }}
    >
      next-step4
    </button>
  ),
}));

vi.mock('./pages/TimeSelectPage', () => ({
  TimeSelectPage: ({ onSelect }: { onSelect: (start: string, end: string) => void }) => (
    <button onClick={() => onSelect('1000', '1030')}>next-step5</button>
  ),
}));

vi.mock('./pages/RequestPage', () => ({
  RequestPage: ({ onNext }: { onNext: (text: string) => void }) => (
    <button onClick={() => onNext('')}>next-step6</button>
  ),
}));

vi.mock('./pages/ConfirmPage', () => ({
  ConfirmPage: ({
    onSlotTaken,
  }: {
    onSlotTaken: (message: string, redirectStep: number) => void;
  }) => (
    <button onClick={() => onSlotTaken('選択された時間枠は既に予約が入っています。', 4)}>
      trigger-slot-taken
    </button>
  ),
}));

import { App } from './App';
import { liffApi } from './api/liff-api';
import { useLiff } from '@/shared-liff/use-liff';
import { getClinicId } from './lib/liff-config';

const SETTINGS: LiffSettings = {
  liff_id: 'liff-123',
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

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('App: clinicId がない場合エラーページを表示する', async () => {
    vi.mocked(getClinicId).mockReturnValue('');
    mockUseLiff();

    const { container } = render(<App />);

    expect(await screen.findByText('クリニックIDが見つかりません')).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('エラーが発生しました');
    expect(container.firstElementChild?.className).toContain('bg-noah-teal-light');
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it('App: 設定取得失敗時にエラーページを表示する', async () => {
    vi.mocked(getClinicId).mockReturnValue('1');
    vi.mocked(liffApi.getSettings).mockRejectedValue(new Error('network error'));
    mockUseLiff();

    const { container } = render(<App />);

    expect(await screen.findByText('設定の取得に失敗しました')).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 1 })).toHaveTextContent('エラーが発生しました');
    expect(container.firstElementChild?.className).toContain('bg-noah-teal-light');
    expect(screen.getByRole('button', { name: '再読み込み' })).toBeInTheDocument();
  });

  it('App: 無効 clinic（設定失敗）でも成功シェルへフォールバックしない（BUG-027 fail-closed）', async () => {
    vi.mocked(getClinicId).mockReturnValue('999');
    vi.mocked(liffApi.getSettings).mockRejectedValue(new Error('not found'));
    mockUseLiff();

    render(<App />);

    expect(await screen.findByText('設定の取得に失敗しました')).toBeInTheDocument();
    expect(screen.queryByText('新規予約')).not.toBeInTheDocument();
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

  it('App: 枠競合時にバナーを表示し alert を使わない', async () => {
    const alertSpy = vi.spyOn(window, 'alert').mockImplementation(() => {});
    vi.mocked(getClinicId).mockReturnValue('1');
    vi.mocked(liffApi.getSettings).mockResolvedValue(SETTINGS);
    vi.mocked(liffApi.getProfile).mockResolvedValue({
      line_user_id: 'U1',
      display_name: 'テストユーザー',
      additional_fields: {},
    });
    mockUseLiff({ idToken: 'mock-token', isReady: true });

    const user = userEvent.setup();
    render(<App />);

    await user.click(await screen.findByText('新規予約'));
    await user.click(await screen.findByText('next-step1'));
    await user.click(await screen.findByText('next-step2'));
    await user.click(await screen.findByText('next-step3'));
    await user.click(await screen.findByText('next-step4'));
    await user.click(await screen.findByText('next-step5'));
    await user.click(await screen.findByText('next-step6'));
    await user.click(await screen.findByText('trigger-slot-taken'));

    expect(await screen.findByRole('alert')).toHaveTextContent(
      '選択された時間枠は既に予約が入っています。',
    );
    expect(alertSpy).not.toHaveBeenCalled();
    // redirectStep=4 → step4（DateSelectPage スタブ）に戻る
    expect(await screen.findByText('next-step4')).toBeInTheDocument();
  });

  it('App: 枠競合バナーは閉じるボタンでクリアできる', async () => {
    vi.mocked(getClinicId).mockReturnValue('1');
    vi.mocked(liffApi.getSettings).mockResolvedValue(SETTINGS);
    vi.mocked(liffApi.getProfile).mockResolvedValue({
      line_user_id: 'U1',
      display_name: 'テストユーザー',
      additional_fields: {},
    });
    mockUseLiff({ idToken: 'mock-token', isReady: true });

    const user = userEvent.setup();
    render(<App />);

    await user.click(await screen.findByText('新規予約'));
    await user.click(await screen.findByText('next-step1'));
    await user.click(await screen.findByText('next-step2'));
    await user.click(await screen.findByText('next-step3'));
    await user.click(await screen.findByText('next-step4'));
    await user.click(await screen.findByText('next-step5'));
    await user.click(await screen.findByText('next-step6'));
    await user.click(await screen.findByText('trigger-slot-taken'));

    await screen.findByRole('alert');
    await user.click(screen.getByRole('button', { name: '閉じる' }));

    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });
});
