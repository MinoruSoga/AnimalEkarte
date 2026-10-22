# EMR-445 A01 品目別の全体割引除外 fixtures

[OpenAPI](../../api.yaml) の7品目マスタ（診察・検査・処置・予防接種・健診・薬剤・物販）に追加した `is_bill_discount_excluded` と、検査・薬剤・物販の `version` に付属するダミーJSON。正本は [A01 選定契約](../../../../docs/work/emr/emr-master-update-selected-contract-20261022.md#a-a01-全体割引除外の設定writer) A.1〜A.4 と、受領依頼 Plane EMR-445 comment `bd7995af`。実医院・実患者データではない。

全fixtureは **NOT_RUN**。作成と YAML/JSON の静的読取だけを行い、OpenAPI instance schema 検証・Go/DB/HTTP 試験・Docker は実行していない。server 実装（EMR-447/449、C）と DDL/model（EMR-446/466、B）は別PRで、本 fixture は runtime 実装済みを示さない。`manifest.json` の `expected_schema` と `semantic_expectation` は期待値であり検証 PASS ではない。

- file 名は `<resource>.<scenario>.<request|response>.json`。HTTP body そのもので metadata wrapper を付けない。
- request の必須項目は各 request schema の既存契約に従って同梱する（既存4種の PATCH は `version` 必須、分離した3種の request は既存の `required: [name]` を写した）。
- 「新設定を含む PATCH の version 必須」は `UpdateMedicineRequest` / `UpdateExaminationTypeRequest` / `UpdateMerchandiseItemRequest` の `anyOf`（`is_bill_discount_excluded` なし、または `version` あり）で表し、`*.patch-without-version` は schema 不適合（400）とする。
- 各品目で boolean の両値を網羅する: POST は省略・true・false・null・文字列（`"true"`、400）、PATCH は省略・true・false・null・文字列、GET は true・false・null・欠落。
- 応答の `is_bill_discount_excluded` は key 必須・nullable。DB NULL（既知未設定）は raw null で返し false へ書き換えない。key 欠落は unknown として不適合。
- 契約 A.6 の fixture 集合 F のうち、競合・無権限/別院・create 補償後の実版・同TX監査 rollback・旧 writer/import 保持・過去 snapshot 不変は runtime 試験の対象で、本 API fixture では扱わない（C/B の実装PRで受領）。

コード生成（`make codegen`）はユーザーのみが実行し、api.yaml はその入力ではない。
