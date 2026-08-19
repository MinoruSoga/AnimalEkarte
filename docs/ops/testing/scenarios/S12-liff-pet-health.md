# S12: LIFF ペットヘルスとアカウント連携

> **目的**: 飼い主が LIFF でアカウント連携（LINE アカウントと飼主レコードの紐付け）を完了でき、ペットヘルスページで自分のペットの健康情報のみが閲覧でき、他の飼主のデータが見えないことを納品前に証明する。
> **所要目安**: 15分 / **深度**: 薄い
> **仕様正本**: [line/architecture.md §2 認証・紐付けロジック](../../../spec/line/architecture.md)・[screens/04-owners-form.md §1.3](../../../spec/screens/04-owners-form.md)・[screens/38-liff-pet-health.md](../../../spec/screens/38-liff-pet-health.md)。実装参照: `frontend/liff/src/pages/PetHealthPage.tsx`・`backend/internal/reservation/liff_service_health_card.go`。

## 前提条件

- 環境: ローカル（seed 003_demo）または STG。**LIFF モックはローカル専用**: LIFF アプリ（`frontend/liff`）は `VITE_LIFF_MOCK=true`（`frontend/liff/src/lib/liff-config.ts` + shared `use-liff.ts`）、バックエンドは `LIFF_MOCK=true` で LINE 認証をバイパスする。バックエンドは release モードでは `LIFF_MOCK` 設定を拒否して起動しない（release guard — `backend/internal/config/config.go` の `LIFF_MOCK must not be set in release mode`）ため、STG/本番では実 LINE アカウントで実施する。
- **モックの限界**: フロント `VITE_LIFF_MOCK=true` の連携ページは連携 API を呼ばず **約 800ms 後に success 表示のみ**返す（`use-liff-link.ts` の `LIFF_MOCK` 分岐 / `LINK_SUCCESS_DISPLAY_MS`）。連携成立・409 の実機証明は、バックエンド `LIFF_MOCK` 経由での API 実行または STG 実 LINE で行う。病院側 UI の「連携済み」表示はモック成功だけでは更新されない点に注意。
- 対象飼主: LINE 未連携（LINE User ID 未設定 — 飼主編集画面の連携セクションが「未連携」）の飼主 1 名。生存ペットが 1 頭以上おり、うち 1 頭にワクチン接種記録があること。
- 隔離確認用に、対象飼主とは別の飼主（ペットあり）をもう 1 名選んでおくこと（手順 7 で使用）。
- 連携トークン: 飼主編集画面の「LINE/Lステップ連携」セクションで、未連携時に表示される「連携用URLを発行」から `POST /api/v1/owners/:id/line/link-token` を実行する（`line_link_tokens` — [architecture.md §2-3](../../../spec/line/architecture.md)）。**runtime PASS（2026-08-01 browser）**: 発行後に `clinic_id` と `token` 付き LIFF URL が読み取り専用欄に表示され「コピー」可能（token 値は report 非掲載）。
- 病院側確認用ログイン: reception ロール（飼主編集画面の連携セクション閲覧に使用。`owners` edit で発行可）。
- 依存シナリオ: なし。**氏名+電話による owner 自動紐付けは行わない**（SEC-CS2-F02 — 予約 LIFF 側。本アプリの連携経路は link-token のみ）。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | `/liff/{clinicId}/`（末尾スラッシュ）で、発行したトークン付き URL（`?clinic_id=...&token=...`）を開く | Vite rewrite が無いと `GET /liff/{clinicId}/src/main.tsx` が 503 で白紙（BUG-017）。成功時は `LiffLinkPage`。`clinic_id` または `token` 欠落時は「無効なURLです。QRコードを再度読み取ってください」 |
| 2 | 連携完了後、病院側の飼主編集画面を開く | （実 API 連携時）連携成功が表示され、「LINE/Lステップ連携」セクションの紐付け状況が「連携済み」になる（[04-owners-form.md §1.3](../../../spec/screens/04-owners-form.md)）。FE モックのみでは病院側は未連携のまま |
| 3 | 連携済み UI で発行導線を探す / 解除後または API で再連携を試す | **連携済み UI には発行無し**。既連携 LINE での再連携は FE が 409 →「このLINEアカウントはすでに連携済みです」 |
| 4 | 期限切れまたは使用済みトークンの URL で連携を試みる | FE:「リンクトークンが無効または期限切れです。スタッフにお声がけください」（BE 400）。連携は成立しない |
| 5 | token なしの URL（`?clinic_id=...` のみ）で LIFF アプリを開く | ペットヘルスページに切り替わり、ヘッダーに飼主名（API の owner_name、未連携時は LINE 表示名）とプロフィール画像が表示される |
| 6 | ペットカードの表示内容を確認する | ペットごとに、ペット名・種/品種・最終来院日（記録がない場合は「記録なし」）・ワクチン記録テーブル（ワクチン名/接種日/次回予定日、予定なしは「—」）が表示される（`PetHealthPage.tsx`）。API は `GET /api/liff/:clinicId/health-card`（profile ではない）。**runtime 2026-08-01 BLOCKED**: 認証 401 でカード未描画 |
| 7 | 別の飼主（前提条件の 2 人目）の LINE アカウント／トークンで同様に連携し、ペットヘルスを開く | その飼主自身のペットのみが表示され、手順 6 の飼主のペットは一切表示されない — 飼主間隔離の実機証明 |
| 8 | `clinic_id` クエリなしの URL でペットヘルスを開く | 「クリニックIDが見つかりません」系エラー（`PetHealthPage` が clinic_id 必須 reject）。他テナントデータは出ない |
| 9 | バックエンド停止などで健康記録の取得を失敗させる（ローカルのみ） | 「データ取得に失敗しました」と再試行ボタン。401（ID Token 失効）は再試行ボタンを出さない |

