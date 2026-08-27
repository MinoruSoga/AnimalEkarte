import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '@/testing/mocks/node';

import { TrimmingCourseSelectPage } from './TrimmingCourseSelectPage';
import type { TrimmingCourse } from '../types/models';
import { AUTO_ADVANCE_HELPER_TEXT } from '../lib/advance-policy';

const BASE_PROPS = {
  clinicId: '1',
  idToken: 'test-id-token',
};

const trimmingCourse: TrimmingCourse = {
  id: 1,
  name: 'フルコース',
  description: '',
  price: 8000,
  sort_order: 1,
};

describe('TrimmingCourseSelectPage（R-F22/R-F23: 共通フェッチ状態管理・ステータス別エラー）', () => {
  it('取得後はトリミングコース一覧を表示する', async () => {
    server.use(http.get('/api/liff/:clinicId/trimming-courses', () => HttpResponse.json([trimmingCourse])));

    render(<TrimmingCourseSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByText('フルコース')).toBeInTheDocument();
  });

  it('API失敗(5xx)時はサーバーエラーメッセージと再試行ボタンを表示する', async () => {
    server.use(http.get('/api/liff/:clinicId/trimming-courses', () => HttpResponse.json(null, { status: 500 })));

    render(<TrimmingCourseSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByRole('alert')).toHaveTextContent('サーバーエラーが発生しました');
    expect(screen.getByRole('button', { name: '再試行' })).toBeInTheDocument();
  });

  it('API失敗(401)時は再ログインメッセージを表示し、再試行ボタンは出さない', async () => {
    server.use(http.get('/api/liff/:clinicId/trimming-courses', () => HttpResponse.json(null, { status: 401 })));

    render(<TrimmingCourseSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByRole('alert')).toHaveTextContent('LINEアプリを再起動して開き直してください');
    expect(screen.queryByRole('button', { name: '再試行' })).not.toBeInTheDocument();
  });

  it('再試行ボタンをクリックするとトリミングコース一覧を再取得する', async () => {
    const user = userEvent.setup();
    let callCount = 0;
    server.use(
      http.get('/api/liff/:clinicId/trimming-courses', () => {
        callCount += 1;
        if (callCount === 1) return HttpResponse.json(null, { status: 500 });
        return HttpResponse.json([trimmingCourse]);
      }),
    );

    render(<TrimmingCourseSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    await user.click(await screen.findByRole('button', { name: '再試行' }));

    expect(await screen.findByText('フルコース')).toBeInTheDocument();
    expect(callCount).toBe(2);
  });

  it('BUG-030: auto-on-select のヘルパー文言を表示し、一覧タップで onSelect する', async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    server.use(http.get('/api/liff/:clinicId/trimming-courses', () => HttpResponse.json([trimmingCourse])));

    render(<TrimmingCourseSelectPage {...BASE_PROPS} onSelect={onSelect} onBack={vi.fn()} />);

    expect(await screen.findByTestId('auto-advance-hint')).toHaveTextContent(AUTO_ADVANCE_HELPER_TEXT);
    expect(screen.queryByRole('button', { name: '次へ' })).not.toBeInTheDocument();

    await user.click(await screen.findByText('フルコース'));
    expect(onSelect).toHaveBeenCalledWith(1, 'フルコース');
  });
});
