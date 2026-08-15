# ステージング環境・フロントエンド検証手順書 (Vercel Frontend Staging Test)

> **目的**: デプロイ後のFE(UI・ログイン・API連携)検証手順を定義する。
> **読者**: デプロイ担当。
> **タイミング**: デプロイ直後のFE確認時。

> **Animal Ekarte**: Vercel デプロイ完了後のフロントエンド・UI・API 連携の検証手順
> **最新更新**: 2026-07-10 | **目的**: 本番リリース前のフロントエンド稼働確認

---

## 1. 目的と対象読者

本ドキュメントは、Vercel へのフロントエンドデプロイ完了直後、および STG 環境 API デプロイ後に、UI・ログイン・API 連携が正常に機能していることを確認するための手順書です。

**対象者**: DevOps / SRE / Team Lead / Frontend Engineer  
**実施頻度**: デプロイ直後（推奨：毎デプロイ時）  
**推定所要時間**: 15 分

---

## 2. Vercel デプロイ状態確認

### 2.1 Vercel Deployments ダッシュボード確認

1. [Vercel Project Dashboard](https://vercel.com/animalekarte/animalekarte-frontend) にアクセス
2. **Deployments** タブを開く
3. 最新デプロイの状態確認：
   - **Status**: `Ready` であることを確認（青色チェックマーク）
   - **Build Log**: エラーがないことを確認
   - **Deployment Time**: デプロイ完了時刻を記録

**失敗時アクション**:
- Status が `Building` → デプロイ進行中。数分待機。
- Status が `Error` → Build Log を確認。[トラブルシューティング §7.1](#71-vercel-ビルドエラー) を参照。
- Status が `Canceled` → デプロイキャンセル。GitHub Actions の再実行検討。

### 2.2 デプロイメント履歴確認

```bash
# 過去 7 日間のデプロイを確認（任意）
# Vercel CLI: vercel list --limit 10
```

**確認項目**:
- 直近 3 デプロイが `Ready` 状態
- ロールバック発生の有無

---

## 3. ブラウザでのフロントエンド動作確認

### 3.1 ページ読み込み確認

**環境**: [https://stg.noah-karte.com](https://stg.noah-karte.com)

1. ブラウザで URL を開く
2. ページ読み込み完了を確認：
   - [ ] ページが 5 秒以内に読み込まれる
   - [ ] ログイン画面が表示される（デフォルト未認証状態）
   - [ ] "Animal Ekarte" ロゴが表示される

**失敗時アクション**:
- Page 読み込み > 10 秒 → [§7.4 Vercel キャッシュ無効化](#74-vercel-キャッシュ無効化) を確認
- White screen / CSS なし → Build error 発生の可能性。[トラブルシューティング](#7-トラブルシューティング) へ。

---

### 3.2 ログイン画面表示確認

1. ログイン画面要素の確認：
   - [ ] Email 入力フィールド表示
   - [ ] Password 入力フィールド表示
   - [ ] "ログイン" ボタン表示
   - [ ] CSS / Tailwind スタイルが正しく適用（見た目の崩れなし）

**確認ポイント**:
- フォーム要素の背景色、テキスト色が正常
- レスポンシブデザイン確認（PC/tablet/mobile で確認推奨）

**失敗時アクション**:
- CSS が反映されていない → [§7.2 CSS 読み込み失敗](#72-css-読み込み失敗) を参照
- Form validation メッセージが表示 → バリデーション正常動作（期待動作）

---

### 3.3 ブラウザ DevTools コンソール確認

1. ブラウザを F12 で開く → **Console** タブ
2. エラーメッセージの有無確認：
   - [ ] **Error** ログなし（警告は許容）
   - [ ] **Network errors** なし

**許容される警告** (無視OK):
- shadcn/ui のコンポーネント初期化時の INFO ログ
- Next.js dev server 接続失敗（STG は production build）

**即座に報告すべきエラー**:
- `Uncaught TypeError`, `Uncaught ReferenceError`
- `Failed to fetch from /api/...`（API 接続エラー）

---

### 3.4 ネットワークリクエスト確認

1. DevTools → **Network** タブを開く
2. ページをリロード（Ctrl+R / Cmd+R）
3. リクエスト確認：
   - [ ] 主要 JS バンドル（`app.js` など）: Status **200**
   - [ ] CSS: Status **200**
   - [ ] 画像アセット: Status **200** または **304** (キャッシュ)
   - [ ] API リクエスト（ある場合）: Status **2xx** または **401** (未認証OK)

**異常基準**:
- JS/CSS が **404** → ビルド出力エラー
- API が **503** または **Connection timeout** → Backend API 非応答（[README.md §4.1](./README.md#41-ヘルスチェック手順) で Backend 確認）
- リクエストの大半が **5xx** → サーバーエラー

---

## 4. Demo アカウントログイン検証

### 4.1 ログイン操作

**アカウント情報**: Stone に保存済み（[CRUD-SMOKE-TEST.md](./CRUD-SMOKE-TEST.md) §2.2 参照）

**手順**:
1. ブラウザで [https://stg.noah-karte.com](https://stg.noah-karte.com) にアクセス
2. Email フィールドに demo account email を入力
3. Password フィールドに password を入力（本ドキュメントに記載しない）
4. "ログイン" ボタンをクリック

**期待結果**:
- ログイン直後、リダイレクト後、ダッシュボード（`/dashboard` or `/`）が表示
- ユーザー名またはクリニック名がヘッダーに表示

**失敗時アクション**:
- "Invalid email or password" → 認証情報確認。Stone を再確認。
- 401 Unauthorized → [§7.3 API 接続エラー](#73-api-接続エラー) を参照。
- リダイレクトループ → Cookie / Token 期限切れの可能性。§7.3 を参照。

---

### 4.2 Cookie / Token 確認

1. DevTools → **Application** タブ → **Cookies**
2. ドメイン: `stg.noah-karte.com` を選択
3. 確認項目：
   - [ ] `access_token` が存在（値を本ドキュメントに記載しない）
   - [ ] `refresh_token` が存在
   - [ ] 両者の `Expires` が未来日時

**異常基準**:
- Cookie がない → ログイン API 失敗
- Cookie の期限が過去 → セッション期限切れ

---

## 5. Settings 画面へのアクセス確認

### 5.1 医院マスタ・権限グループ・スタッフマスタ画面へのアクセス

ログイン後、画面ヘッダーまたはサイドバーから **Settings** へアクセス。

#### Settings 画面一覧：
- [ ] **医院マスタ (Clinic Master)** → `/settings/clinic-master` or 類似
- [ ] **権限グループ (Permission Groups)** → `/settings/permission-groups` or 類似
- [ ] **スタッフマスタ (Staff Master)** → `/settings/staff-master` or 類似

**確認ポイント**:
- 各画面が正常に読み込まれ、テーブル/リストが表示される
- CSS が正しく適用されている
- フォーム（追加/編集）が表示される（権限に基づく）

**失敗時アクション**:
- 403 Forbidden → デモアカウント権限不足。account 確認。
- 404 Not Found → ルーティング設定エラー。[FE build log](https://vercel.com/animalekarte/animalekarte-frontend) を確認。
- ページ読み込み遅延（> 5s） → API 遅延の可能性。[README.md §4.1](./README.md#41-ヘルスチェック手順) で Backend health check。

---

### 5.2 Settings 画面での読み書き権限確認

**医院マスタ（例）**:
1. Settings > 医院マスタ をクリック
2. 既存医院をクリックして詳細表示
3. 編集可能であることを確認（フォーム入力可能 OR 編集ボタン活性化）
4. **保存ボタン**をクリック（実際には Save しなくても OK）

**期待結果**:
- フォーム入力フィールドが読み書き可能
- 編集ボタン / 削除ボタンが活性化（権限に応じて）

**失敗時アクション**:
- すべてのフィールドが読み取り専用 → Demo account 権限不足
- Save 時に API 503 → Backend API 非応答

---

## 6. API 連携確認

### 6.1 Backend API 疎通確認

ブラウザの DevTools Network タブで、Settings 画面へのアクセス時に API リクエストを確認。

```bash
# 例: Clinics リスト API
GET https://api.stg.noah-karte.com/api/v1/clinics
Authorization: Bearer <access_token>
```

**期待結果**:
- Status: **200 OK**
- Response: clinic list（JSON 配列）

**失敗時アクション**:
- Status: **503 Service Unavailable** → [README.md §4.1](./README.md#41-ヘルスチェック手順) で Backend health check
- Status: **401 Unauthorized** → Token 期限切れ。ページリロードして再ログイン。
- Status: **CORS Error** → CORS 設定エラー。Backend CORS 設定確認。

---

### 6.2 API レスポンスタイム確認

DevTools → Network タブで、API リクエストの **Time** 列を確認。

**期待値**:
- API レスポンスタイム < 2s（通常は 100-500ms）

**異常基準**:
- レスポンスタイム > 5s → Database query 遅延 / Backend リソース不足。ログで確認（Cloudflare 正系統は Workers Logs、旧 ECS ロールバック経路は CloudWatch）。
- Connection timeout → Backend 停止。Cloudflare 正系統は `/health` 疎通確認、旧 ECS ロールバック経路は ECS status 確認。

---

## 7. トラブルシューティング

### 7.1 Vercel ビルドエラー

**症状**: Vercel Deployments ダッシュボードで Status = `Error`

**診断**:
1. Vercel ダッシュボード → Deployments → 最新デプロイをクリック
2. Build Logs を確認：
   - `npm run build` のエラーメッセージを探す
   - TypeScript type error / build tool error など

**対処**:
- **Type error**: `pnpm type-check` をローカルで実行。型の不一致を修正。
- **Build tool error**: `pnpm build` をローカルで実行してエラーを再現。
- **Dependency error**: `pnpm install` でロックファイルを再生成。

**既知ビルドエラー**:
- `Cannot find module '@app/...'`: path alias 設定確認（`tsconfig.json`）
- `Tailwind CSS not found`: `globals.css` import 確認

---

### 7.2 CSS 読み込み失敗

**症状**: 
- ページが白い背景でテキストのみ表示
- Tailwind スタイルがない

**診断**:
1. DevTools → Network → CSS リクエストを探す
2. CSS 状態コード確認：
   - `404` → CSS ファイルが生成されていない
   - `200 but blank` → CSS 内容が空

**対処**:
- [§7.4 Vercel キャッシュ無効化](#74-vercel-キャッシュ無効化) を実施
- ブラウザキャッシュクリア: Ctrl+Shift+Delete / Cmd+Shift+Delete
- Vercel deployment ページを再チェック（Status = `Ready` 確認）

---

### 7.3 API 接続エラー

**症状**:
- Console に `Failed to fetch from /api/...`
- Settings 画面でデータ読み込みなし

**診断**:
```bash
# API health check
curl -s https://api.stg.noah-karte.com/health | jq '.status'
```

**対処**:
- Backend health check が `ok` → CORS 設定確認（Backend）
- Backend health check が失敗 → [README.md §4.1](./README.md#41-ヘルスチェック手順) でロールバック判定

---

### 7.4 Vercel キャッシュ無効化

> **訂正(2026-07-10)**: `../infra/architecture.md` の構成図が示す通り、CloudFront は
> バックエンド API（`api.noah-karte.com`）専用であり、Vercel フロントエンド（`stg.noah-karte.com`）は
> CloudFront を経由しない。旧「CloudFront キャッシュ無効化」の記述はアーキテクチャと矛盾していたため訂正する。

Vercel デプロイ直後に CSS/JS が古いバージョンで配信される場合、Vercel はビルド成果物をコンテンツハッシュ付き
ファイル名で配信するためサーバ側キャッシュの明示的無効化は通常不要。古いバージョンが表示される場合は、
まずブラウザキャッシュのクリア、または Vercel ダッシュボードで最新デプロイが `Promote to Production` /
対象環境に正しく反映されているかを確認する。

**期待結果**:
- ブラウザキャッシュクリア後、最新アセットが配信される

---

### 7.5 ブラウザキャッシュクリア手順

**Chrome**:
- Ctrl+Shift+Delete → Storage タブ → "Clear data"
- 対象: Cookies and other site data / Cached images and files

**Safari**:
- Develop → Clear Caches

**Firefox**:
- Ctrl+Shift+Delete → キャッシュ削除

---

## 8. 認証情報保護ポリシー

本ドキュメント作成および検証時、以下の情報は **絶対に記載しないこと**：

- ❌ `password` （パスワード）
- ❌ `access_token` / `refresh_token` （トークン値）
- ❌ Cookie の `Set-Cookie` ヘッダ値
- ❌ Demo account の email（実際のメアド）

**実装方法**:
- Stone に保存した認証情報のみを参照
- curl / API テストは環境変数 (`${TOKEN}`) で実行
- スクリーンショット / ログ出力には認証情報を含めない（デフォルトサニタイズ）
- 本ドキュメント完成後、他の developer と共有する際は敏感情報チェック

---

## 9. 参考資料

- [README.md](./README.md) - デプロイメント・運用ドキュメント（ロールバック判定、コマンド集）
- [STG-CONTINUOUS-OPERATIONS.md](./STG-CONTINUOUS-OPERATIONS.md) - 日常運用チェックリスト
- [CRUD-SMOKE-TEST.md](./CRUD-SMOKE-TEST.md) - CRUD 操作テスト詳細
- [CI-CD-PIPELINE.md](./CI-CD-PIPELINE.md) - デプロイパイプライン
- Vercel Docs: [Deployments](https://vercel.com/docs/concepts/deployments/overview)

