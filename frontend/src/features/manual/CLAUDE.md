# Manual Feature — 取扱説明書ビューア

スタッフ向け取扱説明書（マニュアル）をシステム内に組み込みで配信する Feature。

## アーキテクチャ

| 項目 | 採用 |
|------|------|
| ソース形式 | Markdown (frontmatter 付き) |
| 取り込み | Vite `import.meta.glob('*.md', { query: '?raw', eager: true })` |
| 解析 | 自前 frontmatter parser (`lib/manual-index.ts`) |
| レンダリング | `react-markdown` + `remark-gfm` (テーブル/タスクリスト対応) |
| 画像 | `import.meta.glob('*', { query: '?url', eager: true })` でハッシュ付き URL に解決 |
| 検索 | `fuse.js` fuzzy search (画面別/業務フロー横断) |
| 印刷 | `manual-print.css` で `@media print` ルール |
| ルーティング | `react-router` lazy load。`/manual/:category/:slug` |
| 認可 | permission gating 無し（認証済み全スタッフ閲覧可） |

## ディレクトリ構成

`ls -R features/manual` で再構築可能（Feature Indexing の標準構造 + `content/{screens,workflows,images}`）。

## 編集モード（DB 永続化対応）

ブラウザ上でマニュアルを編集し、**DB に保存** できる機能。

### 編集の保存方式（3 種類）

| 方式 | 必要権限 | 即時反映 | 用途 |
|------|--------|--------|------|
| 「保存」ボタン → DB 保存 | `ResourceManualEdit` edit | ✅ | 管理者の自力修正 |
| 「コピー」 → IT 担当へ | 不要 | × | 一般スタッフが編集案を作る |
| 「ダウンロード」 → IT 担当へ | 不要 | × | 同上 |

### バックエンド統合

- API クライアント: `api/get-manual-articles.ts`, `api/upsert-manual-article.ts`
- 取得: GET /api/v1/manual/articles（認証済全員、失敗時は空配列に fallback）
- 保存: PUT /api/v1/manual/articles/:category/:slug（`ResourceManualEdit` edit 権限）
- 履歴: GET /api/v1/manual/articles/:category/:slug/versions（同 view 権限）
- 削除: DELETE /api/v1/manual/articles/:category/:slug（同 delete 権限 = MD 版に戻す）

### 読み込み時のマージロジック

```
バンドル MD（screenArticles, workflowArticles）
       +
DB オーバーライド（useGetManualArticleOverrides）
       ↓
applyOverrides(base, overrides) で同 slug を置換
       ↓
ManualPage が最終リストを表示
```

DB が空の場合・取得失敗時は MD バンドル版が常に表示される（graceful degradation）。

### 編集機能のフロント実装

- `components/ManualEditor.tsx` — textarea + プレビュー + 編集 / 分割 / プレビューモード + 保存/コピー/ダウンロードボタン
- 編集 / 分割 / プレビュー の 3 モード切替
- 「保存」: `useUpsertManualArticle()` で PUT 送信 → 成功時にエディタを閉じる、失敗時はトースト
- 「コピー」: frontmatter 含む完全な MD をクリップボードへ
- 「ダウンロード」: 同上を .md ファイルとして保存

詳細運用は `content/workflows/27-manual-edit-request.md` 参照。

## Markdown 規約

### frontmatter (MANDATORY)

すべての MD ファイルは以下の frontmatter を含むこと。

```markdown
---
title: 画面/フロー名
order: 1
section: セクション名
---

# 本文タイトル

...
```

| キー | 型 | 必須 | 用途 |
|------|----|----|-----|
| `title` | string | ✅ | 目次・タブで表示される項目名 |
| `order` | number | ✅ | セクション内表示順 (昇順) |
| `section` | string | ✅ | サイドバーのグループ名 |

### セクション分類

