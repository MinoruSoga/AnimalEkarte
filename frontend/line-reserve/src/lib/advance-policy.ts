// BUG-030: LINE 予約フローの advance（進む）ルールを一箇所に定義する。
// - selection-only ステップ: タップで即遷移してよいが、共通ヘルパー文言で「選択すると進む」を示す
// - 入力・複数選択・日付確定・最終確認: 明示的な主 CTA（次へ / 予約を確定する）

import type { FlowStepKey } from './step-progress';

export type FlowAdvanceStep = FlowStepKey;

export type AdvanceMode = 'auto-on-select' | 'explicit-cta';

/** 入力ステップ共通の主 CTA ラベル（最終確認を除く） */
export const EXPLICIT_PRIMARY_CTA_LABEL = '次へ';

/** 最終確認ステップの確定ラベル（既存文言を維持） */
export const FINAL_CONFIRM_CTA_LABEL = '予約を確定する';

/** auto-on-select ステップに必ず表示する共通ヘルパー */
export const AUTO_ADVANCE_HELPER_TEXT = '選択すると次の画面へ進みます';

const FLOW_STEP_ADVANCE_MODE: Record<FlowAdvanceStep, AdvanceMode> = {
  customerInfo: 'explicit-cta',
  courseSelect: 'auto-on-select',
  trimmingCourseSelect: 'auto-on-select',
  // 複数選択のため確定操作が必要
  trimmingOptionSelect: 'explicit-cta',
  staffSelect: 'auto-on-select',
  // カレンダーで選んだあと主 CTA で確定
  dateSelect: 'explicit-cta',
  timeSelect: 'auto-on-select',
  request: 'explicit-cta',
  confirm: 'explicit-cta',
};

export function getAdvanceMode(step: FlowAdvanceStep): AdvanceMode {
  return FLOW_STEP_ADVANCE_MODE[step];
}

export function getAutoAdvanceHelperText(step: FlowAdvanceStep): string | null {
  return getAdvanceMode(step) === 'auto-on-select' ? AUTO_ADVANCE_HELPER_TEXT : null;
}
