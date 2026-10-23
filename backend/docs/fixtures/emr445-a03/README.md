# EMR-445 A03 共通医院時間・個別休診 adapter fixtures

[OpenAPI](../../api.yaml) の `/clinics/{clinic_id}/operating-hours`（GET/POST/PATCH・preview・結果照会）と `/clinic-holidays/common`・`/closing-settings/holidays/common`（月内取得・set・unset・preview）に付属するダミーJSON。正本は [A02/A03 選定契約](../../../../docs/work/emr/emr-clinic-line-time-selected-contract-20261022.md) §1・§2・§5.3–5.5。実医院の営業時間・休診日ではない。

全fixtureは **NOT_RUN**。作成と YAML/JSON の静的読取だけで、OpenAPI instance schema 検証・Go/DB/HTTP 試験・Docker は実行していない。runtime（EMR-448/492）と DDL/model（EMR-446/466）は別PRで、本 fixture は実装済みを示さない。`manifest.json` の `expected_schema` と `semantic_expectation` は期待値であり検証 PASS ではない。

- file は HTTP body そのもの（metadata wrapper なし）。ID/版は §1 の正規十進文字列、operation_key は小文字 UUID のダミー。
- CAS・同時初期化・警告 receipt の根拠再評価・現在認可（失権/異 actor/異 intent）・監査 rollback・旧 shifts mutation の `contract_upgrade_required` 停止・旧 closing レジ休日との独立は runtime 試験の対象（契約 §8）で、本 API fixture では扱わない。
