import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '@/testing/mocks/node';

import { TrimmingOptionSelectPage } from './TrimmingOptionSelectPage';
import type { TrimmingOption } from '../types/models';

const BASE_PROPS = {
  clinicId: '1',
  idToken: 'test-id-token',
  selectedOptionIds: [],
};

const option: TrimmingOption = {
  id: 1,
  name: '爪切り',
  description: '',
  price: 500,
  sort_order: 1,
};

describe('TrimmingOptionSelectPage（R-F22/R-F23: 共通フェッチ状態管理・ステータス別エラー）', () => {
  it('取得後はオプション一覧を表示する', async () => {
    server.use(http.get('/api/liff/:clinicId/trimming-options', () => HttpResponse.json([option])));

    render(<TrimmingOptionSelectPage {...BASE_PROPS} onNext={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByText('爪切り')).toBeInTheDocument();
  });

  it('主CTAラベルは「次へ」であり「次へ進む」ではない', async () => {
    server.use(http.get('/api/liff/:clinicId/trimming-options', () => HttpResponse.json([option])));

    render(<TrimmingOptionSelectPage {...BASE_PROPS} onNext={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByRole('button', { name: '次へ' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '次へ進む' })).not.toBeInTheDocument();
  });

  it('主CTA「次へ」クリックで選択中の optionIds を onNext に渡す', async () => {
    const user = userEvent.setup();
    const onNext = vi.fn();
    server.use(http.get('/api/liff/:clinicId/trimming-options', () => HttpResponse.json([option])));

    render(<TrimmingOptionSelectPage {...BASE_PROPS} onNext={onNext} onBack={vi.fn()} />);

    await user.click(await screen.findByText('爪切り'));
    await user.click(screen.getByRole('button', { name: '次へ' }));

    expect(onNext).toHaveBeenCalledTimes(1);
    expect(onNext).toHaveBeenCalledWith([1]);
  });

  it('API失敗(5xx)時はサーバーエラーメッセージと再試行ボタンを表示する', async () => {
    server.use(http.get('/api/liff/:clinicId/trimming-options', () => HttpResponse.json(null, { status: 500 })));

    render(<TrimmingOptionSelectPage {...BASE_PROPS} onNext={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByRole('alert')).toHaveTextContent('サーバーエラーが発生しました');
    expect(screen.getByRole('button', { name: '再試行' })).toBeInTheDocument();
  });

  it('API失敗(401)時は再ログインメッセージを表示し、再試行ボタンは出さない', async () => {
    server.use(http.get('/api/liff/:clinicId/trimming-options', () => HttpResponse.json(null, { status: 401 })));

    render(<TrimmingOptionSelectPage {...BASE_PROPS} onNext={vi.fn()} onBack={vi.fn()} />);

    expect(await screen.findByRole('alert')).toHaveTextContent('LINEアプリを再起動して開き直してください');
    expect(screen.queryByRole('button', { name: '再試行' })).not.toBeInTheDocument();
  });

  it('再試行ボタンをクリックするとオプション一覧を再取得する', async () => {
    const user = userEvent.setup();
    let callCount = 0;
    server.use(
      http.get('/api/liff/:clinicId/trimming-options', () => {
        callCount += 1;
        if (callCount === 1) return HttpResponse.json(null, { status: 500 });
        return HttpResponse.json([option]);
      }),
    );

    render(<TrimmingOptionSelectPage {...BASE_PROPS} onNext={vi.fn()} onBack={vi.fn()} />);

    await user.click(await screen.findByRole('button', { name: '再試行' }));

    expect(await screen.findByText('爪切り')).toBeInTheDocument();
    expect(callCount).toBe(2);
  });
});
