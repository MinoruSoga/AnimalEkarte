import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '@/testing/mocks/node';

import { PetHealthPage } from './PetHealthPage';
import type { HealthCardResponse } from '../api/liff-api';

const BASE_PROPS = {
  idToken: 'test-id-token',
  displayName: 'テストユーザー',
  pictureUrl: null,
};

const healthCard: HealthCardResponse = {
  owner_name: '山田太郎',
  pets: [],
};

function renderWithClinicId() {
  window.history.pushState({}, '', '/liff/health-card?clinic_id=1');
  return render(<PetHealthPage {...BASE_PROPS} />);
}

function renderWithoutClinicId() {
  window.history.pushState({}, '', '/liff/health-card');
  return render(<PetHealthPage {...BASE_PROPS} />);
}

describe('PetHealthPage（R-F22/R-F23: 共通フェッチ状態管理・ステータス別エラー）', () => {
  it('clinic_id 欠落時は専用文言を表示し、再試行ボタンは出さない（BUG-014）', async () => {
    renderWithoutClinicId();

    expect(await screen.findByRole('alert')).toHaveTextContent('クリニックIDが見つかりません');
    expect(screen.queryByText('しばらく経ってから再度お試しください')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '再試行' })).not.toBeInTheDocument();
  });

  it('取得後は飼い主名を表示する', async () => {
    server.use(http.get('/api/liff/:clinicId/health-card', () => HttpResponse.json(healthCard)));

    renderWithClinicId();

    expect(await screen.findByText('山田太郎')).toBeInTheDocument();
  });

  it('API失敗(5xx)時はサーバーエラーメッセージと再試行ボタンを表示する', async () => {
    server.use(http.get('/api/liff/:clinicId/health-card', () => HttpResponse.json(null, { status: 500 })));

    renderWithClinicId();

    expect(await screen.findByRole('alert')).toHaveTextContent('サーバーエラーが発生しました');
    expect(screen.getByRole('button', { name: '再試行' })).toBeInTheDocument();
  });

  it('API失敗(401)時は再ログインメッセージを表示し、再試行ボタンは出さない', async () => {
    server.use(http.get('/api/liff/:clinicId/health-card', () => HttpResponse.json(null, { status: 401 })));

    renderWithClinicId();

    expect(await screen.findByRole('alert')).toHaveTextContent('LINEアプリを再起動して開き直してください');
    expect(screen.queryByRole('button', { name: '再試行' })).not.toBeInTheDocument();
  });

  it('再試行ボタンをクリックすると健康記録を再取得する', async () => {
    const user = userEvent.setup();
    let callCount = 0;
    server.use(
      http.get('/api/liff/:clinicId/health-card', () => {
        callCount += 1;
        if (callCount === 1) return HttpResponse.json(null, { status: 500 });
        return HttpResponse.json(healthCard);
      }),
    );

    renderWithClinicId();

    await user.click(await screen.findByRole('button', { name: '再試行' }));

    expect(await screen.findByText('山田太郎')).toBeInTheDocument();
    expect(callCount).toBe(2);
  });
});