## 確認観点

- **飼主間隔離**: ヘルスカード API（`GET /api/liff/:clinicId/health-card`）は idToken → `line_customers` → `owner_id` の解決で自分の飼主レコードのペットのみを返す。pet_id 等をクエリで受けない設計のため、URL 操作で他の飼主のデータは閲覧できないこと。
- 連携成功は `audit_logs` に記録される（`line_link_service.go` — #Q22）— DB 参照は USER 実施。
- 未連携の LINE 顧客がペットヘルスを開いた場合、LINE 表示名＋「ペット情報はありません」となり、エラーにはならない（リンク前 UX — `HealthCardResult` 空 pets）。
- 連携トークンは単回使用。トークン値・実 LINE アカウント情報を本ディレクトリやレポートに記録しないこと。
- ヘルスカードのレスポンスはフロントで zod スキーマ検証され、形状不正時は無音で欠落表示にならずエラーになる（`frontend/liff/src/api/liff-api.ts`）。
- 死亡ペット（`DeceasedAt != nil`）はヘルスカードに表示されない（`liff_service_health_card.go` で確認済み）。
- clinic_id 隔離: URL の clinicId と異なるクリニックの飼主・ペットが返らないこと。
- 本シナリオの LIFF アプリはペットヘルス・連携用（`frontend/liff`）であり、LINE 予約アプリ（line-reserve — [S04](S04-liff-reservation-journey.md)）とは別アプリ。予約機能の検証は S04 が正本。

## 実装突合
- 突合日: 2026-08-07
- HEAD: 844e43f69
- 変更:
  - `VITE_LIFF_MOCK` / BE release guard / モック 800ms success-only を再確認
  - エラー文言を FE 実装どおりに更新（400/409/無効 URL/clinic_id 欠落）
  - ヘルス API を `health-card` に統一（誤った profile 参照を除去）
  - SEC-CS2-F02（name+phone 自動紐付け無効）を前提に追記（S04 予約経路と同趣旨）
  - deceased 除外・token 分岐（`App.tsx`）は現行 main で一致

- runtime 2026-08-07: **PASS (API)** — `GET /api/liff/1/health-card` HTTP 200 body owner_name present (pets may be empty under mock). VITE_LIFF_MOCK=true in frontend container. Full UI not run (no login credentials in .env.local).
