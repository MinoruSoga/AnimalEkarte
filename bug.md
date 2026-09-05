# 製品バグ台帳

更新日: 2026-09-05

> **記録対象**: 確認済み製品 FAIL のみ（`docs/ops/testing/TEST_ARCHITECTURE.md` §6）。
> BLOCKED / PARTIAL は書かない。証跡に credential・token・cookie・idToken・個人情報（PHI）を含めない。
> 対応済みは本ファイルから削除する（履歴は Git と `reports/uat-YYYY-MM-DD/`）。
> 起票後は Linear で追跡する。見出し ID の重複禁止。

## BUG-20260905-004: LINE予約設定 PUT が closed_weekdays 省略で 500

- **対象**: V04 / S04 前提 / `PUT /api/v1/clinics/:clinic_id/line-reservation-settings`
- **再現**: 既存行がある clinic に対し、GET した body から `closed_weekdays` キーだけを削除して PUT する（他 jsonb は維持、`additional_fields` は明示可）。
- **期待**: 省略時はデフォルト（空配列 `[]`）を補完して 200 で保存される。または 400 で検証エラー。
- **実際**: `500 internal server error`。Backend log: `ERROR: null value in column "closed_weekdays" of relation "line_reservation_settings" violates not-null constraint (SQLSTATE 23502)`。UPDATE 値は `closed_weekdays=''`（jsonb として null）。`additional_fields` 省略は d3204b3da で修正済み（`[]` 補完）だが、同型の jsonb 列 `closed_weekdays`（および同時省略時の他 jsonb）は未対応。
- **判定**: FAIL（必須 jsonb の省略を永続化前にデフォルトせず 500）。
- **実施日**: 2026-09-05 postfix、`uat/20260905` @ 37044332d、local Docker UAT（MBPM3）。
- **証跡**: `reports/uat-2026-09-05-postfix/precheck-bug003.json` / `precheck-bug003.log`

## BUG-20260905-005: LINE予約設定 PUT が URL clinic_id を無視しセッション医院へ書き込む

- **対象**: `PUT /api/v1/clinics/:clinic_id/line-reservation-settings`
- **再現**: 執行スタッフ（main_clinic_id=1）でログインした状態で `PUT /api/v1/clinics/3/line-reservation-settings` に valid body を送る。
- **期待**: clinic 3 向けに保存されるか、所属外なら 403/404。レスポンス `clinic_id` はパスの医院と一致する。
- **実際**: HTTP 200 だがレスポンス `clinic_id=1`。Backend は `ExtractClinicID`（JWT コンテキストの主医院）を用い URL `:clinic_id` を Save に渡していない。ログ上も `PUT .../clinics/3/...` なのに `WHERE clinic_id = 1`。
- **判定**: FAIL（パス医院と永続化先の不一致。他医院設定の誤更新リスク）。
- **実施日**: 2026-09-05 postfix、`uat/20260905` @ 37044332d、local Docker UAT（MBPM3）。
- **証跡**: `reports/uat-2026-09-05-postfix/precheck-bug003.json` / backend log path clinics/3 → clinic_id=1

## 索引

| ID | status | area | severity | scenario |
|:---|:---|:---|:---|:---|
| BUG-20260905-004 | open | line-reservation-settings | high | V04 / S04 |
| BUG-20260905-005 | open | line-reservation-settings | high | V04 / S04 |

新しい製品 FAIL は [`docs/ops/testing/TEST_ARCHITECTURE.md`](docs/ops/testing/TEST_ARCHITECTURE.md) の規則に従って追記します。env・seed・権限不足による BLOCKED は記録しません。対応済み項目は本ファイルから削除し、履歴は Git と `reports/uat-YYYY-MM-DD/` を参照します。
