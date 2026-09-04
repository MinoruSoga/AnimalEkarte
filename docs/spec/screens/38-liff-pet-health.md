# LIFF 診察券・健康情報 仕様書 (LIFF Pet Health)

## 概要
- **画面の目的**: 飼い主が LINE アプリ内（LIFF）で自身のペットの健康手帳（ワクチン接種記録・最終来院日）を閲覧する。また、スタッフが発行した紐付けトークン付き URL から LINE アカウントと院内の飼い主レコードを紐付ける。
- **URLパターン**: 独立エントリ `frontend/liff`（スタッフ画面とは別ビルド・別配信）。`https://liff.line.me/{LIFF ID}` 経由で起動し、クエリで動作が分岐する:
  - `?token=` あり → LINE アカウント紐付けフロー（`LiffLinkPage`）
  - `?token=` なし → 健康手帳表示（`PetHealthPage`）
  - `?clinic_id=` → 対象クリニックの指定（両フローで必須。LIFF エンドポイント URL 側に含めて運用する）
- **アクセス権限**: スタッフの権限リソースではなく **LINE ID Token 認証**。健康手帳 API は `LiffAuth` ミドルウェア（`liff_auth.go`）が Authorization: Bearer の ID Token を LINE の検証 API へ照合し、`line_customers` の顧客を特定・自動作成する。紐付け API は認証ヘッダ不要で、body 内の link_token + line_id_token による自己認証 + IP レートリミット（10回/分）。

---

## 画面構成

### 1. LoadingPage
LIFF SDK 初期化中に表示される全画面スピナー（`Spinner` + 「読み込み中...」）。

### 2. PetHealthPage（健康手帳）
- **ヘッダー**: LINE プロフィール画像と飼い主名。飼い主名は院内登録名（owner_name）を優先し、未紐付け時は LINE 表示名にフォールバック。
- **ペットカード**（飼い主に紐付く生存ペットごとに 1 枚）:

| 項目 | 内容 |
|:---|:---|
| ペット名・種/品種 | カードヘッダーに表示。 |
| 最終来院日 | 記録がない場合は「記録なし」。 |
| ワクチン記録 | ワクチン名・接種日・次回予定日のテーブル。次回予定日未設定は「—」。記録 0 件のペットではテーブル自体を非表示。 |

- **ペット 0 件時**: 「ペット情報はありません」を表示（LINE 未紐付けの顧客もこの状態になる — 後述）。
- **エラー時**: ステータス別メッセージと再試行ボタン。401（ID Token 失効）は再試行では解消しないため再試行ボタンを出さない（`resolveFetchError` の canRetry 判定）。

### 3. LiffLinkPage（LINE アカウント紐付け）
`useLiffLink` が返す status に応じて 6 状態を表示する。

| status | 表示 | 発生条件 |
|:---|:---|:---|
| loading / linking | スピナー + 「LINEアカウントを連携中...」 | LIFF 初期化中 / API 呼び出し中 |
| success | 「連携が完了しました」+ 閉じるボタン | 紐付け成功（204 No Content） |
| conflict | 「連携済みです」+ 閉じるボタン | 409: 対象の飼い主に既に LINE User ID が設定済み |
| expired | 「リンクが無効です」（再試行ボタンなし） | 400: トークン無効・期限切れ・クリニック不一致 |
| error | 「エラーが発生しました」+ 再試行ボタン | 401（LINE 認証失敗）・その他のエラー・`clinic_id`/`token` 欠落 |

---

## 主要な機能

### 1. 起動から表示までのシーケンス
1. モジュール評価時に `window.location.search` の `token` 有無でフローを分岐（`App.tsx`）。
2. `useLiff`（`frontend/src/shared-liff` の共有フック `use-liff.ts`）が `liff.init` を実行。未ログインなら `liff.login()` で LINE ログインへリダイレクト。
3. ログイン済みなら ID Token と LINE プロフィール（表示名・画像）を取得。初期化失敗時は `ErrorPage` を表示。
4. `PetHealthPage` が `fetchHealthCard`（`liff-api.ts`）で GET `/api/liff/:clinicId/health-card` を呼ぶ。フェッチ状態は共有フック `useFetchState` が管理。
5. サーバ側は `LiffAuth` が ID Token を検証: クリニックの LIFF ID からチャネル ID を抽出し、LINE 検証 API の client_id に使用（クロステナントのトークン流用を防止）。検証成功で `line_customers` を FindOrCreate。
6. レスポンスは zod スキーマ（`healthCardResponseSchema`）で形状検証してから描画。不正な形状は描画せずエラー画面に落とす。

