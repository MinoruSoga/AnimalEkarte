# 製品バグ台帳

更新日: 2026-09-05

> **記録対象**: 確認済み製品 FAIL のみ（`docs/ops/testing/TEST_ARCHITECTURE.md` §6）。
> BLOCKED / PARTIAL は書かない。証跡に credential・token・cookie・idToken・個人情報（PHI）を含めない。
> 対応済みは本ファイルから削除する（履歴は Git と `reports/uat-YYYY-MM-DD/`）。
> 起票後は Linear で追跡する。見出し ID の重複禁止。

## 索引

| ID | status | area | severity | scenario | 層 |
|:---|:---|:---|:---|:---|:---|
| BUG-20260905-001 | open | master-chief-complaint | high | V04 | BE |

未対応の製品 FAIL あり（下記）。

新しい製品 FAIL は [`docs/ops/testing/TEST_ARCHITECTURE.md`](docs/ops/testing/TEST_ARCHITECTURE.md) の規則に従って追記します。env・seed・権限不足による BLOCKED は記録しません。対応済み項目は本ファイルから削除し、履歴は Git と `reports/uat-YYYY-MM-DD/` を参照します。

### BUG-20260905-001

| field | value |
|:--|:--|
| status | open |
| area | master-chief-complaint / inquiries |
| severity | high |
| scenario | V04 master-chief-complaint DELETE |
| 層 | BE |

**現象**: `DELETE /api/v1/masters/chief-complaint-types/:id` が常に **500** を返す。Create/Update/deactivate（`is_active=false`）は成功する。

**証跡** (UAT 2026-09-05 r6, staff 執行, no PHI):
- POST create → 201
- DELETE → 500 `{"error":"internal server error"}`
- BE log: `column inquiries.deleted_at does not exist (SQLSTATE 42703)` at `chief_complaint_repository.go` `CountUsageByChiefComplaintTypeID` (`inquiries.deleted_at IS NULL`)

**期待**: 未使用なら 204、使用中なら 409。500 にしない。

**メモ**: `inquiries` テーブルに `deleted_at` が無いのに usage count SQL が参照している。cleanup は soft-deactivate で代替可。

