# 取扱説明書（マニュアル）画面 仕様書 (Internal Manual)

## 概要
- **画面の目的**: システムを利用するスタッフが、各機能の操作方法や運用ルールをシステム内で直接確認できるようにする。スタッフが自力でマニュアルを編集・修正できる編集機能も提供する。
- **URLパターン**: `/manual`, `/manual/:category/:slug`
- **アクセス権限**:
    - **閲覧**: 認証済全ユーザー（閲覧制限なし）
    - **編集（DB 保存）**: `ResourceManualEdit` (`manual-edit`) リソースの `edit` 権限が必要
    - **編集履歴閲覧**: `ResourceManualEdit` の `view` 権限が必要
    - **オーバーライド削除（MD バンドル版に戻す）**: `ResourceManualEdit` の `delete` 権限が必要

---

## 画面構成

### 1. カテゴリ・ナビゲーション (左カラム)
業務の流れに沿って整理されたメニュー。
- **画面ガイド (Screens)**: 各画面の具体的な入力項目やボタンの意味。
- **業務フロー (Workflows)**: 「初診の受付」「急患時のカルテ保存」など、一連の業務手順。

### 2. 検索機能
キーワード入力により、膨大なマニュアル記事から該当箇所を瞬時に特定します（クライアントサイド Fuse.js 検索）。DB に保存されたオーバーライド版の内容も検索対象になります。

### 3. 本文エリア (右カラム)
構造化された Markdown 形式で表示。
- **視認性**: 実際のキャプチャ画像と臨床的な注意点を織り交ぜた解説。
- **印刷**: 右上の印刷アイコンから A4 印刷可能（サイドバー除外、本文のみ）。
- **編集**: 右上の編集アイコンから編集モードへ遷移。

### 4. 編集モード (`ManualEditor`)
編集アイコンクリックで開く編集 UI。
- **モード切替**: 編集 / 分割（既定） / プレビュー の 3 モード
- **ライブプレビュー**: 編集内容が即時レンダリングされ右側に表示
- **保存（DB）**: 「保存」ボタンで DB に upsert（`ResourceManualEdit` edit 権限が必要）
- **クリップボードコピー**: frontmatter 込みの完全な MD をクリップボードへ
- **.md ダウンロード**: ファイルとしてローカル保存
- **離脱ガード**: 未保存変更がある状態で閉じようとすると確認ダイアログ
- **未保存バッジ**: dirty 状態を視覚化

---

## 主要な機能

### 1. 二層構成のコンテンツ管理
マニュアル本文は以下の二層で管理されます：

| 保存先 | 役割 |
|--------|------|
| MD ファイル (`frontend/src/features/manual/content/`) | 初期版・既定値（Vite ビルド時にバンドル） |
| DB の `manual_articles` テーブル | 編集オーバーライド版（管理者編集分） |

**読み込み時のマージロジック**:
- `useGetManualArticleOverrides()` で DB のオーバーライドを取得
- `applyOverrides(bundledMD, overrides)` で同 slug を置換
- DB 取得失敗時は MD バンドル版が graceful degradation で表示される

### 2. 編集履歴（バージョン管理）
`manual_article_versions` テーブルに、編集ごとにスナップショットを記録：
- 編集者 (`edited_by_staff_id`)
- 編集時刻 (`edited_at`)
- 編集前後の全フィールド（title / order_value / section / body_markdown）
- `audit_logs` テーブルにも `manual_article.upsert` / `manual_article.delete` として記録

### 3. ビルド時解決 + ランタイム取得
- MD ファイル: Vite の `import.meta.glob` でビルド時バンドル（オフライン参照可）
- オーバーライド: ランタイムに API から取得し、バンドル版と merge

---

## 技術仕様

### Backend API
| メソッド | パス | 権限 | 用途 |
|---|---|---|---|
| GET | `/api/v1/manual/articles` | 認証済全員 | オーバーライド一覧取得 |
| GET | `/api/v1/manual/articles/:category/:slug` | 認証済全員 | 単一記事取得 |
| PUT | `/api/v1/manual/articles/:category/:slug` | `manual-edit` edit | 編集保存（upsert） |
| DELETE | `/api/v1/manual/articles/:category/:slug` | `manual-edit` delete | オーバーライド削除（MD 版に戻す） |
| GET | `/api/v1/manual/articles/:category/:slug/versions` | `manual-edit` view | 編集履歴一覧 |

### DB スキーマ
- `manual_articles` (`id`, `category`, `slug`, `title`, `order_value`, `section`, `body_markdown`, `updated_by_staff_id`, `created_at`, `updated_at`)
- `manual_article_versions` (`id`, `article_id`, `title`, `order_value`, `section`, `body_markdown`, `edited_by_staff_id`, `edited_at`)
- マニュアルは医院共通のため `clinic_id` を持たない（`clinics` / `accounts` 等と同様の例外扱い）

### 使用コンポーネント・モジュール
- **`ManualPage`**: ルートレイアウト、編集モード切替、DB オーバーライドのマージ。
- **`ManualSidebar`**: 左目次・ビューモード切替・検索ボックス。
- **`ManualContent`**: `react-markdown` + `remark-gfm` による Markdown 描画。
- **`ManualEditor`**: 編集 UI（textarea + プレビュー + 保存/コピー/ダウンロード）。
- **`useGetManualArticleOverrides` / `useUpsertManualArticle`**: React Query ベースの API hook。
- **`manual-index.ts`**: Frontmatter 解析・MD glob 取り込み・`applyOverrides()` マージロジック。
- **`use-manual-search.ts`**: Fuse.js 検索フック（マージ済記事リストを引数で受け取る）。

---