**screens/**:
- 基本 (ログイン・概要)
- 診療業務 (受付・カルテ・検査・予防接種・トリミング・定期健診)
- 運用・管理 (会計・レジ締め・月次・入院・在庫・シフト)
- 外部連携 (LINE予約・Lステップ)
- システム設定 (マスタ設定)

**workflows/**:
- 来院対応 (新規飼主・既存飼主・入院)
- 日次・月次業務 (締め・集計)
- 定期業務 (ワクチン・在庫)
- 外部連携運用 (LINE×Lステップ)
- 管理者向け (スタッフ登録・トラブル対応)

## ファイル命名規則

```
NN-kebab-case.md     例: 01-login.md, 03-hospitalization-flow.md
```

- `NN` は 2 桁の表示順 (frontmatter の `order` と整合)
- 拡張子は `.md` のみ。frontmatter の `title` がスラグではなくファイル名のスラグ部分が URL になる

## 新規マニュアル項目の追加手順

1. **MD ファイル作成**
   - `content/screens/` または `content/workflows/` に `NN-name.md` を作成
   - frontmatter を必ず記載

2. **スクリーンショット添付（任意）**
   - `content/images/NN-name.png` に画像を配置
   - MD 本文では `![alt](images/NN-name.png)` で参照
   - `ManualContent.tsx` の `resolveImageSrc` が Vite URL に自動置換

3. **ビルド時自動反映**
   - Vite の `import.meta.glob` が新規ファイルを検出
   - HMR で即時反映。再起動不要

4. **検索インデックスの更新**
   - 自動。Fuse.js が `allArticles` を全件再構築

## 画像の取り扱い

- **形式**: PNG / JPEG 推奨
- **配置**: `content/images/` 直下に置く（サブディレクトリ非対応）
- **参照**: `![alt](images/filename.png)` または `![alt](./images/filename.png)`
- **解像度**: 横幅 1280-1920px 程度。長辺 2000px を超える場合は圧縮推奨
- **マスキング**: 本番デプロイ前に、画面に映る個人情報・API キー等を必ずマスキングすること

## CSS ルール

- design-tokens 必須。hex 直書き禁止 → `frontend/CLAUDE.md` の規約に従う
- 印刷時に隠したい要素には `no-print` クラスを付与
- 印刷対象本文には `manual-article` クラスを付与（既に `ManualContent.tsx` で付与済）

## React パターン

- **Hooks ルール厳守**: ManualPage 内で hooks を early return より前に配置すること
- **関数宣言**: `export function ManualPage()` （FC・forwardRef 禁止）
- **Conditional Render**: `&&` 禁止 → `? : null` を使用

## 禁止事項

- ❌ MD 内で生 HTML を多用しない（react-markdown でレンダリングできない要素は避ける）
- ❌ frontmatter なしの MD を追加しない（目次に表示されない）
- ❌ `images/` 以外のディレクトリに画像を置かない（自動解決されない）
- ❌ 動的な MD 読み込み（fetch 等）を実装しない。ビルド時静的解決を維持

## テスト

- テストは対象ファイルと同階層に配置する（`__tests__/` ディレクトリは使わない）。例: `lib/parse-frontmatter.test.ts` — frontmatter parser の単体テスト
- 統合テスト: 必要に応じて `routes/ManualPage.test.tsx` のようにルート/コンポーネントと同階層に追加
- Vitest を使用: `docker compose exec frontend pnpm test:run -- src/features/manual`

## 依存パッケージ

`react-markdown` / `remark-gfm` / `fuse.js`（バージョンは `frontend/package.json` 参照）。追加時は `docker compose exec frontend pnpm install` を実行。

## 将来の拡張候補

- **多言語対応**: `content/ja/`, `content/en/` 構造への切替（現在は日本語固定）
- **PDF 一括エクスポート**: 全項目を結合した PDF 生成（要件確定後）
- **動画埋め込み**: 操作デモ動画の embed 対応
- **編集者ロール**: 管理者が UI 上で MD 編集できる機能（要件確定後）
- **検索ハイライト**: 検索結果のヒット箇所をハイライト表示

## トラブルシューティング

| 症状 | 原因と対処 |
|------|----------|
| 新規 MD が目次に出ない | frontmatter 欠落 → 必須キー (`title`/`order`/`section`) を確認 |
| 画像が表示されない | 配置場所が `content/images/` 直下になっているか、参照パスが `images/xxx.png` 形式か確認 |
| 検索ヒットしない | `searchText` は markdown 記号除去後の文字列。記号のみのクエリは無視される |
| HMR で更新が反映されない | Vite 開発サーバを再起動 (`docker compose restart frontend`) |
| 印刷時にサイドバーが残る | `no-print` クラスが適切に付与されているか、印刷対象が `.manual-root` 配下か確認 |
