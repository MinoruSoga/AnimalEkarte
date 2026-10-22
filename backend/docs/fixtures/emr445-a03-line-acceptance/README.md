# EMR-445 A03 LINE acceptance・非P 救済経路 fixtures

[OpenAPI](../../api.yaml) の `/clinics/{clinic_id}/line-reservation-settings/acceptance`（GET/POST/PATCH・confirm・preview・結果照会）、`/presentation`・`/registration-fields` の PATCH、旧全量 PUT の 409 `contract_upgrade_required` に付属するダミーJSON。正本は [A02/A03 選定契約](../../../../docs/work/emr/emr-clinic-line-time-selected-contract-20261022.md) §3・§5.6。実医院の LINE 設定ではない。

全fixtureは **NOT_RUN**。作成と YAML/JSON の静的読取だけで、OpenAPI instance schema 検証・Go/DB/HTTP 試験・Docker は実行していない。runtime（EMR-448/492）と DDL/model（EMR-446/466、#823 の acceptance_version / reviewed_acceptance_version）は別PRで、本 fixture は実装済みを示さない。`manifest.json` の `expected_schema` と `semantic_expectation` は期待値であり検証 PASS ではない。

- file は HTTP body そのもの（metadata wrapper なし）。ID/版は §1 の正規十進文字列、operation_key は小文字 UUID のダミー。preview の warning code 名と registration-fields の要素 shape はダミー。
- 版の1回増分・真の no-op・reviewed 更新、共通∩LINE の実効値と共通縮小時の LINE 保存値保持、旧 P 直接 UPDATE の閉包、活性化判定、現在認可（失権/異 actor/異 intent）、監査 rollback は runtime 試験の対象（契約 §8）で、本 API fixture では扱わない。
