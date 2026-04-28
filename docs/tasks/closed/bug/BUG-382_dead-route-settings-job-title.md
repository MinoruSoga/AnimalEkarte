# BUG-382: 複数マスタ設定ページが 404 / ドキュメントと実装の命名不整合

**作成日**: 2026-04-15
**Status**: CLOSED
**Priority**: **HIGH** (カルテ編集画面から 2 件の dead route 編集リンクを露出 → 本番運用中の医師作業が 404 で断絶するリスク)
**Affects**: `frontend/src/app/router.tsx`, `features/staff-management`, `features/reservations`, `docs/FUNCTIONAL_TEST_REPORT.md`

---

## 概要

複数のマスタ設定ルートが 404。旧命名の記述がテストレポートに残存し、現行 UI と一致しない。

### 404 になるルート (ブラウザ検証 2026-04-15)

| 旧ルート (404) | 既存レポート参照 / 呼び出し元 | 現行相当ルート |
|------|------|------|
| `/settings/job-title` | 14.7 役職マスタ管理 | `/settings/occupations` (職種マスタ、4 件) |
| `/settings/service-type` | 14.5 サービス種別 / 14.20 サービス種別マスタ色設定 | `/settings/reservation-type` (予約区分マスタ、25 件) |
| `/settings/diagnosis-type` | **カルテ編集画面「診断カテゴリ」欄の「編集」リンク** | `/settings/diagnosis` (カテゴリタブ、8 件) |
| `/settings/diagnosis-name` | **カルテ編集画面「診断病名」欄の「編集」リンク** | `/settings/diagnosis` (病名タブ、20 件) |

**重要**: 2 つの追加 dead route はカルテ編集画面で実運用中の「編集」リンク先。医師が診断マスタを追加編集しようとクリック → 404 → 作業断絶。本番影響大。

### 現行 UI の状態

- `/settings` トップに「役職マスタ」「サービス種別」カード **無し**
- 代わりに「職種マスタ」「予約区分」カードが存在
- `/settings/occupations` は正常表示（タイトル「職種マスタ」、4 件）
- `/settings/reservation-type` は正常表示（タイトル「予約区分マスタ」、25 件、7 グループ構造）

## 再現手順 (ブラウザ検証 2026-04-15)

1. `admin@noavet.jp` でログイン
2. アドレスバーに `http://localhost:3003/settings/job-title` を直接入力
3. 結果: **「ページが見つかりません」** (404)
4. `http://localhost:3003/settings/occupations` を開く
5. 結果: 「職種マスタ」 4 件正常表示
6. `/settings` トップページにもマスタカード「役職マスタ」は存在しない（「職種マスタ」カードのみ）

## 原因仮説

過去リリースで役職マスタ（`job_title`）を職種マスタ（`occupation`）にリネーム・統合したが、以下のいずれかが残存:
- 旧ルート `/settings/job-title` のリダイレクト未設定
- バックエンドに `job_titles` テーブルが残っていて、BUG-112 修正時点の FK 依存チェックが今も `job_titles` を見ているか確認が必要
- テストレポート 14.7 がリネーム前のドキュメントのまま更新漏れ

## 確認事項

1. **DB スキーマ**: `backend/migrations/` に `job_titles` と `occupations` のどちらのテーブルが存在するか
2. **API エンドポイント**: `/api/v1/masters/job-titles` と `/api/v1/masters/occupations` のどちらが生きているか
3. **バックエンドログ**: BUG-112 修正時点 (役職削除時に 409 を返す実装) のコードが現在も `occupations` に対して同じ動作をするか
4. **フロント router**: `/settings/job-title` → `/settings/occupations` へのリダイレクトが必要か（外部ブックマーク互換）

## 修正方針

### Step 1: 命名統一の決定
- リネーム済みなら「職種」に統一（社内用語優先）
- 旧称「役職」を維持するなら `occupations` を `job_titles` に戻す

### Step 2: dead route の処理
- ルーター側で `/settings/job-title` → `/settings/occupations` へ 301 redirect
- もしくは router 定義からの完全削除

### Step 3: ドキュメント更新
- `docs/FUNCTIONAL_TEST_REPORT.md` 14.7 の URL を `/settings/occupations` に修正
- 件数も現行（4 件）に揃える
- 削除されたマスタ（例: 管理者）の影響を確認

## 関連

- **BUG-112**: 役職削除時の FK 依存チェック (`job_titles` に対する 409) が `occupations` で同等動作しているか検証必要
- `docs/FUNCTIONAL_TEST_REPORT.md` 3389-3394 行の 14.7「役職マスタ管理テスト」
- `docs/FUNCTIONAL_TEST_REPORT.md` 3328 行 14.1「役職マスタ」カード（記述ありだが UI 上は存在しない）

## 影響範囲

- **ユーザー影響**: 直接 URL でブックマークしていた管理者は業務停止（404 のため代替経路を自力で探す必要がある）
- **テスト**: FUNCTIONAL_TEST_REPORT の 14.7 全項目が実質 **検証不能**（ルート自体が存在しない）
- **監査・権限**: RBAC の `master-job-title` リソースが存在する場合、フロント側で参照されなくなっている可能性

## 後続アクション

- 「職種マスタ」「役職マスタ」のどちらが正か決定 → 要プロダクトオーナー判断
- 決定後、router / migrations / i18n / docs を一括 rename
