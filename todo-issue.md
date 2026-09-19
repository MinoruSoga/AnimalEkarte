# 未完了 Issue 台帳（repo 正本）

最終照合: 2026-09-19。新規の実装・調査・PO 課題を管理する。完了した実装・仕様どおりの報告・回答済みの操作案内は削除し、履歴は Git と元の UAT 記録を参照する。検証は [todo-verification.md](todo-verification.md)、外部実行は [todo-operations.md](todo-operations.md) を正本とする。

既存 Linear は [BRT-4](https://linear.app/baritechllc/issue/BRT-4) 配下を更新する。新規 Issue を作らない方針を維持し、外部投稿は別途承認後。以下の新規 ID に専用 Linear Issue は今回割り当てていない。

<a id="open"></a>

## Open

着手プラン照合: 2026-09-19。依頼者の追加回答を反映し、4件の「医院に範囲を聞く」待ちを解除した。READY は明記した調査・設計単位の開始可否であり、実環境操作や実装完了ではない。治療 Enter と検索一覧高さは [実機受入](todo-verification.md#uat-followup) が残る。

| ID | 状態 | 残作業・次の一手 |
|---|---|---|
| [UAT-R2-MASTER-PATH](#uat-r2-master-path) | READY（全経路の検証設計） | 金額を持つ全マスタ登録ページを列挙し、新規/編集→再読込→利用→会計を検証。発生ページの回答を待たない |
| [UAT-R2-EXCLUSIVE-LOCK](#uat-r2-exclusive-lock) | READY（競合ケース・不足実装の特定） | カルテ上書きと二重会計の両方を防ぐ。2セッションと再送のケースを既存防御/テストに対応づける |
| [UAT-R2-CHART-FIT](#uat-r2-chart-fit) | READY（寸法再現・UI修正設計） | Windows 8 / Chrome、15.6インチ、1366×625。カルテ全9タブを対象にし、旧Chrome互換性とレイアウトを分離 |
| [UAT-Q3-GENDER-MAP](#uat-q3-gender-map) | 未完了（bundle・運用） | old_db main 統合を確認済み。現行bundle・PostgreSQL検証証拠を照合し、承認後の STG 訂正へ |
| [UAT-Q2-VACCINE-SPECIES](#uat-q2-vaccine-species) | 調査待ち | 承認された STG 集計で、マスタ種欠損・参照取り違え・旧記録そのものを分類 |
| [UAT-Q4-UNPAID-TRIAGE](#uat-q4-unpaid-triage) | 調査待ち | 承認された STG 集計で、旧未精算と支払未紐付けを切り分け |
| [UAT-Q2-TREATMENTS-IMPORT](#uat-q2-treatments-import) | READY（全期間の移行契約設計） | 処置マスタ・患者ごとの処置履歴を全種類・全期間対象として旧列→producer→AE→履歴表示を対応づける |
| [PO-PET-DECEASED-DATA-BACKFILL](#po-pet-deceased-data-backfill) | 限定訂正待ち | 日付根拠がある行だけ対象。対象・監査・復旧・実行承認を確定 |
| [TASK-444-ADDENDUM-CODEGEN](#task-444-addendum-codegen) | DEFERRED | 型生成経路の別スコープと user-run codegen が認められたときに再開 |

要件出所は2026-09-19の本会話の依頼者回答（下表）。今回は文書更新のみで、実装・検証・移行は未実施。個人名を伴う要件責任者・受入者の参照は実行担当が製品仕様変更前に記録する。これを、範囲の再質問やローカル調査の停止理由にはしない。共有STG/PROD操作、外部送信、実データ移行の承認は別に必要。

### 今着手する単位

| ID | 次の担当 / ローカルで作る成果物 | 本体の開始条件 |
|---|---|---|
| UAT-R2-MASTER-PATH | 開発/QA: [全経路票](docs/work/todo-campaign-20260918/UAT-R2-MASTER-PATH.md) を現行route/formへ対応づけ、未カバー経路の失敗テストを用意 | 実行は候補mount・専用fixture・後処理を固定。修正は再現した不一致に限定 |
| UAT-R2-EXCLUSIVE-LOCK | 開発: [競合ケース](docs/work/todo-campaign-20260918/UAT-R2-EXCLUSIVE-LOCK.md) の既存テスト・未カバー箇所と最小修正案 | 2セッションの専用環境。全面占有ロックは採用せず、サーバー側の競合拒否・重複防止を優先 |
| UAT-R2-CHART-FIT | 開発/QA: [寸法・互換性票](docs/work/todo-campaign-20260918/UAT-R2-CHART-FIT.md) の全9タブについてoverflow原因と最小UI修正案 | 1366×625で設計開始可。実機Chrome版/CSS表示領域は検証担当が採取、旧ブラウザ対応方針は別ゲート |
| UAT-Q3-GENDER-MAP | producer/開発: 統合済みrevisionを使用するbundle/検証receiptの照合 | 再統合は不要。bundle/STG は [運用計画](todo-operations.md#uat-q3-gender-map) |
| UAT-Q2-VACCINE-SPECIES | 集計設計済み。[既存票](docs/work/todo-campaign-20260918/UAT-Q2-VACCINE-SPECIES.md) を使い、対象医院/期間/読取条件を実行担当が埋める | 条件未充足なら待機。同じ集計設計を作り直さない |
| UAT-Q4-UNPAID-TRIAGE | 集計設計済み。[既存票](docs/work/todo-campaign-20260918/UAT-Q4-UNPAID-TRIAGE.md) を使い、対象医院/期間/読取条件を実行担当が埋める | 条件未充足なら待機。回収済み事実は旧記録との照合が必要 |
| UAT-Q2-TREATMENTS-IMPORT | 両 repo 担当: [契約差分票](docs/work/todo-campaign-20260918/UAT-Q2-TREATMENTS-IMPORT.md) の旧列・FK・重複キー・精度・参照専用表示を具体化 | 全種類・全期間の設計開始可。原本不明列を推定しない。契約変更・実データ投入は別レビュー/承認 |
| PO-PET-DECEASED-DATA-BACKFILL | 設計済み。[既存票](docs/work/todo-campaign-20260918/PO-PET-DECEASED-DATA-BACKFILL.md) の対象/除外条件へ実行担当が根拠を対応づける | 日付根拠のある対象集合・対象環境・監査・実行承認。根拠なしは除外 |
| TASK-444-ADDENDUM-CODEGEN | DEFERRED 維持。必要性が提起された場合に DTO/tygo/利用箇所の差分を提示 | 別スコープの採用と user-run codegen の実行者・検証範囲 |

<a id="decision-inputs"></a>

### 医院・PO 入力の確定票

回答先は同じ ID。担当者はこの表を依頼文の下書きに使い、既知の回答を再質問しない。個人名や原票を公開台帳へ転記せず、必要なら権限制限された要件記録の参照を残す。

| ID | 確定済み（2026-09-19依頼者回答） | 開発/QAが調べること・後段の入力 |
|---|---|---|
| UAT-R2-MASTER-PATH | 「可能性のあるページをすべて検証」。発生ページの特定を開始条件から外す | 金額を持つ登録/編集フォームと下流経路をコードから全列挙。合成額で保存/再読込/利用を追跡し、再現した問題を修正 |
| UAT-R2-EXCLUSIVE-LOCK | 「両方」＝同じカルテの上書きと同じ会計の二重確定を防ぐ | 既存version/状態検証/行ロック/冪等性を対象操作ごとに照合。別端末が異なる再送キーを使うケースも含める |
| UAT-R2-CHART-FIT | Windows 8、Chrome、15.6インチ、1366（横）×625（縦）でも見切れないこと。既報の125%は維持 | 1366×625を再現基準にし全9タブ/ダイアログを検証。物理寸法とCSS viewportを同一視せず、実機値・Chrome版は検証担当が採取 |
| UAT-Q2-TREATMENTS-IMPORT | 「全部」＝処置名/料金のマスタと患者ごとの全期間の処置履歴。期間や処置種類で意図的に絞らない | 対象医院ごとに提供原本の全期間を対応づける。関連注射・薬剤/処方行も棚卸しして抜けを報告。無関係な業務データの全DB移行や過去分の再請求を許可した意味にはしない |

業務目的は、料金の再入力/転記の解消、見切れによる操作不能の解消、競合による診療記録消失/二重請求の防止、過去診療の参照欠落の解消。回答済みの4点は再質問しない。実機の採取、旧列の調査、テストで決められる事項は担当者の作業にする。全面ロック、全マスタの会計選択への混在、日付捏造を採用しない。

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

- **開始範囲確定（9月19日）:** 医院での元の操作は未特定だが、依頼者が「可能性のあるページをすべて検証」と指定したため再現調査を開始する。対象は [全経路票](docs/work/todo-campaign-20260918/UAT-R2-MASTER-PATH.md) の価格持ち12フォーム群と下流画面。1件再現しても残りを未検証のまま閉じない。
- **source 調査済み（9月17日）:** 商品は [ItemListCard](frontend/src/features/accounting/components/ItemListCard.tsx) → `useGetAllMerchandiseItems` → [useAccountingItemActions](frontend/src/features/accounting/hooks/use-accounting-item-actions.ts) → [createBillingItem](frontend/src/features/accounting/api/create-billing-item.ts) で手動請求へ入る。診察・処置・薬剤は [useTreatmentMaster](frontend/src/hooks/use-treatment-master.ts) → [useTreatmentsTab](frontend/src/features/medical-records/hooks/use-treatments-tab.ts) → [治療 API](frontend/src/features/medical-records/api/treatments.ts) → [確定カルテの未請求治療](backend/internal/medicalrecord/treatment_repository.go) → [請求連携](backend/internal/billing/billing_item_unbilled.go) を通る。商品・治療単価は負数を拒否し0を許容する。`isUnbillableMasterPrice` の null/非有限/負数判定は [請求チェックの検査・ワクチン候補](frontend/src/features/medical-records/lib/medical-record-bill-check-model.ts) の扱いで、全治療 DTO の単価を nullable とする根拠にはしない。
- **手順:** フォームごとに新規・編集、保存 request → API応答 → キャッシュに頼らない再読込 → 適用先の金額を追う。診療項目5タブ、薬剤、商品、入院プラン、ケージ、トリミング2種の単価とキャンペーン割引額を区別し、該当するカルテ/検査/接種/入院/トリミング→会計へ照合する。会計に自動連携しない仕様は根拠付きN/Aにし、保存漏れと経路の違いを分ける。
- **成果物・完了条件:** 全対象の新規/編集/再読込/下流結果がPASSまたは根拠付きN/Aで、金額不一致が解消した経路別receipt。FAILは再現テスト→最小修正→Docker scoped検証。環境不足はBLOCKED、未実行は未実行と残す。全マスタを会計選択に混在させない。[元報告](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-r2-master-path)。

### UAT-R2-EXCLUSIVE-LOCK

- **範囲確定（9月19日）:** 防ぐのは「カルテの上書き」と「二重会計」の両方。画面占有は手段として未採用。未保存離脱の既存防御を維持しつつ、[2セッションのケース票](docs/work/todo-campaign-20260918/UAT-R2-EXCLUSIVE-LOCK.md) から不足箇所を特定する。
- **source 調査済み（9月17日）:** [カルテ保存](frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts) の loaded version → [clinical_plan](backend/internal/medicalrecord/clinical_plan_repository.go) の `expectedVersion` 比較は古い保存を拒否する。[会計](backend/internal/billing/accounting_repository.go) と [明細](backend/internal/billing/billing_item_repository.go) の `FOR UPDATE` は取引内の更新を直列化する。[ReadyPanels](frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) の dirty 状態 → [NavigationBlocker](frontend/src/components/shared/NavigationBlocker/NavigationBlocker.tsx) と [beforeunload](frontend/src/hooks/use-unsaved-changes.ts) は未保存離脱の警告。いずれも、他端末が画面を開くことを禁止する占有ロックではない。
- **手順:** 所見以外の更新APIも含め、古い保存、確定と編集の競合、同一キーの再送、別端末の異なるキーによる同一会計/同一未請求項目の確定、取消/通信断後の再試行を対応づける。[complete](backend/internal/billing/accounting_complete.go) の冪等性は同一キーの再送対策であり、異なるキーの二重処理まで防ぐ証拠にしない。まず既存テストの未カバー箇所をREDにし、不足した状態検証・排他・重複防止だけを修正する。
- **成果物・完了条件:** 勝者の保存内容保持、敗者への競合通知と入力救済/再読込手順、1回分だけの請求/支払/業務監査、失敗時rollback、医院分離を2セッションで確認したreceipt。別患者の別業務を不当に止めない。全面ロックを必要と判断した場合だけ別途占有範囲・期限・解放条件を裁定する。[背景](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-r2-exclusive-lock)。

### UAT-R2-CHART-FIT

- **範囲確定（9月19日）:** Windows 8 / Chrome、15.6インチ、1366×625。対象タブを聞き返さず [全9タブ](frontend/src/features/medical-records/routes/medical-record-form-model.ts) とカルテ内ダイアログを対象にする。[カルテ配置](frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) / [Sidebar](frontend/src/components/shared/Layout/Sidebar.tsx) から固定幅・高さ・overflowを調査する。
- **手順:** [再現票](docs/work/todo-campaign-20260918/UAT-R2-CHART-FIT.md) の1366×625 CSS pxをローカル基準とし、展開/折畳み両方、長文/長一覧、保存/エラー/ダイアログを確認。既報125%の実機CSS表示領域は検証担当が測り追加する。折返し・列配置・スクロール責務を調整し、情報を切り捨てるoverflowや一律縮小では解決しない。
- **互換性の別ゲート:** Windows 8対応はChrome109まで。一方現行Tailwind v4の基準はChrome111以降で、Viteは `esnext`。旧端末対応の合格は寸法調整だけでは出せない。根拠と実機確認/対応方針は再現票へ。OS更新を勝手に前提にせず、依存のdowngradeもこのUI修正に混ぜない。
- **成果物・完了条件:** 基準寸法と実機条件それぞれの前後画像、全9タブの必須情報/入力/保存/フォーカスへ到達できる証拠、root横overflowなし・局所スクロールで末尾まで到達可能の確認。小さすぎる文字、タブ削除、常時sidebar折畳みだけでは合格にしない。最新ChromiumでのPASSは旧ChromeのPASSではない。[背景](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-r2-chart-fit)。

## データ・移行の残件

### UAT-Q3-GENDER-MAP

[元の照合記録](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-q3-gender-map-旧性別コード-34-を雄雌へ直す) に対し、9月17日に producer の source と隔離候補を照合した。元の old_db `main` `2eab89ac` は1/01→male・2/02→femaleのみ。`../old_db-codex-uat-q3-gender-map-20260917`（`fix/codex-uat-q3-gender-map-20260917`）の未コミット候補は3/03→male・4/04→femaleを追加し、classifier/oracle も一致する。`PetOpe_Date` と1900年ガードは変更していない。

[候補の検証報告](.planning/agent-fast-campaign/remaining-local-reconciliation-20260917/evidence/gender-001/receiver/Completion%20Report.md) は38 mapping tests と80% gateが PASS。SQL CASE は SQLite での検証に限り、PostgreSQL・export・STG は未検証。既存の別文書に起因する global sensitive-scan FAIL は残り、変更4パスの scoped scan は PASS。ローカル候補の検証完了を main 統合・bundle 作成・実データ訂正の完了にしない。残る統合・現行契約 bundle と既存 STG の訂正は [運用 TODO](todo-operations.md#uat-q3-gender-map) へ。受入は雄/雌の復元、手術日の捏造なし、訂正後の件数・画面証拠。

**9月19日の読取で更新:** old_db の修正コミット `5fbc3b2` は merge `a2cea37` で main に統合済み（main祖先を照合）。対象は `sql/migration/030_stage.sql`、`scripts/lib/mapping-conversion-families.mjs`、`scripts/test-support/test-mapping-conversion-families.mjs`、`scripts/test-support/test-pet-gender-mapping.mjs` の4パスで、最後のテストも追跡済み。上の9月17–18日の「未コミット候補」は履歴であり、再引継ぎ/再統合のタスクは終了。今回old_dbは編集していない。現行bundle・PostgreSQL・STG訂正の新しい実行証拠は未照合なので、その残件だけを維持する。

### UAT-Q2-VACCINE-SPECIES

- **次の作業:** [集計設計票](docs/work/todo-campaign-20260918/UAT-Q2-VACCINE-SPECIES.md) は作成済み。実行担当が対象医院・期間・読取範囲・承認参照を確定し、現行revisionとの設計差分があれば更新する。species取得/履歴参照の既存調査を最初から繰り返さない。
- **手順・完了条件:** [読取条件](todo-operations.md#uat-data-operations) が揃ったら、[元記録の A/B/C 分類](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-q2-vaccine-species-猫に犬用ワクチンが出る件) に件数と参照関係を対応づけ、マスタ欠損・参照違い・旧記録由来を分ける。機密除去した分類表と訂正要否を成果物にし、実装が必要な分類だけ同じ ID で仕様/検証を確定する。履歴の一括削除や名前だけからの種の推測補完はしない。

### UAT-Q4-UNPAID-TRIAGE

- **次の作業:** [集計設計票](docs/work/todo-campaign-20260918/UAT-Q4-UNPAID-TRIAGE.md) は作成済み。実行担当が対象医院・期間・読取範囲・承認参照を確定し、現行revisionとの計算式の差分があれば更新する。患者名や請求本文は出力しない。
- **手順・完了条件:** [読取条件](todo-operations.md#uat-data-operations) が揃ったら、[元記録](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-q4-unpaid-triage-未納の切り分け消さない) に沿って旧未精算と支払未紐付けを分類し、分類前後の総件数・金額が一致することを確認する。原因別集計と訂正要否を同じ ID に残す。集計で原因が確定しなければ追加照合条件を記録し、請求の一括完了や未納タブの隠蔽へ進めない。

### UAT-Q2-TREATMENTS-IMPORT

[移行範囲の判断記録](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-q2-treatments-import-処置処方の移行今期外候補) の後、**処置移行を今期に含める**方針となり、9月19日に依頼者がマスタ/患者別履歴/期間の問いへ「全部」と回答した。**処置名・料金マスタと患者ごとの処置履歴を全種類・全期間**対象として契約設計を開始する。任意の起止日で減らさず、提供原本の最古から最終抽出までを対象医院別に照合する。関連する注射・薬剤/処方行も棚卸し対象にし、臨床項目の不足を明示する。無関係な全DB移行や履歴の再請求は範囲外。

9月18日の契約照合では、`procedures`（マスタ）と `billing_items`（請求明細）は含むが `treatments` / `prescriptions` は含まない。既存表があることを診療実績の移行済み根拠にしない。両 repo 担当は次表の不足だけを設計資料へ展開する。

| 対応対象 | 現行契約で分かること | 契約案で埋める欄 |
|---|---|---|
| 処置/注射マスタ | `procedures` は既存 | 旧マスタキー→既存 ID の対応と、不明な参照の扱い |
| 診療ごとの処置実績 | `treatments` は対象外 | 元表/列、旧カルテ・ペット・医院との FK、実施日、数量/単位、単価/金額、重複防止キー |
| 処方 | `prescriptions` は対象外 | 全処置履歴に付随する原本の元表/列、薬剤参照、用量/用法/日数、旧表現の保持。復元根拠のない用量や単位は作らない |
| 会計・過去表示 | `billing_items` と過去カルテ導線は既存 | 参照専用/再請求可の判断、二重請求防止、表示先、欠損時の表示、件数/金額の照合規則 |

次の成果物は [既存契約差分票](docs/work/todo-campaign-20260918/UAT-Q2-TREATMENTS-IMPORT.md) の列対応の具体化。元表/列は repo 内の schema/変換定義から確認し、不明な列を推測しない。過去分は参照専用・自動再請求なしを設計の安全側既定とする。合成fixtureは同一医院/ペット/旧カルテ帰属、数量/単位/金額の原本保持、再取込時の重複なし、欠損/未移行の区別を必須とする。原本行数＝取込＋理由付き保留＋根拠付き重複除外を全期間で突合し、黙った除外を許さない。履歴が保留のままなら「全部移行完了」としない。

契約案を両repo担当がレビューし、要件責任者/受入者の参照を記録してからproducer・AE importの変更を同一契約で実装する。合成fixtureの件数/参照/金額検証後、承認済みdisposable rehearsalで保存→再読込→既存過去カルテ導線の表示を確認する。実データ投入は別承認。今回の「全部」という範囲確定だけで21表契約/hash/生成物を先行変更したり、migrationを自動適用したりしない。

### PO-PET-DECEASED-DATA-BACKFILL

[修復計画](bug.md#plan-po-pet-deceased-data-backfill) は **死亡日の根拠がある行だけ訂正**する方針。根拠がない行は補完対象外のまま残し、対象・監査・復旧方法・実行承認を確定する。死亡 write ガードの実装は再開しない。訂正日は推測で埋めず、承認されたデータ操作と照合が完了するまで残件として維持する。

### TASK-444-ADDENDUM-CODEGEN

カルテ追記 response 型は現在 tygo 対象外。先行調査では [response DTO](backend/internal/medicalrecord/medical_record_addendum_response.go)、[tygo 設定](backend/tygo.yaml)、[生成済み response 型](frontend/src/types/generated/medicalrecord-responses.ts) を比較し、対象型と利用箇所を列挙する。生成経路を変更する必要性と別スコープを確定した場合だけ実装する。user-run `make codegen` の後、生成差分が対象契約に限定されることと該当 API の型を確認し、再生成で差分が増えないことを完了条件とする。エージェントの自動 codegen や生成物の手編集で代用しない。背景は [裁定記録](docs/work/development-task-decisions.md#task-444)。

## 更新規則

ID ごとに状態・次の一手・完了条件・根拠を持つ。完了項目は削除し、未完了の検証・運用は専用 TODO へ参照を残す。既存 claim は [AGENTS.md](AGENTS.md) に従って確認し、別担当の作業を取り込まない。
