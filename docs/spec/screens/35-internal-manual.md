# 取扱説明書（システム内マニュアル） 仕様書 (Internal Manual)

## 概要
- **画面の目的**: システムの操作方法、運用ルール、およびトラブルシューティングをスタッフが院内でいつでも閲覧・検索できる統合ヘルプセンター。
- **URLパターン**: 
  - トップ: `/manual`
  - 記事詳細: `/manual/:category/:slug`
- **アクセス権限**: 認証済み全ユーザー（編集は `ResourceManualEdit` 権限者のみ）

---

## 1. 画面構成

### 1.1 カテゴリ・ナビゲーション (左パネル)
- **2 モード**: トップは `screens`（画面別）と `workflows`（業務フロー）。
- **グループ名**: 各 Markdown の frontmatter `section`。固定の「基本／診療業務…」ラベルではない。
- **閲覧**: 認証済みなら静的バンドルを読む。`GET /manual/articles` の override 取得は `ResourceManualEdit` があるときだけ。

### 1.2 記事ビューア (中央パネル)
- **Markdown レンダリング**: 表、リスト、コードブロック、画像を含むリッチなドキュメント表示（`react-markdown` + `remark-gfm`）。

---

## 2. 技術仕様（スタティック import システム）

本マニュアルは、高い保守性と高速な表示を両立するため、Vite の `import.meta.glob` を活用した静的インポートシステムを採用しています。

### 2.1 コンテンツ管理
- **格納場所**: `frontend/src/features/manual/content/` 配下の `.md` ファイル。
- **画像**: 同ディレクトリの `images/` 内に配置し、記事内から相対パスで参照。

### 2.2 検索エンジン
- **クライアントサイド検索**: 全記事のタイトル・セクション名・本文テキストを `fuse.js` で fuzzy search インデックス化し、瞬時の全文検索を実現。

---

## 3. 運用・編集フロー

### 3.1 記事の更新
1.  開発者が `docs/` 配下の正式仕様に基づき、マニュアル用 Markdown を作成。
2.  `src/features/manual/content/` へ配置。
3.  ビルド時に自動的にシステムへ取り込まれます。

### 3.2 DB 連携（編集オーバーライド）
静的 Markdown ファイルを基本としつつ、`ResourceManualEdit` 権限を持つユーザーはシステム画面上（`ManualEditor`）から直接編集が可能です。編集内容は `manual_articles` テーブルにオーバーライド版として保存され、次回表示時に MD ファイル版より優先して読み込まれます。

---

### API連携
| メソッド | エンドポイント | 用途 | 必須権限 | 必須アクション |
|:---|:---|:---|:---|:---|
| GET | `/api/v1/manual/articles` | 記事メタデータの一覧取得 | `manual-edit` | `view` |
| GET | `/api/v1/manual/articles/:category/:slug` | 特定の記事（コンテンツ）の取得 | `manual-edit` | `view` |
| GET | `/api/v1/manual/articles/:category/:slug/versions` | 特定の記事のバージョン履歴一覧取得（BE実装済みだがフロントエンドからは未呼出） | `manual-edit` | `view` |
| PUT | `/api/v1/manual/articles/:category/:slug` | 記事の追加・更新 | `manual-edit` | `edit` |
| DELETE | `/api/v1/manual/articles/:category/:slug` | 記事の削除 | `manual-edit` | `delete` |

---

