import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { server } from '@/testing/mocks/node';
import { createTestWrapper } from '@/testing/utils';
import { LstepSettingsForm } from './LstepSettingsForm';
import type { LstepSettingsResponse } from '../hooks/use-lstep-settings';

const CLINIC_ID = 'clinic-test-1';

const configuredSyncOn: LstepSettingsResponse = {
  lstep_api_key_masked: 'abcd',
  lstep_base_url: 'https://app.lstep.jp',
  line_channel_access_token_masked: 'wxyz',
  line_channel_secret_masked: 'efgh',
  liff_id: '1234567890-xxxxxxxx',
  line_account_name: null,
  is_configured: true,
  last_updated_at: '2026-04-01T00:00:00Z',
  is_sync_enabled: true,
  sync_enabled_at: '2026-03-15T10:00:00Z',
  cpm_version: 'v1',
  dormant_prevention_180_days: 180,
  dormant_prevention_210_days: 210,
  dormant_prevention_240_days: 240,
  dormant_prevention_365_days: 365,
  cpm_v2_coming_threshold: 2,
  cpm_v2_good_threshold: 4,
  cpm_v2_family_threshold: 8,
  cpm_v2_noah_threshold: 13,
  cpm_v1_dormant_days: 240,
  cpm_v1_noah_days: 365,
  cpm_v1_noah_annual_visits: 3,
  cpm_v1_noah_ltv: 80000,
  cpm_v1_core_days: 180,
  cpm_v1_core_annual_visits: 2,
  cpm_v1_core_ltv: 50000,
  cpm_v1_spot_min_amount: 30000,
  cpm_v1_spot_inactive_days: 90,
  cpm_v1_growing_max_days: 90,
  cpm_v1_growing_min_visits: 2,
  cpm_v1_growing_max_visits: 3,
  cpm_v1_ltv_break_low: 20000,
  health_prevention_lookback_days: 365,
  vaccine_deadline_days: 60,
};

const configuredSyncOff: LstepSettingsResponse = {
  ...configuredSyncOn,
  is_sync_enabled: false,
  sync_enabled_at: null,
};

const unconfigured: LstepSettingsResponse = {
  lstep_api_key_masked: null,
  lstep_base_url: null,
  line_channel_access_token_masked: null,
  line_channel_secret_masked: null,
  liff_id: null,
  line_account_name: null,
  is_configured: false,
  last_updated_at: null,
  is_sync_enabled: false,
  sync_enabled_at: null,
  cpm_version: 'v1',
  dormant_prevention_180_days: 180,
  dormant_prevention_210_days: 210,
  dormant_prevention_240_days: 240,
  dormant_prevention_365_days: 365,
  cpm_v2_coming_threshold: 2,
  cpm_v2_good_threshold: 4,
  cpm_v2_family_threshold: 8,
  cpm_v2_noah_threshold: 13,
  cpm_v1_dormant_days: 240,
  cpm_v1_noah_days: 365,
  cpm_v1_noah_annual_visits: 3,
  cpm_v1_noah_ltv: 80000,
  cpm_v1_core_days: 180,
  cpm_v1_core_annual_visits: 2,
  cpm_v1_core_ltv: 50000,
  cpm_v1_spot_min_amount: 30000,
  cpm_v1_spot_inactive_days: 90,
  cpm_v1_growing_max_days: 90,
  cpm_v1_growing_min_visits: 2,
  cpm_v1_growing_max_visits: 3,
  cpm_v1_ltv_break_low: 20000,
  health_prevention_lookback_days: 180,
  vaccine_deadline_days: 30,
};

function setupGetHandler(data: LstepSettingsResponse) {
  server.use(
    http.get(`/api/v1/clinics/${CLINIC_ID}/lstep-settings`, () =>
      HttpResponse.json(data)
    )
  );
}

function createWrapper() {
  return createTestWrapper({ router: true });
}

async function renderAndWait(data: LstepSettingsResponse) {
  setupGetHandler(data);
  render(<LstepSettingsForm />, { wrapper: createWrapper() });
  await waitFor(() => {
    expect(screen.queryByText('読み込み中...')).not.toBeInTheDocument();
  });
}

beforeEach(() => {
  localStorage.setItem('auth_current_clinic:v1', CLINIC_ID);
});

afterEach(() => {
  localStorage.removeItem('auth_current_clinic:v1');
});

// ─────────────────────────────────────────────────────────────
// A: Switch 表示・バッジ・テキスト
// ─────────────────────────────────────────────────────────────

