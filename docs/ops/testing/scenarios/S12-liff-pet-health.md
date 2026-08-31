# S12: LIFF ペットヘルスとアカウント連携

> **目的**: 飼い主が LIFF でアカウント連携（LINE アカウントと飼主レコードの紐付け）を完了でき、ペットヘルスページで自分のペットの健康情報のみが閲覧でき、他の飼主のデータが見えないことを納品前に証明する。
> **所要目安**: 15分 / **深度**: 薄い
> **仕様正本**: [line/architecture.md §2 認証・紐付けロジック](../../../spec/line/architecture.md)・[screens/04-owners-form.md §1.3](../../../spec/screens/04-owners-form.md)・[screens/38-liff-pet-health.md](../../../spec/screens/38-liff-pet-health.md)。実装参照: `frontend/liff/src/pages/PetHealthPage.tsx`・`backend/internal/reservation/liff_service_health_card.go`。

## 前提条件

- ローカル mock lane または USER 管理の実 LINE lane を使う。mock は release mode で禁止され、連携成功画面だけでは backend link 成立を証明しない。
- 使い捨て clinic に owners edit 権限を持つ attached account、未連携 owner A と生存 pet（ワクチン記録あり）、隔離確認用 owner B/pet を作成する。
- URL コピー手順には reservation LIFF ID または L-step LIFF ID が設定済みであることが必須。どちらも未設定の分岐も別途確認する。
- 実 LINE lane では designated synthetic fixture と cleanup 手順を使う。氏名+電話による自動紐付けは使わない。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 0 | reservation LIFF ID と L-step LIFF ID の両方を未設定にして URL 発行を試す | token 作成自体は成功し得るが URL は空。画面に「LIFF IDが未設定…」が表示され、コピー可能 URL があるとは主張しない。以降の URL-copy 手順ではいずれかの LIFF ID を設定する |
| 1 | `/liff/{clinicId}/`（末尾スラッシュ）で、発行したトークン付き URL（`?clinic_id=...&token=...`）を開く | Vite rewrite が無いと `GET /liff/{clinicId}/src/main.tsx` が 503 で白紙（BUG-017）。成功時は `LiffLinkPage`。`clinic_id` または `token` 欠落時は「無効なURLです。QRコードを再度読み取ってください」 |
| 2 | 連携完了後、病院側の飼主編集画面を開く | （実 API 連携時）連携成功が表示され、「LINE/Lステップ連携」セクションの紐付け状況が「連携済み」になる（[04-owners-form.md §1.3](../../../spec/screens/04-owners-form.md)）。FE モックのみでは病院側は未連携のまま |
| 3 | 連携済み UI で発行導線を探す / 解除後または API で再連携を試す | **連携済み UI には発行無し**。既連携 LINE での再連携は FE が 409 →「このLINEアカウントはすでに連携済みです」 |
| 4 | 期限切れまたは使用済みトークンの URL で連携を試みる | FE:「リンクトークンが無効または期限切れです。スタッフにお声がけください」（BE 400）。連携は成立しない |
| 5 | token なしの URL（`?clinic_id=...` のみ）で LIFF アプリを開く | ペットヘルスページに切り替わり、ヘッダーに飼主名（API の owner_name、未連携時は LINE 表示名）とプロフィール画像が表示される |
| 6 | ペットカードの表示内容を確認する | ペットごとに、ペット名・種/品種・最終来院日（記録がない場合は「記録なし」）・ワクチン記録テーブル（ワクチン名/接種日/次回予定日、予定なしは「—」）が表示される（`PetHealthPage.tsx`）。API は `GET /api/liff/:clinicId/health-card`（profile ではない）。 |
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
- 変更:
  - `VITE_LIFF_MOCK` / BE release guard / モック 800ms success-only を再確認
  - エラー文言を FE 実装どおりに更新（400/409/無効 URL/clinic_id 欠落）
  - ヘルス API を `health-card` に統一（誤った profile 参照を除去）
  - SEC-CS2-F02（name+phone 自動紐付け無効）を前提に追記（S04 予約経路と同趣旨）
  - deceased 除外・token 分岐（`App.tsx`）は現行 main で一致
