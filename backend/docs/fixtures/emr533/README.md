# EMR-533 PATCH description null 許容 fixtures

[OpenAPI](../../api.yaml) の `UpdateHospitalizationPlanRequest`（`PATCH /masters/hospitalization-plans/{id}` 専用）に付属するダミーJSON。正本は Plane EMR-533（STATIC-AC-FIX-20261022-S09-NULLABLE）と EMR-528 の既決 PATCH 意味（#806 で収録済みの body_size/billing_unit 三値）。実患者・実医院データではない。

全fixtureは **NOT_RUN**。作成と YAML/JSON の静的読取だけを行い、OpenAPI instance schema 検証・Go/DB/HTTP 試験・Docker は実行していない。`manifest.json` の `expected_schema`（valid/invalid）と `semantic_expectation`（保存意味）は期待値であり検証 PASS ではない。

- 通常の fixture file は HTTP body そのもので、metadata wrapper を付けない。
- `name` は既存 `HospitalizationPlan` の `required: [name]` を PATCH 専用 schema へそのまま写したため、description 軸の fixture にも同梱する（票の「他必須条件が必要なfixtureは既存契約に従う」）。Go の PATCH 実装は name 省略も受け付けるが、この差は本票の範囲外として変更していない。
- 保存意味は既存値を `"old"` とした description 軸の期待値で、endpoint 全体の status や no-op 受理を新たに決めない。
- 対照: POST（`createHospitalizationPlan`）と GET 応答は共用 `HospitalizationPlan` のままで、description を nullable に緩和していない。POST に null を送った場合の runtime 挙動は本票で変更・決定しない。

後日の本人指示時に main ローカルで schema 参照解決・正負 fixture・API 受入を確認する。コード生成（`make codegen`）はユーザーのみが実行し、api.yaml はその入力ではない（tygo は Go 型が入力）。