describe('LstepSettingsForm — A: Switch 表示・バッジ (FEAT-376)', () => {
  it('Switch は見た目のトラック寸法を保ちつつ44px以上の操作領域を持つ', async () => {
    await renderAndWait(configuredSyncOn);

    expect(screen.getByRole('switch', { name: '同期を有効にする' })).toHaveClass(
      'h-11',
      'w-12',
      'relative',
      'before:h-7'
    );
    expect(screen.getByRole('switch', { name: '同期を有効にする' })).not.toHaveClass('border-y-8');
  });

  it('同期ON: Switch が aria-checked="true"、"同期中" バッジが表示される', async () => {
    await renderAndWait(configuredSyncOn);
    expect(screen.getByRole('switch', { name: '同期を有効にする' })).toHaveAttribute(
      'aria-checked',
      'true'
    );
    expect(screen.getByText('同期中')).toBeInTheDocument();
  });

  it('同期OFF: Switch が aria-checked="false"、"同期停止中" バッジが表示される', async () => {
    await renderAndWait(configuredSyncOff);
    expect(screen.getByRole('switch', { name: '同期を有効にする' })).toHaveAttribute(
      'aria-checked',
      'false'
    );
    expect(screen.getByText('同期停止中')).toBeInTheDocument();
  });

  it('未設定: 同期バッジ（"同期中"/"同期停止中"）は表示されない', async () => {
    await renderAndWait(unconfigured);
    expect(screen.queryByText('同期中')).not.toBeInTheDocument();
    expect(screen.queryByText('同期停止中')).not.toBeInTheDocument();
  });

  it('masked API キー "abcd" が "••••abcd" 形式で表示される', async () => {
    await renderAndWait(configuredSyncOn);
    expect(screen.getByText(/••••abcd/)).toBeInTheDocument();
  });

  it('sync_enabled_at が日本語ローカライズ形式（"2026年3月15日"）で表示される', async () => {
    await renderAndWait(configuredSyncOn);
    expect(screen.getByText(/2026年3月15日/)).toBeInTheDocument();
  });

  it('設定済み: "OFFにすると連携処理を一時停止します" の説明コピーが表示される', async () => {
    await renderAndWait(configuredSyncOn);
    expect(
      screen.getByText(/OFFにすると連携処理を一時停止します/)
    ).toBeInTheDocument();
  });
});

// ─────────────────────────────────────────────────────────────
// B: 保存・PATCH 送信
// ─────────────────────────────────────────────────────────────

describe('LstepSettingsForm — B: 保存・PATCH 送信 (FEAT-376)', () => {
  it('"保存" ボタンが表示される', async () => {
    await renderAndWait(configuredSyncOn);
    expect(screen.getByRole('button', { name: '保存' })).toBeInTheDocument();
  });

  it('同期ON状態で保存すると PATCH body に is_sync_enabled=true が含まれる', async () => {
    let capturedBody: Record<string, unknown> | null = null;
    server.use(
      http.patch(`/api/v1/clinics/${CLINIC_ID}/lstep-settings`, async ({ request }) => {
        capturedBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(configuredSyncOn);
      })
    );

    await renderAndWait(configuredSyncOn);

    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: '保存' }));

    await waitFor(() => {
      expect(capturedBody).not.toBeNull();
    });
    expect(capturedBody?.is_sync_enabled).toBe(true);
  });

  it('同期OFF状態で保存すると PATCH body に is_sync_enabled=false が含まれる', async () => {
    let capturedBody: Record<string, unknown> | null = null;
    server.use(
      http.patch(`/api/v1/clinics/${CLINIC_ID}/lstep-settings`, async ({ request }) => {
        capturedBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(configuredSyncOff);
      })
    );

    await renderAndWait(configuredSyncOff);

    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: '保存' }));

    await waitFor(() => {
      expect(capturedBody).not.toBeNull();
    });
    expect(capturedBody?.is_sync_enabled).toBe(false);
  });
});

// ─────────────────────────────────────────────────────────────
// C: 設定削除ダイアログ
// ─────────────────────────────────────────────────────────────

