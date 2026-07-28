# S12: LIFF ペットヘルスとアカウント連携

> **目的**: 飼い主が LIFF でアカウント連携（LINE アカウントと飼主レコードの紐付け）を完了でき、ペットヘルスページで自分のペットの健康情報のみが閲覧でき、他の飼主のデータが見えないことを納品前に証明する。
> **所要目安**: 15分 / **深度**: 薄い
> **仕様正本**: [line/architecture.md §2 認証・紐付けロジック](../../../spec/line/architecture.md)・[screens/04-owners-form.md §1.3](../../../spec/screens/04-owners-form.md)・[screens/38-liff-pet-health.md](../../../spec/screens/38-liff-pet-health.md)。実装参照: `frontend/liff/src/pages/PetHealthPage.tsx`・`backend/internal/reservation/liff_service_health_card.go`。

## 前提条件

- 環境: ローカル（seed 003_demo）または STG。**LIFF モックはローカル専用**: LIFF アプリ（frontend/liff）は `VITE_LIFF_MOCK=true`、バックエンドは `LIFF_MOCK=true` で LINE 認証をバイパスする。バックエンドは release モードでは `LIFF_MOCK` 設定を拒否して起動しない（release guard — `backend/internal/config/config.go`）ため、STG/本番では実 LINE アカウントで実施する。
- **モックの限界**: フロント `VITE_LIFF_MOCK=true` の連携ページは連携 API を呼ばず成功表示のみ返す（`frontend/liff/src/hooks/use-liff-link.ts` のモック分岐）。連携成立・409 の実機証明は、バックエンド `LIFF_MOCK` 経由での API 実行または STG 実 LINE で行う。
- 対象飼主: LINE 未連携（LINE User ID 未設定 — 飼主編集画面の連携セクションが「未連携」）の飼主 1 名。生存ペットが 1 頭以上おり、うち 1 頭にワクチン接種記録があること。
- 隔離確認用に、対象飼主とは別の飼主（ペットあり）をもう 1 名選んでおくこと（手順 7 で使用）。
- 連携トークン: 飼主編集画面の「LINE/Lステップ連携」セクションで、未連携時に表示される「連携用URLを発行」から `POST /api/v1/owners/:id/line/link-token` を実行する（`line_link_tokens` — [architecture.md §2-3](../../../spec/line/architecture.md)）。【要実測】発行後に token と clinic_id 付き LIFF URL が読み取り専用欄に表示され、コピーできること。
- 病院側確認用ログイン: reception ロール（飼主編集画面の連携セクション閲覧に使用）。
- 依存シナリオ: なし。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 発行したトークン付き URL（`?clinic_id=...&token=...`）で LIFF アプリを開く | 連携ページ（`LiffLinkPage`）が表示され、LINE 認証（LIFF ログイン）を経て連携処理が実行される（[architecture.md §2](../../../spec/line/architecture.md)） |
| 2 | 連携完了後、病院側の飼主編集画面を開く | 連携成功が表示され、「LINE/Lステップ連携」セクションの紐付け状況が「連携済み」になる（[04-owners-form.md §1.3](../../../spec/screens/04-owners-form.md)） |
| 3 | 連携済みの状態で再度連携（同じ飼主の新規トークン）を実行する | 409 で拒否され「このLINEアカウントはすでに連携済みです」が表示される（#Q22 実装。実装: `use-liff-link.ts` のステータス分岐） |
| 4 | 期限切れまたは使用済みトークンの URL で連携を試みる | 「リンクトークンが無効または期限切れです」（400）が表示され、連携は成立しない |
| 5 | token なしの URL（`?clinic_id=...` のみ）で LIFF アプリを開く | ペットヘルスページに切り替わり、ヘッダーに飼主名と LINE プロフィール画像が表示される（実装: `App.tsx` の token 有無分岐） |
| 6 | ペットカードの表示内容を確認する | ペットごとに、ペット名・種/品種・最終来院日（記録がない場合は「記録なし」）・ワクチン記録テーブル（ワクチン名/接種日/次回予定日、予定なしは「—」）が表示される。表示項目は実装確認済みの上記のみ — これ以外の健康情報（予防状況サマリ等）の表示有無は【要実測】 |
| 7 | 別の飼主（前提条件の 2 人目）の LINE アカウント／トークンで同様に連携し、ペットヘルスを開く | その飼主自身のペットのみが表示され、手順 6 の飼主のペットは一切表示されない — 飼主間隔離の実機証明 |
| 8 | `clinic_id` クエリなしの URL でペットヘルスを開く | 「クリニックIDが見つかりません」を含むエラー表示となり、他クリニック・他飼主のデータは表示されない（実装: `PetHealthPage.tsx`） |
| 9 | バックエンド停止などで健康記録の取得を失敗させる（ローカルのみ） | 「データ取得に失敗しました」と再試行ボタンが表示され、復旧後に再試行で一覧が表示される（実装: `use-fetch-state.ts` 共通フック） |

## 確認観点

- **飼主間隔離**: ヘルスカード API（`GET /api/liff/:clinicId/health-card`）は idToken → `line_customers` → `owner_id` の解決で自分の飼主レコードのペットのみを返す。pet_id 等をクエリで受けない設計のため、URL 操作で他の飼主のデータは閲覧できないこと（別の飼主で開き直して相互に見えないことを確認）。
- 連携成功は `audit_logs` に記録される（`line_link_service.go` — #Q22）— DB 参照は USER 実施。
- 未連携の LINE 顧客がペットヘルスを開いた場合、LINE 表示名＋「ペット情報はありません」となり、エラーにはならない（リンク前 UX — 実装仕様）。
- 連携トークンは単回使用（`line_link_handler.go` — 使用済みトークンは手順 4 の 400 になる）。トークン値・実 LINE アカウント情報を本ディレクトリやレポートに記録しないこと。
- ヘルスカードのレスポンスはフロントで zod スキーマ検証され、形状不正時は無音で欠落表示にならずエラーになる（実装: `frontend/liff/src/api/liff-api.ts`）。
- 死亡ペット（deceased_at 設定済み）はヘルスカードに表示されない（参照実装 `liff_service_health_card.go` の deceased 除外フィルタで確認済み。仕様文書には明記なし）。
- clinic_id 隔離: URL の clinicId と異なるクリニックの飼主・ペットが返らないこと。
- 本シナリオの LIFF アプリはペットヘルス・連携用（frontend/liff）であり、LINE 予約アプリ（line-reserve — [S04](S04-liff-reservation-journey.md)）とは別アプリ。予約機能の検証は S04 が正本。
