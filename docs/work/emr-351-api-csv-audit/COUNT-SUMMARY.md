# Four-clinic bounded survey

Base `ed04c0237f81ddc2890d39b528d3608a58180c8b`; observation `2026-10-12T18:18:44.981815+00:00`.

The network-none read-only Docker scan hashed the same bytes as the CSV reader and matched each fixed manifest SHA and row count. See [full per-rule counts](aggregate-counts.json) and [DTO source catalog](api-binding-catalog.json). Scan completion is not import/API acceptance.

| observation | hachioji | jouto | shikishima | hakobuneco |
|---|---:|---:|---:|---:|
| table receipts | 21 | 21 | 21 | 21 |
| column empty/nonempty counters | 236 | 236 | 236 | 236 |
| appointments equal start/end (API strict > parity; not DB failure) | 81712 | 126637 | 4981 | 26310 |
| owners.phone / owner_service | 423 | 327 | 57 | 253 |
| owners.postal_code / owner_service | 7 | 141 | 6 | 44 |
| owners.email / owner_service | 106 | 285 | 45 | 62 |
| billing_items.unit_price / createBillingItemRequest | 38988 | 28882 | 5612 | 3840 |
| estimate_items.unit_price / createEstimateItemRequest | 7 | 1 | 0 | 0 |

Counts are overlapping scoped rule observations. Required-name hits can include historical empty text intentionally preserved by ForceNotNull; a DTO rejection does not prove a DB failure. General reservation and trimming have different status DTO domains.

All selected initial-DDL enum/PG post-rounding numeric overflow rules (30), listed uniqueness predicates and relational pairing observations returned zero hits in this frozen cohort. This says nothing about omitted nullability, date, later ALTER or service/lifecycle/permission constraints.

DDL SHA `d792bca4f4d985f1b2acdc7cd9ac0c9af80fe28aa3135712ec70896086c8ba2d`. Investigator SHA `03012229931e6e4a067d4231cac07a938229681ca7e62184634a425e378c93bd`. These receipts and technical manifest/run/build IDs identify inputs, not patients. No source cell values or patient IDs are included.

Followups: EMR-374 existing-floor preflight implementation; EMR-375 field-edit contract; EMR-376 historical/API-domain/correction decisions. No production validators implemented here.
