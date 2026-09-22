# S31: 詳細画面への直接到達（会計・入院・在庫）

> **目的**: 一覧経由ではなく URL で直接 `/accounting/:id`、`/hospitalization/:id`、`/hospitalization/:id/edit`、`/inventory/:id` を開いたとき、権限に応じて正しく詳細が表示され、存在しない ID・権限不足で適切に拒否/ハンドリングされることを納品前に証明する。
> **所要目安**: 15分 / **深度**: 薄い
> **仕様正本**: [screens/11-accounting-detail.md](../../../spec/screens/11-accounting-detail.md)・[screens/08-hospitalization-detail.md](../../../spec/screens/08-hospitalization-detail.md)・[screens/27-inventory-form.md](../../../spec/screens/27-inventory-form.md)。検証キュー: `NOTE2-SWEEP-COVERAGE`（[todo-verification.md](../../../../todo-verification.md)）。

## 前提条件

- ローカルの使い捨て clinic、または承認済みの専用 UAT tenant。
- 各詳細に対応する有効なレコード（会計・入院・在庫）の前提 ID を fixture で確保する。入院は有効な cage 等の必須値を満たすものを作る。固定 ID・存在しない前提 ID は使わない。
- 権限別 actor（参照可・編集可・権限なし）を用意する。
- 依存シナリオ: なし。入院サイクルは S05、会計詳細は S07/S08 を参照する。

## 手順と期待結果

| # | 操作 | 期待結果 |
|:--|:--|:--|
| 1 | 有効な会計 ID で `/accounting/:id` を直接開く | 一覧を経由せず詳細が表示される（`paths.accounting.detail.getHref`）。データが正しくロードされる |
| 2 | 有効な入院 ID で `/hospitalization/:id` を直接開く | 入院詳細が表示される。`paths.hospitalization.detail` 経路が有効 |
| 3 | `/hospitalization/:id/edit` を編集権限ありで開く | 編集画面が開く。`paths.hospitalization.edit` 経路が有効 |
| 4 | 有効な在庫 ID で `/inventory/:id` を直接開く | 在庫詳細が表示される |
| 5 | 権限のない actor で各詳細 URL を直接開く | 閲覧不可または権限エラーで、データが見えない。URL 直打ちで権限を迂回できない |
| 6 | 存在しない/削除済み ID で各詳細を開く | 404 または「見つかりません」等のハンドリングで、真っ白・クラッシュ・別データの誤表示にならない |

## 確認観点

- ルートは `frontend/src/config/paths.ts` の `paths.accounting.detail`（`/accounting/:id`）、`paths.hospitalization.detail`（`/hospitalization/:id`）・`paths.hospitalization.edit`（`/hospitalization/:id/edit`）、`paths.inventory.detail`（`/inventory/:id`）。詳細ルートが登録されていることが前提。
- `NOTE2-SWEEP-COVERAGE` は前提 ID を確保して各詳細へ直接到達するカバレッジ確認。古い「カルテバグでブロック」の判定を流用せず、現行 schema と route で再検証する。
- 権限はサーバー側で検証されること。FE でボタンを隠すだけで URL 直打ちを許す場合は FAIL（L2 認可の欠陥）。
- 存在しない ID は 404/ハンドリングを返すこと。クラッシュ・無限ローディング・別レコードの表示は FAIL。
- 詳細の中身（項目・計算・ステータス）は各専用シナリオ（S05/S07/S08/V02）の対象。本シナリオは到達性と権限の境界のみ。

## 実装突合

- 変更サマリ:
  - `paths.accounting.detail`・`paths.hospitalization.detail`/`edit`・`paths.inventory.detail` のルート定義を現行コードと突合
  - `NOTE2-SWEEP-COVERAGE` の「`/accounting/:id`、`/hospitalization/:id`、`/hospitalization/:id/edit`、`/inventory/:id` の前提 ID 確保と直接到達」を手順 1–6 に対応づけ
  - 権限別・存在しない ID のハンドリングを手順 5–6 で境界確認として明示