describe('LstepSettingsForm — C: 設定削除ダイアログ (FEAT-376)', () => {
  it('"設定削除" ボタンをクリックすると削除確認ダイアログが開く', async () => {
    await renderAndWait(configuredSyncOn);
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: /設定削除/ }));
    const dialog = screen.getByRole('alertdialog');
    expect(dialog).toBeInTheDocument();
    expect(within(dialog).getByText('Lステップ設定を削除しますか？')).toBeInTheDocument();
  });

  it('"削除する" ボタンをクリックすると DELETE エンドポイントが呼ばれる', async () => {
    let deleteCount = 0;
    server.use(
      http.delete(`/api/v1/clinics/${CLINIC_ID}/lstep-settings`, () => {
        deleteCount++;
        return new HttpResponse(null, { status: 200 });
      })
    );

    await renderAndWait(configuredSyncOn);

    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: /設定削除/ }));
    const dialog = screen.getByRole('alertdialog');
    await user.click(within(dialog).getByRole('button', { name: '削除する' }));

    await waitFor(() => {
      expect(deleteCount).toBe(1);
    });
  });

  it('"キャンセル" ボタンをクリックするとダイアログが閉じ DELETE は呼ばれない', async () => {
    let deleteCount = 0;
    server.use(
      http.delete(`/api/v1/clinics/${CLINIC_ID}/lstep-settings`, () => {
        deleteCount++;
        return new HttpResponse(null, { status: 200 });
      })
    );

    await renderAndWait(configuredSyncOn);

    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: /設定削除/ }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'キャンセル' }));

    await waitFor(() => {
      expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
    });
    expect(deleteCount).toBe(0);
  });
});

// ─────────────────────────────────────────────────────────────
// D: 同期無効化確認ダイアログ
// ─────────────────────────────────────────────────────────────

describe('LstepSettingsForm — D: 同期無効化確認ダイアログ (FEAT-376)', () => {
  it('同期ON→OFF: Switch をクリックすると確認ダイアログが表示される', async () => {
    await renderAndWait(configuredSyncOn);
    const user = userEvent.setup();
    await user.click(screen.getByRole('switch', { name: '同期を有効にする' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
    expect(screen.getByText('同期を無効にしますか？')).toBeInTheDocument();
  });

  it('確認ダイアログの説明文が正しい', async () => {
    await renderAndWait(configuredSyncOn);
    const user = userEvent.setup();
    await user.click(screen.getByRole('switch', { name: '同期を有効にする' }));
    expect(
      screen.getByText(
        '同期を無効にすると Lステップへのタグ付与が停止します。よろしいですか？'
      )
    ).toBeInTheDocument();
  });

  it('"無効にする" をクリックするとダイアログが閉じ Switch が aria-checked="false" になる', async () => {
    await renderAndWait(configuredSyncOn);
    const user = userEvent.setup();
    await user.click(screen.getByRole('switch', { name: '同期を有効にする' }));

    const dialog = screen.getByRole('alertdialog');
    await user.click(within(dialog).getByRole('button', { name: '無効にする' }));

    await waitFor(() => {
      expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
    });
    expect(screen.getByRole('switch', { name: '同期を有効にする' })).toHaveAttribute(
      'aria-checked',
      'false'
    );
  });

  it('"キャンセル" をクリックするとダイアログが閉じ Switch は aria-checked="true" のまま', async () => {
    await renderAndWait(configuredSyncOn);
    const user = userEvent.setup();
    await user.click(screen.getByRole('switch', { name: '同期を有効にする' }));

    const dialog = screen.getByRole('alertdialog');
    await user.click(within(dialog).getByRole('button', { name: 'キャンセル' }));

    await waitFor(() => {
      expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
    });
    expect(screen.getByRole('switch', { name: '同期を有効にする' })).toHaveAttribute(
      'aria-checked',
      'true'
    );
  });

  it('同期OFF→ON: Switch をクリックするとダイアログなしで即座に aria-checked="true" になる', async () => {
    await renderAndWait(configuredSyncOff);
    const user = userEvent.setup();
    await user.click(screen.getByRole('switch', { name: '同期を有効にする' }));

    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
    expect(screen.getByRole('switch', { name: '同期を有効にする' })).toHaveAttribute(
      'aria-checked',
      'true'
    );
  });
});

// ─────────────────────────────────────────────────────────────
// F: CPMバージョン・休眠閾値 (Q19+Q21)
// ─────────────────────────────────────────────────────────────

describe('LstepSettingsForm — F: CPMバージョン・休眠閾値 (Q19+Q21 Phase C-2)', () => {
  it('"CPMバージョン" select が表示され defaultValue が選択されている', async () => {
    await renderAndWait({ ...configuredSyncOn, cpm_version: 'v2' });
    const select = screen.getByLabelText('CPMバージョン');
    expect(select).toBeInTheDocument();
    expect((select as HTMLSelectElement).value).toBe('v2');
  });

  it('休眠閾値 (180日) input の defaultValue が表示されている', async () => {
    await renderAndWait({ ...configuredSyncOn, dormant_prevention_180_days: 200 });
    const input = screen.getByLabelText('休眠予防閾値 (180日)');
    expect((input as HTMLInputElement).value).toBe('200');
  });

  it('保存すると PATCH body に cpm_version が含まれる', async () => {
    let capturedBody: Record<string, unknown> | null = null;
    server.use(
      http.patch(`/api/v1/clinics/${CLINIC_ID}/lstep-settings`, async ({ request }) => {
        capturedBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(configuredSyncOn);
      })
    );

    await renderAndWait({ ...configuredSyncOn, cpm_version: 'v1' });
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: '保存' }));

    await waitFor(() => {
      expect(capturedBody).not.toBeNull();
    });
    expect(capturedBody?.cpm_version).toBe('v1');
  });

  it('保存すると PATCH body に dormant_prevention_180_days が数値で含まれる', async () => {
    let capturedBody: Record<string, unknown> | null = null;
    server.use(
      http.patch(`/api/v1/clinics/${CLINIC_ID}/lstep-settings`, async ({ request }) => {
        capturedBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(configuredSyncOn);
      })
    );

    await renderAndWait({ ...configuredSyncOn, dormant_prevention_180_days: 180 });
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: '保存' }));

    await waitFor(() => {
      expect(capturedBody).not.toBeNull();
    });
    expect(capturedBody?.dormant_prevention_180_days).toBe(180);
  });
});

