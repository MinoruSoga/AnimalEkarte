# EMR-536 薬剤 PATCH clear_strength と税率 exact 往復 fixtures

[OpenAPI](../../api.yaml) の `UpdateMedicineRequest`（`PATCH /masters/medicines/{id}` 専用）と共用 `Medicine` 応答に付属するダミーJSON。正本は Plane EMR-536（POSTMERGE-AC-FIX-20261022-C-MEDICINE-EXACT-CLEAR）、EMR-447 Q4 採択（clear_strength）、SPEC-QA-20261022-447-EXACT、PR #812 discussion r4345023830。実患者・実医院・医学的用量の証拠ではない。

全fixtureは **NOT_RUN**。作成と YAML/JSON の静的読取だけを行い、OpenAPI instance schema 検証・Go/DB/HTTP 試験・frontend unit・Docker・実ブラウザは実行していない。`manifest.json` の `expected_schema` と `semantic_expectation` は期待値であり検証 PASS ではない。DB 往復未実行をモックで合格扱いしない。

- `name` は既存 `Medicine` の `required: [name]` を PATCH 専用 schema へそのまま写したため同梱する。
- 税率: 保存済み `"0.1234567890123456789"` / `"0.0000001"` は GET で十進文字列のまま返る。FE は未編集の税率を PATCH に含めず（省略＝保持）、明示選択時だけ十進文字列を送る。名称のみ・備考のみの更新で exact 値が不変であることを再GETで確認する。
- clear_strength: true＝SQL NULL へクリア、false/省略/null＝クリアしない（Go の `bool` binding で null は false）。strength との同時送信は schema で拒否しない（現実装は clear 優先、実効 per_weight は既存検証で 400）。一律の排他 400 は新設しない。FE は明示消去＋none のときだけ true を送り strength を同時送信しない。
- 対照: GET/POST の共用 `Medicine` に clear_strength を追加しない（GET 対照 `get-no-clear-strength`、POST 対照 `post-control-no-clear-strength`）。

後日の本人指示時に main ローカルで schema 参照解決・正負 fixture・API 受入・再GET・frontend unit を確認する。コード生成（`make codegen`）はユーザーのみが実行し、api.yaml はその入力ではない。
