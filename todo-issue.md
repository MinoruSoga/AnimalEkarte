# 未完了 Issue 台帳（repo 正本）

最終照合: 2026-09-15。新規の実装・調査・PO 課題を管理する。完了した実装・仕様どおりの報告・回答済みの操作案内は削除し、履歴は Git と元の UAT 記録を参照する。検証は [todo-verification.md](todo-verification.md)、外部実行は [todo-operations.md](todo-operations.md) を正本とする。

既存 Linear は [BRT-4](https://linear.app/baritechllc/issue/BRT-4) 配下を更新する。新規 Issue を作らない方針を維持し、外部投稿は別途承認後。以下の新規 ID に専用 Linear Issue は今回割り当てていない。

<a id="open"></a>

## Open

着手プラン確認: 2026-09-15。11 ID すべてに個別の計画と根拠がある。各 ID を開き、前提が揃った範囲から着手する。

| ID | 状態 | 残作業・次の一手 |
|---|---|---|
| [UAT-R2-TREATMENT-COMMIT](#uat-r2-treatment-commit) | READY（再現・設計） | 数量確定の Enter / Blur / PATCH を再現し、2回 Enter と Blur の扱いを確定 |
| [UAT-R2-MASTER-LIST-HEIGHT](#uat-r2-master-list-height) | READY（再現・設計） | 治療検索の可視行を増やす案を、viewport・ズーム・操作部の収まりで検証 |
| [UAT-R2-MASTER-PATH](#uat-r2-master-path) | BLOCKED（医院入力） | 登録したマスタ種別・画面・単価・操作順を特定し、保存漏れと正しい会計導線を分離 |
| [UAT-R2-EXCLUSIVE-LOCK](#uat-r2-exclusive-lock) | PO 判断待ち | 二重入力・上書きの防止目的と占有範囲を確認し、既存の競合制御で不足する場面を定義 |
| [UAT-R2-CHART-FIT](#uat-r2-chart-fit) | BLOCKED（端末条件） | 解像度・ズーム・対象タブを確認してから、情報を落とさず余白・行高を調整 |
| [UAT-Q3-GENDER-MAP](#uat-q3-gender-map) | 未完了（old_db 照合・運用） | producer の性別コード修正状況を再確認し、既存 STG 行の訂正を別工程として計画 |
| [UAT-Q2-VACCINE-SPECIES](#uat-q2-vaccine-species) | 調査待ち | 承認された STG 集計で、マスタ種欠損・参照取り違え・旧記録そのものを分類 |
| [UAT-Q4-UNPAID-TRIAGE](#uat-q4-unpaid-triage) | 調査待ち | 承認された STG 集計で、旧未精算と支払未紐付けを切り分け |
| [UAT-Q2-TREATMENTS-IMPORT](#uat-q2-treatments-import) | DEFERRED（PO 未決） | 処置・処方明細を移行対象に含めるか、対象範囲・責任者・受入条件を決定 |
| [PO-PET-DECEASED-DATA-BACKFILL](#po-pet-deceased-data-backfill) | PO 判断待ち | 死亡状態と死亡日の不整合について、日付根拠・対象・監査・復旧を決定 |
| [TASK-444-ADDENDUM-CODEGEN](#task-444-addendum-codegen) | DEFERRED | 型生成経路の別スコープと user-run codegen が認められたときに再開 |

READY は上表の再現・設計範囲を指す。今回の文書整理は、実装、DB 操作、医院への送信を実施する承認ではない。医院側の要件責任者名は元記録に未記載であり、担当者が実装前に確認する。

## 第2報の着手プラン

### UAT-R2-TREATMENT-COMMIT

- **目的・根拠:** 数量確定時の待ちと誤確定を減らす。[医院第2報](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-r2-treatment-commit)。`TreatmentRowEditors.tsx` は Enter で commit、数量欄は Blur でも同じ保存経路を通る。
- **最初の作業:** 1回目・2回目の Enter、IME、Blur、Escape、保存中の再操作を再現し、PATCH 回数と保存完了時間を分けて確認する。Blur で保存するか、1回目の表示をどうするかを確定してから失敗テストを追加する。
- **完了条件:** 1回目の Enter ではサーバー保存せず、2回目の確定値が再読込後も保持される。Escape の取消、権限・確定済み診療の制約を維持。二重 PATCH や失敗時の入力消失がなく、変更対象を Docker の scoped vitest で検証する。測定前に「高速化した」と扱わない。

### UAT-R2-MASTER-LIST-HEIGHT

- **目的・根拠:** 項目選択時の不要なスクロールを減らす。[医院第2報](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-r2-master-list-height)。`TreatmentSearchDialog` の一覧上限は `max-h-[400px]`。
- **最初の作業:** 治療検索を対象として viewport に応じた高さを検討する。任意件数を常に全件表示する約束にせず、検索欄・閉じる操作・候補一覧を同時に使える範囲を決める。
- **完了条件:** 対象端末で可視行が増え、画面外にはみ出さず、検索・キーボード選択・フォーカス復帰が成立する。長い一覧には必要なスクロールを残す。対象 component の Docker scoped vitest と画面確認を行う。

### UAT-R2-MASTER-PATH

- **入力・先行準備:** 医院が使った登録画面、マスタ種別、単価、操作順を確認する。回答前に [ItemListCard](frontend/src/features/accounting/components/ItemListCard.tsx) の商品選択と [請求チェック](frontend/src/features/medical-records/lib/medical-record-bill-check-model.ts) の導線を整理する。
- **手順:** 診療項目・薬剤・商品・ワクチンを識別し、該当画面の保存 request → 再読込した単価 → カルテの治療/接種 → 会計未請求 → 請求明細の順に追う。会計の直接選択は商品が対象のため、保存漏れと操作経路の違いを分ける。
- **成果物・完了条件:** 対象経路、期待値、実際値、原因、最小修正または案内を同じ ID に記録。修正する場合は保存・再読込・金額連携の失敗テストから Docker scoped 検証へ進む。登録経路が未特定の間は実装修正を止め、全マスタを会計選択に混在させない。根拠は [医院回答](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-r2-master-path)。

### UAT-R2-EXCLUSIVE-LOCK

- **入力・先行準備:** PO/医院に防ぎたい事故（上書き、二重会計、同時診療入力、未保存離脱）と対象画面・占有時間を確認する。回答前に [clinical_plan の競合判定](backend/internal/medicalrecord/clinical_plan_repository.go)、[会計の行ロック](backend/internal/billing/accounting_repository.go)、[NavigationBlocker の呼出元](frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) を対応表にする。
- **手順:** 合意した1場面を合成データと2セッションで再現し、競合検出・再読込・再入力で目的を満たせるかを先に判断する。不足時だけロック対象、期限、切断/異常終了時の解放、権限、監査、復旧方法を設計する。
- **成果物・完了条件:** named owner の仕様判断と再現例、採用案、競合・取消・再接続の受入条件を同じ ID に残す。製品判断がない間は全面ロックを実装しない。実装を採用した場合のみ、2セッションの上書き/二重処理防止と医院分離を scoped テスト・対象環境で検証する。[背景](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-r2-exclusive-lock)。

### UAT-R2-CHART-FIT

- **入力・先行準備:** 解像度、ブラウザのズーム、対象タブ、サイドバーの希望状態を確認する。回答前に [カルテ配置](frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) と [Sidebar](frontend/src/components/shared/Layout/Sidebar.tsx) の現行高さ・展開条件を整理する。
- **手順:** 確定した viewport で overflow を再現し、カルテ内の余白・行高・タブ周辺の密度を小さい差分で調整する。自動 collapse と医院の希望が衝突する場合は PO に判断を戻す。
- **成果物・完了条件:** 同じ端末条件の変更前後、可視範囲、操作数を記録し、必須情報・保存操作・キーボードフォーカスが届くことを確認する。情報/タブの削除や小さすぎる文字で収めない。端末条件未確定ではレイアウト実装を止める。[背景](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-r2-chart-fit)。

## データ・移行の残件

### UAT-Q3-GENDER-MAP

[元の照合記録](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-q3-gender-map-旧性別コード-34-を雄雌へ直す) では旧コード 3/4 が不明へ落ちる。old_db の現行実装・テストを producer 担当が確認し、未修正なら承認されたデコードと一致させる。今回 old_db リポジトリの現行状態は未照合なので実装完了とはしない。既存 STG の訂正は [運用 TODO](todo-operations.md#uat-data-operations) に従う。完了条件は雄/雌の復元、手術日の捏造なし、訂正後の件数・画面証拠。

### UAT-Q2-VACCINE-SPECIES

- **入口・先行準備:** [種 filter の repository](backend/internal/medicalrecord/vaccine_repository.go)、[治療マスタ取得](frontend/src/hooks/use-treatment-master.ts)、[接種フォーム](frontend/src/features/vaccinations/hooks/use-vaccination-form.ts) で species の送受信と履歴参照を追う。承認前に、医院別の件数・種欠損・参照不一致を出す集計案を用意する。
- **手順・完了条件:** [読取条件](todo-operations.md#uat-data-operations) が揃ったら、[元記録の A/B/C 分類](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-q2-vaccine-species-猫に犬用ワクチンが出る件) に件数と参照関係を対応づけ、マスタ欠損・参照違い・旧記録由来を分ける。機密除去した分類表と訂正要否を成果物にし、実装が必要な分類だけ同じ ID で仕様/検証を確定する。履歴の一括削除や名前だけからの種の推測補完はしない。

### UAT-Q4-UNPAID-TRIAGE

- **入口・先行準備:** [未納 API](backend/internal/billing/accounting_handler.go) と [未納額計算](backend/internal/billing/unpaid_amount.go) を確認し、status・支払有無・金額帯ごとの集計案を作る。患者名や請求本文は出力しない。
- **手順・完了条件:** [読取条件](todo-operations.md#uat-data-operations) が揃ったら、[元記録](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-q4-unpaid-triage-未納の切り分け消さない) に沿って旧未精算と支払未紐付けを分類し、分類前後の総件数・金額が一致することを確認する。原因別集計と訂正要否を同じ ID に残す。集計で原因が確定しなければ追加照合条件を記録し、請求の一括完了や未納タブの隠蔽へ進めない。

### UAT-Q2-TREATMENTS-IMPORT

[移行範囲の判断記録](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-q2-treatments-import-処置処方の移行今期外候補) を基に、PO が必要な明細・対象期間・責任者・受入条件を決定するまで deferred。先行準備は [21表契約](backend/internal/csvimport/cutover_contract.go) と不足する処置・処方項目の差分表までとする。

再開後は、old_db source → producer → AE import → 過去カルテ表示の項目対応、数量/単位/金額、旧IDと医院境界、重複防止、訂正/復旧を契約化する。合成 fixture で件数・参照・金額を検証してから、承認された disposable rehearsal で保存→再読込→表示を確認する。成果物は PO の範囲判断、両 repo の契約差分、検証結果。実データ投入は別承認とし、21表契約を先行拡張したり、既存の過去カルテ導線を作り直したりしない。

### PO-PET-DECEASED-DATA-BACKFILL

[修復計画](bug.md#plan-po-pet-deceased-data-backfill) の対象・死亡日の根拠・監査・復旧方法を確定する。死亡 write ガードの実装は再開しない。訂正日は推測で埋めず、承認されたデータ操作と照合が完了するまで残件として維持する。

### TASK-444-ADDENDUM-CODEGEN

カルテ追記 response 型は現在 tygo 対象外。先行調査では [response DTO](backend/internal/medicalrecord/medical_record_addendum_response.go)、[tygo 設定](backend/tygo.yaml)、[生成済み response 型](frontend/src/types/generated/medicalrecord-responses.ts) を比較し、対象型と利用箇所を列挙する。生成経路を変更する必要性と別スコープを確定した場合だけ実装する。user-run `make codegen` の後、生成差分が対象契約に限定されることと該当 API の型を確認し、再生成で差分が増えないことを完了条件とする。エージェントの自動 codegen や生成物の手編集で代用しない。背景は [裁定記録](docs/work/development-task-decisions.md#task-444)。

## 更新規則

ID ごとに状態・次の一手・完了条件・根拠を持つ。完了項目は削除し、未完了の検証・運用は専用 TODO へ参照を残す。既存 claim は [AGENTS.md](AGENTS.md) に従って確認し、別担当の作業を取り込まない。