// ─────────────────────────────────────────────────────────────
// G: CPM V2 来院数閾値 (33ca50b2)
// ─────────────────────────────────────────────────────────────

describe('LstepSettingsForm — G: CPM V2 来院数閾値', () => {
  it('各閾値 input の defaultValue がサーバー値で表示される', async () => {
    await renderAndWait({
      ...configuredSyncOn,
      cpm_v2_coming_threshold: 3,
      cpm_v2_good_threshold: 6,
      cpm_v2_family_threshold: 10,
      cpm_v2_noah_threshold: 15,
    });
    expect((screen.getByLabelText('CPM V2 来院数閾値 (出会い→良好)') as HTMLInputElement).value).toBe('3');
    expect((screen.getByLabelText('CPM V2 来院数閾値 (良好→ファミリー)') as HTMLInputElement).value).toBe('6');
    expect((screen.getByLabelText('CPM V2 来院数閾値 (ファミリー→NOAH)') as HTMLInputElement).value).toBe('10');
    expect((screen.getByLabelText('CPM V2 来院数閾値 (NOAH)') as HTMLInputElement).value).toBe('15');
  });

  it('保存すると PATCH body に cpm_v2_coming_threshold が数値で含まれる', async () => {
    let capturedBody: Record<string, unknown> | null = null;
    server.use(
      http.patch(`/api/v1/clinics/${CLINIC_ID}/lstep-settings`, async ({ request }) => {
        capturedBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(configuredSyncOn);
      })
    );

    await renderAndWait({ ...configuredSyncOn, cpm_v2_coming_threshold: 2 });
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: '保存' }));

    await waitFor(() => {
      expect(capturedBody).not.toBeNull();
    });
    expect(capturedBody?.cpm_v2_coming_threshold).toBe(2);
  });

  it('閾値を変更して保存すると PATCH body に変更後の値が含まれる', async () => {
    let capturedBody: Record<string, unknown> | null = null;
    server.use(
      http.patch(`/api/v1/clinics/${CLINIC_ID}/lstep-settings`, async ({ request }) => {
        capturedBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(configuredSyncOn);
      })
    );

    await renderAndWait({ ...configuredSyncOn, cpm_v2_noah_threshold: 13 });
    const user = userEvent.setup();
    const noahInput = screen.getByLabelText('CPM V2 来院数閾値 (NOAH)');
    await user.clear(noahInput);
    await user.type(noahInput, '20');
    await user.click(screen.getByRole('button', { name: '保存' }));

    await waitFor(() => {
      expect(capturedBody).not.toBeNull();
    });
    expect(capturedBody?.cpm_v2_noah_threshold).toBe(20);
  });
});

// ─────────────────────────────────────────────────────────────
// H: CPM V1 判定閾値 (8ca5181b)
// ─────────────────────────────────────────────────────────────

