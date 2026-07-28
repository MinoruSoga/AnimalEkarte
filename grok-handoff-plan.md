# GrokAgent 作業単位ハンドオフ

作成日: 2026-07-28
現行照合基準: HEAD `09ea4716455fc013b8ad99c5d7e8f1383a92cf40`

## 正本と対象

本書は `3-session-agent.html#ledger` の派生ビューである。状態、対象パス、行番号、完了・削除判断が食い違う場合は、現行コード・git履歴を実測したうえで台帳を正とする。本書へ実装契約を複製せず、各作業単位は台帳の同名sectionを実行時の正本として読む。

現行台帳は23 section。`TASK-251` は別のIssue/unit所有で本書から直接割り当てず、依存関係としてのみ参照する。本書が扱う22 IDは次のとおり（うち `TASK-461` はBLOCKEDのため現在は割当不可、割当可能21 ID）。

`TASK-ADR003` / `SEC-DUR-01` / `BUG-440` / `BUG-433` / `BUG-437` / `TASK-445` / `BUG-448` / `E2E-BASELINE-DEBT-01` / `BUG-454` / `BUG-455` / `BUG-456` / `BUG-458` / `BUG-459` / `BUG-460` / `TASK-461` / `BUG-463` / `BUG-464` / `BUG-465` / `BUG-466` / `TASK-467` / `TASK-468` / `TASK-469`

次は割当対象外。

- `BUG-449`: API/UIと死んだ基準値経路の撤去までcode-complete。残る臨床値投入はUSER/master-data境界のため台帳sectionを削除済み。
- `TASK-444-S1`: `226a7dc29` で完了。恒久段S2は未起票であり、本書から実行単位を作らない。
- `SEC-SWEEP-02`: residual 0で完了し台帳sectionを削除済み。旧S1/B1詳細を実行単位として再利用しない。
- `BUG-453` / `BUG-457`: CLOSED。再割当しない。

`BUG-454-B1` は `47b7b5b5a` で完了済み。`BUG-454` の残りはDB hardeningだけで、`TASK-445` と同じmigration所有へ統合する。

## セッション分割（4グループ）

USER指示（2026-07-28）: 「4セッションぐらいでいい」「worktreeを使ってもいいが、1つ1つの対応が終わり次第mainに統合するように」。

worktreeを使う場合も、グループ全体を溜めず、1 unitのAcceptance Checklistが全PASSになった時点で統合を終えてから次へ進む。共有working treeでは同時writerを置かない。他sessionの差分をrevert・stash・上書きしない。

| # | セッション | 対象台帳ID / subunit | 所有面・順序 |
|---|---|---|---|
| 1 | billing + inventory | `TASK-ADR003` / `BUG-463` / `BUG-440` / `TASK-445`（`BUG-454` DDLを含む）/ `BUG-455-S3` / `SEC-DUR-01-BILL-B1,BILL-B2` / `BUG-466` / `BUG-465` / `BUG-455-S4` | `backend/internal/billing` と `backend/internal/inventory`。`TASK-ADR003` はmethod↔system_key値一致、`TASK-445` はpaymentsを含むclinic-axis hardeningを所有する。`TASK-251` U2b-full第2段は現HEAD未着地なので、各billing unit着手直前に `git log` とownership diffを再実測する。競合時はinventoryを先行する。 |
| 2 | medicalrecord | `BUG-448` / `SEC-DUR-01-MR-T1,MR-S1` / `BUG-455-S5` | `backend/internal/medicalrecord`。`BUG-448` は保存済みpromptの実行状況を再確認し、未着地なら台帳の現行evidenceから開始する。 |
| 3 | pet / reservation / staff / trimming / lstep / lintscan | `BUG-455-S2,S6,S7,S8` / `BUG-464` / `BUG-437` | 各domainの所有パスは交差しないが、`DBOrTx` inventory gateを変更するsliceは `backend/internal/lintscan/dbortx_inventory_lint_test.go` のwriterを直列化する。 |
| 4 | frontend + docs/decision | `E2E-BASELINE-DEBT-01` / `BUG-456` / `BUG-458` / `BUG-433` / `TASK-467` / `TASK-468` / `TASK-469` | `frontend/`、`backend/CODING_RULES.md`、`BE-refactor.md`。同じdocs fileを複数writerへ渡さない。`BUG-433` の短期境界は完了済みで、恒久段を未起票のまま実装しない。 |

