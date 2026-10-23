# EMR-445 A02 担当者別 trimming 開始 policy fixtures

[OpenAPI](../../api.yaml) の `/masters/reservation-types/{type_id}/staffs/{staff_id}/line-start-policy`（GET/PUT/DELETE・preview）と `/line-start-policy-operations/{operation_key}` に付属するダミーJSON。正本は [A02/A03 選定契約](../../../../docs/work/emr/emr-clinic-line-time-selected-contract-20261022.md) §4・§5.7。実医院の担当者設定ではない。

全fixtureは **NOT_RUN**。作成と YAML/JSON の静的読取だけで、OpenAPI instance schema 検証・Go/DB/HTTP 試験・Docker は実行していない。runtime（EMR-448/492）と DDL/model（EMR-446/466、#824 の line_staff_start_policies・492 anchor 後続差分）は別PRで、本 fixture は実装済みを示さない。`manifest.json` の `expected_schema` と `semantic_expectation` は期待値であり検証 PASS ではない。

- file は HTTP body そのもの（metadata wrapper なし）。ID/版は §1 の正規十進文字列、operation_key は小文字 UUID のダミー。
- times の一意昇順・date_overrides の日付一意昇順、tombstone の create 拒否、assignment 変更/失権、元 anchor による lookup、同時実行・監査 rollback、④ 予約適格性との同時評価は runtime 試験の対象（契約 §8）で、本 API fixture では扱わない（`put-duplicate-time` は uniqueItems で schema 上も invalid）。