### 2. LINE アカウント紐付けフロー
1. スタッフ側 API（POST `/api/v1/owners/:id/line/link-token`、`owners` の edit 権限）が 24 時間有効の単回使用トークンを発行し、`https://liff.line.me/{LIFF ID}?token={token}&clinic_id={clinicID}` 形式の URL を返す（`GenerateLineLinkToken` → `line_link_service.go`）。`LiffLinkPage` は `token` と `clinic_id` の両方をクエリから読むため、`clinic_id` 欠落時は「無効なURLです」で即エラーになる（SD-14 で修正・旧実装は `clinic_id` 欠落のまま発行していた）。飼い主情報画面（[04-owners-form.md](./04-owners-form.md)）の LINE/Lステップ連携セクション（`LineIntegrationCard`）の未連携時分岐に、この発行 API を呼ぶ `LineLinkTokenSection`（「連携用URLを発行」ボタン → 発行後は読み取り専用入力欄に URL 表示 + コピー ボタン、`useGenerateLineLinkToken` mutation 経由）を SD-14 で追加した。
2. 飼い主が URL を開くと `useLiffLink`（`use-liff-link.ts`）が LIFF 認証完了後に POST `/api/liff/:clinicId/link` へ link_token と LINE ID Token を送信。
3. サーバ側（`LinkLiffAccount`）は ①LINE ID Token 検証 → ②トークンの実在・期限・クリニック一致検証 → ③飼い主の既存 LINE User ID 有無チェック（既設定なら 409）→ ④LINE User ID 更新 → ⑤トークンの単回 CAS 消費 → ⑥監査ログ記録と更新結果の再取得、の順で処理する。④〜⑥は同一 transaction で実行し、いずれかが失敗した場合は飼い主更新・トークン消費・監査を全て rollback する。上書き・再リンクはこの endpoint では未対応で、strict JSON decode により `force` などの未知フィールドも拒否する。

### 3. 未紐付け時の挙動
health-card API は LINE 顧客が飼い主未紐付けでも 200 を返し、owner_name = LINE 表示名 + pets 空配列にフォールバックする（`liff_service_health_card.go`）。画面上は「ペット情報はありません」となり、紐付け前でもエラーにはならない。

### 4. 臨床安全・データ保護
- **死亡ペットの除外**: 死亡日（deceased_at）が設定されたペットは健康手帳に表示しない。
- **レスポンス形状検証**: zod による実行時検証で、想定外データの誤表示を防ぐ。
- **テナント分離**: ID Token の client_id 照合をクリニックごとの LINE チャネル ID で行い、トークンにはクリニック一致検証がある。
- **トークン保護**: 新規 raw token は 32 random bytes の unpadded base64url（43文字）で、DB には SHA-256 digest のみを保存する。24 時間期限・単回使用。64桁 hex raw token は、発行済み legacy 行の期限内検索だけに対応する。紐付け操作は監査ログに記録される。
- **レートリミット**: `/link` は 10回/分、読み取り系は 30回/分（IP ベース）。

### 5. アクセシビリティ
- エラーメッセージは role="alert" で通知。装飾絵文字（⚠️ / ✅ / ℹ️）は aria-hidden="true" でスクリーンリーダーから除外。
- プロフィール画像には表示名の alt を付与。

---

## 技術仕様

### 開発サーバー（`frontend/vite.config.ts`）
ローカル dev では `/liff/{clinicId}/src/*` を実ファイルへ rewrite する（`line-reserve` と同型）。rewrite が無いと `GET /liff/{clinicId}/src/main.tsx` が 503 になり画面が白紙になる。

### 使用コンポーネント
- **`PetHealthPage`** / **`LiffLinkPage`** / **`LoadingPage`**: `frontend/liff/src/pages` の 3 画面。
- **`ErrorBoundary`** / **`ErrorPage`** / **`Spinner`**: `frontend/src/shared-liff` の共有部品（line-reserve アプリと共用）。ルート全体を `ErrorBoundary` でラップ（`main.tsx`）。
- **`useLiff`**: LIFF SDK 初期化・ID Token / プロフィール取得の共有フック。
- **`useLiffLink`**: 紐付け API 呼び出しとステータス遷移。
- **`useFetchState`** + **`resolveFetchError`**: GET フェッチ状態管理とステータス別エラーメッセージ解決（`handle-fetch-error.ts`）。
- **モックモード**: `VITE_LIFF_MOCK=true`（FE、`liff-config.ts` の LIFF_MOCK）で LIFF SDK をバイパスしモックトークンで動作。バックエンドも環境変数 LIFF_MOCK=true（release モード以外）で認証をバイパスする。

### API連携
| メソッド | エンドポイント | 用途 | 認証方式 |
|:---|:---|:---|:---|
| GET | `/api/liff/:clinicId/health-card` | 健康手帳データ（飼い主名・ペット・ワクチン記録）の取得 | LINE ID Token（`LiffAuth`） |
| POST | `/api/liff/:clinicId/link` | link_token + line_id_token による飼い主への LINE User ID 紐付け | body 内トークン自己認証（10回/分） |
| POST | `/api/v1/owners/:id/line/link-token` | スタッフ側：紐付けトークンと LIFF URL の発行（[04-owners-form.md](./04-owners-form.md) の `LineLinkTokenSection` から呼び出し） | スタッフ JWT（`owners` / edit） |

---
