/**
 * EMR-487 Slice A: 新しい発行・再印刷を始められるか（回復の fence が使えるか）。
 *
 * fence / expected_recovery_generation / 新 key での再開は Slice B で、まだ存在しない。fence が無い間は、
 * 発見が 0 件・complete・成功/終結済みのみでも、遅れて到着する旧世代の要求が成功しうるため、
 * 新しい操作を開始してはならない（契約 §11）。この値は固定の定数で、env・server・storage による切替は持たない。
 * Slice B が fence を実装した変更で、この値と関門を一緒に更新する。
 * テストだけが vi.mock でこの値を差し替えて既存フローを検証する。
 */
export const CERTIFICATE_RECOVERY_FENCE_AVAILABLE = false;

export const RECOVERY_FENCE_PAUSED_TEXT =
  "回復の安全確認（fence）が利用可能になるまで、新しい発行・再印刷は一時停止しています。";
