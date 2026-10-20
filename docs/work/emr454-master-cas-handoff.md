# EMR-454/455 master CAS 統合引継ぎ

## 今回の範囲

`002_master_insurance_versions.sql` が consultations/procedures/vaccines に正の bigint version と既存 `app_private.entity_version_guard()` の更新triggerを追加する。001・共有testdb/fixtureは編集しない。modelは454単独所有、vaccine専有API/service/repositoryは455が使用する。既存checkup_types版を再実装しない。

更新は編集開始時version必須（欠落/null/非正/非整数は400）、clinic+id+versionをWHERE照合する。同医院に対象が残る版不一致は409、未存在/他医院は404。versionをSETせずtriggerが内容変更時にのみ増分する。新規作成は版不要、成功応答は同TXの保存済み行を再読取するためinactive初期化の追加UPDATE後の版も返す。

前値の対象/親lock、変更、再読取、保険設定auditを同TXに維持する。取得し直した現行版・updated_at・接種/治療source版で期待版を補わない。旧is_non_insurance NULLのbackfillはしない。DEFAULT1はmaster CASの初期値で、過去のsource採用版の復元ではない。

## 466共有基盤担当への最小差分

- model.Consultation/Procedure/VaccineにVersion int64（column version bigint NOT NULL DEFAULT1）を追加。既存model登録を維持し、新規table登録は不要。
- AutoMigrateは002のCHECK/triggerを再現しない。CAS/trigger受入は001+002を正規順で適用する専用fixtureで行い、通常testdbのfake PASSを実storage保証へ昇格しない。
- 実DDL fixtureの追加範囲：3表version>0、初期版、業務変更+1/no-op不変/版のみ書換拒否、clinic+期待版競合、2接続の同版同時更新、監査失敗rollback、inactive create再読取。
- この依頼は登録差分の引継ぎであり、共有internal/testdb・既存schema fixtureへの編集や既存DB適用を行っていない。002番号は最新main直下001のみを確認して454が所有。merge前に再度採番衝突を照合する。

## receipt consumerの設計候補（未実装・未標準化）

CAS成功はreceipt完成を意味しない。採用済みoperation.Runner/Participantを再利用し、独自receipt表・先行claimを書かない。

1. 認証contextからclinic/staff Scopeを作る。Contentは対象種別・対象ID（update）と省略/falseを区別した業務変更値をcanonical化し、Applyも同じContentから導出する。CAS期待版はPreconditionsに置く。
2. Participantが現在grant、昇順target/parent lock、CASを再検証し、同じTXでmaster変更・audit・成功Resultを書き込む。ResultRefs/CommittedVersionsは保存済みID/版を使う。
3. unknownでは新しいkeyで再作成せず、元のscope/keyで照会・再認可する。DB内master Applyに外部送信や補償を混ぜない。
4. 未選定事項：owner_kind/intent/Content/Resultの正確なwire名、成功前のbindingから対象を復元して認可する方法、create未採番時の対応、操作keyと編集変更時の扱い。共有446/489で一つのconsumer契約を選んでから実装する。Content/fingerprintを対象metadataの代用として推測復元しない。

## 適用・検証状態

DDLソースの追加だけであり、DBへのmigration適用・reset・STG/本番操作・mergeは行っていない。採用/適用後の利用前にはAGENTSの手順に従う手動migrationが必要。限定source/SQLDryRun/fake/httptestは実DB trigger・2接続・実HTTP/UI/UATのPASSではない。

## 段階採用のコード分離（2026-10-20 追補）

計画のみの状態から、基盤と専有consumerの活性化を実際に分離した。

1. **#735 基盤**：3model/002・共通audit/parent lock・診察/処置CAS。vaccine/checkupのproduction constructor、専有API request/response、FE版必須化・保険送信は先取りしない。旧GETに版がないconsumerでも既存項目を編集できる。共通nullable型/transform/insurance helperはここだけで所有する。
2. **#730 vaccine consumer**：専有監査/CAS writer・responseと同時にcomposition、専有API記述、共有saveのvaccine活性化、編集開始版必須化を採用する。default reference servicesを構築してからvaccineだけaudited serviceへ差し替え、未採用checkupの契約を変更しない。
3. **#732 checkup consumer**：既存の専有constructor配線と同時に専有API記述、共有saveのcheckup活性化、編集開始版必須化を採用する。

`git merge-tree` の一時treeで基盤→vaccine→checkupを合成し、content conflictがないことを確認した。PRブランチ同士のmerge・ref更新・公開はこの検証で行っていない。各worktree全体の単純上書きで共有ファイルを失う方式は使わない。

段階別FEは89/104/114 tests＋scoped tsc/eslint/format PASS。所有するactual-source限定Goは247/390/523 test/subtest nodes＋vet PASS。実際のcompositionにあるmaster constructor呼出しを取り出し、同じactual service/portに対する署名契約もコンパイルした。APIはYAML・内部参照・段階別版契約を限定検証した。これらはcmd/api全体のbuildや実HTTPのPASSではない。

前2段階で未採用の旧consumerまで含めた試験は、既存main同一のvaccine/checkup service/responseのSortOrder int対model *int不整合により4箇所/2箇所でcompile失敗した。失敗証跡を保持し、未採用旧consumerを除いた所有scopeのPASSと区別した。4ファイルがmainと同一であることを照合し、各専有consumerに存在する修正を基盤へ複製していない。全backend/API独立buildは、sharedkernel/testdb等も含む既存model caller移行が済むまで未成立である。

003画像台帳を追加した#731との統合では、ERD index/manifestを両migrationを含む採用treeから再生成する必要がある。002/003のDB適用は未実施。receipt保存/照会、source adoption/staff overrideもこの段階分離で完成扱いにしない。
