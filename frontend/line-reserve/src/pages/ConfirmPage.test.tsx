import { act, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '@/testing/mocks/node';

const { isInClientMock, sendMessagesMock } = vi.hoisted(() => ({
  isInClientMock: vi.fn(() => true),
  sendMessagesMock: vi.fn(() => Promise.resolve()),
}));

// 実 SDK には触れず、LINE トーク画面への通知を観測可能な副作用として置き換える。
vi.mock('@line/liff', () => ({
  default: {
    isInClient: isInClientMock,
    sendMessages: sendMessagesMock,
  },
}));

vi.mock('../lib/liff-config', async (importOriginal) => {
  const actual = await importOriginal<typeof import('../lib/liff-config')>();
  return { ...actual, LIFF_MOCK: false };
});

import { ConfirmPage } from './ConfirmPage';
import type { ReservationFlow } from '../types/models';

const baseFlow: ReservationFlow = {
  customerInfo: {
    name: '山田花子',
    phone: '090-1234-5678',
    ownerName: '山田太郎',
    pets: [{ name: 'ポチ', type: '柴犬', isNew: false }],
  },
  courseId: 10,
  courseName: '一般診察',
  courseCategory: 'general',
  staffId: 0,
  staffName: '',
  date: '2026-08-01',
  startTime: '1000',
  endTime: '1030',
  requestText: '',
  trimmingCourseId: null,
  trimmingCourseName: '',
  trimmingOptionIds: [],
};

const BASE_PROPS = {
  clinicId: '1',
  idToken: 'test-id-token',
  flow: baseFlow,
};

function createDeferred() {
  let resolve!: () => void;
  const promise = new Promise<void>((resolvePromise) => {
    resolve = resolvePromise;
  });
  return { promise, resolve };
}

describe('ConfirmPage（R-F4: 予約作成フロー・確認/送信ステップ）', () => {
  beforeEach(() => {
    isInClientMock.mockReset();
    isInClientMock.mockReturnValue(true);
    sendMessagesMock.mockReset();
    sendMessagesMock.mockResolvedValue(undefined);
  });

  it('予約内容（お名前・電話番号・ペット・日時）を表示する', () => {
    render(
      <ConfirmPage {...BASE_PROPS} onConfirm={vi.fn()} onSlotTaken={vi.fn()} onBack={vi.fn()} />,
    );

    expect(screen.getByText('山田花子')).toBeInTheDocument();
    expect(screen.getByText('090-1234-5678')).toBeInTheDocument();
    expect(screen.getByText('ポチ（柴犬）')).toBeInTheDocument();
    expect(screen.getByText('10:00 〜 10:30')).toBeInTheDocument();
  });

  it('「予約を確定する」クリック→送信成功で onConfirm(id, notes) が呼ばれる', async () => {
    const user = userEvent.setup();
    server.use(
      http.post('/api/liff/:clinicId/reservations', () =>
        HttpResponse.json({ id: 42, notes: 'R-20260801-0001' }),
      ),
    );
    const onConfirm = vi.fn();

    render(
      <ConfirmPage {...BASE_PROPS} onConfirm={onConfirm} onSlotTaken={vi.fn()} onBack={vi.fn()} />,
    );

    await user.click(screen.getByRole('button', { name: '予約を確定する' }));

    await waitFor(() => {
      expect(onConfirm).toHaveBeenCalledWith(42, 'R-20260801-0001');
    });
  });

  it('送信中の再クリックでは予約POSTを重複送信せず、応答後に完了する', async () => {
    const user = userEvent.setup();
    const responseGate = createDeferred();
    let reservationPostCount = 0;
    server.use(
      http.post('/api/liff/:clinicId/reservations', async () => {
        reservationPostCount += 1;
        await responseGate.promise;
        return HttpResponse.json({ id: 42, notes: 'R-20260801-0001' });
      }),
    );
    const onConfirm = vi.fn();

    render(
      <ConfirmPage {...BASE_PROPS} onConfirm={onConfirm} onSlotTaken={vi.fn()} onBack={vi.fn()} />,
    );

    await user.click(screen.getByRole('button', { name: '予約を確定する' }));

    const pendingButton = await screen.findByRole('button', { name: '送信中...' });
    expect(pendingButton).toBeDisabled();
    await waitFor(() => {
      expect(reservationPostCount).toBe(1);
    });

    await user.click(pendingButton);
    expect(reservationPostCount).toBe(1);

    await act(async () => {
      responseGate.resolve();
    });

    await waitFor(() => {
      expect(onConfirm).toHaveBeenCalledWith(42, 'R-20260801-0001');
    });
  });

  it('LINEメッセージ送信の完了後に onConfirm を呼ぶ', async () => {
    const user = userEvent.setup();
    const messageGate = createDeferred();
    const callOrder: string[] = [];
    server.use(
      http.post('/api/liff/:clinicId/reservations', () =>
        HttpResponse.json({ id: 42, notes: 'R-20260801-0001' }),
      ),
    );
    sendMessagesMock.mockImplementation(async () => {
      callOrder.push('sendMessages:start');
      await messageGate.promise;
      callOrder.push('sendMessages:complete');
    });
    const onConfirm = vi.fn(() => {
      callOrder.push('onConfirm');
    });

    render(
      <ConfirmPage {...BASE_PROPS} onConfirm={onConfirm} onSlotTaken={vi.fn()} onBack={vi.fn()} />,
    );

    await user.click(screen.getByRole('button', { name: '予約を確定する' }));

    await waitFor(() => {
      expect(sendMessagesMock).toHaveBeenCalledTimes(1);
    });
    expect(sendMessagesMock).toHaveBeenCalledWith([
      expect.objectContaining({
        type: 'text',
        text: expect.stringContaining('■ 予約番号: R-20260801-0001'),
      }),
    ]);
    expect(onConfirm).not.toHaveBeenCalled();

    await act(async () => {
      messageGate.resolve();
    });

    await waitFor(() => {
      expect(onConfirm).toHaveBeenCalledWith(42, 'R-20260801-0001');
    });
    expect(callOrder).toEqual([
      'sendMessages:start',
      'sendMessages:complete',
      'onConfirm',
    ]);
  });

  it('送信失敗（500）で「予約の確定に失敗しました」を表示し、再送信できる状態に戻す', async () => {
    const user = userEvent.setup();
    server.use(
      http.post('/api/liff/:clinicId/reservations', () => HttpResponse.json(null, { status: 500 })),
    );
    const onConfirm = vi.fn();

    render(
      <ConfirmPage {...BASE_PROPS} onConfirm={onConfirm} onSlotTaken={vi.fn()} onBack={vi.fn()} />,
    );

    await user.click(screen.getByRole('button', { name: '予約を確定する' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('予約の確定に失敗しました');
    expect(onConfirm).not.toHaveBeenCalled();
    expect(screen.getByRole('button', { name: '予約を確定する' })).not.toBeDisabled();
  });

  it('409（枠が既に埋まっている）で onSlotTaken(message, redirect_step) が呼ばれ、onConfirmは呼ばれない', async () => {
    const user = userEvent.setup();
    server.use(
      http.post('/api/liff/:clinicId/reservations', () =>
        HttpResponse.json(
          { error: '選択された時間枠は既に予約が入っています。', code: 'slot_taken', redirect_step: 5 },
          { status: 409 },
        ),
      ),
    );
    const onConfirm = vi.fn();
    const onSlotTaken = vi.fn();

    render(
      <ConfirmPage {...BASE_PROPS} onConfirm={onConfirm} onSlotTaken={onSlotTaken} onBack={vi.fn()} />,
    );

    await user.click(screen.getByRole('button', { name: '予約を確定する' }));

    await waitFor(() => {
      expect(onSlotTaken).toHaveBeenCalledWith('選択された時間枠は既に予約が入っています。', 5);
    });
    expect(onConfirm).not.toHaveBeenCalled();
  });

  it('409では onSlotTaken を呼び、エラーalertを表示しない', async () => {
    const user = userEvent.setup();
    server.use(
      http.post('/api/liff/:clinicId/reservations', () =>
        HttpResponse.json(
          { error: '選択された時間枠は既に予約が入っています。', code: 'slot_taken', redirect_step: 5 },
          { status: 409 },
        ),
      ),
    );
    const onSlotTaken = vi.fn();

    render(
      <ConfirmPage
        {...BASE_PROPS}
        onConfirm={vi.fn()}
        onSlotTaken={onSlotTaken}
        onBack={vi.fn()}
      />,
    );

    await user.click(screen.getByRole('button', { name: '予約を確定する' }));

    await waitFor(() => {
      expect(onSlotTaken).toHaveBeenCalledWith('選択された時間枠は既に予約が入っています。', 5);
    });
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });
});
