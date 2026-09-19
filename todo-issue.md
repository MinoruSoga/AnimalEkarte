<a id="未完了-issue-台帳repo-正本"></a>

# Issue 台帳（着手区分別・repo 正本）

最終照合: 2026-09-19。新規の実装・調査・PO 課題を管理する。完了した実装・仕様どおりの報告・回答済みの操作案内は 未対応エリアに戻さず、[Slack 出典対応表](#slack-source-map) に処理区分を残す。検証は [todo-verification.md](todo-verification.md)、外部実行は [todo-operations.md](todo-operations.md) を正本とする。

既存 Linear は [BRT-4](https://linear.app/baritechllc/issue/BRT-4) 配下を更新する。新規 Issue を作らない方針を維持し、外部投稿は別途承認後。以下の新規 ID に専用 Linear Issue は今回割り当てていない。

今回の依頼先はこのファイル。[5エリアの着手プラン](#slack-tasks) と [50親投稿の対応表](#slack-source-map) を管理する。Plane への登録は行っていない。

<a id="open"></a>
<a id="slack-tasks"></a>
<a id="slack-全件からの着手プラン2026-09-19"></a>

## 着手区分（5エリア）

各タスク本文は、現在の主な開始条件に従って1か所だけに置く。区分は優先順位や実装完了を示すものではない。待ちエリアでも先行できる調査・準備は本文に残し、複数の条件がある場合は各タスクの停止条件をすべて維持する。

| エリア | 件数 | 読み方 |
|---|---|---|
| [1. 今すぐ調査・設計に着手できる](#area-ready) | 17課題 | ローカル調査・設計を開始できる。後段の実装・実行条件は本文で確認 |
| [2. PO判断待ち](#area-po) | 6課題 | 判断材料は準備できる。仕様を確定せず実装には進まない |
| [3. 資料・環境・実機確認待ち](#area-evidence) | 13課題 | 必要な資料・環境・承認・受入条件を揃える |
| [4. 納品後対応・別スコープ待ち](#area-deferred) | 3課題 | 納品後2件と、納品に依存しない別スコープ1件を区別 |
| [5. 回答済み・実装済み（受入確認は別管理）](#area-resolved) | 7トピック | 回答履歴・既存実装を保持。医院受入完了とは扱わない |

課題本文は計39件（既存9件＋Slack個別プラン30件）。回答済み・実装済みの7トピックは、新しい実装課題を重複作成せず参照で管理する。45トピックとの対応は [トピック別の処理先](#トピック別の処理先) を参照。

着手プラン照合: 2026-09-19。依頼者の追加回答を反映し、4件の「医院に範囲を聞く」待ちを解除した。READY は明記した調査・設計単位の開始可否であり、実環境操作や実装完了ではない。治療 Enter と検索一覧高さは [実機受入](todo-verification.md#uat-followup) が残る。

対象は Vault 相対パス `50_Projects/顧客案件/ノア動物病院電子カルテ/会話ログ/Slack_電子カルテ開発_曽我/2026-08-20以降_不具合等Slack全文.md`（SHA-256 `9d54ab0d7c4432dade86fcb75c82fef55cbe91243745fbe594dc4b543a286e03`）。収録実期間は8月29日〜9月17日、50親投稿＋82返信＝132出現、重複を除く130メッセージ。以下の `source` は同ファイルの行と親 timestamp を指し、返信も対象にする。添付は名称・存在のみを確認しており、画像・PDFの中身は UNKNOWN。個人・患者データ、認証情報、外部シートのURL/本文は転記しない。

9月16日の方針は「不具合＝納品前、要望＝納品後、質問＝操作案内」。分類が未確定の項目は PO 判断待ちとし、全要望を納品前の不具合に昇格させない。以下の READY はローカル調査・設計のみ。実装前に [設計思想](docs/product-philosophy.md) の要件責任者・目的・削減工程・測定方法を確定し、臨床仕様は獣医師/PO の受入者参照を残す。所要時間や合格閾値は未測定で、数値を仮定しない。共有環境、実データ、実機、外部送信は各既存運用ゲートに従う。Slack 内の過去日程・納期の曜日表記は履歴であり、今回の新しい期限ではない。

本節は依頼されたローカル課題整理であり、外部 Q&A シートの複製ではない。今回の登録先は本書のみ。Plane/Linear/Backlogへの書込みは実施していない。外部チケットの現行状態は今回の文書判定に用いない。

<a id="area-ready"></a>
<a id="今着手する単位"></a>
<a id="第2報の着手プラン"></a>

## 1. 今すぐ調査・設計に着手できる

ここで着手できるのは、各本文の「最初の作業」にあるローカル調査・再現条件整理・設計です。実装・共有環境での実行を一律に許可する区分ではありません。バイタルの時刻扱いやコピーの保存意味など、後段のPO判断は各タスクの停止条件を維持します。

既存課題の担当・成果物・開始条件:

| ID | 状態・次の一手 | 担当・ローカル成果物 | 本体の開始条件 |
|---|---|---|---|
| [UAT-R2-MASTER-PATH](#uat-r2-master-path) | READY（全経路の検証設計） / 金額を持つ全マスタ登録ページを列挙し、新規/編集→再読込→利用→会計を検証。発生ページの回答を待たない | 開発/QA: [全経路票](docs/work/todo-campaign-20260918/UAT-R2-MASTER-PATH.md) を現行route/formへ対応づけ、未カバー経路の失敗テストを用意 | 実行は候補mount・専用fixture・後処理を固定。修正は再現した不一致に限定 |
| [UAT-R2-EXCLUSIVE-LOCK](#uat-r2-exclusive-lock) | READY（競合ケース・不足実装の特定） / カルテ上書きと二重会計の両方を防ぐ。2セッションと再送のケースを既存防御/テストに対応づける | 開発: [競合ケース](docs/work/todo-campaign-20260918/UAT-R2-EXCLUSIVE-LOCK.md) の既存テスト・未カバー箇所と最小修正案 | 2セッションの専用環境。全面占有ロックは採用せず、サーバー側の競合拒否・重複防止を優先 |
| [UAT-R2-CHART-FIT](#uat-r2-chart-fit) | READY（寸法再現・UI修正設計） / Windows 8 / Chrome、15.6インチ、1366×625。カルテ全9タブを対象にし、旧Chrome互換性とレイアウトを分離 | 開発/QA: [寸法・互換性票](docs/work/todo-campaign-20260918/UAT-R2-CHART-FIT.md) の全9タブについてoverflow原因と最小UI修正案 | 1366×625で設計開始可。実機Chrome版/CSS表示領域は検証担当が採取、旧ブラウザ対応方針は別ゲート |
| [UAT-Q2-TREATMENTS-IMPORT](#uat-q2-treatments-import) | READY（全期間の移行契約設計） / 処置マスタ・患者ごとの処置履歴を全種類・全期間対象として旧列→producer→AE→履歴表示を対応づける | 両 repo 担当: [契約差分票](docs/work/todo-campaign-20260918/UAT-Q2-TREATMENTS-IMPORT.md) の旧列・FK・重複キー・精度・参照専用表示を具体化 | 全種類・全期間の設計開始可。原本不明列を推定しない。契約変更・実データ投入は別レビュー/承認 |

### UAT-R2-MASTER-PATH

- **Slack 追加根拠:** `1789453043.269699` の9月16日返信は「治療プランから入力したマスタの金額」、`1789540156.256659` はマスタがない場合と0円の質問。[MedicalRecordDiagnosisPlan](frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx#L13-L19) は `treatments` API を使い、[手入力とマスタ単価設定](frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx#L135-L162) がある。同名の backend `treatment_plans` をこの画面の保存先とみなさない。[request](backend/internal/medicalrecord/treatment_request.go#L7-L25) は price を受け、[service](backend/internal/medicalrecord/treatment_service.go#L172-L180) は0以上を許容し、[未請求明細](backend/internal/billing/billing_item_unbilled.go#L124-L142) は治療を扱う。原因は未確定。該当タブでも選択→request→保存→再読込→未請求→確定会計を追い、0円・欠損・未保存を分離する。マスタなしは既存手入力の運用案内を検証し、0円がデモだからとは断定しない。
- **開始範囲確定（9月19日）:** 医院での元の操作は未特定だが、依頼者が「可能性のあるページをすべて検証」と指定したため再現調査を開始する。対象は [全経路票](docs/work/todo-campaign-20260918/UAT-R2-MASTER-PATH.md) の価格持ち12フォーム群と下流画面。1件再現しても残りを未検証のまま閉じない。
- **source 調査済み（9月17日）:** 商品は [ItemListCard](frontend/src/features/accounting/components/ItemListCard.tsx) → `useGetAllMerchandiseItems` → [useAccountingItemActions](frontend/src/features/accounting/hooks/use-accounting-item-actions.ts) → [createBillingItem](frontend/src/features/accounting/api/create-billing-item.ts) で手動請求へ入る。診察・処置・薬剤は [useTreatmentMaster](frontend/src/hooks/use-treatment-master.ts) → [useTreatmentsTab](frontend/src/features/medical-records/hooks/use-treatments-tab.ts) → [治療 API](frontend/src/features/medical-records/api/treatments.ts) → [確定カルテの未請求治療](backend/internal/medicalrecord/treatment_repository.go) → [請求連携](backend/internal/billing/billing_item_unbilled.go) を通る。商品・治療単価は負数を拒否し0を許容する。`isUnbillableMasterPrice` の null/非有限/負数判定は [請求チェックの検査・ワクチン候補](frontend/src/features/medical-records/lib/medical-record-bill-check-model.ts) の扱いで、全治療 DTO の単価を nullable とする根拠にはしない。
- **手順:** フォームごとに新規・編集、保存 request → API応答 → キャッシュに頼らない再読込 → 適用先の金額を追う。診療項目5タブ、薬剤、商品、入院プラン、ケージ、トリミング2種の単価とキャンペーン割引額を区別し、該当するカルテ/検査/接種/入院/トリミング→会計へ照合する。会計に自動連携しない仕様は根拠付きN/Aにし、保存漏れと経路の違いを分ける。
- **成果物・完了条件:** 全対象の新規/編集/再読込/下流結果がPASSまたは根拠付きN/Aで、金額不一致が解消した経路別receipt。FAILは再現テスト→最小修正→Docker scoped検証。環境不足はBLOCKED、未実行は未実行と残す。全マスタを会計選択に混在させない。[元報告](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-r2-master-path)。

### UAT-R2-EXCLUSIVE-LOCK

- **Slack 追加根拠:** `1789453043.269699` の9月16日返信に「別端末で会計待ち→診察中へ戻し、会計後に追加マスタが反映されない」「問い合わせ用に入れた明細の消し忘れ」がある。[会計確認](backend/internal/billing/billing_confirmation_service.go) と受付・カルテの状態遷移、会計確定が読む明細集合を時系列にし、会計を開く→別端末で戻す→追加→確定の再現を既存2セッション票へ追加する。閲覧だけで全面占有する設計は未採用。問い合わせ用の価格確認を請求行へ書かずに済ませる導線も調査し、PO と編集可能状態・確定時の不一致通知・再確認の責任を決める。完了は後追い明細の黙った脱落、二重請求、照会目的の混入が防げる証拠。画面ロックや確認ダイアログ単独を安全性の根拠にしない。
- **範囲確定（9月19日）:** 防ぐのは「カルテの上書き」と「二重会計」の両方。画面占有は手段として未採用。未保存離脱の既存防御を維持しつつ、[2セッションのケース票](docs/work/todo-campaign-20260918/UAT-R2-EXCLUSIVE-LOCK.md) から不足箇所を特定する。
- **source 調査済み（9月17日）:** [カルテ保存](frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts) の loaded version → [clinical_plan](backend/internal/medicalrecord/clinical_plan_repository.go) の `expectedVersion` 比較は古い保存を拒否する。[会計](backend/internal/billing/accounting_repository.go) と [明細](backend/internal/billing/billing_item_repository.go) の `FOR UPDATE` は取引内の更新を直列化する。[ReadyPanels](frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) の dirty 状態 → [NavigationBlocker](frontend/src/components/shared/NavigationBlocker/NavigationBlocker.tsx) と [beforeunload](frontend/src/hooks/use-unsaved-changes.ts) は未保存離脱の警告。いずれも、他端末が画面を開くことを禁止する占有ロックではない。
- **手順:** 所見以外の更新APIも含め、古い保存、確定と編集の競合、同一キーの再送、別端末の異なるキーによる同一会計/同一未請求項目の確定、取消/通信断後の再試行を対応づける。[complete](backend/internal/billing/accounting_complete.go) の冪等性は同一キーの再送対策であり、異なるキーの二重処理まで防ぐ証拠にしない。まず既存テストの未カバー箇所をREDにし、不足した状態検証・排他・重複防止だけを修正する。
- **成果物・完了条件:** 勝者の保存内容保持、敗者への競合通知と入力救済/再読込手順、1回分だけの請求/支払/業務監査、失敗時rollback、医院分離を2セッションで確認したreceipt。別患者の別業務を不当に止めない。全面ロックを必要と判断した場合だけ別途占有範囲・期限・解放条件を裁定する。[背景](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-r2-exclusive-lock)。

### UAT-R2-CHART-FIT

- **範囲確定（9月19日）:** Windows 8 / Chrome、15.6インチ、1366×625。対象タブを聞き返さず [全9タブ](frontend/src/features/medical-records/routes/medical-record-form-model.ts) とカルテ内ダイアログを対象にする。[カルテ配置](frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) / [Sidebar](frontend/src/components/shared/Layout/Sidebar.tsx) から固定幅・高さ・overflowを調査する。
- **手順:** [再現票](docs/work/todo-campaign-20260918/UAT-R2-CHART-FIT.md) の1366×625 CSS pxをローカル基準とし、展開/折畳み両方、長文/長一覧、保存/エラー/ダイアログを確認。既報125%の実機CSS表示領域は検証担当が測り追加する。折返し・列配置・スクロール責務を調整し、情報を切り捨てるoverflowや一律縮小では解決しない。
- **互換性の別ゲート:** Windows 8対応はChrome109まで。一方現行Tailwind v4の基準はChrome111以降で、Viteは `esnext`。旧端末対応の合格は寸法調整だけでは出せない。根拠と実機確認/対応方針は再現票へ。OS更新を勝手に前提にせず、依存のdowngradeもこのUI修正に混ぜない。
- **成果物・完了条件:** 基準寸法と実機条件それぞれの前後画像、全9タブの必須情報/入力/保存/フォーカスへ到達できる証拠、root横overflowなし・局所スクロールで末尾まで到達可能の確認。小さすぎる文字、タブ削除、常時sidebar折畳みだけでは合格にしない。最新ChromiumでのPASSは旧ChromeのPASSではない。[背景](docs/work/stg-uat-clinic-feedback-q1-q4.md#uat-r2-chart-fit)。

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

### SLACK-MANUAL-URINE

- **source / 状態:** MANUAL-URINE。出典212–294の9月9日返信で敷島/猫は尿試験紙を目視、城東/八王子は機器測定と明示。手入力の受入設計 READY。
- **現行根拠・最初の作業:** [検査フォーム](frontend/src/features/examinations/components/ExaminationFormFields.tsx) と [検査items request](backend/internal/medicalrecord/examination_request.go#L158) の値・単位・基準値を追い、手動結果の保存→再読込→カルテ表示を合成fixtureで設計する。自動受信へ無理に統合せず、測定方法と原本由来を保持する。
- **完了 / PO・停止:** 各院の実際の項目/定性表現/単位/基準値について医院が承認した期待表と保存・表示の一致。基準値や陽性/陰性の臨床的意味は推定しない。不明項目は PO/検査担当へ確認し、機器測定結果と手入力結果の混同・上書きを防ぐ。

### SLACK-STAFF-SELECT

- **source / 状態:** STAFF-SELECT。出典643–697、重複725–735。iPadと一部PCで獣医師を選べず、iPadでChromeへ変更しても不可、治療マスタ選択は可能という条件を保持する。再現調査 READY。
- **現行根拠・最初の作業:** [予約フォーム](frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx)、[候補の肯定的capability filter](frontend/src/components/shared/ReservationFormModal/filter-staff-candidates.ts)、[SearchableSelect](frontend/src/components/ui/searchable-select.tsx) を追う。候補0件/取得中/失敗と、候補があるのにタッチ・popoverで選べないケースを分離する。同じ医院・日付・予約区分・権限・staff所属を固定し、動くPCと差分を採取する。
- **完了 / PO・停止:** 対象PC/iPadの版を明記し、候補表示→選択→保存→再読込と無資格者の除外を確認。POに全staff開放を提案して回避せず、capability判定を保持する。対象端末未入手なら実機分はBLOCKED、再現なしでブラウザの古さを原因にしない。

### SLACK-OWNER-HEIGHT

- **source / 状態:** OWNER-HEIGHT。出典736–763。最大化済みでも飼主検索結果をスクロールできない。Win8/Chrome、1366×625、15.6インチ。マスタ候補の高さとは別の再現対象。
- **現行根拠・最初の作業:** [飼主/ペット一覧](frontend/src/features/owners/components/OwnersListTable.tsx) と [飼主検索modal](frontend/src/components/shared/OwnerSearchModal/OwnerSearchModal.tsx) の呼出元・スクロール責務を照合し、どちらのsurfaceかを既存条件から絞る。候補多数/0件、検索欄と末尾行、キーボード選択をケース化し、[CHART-FIT](#uat-r2-chart-fit) の実機viewport/zoom採取と互換性ゲートを共用する。
- **完了 / PO・停止:** 検索・選択・閉じるへ到達できる前後証拠と、1366×625基準/実機双方の結果。物理寸法をCSS viewportと同一視せず、単に再度最大化を案内して閉じない。正確な画面/旧Chromeが未確認なら該当実機受入を残す。

### SLACK-LATENCY

- **source / 状態:** LATENCY。出典883–896の「数量入力など反映に時間がかかりすぎる」。2回Enterの受入とは独立の計測課題。
- **現行根拠・最初の作業:** [TreatmentQuantityCell](frontend/src/features/medical-records/components/TreatmentsTab/TreatmentQuantityCell.tsx) → [治療hook](frontend/src/features/medical-records/hooks/use-treatments-tab.ts) → [治療API](frontend/src/features/medical-records/api/treatments.ts) の入力表示、commit、通信、再取得を分けて採時する。行数・端末・回線・revision・IME条件を固定し、どこで待つかを特定する。
- **完了 / PO・停止:** 同じケースの前後計測と保存値一致、既存Enter/Blur/IME挙動の維持。許容時間は現場の目的と実測から合意し、架空の性能値/原因を置かない。対象環境不足は計測BLOCKED。根拠がない一括debounce/楽観更新を先行実装しない。

### SLACK-COMPLAINT

- **source / 状態:** COMPLAINT。出典929–937の「主訴区分は空欄で入力」。空欄許可は依頼済みであり、可否を再質問せず保存経路を調査する。
- **現行根拠・最初の作業:** [InterviewChiefComplaint](frontend/src/features/medical-records/components/InterviewChiefComplaint.tsx#L78-L103) は `number | null` と空文字→nullを扱い、[作成request](backend/internal/medicalrecord/medical_record_request.go#L136) も任意ID。新規の未選択と、既存選択を解除した場合を分けてUI→request→保存→再読込を追う。クリア操作の有無と「省略」とnullの違いを確認する。
- **完了 / PO・停止:** 区分未設定で主訴本文を失わず保存でき、既存区分の意図的解除も受入条件に対応する。失敗した画面/経路だけ最小修正し、該当しなければ操作案内へ。納品区分・受入者の確認は別途だが、空欄許可そのものを新しい仕様待ちに戻さない。

### SLACK-VITALS

- **source / 状態:** VITALS。出典929–937。「名前のすぐ近く」「クリックなしで表示」「時間不要」。表示/入力操作の削減と診療記録の時刻保持を分離する設計 READY。
- **現行根拠・最初の作業:** [VitalsTab](frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx) と [ヘッダー配置](frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx) で常時表示の候補を比較する。[Vitalモデル](backend/internal/model/vital.go#L22) / [作成request](backend/internal/medicalrecord/vital_request.go#L13) は `recorded_at` を保持し、時刻順を用いる。既存保存経路を再利用し、ヘッダーへ別の永続値を作らない。
- **完了 / PO・停止:** 常時表示する項目、最新値と今回測定の識別、空欄/未保存/複数測定の扱いを臨床POが確認し、全9タブ/狭い画面でも入力・保存できる受入条件。時刻入力/表示の省略案は記録時刻の根拠を保つ設計とし、「時間不要」を保存時刻削除・測定時刻捏造へ読み替えない。臨床上の時刻扱いが決まるまでその変更は停止。

### SLACK-MICROCHIP

- **source / 状態:** MICROCHIP。出典938–943。マイクロチップ番号を名前の近くへ表示したい。表示設計 READY。
- **現行根拠・最初の作業:** [Petの既存フィールド](backend/internal/model/pet.go#L55) → [変換](frontend/src/lib/transforms/pet.ts#L107) は存在する。[MedicalRecordStickyHeader](frontend/src/features/medical-records/components/MedicalRecordStickyHeader.tsx) と [PatientContextHeader](frontend/src/components/shared/PatientContextHeader/PatientContextHeader.tsx) のprops/表示へ渡る経路を確認し、既存値を一度だけ表示する最小案を作る。
- **完了 / PO・停止:** 番号有無・長い表示・患者切替・1366×625で正しい患者の値が欠けず読めること。対象ヘッダーと表示目的/権限をPOが確認し、二重保存や架空の番号を作らない。実装後はAPI再取得/切替時の古い値残留を受入で確認する。

### SLACK-DETAILS

- **source / 状態:** DETAILS。出典949–955。飼主/ペット名または隣から詳細を見たい。詳細閲覧導線の調査 READY、添付画面の内容は未確認。
- **現行根拠・最初の作業:** [ReadyPanels](frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx#L218) の `onOwnerClick` は `handleOpenOwnerSearch`（飼主差替え検索）であり、詳細閲覧ではない。[共通ヘッダー](frontend/src/components/shared/PatientContextHeader/PatientContextHeader.tsx#L117-L137) は飼主ボタン/ペット名span。[既存飼主フォーム](frontend/src/features/owners/routes/OwnerForm.tsx) / [ペット編集](frontend/src/features/owners/components/PetEditModal.tsx) の閲覧・編集権限を整理し、読み取り詳細を未保存のカルテ入力を失わず開く案を比較する。
- **完了 / PO・停止:** 飼主/ペットそれぞれで正しい詳細に到達し、閉じる/戻るでカルテ入力が保持され、患者の差替えを誤発火させないこと。POが閲覧だけか編集も必要か、表示項目と権限を確定。新しい重複画面や患者変更を前提にせず、既存画面を再利用できる範囲を先に調べる。

### SLACK-VACCINE-MULTI

- **source / 状態:** VACCINE-MULTI。出典968–975。「登録できない」と「初診/同日に2–3件入力」を別ケースとして調査 READY。
- **現行根拠・最初の作業:** [カルテ内フォーム](frontend/src/features/medical-records/hooks/use-medical-record-vaccination-form.ts) と [独立フォーム](frontend/src/features/vaccinations/hooks/use-vaccination-form.ts)、[作成API](frontend/src/features/vaccinations/api/create-vaccination.ts)、[接種service](backend/internal/medicalrecord/vaccination_service.go) を照合する。保存済み/新規カルテ、対象ペット、必須項目、権限、エラー表示、species候補を固定し、単件成功→同日別接種2–3件を順に保存→一覧再読込をまず再現する。
- **完了 / PO・停止:** 登録失敗の原因/修正receiptと、各接種の実施日・lot・次回予定・金額が混ざらず保存される証拠。一括入力が必要な操作数/失敗時の部分保存の扱いはPO裁定後に設計する。単件POSTの存在だけで同日複数不可やbatch API必須と決めない。[種別課題](#uat-q2-vaccine-species) の承認済み条件を保持し、接種間隔などの臨床判断を代行しない。

### SLACK-PLAN-MANUAL

- **source / 状態:** PLAN-MANUAL。出典987–994。治療プランと治療の違いの質問、プランへの手入力要望。既存手入力の調査/案内 READY。
- **現行根拠・最初の作業:** [MedicalRecordDiagnosisPlan](frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx#L78-L84) は治療APIを使い、[手入力行](frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx#L135-L147) が既存。[治療タブhook](frontend/src/features/medical-records/hooks/use-treatments-tab.ts) と比較して、画面名→対象データ→保存時点→未請求/会計の対応を説明できる表を作る。別の [treatment_plan request](backend/internal/medicalrecord/treatment_plan_request.go) の存在を、このUIが予定専用である根拠にはしない。
- **完了 / PO・停止:** 現行配信版で手入力の場所・保存・再読込・会計との関係を示し、既存機能で満たせば操作質問として閉じる。価格欠落は [MASTER](#uat-r2-master-path) へ統合する。真に別の「予定だけで未実施/非請求」の意味が必要と分かった場合だけ、実施への移行条件を臨床POが裁定するまでその仕様変更を停止する。予定を自動で実施・請求済みにしない。

### SLACK-COPY

- **source / 状態:** COPY。出典987–994の前回治療/主訴のコピー要望。転記削減の設計 READY、複写の保存意味はPO確認が必要。
- **現行根拠・最初の作業:** [問診履歴](frontend/src/features/medical-records/components/InterviewHistory.tsx#L76-L99) は過去詳細へのリンク。主訴の定型文挿入や接種履歴の複製は、前回治療/主訴コピーが実装済みの根拠ではない。[治療API](frontend/src/features/medical-records/api/treatments.ts) と主訴保存経路を調べ、同一医院/ペットの前回記録を参照して選択的に取り込む案を作る。
- **完了 / PO・停止:** コピー元の記録/日付が追え、追記/置換、現入力との競合、数量/単価/日付/薬剤の再確認、対象項目を臨床POが決めた受入条件。保存前previewと取消で元記録・現入力を失わず、旧ID/請求済み状態の流用や重複請求を防ぐ。過去の投薬判断を未確認で再実施する仕様や「前回」の曖昧な選択は停止。

### SLACK-CAMERA

- **source / 状態:** CAMERA。出典995–999。撮影してそのままカルテへ追加したい。対象端末と最小導線の調査 READY。
- **現行根拠・最初の作業:** [ImageGalleryFilter](frontend/src/features/medical-records/components/ImageGalleryFilter.tsx#L107-L108) は画像/PDF file inputで、明示的なcapture指定は確認できない。[カルテ画像](frontend/src/features/medical-records/components/MedicalRecordImage.tsx#L58-L66) の既存uploadへつなぐ経路を調べ、対象iPad/スマートフォンで現file pickerに撮影選択が出るかを先に実機確認する。PCカメラや独自カメラUIが必須とは推定しない。
- **完了 / PO・停止:** 対象端末で撮影→確認/取消→正しいカルテへ保存→再読込ができ、向き・許可拒否・容量/形式エラー・通信失敗を説明できること。POが必要端末/用途/保存権限を確定し、既存の医院境界・死亡/確定制約を保持。カメラ権限や安全なfixtureなしの実機試験は停止し、実患者写真を試験へ流用しない。

### SLACK-INTAKE

- **source / 状態:** INTAKE。出典1000–1063。受付/看護の窓口集約、1内容1件、不具合/要望/質問の分類、確認済み重複行は不要との回答。外部シートの現状は UNKNOWN。
- **現行根拠・最初の作業:** [外部Q&A運用規則](docs/ops/backlog-spreadsheet.md) と本書の45キー/50親対応を使い、各課題のsource・現行コード・再現/質問・分類根拠を点検する。9月16日の細分化要望は個別プランを保ち、この窓口課題だけにまとめない。
- **完了 / PO・停止:** 各topicが既存ID、新規調査、回答済み、重複、対象外のいずれかに対応し、曖昧な納品前/後の分類にはPOへの限定質問がある。外部照合時は承認済locator/対象行/更新範囲を確定するまで停止。黒塗り行の不要回答を無制限削除の許可にせず、既に消去済みとも断定しない。

<a id="area-po"></a>

## 2. PO判断待ち

対象欄、権限、臨床上の意味、採用範囲などの判断を先に確定する課題です。判断材料となる既存コードの照合・案の比較は進められますが、未確定の仕様を選んで実装には進みません。判断してほしい内容は各タスクの「完了 / PO・停止」に記載しています。

### SLACK-CROSS-CLINIC

- **source / 状態:** CROSS-CLINIC。出典643–697、725–735、832–847。山梨の他院患者を検索して受付したい。代表アカウントという返信があり、単なる同院検索の不具合と分ける。
- **現行根拠・最初の作業:** [飼主一覧loader](frontend/src/features/owners/loaders.ts#L141-L159) / [カルテ一覧hook](frontend/src/hooks/use-medical-records.ts#L60-L61) は `clinic_ids` を扱う。[認証middleware](backend/internal/middleware/auth.go) と予約/受付側の所属検証を追い、アカウント所属、選択医院、検索範囲、記録の医院、操作医院の対応表を作る。
- **完了 / PO・停止:** POが対象医院の明示リスト（八王子を含むかも明記）、閲覧/受付/編集/会計それぞれの権限と新規診療の帰属を裁定し、許可・拒否双方のケースと監査方法が確定。安全な調査は開始可だが、裁定前の医院境界変更は停止。`clinic_id` 検証の削除や全院無条件共有で解決しない。

### SLACK-BACKGROUND

- **source / 状態:** BACKGROUND。出典929–937の原文は「背景分に皮下点滴を追加」。対象画面・欄・臨床的用途が特定できておらず、PO確認待ち。
- **現行根拠・最初の作業:** [問診の定型文挿入](frontend/src/features/medical-records/components/InterviewChiefComplaint.tsx#L106-L128) と [診察/治療プラン](frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx)、治療マスタの既存候補を比較し、入力欄を特定するための画面対応図を作る。「背景分」を定型文、既往歴、治療行のいずれとも推定しない。
- **完了 / PO・停止:** 依頼者/臨床責任者が対象欄と目的、追加内容、マスタか自由文か、実施/請求との関係を確定した受入条件。特定前は内容追加を停止し、用量・方法・臨床文面を作らない。既存項目で目的を満たす場合は重複登録せず案内する。

### SLACK-DANGER

- **source / 状態:** DANGER。出典944–948。危険度に応じ赤/黄で分かる表示を求める。既存高危険表示の保持を前提に追加箇所を設計する。
- **現行根拠・最初の作業:** [Pet danger level/reason](backend/internal/model/pet.go#L59-L60)、[変換](frontend/src/lib/transforms/pet.ts#L118-L119)、[一覧の「高」危険バッジと理由](frontend/src/features/owners/components/OwnersListTable.tsx#L203-L222) は既存。[一覧テスト](frontend/src/features/owners/components/OwnersListTable.report.test.tsx#L179-L207) は高だけの表示を規定する。カルテヘッダーへ値が届くかを確認し、既存一覧の赤系表示を未実装と扱わない。
- **完了 / PO・停止:** 高/中/低・不明、赤/黄の対応、表示場所、理由の閲覧を臨床POと決め、色以外の文字/アイコンでも区別できる。既存「高」の理由必須・表示を保持し、患者切替と狭い画面で確認する。黄色の意味を推測して中/低へ自動割当しない。未裁定なら色意味の変更は停止。

### SLACK-STORY

- **source / 状態:** STORY。出典949–955の「名前の由来と出逢いのストーリー欄」。DETAILS とは別の記録要望、PO判断待ち。
- **現行根拠・最初の作業:** [Pet](backend/internal/model/pet.go#L58-L66) の取得区分/備考と [ペット編集項目](frontend/src/features/owners/components/PetEditModalFields.tsx) の利用を調べ、二重入力なしに目的を満たせるかを検討する。取得区分をストーリー全文の保存先とみなさない。
- **完了 / PO・停止:** 記録者/読む人/業務目的、既存備考で足りるか、必要な表示先・上限・編集権限・個人情報の範囲が確定した採否資料。削減工程がない欄追加は目的を再検討し、独立フィールド/DB追加は採用まで停止。診療記録や会計への自動転記はしない。

### SLACK-VACCINE-PRINT

- **source / 状態:** VACCINE-PRINT。出典956–960の紙のワクチン証明書の質問。専用証明書の対応範囲確認が必要。
- **現行根拠・最初の作業:** [カルテprint hook](frontend/src/features/medical-records/hooks/use-medical-record-form-modals.ts#L30-L35) と [診療カルテ印刷](frontend/src/features/medical-records/components/MedicalRecordPrintView.tsx#L66-L67) は一般カルテ用。調査範囲では専用証明書route/rendererを確認できていないため「既に印刷可」と回答しない。[接種フォーム仕様](docs/spec/screens/15-vaccinations-form.md) と保存項目を、承認済みの証明書見本/必要項目へ対応づける。
- **完了 / PO・停止:** 証明書の用途・種類・記載内容・発行者/日付・用紙/レイアウトを臨床POが承認し、正しい接種記録だけからpreview→PDF→実印刷できる受入条件。公的文書要件や未収録の見本を推測せず、専用仕様が決まるまで実装は停止。一般カルテ印刷と証明書を同一視しない。

### SLACK-DECEASED

- **source / 状態:** DECEASED。出典982–986。死亡後も飼主との連絡を記録したい。死亡日の補完とは別の用途、記録範囲のPO判断待ち。
- **現行根拠・最初の作業:** [死亡write guard](backend/internal/sharedkernel/pet_not_deceased.go)、[カルテの死亡ガード回帰](backend/internal/medicalrecord/medical_record_deceased_pet_test.go)、既存飼主/ペット備考の書込み経路を比較する。非診療の連絡メモと新規診療/処置/請求を分け、適切な既存記録先または限定された連絡記録案を提示する。
- **完了 / PO・停止:** 記録対象、閲覧/編集者、時刻/記録者/監査、過去カルテとの関係をPOが確定し、許可した連絡だけが保存でき、新規臨床writeや請求は禁止のままという受入条件。用途未確定ならガード変更を停止。死亡フラグ解除・死亡日捏造で回避せず、[日付根拠のある限定訂正](#po-pet-deceased-data-backfill) を再開/統合しない。

<a id="area-evidence"></a>
<a id="データ移行の残件"></a>

## 3. 資料・環境・実機確認待ち

資料・対象環境・実行承認・受入者・実機の証拠が揃ってから、本体の検証やデータ操作を進める課題です。必要な資料の特定、既存票の更新、試験条件の整理は各本文に従って先行できます。設計済みの集計票を作り直したり、環境待ちを実装不備と断定したりしません。会計通し確認は最優先のまま、必要条件を揃えます。[全種類・全期間の処置移行契約設計](#uat-q2-treatments-import) はエリア1で着手できます。

既存課題の担当・成果物・開始条件:

| ID | 状態・次の一手 | 担当・ローカル成果物 | 本体の開始条件 |
|---|---|---|---|
| [UAT-Q3-GENDER-MAP](#uat-q3-gender-map) | 未完了（bundle・運用） / old_db main 統合を確認済み。現行bundle・PostgreSQL検証証拠を照合し、承認後の STG 訂正へ | producer/開発: 統合済みrevisionを使用するbundle/検証receiptの照合 | 再統合は不要。bundle/STG は [運用計画](todo-operations.md#uat-q3-gender-map) |
| [UAT-Q2-VACCINE-SPECIES](#uat-q2-vaccine-species) | 調査待ち / 承認された STG 集計で、マスタ種欠損・参照取り違え・旧記録そのものを分類 | 集計設計済み。[既存票](docs/work/todo-campaign-20260918/UAT-Q2-VACCINE-SPECIES.md) を使い、対象医院/期間/読取条件を実行担当が埋める | 条件未充足なら待機。同じ集計設計を作り直さない |
| [UAT-Q4-UNPAID-TRIAGE](#uat-q4-unpaid-triage) | 調査待ち / 承認された STG 集計で、旧未精算と支払未紐付けを切り分け | 集計設計済み。[既存票](docs/work/todo-campaign-20260918/UAT-Q4-UNPAID-TRIAGE.md) を使い、対象医院/期間/読取条件を実行担当が埋める | 条件未充足なら待機。回収済み事実は旧記録との照合が必要 |
| [PO-PET-DECEASED-DATA-BACKFILL](#po-pet-deceased-data-backfill) | 限定訂正待ち / 日付根拠がある行だけ対象。対象・監査・復旧・実行承認を確定 | 設計済み。[既存票](docs/work/todo-campaign-20260918/PO-PET-DECEASED-DATA-BACKFILL.md) の対象/除外条件へ実行担当が根拠を対応づける | 日付根拠のある対象集合・対象環境・監査・実行承認。根拠なしは除外 |

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

### PO-PET-DECEASED-DATA-BACKFILL

[修復計画](bug.md#plan-po-pet-deceased-data-backfill) は **死亡日の根拠がある行だけ訂正**する方針。根拠がない行は補完対象外のまま残し、対象・監査・復旧方法・実行承認を確定する。死亡 write ガードの実装は再開しない。訂正日は推測で埋めず、承認されたデータ操作と照合が完了するまで残件として維持する。

### SLACK-BILLING-UAT

- **source / 状態:** BILLING-UAT。出典101–126、1000–1063。9月16日の最優先は、新カルテ単独の会計→精算書/領収書PDF→物理プリンタ→締め。八王子は新規カルテで実施する方針。
- **現行根拠・最初の作業:** [会計詳細](frontend/src/features/accounting/routes/AccountingDetail.tsx#L97-L106)、[印刷preview](frontend/src/features/accounting/components/AccountingDetailPanels.tsx#L315)、[明細兼領収書](frontend/src/features/accounting/components/AccountingDocument.tsx#L131)、[会計確定](backend/internal/billing/accounting_handler.go#L132)、[締め画面](frontend/src/features/cash-register/routes/CashRegisterClosePage.tsx#L73) から通し票を作る。顧客と品目・金額を確認→確定/支払→ブラウザ印刷でPDF保存/実印刷→締めpreview/実査/保存を分け、対象医院・build・権限・合成fixture・後処理を固定する。
- **完了 / PO・停止:** 各段の金額/品目/支払/未納/締めの一致と、実プリンタの用紙・縮尺・欠落なしのreceipt。コード上の `window.print()` は専用PDF生成やプリンタ成功の証拠ではない。過去Slackの窓口を参照し現受入担当/日程を確認する。実請求・締め操作は運用承認後、止まったroute/actionを個別不具合にし、Smaregi待ちにはしない。

### SLACK-CLINICAL-UAT

- **source / 状態:** CLINICAL-UAT。出典704–724、1000–1063。各拠点でカルテ入力・処置・検査・予約の診療/看護一連を確認する。
- **現行根拠・最初の作業:** [全9タブ定義](frontend/src/features/medical-records/routes/medical-record-form-model.ts)、[臨床E2E対象](frontend/scripts/run-e2e.sh)、[臨床受入キュー](todo-verification.md#todo-v-clinical-e2e--qa-full-clinical-e2e) を医院×職種×旧業務→新導線へ対応づける。自動テストの対象と実際の現場手順を区別し、[UAT-SCHEDULE](#slack-uat-schedule) の未実施者/導線を埋める。
- **完了 / PO・停止:** 同一患者の入力→保存→処置/検査→予約→再読込、医院/権限分離と中断時の復帰を承認fixtureで示す。未達を一件一原因で記録し、受付/看護窓口へ集約するための下書きを作る。実送信は別承認。臨床手順・受入者が不明ならそのケースを止め、自動テスト成功を全院受入にしない。

### SLACK-ACCESS

- **source / 状態:** ACCESS。出典32–100、127–173、302–402、483–505。全院のログイン、デモ権限、スタッフ別アカウントは運用 readiness。9月9日のログイン成功報告は一部の例であり、全員の確認済みではない。
- **現行根拠・最初の作業:** [demo環境判定](backend/internal/seedlogin/env.go#L19) と [スタッフ発行条件](docs/ops/deploy/STAFF_ACCOUNT_PROVISIONING.md) を使い、[P5](todo-operations.md#p5--staff-provision) / H3-9 の入力不足を医院・役割・デモ/正式アカウント別に整理する。7月の名簿・転送原文は未収録のため、既知の氏名から名簿を作らず権限者に現行参照を取得してもらう。
- **完了 / PO・停止:** 対象clinic/staff/権限/実行者/環境/発行方式の非機密対応表と、承認後のログイン・許可操作/拒否操作のreceipt。正式アカウントのメール・権限方針は PO、認証情報の変更は別承認。remote方式未確定や名簿不足ならその実行を止め、デモ権限を本番へ流用しない。

### SLACK-UAT-SCHEDULE

- **source / 状態:** UAT-SCHEDULE。出典32–76、127–173、295–301、483–505、600–642、698–724。医院・交代勤務者を網羅するテスト準備。
- **現行根拠・最初の作業:** [UAT環境条件](docs/ops/testing/UAT-ENV-SETUP.md)、[既存検証キュー](todo-verification.md#直近-uat-の残作業) に、医院×受付/獣医師/看護×対象導線×配信revision×実施/未実施を対応づける。過去の候補日・会議調整から新しい期限を作らず、現在のテスト可能枠と結果保管先を運用担当が採取する。
- **完了 / PO・停止:** 未実施と失敗の理由、再確認者・連絡窓口の権限付き参照、次の実施枠が明確。日程/受入者は医院側、最優先は [会計通し](#slack-billing-uat)。ログイン/fixtureが不足する実機実行は停止し、机上チェックを医院受入に数えない。

### SLACK-HAC-IMPORT

- **source / 状態:** HAC-IMPORT。出典77–100、127–159、295–423。9月11日にBAKから抽出したCSVの送付報告があるが、現在の完全性・受領・投入結果は UNKNOWN。
- **現行根拠・最初の作業:** [CSV import契約](docs/ops/deploy/CLINIC_CSV_IMPORT.md)、[H0-2 / HAC-CSV-1](todo-operations.md#h0-2--hac-csv-1)、[八王子レーン](todo-operations.md#ae-stg-uat-lane3-hac) の同じ残件へ接続する。producer/consumer revision・医院・対象表・行数/hash・欠損・正式/rehearsal用途を受領receiptと照合し、送付報告と検証済みbundleを区別する。
- **完了 / PO・停止:** 原本→変換→受入→履歴表示の件数/参照/金額が追え、未受領/保留表も報告済み。入力不足・契約不一致なら投入停止。対象・読取/投入承認・rollbackは運用側で確定する。過去履歴が揃う前でも、合成の新規カルテによる会計UATは別に準備できる。

### SLACK-LAB

- **source / 状態:** LAB。出典212–294、976–981。各院の機種差、敷島のみの「コアグ」、9月16日の検査受信可否/時期の質問。実機受信状態は UNKNOWN。
- **現行根拠・最初の作業:** [受信service](backend/internal/medicalrecord/lab_device_receive_service.go#L54)、[受信画面](frontend/src/features/lab-device/routes/LabDeviceBoard.tsx#L87)、[接続条件](docs/ops/deploy/LAB_DEVICE_CONNECTIVITY.md)、[実機UAT票](docs/ops/testing/scenarios/LAB_DEVICE_CLIENT_UAT.md) を医院×機器×送信形式×対応状態に対応づける。城東の既存対応、decoder-only、通常運用未承認、未知機器を分ける。猫/八王子の型番やコアグのプロトコルは写真を見ていないため未確定。
- **完了 / PO・停止:** 承認された機器ごとに受信→患者/検査紐付け→数値/単位/基準値→再送重複/失敗表示を照合したreceipt。不足機器仕様・接続/認証情報の供給承認・医院受入がない場合は該当機器をBLOCKEDとする。実装済み受信コードだけで連携済みや対応日を回答しない。手動尿検査は次項へ分ける。

### SLACK-EXAM-HISTORY

- **source / 状態:** EXAM-HISTORY。出典506–529。DrOne単体データは移行対象外、旧カルテ/BAKの検査履歴は対象という回答・謝辞あり。後者の移行/表示は受入待ち。
- **現行根拠・最初の作業:** [カルテ検査履歴](frontend/src/features/medical-records/components/MedicalRecordExamination.tsx#L40-L60) と [取得API](frontend/src/features/medical-records/api/get-record-examinations.ts)、[CSV契約](docs/ops/deploy/CLINIC_CSV_IMPORT.md) から対象表・旧ID・医院/ペット・検査項目・日付・単位を対応づける。履歴一覧の上限/日付条件と取込欠損を分け、[HAC](#slack-hac-import) の受入receiptに接続する。
- **完了 / PO・停止:** 対象原本の件数・値・表示先、除外したDrOne由来の範囲を根拠付きで説明できる。承認原本がなければ実データ検証は停止。除外回答を勝手に再開せず、範囲変更は新たな PO 裁定とする。

### SLACK-RESERVATION-REFERENCE

- **source / 状態:** RESERVATION-REFERENCE。出典643–697。スタッフ/シフト作成後の「参照先が存在しません」は9月13日に対応報告あり。報告時の画面・現行配信版の同一性は未確認。
- **現行根拠・最初の作業:** [予約handler](backend/internal/reservation/reservation_handler.go#L31) と [予約作成service](backend/internal/reservation/reservation_service.go#L317) のclinic/staff/capability/owner/pet参照を、合成の同医院ケースに対応づける。新規staff作成後の所属・対応可能区分も確認し、失敗request/応答を非機密で収集する。
- **完了 / PO・停止:** 対象buildの同じ操作で保存・再読込でき、無効/他院参照は拒否されるreceipt。再現しなければ「対応報告後の確認済み」として閉じ、原因を捏造して再実装しない。環境/操作が不明なら受入待ち。選択UI自体の障害は次項。

### SLACK-BACKLOG61

- **source / 状態:** BACKLOG61。出典1064–1070、No.61用「内容締めの集計表 表示.pdf」の添付のみ。原票・期待内容 UNKNOWN、限定調査待ち。
- **現行根拠・最初の作業:** [締め画面](frontend/src/features/cash-register/routes/CashRegisterClosePage.tsx) / [締めservice](backend/internal/billing/cash_register_service.go) は比較先の候補であり、この添付の正体や不具合箇所とは未確定。権限者がNo.61本文とPDFを安全に取得し、対象画面/集計軸/期待額/実際/再現条件を機密除去して対応づける。
- **完了 / PO・停止:** 本文/添付の出所と対象コードが一致し、期待値を含む再現・受入条件が定義されること。資料なしで帳票欄・税/端数・集計式を作らず、実装はBLOCKED。外部シート本文やPDF全体を本書へ複製しない。

<a id="area-deferred"></a>

## 4. 納品後対応・別スコープ待ち

OCR・Smaregiは納品後の採用判断まで延期します。`TASK-444-ADDENDUM-CODEGEN` は納品時期とは無関係の別スコープ採用・user-run codegen待ちです。いずれも記載した再開条件を満たすまでは、実装・接続・自動実行を開始しません。

既存課題の担当・成果物・開始条件:

| ID | 状態・次の一手 | 担当・ローカル成果物 | 本体の開始条件 |
|---|---|---|---|
| [TASK-444-ADDENDUM-CODEGEN](#task-444-addendum-codegen) | DEFERRED / 型生成経路の別スコープと user-run codegen が認められたときに再開 | DEFERRED 維持。必要性が提起された場合に DTO/tygo/利用箇所の差分を提示 | 別スコープの採用と user-run codegen の実行者・検証範囲 |

### SLACK-OCR

- **source / 状態:** OCR。出典424–460。納品後に必要性を相談する DEFERRED。現時点の合意は画像/PDFをアップロードして閲覧する運用。
- **現行根拠・再開時の作業:** [カルテ画像](frontend/src/features/medical-records/components/MedicalRecordImage.tsx#L45-L66) の保存・閲覧を現行受入に使う。納品後に採用が決まった場合だけ、対象帳票/件数、手入力削減の目的、認識結果の人による確認、値・単位・患者誤紐付けの防止を設計し、providerの当時の仕様・料金・保持条件を調査する。
- **完了 / PO・停止:** 対象データ・精度評価方法・費用上限・個人情報の扱い・停止/失敗通知が承認され、手動確認で安全性を測れる採否資料。Slackの古い参考価格を現在の費用とせず、無承認の外部送信/有料API利用/自動カルテ確定へ進まない。納品・採用の契機がなければ延期を維持する。

### SLACK-SMAREGI

- **source / 状態:** SMAREGI。出典101–126、483–505、530–559、600–642。9月16日1000–1063の「納品後」が先の急ぎ依頼を更新する。DEFERRED。
- **現行根拠・再開時の作業:** [会計確定](backend/internal/billing/accounting_complete.go) は既存。backend/frontend/docsの名称検索では専用連携実装を確認できない。採用後に公式API契約を調べ、確定会計→取り置き→受付選択→品目/金額カート、返金時の顧客単位履歴、trial/本番の設定差分を一つずつ対応づける。二重送信、再試行、取消/返金、送信失敗の確認方法を設計する。
- **完了 / PO・停止:** API可否・項目対応・顧客情報の最小範囲・ID紐付け・試験/本番切替・重複防止/復旧の契約が揃う。1円単位切捨ての具体則、計算段階、税・割引との順序は受付/PO未回答。これらと有料利用/外部送信承認がない実装・接続は停止。現行の [会計UAT](#slack-billing-uat) は独立して進める。

### TASK-444-ADDENDUM-CODEGEN

カルテ追記 response 型は現在 tygo 対象外。先行調査では [response DTO](backend/internal/medicalrecord/medical_record_addendum_response.go)、[tygo 設定](backend/tygo.yaml)、[生成済み response 型](frontend/src/types/generated/medicalrecord-responses.ts) を比較し、対象型と利用箇所を列挙する。生成経路を変更する必要性と別スコープを確定した場合だけ実装する。user-run `make codegen` の後、生成差分が対象契約に限定されることと該当 API の型を確認し、再生成で差分が増えないことを完了条件とする。エージェントの自動 codegen や生成物の手編集で代用しない。背景は [裁定記録](docs/work/development-task-decisions.md#task-444)。

<a id="area-resolved"></a>

## 5. 回答済み・実装済み（受入確認は別管理）

回答済み操作2トピックと、コード対応・導線がある5トピックを分けて保持する。ここへの配置は実機・医院受入の完了を意味しない。未完了の受入は [検証 TODO](todo-verification.md#uat-followup) で管理する。

<a id="slack-answered"></a>

### 回答済み操作と実装済み部分の扱い

- **SHIFT:** 出典560–599。スタッフ有効/医師、対象院・予約日の勤務シフトを確認する案内と謝辞あり。[予約枠判定](backend/internal/reservation/reservation_repository_slots.go#L144) と [予約からカルテの契約](docs/spec/reservation-to-record-flow.md) を現行手順の根拠とする。担当明示時は別の競合/資格検証があり、任意医師指定をシフト設定の代用にしない。新しい不具合として再実装せず、実際の保存成功は未確認として [予約再確認](#slack-reservation-reference) に接続する。
- **RESERVATION-EDIT:** 出典907–923。予約→対象日→カード→権限付き編集→保存の回答・謝辞あり。[ReservationManagement](frontend/src/features/reservations/routes/ReservationManagement.tsx#L217-L240) に月から週への移動・詳細・編集導線がある。操作質問として保持し、新規実装は起こさない。後日再現した権限/保存問題だけを別途切り分ける。
- **SEARCH-AND / INSURANCE / ENTER / MASTER-HEIGHT / HISTORY:** [既存UATキュー](todo-verification.md#uat-followup) の同じIDを使用する。検索の空白区切りAND/電話・番号維持、保険新規50/70・旧90/100保持、2回Enter/IME、一覧到達性、同一ペットの過去カルテ導線はそれぞれ受入範囲を保つ。ローカル実装/過去テストの存在は今回の医院受入 PASS ではない。処置履歴が未移行なら [全期間移行](#uat-q2-treatments-import) へ接続し、導線を再実装して解決済みにしない。

## 共通の前提・確定事項

要件出所は2026-09-19の本会話の依頼者回答（[確定票](#decision-inputs)）。今回は文書更新のみで、実装・検証・移行は未実施。個人名を伴う要件責任者・受入者の参照は実行担当が製品仕様変更前に記録する。これを、範囲の再質問やローカル調査の停止理由にはしない。共有STG/PROD操作、外部送信、実データ移行の承認は別に必要。

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

## 出典とトピックの対応

### トピック別の処理先

45キーを以下へ対応づける。30件はSlack個別プラン、8キーは既存9課題（EXISTING-OTHER は2件）、5キーは実装済み部分の受入先、2キーは回答済み操作案内。課題本文は上の5エリアへ配置している。本文の保存やローカル検証を、報告されたソフトウェア不具合の解消とは扱わない。

| Topic key | 処理区分・担当課題 |
|---|---|
| ACCESS | 運用準備 → [SLACK-ACCESS](#slack-access) |
| UAT-SCHEDULE | 受入準備 → [SLACK-UAT-SCHEDULE](#slack-uat-schedule) |
| HAC-IMPORT | 既存運用へ接続 → [SLACK-HAC-IMPORT](#slack-hac-import) |
| LAB | 調査・実機受入待ち → [SLACK-LAB](#slack-lab) |
| MANUAL-URINE | 手入力経路の受入設計 → [SLACK-MANUAL-URINE](#slack-manual-urine) |
| OCR | 納品後 DEFERRED → [SLACK-OCR](#slack-ocr) |
| EXAM-HISTORY | 移行元の区別・受入 → [SLACK-EXAM-HISTORY](#slack-exam-history) |
| SMAREGI | 納品後 DEFERRED → [SLACK-SMAREGI](#slack-smaregi) |
| SHIFT | 回答済み → [操作案内の保持](#slack-answered) |
| RESERVATION-REFERENCE | 対応報告あり・再確認待ち → [SLACK-RESERVATION-REFERENCE](#slack-reservation-reference) |
| STAFF-SELECT | 再現調査 → [SLACK-STAFF-SELECT](#slack-staff-select) |
| CROSS-CLINIC | 権限境界の設計・PO → [SLACK-CROSS-CLINIC](#slack-cross-clinic) |
| OWNER-HEIGHT | 再現調査 → [SLACK-OWNER-HEIGHT](#slack-owner-height) |
| SEARCH-AND | コード対応済み・UAT待ち → [UAT-Q1-SEARCH-AND](todo-verification.md#uat-q1-search-and) |
| HISTORY | 導線コードあり・UAT待ち → [UAT-Q2-HISTORY-NAV](todo-verification.md#uat-q2-history-nav) |
| SPECIES | 既存調査 → [UAT-Q2-VACCINE-SPECIES](#uat-q2-vaccine-species) |
| GENDER | old_db 統合済み・bundle/実データ受入待ち → [UAT-Q3-GENDER-MAP](#uat-q3-gender-map) |
| UNPAID | 原因未確定・既存調査 → [UAT-Q4-UNPAID-TRIAGE](#uat-q4-unpaid-triage) |
| INSURANCE | 新規50/70・旧90/100保持の受入 → [UAT-Q4-INSURANCE-RATES](todo-verification.md#uat-q4-insurance-rates) |
| MASTER | 9月16日追加証拠を統合 → [UAT-R2-MASTER-PATH](#uat-r2-master-path) |
| CONCURRENCY | 9月16日追加事故を統合 → [UAT-R2-EXCLUSIVE-LOCK](#uat-r2-exclusive-lock) |
| LATENCY | Enter操作と分離して計測 → [SLACK-LATENCY](#slack-latency) |
| ENTER | コード対応済み・実機IME待ち → [UAT-R2-TREATMENT-COMMIT](todo-verification.md#uat-r2-treatment-commit) |
| MASTER-HEIGHT | コード対応済み・実画面受入待ち → [UAT-R2-MASTER-LIST-HEIGHT](todo-verification.md#uat-r2-master-list-height) |
| CHART-FIT | 既存全9タブ計画 → [UAT-R2-CHART-FIT](#uat-r2-chart-fit) |
| RESERVATION-EDIT | 回答・謝辞あり → [操作案内の保持](#slack-answered) |
| COMPLAINT | 任意入力の再現調査 → [SLACK-COMPLAINT](#slack-complaint) |
| BACKGROUND | 対象欄の特定・PO → [SLACK-BACKGROUND](#slack-background) |
| VITALS | 表示位置と記録時刻を分離 → [SLACK-VITALS](#slack-vitals) |
| MICROCHIP | 既存値の表示設計 → [SLACK-MICROCHIP](#slack-microchip) |
| DANGER | 既存高危険表示の保持・PO → [SLACK-DANGER](#slack-danger) |
| DETAILS | 詳細閲覧導線の調査 → [SLACK-DETAILS](#slack-details) |
| STORY | 記録目的/保存先のPO → [SLACK-STORY](#slack-story) |
| VACCINE-PRINT | 専用証明書の仕様確認 → [SLACK-VACCINE-PRINT](#slack-vaccine-print) |
| VACCINE-MULTI | 登録失敗と複数入力を分離 → [SLACK-VACCINE-MULTI](#slack-vaccine-multi) |
| DECEASED | 死亡後連絡記録のPO → [SLACK-DECEASED](#slack-deceased) |
| PLAN-MANUAL | 画面/API対応と案内 → [SLACK-PLAN-MANUAL](#slack-plan-manual) |
| COPY | 前回参照と転記の安全設計 → [SLACK-COPY](#slack-copy) |
| CAMERA | 撮影から保存の経路調査 → [SLACK-CAMERA](#slack-camera) |
| BILLING-UAT | 最優先の実機受入 → [SLACK-BILLING-UAT](#slack-billing-uat) |
| CLINICAL-UAT | 各院の診療/看護受入 → [SLACK-CLINICAL-UAT](#slack-clinical-uat) |
| INTAKE | 個別課題の分類と照合 → [SLACK-INTAKE](#slack-intake) |
| BACKLOG61 | 原票・PDF内容待ち → [SLACK-BACKLOG61](#slack-backlog61) |
| TREATMENTS-IMPORT | 既存の全種類/全期間回答保持 → [UAT-Q2-TREATMENTS-IMPORT](#uat-q2-treatments-import) |
| EXISTING-OTHER | 保存 → [死亡日限定訂正](#po-pet-deceased-data-backfill)、[追記型生成](#task-444-addendum-codegen) |

<a id="slack-source-map"></a>

### 全50親投稿・返信の出典対応表

行範囲は返信を含む。応答・謝辞・日程調整は各行の履歴へ含め、返信中の別要件をトピックへ分割した。親として再掲された2件は同一イベントであり、重複の実装課題を作らない。50親ブロックの内訳は通常対応44、重複再掲2、対象外4（参加通知3、表紙PDFの提出・謝辞1）。対象外は製品課題を追加しない判断であり、助成金受理の確認済みを意味しない。

| 出典行 | 親 timestamp | 処理先・理由 |
|---|---|---|
| 32–41 | `1787984405.983319` | ACCESS / UAT-SCHEDULE |
| 42–57 | `1787988264.830289` | UAT-SCHEDULE。候補日は履歴 |
| 58–76 | `1787992807.018379` | ACCESS / UAT-SCHEDULE |
| 77–100 | `1788261491.711849` | ACCESS / HAC-IMPORT。伏字の秘密は入力に使わない |
| 101–126 | `1788406963.496419` | SMAREGI / BILLING-UAT。購入内容を顧客と確認する要件を保持 |
| 127–159 | `1788422744.682529` | ACCESS / UAT-SCHEDULE / HAC-IMPORT。旧DB共通という返信は新システムの医院分離解除の根拠にしない |
| 160–173 | `1788427309.960679` | ACCESS / UAT-SCHEDULE |
| 174–178 | `1788529344.397139` | 対象外: 参加通知 |
| 179–183 | `1788662604.083949` | 対象外: 参加通知 |
| 184–211 | `1788671384.944389` | 対象外: 表紙付き提案PDFの送付・謝辞あり。助成金受理は未確認 |
| 212–294 | `1788686430.832129` | LAB / MANUAL-URINE。猫・八王子の写真、敷島コアグの追加返信も含む |
| 295–301 | `1788764745.291709` | HAC-IMPORT / UAT-SCHEDULE |
| 302–402 | `1788832046.595209` | ACCESS / HAC-IMPORT。9月11日のCSV送付は取込完了の証拠ではない |
| 403–423 | `1788853449.630089` | HAC-IMPORT。過去の「2日後」は新期限へ転用しない |
| 424–460 | `1788866843.926759` | OCR。納品後へ延期、当面の画像/PDF閲覧を保持 |
| 461–482 | `1788931012.789169` | 重複再掲: 302–402の返信。ACCESS / HAC-IMPORT に統合 |
| 483–505 | `1788934802.744769` | ACCESS / SMAREGI / UAT-SCHEDULE。7月転送原文は未収録 |
| 506–529 | `1788940302.976299` | EXAM-HISTORY。DrOne単体の移行除外は回答済み、旧カルテ検査履歴の受入は残る |
| 530–559 | `1789109051.075399` | SMAREGI。取り置き・顧客・端数・trial、本件は後日の延期に従う |
| 560–599 | `1789176984.729069` | SHIFT。操作回答・謝辞あり、予約成功の実証とは分離 |
| 600–642 | `1789189178.939309` | SMAREGI / UAT-SCHEDULE。会議日程調整は回答済み、新しい会議タスクなし |
| 643–697 | `1789191550.389859` | RESERVATION-REFERENCE / STAFF-SELECT / CROSS-CLINIC。Chromeでも選択不能の返信を保持 |
| 698–703 | `1789259845.886529` | UAT-SCHEDULE。テスト日の通知 |
| 704–724 | `1789276317.113699` | UAT-SCHEDULE / CLINICAL-UAT。敷島の交代勤務者の網羅 |
| 725–735 | `1789277858.989679` | 重複再掲: 643–697の返信。STAFF-SELECT / CROSS-CLINIC に統合 |
| 736–763 | `1789351591.207249` | OWNER-HEIGHT。最大化済み、Win8/Chrome・1366×625・15.6インチ |
| 764–783 | `1789388624.382679` | SEARCH-AND。実装報告と謝辞、UAT は既存検証先 |
| 784–794 | `1789389491.915289` | HISTORY。導線/移行済みデータの受入 |
| 795–805 | `1789389696.355619` | SPECIES。猫履歴に犬用、原因未確定 |
| 806–815 | `1789390572.658379` | GENDER。旧3/4コード修正後の実データ受入 |
| 816–831 | `1789391240.242759` | UNPAID / INSURANCE。移行不備との当時の推測は原因確定にしない |
| 832–847 | `1789437234.502929` | CROSS-CLINIC。代表アカウントという追加条件 |
| 848–882 | `1789453043.269699` | MASTER / CONCURRENCY。9月16日の価格・再診察・追加明細・照会混入も別ケース化 |
| 883–896 | `1789459005.162369` | LATENCY / ENTER / MASTER-HEIGHT。3要件を分離 |
| 897–906 | `1789459375.695359` | CHART-FIT |
| 907–923 | `1789463634.587649` | RESERVATION-EDIT。案内・謝辞あり |
| 924–928 | `1789538822.397429` | 対象外: 参加通知 |
| 929–937 | `1789539164.115569` | COMPLAINT / BACKGROUND / VITALS。3要件を分離 |
| 938–943 | `1789539363.704399` | MICROCHIP |
| 944–948 | `1789539576.256969` | DANGER |
| 949–955 | `1789539700.846829` | DETAILS / STORY。詳細導線と記録欄を分離 |
| 956–960 | `1789539775.666449` | VACCINE-PRINT |
| 961–967 | `1789540156.256659` | MASTER。マスタなし/0円/手入力案内 |
| 968–975 | `1789540169.467719` | VACCINE-MULTI。登録失敗と同日2–3件入力を分離 |
| 976–981 | `1789540531.163329` | LAB。受信可否と対応時期は実機条件確認後 |
| 982–986 | `1789540849.748249` | DECEASED。死亡日訂正とは異なる連絡記録 |
| 987–994 | `1789541345.488879` | PLAN-MANUAL / COPY。画面の違い・手入力と前回複写を分離 |
| 995–999 | `1789541513.692419` | CAMERA |
| 1000–1063 | `1789550614.370909` | BILLING-UAT / CLINICAL-UAT / INTAKE / SMAREGI。後者延期を優先、黒塗り重複行の削除回答は外部削除実施証拠ではない |
| 1064–1070 | `1789633031.590089` | BACKLOG61。PDF内容は UNKNOWN |

## 更新規則

ID ごとに状態・次の一手・完了条件・根拠を持つ。主な開始条件が変わったら本文を該当エリアへ移し、件数・トピック対応を更新する。同じIDの本文を複数エリアへ複製しない。完了項目は未対応エリアから外し、エリア5とSlack対応表に回答済み/完了/重複/対象外の根拠と移管先を残す。未完了の検証・運用は専用 TODO へ参照を残す。既存 claim は [AGENTS.md](AGENTS.md) に従って確認し、別担当の作業を取り込まない。