describe('LstepSettingsForm — H: CPM V1 判定閾値', () => {
  it('各閾値 input の defaultValue がサーバー値で表示される', async () => {
    await renderAndWait({
      ...configuredSyncOn,
      cpm_v1_dormant_days: 240,
      cpm_v1_noah_ltv: 80000,
      cpm_v1_core_days: 180,
      cpm_v1_spot_min_amount: 30000,
      cpm_v1_growing_max_days: 90,
    });
    expect((screen.getByLabelText('CPM V1 休眠判定 (日数)') as HTMLInputElement).value).toBe('240');
    expect((screen.getByLabelText('CPM V1 NOAH LTV (円)') as HTMLInputElement).value).toBe('80000');
    expect((screen.getByLabelText('CPM V1 コア 在籍日数') as HTMLInputElement).value).toBe('180');
    expect((screen.getByLabelText('CPM V1 スポット 単回最大金額 (円)') as HTMLInputElement).value).toBe('30000');
    expect((screen.getByLabelText('CPM V1 グローイング 在籍最大日数') as HTMLInputElement).value).toBe('90');
  });

  it('保存すると PATCH body に cpm_v1_dormant_days が数値で含まれる', async () => {
    let capturedBody: Record<string, unknown> | null = null;
    server.use(
      http.patch(`/api/v1/clinics/${CLINIC_ID}/lstep-settings`, async ({ request }) => {
        capturedBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(configuredSyncOn);
      })
    );

    await renderAndWait({ ...configuredSyncOn, cpm_v1_dormant_days: 240 });
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: '保存' }));

    await waitFor(() => {
      expect(capturedBody).not.toBeNull();
    });
    expect(capturedBody?.cpm_v1_dormant_days).toBe(240);
  });

  it('閾値を変更して保存すると PATCH body に変更後の値が含まれる', async () => {
    let capturedBody: Record<string, unknown> | null = null;
    server.use(
      http.patch(`/api/v1/clinics/${CLINIC_ID}/lstep-settings`, async ({ request }) => {
        capturedBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(configuredSyncOn);
      })
    );

    await renderAndWait({ ...configuredSyncOn, cpm_v1_noah_ltv: 80000 });
    const user = userEvent.setup();
    const noahLtvInput = screen.getByLabelText('CPM V1 NOAH LTV (円)');
    await user.clear(noahLtvInput);
    await user.type(noahLtvInput, '100000');
    await user.click(screen.getByRole('button', { name: '保存' }));

    await waitFor(() => {
      expect(capturedBody).not.toBeNull();
    });
    expect(capturedBody?.cpm_v1_noah_ltv).toBe(100000);
  });
});

describe('LstepSettingsForm — I: 健診・予防タグ判定閾値', () => {
  it('各閾値 input の defaultValue がサーバー値で表示される', async () => {
    await renderAndWait({
      ...configuredSyncOn,
      health_prevention_lookback_days: 365,
      vaccine_deadline_days: 60,
    });
    expect((screen.getByLabelText('健診・来院履歴ルックバック日数') as HTMLInputElement).value).toBe('365');
    expect((screen.getByLabelText('ワクチン期限日数') as HTMLInputElement).value).toBe('60');
  });

  it('保存すると PATCH body に health_prevention_lookback_days が数値で含まれる', async () => {
    let capturedBody: Record<string, unknown> | null = null;
    server.use(
      http.patch(`/api/v1/clinics/${CLINIC_ID}/lstep-settings`, async ({ request }) => {
        capturedBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(configuredSyncOn);
      })
    );

    await renderAndWait({ ...configuredSyncOn, health_prevention_lookback_days: 365, vaccine_deadline_days: 60 });
    const user = userEvent.setup();
    await user.click(screen.getByRole('button', { name: '保存' }));

    await waitFor(() => {
      expect(capturedBody).not.toBeNull();
    });
    expect(capturedBody?.health_prevention_lookback_days).toBe(365);
    expect(capturedBody?.vaccine_deadline_days).toBe(60);
  });

  it('閾値を変更して保存すると PATCH body に変更後の値が含まれる', async () => {
    let capturedBody: Record<string, unknown> | null = null;
    server.use(
      http.patch(`/api/v1/clinics/${CLINIC_ID}/lstep-settings`, async ({ request }) => {
        capturedBody = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json(configuredSyncOn);
      })
    );

    await renderAndWait({ ...configuredSyncOn, health_prevention_lookback_days: 365, vaccine_deadline_days: 60 });
    const user = userEvent.setup();
    const lookbackInput = screen.getByLabelText('健診・来院履歴ルックバック日数');
    await user.clear(lookbackInput);
    await user.type(lookbackInput, '180');
    await user.click(screen.getByRole('button', { name: '保存' }));

    await waitFor(() => {
      expect(capturedBody).not.toBeNull();
    });
    expect(capturedBody?.health_prevention_lookback_days).toBe(180);
  });
});