別枠のruntime実証:

- `BUG-459`: code landed at `114bdf23a`。POST/PATCH runtime proofはUNREPORTED。direct APIは `/api/v1/hospitalizations/{id}/care-plan-items`。
- `BUG-460`: codeは仕様どおり。reserved詳細のbrowser proofはUNREPORTED。
- `TASK-461`: 2026-07-28 attemptはpersona/exact-ID stop gate不一致のためBLOCKED・invalidated。現行fixtureと契約を整合した新しい実行契約ができるまで再実行しない。

runtime writerは専用Chrome portとfixture ownershipを確認し、API create/PATCH/DELETEと記録更新を直列化する。

## 優先再実測

`BUG-463`〜`BUG-466` は着手前再実測が必須。2026-07-28の本台帳監査では4件とも現行実装と一致し、OPENのままだった。

- `BUG-463`: `routes.go:111-112`、`billing_item_service.go:282,404-454`、`billing_item_repository.go:145-153,568-576`、`cash_register_service.go:407-414`
- `BUG-464`: `lstep_tag_summary_service.go:131-133`、`lstep_tag_summary_handler.go:43-59`、`httpapi/response.go:15-20`
- `BUG-465`: `inventory/repository.go:117-121`、`merchandise_item_repository.go:55-59`
- `BUG-466`: `inventory_request.go:38,75`、`inventory/repository.go:133-144`、`migrations/001_init.sql:668`

上記はスナップショットであり、実装sessionは台帳記載の `rg` / `Read` を再実行してから変更する。

## 共通実行契約

1. 一度に台帳の1 sectionまたは明示subunitだけを所有する。
2. 着手前にHEAD、対象パスの既存diff、先行commit、保存済みpromptの実行状況を確認する。
3. production writerは台帳に列挙されたpathと必要なfocused testだけを所有する。shared fileは単一writerへ直列化する。
4. Dockerのscoped verificationだけを自動実行する。full-project test/lint/build、migration適用、DB reset、codegen、push/mergeは行わない。
5. absent runtime/browser proofをPASSへ昇格しない。`UNREPORTED` / `BLOCKED` を維持する。
6. 1 unit完了ごとに、台帳の運用規約に従って完了sectionを削除する。完了履歴はgitと実装testを正本とする。

## 現行補足

- `BUG-455`: censusは39 field / 26 model file、実害あり28・実害なし11、PermissionGroup完了後の残りは27 field / 7 slice。AUTHのreachable履歴は `0c2bce44a` / `77fbbacaf`。残sliceは台帳のrepository allowlistとDB-backed false/default testを正本とする。
- `BUG-433`: 現HEADの例は生成Pet 33 property、予約wire DTO 10 field、差23。既存import allowlistは268 entry。旧「31対9」「268 file / 294 site」は使わない。
- `BUG-456`: 結果値・基準値は `f361801ee` で解消済み。unitの静的経路は存在し、根本原因とpost-fix runtimeはUNREPORTED。
- `SEC-DUR-01`: reachable履歴は `bf4e210e2` / `576ed17ce` / `4d83e8cad`。残りは `MR-T1` / `MR-S1` / `BILL-B1` / `BILL-B2` で、いずれも現HEAD未着地。
- `TASK-467` / `TASK-468` / `TASK-469`: docs/decision unit。production codeを同梱しない。詳細な不一致行と境界は台帳sectionを正本とする。
