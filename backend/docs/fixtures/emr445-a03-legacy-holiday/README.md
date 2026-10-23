# EMR-445 A03 旧 shifts 休日 mutation 停止 fixtures

[OpenAPI](../../api.yaml) の旧 `POST /clinic-holidays`・`DELETE /clinic-holidays/{date}` に追記した 409 `contract_upgrade_required`（`LineTimeContractUpgradeRequired`）のダミーJSON。正本は [A02/A03 選定契約](../../../../docs/work/emr/emr-clinic-line-time-selected-contract-20261022.md) §2.4。

全fixtureは **NOT_RUN**。作成と JSON の静的読取だけで、OpenAPI instance schema 検証・Go/DB/HTTP 試験・Docker は実行していない。活性化判定・zero-write・旧 closing レジ休日（`/closing-settings/holidays`、停止対象外）との独立編集・予約/common reader の旧表非参照は runtime 試験の対象（契約 §8）で、本 fixture では扱わない。
