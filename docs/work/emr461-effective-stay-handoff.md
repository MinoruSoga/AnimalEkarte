# EMR-461 / EMR-499: 734の先行実装と後続の保存接続

2026-10-20の親ChatGPT Dotsの採用決定を記録する。旧head `06ae59aca8f3912d567ff87397f00d581f68c158` の検討案は[当時のhandoff](https://github.com/MinoruSoga/AnimalEkarte/blob/06ae59aca8f3912d567ff87397f00d581f68c158/docs/work/emr461-effective-stay-handoff.md)に保全し、現契約は本書で更新する。本書は親決定の記録であり、列型・migration番号など未提示の詳細を独自に確定しない。

**2段階の完成条件**: PR734はsource先行。後続499 PRが保存意味論と461の実write接続まで所有する。734だけで機能完了・利用開始可とは扱わず、旧保存caller・swapを第三の未定チケットへ移さない。今回の更新は文書のみで、後続実装・migration編集は別途の着手指示と共有claim調整を待つ。

## 1. 実装済み（sourceに存在、実行受入とは別）

基準source headは `0e79ae86bcbcda1de7dcc87d97bd33ad014a0fce`。公開wireの正本はmain採用済みPR740（照合main `4debcf9961fdd8735afe57b2886ae354ac0b1a5a`）。

| 範囲 | 現source | 成立範囲・制限 |
|---|---|---|
| 正常movementの単一core | 実admit→admitted、矛盾しないadmit→dischargeと開始前cancel/no_show→released。全履歴をrevision順で検査し、不正・未解決履歴をstay単位のunknownにする | 予定日/status/最新eventだけから予約・退院を生成しない。訂正resolver、予約proof、通常writerの拒否接続は未実装 |
| core reader / 461 adapter | 医院付きStayCaseID、同TXのstay行ID昇順FOR SHAREとmovement read。候補/保存側の集計は同じcore面を使い、hospitalization IDだけを自己除外する | 原event不変・INSERT時stay lockは既存001と静的に照合しただけ。全writer参加、一貫snapshot、複合CASの成立証拠ではない |
| 740 occupancy read | clinic/編集対象を文字列IDでecho、knownは両配列非NULL、unknownは200かつoccupied/candidate両配列NULL、実取得障害500 | 部分候補や空配列でunknownを偽装しない。患者/滞在明細を返さない |
| 現在割当と候補 | current_assignmentを別fieldで保持。同院の有効な独立/代表のみ候補、構成員に他滞在の占有があれば代表も除外 | core不明・候補外・無効・旧構成員という理由だけで現値を消さない。現値保持は新規候補への追加許可ではない |
| lossless ID / 認可 | 正規なPostgreSQL bigint正数文字列。未許可/重複queryと範囲外は400。選択医院createまたはedit、自己除外には同医院editも必須。scope外/不存在/削除済みは同じ404 | view/master編集権限を暗黙の前提にしない。実認可DBの受入は未実施 |
| APIとfixture | api.yamlのcage-occupancy/cage-swapおよび関連3 schemaは上記main契約本文に一致。HTTP境界/純粋projection/同contextの限定fixtureを追加 | swap DTO/write実装の追随完了ではない。fixture追加・静的確認をunit/実DB PASSとしない |

source配置: [正常core](../../backend/internal/medicalrecord/stay_cage_effective_state.go)、[reader](../../backend/internal/medicalrecord/stay_cage_effective_state_reader.go)、[adapter](../../backend/internal/medicalrecord/cage_occupancy_projection.go)、[HTTP](../../backend/internal/medicalrecord/hospitalization_cage_occupancy_http.go)、[候補projection](../../backend/internal/medicalrecord/hospitalization_cage_occupancy_read.go)。

## 2. 未実装（後続499が実装まで所有）

### 親Dotsが採用した保存・解決契約

| 契約 | 採用内容 | 現sourceとの差 |
|---|---|---|
| Q1 予約proof | 既存immutable `appointment_change_events.after_values.stay_acceptance_changes` の型付きv1配列を明示受入の証跡とする。同院・同appointment・同stayの受入系列、現在appointment.versionに一致するevent、初回/現在eventの成功operation receipt、その参照・保存版を検証する | decoder、証跡writer、現効力resolver、readerへの接続なし。予定日・status・行の存在だけでreservedを作らず、旧行を推測backfillしない。別stayへの付替えを新しい通常操作として追加しない。専用writerの意味が未成立なら拒否して旧割当を保持する |
| Q2 訂正保存 | 新correction effect子表にreplace/voidと完全置換payload。親のnullable `correction_effect_schema_version` は新correction=1、旧=NULL。親子同TX完全性はdeferred guard、子にもimmutable/RLS/同医院・同stayのscope保護 | 追加migration/model/ERD、writer、deferred guardとも未実装。001は変更せず、446/466とclaim調整後に追加差分を作る。旧correctionの意味を推測して埋めない |
| Q2 effect解決 | 現在leafの完全effectで解決し、中間の旧effect不明だけで修復を妨げない。構造不正・分岐・現在effect不明はunknown。既知のcancel/no_show.actual_atは保持しNULL強制しない。inpatientの医師/入院記録必須をhotelへ拡張しない | 現coreはcorrectionを一律unknownにする。さらにadmitのAdmissionRecordRefをhotelにも要求しており、この点は採用契約より厳しい。resolver/対象fixtureで解消が必要 |
| 版とCAS | `stay.version` + movement head revision（0も有効）+ 必要なappointment.version + leaf期待値の複合CAS。保存時に現在根拠を同TXで再評価する | 現EvidenceVersionは検査したmovement revisionだけで、adapterは正数を要求する。予約証跡を含む複合版ではなく、CAS/fence/stay.versionの代用にしない。0を避ける偽正数、versionだけの偽更新を作らない |
| Q3 通常遷移 | 通常writerは未開始→admit→discharge、未開始→cancel/no_showを検証し、discharge単独・admit後cancel/no_show・退院後再admit・実時刻逆転を拒否する。訂正解決後も通常遷移を検証し、不正ならunknown | 正常履歴readerのunknown判定は存在するが、旧保存callerから単一writerへの接続はない。readerの拒否をwriterの拒否実装と数えない |
| Q4 gate / lock | 既選定 `reservation-graph:write:v1` を最外業務TXの先頭で取得→receipt/binding→各ownerの既存lock順。新gate namespaceを作らず、ambient participantは別TX/入れ子Runnerを開始しない | 関連blocking writerのgate実参加は未実装。経路別の実参加、または当該経路停止を後続499で明記し、単一reader注入で全体安全と主張しない |
| 候補readの一貫性 | 候補readは一貫snapshot。保存は候補readの結果を確保とみなさず、同TXの現在根拠を再評価する | ambient TXへの参加と短いread TXはあるが、現WithTxは明示のsnapshot設定なし。複数queryを同TXへ置いただけではsnapshot成立を認定しない |
| 461の保存境界 | 変更save入口で旧日付/statusから有効状態を捏造しない。新規非NULL割当のunknownは409。未指定/維持/明示clearの既存意味を壊さない | 旧Create/Update/退院/swapに旧status/type/start/end依存が残る。現typed unknownは保存側では503へ伝播し、採用409のwrite接続は未実装。旧statusにより占有再検査をskipする分岐も置換対象 |

### 具体的な保存callerとコンパイル残差

| 接続箇所 | 未完のコードと後続499の責務 |
|---|---|
| [hospitalization_service.go](../../backend/internal/medicalrecord/hospitalization_service.go) / [create](../../backend/internal/medicalrecord/hospitalization_service_create.go) | Createの予定日由来default status、結果statusを使う占有判定、Updateの旧status遷移・旧列update mapを499の単一core/writerへ接続。未指定/維持/clear/新規非NULLを区別し、unknown新規割当を409で拒否する |
| [hospitalization_repository.go](../../backend/internal/medicalrecord/hospitalization_repository.go) / [退院](../../backend/internal/medicalrecord/hospitalization_discharge.go) | `status != discharged` の旧SQL、退院更新と会計・監査・成功receiptを採用された状態根拠/owner/TXへ接続。既存の部分write防止を保持する |
| [swap service](../../backend/internal/medicalrecord/hospitalization_service_cage_swap.go) / [request DTO](../../backend/internal/medicalrecord/hospitalization_request.go) / [handler](../../backend/internal/medicalrecord/hospitalization_handler.go) | 旧status依存、数値ID DTO、公開740の409 code、両対象の現在認可/期待割当・原子性、同TXの有効状態再検査を整合する。結果不明時にreceipt保証が既にあると推定しない |
| [model.Hospitalization](../../backend/internal/model/hospitalization.go)との型不一致 | modelに存在しないStatus/StartDate/EndDate/HospitalizationTypeを旧callerが参照・struct literalへ設定している。**現在コードはこの型不整合でコンパイル不能であり、「試験待ちのみ」ではない。** 旧列/fieldを復活して隠さず、499の実接続で解消する。現セッションでGo compilerを実行したという主張ではない |

### 後続499の着手プラン（実装開始指示前の記録）

1. 共有本体で499と関係claim/branch/worktree/dirtyを再確認し、446/466と追加migration/model/ERDの所有範囲・番号を調整する。現在の読取では `claim/EMR-461` と `claim/EMR-466` が存在し、499/446 claimは見当たらないが、未claimを着手許可とは扱わない。今回499のclaim取得・移管・migration番号予約は行わない。
2. 関連blocking writer（予約受入/取消/未来院/接続、開始/退院/訂正、ケージcreate/edit/swap等）を実際のowner呼出しまで列挙し、既存gate/receipt/binding/lock順とambient参加を対応付ける。未参加のまま安全と認定せず、実参加または経路停止を決めて実装する。第三の未定チケットへ残さない。
3. 446/466と調整した追加保存差分を前提に、型付きv1予約proof、schema-version付き訂正親子、完全leaf resolver、複合CASを単一499 core/writerに実装する。旧行推測backfill、通常付替え/再入院、ホテルへの不要な必須条件を追加しない。
4. 最外TX先頭の既選定gate→receipt/binding→owner既存lockを接続し、保存時の現在根拠再評価と候補readの一貫snapshotを成立させる。ambient参加者に新TX/入れ子Runnerを作らない。
5. 上記のCreate/Update/退院/swapを実writerまで接続し、削除済みfield/SQL依存、文字列ID、409コードを解消する。新規非NULLunknownの409と、未指定/維持/明示clearを別fixtureで確認する。
6. 既存の簡単なsource確認で進められる差分を進め、対象head・残差・確認内容を固定してレビューする。compile/unit/vet/gofmtを実行していなければ未実施と記録し、Docker検査を進行停止・マージ必須条件にしない。実DB/結合/実機は本人が指示した時にmainローカルで行い、検証専用環境を新設しない。

## 3. 後日試験（未実施・未検証）

| 検査 | 必要な証拠 / 現在の扱い |
|---|---|
| compile / unit / vet / gofmt | 現変更では未実施。Docker検査の実行・環境復旧を進行停止やマージ必須条件にしない。別途実施時は734の限定fixtureと後続499のproof/leaf/複合CAS/保存境界、対象headを記録する。2節の型不整合・保存未接続はコードの未実装として独立に扱う |
| 実DB保存制約 | 本人指示後、追加migration適用、親子deferred completeness、immutable/RLS/scope、成功receiptとの同TX・rollbackを確認。001変更・推測backfillなし |
| 競合 / snapshot | 本人指示後、候補read対movement、訂正対ケージ保存、逆順swap、予約接続対開始でlock待ち後の再検査、複合CAS、snapshot、rollback後の変更/成功receipt不残存を確認。context markerやfile-list fixtureを代用にしない |
| API / 結合 / UI / 実機 | 本人指示後、740 wire・scope/認可・lossless ID・unknownと取得障害・現在割当保持・未指定/維持/clear・新規409・交換原子性を確認。734だけで利用開始可とはしない |

現sourceで確認済みなのはdiff-check、inventory JSON parse、Ruby PsychによるOpenAPI YAML構文parse、cageの2 path/3 schemaと指定main契約の本文一致。`verify-agent-task.py --base 06ae59aca --plan` は対象列挙だけで実行件数0。今回の更新は文書のみで、runtime検査は実行していない。検査未実施をコード未実装の代わりにせず、Docker実行を進行停止・マージ必須条件にしない。

Docker拒否の記録（検査未実施の履歴であり、進行停止・マージ必須条件ではない）:

```sh
docker ps --format '{{.Names}}\t{{.Image}}'
```

`unix:///Users/minoru/.orbstack/run/docker.sock` への接続が `permission denied`。同操作の再試行、権限拒否の別経路迂回、ホストGo/npm代替、検証環境新設を行わない。DB適用/reset、実DB/結合/実機、STG/prod/storage操作も行っていない。

### 所有権・保全の証跡

- 旧ownerの `task-3/pr734` はhead `06ae59aca8f3912d567ff87397f00d581f68c158`・dirty空を確認し、親Dotsが旧ownerの734編集停止とtask-6への正式handoffを承認した。旧worktreeは保全。
- 共有本体の `claim/EMR-461` は旧tip `5c5219965ad44258b8ccd918ffe52179a6139ade` と未checkoutを確認後、通常の非強制削除→再取得を実施。新tipは `e60ba31e86f88e0e0a6a870989ac491fd8872654`。他claim/branchの解放・削除なし。このtipは取得基点であり、修正head/CAS/状態根拠版ではない。
- 新owner領域: `task-6/AnimalEkarte-pr734`、local branch `fix/pr734-emr461-499-20261020`。通常push先はPR734の `feature/emr461-cage-recheck-20261020`。source先行head `0e79ae86bcbcda1de7dcc87d97bd33ad014a0fce` はpush済み・Draft維持。
