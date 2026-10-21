// EMR-474 operation-recovery: 明示の次操作 fence（Slice B）が利用可能か。
// Slice A では常に false。発見の結果（0 件・complete・保留理由なし・経過時間）から新規送信を許可しない。
// 静的 constant のみ。環境変数・server 応答・storage・URL では切り替えない。
// true への変更は Slice B（control / closure / fence / 全 writer の wire が揃った単位）の別 PR でのみ行う。
// 既存の editor を使うページテストは vi.mock でこの module を true にして Slice B 相当の経路を検証する。
export const LIFF_REVISIT_INTAKE_RECOVERY_FENCE_AVAILABLE = false;
