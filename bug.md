# 製品バグ台帳

更新日: 2026-09-05

## BUG-20260905-001: パスワードリセット申請が不正なメール形式を成功扱いにする

- **対象**: V05-3 / `/forgot-password`
- **再現**: メールアドレスに `a@` を入力して「リセットリンクを送信」を押す。
- **期待**: 形式エラーが表示され、送信されない。
- **実際**: 「パスワードリセットのリンクをメールに送信しました。メールをご確認ください。」と成功表示になる。
- **判定**: FAIL（不正形式を受理して成功表示するため）。
- **実施日**: 2026-09-05、`uat/20260905`、local Docker UAT。

## BUG-20260905-002: LIFF ヘルスカードが token なし URL を正常表示する

- **対象**: S12 手順1 / V05-5 手順1 / `/liff/health-card?clinic_id=1`
- **再現**: `token` を付けず、`clinic_id=1` だけで URL を開く。
- **期待**: 無効 URL としてエラー表示し、連携・個人情報画面を表示しない。
- **実際**: 「飼い主 テストユーザー ペット情報はありません」と正常画面が表示される。
- **判定**: FAIL（必須 link token 欠落を拒否していない）。
- **実施日**: 2026-09-05、`uat/20260905`、local Docker UAT。

## BUG-20260905-003: LINE予約設定 PUT が additional_fields 未指定の INSERT で 500

- **対象**: S04 前提 / `PUT /api/v1/clinics/:clinic_id/line-reservation-settings`
- **再現**: 対象 clinic に `line_reservation_settings` 行が無い（または INSERT 分岐に入る）状態で、valid な body（`status=running`, `time_slot_mode`, `no_staff_mode` 等）を送り、`additional_fields` を省略する。
- **期待**: デフォルト JSON（空配列/`[]`）を補完して 200/201 で保存される。
- **実際**: `500 internal server error`。Backend log: `ERROR: null value in column "additional_fields" of relation "line_reservation_settings" violates not-null constraint (SQLSTATE 23502)`。INSERT 値は `additional_fields=''`（jsonb として null）。
- **回避**: UAT では直接 SQL INSERT で行を先行作成。既存行への PUT（UPDATE）で `additional_fields` を明示すると 200。
- **判定**: FAIL（必須 jsonb の省略を永続化前にデフォルトせず 500 になる）。
- **実施日**: 2026-09-05、`uat/20260905`、local Docker UAT（MBPM3）。

新しい製品 FAIL は [`docs/ops/testing/TEST_ARCHITECTURE.md`](docs/ops/testing/TEST_ARCHITECTURE.md) の規則に従って追記します。env・seed・権限不足による BLOCKED は記録しません。対応済み項目は本ファイルから削除し、履歴は Git と `reports/uat-YYYY-MM-DD/` を参照します。
