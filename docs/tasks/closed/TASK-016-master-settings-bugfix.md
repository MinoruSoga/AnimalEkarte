# TASK-016: マスタ設定ページ バグ修正（トリミング・診断病名）

**作成日**: 2026-03-18
**ステータス**: Open
**依頼元**: ユーザー

---

## 概要

トリミングマスタページのランタイムエラーと、診断病名マスタページのデータ未表示バグを修正する。
両方ともバックエンドのレスポンス層のバグ。

## 依頼内容（原文）

> トリミングマスタページにてエラーです。
> 診断病名マスタページにて、DBデータが表示されていません。

## 仕様確認ログ

確認事項なし。バグの根本原因がコード調査で特定済み。

## サブタスク分解

| # | サブタスク | 領域 | イシュー | 依存 | 完了 |
|---|----------|------|---------|------|------|
| 1 | トリミングマスタ: ダングリングポインタ修正 | BE | BE-043 | - | [x] |
| 2 | 診断病名マスタ: レスポンス形式修正（paginated → plain array） | BE | BE-044 | - | [x] |

## 受入条件（Acceptance Criteria）

- [ ] AC-1: `/settings/trimming` ページがエラーなく表示され、コース/オプション一覧が正しくロードされる
- [ ] AC-2: コース作成・編集・削除が正常に動作し、`target_size` フィールドが正しく保存・表示される
- [ ] AC-3: `/settings/diagnosis` ページで診断病名カテゴリ一覧（8件のシードデータ）が表示される
- [ ] AC-4: `/settings/diagnosis` ページで診断病名一覧（20件のシードデータ）が表示される
- [ ] AC-5: `docker compose exec backend go build ./cmd/api` がエラーなくパスする

## 技術的判断

| 判断事項 | 採用案 | 理由 | 却下案 |
|---------|--------|------|--------|
| TargetSize の型 | 値型 `string` + `omitempty` | ポインタ不要、JSON 空文字は省略 | ポインタ型 `*string` を修正 |
| 診断病名 API レスポンス | プレーン配列 | フロントエンドが `[]` を期待、他のマスタ API と統一 | ページネーション付き |

## 影響範囲

### Backend
- `backend/internal/handler/trimming_master_response.go` — TargetSize フィールドの型とポインタ処理修正
- `backend/internal/handler/diagnosis_handler.go` — ListDiagnosisCategories, ListDiagnosisNames のレスポンス形式修正

### Frontend
- 変更なし（フロントエンドの実装は正しい）

## 参照実装

- `backend/internal/handler/animal_species_handler.go` — プレーン配列を返すマスタ API の参照例

## リスク・懸念事項

特になし。バックエンドのレスポンス層のみの修正で、DB やモデルに変更なし。

## 未解決事項

- なし

## 実装順序

1. BE-043: トリミングマスタ レスポンス修正
2. BE-044: 診断病名マスタ レスポンス修正
※ 依存関係なし。並行実装可能。

## 関連イシュー

- BE-043: [トリミングマスタ ダングリングポインタ修正](../../backend/issues/open/BE-043-trimming-master-dangling-pointer.md)
- BE-044: [診断病名マスタ レスポンス形式修正](../../backend/issues/open/BE-044-diagnosis-master-response-format.md)
