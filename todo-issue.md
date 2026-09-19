# 未完了 Issue 台帳（repo 正本）

最終照合: 2026-09-18。新規の実装・調査・PO 課題を管理する。完了した実装・仕様どおりの報告・回答済みの操作案内は削除し、履歴は Git と元の UAT 記録を参照する。検証は [todo-verification.md](todo-verification.md)、外部実行は [todo-operations.md](todo-operations.md) を正本とする。

既存 Linear は [BRT-4](https://linear.app/baritechllc/issue/BRT-4) 配下を更新する。新規 Issue を作らない方針を維持し、外部投稿は別途承認後。以下の新規 ID に専用 Linear Issue は今回割り当てていない。

<a id="open"></a>

## Open

着手プラン照合: 2026-09-18。未完了の9 ID を点検し、以下に準備単位・判断入力・調査出力を補完した。治療 Enter と検索一覧高さは [実機受入](todo-verification.md#uat-followup) が残る。下表の状態は本体タスクの状態で、準備着手の可否は次表を参照する。

| ID | 状態 | 残作業・次の一手 |
|---|---|---|
| [UAT-R2-MASTER-PATH](#uat-r2-master-path) | BLOCKED（医院入力） | 登録したマスタ種別・画面・単価・操作順を特定し、保存漏れと正しい会計導線を分離 |
| [UAT-R2-EXCLUSIVE-LOCK](#uat-r2-exclusive-lock) | PO 判断待ち | 二重入力・上書きの防止目的と占有範囲を確認し、既存の競合制御で不足する場面を定義 |
| [UAT-R2-CHART-FIT](#uat-r2-chart-fit) | BLOCKED（端末条件） | ノートPC・125% は記録済み。解像度・対象タブを確認してから余白・行高を調整 |
| [UAT-Q3-GENDER-MAP](#uat-q3-gender-map) | 未完了（統合・運用） | 隔離候補のローカル検証済み。old_db main 統合・bundle と承認後の STG 訂正を進める |
| [UAT-Q2-VACCINE-SPECIES](#uat-q2-vaccine-species) | 調査待ち | 承認された STG 集計で、マスタ種欠損・参照取り違え・旧記録そのものを分類 |
| [UAT-Q4-UNPAID-TRIAGE](#uat-q4-unpaid-triage) | 調査待ち | 承認された STG 集計で、旧未精算と支払未紐付けを切り分け |
| [UAT-Q2-TREATMENTS-IMPORT](#uat-q2-treatments-import) | BLOCKED（範囲詳細） | 処置移行は今期に含む。種類・対象期間・責任者・受入条件を確定 |
| [PO-PET-DECEASED-DATA-BACKFILL](#po-pet-deceased-data-backfill) | 限定訂正待ち | 日付根拠がある行だけ対象。対象・監査・復旧・実行承認を確定 |
| [TASK-444-ADDENDUM-CODEGEN](#task-444-addendum-codegen) | DEFERRED | 型生成経路の別スコープと user-run codegen が認められたときに再開 |

既知の医院入力と source 調査を反映した。追加の医院・PO 事実は今回得られていない。今回の文書整理は、実装、DB 操作、医院への送信を実施する承認ではない。医院側の要件責任者名は元記録に未記載であり、担当者が実装前に確認する。

### 今着手する準備単位

| ID | 次の担当 / ローカルで作る成果物 | 本体の開始条件 |
|---|---|---|
| UAT-R2-MASTER-PATH | 開発: 既知の2導線に医院の1操作を対応づける再現票。未回答欄は下の確定票へ | 登録経路・期待単価・再現条件・要件責任者の確定 |
| UAT-R2-EXCLUSIVE-LOCK | 開発: 所見の古い保存、二重精算、未保存離脱の3ケースについて既存防御と不足を記した判断票 | PO が防止する事故・対象資源・受入例を選ぶ。既存防御で足りれば追加実装なし |
| UAT-R2-CHART-FIT | 開発/QA: 対象タブ・CSS viewport・ズーム・サイドバー状態・隠れる操作を記す再現票 | 医院端末の実測値と対象タブ。候補解像度を医院の値として採用しない |
| UAT-Q3-GENDER-MAP | producer/開発: 既存候補の4パス、base、差分、untracked test、検証証拠の引継ぎ一覧 | 所有者の引継ぎと統合判断。bundle/STG は [運用計画](todo-operations.md#uat-q3-gender-map) |
| UAT-Q2-VACCINE-SPECIES | 開発: 下の分類規則・出力列から read-only 集計案と合成期待表を作る | 医院/期間/読取範囲の確定。種の臨床判断と履歴訂正は集計後 |
| UAT-Q4-UNPAID-TRIAGE | 開発: 下の未納式に基づく集計案・二重計上検出・合成期待表を作る | 医院/期間/読取範囲の確定。回収済み事実は旧記録と照合 |
| UAT-Q2-TREATMENTS-IMPORT | 両 repo 担当: 下の契約差分表を source 列・変換・参照先・未解決の4列で具体化 | PO/医院が種類・期間・目的・受入者を確定 |
| PO-PET-DECEASED-DATA-BACKFILL | 開発/運用: 根拠あり/なし・競合・既訂正の合成ケースと dry-run/復旧の設計 | 日付根拠のある対象集合・対象環境・監査・実行承認 |
| TASK-444-ADDENDUM-CODEGEN | DEFERRED 維持。必要性が提起された場合に DTO/tygo/利用箇所の差分を提示 | 別スコープの採用と user-run codegen の実行者・検証範囲 |

<a id="decision-inputs"></a>

### 医院・PO 入力の確定票

回答先は同じ ID。担当者はこの表を依頼文の下書きに使い、既知の回答を再質問しない。個人名や原票を公開台帳へ転記せず、必要なら権限制限された要件記録の参照を残す。

| ID | 確定済み | 回答が必要な欄 / 回答後の判断 |
|---|---|---|
| UAT-R2-MASTER-PATH | 商品は会計へ直接、診療項目/薬剤はカルテ治療経由 | マスタ種別、登録画面、入力額、保存→再読込の値、会計へ追加した操作順、期待値、要件責任者。保存不良なら修正、正しい保存で入口違いなら手順案内 |
| UAT-R2-EXCLUSIVE-LOCK | 画面占有ロックなし。version 競合検出・会計行ロック・離脱警告は既存 | 同じカルテ/会計で起きる具体例、止めたい操作、解除条件、要件責任者。まず競合拒否で足りるか裁定し、不足する1場面だけ設計 |
| UAT-R2-CHART-FIT | ノートPC、ブラウザ125%、サイドバー展開希望 | OS表示倍率、画面解像度、ブラウザ表示領域（CSS px）、対象タブ、見えない情報/操作、要件責任者。ブラウザ125%だけから viewport を算出しない |
| UAT-Q2-TREATMENTS-IMPORT | 処置移行は今期対象。現行21表に処置実績/処方なし | 処置・注射・処方それぞれの採否、起止日と日付基準、参照専用か請求再利用か、数量/単位/金額の必要性、欠損時の扱い、要件責任者/受入者。契約案への回答後に実装範囲を固定 |

未回答は推奨案で埋めず、その欄を本体の BLOCKED 理由として残す。全面ロック、全マスタの会計選択への混在、日付捏造の採用を前提にしない。

<a id="data-investigation-contracts"></a>

### 集計設計の確定事項（9月18日の source 照合）

**ワクチン種:** 現行 [接種フォーム](frontend/src/features/vaccinations/hooks/use-vaccination-form.ts) は有効マスタだけを選び、[取得 hook](frontend/src/hooks/use-treatment-master.ts) は species を送らない。[repository](backend/internal/medicalrecord/vaccine_repository.go) の species 指定は完全一致で、`cat` に `both` が自動で含まれる契約ではない。単に `species=cat` を追加する修正計画では既存の受入条件を満たさない。old_db `030_stage.sql` の種 NULL と、アプリ選択条件の問題を分ける。

- 集計単位は対象医院内の接種。pet/vaccine の参照先も同一医院かを検証し、不一致・削除済み参照を通常集計へ混ぜない。出力は医院の匿名ラベル、ペット種、マスタ種（欠損を独立）、承認されたマスタ識別子、件数、参照不一致件数。個体名・履歴本文・実個体 ID は共有しない。
- A=マスタ種欠損、B=旧記録との照合で参照違いを確認、C=旧記録と一致、未照合=UNKNOWN。A と C は同時に成立し得るので「種品質」と「履歴照合結果」を別列にする。名称だけから B/C を決めない。
- 合成期待表は猫/犬 × `cat`/`dog`/`both`/未設定、無効マスタ、他医院参照、既存履歴を含める。調査後の選択 UI は猫に `cat`/`both`、犬に `dog`/`both` を残す。未設定・犬猫以外の扱いは医院判断を記録し、既存履歴を新規候補の条件で消さない。

**未納:** [現行計算](backend/internal/billing/unpaid_amount.go) に合わせ、payment なしの waiting は `billings.total_amount`、payment ありの waiting/completed は `max(0, total_amount - insurance_amount - discount_amount - billing_amount)`（payment の各額）で算出する。他 status は0。completed の訂正残も対象であり、waiting 件数だけで未納全体を説明しない。

- 対象医院・支払予定日の範囲を [未納一覧仕様](docs/spec/screens/30-unpaid-list.md) と合わせる。請求1件1行を母集団にし、同一医院の未削除 payment との対応件数を検査する。複数対応・医院不一致は別の異常件数に出して停止し、payment_splits の直接 JOIN で請求額を増幅しない。
- 出力は status、payment 有無、未納額帯（0円 / 1–9,999円 / 10,000円以上）、請求件数、請求額合計、未納額合計、旧記録との照合済み/未照合。金額帯は調査用区分で、回収判断や製品仕様の変更ではない。
- 合成期待表: waiting/paymentなし/請求1,000円→未納1,000円、completed/paymentあり/総額1,000円・保険0・割引0・支払額600円→400円、同支払額1,000円→0円。waiting/paymentありにも同じ差額式を使う。分類前後の件数と未納額を照合し、payment の有無だけで「実未納/突合失敗」と断定しない。

以上はローカル集計設計の入力。実データの取得・更新は [運用 TODO](todo-operations.md#uat-data-operations) の条件を満たしてから行う。

## 第2報の着手プラン

### UAT-R2-MASTER-PATH

- **入力待ち:** 医院が使った登録画面、マスタ種別、単価、操作順は未特定。
- **source 調査済み（9月17日）:** 商品は [ItemListCard](frontend/src/features/accounting/components/ItemListCard.tsx) → `useGetAllMerchandiseItems` → [useAccountingItemActions](frontend/src/features/accounting/hooks/use-accounting-item-actions.ts) → [createBillingItem](frontend/src/features/accounting/api/create-billing-item.ts) で手動請求へ入る。診察・処置・薬剤は [useTreatmentMaster](frontend/src/hooks/use-treatment-master.ts) → [useTreatmentsTab](frontend/src/features/medical-records/hooks/use-treatments-tab.ts) → [治療 API](frontend/src/features/medical-records/api/treatments.ts) → [確定カルテの未請求治療](backend/internal/medicalrecord/treatment_repository.go) → [請求連携](backend/internal/billing/billing_item_unbilled.go) を通る。商品・治療単価は負数を拒否し0を許容する。`isUnbillableMasterPrice` の null/非有限/負数判定は [請求チェックの検査・ワクチン候補](frontend/src/features/medical-records/lib/medical-record-bill-check-model.ts) の扱いで、全治療 DTO の単価を nullable とする根拠にはしない。
- **手順:** 診療項目・薬剤・商品・ワクチンを識別し、該当画面の保存 request → 再読込した単価 → カルテの治療/接種 → 会計未請求 → 請求明細の順に追う。会計の直接選択は商品が対象のため、保存漏れと操作経路の違いを分ける。
- **成果物・完了条件:** 対象経路、期待値、実際値、原因、最小修正または案内を同じ ID に記録。修正する場合は保存・再読込・金額連携の失敗テストから Docker scoped 検証へ進む。登録経路が未特定の間は実装修正を止め、全マスタを会計選択に混在させない。根拠は [医院回答](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-r2-master-path)。

### UAT-R2-EXCLUSIVE-LOCK

- **入力・先行準備:** PO/医院に防ぎたい事故（上書き、二重会計、同時診療入力、未保存離脱）と対象画面・占有時間を確認する。全面ロックは導入しない方針を維持し、目的・具体的な衝突事例は未確定。
- **source 調査済み（9月17日）:** [カルテ保存](frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts) の loaded version → [clinical_plan](backend/internal/medicalrecord/clinical_plan_repository.go) の `expectedVersion` 比較は古い保存を拒否する。[会計](backend/internal/billing/accounting_repository.go) と [明細](backend/internal/billing/billing_item_repository.go) の `FOR UPDATE` は取引内の更新を直列化する。[ReadyPanels](frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) の dirty 状態 → [NavigationBlocker](frontend/src/components/shared/NavigationBlocker/NavigationBlocker.tsx) と [beforeunload](frontend/src/hooks/use-unsaved-changes.ts) は未保存離脱の警告。いずれも、他端末が画面を開くことを禁止する占有ロックではない。
- **手順:** 合意した1場面を合成データと2セッションで再現し、競合検出・再読込・再入力で目的を満たせるかを先に判断する。不足時だけロック対象、期限、切断/異常終了時の解放、権限、監査、復旧方法を設計する。
- **成果物・完了条件:** named owner の仕様判断と再現例、採用案、競合・取消・再接続の受入条件を同じ ID に残す。製品判断がない間は全面ロックを実装しない。実装を採用した場合のみ、2セッションの上書き/二重処理防止と医院分離を scoped テスト・対象環境で検証する。[背景](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-r2-exclusive-lock)。

### UAT-R2-CHART-FIT

- **入力・先行準備:** ノートPC・ブラウザ125% は記録済み。解像度、対象タブ、サイドバーの希望状態を確認する。回答前に [カルテ配置](frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) と [Sidebar](frontend/src/components/shared/Layout/Sidebar.tsx) の現行高さ・展開条件を整理する。
- **手順:** 確定した viewport で overflow を再現し、カルテ内の余白・行高・タブ周辺の密度を小さい差分で調整する。自動 collapse と医院の希望が衝突する場合は PO に判断を戻す。
- **成果物・完了条件:** 同じ端末条件の変更前後、可視範囲、操作数を記録し、必須情報・保存操作・キーボードフォーカスが届くことを確認する。情報/タブの削除や小さすぎる文字で収めない。端末条件未確定ではレイアウト実装を止める。[背景](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-r2-chart-fit)。

## データ・移行の残件

### UAT-Q3-GENDER-MAP

[元の照合記録](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-q3-gender-map-旧性別コード-34-を雄雌へ直す) に対し、9月17日に producer の source と隔離候補を照合した。元の old_db `main` `2eab89ac` は1/01→male・2/02→femaleのみ。`../old_db-codex-uat-q3-gender-map-20260917`（`fix/codex-uat-q3-gender-map-20260917`）の未コミット候補は3/03→male・4/04→femaleを追加し、classifier/oracle も一致する。`PetOpe_Date` と1900年ガードは変更していない。

[候補の検証報告](.planning/agent-fast-campaign/remaining-local-reconciliation-20260917/evidence/gender-001/receiver/Completion%20Report.md) は38 mapping tests と80% gateが PASS。SQL CASE は SQLite での検証に限り、PostgreSQL・export・STG は未検証。既存の別文書に起因する global sensitive-scan FAIL は残り、変更4パスの scoped scan は PASS。ローカル候補の検証完了を main 統合・bundle 作成・実データ訂正の完了にしない。残る統合・現行契約 bundle と既存 STG の訂正は [運用 TODO](todo-operations.md#uat-q3-gender-map) へ。受入は雄/雌の復元、手術日の捏造なし、訂正後の件数・画面証拠。

9月18日の読取でも producer main は `2eab89a`、候補は未コミット。引継ぎ対象は `sql/migration/030_stage.sql`、`scripts/lib/mapping-conversion-families.mjs`、`scripts/test-support/test-mapping-conversion-families.mjs`、**untracked の `scripts/test-support/test-pet-gender-mapping.mjs`** の4パス。producer main の別作業 WIP は対象外。通常の `git diff` だけでは4本目が抜けるため、所有者・保存先・4本の hash・既存検証との一致を確認してから統合案を作る。今回 old_db の編集・統合はしていない。

### UAT-Q2-VACCINE-SPECIES

- **入口・先行準備:** [種 filter の repository](backend/internal/medicalrecord/vaccine_repository.go)、[治療マスタ取得](frontend/src/hooks/use-treatment-master.ts)、[接種フォーム](frontend/src/features/vaccinations/hooks/use-vaccination-form.ts) で species の送受信と履歴参照を追う。承認前に、医院別の件数・種欠損・参照不一致を出す集計案を用意する。
- **手順・完了条件:** [読取条件](todo-operations.md#uat-data-operations) が揃ったら、[元記録の A/B/C 分類](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-q2-vaccine-species-猫に犬用ワクチンが出る件) に件数と参照関係を対応づけ、マスタ欠損・参照違い・旧記録由来を分ける。機密除去した分類表と訂正要否を成果物にし、実装が必要な分類だけ同じ ID で仕様/検証を確定する。履歴の一括削除や名前だけからの種の推測補完はしない。

### UAT-Q4-UNPAID-TRIAGE

- **入口・先行準備:** [未納 API](backend/internal/billing/accounting_handler.go) と [未納額計算](backend/internal/billing/unpaid_amount.go) を確認し、status・支払有無・金額帯ごとの集計案を作る。患者名や請求本文は出力しない。
- **手順・完了条件:** [読取条件](todo-operations.md#uat-data-operations) が揃ったら、[元記録](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-q4-unpaid-triage-未納の切り分け消さない) に沿って旧未精算と支払未紐付けを分類し、分類前後の総件数・金額が一致することを確認する。原因別集計と訂正要否を同じ ID に残す。集計で原因が確定しなければ追加照合条件を記録し、請求の一括完了や未納タブの隠蔽へ進めない。

### UAT-Q2-TREATMENTS-IMPORT

[移行範囲の判断記録](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-q2-treatments-import-処置処方の移行今期外候補) の後、**処置移行を今期に含める**方針が記録された。種類・対象期間・責任者・受入条件は未確定で、詳細確定まで実装を保留する。先行準備は [21表契約](backend/internal/csvimport/cutover_contract.go) と不足する処置・処方項目の差分表までとする。

9月18日の契約照合では、`procedures`（マスタ）と `billing_items`（請求明細）は含むが `treatments` / `prescriptions` は含まない。既存表があることを診療実績の移行済み根拠にしない。両 repo 担当は次表の不足だけを設計資料へ展開する。

| 対応対象 | 現行契約で分かること | 契約案で埋める欄 |
|---|---|---|
| 処置/注射マスタ | `procedures` は既存 | 旧マスタキー→既存 ID の対応と、不明な参照の扱い |
| 診療ごとの処置実績 | `treatments` は対象外 | 元表/列、旧カルテ・ペット・医院との FK、実施日、数量/単位、単価/金額、重複防止キー |
| 処方 | `prescriptions` は対象外 | 今期採否、元表/列、薬剤参照、用量/用法/日数、旧表現の保持と単位変換可否 |
| 会計・過去表示 | `billing_items` と過去カルテ導線は既存 | 参照専用/再請求可の判断、二重請求防止、表示先、欠損時の表示、件数/金額の照合規則 |

元表/列は repo 内の schema/変換定義から確認し、分からない列を推測しない。合成 fixture の受入案は「同一医院・同一ペット・同一旧カルテへの帰属」「数量/単位/金額の原本一致」「再取込で重複しない」「欠損と未移行が区別できる」。医院の採否を得るまでは案であり、契約の表数・hash・生成物を変更しない。

詳細確定後は、old_db source → producer → AE import → 過去カルテ表示の項目対応、数量/単位/金額、旧IDと医院境界、重複防止、訂正/復旧を契約化する。合成 fixture で件数・参照・金額を検証してから、承認された disposable rehearsal で保存→再読込→表示を確認する。成果物は PO の範囲判断、両 repo の契約差分、検証結果。実データ投入は別承認とし、21表契約を先行拡張したり、既存の過去カルテ導線を作り直したりしない。

### PO-PET-DECEASED-DATA-BACKFILL

[修復計画](bug.md#plan-po-pet-deceased-data-backfill) は **死亡日の根拠がある行だけ訂正**する方針。根拠がない行は補完対象外のまま残し、対象・監査・復旧方法・実行承認を確定する。死亡 write ガードの実装は再開しない。訂正日は推測で埋めず、承認されたデータ操作と照合が完了するまで残件として維持する。

### TASK-444-ADDENDUM-CODEGEN

カルテ追記 response 型は現在 tygo 対象外。先行調査では [response DTO](backend/internal/medicalrecord/medical_record_addendum_response.go)、[tygo 設定](backend/tygo.yaml)、[生成済み response 型](frontend/src/types/generated/medicalrecord-responses.ts) を比較し、対象型と利用箇所を列挙する。生成経路を変更する必要性と別スコープを確定した場合だけ実装する。user-run `make codegen` の後、生成差分が対象契約に限定されることと該当 API の型を確認し、再生成で差分が増えないことを完了条件とする。エージェントの自動 codegen や生成物の手編集で代用しない。背景は [裁定記録](docs/work/development-task-decisions.md#task-444)。

## 更新規則

ID ごとに状態・次の一手・完了条件・根拠を持つ。完了項目は削除し、未完了の検証・運用は専用 TODO へ参照を残す。既存 claim は [AGENTS.md](AGENTS.md) に従って確認し、別担当の作業を取り込まない。
