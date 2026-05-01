import { describe, it, expect } from 'vitest';
import { readFileSync } from 'fs';
import { resolve } from 'path';

const formSource = readFileSync(
  resolve(process.cwd(), 'src/features/settings/integrations/lstep/LstepSettingsForm.tsx'),
  'utf-8'
);

const hooksSource = readFileSync(
  resolve(process.cwd(), 'src/features/settings/integrations/lstep/hooks/useLstepSettings.ts'),
  'utf-8'
);

/**
 * FEAT-376: Lステップ設定画面 — 同期ON/OFF・状態表示
 *
 * ソースコード検査による静的解析テスト。
 * - Switch コンポーネントの使用
 * - is_sync_enabled の型定義とPATCHリクエスト送信
 * - sync_enabled_at の表示
 * - 同期状態バッジ（同期中 / 同期停止中）
 * - 同期OFF = 一時停止であることを示すコピー
 */

describe('LstepSettingsForm — 同期ON/OFFトグル (FEAT-376)', () => {
  it('Switch コンポーネントをインポートしている', () => {
    expect(formSource).toContain("import { Switch } from");
  });

  it('Switch が checked と onCheckedChange で制御されている', () => {
    expect(formSource).toContain('checked={isSyncEnabled}');
    expect(formSource).toContain('onCheckedChange={setSyncEnabledOverride}');
  });

  it('aria-label="同期を有効にする" が設定されている（アクセシビリティ）', () => {
    expect(formSource).toContain('aria-label="同期を有効にする"');
  });

  it('syncEnabledOverride を undefined で初期化し、サーバー値にフォールバックする', () => {
    expect(formSource).toContain('useState<boolean | undefined>(undefined)');
    expect(formSource).toContain('syncEnabledOverride ?? settings?.is_sync_enabled');
    expect(formSource).toContain('?? false');
  });

  it('PATCH リクエストに is_sync_enabled を常に含める', () => {
    expect(formSource).toContain('req.is_sync_enabled = isSyncEnabled');
  });

  it('保存成功後に syncEnabledOverride をリセットする', () => {
    expect(formSource).toContain('setSyncEnabledOverride(undefined)');
  });
});

describe('LstepSettingsForm — 同期状態バッジ (FEAT-376)', () => {
  it('"同期中" バッジテキストが存在する', () => {
    expect(formSource).toContain('同期中');
  });

  it('"同期停止中" バッジテキストが存在する', () => {
    expect(formSource).toContain('同期停止中');
  });

  it('同期状態バッジは isConfigured の場合のみ表示される', () => {
    // isConfigured を条件に同期バッジを出す構造
    expect(formSource).toContain('{isConfigured ? (');
    expect(formSource).toContain('同期中');
    expect(formSource).toContain('同期停止中');
  });
});

describe('LstepSettingsForm — sync_enabled_at 表示 (FEAT-376)', () => {
  it('sync_enabled_at を条件付きで表示する', () => {
    expect(formSource).toContain('settings?.sync_enabled_at');
  });

  it('formatSyncDate ヘルパーで日付を日本語ローカライズしている', () => {
    expect(formSource).toContain('formatSyncDate');
    expect(formSource).toContain('"ja-JP"');
  });

  it('同期有効化日時のラベルが表示される', () => {
    expect(formSource).toContain('同期有効化日時');
  });
});

describe('LstepSettingsForm — 同期OFF = 一時停止のコピー (FEAT-376)', () => {
  it('OFFは連携処理の一時停止であることを説明するコピーがある', () => {
    expect(formSource).toContain('OFFにすると連携処理を一時停止します');
  });

  it('APIキー設定は保持されることを明示している', () => {
    expect(formSource).toContain('APIキー設定は保持されます');
  });
});

describe('useLstepSettings — 型定義 (FEAT-376)', () => {
  it('LstepSettingsResponse に is_sync_enabled が定義されている', () => {
    expect(hooksSource).toContain('is_sync_enabled: boolean');
  });

  it('LstepSettingsResponse に sync_enabled_at が定義されている', () => {
    expect(hooksSource).toContain('sync_enabled_at: string | null');
  });

  it('LstepSettingsRequest に is_sync_enabled が定義されている', () => {
    expect(hooksSource).toMatch(/LstepSettingsRequest[\s\S]*?is_sync_enabled\?:\s*boolean/);
  });
});
