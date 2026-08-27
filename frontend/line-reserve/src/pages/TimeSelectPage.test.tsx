import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '@/testing/mocks/node';

import { TimeSelectPage } from './TimeSelectPage';
import type { AvailableTime } from '../types/models';
import { AUTO_ADVANCE_HELPER_TEXT } from '../lib/advance-policy';

const BASE_PROPS = {
  clinicId: '1',
  idToken: 'test-id-token',
  courseId: 10,
  staffId: 0,
  date: '2026-08-01',
  isTrimming: false,
};

describe('TimeSelectPage（R-F4: 予約作成フロー・枠選択ステップ）', () => {
  it('読み込み中はローディング表示、取得後は空き時間一覧を表示する', async () => {
    const times: AvailableTime[] = [
      { start_time: '1000', end_time: '1030' },
      { start_time: '1100', end_time: '1130', display_time: '11:00〜' },
    ];
    server.use(
      http.get('/api/liff/:clinicId/available-times', () => HttpResponse.json(times)),
    );

    render(<TimeSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    expect(screen.getByText('読み込み中...')).toBeInTheDocument();

    expect(await screen.findByText(/10:00/)).toBeInTheDocument();
    expect(screen.getByText(/11:00〜/)).toBeInTheDocument();
  });

  it('時間枠をクリックすると onSelect(start_time, end_time) が呼ばれる', async () => {
    const user = userEvent.setup();
    const times: AvailableTime[] = [{ start_time: '1000', end_time: '1030' }];
    server.use(
      http.get('/api/liff/:clinicId/available-times', () => HttpResponse.json(times)),
    );
    const onSelect = vi.fn();

    render(<TimeSelectPage {...BASE_PROPS} onSelect={onSelect} onBack={vi.fn()} />);

    const button = await screen.findByRole('button', { name: /10:00/ });
    await user.click(button);

    expect(onSelect).toHaveBeenCalledWith('1000', '1030');
  });

  it('空き時間が0件のとき「この日の空き時間はありません」と表示する', async () => {
    server.use(
      http.get('/api/liff/:clinicId/available-times', () => HttpResponse.json([])),
    );

    render(<TimeSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByText('この日の空き時間はありません')).toBeInTheDocument();
  });

  it('API失敗(5xx)時はサーバーエラーメッセージと再試行ボタンを表示する（R-F22）', async () => {
    server.use(
      http.get('/api/liff/:clinicId/available-times', () => HttpResponse.json(null, { status: 500 })),
    );

    render(<TimeSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByRole('alert')).toHaveTextContent('サーバーエラーが発生しました');
    expect(screen.getByRole('button', { name: '再試行' })).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.queryByText('読み込み中...')).not.toBeInTheDocument();
    });
  });

  it('API失敗(401)時は再ログインを促すメッセージを表示し、再試行ボタンは出さない（R-F22）', async () => {
    server.use(
      http.get('/api/liff/:clinicId/available-times', () => HttpResponse.json(null, { status: 401 })),
    );

    render(<TimeSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByRole('alert')).toHaveTextContent('LINEアプリを再起動して開き直してください');
    expect(screen.queryByRole('button', { name: '再試行' })).not.toBeInTheDocument();
  });

  it('再試行ボタンをクリックすると空き時間を再取得する（R-F22/R-F23）', async () => {
    const user = userEvent.setup();
    let callCount = 0;
    server.use(
      http.get('/api/liff/:clinicId/available-times', () => {
        callCount += 1;
        if (callCount === 1) return HttpResponse.json(null, { status: 500 });
        return HttpResponse.json([{ start_time: '1000', end_time: '1030' }]);
      }),
    );

    render(<TimeSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    await user.click(await screen.findByRole('button', { name: '再試行' }));

    expect(await screen.findByText(/10:00/)).toBeInTheDocument();
    expect(callCount).toBe(2);
  });

  it('BUG-030: auto-on-select のヘルパー文言を表示し、次へボタンは出さない', async () => {
    const times: AvailableTime[] = [{ start_time: '1000', end_time: '1030' }];
    server.use(
      http.get('/api/liff/:clinicId/available-times', () => HttpResponse.json(times)),
    );

    render(<TimeSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByTestId('auto-advance-hint')).toHaveTextContent(AUTO_ADVANCE_HELPER_TEXT);
    expect(screen.queryByRole('button', { name: '次へ' })).not.toBeInTheDocument();
  });
});
