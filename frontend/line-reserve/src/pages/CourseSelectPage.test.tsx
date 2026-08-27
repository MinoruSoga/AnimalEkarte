import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '@/testing/mocks/node';

import { CourseSelectPage } from './CourseSelectPage';
import type { Course } from '../types/models';
import { AUTO_ADVANCE_HELPER_TEXT } from '../lib/advance-policy';

const BASE_PROPS = {
  clinicId: '1',
  idToken: 'test-id-token',
};

const course: Course = {
  id: 10,
  name: '一般診察',
  short_name: '一般',
  show_short_name: false,
  duration_minutes: 30,
  reservation_comment: '',
  reservation_image_url: '',
  sort_order: 1,
  category: 'general',
};

describe('CourseSelectPage（R-F22/R-F23: 共通フェッチ状態管理・ステータス別エラー）', () => {
  it('読み込み中はローディング表示、取得後はコース一覧を表示する', async () => {
    server.use(http.get('/api/liff/:clinicId/courses', () => HttpResponse.json([course])));

    render(<CourseSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    expect(screen.getByText('読み込み中...')).toBeInTheDocument();
    expect(await screen.findByText('一般診察')).toBeInTheDocument();
  });

  it('API失敗(5xx)時はサーバーエラーメッセージと再試行ボタンを表示する', async () => {
    server.use(http.get('/api/liff/:clinicId/courses', () => HttpResponse.json(null, { status: 500 })));

    render(<CourseSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByRole('alert')).toHaveTextContent('サーバーエラーが発生しました');
    expect(screen.getByRole('button', { name: '再試行' })).toBeInTheDocument();
  });

  it('API失敗(401)時は再ログインメッセージを表示し、再試行ボタンは出さない', async () => {
    server.use(http.get('/api/liff/:clinicId/courses', () => HttpResponse.json(null, { status: 401 })));

    render(<CourseSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByRole('alert')).toHaveTextContent('LINEアプリを再起動して開き直してください');
    expect(screen.queryByRole('button', { name: '再試行' })).not.toBeInTheDocument();
  });

  it('再試行ボタンをクリックするとコース一覧を再取得する', async () => {
    const user = userEvent.setup();
    let callCount = 0;
    server.use(
      http.get('/api/liff/:clinicId/courses', () => {
        callCount += 1;
        if (callCount === 1) return HttpResponse.json(null, { status: 500 });
        return HttpResponse.json([course]);
      }),
    );

    render(<CourseSelectPage {...BASE_PROPS} onSelect={vi.fn()} onBack={vi.fn()} />);

    await user.click(await screen.findByRole('button', { name: '再試行' }));

    expect(await screen.findByText('一般診察')).toBeInTheDocument();
    expect(callCount).toBe(2);
  });

  it('BUG-030: auto-on-select のヘルパー文言を表示し、一覧タップで onSelect する（次へボタンは出さない）', async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    server.use(http.get('/api/liff/:clinicId/courses', () => HttpResponse.json([course])));

    render(<CourseSelectPage {...BASE_PROPS} onSelect={onSelect} onBack={vi.fn()} />);

    expect(await screen.findByTestId('auto-advance-hint')).toHaveTextContent(AUTO_ADVANCE_HELPER_TEXT);
    expect(screen.queryByRole('button', { name: '次へ' })).not.toBeInTheDocument();

    await user.click(await screen.findByText('一般診察'));
    expect(onSelect).toHaveBeenCalledWith(10, '一般診察', 'general');
  });
});
