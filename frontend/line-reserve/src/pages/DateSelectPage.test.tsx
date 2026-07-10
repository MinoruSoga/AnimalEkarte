import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '@/testing/mocks/node';

import { DateSelectPage } from './DateSelectPage';

const BASE_PROPS = {
  clinicId: '1',
  idToken: 'test-id-token',
  courseId: 10,
  staffId: 0,
  selectedDate: '',
  bookingWindow: 30,
};

describe('DateSelectPage（R-F22/R-F23: 共通フェッチ状態管理・ステータス別エラー）', () => {
  it('API失敗(5xx)時はサーバーエラーメッセージと再試行ボタンを表示する', async () => {
    server.use(http.get('/api/liff/:clinicId/available-dates', () => HttpResponse.json(null, { status: 500 })));

    render(<DateSelectPage {...BASE_PROPS} onSelect={vi.fn()} onNext={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByRole('alert')).toHaveTextContent('サーバーエラーが発生しました');
    expect(screen.getByRole('button', { name: '再試行' })).toBeInTheDocument();
  });

  it('API失敗(401)時は再ログインメッセージを表示し、再試行ボタンは出さない', async () => {
    server.use(http.get('/api/liff/:clinicId/available-dates', () => HttpResponse.json(null, { status: 401 })));

    render(<DateSelectPage {...BASE_PROPS} onSelect={vi.fn()} onNext={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByRole('alert')).toHaveTextContent('LINEアプリを再起動して開き直してください');
    expect(screen.queryByRole('button', { name: '再試行' })).not.toBeInTheDocument();
  });

  it('再試行ボタンをクリックすると空き日程を再取得する', async () => {
    const user = userEvent.setup();
    let callCount = 0;
    server.use(
      http.get('/api/liff/:clinicId/available-dates', () => {
        callCount += 1;
        if (callCount === 1) return HttpResponse.json(null, { status: 500 });
        return HttpResponse.json({ dates: [{ date: '2026-08-01', available: true }], window: null });
      }),
    );

    render(<DateSelectPage {...BASE_PROPS} onSelect={vi.fn()} onNext={vi.fn()} onBack={vi.fn()} />);

    await user.click(await screen.findByRole('button', { name: '再試行' }));

    expect(callCount).toBe(2);
  });
});
