# FE-refactor 第12期（FE12）— 84リーフルート 3軸監査・リファクタ計画

> 起票: 2026-07-24（要件責任者: 曽我）
> 業務目的: 全ページをデザイン正本・React実装ベストプラクティス・プロジェクト規約へ収束させるため、証跡付きで実行可能なリファクタ単位を確定する。
> **本ファイルは対応後削除する使い捨てトラッカー**。恒久規約は `DESIGN.md`、`docs/spec/design-system.md`、`docs/spec/ui-design-compliance.md`、`frontend/CLAUDE.md`。
> **本期は計画起票のみ**。コード・テスト・設定・恒久ドキュメントは変更しない。

## 決裁事項・判定境界

- 軸①の色は `docs/spec/design-system.md` が正本。brand `#0075DE`、業務ステータス色、臨床semantic色、nav canvas-softを維持し、DESIGN.mdのsticker paletteを色判定へ適用しない。
- 軸①のタイポグラフィ・形状・余白・エレベーション・寸法は `DESIGN.md` 字義が正本。
- 軸②は `vercel-react-best-practices` のカテゴリ名とrule名で根拠付ける。Vite SPAでNext.js/SSR/RSCを使わないため、Server-Side Performanceカテゴリは対象外。
- 軸③は `frontend/CLAUDE.md` / `frontend/src/features/CLAUDE.md` のMANDATORY規約違反、または `frontend/src`・`frontend/liff/src`・`frontend/line-reserve/src` 間/feature間の実重複だけを指摘根拠とする。
- `lib/design-tokens.ts` 805行、`ManualPage`の独自shell、C8 allowlist 14件、`@/shared-liff` alias構造は裁定済みのため再指摘しない。
- product-philosophyの順序は ①要件を疑う → ②削除 → ③簡素化 → ④サイクル短縮 → ⑤自動化。全taskで削除・統合・置換を先に判定し、追加だけの解決を禁止する。臨床安全は全原則に優先する。

## 判定記号

- `✅`: 適合。現在の証拠からtask不要。
- `FE12-XX`: 指摘あり。下記task行へ接続。
- `△`: 実装時にruntime計測または臨床レビューが必要。対応taskへ接続。
- `🔒`: 対象外または裁定済み例外。
- `保留セル`: RED段階だけで使用し、最終版では0件にする。

## Acceptance Checklist（実装前確定）

- [ ] **AC-01 84ルート×3軸の完全性** | Expected behavior: `ui-design-compliance.md` §2の84ルートが重複・欠落なく各3軸に `✅` / `FE12-XX` / `△ FE12-XX` / `🔒` の判定を持つ | Target surface: 本書「ページ×3軸 判定表」 | Verification method: §2と本書のmarker内からパス列を抽出して `sort | diff`、行数・重複・保留セルを検査 | PASS evidence: source=84、tracker=84、missing=0、extra=0、duplicate=0、保留セル=0。
- [ ] **AC-02 task行の実行可能性** | Expected behavior: 全指摘が一意な `FE12-XX` 行を持ち、priority・対象route・file:line証跡・根拠・修正方針・②削除判定・将来のscoped検証を含む | Target surface: 本書「FE12 task backlog」 | Verification method: marker内のMarkdown表を `awk -F'|'` で列数・ID連番・必須列非空・file:line・P0/P1/P2・`docker compose exec frontend npx vitest run <path>`等のscoped commandとして構造検査 | PASS evidence: malformed=0、duplicate ID=0、orphan ID=0。
- [ ] **AC-03 色正本の誤判定ゼロ** | Expected behavior: brand `#0075DE`・業務status色・臨床semantic色・nav canvas-softをDESIGN.md違反にしない | Target surface: 軸①判定とtask行 | Verification method: 決裁事項を目視照合し、task marker内だけを抽出して色値とDESIGN.md違反の同居を検索 | PASS evidence: forbidden false-positive=0、正本分割の明記あり。
- [ ] **AC-04 Vercel rule traceability** | Expected behavior: 軸②の全指摘がskillのカテゴリ名・rule名を引用し、Server-Side PerformanceをVite SPA対象外とする | Target surface: 軸②判定とtask行 | Verification method: 軸②taskを抽出し、カテゴリが `Eliminating Waterfalls` / `Bundle Size Optimization` / `Client-Side Data Fetching` / `Re-render Optimization` / `Rendering Performance` / `JavaScript Performance` / `Advanced Patterns` のいずれか、ruleがskillの公開名と一致することを照合 | PASS evidence: uncited axis② task=0、Server-Side N/A明記=1。
- [ ] **AC-05 軸③根拠の限定** | Expected behavior: 軸③指摘がMANDATORY規約または実重複に根拠付く | Target surface: 軸③判定とtask行 | Verification method: 軸③taskの根拠列を構造確認し、規約指摘は該当見出し、重複指摘は2箇所以上のfile:lineを持つことを照合 | PASS evidence: unsupported axis③ task=0、3ツリー走査記録あり。
- [ ] **AC-06 既知残債の計画化** | Expected behavior: C18 raw cell 22ファイル204件とC6a臨床安全UIがそれぞれtask化される | Target surface: 軸①task | Verification method: `rg -n 'C18.*22ファイル.*204件|C6a.*臨床安全' FE-refactor.md` とtask参照を確認 | PASS evidence: C18 task=1以上、C6a task=1以上、双方に段階的方針とscoped検証あり。
- [ ] **AC-07 design-audit実測反映** | Expected behavior: `docker compose exec frontend pnpm design-audit` のexit status・違反件数を引用し、fail時はP0 task化 | Target surface: 「機械監査baseline」とtask backlog | Verification method: 許可されたコマンドを実行し、出力と本書引用を照合 | PASS evidence: command/exit/件数が一致。exit非0なら対応P0 IDが存在。
- [ ] **AC-08 本run allowlist** | Expected behavior: 本runが書いたrepo pathは`FE-refactor.md`だけで、開始後も動く他sessionのWIPを本runへ誤帰属しない | Target surface: worktreeとexecutor write log | Verification method: `git status --porcelain | sort | diff "${TMPDIR:-/tmp}/fe12-git-baseline.txt" -` と全write tool targetを突合 | PASS evidence: 本run write targetは`FE-refactor.md`のみ。baseline後の他session backend driftは外部変動として列挙し、revert/stage/commitしていない。
- [ ] **AC-09 ②削除判定** | Expected behavior: 全taskに②削除判定があり、追加だけ・削除ゼロの修正方針がない | Target surface: task backlog | Verification method: task marker内を `awk -F'|'` で検査し削除判定列非空、修正方針に削除/統合/置換/既存コード追加0のいずれかを要求 | PASS evidence: missing deletion verdict=0、add-only task=0。
- [ ] **AC-10 verification-loop適用境界** | Expected behavior: docs-only・plan-onlyとして禁止されたfull lint/test/build/type-check/installを実行せず、許可されたdesign-audit・構造検査・diff review・secret/private-data目視だけを行う | Target surface: 実行ログと本書Run Summary | Verification method: 実行コマンド履歴と `git diff -- FE-refactor.md` の構造/機密情報レビュー | PASS evidence: prohibited command=0、コード変更=0、秘密・個人/臨床データ記載=0。coverage/lint/build/type-checkはCI所掌として未実行理由を記録。
- [ ] **AC-11 trackerの追跡可能性** | Expected behavior: 成果物が既存tracked fileであり、新規/ignored artifactではない | Target surface: `FE-refactor.md` | Verification method: `git ls-files --error-unmatch FE-refactor.md` と `git status --short -- FE-refactor.md` | PASS evidence: tracked path=`FE-refactor.md`、statusは本runの変更のみ。commit/stageは行わないためcached path listingは非該当。

## TDD状態

- RED: 正本84ルートに対し現行trackerの3軸判定行は0、FE12 task行は0。
- GREEN条件: AC-01〜AC-11がすべてPASSし、保留セル・orphan task・根拠なし指摘が0。

## 機械監査baseline・静的走査

- 実行: `docker compose exec frontend pnpm design-audit`
- 結果: **exit 0**、`design-system-audit: PASS — 違反 0 件`。C1/C3/C5/C6/C7〜C17/C18 override/C19は各0件、C18 rawは既知baseline **22ファイル204件（non-gating）**。
- 解釈: strict対象を再訴訟せず、軸①はC18 raw移行（FE12-01）と静的判定不能なC6a（FE12-02）へ集中する。色の違反は起票していない。
- 3ツリー走査: production TS/TSXの完全同一hash重複は0組。`React.FC` / `forwardRef` / production JSX `&&` / `__tests__` のMANDATORY違反は0件。意図的な `@/shared-liff` aliasは除外した。
- 軸②: Eliminating Waterfalls / Bundle Size Optimization / Client-Side Data Fetching / Re-render Optimization / Rendering Performance / JavaScript Performance / Advanced Patternsを適用。**Server-Side PerformanceはVite SPA（SSR/RSCなし）のため対象外**。

## ページ×3軸 判定表（GREEN）

<!-- FE12-ROUTE-TABLE-START -->
| エリア | ページ | パス | コンポーネント | 軸① デザイン | 軸② React BP | 軸③ 規約/重複 | 証跡・注記 |
|---|---|---|---|---|---|---|---|
| 認証/共通 | ログイン | /login | Login | ✅ | ✅ | ✅ | 静的strictとページ実装に追加指摘なし |
| 認証/共通 | パスワードを忘れた方 | /forgot-password | ForgotPasswordPage | ✅ | ✅ | ✅ | 同上 |
| 認証/共通 | パスワード再設定 | /reset-password | ResetPasswordPage | ✅ | ✅ | ✅ | 同上 |
| 認証/共通 | 飼主カルテレポート | /owners/:id/report | OwnerReport | ✅ | ✅ | ✅ | C18 raw: Checkup/Examination history |
| 認証/共通 | 404 Not Found | * | inline element | 🔒 | 🔒 | 🔒 | 対象外1件: lazy pageを持たないinline fallback |
| 会計 | 会計一覧 | /accounting | AccountingList | ✅ | 🔒 | ✅ | `AccountingList.tsx:17`; 金額表示共通化 |
| 会計 | 会計 - ペット選択 | /accounting/select-pet | AccountingPetSelection | ✅ | ✅ | ✅ | 追加指摘なし |
| 会計 | 会計登録 | /accounting/new | AccountingDetail | ✅ | 🔒 | ✅ | ItemList/Refund raw cell・通貨表現 |
| 会計 | 会計詳細 | /accounting/:id | AccountingDetail | ✅ | 🔒 | ✅ | 同上 |
| 会計 | レジ締め | /accounting/close | CashRegisterClosePage | ✅ | 🔒 | ✅ | Billing/summary raw cell |
| 会計 | レジ締め履歴 | /accounting/close/history | CashRegisterHistoryPage | ✅ | 🔒 | ✅ | route raw cell `CashRegisterHistoryPage.tsx` |
| 会計 | 月次集計レポート | /accounting/reports | AccountingReportsPage | ✅ | 🔒 | ✅ | DailyBreakdown raw cell |
| 在庫/見積/シフト | 在庫一覧 | /inventory | InventoryList | ✅ | 🔒 | ✅ | `InventoryList.tsx:10` |
| 在庫/見積/シフト | 在庫登録 | /inventory/new | InventoryForm | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| 在庫/見積/シフト | 在庫編集 | /inventory/:id | InventoryForm | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| 在庫/見積/シフト | 見積一覧 | /estimates | EstimateList | ✅ | 🔒 | ✅ | 禁止`utils/`をfeature indexから参照 |
| 在庫/見積/シフト | 見積作成 | /estimates/new | EstimateForm | ✅ | 🔒 | ✅ | 併発: FE12-04 |
| 在庫/見積/シフト | 見積詳細 | /estimates/:id | EstimateDetail | ✅ | 🔒 | ✅ | `EstimateDetail.tsx:5` |
| 在庫/見積/シフト | 見積編集 | /estimates/:id/edit | EstimateForm | ✅ | 🔒 | ✅ | 併発: FE12-04 |
| 在庫/見積/シフト | シフトカレンダー | /shifts | ShiftCalendarPage | ✅ | ✅ | ✅ | 追加指摘なし |
| カルテ | カルテ一覧 | /medical-records | MedicalRecords | △ FE12-02 | 🔒 | ✅ | C6a danger/異常/RBACレビュー |
| カルテ | カルテ作成 - ペット選択 | /medical-records/select-pet | MedicalRecordPetSelection | △ FE12-02 | ✅ | ✅ | C6a死亡/危険ペット選択 |
| カルテ | カルテ作成 | /medical-records/new | MedicalRecordForm | △ FE12-02 | 🔒 | ✅ | 併発: C18 FE12-01、通貨 FE12-10 |
| カルテ | カルテ編集 | /medical-records/:id | MedicalRecordForm | △ FE12-02 | 🔒 | ✅ | 併発: C18 FE12-01、通貨 FE12-10 |
| 入院/ホテル | 入院・ホテル一覧 | /hospitalization | HospitalizationList | △ FE12-02 | 🔒 | ✅ | C6a死亡表示・操作抑止 |
| 入院/ホテル | 入院・ホテル登録 - ペット選択 | /hospitalization/select-pet | HospitalizationPetSelection | △ FE12-02 | ✅ | ✅ | C6a死亡/危険ペット選択 |
| 入院/ホテル | 入院・ホテル登録 | /hospitalization/new | HospitalizationForm | △ FE12-02 | 🔒 | ✅ | 併発: C18 FE12-01、通貨 FE12-10 |
| 入院/ホテル | 入院・ホテル詳細 | /hospitalization/:id | HospitalizationDetail | △ FE12-02 | ✅ | ✅ | 併発: C18 FE12-01 |
| 入院/ホテル | 入院・ホテル編集 | /hospitalization/:id/edit | HospitalizationForm | △ FE12-02 | 🔒 | ✅ | 併発: C18 FE12-01、通貨 FE12-10 |
| トリミング | トリミング一覧 | /trimming | TrimmingList | ✅ | 🔒 | ✅ | `TrimmingList.tsx:12` |
| トリミング | トリミング登録 - ペット選択 | /trimming/select-pet | TrimmingPetSelection | ✅ | ✅ | ✅ | 追加指摘なし |
| トリミング | トリミング登録 | /trimming/new | TrimmingForm | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| トリミング | トリミング編集 | /trimming/:id | TrimmingForm | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| 検査 | 検査一覧 | /examinations | ExaminationsList | △ FE12-02 | 🔒 | ✅ | C6a異常値/RBAC |
| 検査 | 検査登録 - ペット選択 | /examinations/select-pet | ExaminationPetSelection | △ FE12-02 | ✅ | ✅ | C6a死亡/危険ペット選択 |
| 検査 | 検査登録 | /examinations/new | ExaminationForm | △ FE12-02 | ✅ | ✅ | global listener重複 |
| 検査 | 検査編集 | /examinations/:id | ExaminationForm | △ FE12-02 | ✅ | ✅ | 同上 |
| ワクチン | ワクチン一覧 | /vaccinations | VaccinationList | △ FE12-02 | 🔒 | ✅ | C6a期限超過/死亡/RBAC |
| ワクチン | ワクチン接種 - ペット選択 | /vaccinations/select-pet | VaccinationPetSelection | △ FE12-02 | ✅ | ✅ | C6a死亡/危険ペット選択 |
| ワクチン | ワクチン登録 | /vaccinations/new | VaccinationForm | △ FE12-02 | 🔒 | ✅ | 併発: axis② FE12-04 |
| ワクチン | ワクチン編集 | /vaccinations/:id | VaccinationForm | △ FE12-02 | 🔒 | ✅ | 併発: axis② FE12-04 |
| 定期健診 | 定期健診一覧 | /checkups | CheckupsList | △ FE12-02 | 🔒 | ✅ | C6a要フォロー表示 |
| 定期健診 | 定期健診登録 - ペット選択 | /checkups/select-pet | CheckupPetSelection | △ FE12-02 | ✅ | ✅ | C6a死亡/危険ペット選択 |
| 定期健診 | 定期健診登録 | /checkups/new | CheckupForm | △ FE12-02 | 🔒 | ✅ | C6a臨床status |
| 受付/飼主/予約 | 受付 | / | Reception | △ FE12-02 | ✅ | ✅ | C6a危険/死亡ペットの受付表示 |
| 受付/飼主/予約 | 飼主一覧 | /owners | OwnersList | △ FE12-02 | 🔒 | ✅ | C6a danger/deceased filter、併発: C18 FE12-01 |
| 受付/飼主/予約 | 飼主登録 | /owners/new | OwnerForm | △ FE12-02 | 🔒 | ✅ | OwnerSearchModal C18併発 |
| 受付/飼主/予約 | 飼主編集 | /owners/:id | OwnerForm | △ FE12-02 | 🔒 | ✅ | OwnerSearchModal C18併発 |
| 受付/飼主/予約 | 集計ダッシュボード | /aggregation | AggregationDashboardPage | ✅ | 🔒 | ✅ | `AggregationDashboardPage.tsx:3` |
| 受付/飼主/予約 | 予約管理 | /reservations | ReservationManagement | ✅ | 🔒 | ✅ | `ReservationManagement.tsx:7` |
| 運用/Lステップ | Lステップ健診連携 | /lstep/checkup-sync | CheckupSyncPage | ✅ | 🔒 | ✅ | `CheckupSyncPage.tsx:4` |
| 運用/Lステップ | Lステップ配信モニター | /lstep/delivery-monitor | LstepDeliveryMonitorPage | ✅ | 🔒 | ✅ | `LstepDeliveryMonitorPage.tsx:2` |
| 運用/Lステップ | Lステップ分析 | /lstep/analytics | LstepAnalyticsPage | ✅ | 🔒 | ✅ | C18 raw 26件（3 component） |
| 運用/LINE予約 | LINE予約設定(index) | /line-reservation | LineReservationSettings | ✅ | ✅ | ✅ | 本体route。別app axis③はFE12-08 / FE12-09 |
| 運用/LINE予約 | LINE予約設定 | /line-reservation/settings | LineReservationSettings | ✅ | ✅ | ✅ | 同上 |
| 運用/LINE予約 | LINE予約ページエディタ | /line-reservation/page-editor | LineReservationPageEditor | ✅ | ✅ | ✅ | 同上 |
| 運用/LINE予約 | LINE予約枠設定 | /line-reservation/slots | LineReservationSlotsSettings | ✅ | 🔒 | ✅ | `LineReservationSlotsSettings.tsx:3` |
| 運用/医院 | 医院マスタ設定 | /settings/clinic | ClinicMasterList | ✅ | ✅ | ✅ | 追加指摘なし |
| 運用/マニュアル | マニュアル | /manual | ManualPage | ✅ | 🔒 | ✅ | 独自shellは裁定済み、icon importのみ対象 |
| 運用/マニュアル | マニュアル記事 | /manual/:category/:slug | ManualPage | ✅ | 🔒 | ✅ | 同上 |
| 設定/マスタ | 設定トップ | /settings | MasterSettingsIndex | ✅ | 🔒 | ✅ | `MasterSettingsIndex.tsx:7` |
| 設定/マスタ | 職員マスタ | /settings/staff | StaffSettings | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| 設定/マスタ | 診療項目マスタ | /settings/treatment-items | TreatmentPlanMaster | ✅ | ✅ | ✅ | global listener重複 |
| 設定/マスタ | 診断名マスタ | /settings/diagnosis | DiagnosisSettings | ✅ | ✅ | ✅ | 同上 |
| 設定/マスタ | 動物種マスタ | /settings/animal-species | AnimalSpeciesSettings | ✅ | ✅ | ✅ | 同上 |
| 設定/マスタ | トリミングマスタ | /settings/trimming | TrimmingSettings | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| 設定/マスタ | トリミングコース種別マスタ | /settings/trimming-course-type | TrimmingCourseTypeSettings | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| 設定/マスタ | 薬剤マスタ | /settings/medicine | MedicineSettings | ✅ | ✅ | ✅ | global listener重複 |
| 設定/マスタ | 予約種別マスタ | /settings/reservation-type | ReservationTypeSettings | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| 設定/マスタ | 入院・ホテルマスタ | /settings/hospitalization | HospitalizationSettings | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| 設定/マスタ | ケージマスタ | /settings/cage | CageSettings | ✅ | 🔒 | ✅ | 併発: 通貨 FE12-10 |
| 設定/マスタ | 物販品マスタ | /settings/merchandise-items | MerchandiseItemSettings | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| 設定/マスタ | 保険マスタ | /settings/insurance | InsuranceSettings | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| 設定/マスタ | 職種マスタ | /settings/occupations | OccupationSettings | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| 設定/マスタ | 権限グループマスタ | /settings/permission-groups | PermissionGroupSettings | △ FE12-02 | 🔒 | ✅ | 併発: C18 FE12-01、RBAC review |
| 設定/マスタ | 問診テンプレートマスタ | /settings/inquiry-templates | InterviewTemplateSettings | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| 設定/マスタ | 主訴マスタ | /settings/interview/chief-complaint | ChiefComplaintSettings | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| 設定/マスタ | 問診テンプレート(interview) | /settings/interview/templates | InterviewTemplateSettings | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| 設定/マスタ | シフトテンプレートマスタ | /settings/shift-templates | ShiftTemplateSettings | ✅ | ✅ | ✅ | global listener重複 |
| 設定/マスタ | 締め時間設定 | /settings/closing-time | ClosingSettingsPage | ✅ | 🔒 | ✅ | StandardClosingTime raw cell |
| 設定/マスタ | 支払方法マスタ | /settings/payment-methods | PaymentMethodSettings | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| 設定/マスタ | 割引キャンペーンマスタ | /settings/campaigns | CampaignSettings | ✅ | 🔒 | ✅ | 併発: axis② FE12-04 |
| 設定/マスタ | Lステップ連携設定 | /settings/integrations/lstep | LstepSettingsPage | ✅ | ✅ | ✅ | 追加指摘なし |
| 設定/マスタ | Lステップタグ管理 | /settings/lstep/tags | LstepTagManagementPage | ✅ | 🔒 | ✅ | `LstepTagManagementPage.tsx:2` |
<!-- FE12-ROUTE-TABLE-END -->

## FE12 task backlog

同じ根因をページ数だけ複製しない。`対象route` は修正影響範囲であり、ページ表の主IDまたは「併発」注記から参照する。FE12-07〜09は3ツリー横断監査で検出したglobal taskのため、本体84ページの適合判定を機械的に❌へ変えない。

<!-- FE12-TASK-TABLE-START -->
| ID | Priority | 軸 | 対象route | 証跡(file:line) | 根拠(category/rule or MANDATORY/duplication) | 修正方針 | ②削除判定 | 将来のscoped検証 |
|---|---|---|---|---|---|---|---|---|
| FE12-01 | P1 | ① | owner report、会計new/detail/close/history/reports、medical-record form、hospitalization form/detail、Lstep analytics、permission-groups、closing-time | `frontend/scripts/design-system-audit.mjs:135-161`、`frontend/src/features/hospitalization/components/HospitalizationTreatmentTable.tsx:84` | C18 / DESIGN.md `ex-data-table-cell`; 22ファイル204 raw cellのnon-gating ratchet | 22ファイルを表primitive移行可能、帳票/特殊matrix、構造cellの3群へ再分類し、標準表だけを`TableHead`/`TableCell`へ置換する。1バッチごとにbaseline件数を減らし、allowlistは拡大しない。**Batch 1 COMPLETE (2026-07-24)**: 標準data table 17ファイルを移行し、C18 baselineを204→44へ削減。**Batch 2 COMPLETE (2026-07-24)**: form内table 3ファイル・34 raw cellを移行し、baselineを44→10へ削減。下記実行記録を参照 | **可**。raw `th`/`td`と重複classを削除し既存primitiveへ統合。追加だけは禁止 | `docker compose exec frontend pnpm design-audit` |
| FE12-02 | P0 | ① | owners/reception、medical-records、hospitalization、examinations、vaccinations、checkups、permission-groups | `docs/spec/ui-design-compliance.md:14`、`frontend/src/features/medical-records/routes/MedicalRecords.tsx:261-270`、`frontend/src/features/hospitalization/components/HospitalizationListView.tsx:78` | C6a / design-system.md §2.4・§9。危険/死亡/RBAC非活性は静的網羅不能で臨床安全が最優先 | danger・死亡・異常値・期限超過・権限なしの各sentinelをコードレビューし、色だけでなく文言/操作抑止/accessible nameを確認する。不一致時だけ誤った装飾を既存semantic tokenへ置換し、正常表示の追加はしない。**U1〜U9 COMPLETE (2026-07-24〜25)**: F1/F2/F5〜F8/F10〜F15/F17/F18を解消（〜`c46f51141`）。**U10（F3+F4 death token置換）はF16決裁待ちでBLOCKED**。下記実行記録を参照 | **条件付き可**。適合ならコード追加0。不適合なら誤った色・操作・重複分岐を削除/置換 | `docker compose exec frontend npx vitest run src/features/hospitalization/components/HospitalizationListView.test.tsx src/features/medical-records/routes/MedicalRecords.test.tsx src/features/owners/routes/OwnerForm.permissions.test.tsx src/features/master/routes/MasterReorderPermissionGuards.test.tsx` |
| FE12-03 | P1 | ② | page表で参照した本体route（production 208ファイル） | `frontend/src/features/accounting/routes/AccountingList.tsx:17`、`frontend/src/features/estimates/routes/EstimateList.tsx:10`、`frontend/src/features/master/routes/StaffSettings.tsx:2` | Bundle Size Optimization / `bundle-barrel-imports`。ただしproject Feature Indexing barrelはMANDATORYのため対象外、第三者`lucide-react` root importだけを対象 | Viteのchunk実測を先に取り、対応可能ならicon単位subpath importまたは既存Vite最適化へ置換する。feature indexは維持し、独自icon wrapperは新設しない。**不要と判定 (2026-07-25)**: 前提のchunk実測を取得した結果、`vendor-icons` は **28.00 kB / gzip 9.10 kB** でrolldownが既にtree-shake済み。208ファイルを触る根拠が無い（①要件を疑う→存在すべきでない工程）。実バンドル問題は `manual` 522.69 kB 他であり本行とは別事象。下記実行記録を参照 | **着手しない**。前提が実測で否定された | `rg -n "from .lucide-react." frontend/src`、`docker compose exec frontend npx vitest run src/features/accounting/routes/AccountingRouteGuards.test.tsx src/features/estimates/routes/EstimateList.test.tsx src/features/master/routes/StaffSettings.test.tsx` |
| FE12-04 | P1 | ②/③ | dirty form群・shift/master side peek群 | `frontend/src/hooks/use-side-peek-dirty.ts:31-40`、`frontend/src/hooks/use-unsaved-changes.ts:14-25` | Client-Side Data Fetching / `client-event-listeners`; `frontend/CLAUDE.md` Shared Helper配置MANDATORY / duplication | 共通beforeunload購読を`src/hooks`の既存hookへ統合し、dirty状態APIの差だけをadapterなしで吸収する。2重listener実装と同一handlerを削除する。**COMPLETE (2026-07-25)**: commit `7246dd931`。下記実行記録を参照 | **可**。2実装を1実装へ統合してlistener/effect重複を削除。追加だけは禁止 | `docker compose exec frontend npx vitest run src/hooks/use-unsaved-changes.test.tsx src/features/estimates/routes/EstimateForm.test.tsx src/features/master/routes/MasterReorderPermissionGuards.test.tsx` |
| FE12-05 | P2 | ② | `/hospitalization/:id` | `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx:51-60` | Re-render Optimization / `rerender-simple-expression-in-memo` | primitive文字列比較と`clampDate`の単純導出をinline計算へ置換し、依存配列と不要importを削除する。挙動・初期dateは維持。**COMPLETE (2026-07-25)**: commit `4603130c2`。下記実行記録を参照 | **可**。2つの`useMemo`と不要importを削除し、追加0 | `docker compose exec frontend npx vitest run src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.test.tsx` |
| FE12-06 | P1 | ③ | `/estimates`、`/estimates/new`、`/estimates/:id`、`/estimates/:id/edit` | `frontend/src/features/estimates/utils/estimate-status-options.ts:1-20`、`frontend/src/features/estimates/utils/is-estimate-expired.ts:1-9`、`frontend/src/features/estimates/utils/is-estimate-locked-status.ts:1-13` | `frontend/CLAUDE.md` Shared Helper配置・Shared Constants配置・Feature Indexing MANDATORY | status定数を`src/constants`、再利用helperを`src/lib`の既存責務へ統合し、feature index経由の公開境界を維持する。単一consumerならconsumerへinlineしてファイル自体を消す。**COMPLETE (2026-07-25)**: commit `ef32784d3`。移設先は層逆転lint+共有定数所有規約によりfeature-local `constants/` / `lib/` へ裁定変更。下記実行記録を参照 | **可**。禁止`utils/` directoryと不要な薄いhelperを削除/統合。移動だけでファイル増加しない | `find frontend/src frontend/liff/src frontend/line-reserve/src -type d -name utils -print`、`docker compose exec frontend npx vitest run src/features/estimates/routes/EstimateList.test.tsx src/features/estimates/routes/EstimateForm.test.tsx src/features/estimates/routes/EstimateDetail.test.tsx` |
| FE12-07 | P1 | ③ | 3ツリーglobal（generated model consumers） | `frontend/src/types/generated/models.ts:210-216`、`frontend/src/types/generated/models.ts:497`、`frontend/src/types/generated/models.ts:1114`、`frontend/src/types/generated/models.ts:1816-1822` | `frontend/CLAUDE.md` Type Safety MANDATORY。production generated modelに`any` 17件 | 生成元mappingを修正し、JSONは`unknown`+境界schema、UUIDは`string`へ置換する。生成物の手編集はせず、consumerの不要cast/防御分岐も型確定後に削除する。**COMPLETE (2026-07-25)**: commit `0bafc2770`。`backend/tygo.yaml`へ型mapping 4種追加→`make codegen`で`any` 17→0。consumer修正は0件（全消費側が既にunknown受け/別型正本/消費者ゼロ）。下記実行記録を参照 | **可**。17個の`any`と派生castを削除/置換。schemaは既存境界へ統合し追加だけにしない | `rg -n "\bany\b" frontend/src/types/generated/models.ts`、`docker compose exec frontend npx vitest run src/features/cash-register/category-breakdown.test.ts src/features/estimates/api/transforms.test.ts` |
| FE12-08 | P2 | ③ | line-reserve 3画面（本体84ページ別デザイン監査外） | `frontend/line-reserve/src/pages/ConfirmPage.tsx:32-42`、`frontend/line-reserve/src/pages/TimeSelectPage.tsx:19-22`、`frontend/line-reserve/src/pages/CompletePage.tsx:19-25` | `frontend/CLAUDE.md` Shared Helper配置MANDATORY / 3-tree duplication | 既存`@/shared-liff/jst-date`へ契約一致分を統合し、padding有無を明示引数で維持する。時刻formatterも1箇所へ統合して画面local wrapperを削除する。**COMPLETE (2026-07-25)**: commit `3cf2eb8e6`。下記実行記録を参照 | **可**。date/time formatter 6関数を既存shared helperへ統合・削除し、新規helper treeは作らない | `docker compose exec frontend npx vitest run line-reserve/src/pages/ConfirmPage.test.tsx line-reserve/src/pages/TimeSelectPage.test.tsx` |
| FE12-09 | P1 | ③ | line-reserve ConfirmPage（本体84ページ別デザイン監査外） | `frontend/line-reserve/src/pages/ConfirmPage.tsx:88-139`、`frontend/line-reserve/src/pages/ConfirmPage.tsx:223-225` | `frontend/CLAUDE.md` React 19 Patterns MANDATORY（`useActionState` + `<form action>` + SubmitButton） | manual `submitting` state/try-finally actionをReact 19 form actionへ置換し、既存error/409分岐をaction stateに統合する。二重送信防止とLIFF message順序を維持。**COMPLETE (2026-07-25)**: commit `ef2961906`。下記実行記録を参照 | **可**。`submitting` state、setter、manual pending/finallyを削除し、既存form patternへ統合 | `docker compose exec frontend npx vitest run line-reserve/src/pages/ConfirmPage.test.tsx` |
| FE12-10 | P2 | ③ | accounting/reports/cash-register、medical-records、hospitalization、master金額表示 | `frontend/src/lib/format/number.ts:28-40`、`frontend/src/features/accounting/components/ItemListCard.tsx:311-314`、`frontend/src/features/hospitalization/components/HospitalizationCostSummary.tsx:38-100` | `frontend/CLAUDE.md` Shared Helper配置MANDATORY / feature間duplication | `¥`/`￥`、符号、0の扱いが既存契約と一致する箇所だけ`formatCurrency`/`formatCurrencyOrDash`へ置換する。帳票固有・差額固有は別契約として残し、無理な一律化をしない。**COMPLETE (2026-07-25)**: commit `6e0fe1747`。下記実行記録を参照 | **可**。一致するinline `toLocaleString`と重複分岐を削除し、既存helperへ統合。新helper追加0 | `docker compose exec frontend npx vitest run src/features/accounting/routes/AccountingDetail.test.tsx src/features/cash-register/category-breakdown.test.ts src/features/hospitalization/components/HospitalizationCostSummary.test.tsx` |
| FE12-11 | P1 | ①/② | line-reserve 全画面（webfont読込） | `frontend/line-reserve/src/index.css:2`（削除前）、`frontend/line-reserve/src/pages/TopPage.tsx:15`（削除前）、`frontend/line-reserve/index.html`（link不在） | `DESIGN.md` タイポグラフィ正本（FE11）／CSS仕様「@importは@charset・@layer以外の全ルールに先行」 | Google Fonts `@import` が `@import "tailwindcss"` の後段にありブラウザに無視されていた＝webfontが一度も読まれていない。仕様外の Montserrat は削除し、正本側の Noto Sans JP は `<link>` で実効化する。**COMPLETE (2026-07-25)**: 未コミット。下記実行記録を参照 | **可**。dead `@import` 1行と仕様外 fontFamily override 1件を削除、`<link>` へ置換 | `docker compose exec frontend pnpm build`（CSS警告消滅）・`npx eslint line-reserve/src/pages/TopPage.tsx` |
| FE12-12 | P1 | ②/③ | Sidebar change-password dialog、trimming master-select modal | `frontend/src/components/shared/Layout/Sidebar.tsx:5`、`frontend/src/features/trimming/components/TrimmingLeftColumn.tsx:10` | Vite `INEFFECTIVE_DYNAMIC_IMPORT` / `frontend/CLAUDE.md` Import境界Lint MANDATORY | eager consumerを所有concrete moduleへ向け、barrel経由でlazy payloadを静的巻き込みする2経路を削除する。**COMPLETE (2026-07-25)**: commit `0804c400e`。`MasterSelectModal` chunk 0.07→2.47 kB、該当warning消滅。下記実行記録を参照 | **可**。見せかけのlazy importと専用`Suspense`を削除し、公開barrelとlazy modal自体は維持 | `docker compose exec frontend pnpm build`、`docker compose exec frontend npx vitest run src/components/shared/Layout/Sidebar.test.tsx src/features/trimming/components/TrimmingFormColumns.test.tsx src/features/trimming/components/TrimmingListTable.test.tsx` |
| FE12-13 | P1 | ①/②/③ | route判定表 119 axis cells | route table、関連実行ledgers、current route code | 完了unit・裁定済み例外とroute判定表のdriftをcurrent code証跡で解消する | axis cellだけを`✅`/`🔒`へ置換し、held clinical cellsを保持する。**PARTIAL / BLOCKED (2026-07-25)**: 110セル反映。C18 audit非greenの8セルと残存inline通貨の1セルは未反映。下記ledger参照 | **可**。stale task ID 110件を最終記号へ置換。source追加・変更0 | column限定distribution、marker限定field count、`docker compose exec frontend pnpm design-audit`、current-code route evidence |
<!-- FE12-TASK-TABLE-END -->

## C18 解消バッチ

1. **標準data table**: accounting、accounting-reports、cash-register、medical-records、Lstep、permission-groupを既存table primitiveへ置換し、各バッチでraw baselineを減算する。
2. **form内table**: OwnerSearchModal、HospitalizationTreatmentTable、closing-timeはviewportとfocus順を保ったままcell padding/typographyを正本化する。
3. **非対象の保全**: print、ManualContent、owner-reportの裁定済み特殊matrix、nested structural cellをbaseline解消のために無理に標準化しない。allowlist追加で数を隠さない。

### FE12-01 Batch 1 実行記録（2026-07-24）

- **Status**: COMPLETE。保存プロンプト SHA-256 `ae49f2ee99d3b096f95ede7c83a4be022b098a2fb1ef21ab2766150d94592e05` を validator PASS 後に実行した。
- **変更scope**: accounting 2、accounting-reports 1、cash-register 3、medical-records 7、Lstep 3、master 1の計17 TSXと `frontend/scripts/design-system-audit.mjs` のC18 baseline Map、ならびに本台帳のみ。テストファイル変更なし。stage / commit / stash / pushなし。
- **移行結果**: ratchet対象160件に既準拠raw header 7件とempty-state 3件を加えたraw opening 170件を `TableHead` / `TableCell` へ移行。対象17ファイルのraw `<th>` / `<td>` openingは0件。空状態は `data-empty-state` と `colSpan={6,10,7}` を保持。17 baseline entryを削除し、残り5 entry / 44件。allowlist・検査ロジック・正規表現は変更なし。
- **対象TSX**:
  - `frontend/src/features/accounting/components/ItemListCard.tsx`
  - `frontend/src/features/accounting/components/RefundSection.tsx`
  - `frontend/src/features/accounting-reports/components/DailyBreakdownTable.tsx`
  - `frontend/src/features/cash-register/components/BillingDetailTable.tsx`
  - `frontend/src/features/cash-register/components/UnifiedClosingSummaryTable.tsx`
  - `frontend/src/features/cash-register/routes/CashRegisterHistoryPage.tsx`
  - `frontend/src/features/medical-records/components/CheckupsTab/CheckupsTabRows.tsx`
  - `frontend/src/features/medical-records/components/CheckupsTab/CheckupsTabTable.tsx`
  - `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentRow.tsx`
  - `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentRowParts.tsx`
  - `frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTabParts.tsx`
  - `frontend/src/features/medical-records/components/VitalsTab/VitalsTabRows.tsx`
  - `frontend/src/features/medical-records/components/VitalsTab/VitalsTabTable.tsx`
  - `frontend/src/features/lstep/components/LstepCsvImportSection.tsx`
  - `frontend/src/features/lstep/components/LstepDeliveryStatsSection.tsx`
  - `frontend/src/features/lstep/components/LstepVisitConversionSection.tsx`
  - `frontend/src/features/master/components/PermissionRuleTable.tsx`
- **Validator gate**: `node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-01-c18-batch1.md` → `Prompt Craft Harness Validation: PASS`、exit 0。
- **Scoped test baseline/final**: `docker compose exec frontend npx vitest run src/features/accounting src/features/accounting-reports src/features/cash-register src/features/medical-records src/features/lstep src/features/master` → baseline `Test Files 94 passed (94)` / `Tests 706 passed | 3 skipped (709)` / exit 0、finalも同一 / exit 0、新規failure 0。
- **Design audit**: `docker compose exec frontend pnpm design-audit` → `C18 table cell override — 0 件`、`C18 raw legacy baseline — 44 件（non-gating ratchet）`、`C19 table row onClick — 0 件`、`PASS — 違反 0 件`、exit 0。
- **Scoped lint/diff**: 17 TSXとaudit scriptへの `docker compose exec frontend npx eslint ... --max-warnings 0` はexit 0。`git diff --check -- <18 paths>` はexit 0。開始baselineから新規にdirty化したpathは指定18 pathのみ。
- **Independent review**: 汎用reviewerとTypeScript/React reviewerが独立確認し、双方 `Approve — PASS`。CRITICAL / HIGH / MEDIUM / LOWのactionable findingなし。列内容・順序、幅、`text-right` / `text-center`、`colSpan`、empty state、C19非回帰、audit Map-only差分を確認した。
- **Assumption deviation**: なし。prompt記載の `rg -c "<th|<td"` は `<thead>` も一致し、0件時exit 1の扱いが曖昧なため、意図どおりopening tagだけを判定する `if rg -n '<(th|td)\b' <17 paths>; then FAIL; else PASS; fi` を使用し `RAW_CELL_CHECK=PASS total=0` を確認した。
- **残余risk / 手動full gate**: 本unitの禁止事項に従い全体 `type-check` / `test:run` / `build` は未実行。統合前に手動で `docker compose exec frontend pnpm type-check`、`docker compose exec frontend pnpm test:run`、`docker compose exec frontend pnpm build` を実行する。

### FE12-01 Batch 2 実行記録（2026-07-24）

- **Status**: COMPLETE。保存プロンプト SHA-256 `84485838c1ecbd5b4a57ca6ebc422bff8a24d72d3f924b4dcdcc394626cbabe9` を validator PASS 後に実行した。initial independent reviewで検出したaudit専用testの旧baseline fixtureは、2026-07-24 generating session裁定で同testをallowlistへ追加して解消した。
- **変更scope**: `OwnerSearchModal.tsx`、`HospitalizationTreatmentTable.tsx`、`StandardClosingTimeSection.tsx` の計3 TSX、`frontend/scripts/design-system-audit.mjs` のC18 baseline Map、`frontend/scripts/design-system-audit.test.mjs` のratchet fixture、ならびに本台帳のみ。隣接component test 3件はmarkup非依存で既存の選択操作、read-only、focusable scroll region、checkbox hit target、range再計算を検証済みのため変更なし。stage / commit / stash / pushなし。
- **移行結果**: 対象3ファイルのraw `<th>` / `<td>` opening 34件を `TableHead` / `TableCell` へ移行し、0件を確認した。列内容・順序、table / thead / tbody / tr、横scroll viewport、focus順、入力DOM順、`scope="row"`、clinical read-only / 削除条件を維持した。該当3 baseline entryだけを削除し、残り2 entry / 10件。owner-report 2 entry、allowlist、検査ロジック、正規表現は変更なし。
- **Validator gate**: `node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-01-c18-batch2.md` → `Prompt Craft Harness Validation: PASS`、exit 0。
- **Scoped test baseline/final**: `docker compose exec frontend npx vitest run src/components/shared/OwnerSearchModal src/features/hospitalization src/features/closing-settings src/features/owners` → baseline `Test Files 41 passed (41)` / `Tests 243 passed (243)` / exit 0、finalも同一 / exit 0、新規failure 0。
- **Design audit**: `docker compose exec frontend pnpm design-audit` → `C18 table cell override — 0 件`、`C18 raw legacy baseline — 10 件（non-gating ratchet）`、`C19 table row onClick — 0 件`、`PASS — 違反 0 件`、exit 0。
- **Audit regression test**: `docker compose exec frontend node --test scripts/design-system-audit.test.mjs` → `tests 61` / `pass 61` / `fail 0` / exit 0。注入APIは無いため、fixtureを現存baselineの `CheckupHistorySection.tsx`（4件）へ切り替え、5 cellsに対しbaseline 4 / violation 1 / line 5を検証する最小修正とした。
- **Scoped lint/diff**: 3 TSXとaudit scriptへの既存scoped ESLint、追加testへの `docker compose exec frontend npx eslint scripts/design-system-audit.test.mjs --max-warnings 0` はともにexit 0。拡張後allowlist 9 pathへの `git diff --check` はexit 0。隣接component test差分・対象pathのcached差分は0。本台帳は開始時dirtyだった既存内容を保全した。final status diffのunit-owned contentはallowlist内6 pathのみ。並行session由来driftは2026-07-24 generating session裁定に従いwrite attributionから除外し、AC-08をPASSとした。
- **Independent review**: initial React/a11y reviewerは `Approve / PASS`（finding 0）。initial汎用reviewerのHIGH finding（audit test `actual 0 / expected 10`）は上記fixture修正で解消。final汎用reviewerは現ツリーでaudit test 61/61、design-audit 10件/PASS、raw 0、diff-check、cached path 0、allowlist attributionを再確認し `Approve / PASS`。非blocking LOWとしてFE12-01行のhistorical line証跡精度のみ指摘し、コード・testへのactionable findingはなし。
- **Failure Signature log**:
  - AC-08 final global status gate、initial actual=unit-owned pathに加えて並行session drift。generating sessionの明示裁定によりunit write attributionで再判定し、allowlist内6 pathのみのためPASS。
  - AC-11 audit専用test、initial actual=60/61 PASS（`0 !== 10`）、attempt 1。allowlist拡張後に現存baseline fixtureへ最小修正し、attempt 2で61/61 PASS。
- **Assumption deviation**: なし。raw opening tag判定はBatch 1と同じ `if rg -n '<(th|td)\b' <3 paths>; then FAIL; else PASS; fi` を使用し `RAW_CELL_CHECK=PASS total=0` を確認した。
- **残余risk / 手動full gate**: 本unitの禁止事項に従い全体 `type-check` / `test:run` / `build` は未実行。統合前に手動で `docker compose exec frontend pnpm type-check`、`docker compose exec frontend pnpm test:run`、`docker compose exec frontend pnpm build` を実行する。

## FE12-04 実行記録（2026-07-25）

- **実施**: 外部エージェントCLIへprompt-craft-agent委任（4往復: gate過剰設計によるBLOCKED 2回→契約修正→Resume modeで完遂）。`useSidePeekDirty`を`useUnsavedChanges()`の内部合成へ変更し、逐語重複していたbeforeunload effectと独自stateを削除。`isDirtyRef`・`confirmDiscard`（confirm文言・ref経由stale-closure防止含む）と両hookの公開APIは不変。consumer 27ファイル（useSidePeekDirty=19・useUnsavedChanges=8）無変更。
- **成果物**: commit `7246dd931`（4ファイル・254 insertions/19 deletions）— hooks 2本 + characterizationテスト新設2本（`use-unsaved-changes.test.tsx`・`use-side-peek-dirty.test.tsx`）。
- **検証**: 構造gate（`src/hooks`配下のbeforeunload実装ファイル=1）、新設テスト2 files / 8 tests green、consumer回帰（use-unsaved-changes / EstimateForm / MasterReorderPermissionGuards）3 files / 15 tests green、scoped ESLint 4ファイル指摘0、React/TypeScript独立レビュー両Approve（CRITICAL/HIGH/MEDIUM 0。TypeScript reviewerのdouble assertion指摘1件は実DOM Eventへ是正済み）。
- **備考**: 本行の「将来のscoped検証」が参照する`use-unsaved-changes.test.tsx`は着手時点で未存在 — 本タスクで新設して初めて成立した（計画のspec穴として記録）。台帳転記はFE12-02レーンとの衝突回避のため実装セッションでなく生成側セッションが実施。

## C6a 臨床安全レビュー

- 危険/死亡: OwnersList、Reception、各PetSelection、HospitalizationListでsemantic色・明示文言・操作抑止が同時に保たれること。
- 異常/期限: MedicalRecords、Examinations、Vaccinations、Checkupsで正常値をdanger表示せず、異常/期限超過だけをsemantic色と非色手掛かりで示すこと。
- RBAC: PermissionGroupSettingsと各clinical actionで非活性表示だけに頼らず、mutation不能・accessible name維持を確認する。
- 静的レビューで閉じず、実装期には既存sentinel fixtureのscoped component testを先にRED化してから修正する。

## バグ疑い（todo.md 転記候補）

- なし。今回の検出はリファクタ範囲内の設計/性能/規約負債であり、独立した機能バグとして断定できる証拠はなかった。

## Run Summary（計画起票時）

- Binding pre-read: `AGENTS.md`、`.claude/CLAUDE.md`、`.claude/rules/claude-code-usage.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/refs/typescript-react.md`、`.claude/refs/accessibility-rules.md`、`frontend/CLAUDE.md`、`frontend/src/features/CLAUDE.md`、`DESIGN.md`、`docs/spec/design-system.md`、`docs/spec/ui-design-compliance.md`、`docs/product-philosophy.md`、`~/.agents/skills/vercel-react-best-practices/SKILL.md`、`~/.agents/skills/verification-loop/SKILL.md`。binding矛盾なし。
- Harness: tdd/checklist-first。verification-loopを実際に読了。REDは84 route固定・252セル保留、GREENは84 route・252セル確定と10 taskの構造化。
- Execution loop: sequential。area batch 1〜14を順に監査し、各batchで最新証拠を表へ反映。loop monitoring専用facilityは使用せず、iteration 1=RED完全性、2=design-audit PASS、3=3軸GREEN、4=最終reconciliation。resource limit/budgetは指定なしのため各iterationともn/a。
- 計画のみの裁定: lint/test/build/type-check/installは本unitの禁止コマンドかつCI所掌。実装・De-Sloppifyはコード/テスト変更0のため非該当。将来commandは実装期にscopedで使う。
- 3-tree global overlay: FE12-07（本体generated consumer）、FE12-08 / FE12-09（line-reserve）。LIFF/line-reserveのページ別デザイン判定は行っていない。
- Allowlist attribution: 開始baselineとの差分には並行sessionのbackend WIP driftが含まれたが、本runのwrite tool targetは全回`FE-refactor.md`だけ。外部WIPは読取り以外に触れず、stash/revert/stage/commitしていない。

## FE12-02 C6a レビュー結果

<!-- FE12-02-REVIEW-START -->

### Acceptance Checklist（RED → GREEN）

最初の台帳書込みでは下記8項目と空の29面/104判定セルをREDとして固定し、その後に静的証跡を埋めた。

| ID | Expected behavior | Target surface | Verification method | PASS evidence |
|---|---|---|---|---|
| AC-01 | Lens A/B/C の全対象を空欄・保留なく判定する | 下記3判定表 | Node構造検査で行数・判定値を検査 | `FE12-02 matrix: PASS (30/30 rows; 104/104 verdict cells)` |
| AC-02 | Lens A は semantic token・非色文言・操作抑止・accessible name を個別判定する | A-01〜A-13 | sentinel→DOM→handlerを `file:line` 追跡 | 13面/52セル確定、追加面2件を含む |
| AC-03 | Lens B は偽陽性と欠落の両方向を判定する | B-01〜B-08 | 正常/未来/未判定と異常/期限切れを対で追跡 | 8面/16セル確定 |
| AC-04 | Lens C は mutation fail-closed・token・accessible name・route guard を個別判定する | C-01〜C-09 | route→UI→callback→mutationを追跡 | 9面/36セル確定 |
| AC-05 | 全不適合を連番findingへ結び、必要4項目を記録する | FE12-02-F1〜F18 | ID連続性と4ラベルをNode構造検査 | `FE12-02 findings: PASS (18/18 consecutive; schema 18/18)` |
| AC-06 | 要実測を具体的なブラウザ手順へ結ぶ | M-01〜M-05 | route/fixture/persona/操作/期待結果/証跡を目視・構造確認 | orphan `要実測` 0、全5手順に6項目あり |
| AC-07 | 指定4 test fileを1コマンドで実行する | prompt指定test | exact scoped vitest command | exit 0、`Test Files  4 passed (4)`、`Tests  19 passed (19)` |
| AC-08 | write targetを `FE-refactor.md` のみに限定する | worktree | 開始baselineと最終statusをdiff | **BLOCKED**: 最終記録snapshotで `FE-refactor.md` に加え並行backend WIP 47 tracked + 5 untracked pathが開始baseline後に増加し、global gateを再現不能。本unitのwrite tool targetはledgerだけだが代替してPASSとはしない。 |

判定値は `適合 / 不適合 / 要実測` の3種だけとする。`要実測` は適合扱いしない。

### Lens A — 死亡/危険ペット sentinel

| ID | 対象面 | semantic token | 非色文言 | 操作抑止 | accessible name | 証跡 |
|---|---|---|---|---|---|---|
| A-01 | OwnersList | 不適合 F16 | 適合 | 要実測 M-01 | 適合 | `OwnersListTable.tsx:196-217` はdangerをsemantic表示するが、死亡は `status-helpers.ts:176-178` のgeneric `BADGE.grayHover`で非生存状態を一括表示する。操作名は同`:189-191,235-237`。 |
| A-02 | Reception当日受付 | 不適合 F2 | 不適合 F2 | 不適合 F2 | 不適合 F2 | `reception/api/transforms.ts:79-109` がpet status/dangerをview modelへ写さず、`AppointmentCard.tsx:159-216` は表示もmini-action抑止もできない。 |
| A-03 | 共有PetSelection | 不適合 F1 | 不適合 F1 | 不適合 F1 | 不適合 F1 | 死亡分岐自体は `PetSelectionResultsTable.tsx:46-95` でgray/「死亡・選択不可」/native disabled/ID付きnameを持つが、`use-pet-selection-page.ts:32,63-67` は死亡を取得せずhandlerも無条件navigate、danger sentinelは表にない。 |
| A-04 | AccountingPetSelection | 不適合 F1 | 不適合 F1 | 不適合 F1 | 不適合 F1 | `accounting/routes/AccountingPetSelection.tsx:9-16` がA-03をそのまま使用。固有sentinel分岐なし。 |
| A-05 | MedicalRecordPetSelection | 不適合 F1 | 不適合 F1 | 不適合 F1 | 不適合 F1 | `medical-records/routes/MedicalRecordPetSelection.tsx:9-16` がA-03をそのまま使用。 |
| A-06 | HospitalizationPetSelection | 不適合 F1 | 不適合 F1 | 不適合 F1 | 不適合 F1 | `hospitalization/routes/HospitalizationPetSelection.tsx:9-16` がA-03をそのまま使用。 |
| A-07 | TrimmingPetSelection | 不適合 F1 | 不適合 F1 | 不適合 F1 | 不適合 F1 | `trimming/routes/TrimmingPetSelection.tsx:9-16` がA-03をそのまま使用。 |
| A-08 | ExaminationPetSelection | 不適合 F1 | 不適合 F1 | 不適合 F1 | 不適合 F1 | `examinations/routes/ExaminationPetSelection.tsx:9-16` がA-03をそのまま使用。 |
| A-09 | VaccinationPetSelection | 不適合 F1 | 不適合 F1 | 不適合 F1 | 不適合 F1 | `vaccinations/routes/VaccinationPetSelection.tsx:9-16` がA-03をそのまま使用。 |
| A-10 | CheckupPetSelection | 不適合 F1 | 不適合 F1 | 不適合 F1 | 不適合 F1 | `checkups/routes/CheckupPetSelection.tsx:9-16` がA-03をそのまま使用。 |
| A-11 | HospitalizationListView | 不適合 F3 | 適合 | 適合 | 適合 | `HospitalizationListView.tsx:47-59,77-84` は死亡行をgeneric `opacity-40/C.text40` にする一方、「死亡」を表示しdetail/editを除去。test `:65-73` も抑止を確認。 |
| A-12 | 追加面 HospitalizationBoard | 不適合 F4 | 不適合 F4 | 適合 | 不適合 F4 | `HospitalizationBoard.tsx:39-45,69-82,100-110` は死亡時drag/openを止めるが、generic opacity/borderだけで死亡文言・accessible cueなし。 |
| A-13 | 追加面 Pet死亡記録 | 不適合 F5 | 不適合 F17 | 不適合 F5/F17 | 適合 | `PetDeceasedRecordButton.tsx:27-56` はgeneric muted色、`canEdit`未検査に加え、statusでなく`deceasedAt`だけで記録済みを判定する。`PetCareSection.test.tsx:82-86` は死亡statusでも日時欠落時に「死亡を記録」が出る現状を固定。 |

### Lens B — 異常値/期限超過 sentinel

| ID | 対象面 | 偽陽性なし | 欠落なし | 証跡 |
|---|---|---|---|---|
| B-01 | MedicalRecords 無効担当医 | 適合 | 適合 | `MedicalRecords.tsx:261-270` は有効staffを通常表示し、無効時だけ `C.danger`、`AlertTriangle`、`aria-label="無効な担当医: ..."` を付与。test `MedicalRecords.test.tsx:212-219`。 |
| B-02 | Medical-records embedded examinations | 適合 | 不適合 F6 | `ExaminationGroup.tsx:73-77,97-107` はnormalをgreen、高値をdanger+HIGHにするが、低値をclinical blueでなくbrand tokenにする。 |
| B-03 | Examinations form | 適合 | 不適合 F6 | `ExamItemsTable.tsx:38-60,117-121` はnormalにHIGH/LOWなし、高値はdanger、低値は `C.textBrand/C.bgBrand5/C.bgBrandLight8`。 |
| B-04 | Examinations list/card | 適合 | 要実測 M-02 | `ExaminationsList.tsx:224-226` と `ExaminationCard.tsx:73-76` は通常色のsummaryだけで、items由来HIGH/LOWを一覧へ出す仕様か静的に確定不能。 |
| B-05 | Vaccinations list | 適合 | 不適合 F7 | `VaccinationList.tsx:230` は全`nextDate`をplain `C.text`で表示するため正常値の誤強調はないが、期限超過も識別不能。 |
| B-06 | VaccinationCard | 不適合 F8 | 適合 | `VaccinationCard.tsx:13-18` の`new Date(nextDate) < new Date()`は当日を時刻経過後に期限超過化し得る。実際の過去日は同`:60-73`でdanger+icon+「期限超過」。 |
| B-07 | Checkups期限badge | 適合 | 適合 | `CheckupAlertBadge.tsx:9-19` は過去を「期限切れ」、30日内を「期限間近」、未来/undefinedを非表示。`CheckupsList.test.tsx:152-164,185-210`。 |
| B-08 | Checkups要フォロー | 適合 | 不適合 F9 | `CheckupsList.tsx:297-305` は期限badgeとraw resultのみ。`checkups/api/types.ts:1-15` にfollow-up sentinelがなく、要フォローをsemantic色+文言へ写す経路がない。 |

### Lens C — RBAC非活性

| ID | 対象面 | mutation fail-closed | token | accessible name | route guard整合 | 証跡 |
|---|---|---|---|---|---|---|
| C-01 | PermissionGroup route/view | 適合 | 適合 | 適合 | 適合 | `settings-routes.tsx:193-203`→`RequirePermission.tsx:20-32`→`features/auth/hooks/use-auth.tsx:184-192`でviewをexact判定し、AccessDeniedはtoken+理由文。 |
| C-02 | PermissionGroup CRUD/rules | 不適合 F10 | 不適合 F10 | 不適合 F10 | 不適合 F10 | `MasterPageShell.tsx:49-69` と `MasterCRUDPage.tsx:97-149` はUIをhide/readOnlyにするが、`use-master-save.ts:57-104` と `use-master-crud.ts:213-241` のmutation境界は権限を受け取らない。`PermissionGroupSidePanel.tsx:99-145` はreadOnlyを全fieldへ伝播せず、`PermissionRuleTableRow.tsx:49-60` のcheckboxに関連nameがない。 |
| C-03 | PermissionGroup reorder | 適合 | 適合 | 適合 | 適合 | `PermissionGroupSettings.tsx:76-82` はmutation直前に`if (!canEdit) return`、table `:135-140` は`canEdit`を渡す。`PermissionGroupSortableTable.tsx:48-72` はdragDisabled+ID/name付きdragLabel。指定testもPASS。 |
| C-04 | MedicalRecords top-level C/U/D | 不適合 F11 | 不適合 F11 | 要実測 M-03 | 不適合 F11 | save actionは `use-medical-record-save-action.ts:87-94` で拒否するが、delete callback `MedicalRecordForm.tsx:186-198` は再検査せず、fieldset `:314` は`isFinalized`だけ。`:id` route `clinical-care-routes.tsx:62-66` にedit guardなし。 |
| C-05 | MedicalRecords quick patch/child | 不適合 F12 | 不適合 F12 | 不適合 F12 | 不適合 F12 | UIは `MedicalRecordFormPanels.tsx:70-180` で一部disabled/hideするが、owner変更は`canEdit`を使わず、`use-medical-record-quick-patch-actions.ts:40-143` と `use-medical-record-owner-change.ts:47-88` のmutationは権限を受け取らない。 |
| C-06 | Hospitalization top-level C/U/D | 不適合 F13 | 不適合 F13 | 要実測 M-03 | 適合 | create/edit routeは `clinical-care-routes.tsx:96-139` でaction guard、UIは `HospitalizationForm.tsx:168-192` でhide/fieldset disabled。しかしgeneric opacityに依存し、action `use-hospitalization-form.ts:45-70` とdelete `HospitalizationForm.tsx:109-120` は権限を再検査しない。 |
| C-07 | Hospitalization child mutations | 不適合 F18 | 不適合 F18 | 要実測 M-04 | 不適合 F18 | `CarePlanTab.tsx:40-70` と `DailyRecordsTab.tsx:83-130` はaction別permissionをUI表示に使うだけでmutation直前に再検査しない。残る複合controlのname/操作実効性だけM-04で実測する。 |
| C-08 | Vaccinations C/U/D | 不適合 F14 | 不適合 F14 | 要実測 M-03 | 不適合 F14 | `VaccinationForm.tsx:37-52,158-183` はhide/fieldset disabledだがgeneric opacityに依存し、`use-vaccination-form.ts:195-238,315-328` のsave/delete mutationは権限を受け取らず、`:id` route `clinical-care-routes.tsx:303-307` はview guardのみ。 |
| C-09 | Examinations C/U/D/items | 不適合 F15 | 不適合 F15 | 要実測 M-03 | 不適合 F15 | `ExaminationForm.tsx:36-64,190-234` はhide/fieldset disabledだがgeneric opacityに依存し、`use-examination-form.ts:215-273,288-298` のparent/items/delete mutationは権限を受け取らず、`:id` route `clinical-care-routes.tsx:247-252` はview guardのみ。 |

### Findings

#### FE12-02-F1: 共有PetSelectionが死亡/危険sentinelをend-to-endで運ばない
- Evidence: `frontend/src/hooks/use-pet-selection-page.ts:32,63-67`、`frontend/src/components/shared/PetSelection/PetSelectionResultsTable.tsx:46-95`、7 wrapper各`:9-16`。
- Current code: `const { data: pets } = useGetPets();` と `navigate(selectPath(pet.id))`。tableは死亡時だけ`disabled`/「死亡・選択不可」だがdanger分岐なし。
- Missing: 通常queryで死亡個体を得る契約、dangerのsemantic/非色cue、callback境界の死亡拒否。
- Minimal fix: shared hookで既存`includeDeceased` optionを指定し、`handleSelect`冒頭で死亡をreturnする。tableへ既存danger token+明示文言を追加し、7 wrapper固有追加はしない。

#### FE12-02-F2: Reception view modelが死亡/危険情報を捨てる
- Evidence: `frontend/src/features/reception/api/transforms.ts:79-109`、`frontend/src/features/reception/components/AppointmentCard.tsx:159-216`。
- Current code: transformはpetのid/name/typeだけを保持し、card mini-actionはrecords/accounting/hospitalizationへ無条件navigateする。
- Missing: status/danger_levelの伝播、semantic badge+文言、死亡時の禁止操作、stateを含むaccessible name。
- Minimal fix: generated `reservation.pet` の既存fieldをview modelへ写し、cardで既存semantic tokenを使う。死亡時mini-actionは削除/非活性化し、危険時は警告を追加する。

#### FE12-02-F3: HospitalizationListViewの死亡表現がgeneric opacity
- Evidence: `frontend/src/features/hospitalization/components/HospitalizationListView.tsx:47-50,77-84`。
- Current code: 死亡行は`opacity-40`、文言は`C.text40`。操作抑止と「死亡」は既にある。
- Missing: 死亡という臨床意味に紐づくstatus/death semantic token。
- Minimal fix: generic opacity/text指定を既存death/status tokenへ置換する。文言と抑止ロジックは増やさない。

#### FE12-02-F4: HospitalizationBoardの死亡cardがopacity-only
- Evidence: `frontend/src/features/hospitalization/components/HospitalizationBoard.tsx:39-45,69-82,100-110`。
- Current code: `C.bgPage/C.borderPrimary20/opacity`だけで、drag/openは停止するが「死亡」を描画しない。
- Missing: semantic death token、非色文言、accessibility tree上の死亡状態。
- Minimal fix: generic deceased styleをstatus tokenへ置換し、既存card本文へ「死亡」badge/accessible textを1個追加する。抑止分岐は再利用する。

#### FE12-02-F5: Pet死亡記録componentが`canEdit`を自己検査しない
- Evidence: `frontend/src/components/shared/PetDeceasedRecordButton/PetDeceasedRecordButton.tsx:27-56`、`frontend/src/features/owners/components/PetCareSection.tsx:113-121`。
- Current code: alive時の「死亡を記録」はgeneric muted色でdialogを開き、`canEdit`を分岐に使わない。現callerは祖先fieldsetにも依存する。
- Missing: component境界の権限拒否とdanger/death semantic token。
- Minimal fix: button/dialog openを`canEdit`で拒否し、generic muted classを既存death/danger tokenへ置換する。caller側の重複分岐は追加しない。

#### FE12-02-F6: 異常低値がclinical blueでなくbrand token
- Evidence: `frontend/src/features/examinations/components/ExamItemsTable.tsx:49-56,117-121`、`frontend/src/features/medical-records/components/ExaminationGroup.tsx:73-77,97-103`。
- Current code: LOWに`C.textBrand/C.borderBrand/C.bgBrand5/C.bgBrandLight8`を使用し、HIGHは`C.danger`。
- Missing: `docs/spec/design-system.md:84-87` の異常低値blue semantic。
- Minimal fix: LOW専用のbrand classを既存status-blue tokenへ置換する。既存`LOW`文言は維持し新規componentは作らない。

#### FE12-02-F7: Vaccinations一覧が期限超過を表示しない
- Evidence: `frontend/src/features/vaccinations/routes/VaccinationList.tsx:230`、`frontend/src/features/vaccinations/api/transforms.ts:13`。
- Current code: 全nextDateを`font-mono ${C.text}`で表示する。
- Missing: 過去日だけのdanger token、icon/「期限超過」文言。
- Minimal fix: B6と共用する日付のみ比較helperを再利用し、past行だけ既存danger token+文言へ置換する。未来/空欄は現状維持。

#### FE12-02-F8: VaccinationCardが当日を期限超過にし得る
- Evidence: `frontend/src/features/vaccinations/components/VaccinationCard.tsx:13-18,60-73`。
- Current code: `new Date(nextDate) < new Date()`でdate-onlyと現在時刻を比較する。
- Missing: clinic local dateのstrict past比較。
- Minimal fix: 既存JST ISO date helperとの文字列/date-only比較へ置換する。danger/icon/「期限超過」の表示分岐は維持する。

#### FE12-02-F9: Checkupsに要フォローsentinel経路がない
- Evidence: `frontend/src/features/checkups/routes/CheckupsList.tsx:297-305`、`frontend/src/features/checkups/api/types.ts:1-15`。
- Current code: 次回期限badgeとraw result文字列だけを表示する。
- Missing: 要フォローの型/API fieldからsemantic token+非色文言までの契約。
- Minimal fix: まず既存backend fieldの有無を確認し、存在すればtransform/displayへ写す。存在しない場合のみ仕様責任者が最小field追加を別unitで裁定する。

#### FE12-02-F10: PermissionGroup CRUD/rules mutation境界がUI権限だけに依存
- Evidence: `frontend/src/features/master/hooks/use-master-save.ts:57-104`、`frontend/src/features/master/hooks/use-master-crud.ts:213-241`、`frontend/src/features/master/routes/PermissionGroupSettings.tsx:45-143`。
- Current code: create/edit/delete UIはhide/readOnlyだが、save/delete/rules mutation handlerは`canCreate/canEdit/canDelete`を受け取らない。
- Missing: mutation直前のaction別fail-closed、readOnlyの全field伝播、rule checkboxのresource/actionを含むaccessible name、view-only routeとの明示対応。
- Minimal fix: 共通save/delete handlerへaction別booleanを渡し、mutation直前にreturnする。readOnlyを既存panel fieldへ一括伝播し、checkboxへ関連labelを付ける。個別pageへの重複guard追加は避ける。

#### FE12-02-F11: MedicalRecord top-level edit/delete guardが非対称
- Evidence: `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:143-198,314`、`frontend/src/app/routes/clinical-care-routes.tsx:62-66`、`frontend/src/features/medical-records/hooks/use-medical-record-save-action.ts:87-94`。
- Current code: save actionだけは`!canEdit`を拒否するが、delete handlerは再検査せず、fieldsetは`isFinalized`のみ、edit routeはview guardのみ。
- Missing: delete mutation境界、非編集personaの全fieldset、route edit actionの一貫した防御。
- Minimal fix: delete callbackへ`canDelete`拒否、fieldsetへ`!canSubmit`を統合し、`:id`の編集用途をroute `action="edit"`またはview/edit別routeへ整理する。

#### FE12-02-F12: MedicalRecord quick patchがmutation境界で権限を検査しない
- Evidence: `frontend/src/features/medical-records/hooks/use-medical-record-quick-patch-actions.ts:40-143`、`frontend/src/features/medical-records/components/MedicalRecordFormPanels.tsx:70-180`、`frontend/src/features/medical-records/hooks/use-medical-record-owner-change.ts:47-88`。
- Current code: header controlは一部disabled/hideだがdoctor/visit/date/next-date/finalize handlersは`canEdit`なしでmutationし、owner変更button/handlerも権限を検査しない。
- Missing: programmatic callback/raceに対するfail-closed、owner変更controlの非活性token/name。
- Minimal fix: hooksへ`canEdit/isFinalized`を渡し、各mutation共通入口でreturnする。owner buttonにも同じpermissionを渡す。既存UI disabledは維持し、handlerごとの重複はhelperへ集約する。

#### FE12-02-F13: Hospitalization save/delete actionが権限を受け取らない
- Evidence: `frontend/src/features/hospitalization/hooks/use-hospitalization-form.ts:45-70`、`frontend/src/features/hospitalization/routes/HospitalizationForm.tsx:45-46,109-120,168-192`。
- Current code: route/fieldset/buttonはguardするが、form actionとdelete callbackはpermissionを再検査しない。
- Missing: mutation境界の`canCreate/canEdit/canDelete`拒否。
- Minimal fix: action hookへmode別permissionを渡し、create/update直前で拒否する。delete callback先頭に`!canDelete` returnを追加する。

#### FE12-02-F14: Vaccination save/deleteがUI guardだけで、edit routeもview
- Evidence: `frontend/src/features/vaccinations/routes/VaccinationForm.tsx:37-52,158-183`、`frontend/src/features/vaccinations/hooks/use-vaccination-form.ts:195-238,315-328`、`frontend/src/app/routes/clinical-care-routes.tsx:303-307`。
- Current code: fieldset/buttonを非活性/非表示にするがhook mutationはpermissionを受け取らず、`:id`にedit action guardがない。
- Missing: mutation境界fail-closedとroute/component action整合。
- Minimal fix: hookへmode別permissionを渡してsave/delete直前で拒否し、`:id` routeを`action="edit"`で保護する。

#### FE12-02-F15: Examination parent/items/delete mutationがUI guardだけ
- Evidence: `frontend/src/features/examinations/routes/ExaminationForm.tsx:36-64,190-234`、`frontend/src/features/examinations/hooks/use-examination-form.ts:215-273,288-298`、`frontend/src/app/routes/clinical-care-routes.tsx:247-252`。
- Current code: fieldset/buttonは非活性/非表示だがparent/item replacement/delete mutationはpermissionを受け取らず、`:id` routeはview guardのみ。
- Missing: parentとitemsを同じ権限境界で拒否するfail-closed、edit route整合。
- Minimal fix: action hookへmode別permissionを渡し、parent mutation前に一度拒否してitemsへ到達不能にする。deleteも同様、routeへedit guardを追加する。

#### FE12-02-F16: OwnersListの死亡表示がgeneric gray token
- Evidence: `frontend/src/features/owners/components/OwnersListTable.tsx:214-217`、`frontend/src/lib/status-helpers.ts:176-178`。
- Current code: 死亡を含む非生存statusをすべて`BADGE.grayHover`へまとめる。
- Missing: 死亡という臨床意味に紐づく明示的semantic token、またはgeneric grayを死亡正本とする仕様根拠。
- Minimal fix: status helperの死亡分岐だけを既存death/status tokenへ置換する。適切な既存tokenがない場合に限り、設計正本でtoken名を先に裁定する。

#### FE12-02-F17: 死亡statusと死亡日時の不整合時に再登録導線が出る
- Evidence: `frontend/src/components/shared/PetDeceasedRecordButton/PetDeceasedRecordButton.tsx:36-56`、`frontend/src/features/owners/components/PetCareSection.test.tsx:82-86`。
- Current code: `deceasedAt`だけで記録済みを判定するため、`status=死亡/deceasedAt=null`で「死亡を記録」を表示する。
- Missing: statusと日時の整合性判定、再登録を止める不整合表示。
- Minimal fix: existing pet statusをcomponent境界へ渡し、死亡statusなら日時欠落でも登録buttonを削除する。代わりにデータ不整合の明示文言を表示し、修復は別の監査付き経路へ限定する。

#### FE12-02-F18: Hospitalization child mutationがUI permissionだけに依存
- Evidence: `frontend/src/features/hospitalization/components/CarePlanTab.tsx:40-70`、`frontend/src/features/hospitalization/components/DailyRecordsTab.tsx:83-130`。
- Current code: `canCreate/canEdit/canDelete`はcontrol表示に使うが、care plan/daily/vital/care log/noteのmutation callback直前では再検査しない。
- Missing: action別permissionのmutation境界fail-closed、権限剥奪race/programmatic callbackの拒否。
- Minimal fix: 各tabの共通mutation callback入口へaction別permissionを渡し、mutation前にreturnする。UI側の既存条件は維持し、拒否回帰testをcallback単位で追加する。

### 要実測項目

- **M-01 OwnersList操作範囲**: route=`/owners?include_deceased=true`、fixture=同一ownerにalive/high-danger/deceased、persona=owners editあり。各rowをpointer/keyboardで開き、仕様責任者が死亡/危険時に許可する操作を確定する。期待=警告/nameを失わず禁止操作だけ不能。証跡=screenshot、accessibility tree、GET以外のnetwork log。
- **M-02 Examinations一覧意味とlayout**: route=`/examinations` とmedical-record履歴、fixture=high/low/normal同居、persona=view。1440/1200/800/500pxでsummary/cardを確認。期待=一覧で異常cueが必要ならHIGH/LOWが非色でも識別でき、overflow/重なりなし。証跡=screenshot、accessible text、computed token。
- **M-03 RBAC非活性の理由/name**: routes=`/settings/permission-groups`、`/medical-records/:id`、`/hospitalization/:id/edit`、`/vaccinations/:id`、`/examinations/:id`、fixture=既存record、persona=view-only/create-only/edit-without-delete。pointer/keyboard/programmatic submitを試す。期待=禁止controlはnameと理由を保ち、POST/PATCH/PUT/DELETE 0件。証跡=accessibility tree、screenshot、network log。
- **M-04 Hospitalization child control実効性**: routes=`/hospitalization` boardとdetail、fixture=admitted record+dailies/vitals/care logs/plan、persona=view-only。修正後にdrag/status/daily/vital/log/note/planをpointer/keyboardで試す。期待=全mutation 0件かつ非活性token/name保持。証跡=操作別network log、accessibility tree、screenshot。
- **M-05 Clinical sentinel responsive**: routes=A/B対象全route、fixture=death/high-danger/high/low/past/today/future/empty、persona=通常view。1440/1200/800/500pxで文言、badge、日付、controlを確認。期待=cueのwrap/clip/overlapなし、色だけに依存しない。証跡=4 viewport screenshot、accessible name、console/network log。

### Completion Report

- Run status: BLOCKED

#### Checklist Results

| Checklist item | Expected behavior | Actual behavior | Status | Verification method | Evidence |
|---|---|---|---|---|---|
| AC-01〜04 | 30面/104セルを証跡付きで確定 | 13+8+9面、全セルが3判定値 | PASS | Node構造検査 | `FE12-02 matrix: PASS (30/30 rows; 104/104 verdict cells)` |
| AC-05 | F1〜F18が連続し4要素を持つ | 18 findingを連番・4ラベルで記録 | PASS | Node構造検査 | `FE12-02 findings: PASS (18/18 consecutive; schema 18/18)` |
| AC-06 | 要実測を具体化 | M-01〜M-05へroute/fixture/persona/操作/期待/証跡を記録 | PASS | structured fresh pass | orphan 0、6項目欠落0 |
| AC-07 | scoped test 4件 | 4 files/19 tests green | PASS | exact vitest command | `Test Files  4 passed (4)` / `Tests  19 passed (19)` / exit 0 |
| AC-08 | ledger以外を変更しない | 本unitのwrite targetはledgerだけだが、global baseline後に並行backend WIPが52 path増加 | BLOCKED | exact status diff | `ALLOWLIST_DIFF_EXIT=1`、`ADDED_STATUS_COUNTS ledger=1 external_tracked=47 external_untracked=5`。所有者外のため変更/revert不可。 |
| IR-01 | 全判定セルのevidence qualityを独立照合 | 初回FAILのF16〜F18を反映して再review | PASS | independent reviewer | 104/104判定セルを再照合、未反映actionable finding 0。AC-08 blockerは維持。 |

#### Run Summary

- Changed files: `FE-refactor.md` のみ。code/test/style/token変更0、stage/commit/stash/push 0。
- Failure Signature log: `AC-08` — expected=開始baselineとの差分がledgerのみ、actual=`ADDED_STATUS_COUNTS ledger=1 external_tracked=47 external_untracked=5`、check=`git status --porcelain | sort | diff "${TMPDIR:-/tmp}/fe12-02-git-baseline.txt" -`、attempt=1、hypothesis=並行session WIP drift、attempted fix=所有者外WIPを触らずwrite attributionと最新statusを照合、result=BLOCKED（必要条件: 並行ownerがWIPを確定した後、fresh baselineから再実行）。
- Staged plan ledger: `FE-refactor.md` / FE12-02 entry appended。
- Risk Tier: Local write | Safety boundary events: none。
- Binding pre-read: `AGENTS.md`、`~/.agents/codex/AGENTS.md`、`.codex/config.toml`、`.claude/CLAUDE.md`、`.claude/rules/claude-code-usage.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/refs/typescript-react.md`、`.claude/refs/accessibility-rules.md`、`frontend/CLAUDE.md`、`frontend/src/features/CLAUDE.md`、`frontend/src/hooks/CLAUDE.md`、`docs/product-philosophy.md`、`docs/spec/design-system.md` §2.4/§9、`docs/spec/ui-design-compliance.md` §1 C6a、`FE-refactor.md` FE12-02/C6a、`frontend/src/lib/design-tokens.ts`、auth/permission実装。binding矛盾なし。
- 開始baseline: `git status --porcelain | sort > "${TMPDIR:-/tmp}/fe12-02-git-baseline.txt"`、2026-07-24T17:06:01+0900、631行。最初のledger write前に取得。`FE-refactor.md` targeted statusは空で、FE12-01同時実行痕跡なし。
- Scoped test command: `docker compose exec frontend npx vitest run src/features/hospitalization/components/HospitalizationListView.test.tsx src/features/medical-records/routes/MedicalRecords.test.tsx src/features/owners/routes/OwnerForm.permissions.test.tsx src/features/master/routes/MasterReorderPermissionGuards.test.tsx`。
- Verbatim test result: `Test Files  4 passed (4)`、`Tests  19 passed (19)`、exit 0。非fatal warningはCompose DB env未設定、Vite React SWC推奨、OwnerFormの`No HydrateFallback`。
- Root cause summary: sentinelの局所表示は存在するが、API transform/shared hook/action hook/routeの境界で意味データまたはpermissionが途切れる。最小修正は共有境界での既存field/token/guard再利用を優先し、18 findingとして分離した。
- Harness: tdd/checklist-first。`verification-loop`、`security-review`、`tdd-workflow`を実際に読了。RED=空の29面/104セル、GREEN=30面（Bをcard/list分離）/104セル確定。コード変更0のため80% coverage、lint、build、De-Sloppifyは非該当。sequential loopはAC-01〜07/IR-01 PASS、AC-08 genuine blockerで停止。
- Subagents: plannerがchecklist/matrix、Lens A explorerが死亡/危険、Lens B explorerが異常/期限、Lens C security reviewerがRBACを調査。Aのnative disabledだけをfail-openとする主張は退け、通常query欠落・danger欠落・handler境界を合わせてF1とした。Bのlocal IDはglobal F6〜F9へ再採番。最終統合はmain agentがcurrent codeで再照合した。
- Independent Review: 別reviewerの初回fresh passはA-01/A-13/C-07とallowlist証跡をFAIL。F16〜F18を追加し、AC-08をBLOCKEDへ訂正後に再reviewした。未反映actionable finding 0。
- Validator: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-02-c6a-review.md` → `Prompt Craft Harness Validation: PASS`、`Profile: standard (declared-risk-tier)`、`Target: agent (detected)`、exit 0。
- Assumption deviation: Scope外追加面としてHospitalizationBoardとPet死亡記録をA-12/A-13へ追加。BのVaccinations list/cardを分離したためRED 29面からGREEN 30面になった。コード/テスト修正unitは開始していない。
- Prompt-defect notes / Harness Improvement Feedback: P2 — `frontend/src/hooks/use-auth.ts` を正本と記すが実装は `frontend/src/features/auth/hooks/use-auth.tsx`。P2 — allowlist commandは成功時に差分0を想定する表現だが、許可されたledger変更があるためexit 1が正常であり、差分内容を検査する必要がある。eval captureは生成側session所掌。
- Remaining risks/follow-ups: F1〜F18は未修正。M-01〜M-05は本unitのout-of-scopeで未実測。AC-08は並行WIP確定後のfresh baseline再実行が必要。後続修正unitはfinding単位でTDD化し、臨床fixtureを正常/死亡/危険/high/low/past/today/future/RBAC personaへ拡張する。全体lint/test/build/type-checkはprompt禁止により未実行。

### 生成側裁定（2026-07-24・Mode 3 reconciliation）

- **AC-08 → PASS（帰属裁定）・Run status=COMPLETE へ再分類**。非backendドリフトが `BE-refactor.md`・`docs/ops/README.md`（並行セッション管轄文書）のみであることを executor ベースライン（631行・17:06）との突合で独立確認。write log=ledger のみと整合し、fresh baseline 再実行は不要（帰属で解決する構造問題であり、活動中の共有ツリーで global diff gate は恒久に充足不能）。
- スポット照合3件（F1 `use-pet-selection-page.ts` の `useGetPets()` 無option+無条件navigate / F8 `VaccinationCard` の date-only×現在時刻比較 / F13 hospitalization action hook の権限検査不在）は全て実コードで裏付け確認。
- prompt側欠陥2件を受理: ①`use-auth.ts` パス誤記（実体= `features/auth/hooks/use-auth.tsx`）②allowlist gate は「diff exit 0」でなく「差分内容⊆allowlist」で判定すべき（ledger 変更が正当にあるため exit 1 が正常）。以後の unit プロンプトへ反映する。
- **Finding 裁定**: F1〜F18 のうち16件を修正unit化承認（下記グルーピング）、2件を決裁待ちへ分離。Lens C の fail-open 群（F10〜F15/F18）は backend RBAC が最終防壁のため P0 セキュリティ穴ではなく defense-in-depth + C6a 整合の P1 と裁定。
  - U1: F2（Reception sentinel伝播・実質P0） / U2: F1（共有PetSelection） / U3: F5+F17（死亡記録button） / U4: F7+F8（vaccination期限判定+一覧） / U5: F6（LOW→clinical blue） / U6: F11+F12（medical-records mutation境界） / U7: F13+F18（hospitalization mutation境界） / U8: F14+F15（vaccination/examination mutation境界+edit route guard） / U9: F10（permission-group mutation境界+a11y） / U10: F3+F4（death token置換・F16決裁後）
  - **決裁待ち（要件責任者=曽我）**: F16=死亡表示の正本token裁定（generic gray を死亡正本とするか、death semantic token を §2.4 へ新設するか）／ F9=checkups「要フォロー」sentinel の要否（backend field 実在確認と①要件の妥当性判断が先）

### U1/F2 実行 ledger（2026-07-24）

- Status: COMPLETE — 実装・deterministic gate・allowlist再照合・Independent Review は全て PASS。
- Product gate: 要件責任者=曽我、目的=死亡個体からカルテ/施術・会計・入院へ誤到達する受付動線を物理的に除去する臨床安全修正。新規画面・入力・確認dialog・自動化は追加せず、既存card clickは維持する。
- Changed files: `frontend/src/features/reception/api/transforms.ts`、`frontend/src/features/reception/api/transforms.test.ts`、`frontend/src/features/reception/components/AppointmentCard.tsx`、`frontend/src/features/reception/components/AppointmentCard.test.tsx`、本ledger。
- Root cause / minimal patch: Reception transformがgenerated `Pet.status` / `Pet.danger_level`をview modelへ写さず、card側がsentinelを判定できなかった。generated `PetStatus` / `DangerLevel`由来のoptional fieldを追加し、既存の完全object literal consumerを壊さずadditiveに伝播。`PetStatusDeceased` positive matchでmini-action群全体を非描画し、`DangerLevelHigh`は既存tokenの警告badgeだけを追加した。
- Saved Prompt Validation Gate: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-u1-reception-sentinel.md` → `Prompt Craft Harness Validation: PASS`、`Profile: standard (declared-risk-tier)`、`Target: agent (detected)`、exit 0。
- 開始baseline: `git status --porcelain | sort > "${TMPDIR:-/tmp}/fe12-u1-git-baseline.txt"` を最初の編集前に実行。対象5pathの開始statusは空。`docker compose exec frontend npx vitest run src/features/reception` → `Test Files  11 passed (11)`、`Tests  90 passed (90)`、exit 0。出力は `${TMPDIR:-/tmp}/fe12-u1-vitest-baseline.txt` に保存。
- 開始design baseline: `docker compose exec frontend pnpm design-audit` → `design-system-audit: C18 raw legacy baseline — 10 件（non-gating ratchet）`、`design-system-audit: PASS — 違反 0 件`、exit 0。
- TDD RED: `docker compose exec frontend npx vitest run src/features/reception/api/transforms.test.ts src/features/reception/components/AppointmentCard.test.tsx` → `Test Files  2 failed (2)`、`Tests  5 failed | 42 passed (47)`、exit 1。失敗はsentinel写像1件、死亡badge3件、危険badge1件で、旧実装の欠落と一致。
- TDD GREEN: 同command → `Test Files  2 passed (2)`、`Tests  47 passed (47)`、exit 0。
- Reception regression: `docker compose exec frontend npx vitest run src/features/reception` → `Test Files  11 passed (11)`、`Tests  98 passed (98)`、`FE12_FINAL_RECEPTION_VITEST_EXIT=0`。baseline比の新規失敗0、新規8 case全PASS。
- Scoped lint: `docker compose exec frontend npx eslint src/features/reception/api/transforms.ts src/features/reception/api/transforms.test.ts src/features/reception/components/AppointmentCard.tsx src/features/reception/components/AppointmentCard.test.tsx --max-warnings 0` → stdout 0行、`FE12_FINAL_SCOPED_ESLINT_EXIT=0`。
- Design regression: `docker compose exec frontend pnpm design-audit` → `design-system-audit: C18 raw legacy baseline — 10 件（non-gating ratchet）`、`design-system-audit: PASS — 違反 0 件`、`FE12_FINAL_DESIGN_AUDIT_EXIT=0`。開始時と件数不変。
- De-Sloppify / review correction: 初版の直接field追加はReturnType上required keyとなり、既存完全object literal consumerを型上壊し得るHIGHをfresh reviewで検出。allowlist内transformだけでtyped optional spreadへ修正し、新規Reservation/Pet test fixtureもgenerated完全形 + `satisfies`へ強化した。scope外consumerは編集していない。
- Post-review-fix reconciliation: `docker compose exec frontend npx vitest run src/features/reception` → `Test Files  11 passed (11)`、`Tests  98 passed (98)`、`FE12_POST_REVIEW_FIX_VITEST_EXIT=0`。同4pathのscoped ESLint → `FE12_POST_REVIEW_FIX_ESLINT_EXIT=0`。`docker compose exec frontend pnpm design-audit` → `design-system-audit: C18 raw legacy baseline — 10 件（non-gating ratchet）`、`design-system-audit: PASS — 違反 0 件`、`FE12_POST_REVIEW_FIX_DESIGN_AUDIT_EXIT=0`。`git diff --check` → stdout 0行、`FE12_POST_REVIEW_FIX_DIFF_CHECK_EXIT=0`。
- Source contract check: 追加行のraw `"deceased"` / `"high"`比較、生hex、inline `style`をpost-fix diffで再検査 → `FE12_RECONCILED_GENERATED_CONSTANT_CHECK=PASS`、`FE12_RECONCILED_RAW_HEX_CHECK=PASS`、`FE12_RECONCILED_INLINE_STYLE_CHECK=PASS`。
- Allowlist: `git status --porcelain | sort | diff "${TMPDIR:-/tmp}/fe12-u1-git-baseline.txt" -` → `M FE-refactor.md`、`M frontend/src/features/reception/api/transforms.test.ts`、`M frontend/src/features/reception/api/transforms.ts`、`M frontend/src/features/reception/components/AppointmentCard.test.tsx`、`M frontend/src/features/reception/components/AppointmentCard.tsx` の5行のみ、`FE12_RECONCILED_ALLOWLIST_DIFF_EXIT=1`（許可変更があるためexpected）。差分内容はallowlist⊆5path、並行owner由来の新規drift 0。
- Failure Signature log: attempt 1 / transform — expected=`petStatus=deceased`・`petDangerLevel=high`、actual=両field `undefined`、verification=RED command、error=`AssertionError: expected undefined to be 'deceased'`、fix=generated fieldの直接写像、result=GREEN。attempt 1 / card — expected=死亡/危険badgeと死亡時action不在、actual=badge不在・action表示、verification=RED command、error=`TestingLibraryElementError: Unable to find an element with the text`、fix=generated定数比較・正本badge・single positive deceased branch、result=GREEN。attempt 1 / additive type — expected=既存consumerを壊さないoptional field、actual=初版がrequired-with-undefined、verification=fresh reviewer source trace、error=既存完全literalでmissing-property型退行、fix=typed optional spread + generated完全fixture、result=post-fix gate PASS・reviewer Approve。同一signatureの再発なし。
- Runtime contract / Assumption: `petStatus === deceased` の時だけactionを抑止する。`petStatus=undefined`（pet未紐付け）は判定不能としてbadgeなし・従来action維持。危険highは警告のみでaction維持。card clickは死亡時も維持。disabled化への切替えは不要だった。
- Independent Review: general reviewer + TypeScript/React reviewerの2 fresh passで最終 `CRITICAL/HIGH/MEDIUM/LOW actionable finding 0`、Approve。fail-open分岐、alive/undefined退行、general/trimming/hospitalization全action、generated型、正本token、accessible text、allowlistを再照合した。
- Binding pre-read: `AGENTS.md`、`~/.agents/codex/AGENTS.md`、`.codex/config.toml`、`.claude/CLAUDE.md`、`.claude/rules/claude-code-usage.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/refs/typescript-react.md`、`.claude/refs/accessibility-rules.md`、`frontend/CLAUDE.md`、`frontend/src/features/CLAUDE.md`、`docs/product-philosophy.md`、`FE-refactor.md` FE12-02 C6a/F2/U1、`PatientInfoCard.tsx`、`OwnersListTable.tsx`、generated `models.ts`、`tdd-workflow`、`react-testing`、`verification-loop`。binding矛盾なし。
- Harness / loop: tdd + sequentialを実運用。planner、explorer、tdd-guideのread-only passを統合し、RED→GREEN→De-Sloppify→full reception regressionまで完了。全projectのtype-check/test/buildとcoverage ratchetはprompt禁止のため未実行で、CI/ユーザー手動gateとして残す。
- Prompt-defect / Harness Improvement Feedback: P2 — prompt記載の既存test件数（transforms 29 / AppointmentCard 13）は実体（28 / 11）とdriftしていた。件数固定をPASS条件にせず、named caseとbaseline/final差分で判定した。その他のeval regression captureは不要。
- Remaining scope / risk: U2〜U10、modal、Kanban、hooks、paths、backend、generated型は未変更。別のalive petへ予約編集した直後は既存kanban mergeが旧danger sentinelをquery invalidation/refetchまで一時保持し得るが、deceased petは通常selectionへ出ず、本unitの死亡fail-open経路ではない。hooksは明示out-of-scopeのためnon-blocking follow-upとして記録。frontend CI coverage baseline 43.78% / tolerance 0.5ppのratchetはremote CIまたは手動full coverageで確認する。

### U2/F1 実行 ledger（2026-07-24）

- Status: COMPLETE — shared PetSelectionの死亡取得・callback拒否・危険非色cue、deterministic gate、De-Sloppify、Independent Reviewは全てPASS。
- Product gate: 要件責任者=曽我、目的=7臨床フローの入口で死亡個体を「存在しない」と誤認させず、死亡個体への遷移をcallback境界でも物理的に拒否する臨床安全修正。新規画面・入力・確認dialog・自動化・wrapper固有分岐は追加していない。
- Changed files: `frontend/src/hooks/use-pet-selection-page.ts`、`frontend/src/hooks/use-pet-selection-page.test.ts`、`frontend/src/components/shared/PetSelection/PetSelectionResultsTable.tsx`、`frontend/src/components/shared/PetSelection/PetSelectionResultsTable.test.tsx`、本ledger。
- Root cause / minimal patch: shared hookが`useGetPets()`をoptionなしで呼び死亡個体をtableへ届かせず、`handleSelect`も無条件navigateしていた。`useGetPets(undefined, { includeDeceased: true })`と`pet.status === "死亡"`のpositive guardを追加し、tableのペット名cellに`pet.dangerLevel === "高"`時だけOwnersList正本token+固定文言「⚠ 危険」を追加した。既存死亡badge・grayscale・disabled・aria-label・文言、7 wrapper、`use-pet.ts`、transform、query key定義は未変更。
- Saved Prompt Validation Gate: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-u2-pet-selection-sentinel.md` → `Prompt Craft Harness Validation: PASS`、`Profile: standard (declared-risk-tier)`、`Target: agent (detected)`、exit 0。prompt SHA-256=`7096cfcfb591c03a5b70a61811bab2931bf271228ff002c95f609ff2e0d03f00`。
- 開始baseline: `git status --porcelain | sort > "${TMPDIR:-/tmp}/fe12-u2-git-baseline.txt"`を最初の編集前に実行。対象5pathの開始statusは空。`docker compose exec frontend npx vitest run src/hooks/use-pet-selection-page.test.ts src/components/shared/PetSelection` → `Test Files  2 passed (2)`、`Tests  7 passed (7)`、`VITEST_BASELINE_EXIT=0`。出力は`${TMPDIR:-/tmp}/fe12-u2-vitest-baseline.txt`へ保存。
- 開始design baseline: `docker compose exec frontend pnpm design-audit` → `design-system-audit: C18 raw legacy baseline — 10 件（non-gating ratchet）`、`design-system-audit: PASS — 違反 0 件`、exit 0。
- TDD RED: `docker compose exec frontend npx vitest run src/hooks/use-pet-selection-page.test.ts src/components/shared/PetSelection/PetSelectionResultsTable.test.tsx` → `Test Files  2 failed (2)`、`Tests  3 failed | 7 passed (10)`。failureはincludeDeceased optionが`undefined`、死亡callbackがnavigateを1回実行、危険badgeが不在の3件で旧実装の欠落と一致。
- TDD GREEN / final regression: `docker compose exec frontend npx vitest run src/hooks/use-pet-selection-page.test.ts src/components/shared/PetSelection` → `Test Files  2 passed (2)`、`Tests  11 passed (11)`、exit 0。baseline比の新規失敗0、追加4case（includeDeceased、死亡拒否、生存/undefined通過、危険badge表示/非表示）PASS。
- Scoped lint: `docker compose exec frontend npx eslint src/hooks/use-pet-selection-page.ts src/hooks/use-pet-selection-page.test.ts src/components/shared/PetSelection/PetSelectionResultsTable.tsx src/components/shared/PetSelection/PetSelectionResultsTable.test.tsx --max-warnings 0` → ESLint stdout 0行（Composeの未設定DB env warningのみ）、exit 0、警告0。
- Design regression: `docker compose exec frontend pnpm design-audit` → C1/C3/C5〜C19 gating項目0件、`C18 raw legacy baseline — 10 件（non-gating ratchet）`、`PASS — 違反 0 件`、exit 0。開始時から件数不変。
- Source/diff contract: positive sentinel=`use-pet-selection-page.ts:66 pet.status === "死亡"`、`PetSelectionResultsTable.tsx:60 pet.dangerLevel === "高"`。正本badge class/textは同`:61-62`。added raw hex / inline style=`FORBIDDEN_ADDITION_CHECK=PASS`、既存production死亡UI差分=`DEATH_UI_UNCHANGED_CHECK=PASS`、`git diff --check` stdout 0行/exit 0、added `any`/raw HTML/secret pattern=`TYPE_SECURITY_ADDITION_CHECK=PASS`。
- Shared propagation / cache contract: `rg`でAccounting/Checkup/Examination/Hospitalization/MedicalRecord/Trimming/Vaccinationの7 wrapperがshared hook/tableを消費することを確認し、wrapper diffは空。`query-keys.ts:275-280`は`includeDeceased: true`を通常`["pets"]`と別keyへ写すためcache衝突なし。
- Allowlist: `git status --porcelain | sort | diff "${TMPDIR:-/tmp}/fe12-u2-git-baseline.txt" -` → 本unit所有の追加差分は`FE-refactor.md`と4 TS/TSXの計5行、`FE12_U2_ALLOWLIST_DIFF_EXIT=1`（許可変更があるためexpected）。並行ownerの`backend/internal/reservation/reservation_clinic_isolation_test.go`と`reservation_repository.go`が開始時`M `から完了時`MM`へ変化した2path driftも表示されたが、本runのwrite target/log外であり編集・revert・stageしていない。owned target statusと`git diff --name-only -- <5 allowlist paths>`は正確に5path。
- Failure Signature log: attempt 1 / AC-01 — expected=`useGetPets(undefined,{includeDeceased:true})`、actual=第2引数`undefined`、verification=TDD RED、error=`expected vi.fn() to be called with arguments`、fix=shared hook option追加、result=GREEN。attempt 1 / AC-02 — expected=死亡petでnavigate 0回、actual=1回、verification=TDD RED、error=`expected vi.fn() to not be called at all`、fix=URL生成前のpositive guard、result=GREEN。attempt 1 / AC-04 — expected=「⚠ 危険」1件、actual=0件、verification=TDD RED、error=`Unable to find an element with the text: ⚠ 危険`、fix=正本badge分岐、result=GREEN。同一signature再発なし。
- Runtime contract / Assumption: `status === "死亡"`だけを拒否する。`status=undefined`（判定不能）はprompt裁定どおり従来のnavigateを維持し、専用testで固定した。危険「高」は警告表示のみで選択を抑止しない。toast/dialogは追加していない。
- De-Sloppify: product behaviorに直結する4追加caseと既存死亡無退行caseだけを保持。console、commented-out code、catch-all、helper、map、rename、import並べ替え、drive-by cleanupなし。cleanup後のscoped Vitest 11/11、ESLint、design-audit、diff-checkは全てPASS。
- Independent Review: general reviewer、React hooks/a11y reviewer、healthcare reviewerの3 fresh read-only passが`CRITICAL/HIGH/MEDIUM/LOW actionable finding 0`、Approve。fail-open分岐、undefined挙動、既存死亡UI、7 wrapper波及、query key、accessible非色cue、allowlistを再照合。reviewerによるfile/git writeなし。
- Binding pre-read: `AGENTS.md`、`~/.agents/codex/AGENTS.md`、`.codex/config.toml`、`.claude/CLAUDE.md`、`.claude/rules/claude-code-usage.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/refs/typescript-react.md`、`.claude/refs/accessibility-rules.md`、`frontend/CLAUDE.md`、`frontend/src/hooks/CLAUDE.md`、`frontend/src/components/shared/CLAUDE.md`、`docs/product-philosophy.md`、`FE-refactor.md` FE12-02 C6a/F1/U2/U1、`use-pet.ts`、`OwnersListTable.tsx`、`transforms/pet.ts`、`query-keys.ts`、`tdd-workflow`、`react-testing`、`verification-loop`。binding矛盾なし。
- Harness / orchestration: tdd + sequential（code/test変更時De-Sloppify overlay）を実運用。plannerのread-only passを統合し、main agentがRED→GREEN→reconciliation、3 reviewerがIndependent Reviewを担当。stop condition=全Acceptance Checklist PASS。loop monitoringはlong-lived loopでないため非該当。prompt-defect / Eval Regression Capture / Harness Improvement Feedbackはnone needed。
- Coverage / full gates: executable正本はfrontend statements 43.78%、tolerance 0.5ppのCI ratchetで、`coverage.thresholds`は未設定。coverage artifact生成はallowlist外書込みとなり、full `test:coverage` / `type-check` / `test:run` / `build`はprompt禁止のため未実行。CIまたはユーザー手動で`docker compose exec frontend pnpm type-check`、`docker compose exec frontend pnpm test:run`、`docker compose exec frontend pnpm build`を実行する。
- Remaining scope / risk: U3〜U10、M-05 responsive browser実測、wrapper個別E2E、backend/API/generated型/token変更は未着手。badgeによる狭幅cellの実ブラウザwrap/clipは本unitの静的/RTL gateでは未実測。stage/commit/stash/push 0。

### U3/F5+F17 実行 ledger（2026-07-24）

- Status: COMPLETE — 共有component自身の権限拒否、死亡status/日時不整合時の再登録遮断、既存danger token置換、caller status中継、TDD、deterministic gate、De-Sloppify、Independent Reviewを完了。
- Product gate: 要件責任者=曽我、目的=権限なしstaffへの死亡登録導線と、`status=死亡/deceasedAt`欠落時の監査経路を迂回し得る再登録導線を物理的に除去する臨床安全修正。新規画面・入力・確認dialog・修復機能・自動化は追加していない。
- Changed files: `frontend/src/components/shared/PetDeceasedRecordButton/PetDeceasedRecordButton.tsx`、同`PetDeceasedRecordButton.test.tsx`、`frontend/src/features/owners/components/PetCareSection.tsx`、同`PetCareSection.test.tsx`、本ledger。
- Root cause / minimal patch: 共有componentがalive分岐で`canEdit`を使用せず、記録済み判定も`deceasedAt`だけだった。optional `petStatus`を追加し、`deceasedAt`あり→既存Banner、`petStatus === "死亡"`→固定不整合文言、`canEdit === true`→既存button/dialog、それ以外→`null`の順へ再構成。buttonの`C.text50/C.hoverText`を`C.danger`へ置換し、callerは`petStatus={formData.status}`の1行だけを中継した。
- Saved Prompt Validation Gate: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-u3-deceased-record-guard.md` → `Prompt Craft Harness Validation: PASS`、`Profile: standard (declared-risk-tier)`、`Target: agent (detected)`、exit 0。SHA-256=`d63756a982fd27afb523f470f8ddd806c8f5c83de533b663e918988c28475bfc`。
- 開始baseline: `git status --porcelain | sort > "${TMPDIR:-/tmp}/fe12-u3-git-baseline.txt"`を最初の編集前に実行し、allowlist 5pathはclean。`docker compose exec frontend npx vitest run src/components/shared/PetDeceasedRecordButton src/features/owners/components/PetCareSection.test.tsx` → `Test Files  3 passed (3)`、`Tests  21 passed (21)`、`FE12_U3_VITEST_BASELINE_EXIT=0`。出力は`${TMPDIR:-/tmp}/fe12-u3-vitest-baseline.txt`へ保存。
- 開始design baseline: `docker compose exec frontend pnpm design-audit` → C1/C3/C5〜C19 gating項目0件、`design-system-audit: C18 raw legacy baseline — 10 件（non-gating ratchet）`、`design-system-audit: PASS — 違反 0 件`、`FE12_U3_DESIGN_BASELINE_EXIT=0`。
- TDD RED: `docker compose exec frontend npx vitest run src/components/shared/PetDeceasedRecordButton/PetDeceasedRecordButton.test.tsx src/features/owners/components/PetCareSection.test.tsx` → `Test Files  2 failed (2)`、`Tests  4 failed | 9 passed (13)`、`FE12_U3_TDD_RED_EXIT=1`。失敗は権限なしbutton残存、`C.danger`欠落、component/callerの固定不整合文言欠落の4件で旧F5/F17と一致。Banner優先・`petStatus`未指定互換はRED時点からPASS。
- TDD GREEN: 同command → `Test Files  2 passed (2)`、`Tests  13 passed (13)`、`FE12_U3_TDD_GREEN_EXIT=0`。
- Final scoped regression: `docker compose exec frontend npx vitest run src/components/shared/PetDeceasedRecordButton src/features/owners/components/PetCareSection.test.tsx` → `Test Files  4 passed (4)`、`Tests  26 passed (26)`、`FE12_U3_FINAL_SCOPED_VITEST_EXIT=0`。baseline比の新規失敗0、新規5 case（権限拒否、alive dialog、danger class、不整合遮断、Banner優先、未指定互換を5 testで被覆）と反転F17 caseは全PASS。既存Dialog error caseの`Non-Axios Error: network error` stderrはbaseline/final共通の期待済みtest出力。
- Scoped lint: `docker compose exec frontend npx eslint src/components/shared/PetDeceasedRecordButton/PetDeceasedRecordButton.tsx src/components/shared/PetDeceasedRecordButton/PetDeceasedRecordButton.test.tsx src/features/owners/components/PetCareSection.tsx src/features/owners/components/PetCareSection.test.tsx --max-warnings 0` → ESLint stdout 0行（Composeの未設定DB env warningのみ）、`FE12_U3_SCOPED_ESLINT_EXIT=0`。
- Design regression: `docker compose exec frontend pnpm design-audit` → C1/C3/C5〜C19 gating項目0件、`C18 raw legacy baseline — 10 件（non-gating ratchet）`、`PASS — 違反 0 件`、`FE12_U3_FINAL_DESIGN_AUDIT_EXIT=0`。開始時から全件数不変。
- Source/diff contract: production positive matchは`PetDeceasedRecordButton.tsx`の`petStatus === "死亡"`と`canEdit === true`。`PetCareSection.tsx` zero-context diffは`petStatus={formData.status}`の追加1行だけ。added raw hex / inline style / `any` / `C.text50` / `C.hoverText`=`FE12_U3_FORBIDDEN_ADDITION_CHECK=PASS`、secret/console/raw HTML/catch-all=`FE12_U3_SECURITY_SLOPPY_ADDITION_CHECK=PASS`、allowlist scoped `git diff --check`=`FE12_U3_SCOPED_DIFF_CHECK_EXIT=0`。
- Allowlist: `git status --porcelain | sort | diff "${TMPDIR:-/tmp}/fe12-u3-git-baseline.txt" -` → `FE12_U3_ALLOWLIST_DIFF_EXIT=1`（本unitの許可変更があるためexpected）。本unit帰属は上記5pathのみ。実装完了時のtargeted statusはproduction/test 4path、ledger追記後は正確に5path。開始baselineにあった大規模backend staged migrationが並行ownerにより本run中に確定され、global diffにはそのstatus消失が大量表示されたが、開始時target 4 tracked pathのSHA-256とclean status、新規testの不存在、完了時targeted status/`git diff --name-only`/`git ls-files --others`で本unit writeを分離。外部WIPは編集・revert・stageしていない。
- Failure Signature log: attempt 1 / F5 permission — expected=alive+`canEdit=false`でbutton不在、actual=`死亡を記録` button残存、verification=TDD RED、error=`expected document not to contain element`、fix=`canEdit === true` branch内だけへbutton/dialogを限定、result=GREEN。attempt 1 / danger token — expected=`C.danger`、actual=`C.text50`+`C.hoverText`、verification=TDD RED、error=`expected ... to contain 'text-[#C0392B]'`、fix=既存danger tokenへ置換、result=GREEN。attempt 1 / F17 component+caller — expected=固定不整合文言+button不在、actual=文言なし+button表示、verification=TDD RED、error=`Unable to find an element with the text`（2 case）、fix=positive status branch+caller prop中継、result=GREEN。同一signature再発なし。
- Runtime contract / Assumption: `deceasedAt`を最優先し、`petStatus === "死亡"`だけを不整合として遮断する。`petStatus=undefined`はprompt裁定どおりalive互換だが、唯一のlive callerは同時に`formData.status`を渡すため実質到達しない。live editはOwnersList→`/owners/:id`→owner detail DTO→`deceased_at`→`deceasedAt`と伝播し、日時を持たないOwnersList内Pet modalは`petModal.open`呼出しがなくdormantであることをsource traceで確認。
- De-Sloppify: product behaviorに直結する新規5 testと反転1 caseだけを保持。新規helper/component、console、commented-out code、catch-all、drive-by cleanup、Banner/Dialog変更なし。React reviewerのMEDIUM（Banner優先testが`petStatus="生存"`で競合分岐を実証していない）を受理し、`petStatus="死亡"`へ修正。修正後のscoped Vitest → `Test Files  4 passed (4)` / `Tests  26 passed (26)` / `FE12_U3_POST_REVIEW_VITEST_EXIT=0`、ESLint → `FE12_U3_POST_REVIEW_ESLINT_EXIT=0`、design-audit → raw 10 / violations 0 / `FE12_U3_POST_REVIEW_DESIGN_AUDIT_EXIT=0`。
- Independent Review: planner/TDD passを実装前に統合。security pre-reviewの「OwnersList薄いDTOで正常死亡も偽警告」というHIGHは、同routeのPet modalが到達不能でlive editはdetail DTOを使うsource traceによりreject。「unknown statusも拒否」はpromptが明示するoptional prop互換と唯一caller中継に反するためreject。backend専用死亡APIが既死亡状態をConflict拒否せず再上書き可能というMEDIUMは本unit外の残余riskとしてaccept。最終healthcare reviewerはU3 APPROVE。React reviewerはproduction CRITICAL/HIGH 0、MEDIUM test指摘の修正後follow-upでactionable finding 0 / Approve。general reviewerはproduction actionable finding 0で初回ledger未記録だけをBlockし、本entry追記後follow-upでCRITICAL/HIGH/MEDIUM/LOW actionable finding 0 / Approve。
- Binding pre-read: `AGENTS.md`、`~/.agents/codex/AGENTS.md`、`.codex/config.toml`、`.claude/CLAUDE.md`、`.claude/rules/claude-code-usage.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/refs/typescript-react.md`、`.claude/refs/accessibility-rules.md`、`frontend/CLAUDE.md`、`frontend/src/features/CLAUDE.md`、`frontend/src/components/shared/CLAUDE.md`、`docs/product-philosophy.md`、`docs/spec/design-system.md` §2.4、`DESIGN.md`、`FE-refactor.md` FE12-02 C6a/F5/F17/U3/U1/U2、component/Banner/caller/tests、`tdd-workflow`、`react-testing`、`verification-loop`、`security-review`。binding矛盾なし。
- Harness / orchestration: tdd + sequential（code/test変更時De-Sloppify overlay）を実運用。planner、TDD、securityのread-only passを統合し、main agentがRED→GREEN→reconciliation、general/React/healthcareの3 fresh reviewerがIndependent Reviewを担当。stop condition=全Acceptance Checklist PASS。long-lived loopではないためloop monitoring非該当。
- Coverage / full gates: executable正本はfrontend statements 43.78%、tolerance 0.5ppのCI ratchetで、`vite.config.ts`にcoverage thresholdなし。coverage artifact生成はallowlist外書込みとなり、full `test:coverage` / `type-check` / `test:run` / `build`はprompt禁止のため未実行。CIまたはユーザー手動で`docker compose exec frontend pnpm type-check`、`docker compose exec frontend pnpm test:run`、`docker compose exec frontend pnpm build`を実行する。
- Prompt-defect / Harness Improvement Feedback: none needed。promptはvalidator PASSで、allowlist帰属・TDD・bounded retry・review/reconciliationを実運用できた。
- Remaining scope / risk: U4〜U10、M-05 responsive browser実測、backend/API/generated型/death token変更は未着手。backend死亡APIの既死亡再登録Conflict guardは別unit候補。stage/commit/stash/push 0。

### U4/F7+F8 実行 ledger（2026-07-24）

- Status: COMPLETE — `isPastJSTDate` のJST date-only厳密過去判定、VaccinationCard当日境界修正、VaccinationList期限超過表示、TDD、scoped regression、scoped lint、design-audit、De-Sloppify、Independent Review、allowlist照合を完了。git stage/commit/stash/pushは0。
- Product gate: 要件責任者=曽我、目的=再接種スケジューリングの臨床判断材料である期限表示を、JST当日は誤警告せず、過去日は一覧で例外抽出可能にすること。新規画面・列・入力・確認dialog・自動化・design tokenは追加していない。
- Changed files: `frontend/src/lib/jst-date.ts`、`frontend/src/lib/jst-date.test.ts`、`frontend/src/features/vaccinations/components/VaccinationCard.tsx`、`frontend/src/features/vaccinations/components/VaccinationCard.test.tsx`、`frontend/src/features/vaccinations/routes/VaccinationList.tsx`、`frontend/src/features/vaccinations/routes/VaccinationList.test.tsx`、本ledger。
- Root cause / minimal patch: `VaccinationCard.tsx` の `new Date(nextDate) < new Date()` がdate-only文字列を現在時刻と比較し、JST当日の午後を期限超過化し得た。共有 `jst-date.ts` に形式guard後の `date < todayJSTISO()` を使う `isPastJSTDate` を追加し、card/listの両方を同helperへ統一した。listは過去日のcellだけ既存 `C.danger`、`AlertTriangle`、日付、「（期限超過）」を表示し、未来日・空欄の既存cell表現を維持した。
- Saved Prompt Validation Gate: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-u4-vaccination-overdue.md` → `Prompt Craft Harness Validation: PASS`、`Profile: standard (declared-risk-tier)`、`Target: agent (detected)`、exit 0。
- 開始baseline: `git status --porcelain | sort > "${TMPDIR:-/tmp}/fe12-u4-git-baseline.txt"` → baselineは `BE-refactor.md`、`backend/cmd/csv-import/main_test.go`、`frontend/e2e/fixtures/ui-design-clinical.ts`、`frontend/src/features/medical-records/components/TreatmentsTab/treatments-tab-model.ts`、`backend/cmd/migrate/sql_migrations_integration_test.go` の5並行WIP path。indexは空。`docker compose exec frontend npx vitest run src/lib/jst-date.test.ts src/features/vaccinations` → `Test Files  7 passed (7)`、`Tests  72 passed (72)`、`baseline_vitest_exit=0`。出力は `${TMPDIR:-/tmp}/fe12-u4-vitest-baseline.txt` に保存。
- TDD RED: `docker compose exec frontend npx vitest run src/lib/jst-date.test.ts src/features/vaccinations/components/VaccinationCard.test.tsx src/features/vaccinations/routes/VaccinationList.test.tsx` → `Test Files  2 failed | 1 passed (3)`、`Tests  6 failed | 18 passed (24)`、`red_exit=1`。helper未実装の `TypeError: isPastJSTDate is not a function` と、過去日cellが `text-[#000000]` のままというF7欠落を確認。card当日RED fixtureは旧UTC parseでは再現しなかったため、JST当日午後 `2026-07-24T06:00:00Z` に補正し、同じTDD gateで修正を検証した。
- TDD GREEN: 同scoped command → `Test Files  3 passed (3)`、`Tests  24 passed (24)`、`green_exit=0`。
- Final scoped regression: `docker compose exec frontend npx vitest run src/lib/jst-date.test.ts src/features/vaccinations` → `Test Files  8 passed (8)`、`Tests  84 passed (84)`、`final_scoped_vitest_exit=0`。baseline比の新規失敗0、新規helper 5 case、card 4 case、list 3 caseが全PASS。
- Scoped lint: `docker compose exec frontend npx eslint src/lib/jst-date.ts src/lib/jst-date.test.ts src/features/vaccinations/components/VaccinationCard.tsx src/features/vaccinations/components/VaccinationCard.test.tsx src/features/vaccinations/routes/VaccinationList.tsx src/features/vaccinations/routes/VaccinationList.test.tsx --max-warnings 0` → ESLint出力なし（Composeの未設定DB env warningのみ）、`scoped_eslint_exit=0`。
- Design regression: `docker compose exec frontend pnpm design-audit` → `C18 raw legacy baseline — 10 件（non-gating ratchet）`、`C18 table cell override — 0 件`、`C19 table row onClick — 0 件`、`PASS — 違反 0 件`。追加hex/inline style/table-cell override/row onClickなし。既存FE12 ledgerのraw baseline 10件から増加なし。
- Source/diff contract: `rg -n "new Date\\(" frontend/src/features/vaccinations/components/VaccinationCard.tsx` → no match、`card_date_parse_rg_exit=1`。判定実装は `VaccinationCard.tsx` と `VaccinationList.tsx` の `isPastJSTDate`呼出しだけで、`jst-date.ts`のhelperは `/^\\d{4}-\\d{2}-\\d{2}$/` guardと `date < todayJSTISO()` のみ。`git diff --check` → stdout 0行、`diff_check_exit=0`。
- Allowlist: 完了時 `git status --porcelain | sort | diff "${TMPDIR:-/tmp}/fe12-u4-git-baseline.txt" -` は並行WIPの状態変化と本unitの7許可pathだけを表示。並行WIPは編集・revert・stageしていない。本unit所有のtargeted statusは上記6 TS/TSX path + `FE-refactor.md` の7path。
- Failure Signature log: attempt 1 / helper implementation — expected=`isPastJSTDate`が5境界を返す、actual=未定義関数、verification=TDD RED、error=`TypeError: isPastJSTDate is not a function`、fix=共有helper追加、result=GREEN。attempt 1 / F7 list — expected=過去cellがdanger+icon+文言、actual=`text-[#000000]`のみ、verification=TDD RED、error=`Expected ... text-[#C0392B] / Received ... text-[#000000]`、fix=既存danger表現をcell内へ追加、result=GREEN。同一failure signature再発なし。
- Runtime contract / Assumption: `nextDate`は `api/transforms.ts:13` の `"YYYY-MM-DD"` または空文字契約。形式guardはprompt裁定どおり軽量regexのみで、実在日付の追加parseや新規tokenを導入していない。一覧の超過表現は列を増やさず、既存card表現をcell内横並びで再現した。
- De-Sloppify: 新規テストはhelperの境界5種、cardの当日/過去/未来/空、listの過去/未来/空というproduct behaviorだけを保持。重複helper呼出しは `renderRow` 内の `overdue` へ集約。console、commented-out code、catch-all、drive-by cleanup、raw color、onClick、`data-c18-structural-cell`の追加なし。cleanup後scoped Vitest、ESLint、design-audit、diff-checkはPASS。
- Independent Review: Codexのsubagent/Task primitiveはこのruntimeで利用できないため、main agentがfresh read-only reviewを実施。観点は当日境界の `<`、JST/UTC境界、未来/空欄cell無退行、card/listのhelper共有、実時刻依存テスト、allowlist逸脱。CRITICAL/HIGH/MEDIUM actionable finding 0、Approve。`onClick`検出は既存のカード/操作UIのみで、本unitのtable rowには追加していない。
- Binding pre-read: `AGENTS.md`、`~/.agents/codex/AGENTS.md`、`.codex/config.toml`（実体なし）、`.claude/CLAUDE.md`、`.claude/rules/claude-code-usage.md`、`.claude/rules/go-gin-backend-guidelines.md`、`frontend/CLAUDE.md`、`frontend/src/features/CLAUDE.md`、`.claude/refs/typescript-react.md`、`.claude/refs/accessibility-rules.md`、`docs/product-philosophy.md`、`FE-refactor.md` FE12-02 C6a/F7/F8/U1/U2/U3、対象production/test files、`tdd-workflow`、`react-testing`、`verification-loop`。binding矛盾なし。
- Harness / orchestration: `tdd` + `sequential`を実運用。`~/.agents/skills/tdd-workflow/SKILL.md`、`react-testing/SKILL.md`、`verification-loop/SKILL.md`を読了し、RED→GREEN→De-Sloppify→scoped regressionを実施。subagentは利用不可のため独立fresh manual reviewへfallback。stop condition=全Acceptance Checklist PASS。long-lived loopではないためloop monitoring非該当。
- Coverage / full gates: project正本はfrontend `43.78%` baseline + tolerance `0.5pp` ratchetであり、coverage artifact生成はallowlist外書込みになるため本runの `test:coverage` は未実行。prompt禁止の全体 `docker compose exec frontend pnpm type-check`、`docker compose exec frontend pnpm test:run`、`docker compose exec frontend pnpm build` はユーザー/CI手動gateとして残す。scoped behavior/lint/design gatesは上記のとおりPASS。
- Prompt-defect / Harness Improvement Feedback: P2 — promptはcard当日fixture例として `2026-07-23T15:30:00Z` を示すが、現行ECMAScript date-only UTC parseでは旧bugを再現しないため、TDD card fixtureをJST当日午後 `2026-07-24T06:00:00Z`へ補正した。validator/contractの欠陥ではなく、生成側が「JST当日」を再現する時刻をより明示すべきという改善メモ。その他はnone needed。
- Remaining scope / risk: U5〜U10、他featureの日付比較、vaccination form/detail/hooks/API mutation、M-05 responsive browser実測、full type-check/test/build/coverageは未着手または手動gate。stage/commit/stash/push 0。

### U5/F6 実行 ledger（2026-07-24）

- Status: COMPLETE — LOW の臨床 semantic blue 置換、TDD RED→GREEN、scoped regression、scoped ESLint、design-audit、De-Sloppify、Independent Review、allowlist照合を完了。git stage/commit/stash/pushは0。
- Product gate: 要件責任者=曽我、目的=検査異常低値をCTA・リンクのbrand構造色から臨床semantic blueへ分離し、LOWを異常値として認知可能にすること。LOW文言、DOM、grid列、badge variant/size/shape、高値・正常・未判定分岐、新規token/componentは変更していない。
- Changed files: `frontend/src/features/examinations/components/ExamItemsTable.tsx`、`frontend/src/features/examinations/components/ExamItemsTable.test.tsx`、`frontend/src/features/medical-records/components/ExaminationGroup.tsx`、`frontend/src/features/medical-records/components/ExaminationGroup.test.tsx`、本ledger。
- Root cause / minimal patch: `ExamItemsTable.tsx:53,120` と `ExaminationGroup.tsx:76,100` のLOW分岐が `C.textBrand` / `C.borderBrand` / `C.bgBrand5` / `C.bgBrandLight8` を使っていたため、既存 `C.textStatusBlue` / `C.borderBlue400` / `C.bgStatusBlueLight` へ4箇所だけ置換した。HIGHの`C.danger`/`C.bgDanger*`、normalの`C.textStatusGreen`、pendingの`C.text45`は維持した。
- Saved Prompt Validation Gate: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-u5-low-clinical-blue.md` → `Prompt Craft Harness Validation: PASS`、`Profile: standard (declared-risk-tier)`、`Target: agent (detected)`、exit 0。
- Binding pre-read: `AGENTS.md`、`~/.agents/codex/AGENTS.md`、`.codex/config.toml`、`.claude/CLAUDE.md`、`.claude/rules/claude-code-usage.md`、`.claude/rules/go-gin-backend-guidelines.md`、`frontend/CLAUDE.md`、`frontend/src/features/CLAUDE.md`、`docs/product-philosophy.md`、`docs/spec/design-system.md` §2.4、`FE-refactor.md` FE12-02 C6a/F6/U1〜U4、`frontend/src/lib/design-tokens.ts:326-352`、`.claude/refs/typescript-react.md`、`.claude/refs/accessibility-rules.md`、対象production/test files、`/Users/minoru/.agents/skills/tdd-workflow/SKILL.md`、`/Users/minoru/.agents/skill-library/react-testing/SKILL.md`、`/Users/minoru/.agents/skills/verification-loop/SKILL.md`。binding矛盾なし。
- Start baselines: `git status --porcelain | sort > "${TMPDIR:-/tmp}/fe12-u5-git-baseline.txt"` → ` M backend/cmd/csv-import/main_test.go`、` M frontend/e2e/fixtures/ui-design-clinical.ts`、` M frontend/src/features/medical-records/components/TreatmentsTab/treatments-tab-model.ts`、`?? backend/cmd/migrate/sql_migrations_integration_test.go`、exit 0。`docker compose exec frontend npx vitest run src/features/examinations/components/ExamItemsTable.test.tsx src/features/medical-records/components/ExaminationGroup.test.tsx` → `Test Files  2 passed (2)`、`Tests  36 passed (36)`、`FE12_U5_VITEST_BASELINE_EXIT=0`; output saved to `${TMPDIR:-/tmp}/fe12-u5-vitest-baseline.txt`.
- TDD RED: the same scoped Vitest command after test assertions → `Test Files  2 failed (2)`、`Tests  2 failed | 34 passed (36)`、exit 1. Exact failures were the LOW badge expected `text-blue-700 border-blue-400 bg-blue-50` but received old brand classes in both named test files; HIGH assertions passed.
- TDD GREEN / final scoped regression: `docker compose exec frontend npx vitest run src/features/examinations/components/ExamItemsTable.test.tsx src/features/medical-records/components/ExaminationGroup.test.tsx` → `Test Files  2 passed (2)`、`Tests  36 passed (36)`、exit 0. Baseline比の新規失敗0、追加LOW/HIGH/pending assertionsを含む全36件PASS。
- Source contract: `rg -n "textBrand|borderBrand|bgBrand5|bgBrandLight8" frontend/src/features/examinations/components/ExamItemsTable.tsx frontend/src/features/medical-records/components/ExaminationGroup.tsx` → hitなし、`FE12_U5_BRAND_REMAINING_EXIT=1`（rgの0-hit exit）。`rg` source review confirms `ExamItemsTable.tsx:53`=`C.textStatusBlue C.borderBlue400 C.bgStatusBlueLight`、`:119`=`C.bgDanger8`、`:120`=`C.bgStatusBlueLight`; `ExaminationGroup.tsx:74`=`C.danger`、`:76`=`C.textStatusBlue`、`:100`=`C.textStatusBlue C.borderBlue400 C.bgStatusBlueLight`、`:106`=`C.textStatusGreen`。
- Diff/whitespace gate: target `git diff --stat` and `git diff -w --stat` both report `4 files changed, 25 insertions(+), 10 deletions(-)`; `git diff --check` → stdout 0行, exit 0. Unified diff contains only the four production token substitutions plus adjacent test token assertions; `frontend/src/features/reservations/components/WeekViewDayColumn.tsx` remains an existing `bgBrandLight8` reservation consumer and was not changed; `frontend/src/lib/design-tokens.ts` was not changed.
- Scoped lint: first repository-relative container invocation failed with exact `No files matching the pattern "frontend/src/features/examinations/components/ExamItemsTable.tsx" were found.` / exit 2. Failure Signature attempt 1 was resolved by using container-relative paths. `docker compose exec frontend npx eslint src/features/examinations/components/ExamItemsTable.tsx src/features/examinations/components/ExamItemsTable.test.tsx src/features/medical-records/components/ExaminationGroup.tsx src/features/medical-records/components/ExaminationGroup.test.tsx --max-warnings 0` → ESLint diagnostics 0行、exit 0。
- Design regression: `docker compose exec frontend pnpm design-audit` → `C1/C3/C5〜C19 gating項目 0 件`、`C18 raw legacy baseline — 10 件（non-gating ratchet）`、`design-system-audit: PASS — 違反 0 件`、exit 0. Counts are unchanged from the run baseline and no new audit surface was introduced.
- Allowlist/index: before ledger append, `git status --porcelain | sort | diff "${TMPDIR:-/tmp}/fe12-u5-git-baseline.txt" -` showed only the four U5 target paths plus concurrently changing external `BE-refactor.md`; the baseline WIP paths were not edited by this unit. `git diff --cached --name-only` produced no paths and no git add/commit/stash/push was executed. Final post-ledger status and target-path ownership were rechecked after this append.
- Final reconciliation evidence: repeated scoped Vitest → `Test Files  2 passed (2)`, `Tests  36 passed (36)`, exit 0; repeated scoped ESLint → no diagnostics, exit 0; final brand `rg` → no hits, `FE12_U5_FINAL_BRAND_REMAINING_EXIT=1`; repeated design-audit → `C18 raw legacy baseline — 10 件（non-gating ratchet）`, `C18 table cell override — 0 件`, `C19 table row onClick — 0 件`, `PASS — 違反 0 件`, exit 0. Final target `git diff --stat` and `git diff -w --stat` both → `4 files changed, 25 insertions(+), 10 deletions(-)`; final status delta snapshot → `M BE-refactor.md` and `?? q&a.html` (concurrent external WIP), `M FE-refactor.md`, the four U5 TSX paths; `FE12_U5_ALLOWLIST_DIFF_EXIT=1` because the raw global delta includes those external paths; `git diff --cached --name-only` empty and `FE12_U5_INDEX_PATHS=0`. U5-owned delta is allowlist-complete; external WIP was not edited, reverted, staged, or committed.
- Harness / orchestration: `tdd` + `sequential` (De-Sloppify overlay)実運用。`/Users/minoru/.agents/skills/tdd-workflow/SKILL.md`、`/Users/minoru/.agents/skill-library/react-testing/SKILL.md`、`/Users/minoru/.agents/skills/verification-loop/SKILL.md`を読了し、test-first RED→GREEN→reconciliationを実施。subagent/Task primitiveはこのruntimeで利用不可のため、独立fresh manual reviewへfallback。review観点はLOW/HIGH取り違え、スコープ外brand消費者への波及、whitespace、allowlist、臨床誤認で、CRITICAL/HIGH/MEDIUM actionable finding 0、Approve。
- Failure Signature log: attempt 1 / scoped ESLint path — expected=changed TSX files linted in Docker, actual=container `/app` could not resolve repository-relative `frontend/...`, verification=scoped ESLint, error=`No files matching the pattern "frontend/src/features/examinations/components/ExamItemsTable.tsx" were found.` exit 2, fix=rerun with `src/...` container-relative paths, result=exit 0. Same failure signature did not recur.
- De-Sloppify: retained only product-facing LOW blue, HIGH danger, pending token assertions and the four required substitutions; no console, commented-out code, broad catch, new helper/component, whitespace cleanup, or drive-by changes were introduced. Post-cleanup scoped Vitest, ESLint, design-audit, diff-check all PASS.
- Coverage / full gates: project policy is frontend `43.78%` statement baseline with `0.5pp` ratchet; no project-enforced fixed 80% threshold was found in `vite.config.ts`/coverage policy. Coverage artifact generation and prohibited full `docker compose exec frontend pnpm type-check`, `docker compose exec frontend pnpm test:run`, `docker compose exec frontend pnpm build` were not run; they remain user/CI manual follow-ups.
- Prompt-defect / Harness Improvement Feedback: none needed. The prompt validator and required reconciliation/review gates operated as specified; the only command correction was the project-documented `/app` Docker path resolution and is recorded above.
- Remaining scope / risk: U6〜U10、F1〜F5/F7〜F18、M-05 responsive browser実測、full type-check/test/build/coverageは未着手または手動gate。stage/commit/stash/push 0。

### U6/F11+F12 実行 ledger（2026-07-24）

- Status: COMPLETE — medical-records mutation境界のfail-closed guard、TDD、scoped regression、scoped ESLint、design-audit、De-Sloppify、Independent Review、allowlist帰属を完了。git stage/commit/stash/pushは0。
- Product gate: 要件責任者=曽我、目的=非編集personaのprogrammatic/race経路からカルテ一次記録のquick patch・owner変更・deleteを物理的に遮断する臨床安全defense-in-depth。新規画面・入力・確認dialog・permission hook・backend/API/generated型・design tokenは追加していない。
- Root cause / minimal patch: quick patch 5 handlerとowner変更handlerがmutation直前のpermissionを持たず、MedicalRecordFormのdelete callbackとfieldsetもUI状態に依存していた。各hookにstrict positive match `canEdit === true` の共通helperを1つだけ置き、deleteには`canDelete === true` guard、fieldsetには`isFinalized || !canSubmit`、owner controlには既存のread-only spanパターンを適用した。
- Changed files: `frontend/src/features/medical-records/hooks/use-medical-record-quick-patch-actions.ts`、同`use-medical-record-quick-patch-actions.test.ts`、`frontend/src/features/medical-records/hooks/use-medical-record-owner-change.ts`、同`use-medical-record-owner-change.test.ts`、`frontend/src/features/medical-records/routes/MedicalRecordForm.tsx`、同`MedicalRecordForm.permissions.test.tsx`、`frontend/src/features/medical-records/components/MedicalRecordFormPanels.tsx`、本ledger。`clinical-care-routes.tsx`は変更なし。
- Saved Prompt Validation Gate: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-u6-medical-record-mutation-guards.md` → `Prompt Craft Harness Validation: PASS`、`Profile: standard (declared-risk-tier)`、`Target: agent (detected)`、exit 0。SHA-256=`1d465f1a4633ea7991e726eac5056e74e41ff0d192257892d41b418234850ae4`。
- Binding pre-read: `AGENTS.md`、`~/.agents/codex/AGENTS.md`、`.codex/config.toml`、`.claude/CLAUDE.md`、`.claude/rules/claude-code-usage.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/refs/typescript-react.md`、`.claude/refs/accessibility-rules.md`、`frontend/CLAUDE.md`、`frontend/src/features/CLAUDE.md`、`docs/product-philosophy.md`、本節FE12-02 C6a/F11/F12/U1〜U5、save-action guard、OwnerForm permissions precedent、clinical-care route、対象production/test files、`tdd-workflow`、`react-testing`、`verification-loop`、`security-review`。binding矛盾なし。
- 開始baseline: `git status --porcelain | sort > "${TMPDIR:-/tmp}/fe12-u6-git-baseline.txt"` exit 0。`docker compose exec frontend npx vitest run src/features/medical-records` → `Test Files  1 failed | 29 passed (30)`、`Tests  1 failed | 233 passed (234)`、exit 1。既存 failureは`use-medical-record-form.auto-create.test.ts`のvisitDate appointment timestamp assertionであり、U6変更前から存在。出力は`${TMPDIR:-/tmp}/fe12-u6-vitest-baseline.txt`に保存。
- TDD RED: `docker compose exec frontend npx vitest run src/features/medical-records/hooks/use-medical-record-quick-patch-actions.test.ts src/features/medical-records/hooks/use-medical-record-owner-change.test.ts src/features/medical-records/routes/MedicalRecordForm.permissions.test.tsx` → `Test Files  3 failed (3)`、`Tests  3 failed | 2 passed (5)`、exit 1。権限なしquick patch 5件、owner mutation、fieldsetで旧fail-open/未無効化を確認。初回実行はfrontend依存起動途中の`Cannot find package 'jsdom'` 3 worker errorとなったため、依存インストールは行わずcontainer起動完了後に本REDを再実行した。
- TDD GREEN: 同scoped command → `Test Files  3 passed (3)`、`Tests  5 passed (5)`、exit 0。新規ケースは権限なしmutation不呼出、権限あり従来payload、delete拒否、fieldset非活性を被覆。
- Final scoped regression: `docker compose exec frontend npx vitest run src/features/medical-records` → `Test Files  1 failed | 32 passed (33)`、`Tests  1 failed | 238 passed (239)`、exit 1。新規失敗0、U6 5 cases PASS。残る1 failureはbaselineと同じauto-create visitDate timestamp assertion。
- Scoped ESLint: `docker compose exec frontend npx eslint src/features/medical-records/hooks/use-medical-record-quick-patch-actions.ts src/features/medical-records/hooks/use-medical-record-quick-patch-actions.test.ts src/features/medical-records/hooks/use-medical-record-owner-change.ts src/features/medical-records/hooks/use-medical-record-owner-change.test.ts src/features/medical-records/routes/MedicalRecordForm.tsx src/features/medical-records/routes/MedicalRecordForm.permissions.test.tsx src/features/medical-records/components/MedicalRecordFormPanels.tsx --max-warnings 0` → diagnostics 0、exit 0。
- Design regression: `docker compose exec frontend pnpm design-audit` → C1/C3/C5〜C19 gating各0件、`C18 raw legacy baseline — 10 件（non-gating ratchet）`、`C18 table cell override — 0 件`、`C19 table row onClick — 0 件`、`design-system-audit: PASS — 違反 0 件`、exit 0。開始時から件数不変。
- Route guard decision: `rg`でmedical-record detail routeは`clinical-care-routes.tsx:17-67`の`/medical-records/:id`のみ、`paths.ts:59-70`にも同一detail pathのみを確認。別read-only detail routeは実在しないため、`:id`へ`action="edit"` guardは適用しなかった。既存resource view guardとcomponent/mutation boundaryでread accessを維持する。この不適用判断は本ledgerのdeviationとして記録する。
- Whitespace/diff: 対象production 4 filesの`git diff --stat`と`git diff -w --stat`はともに`4 files changed, 31 insertions(+), 17 deletions(-)`。`git diff --check`はstdout 0行、exit 0。TreatmentsTab配下は閲覧以外の接触なし。
- Allowlist: 本unit-owned deltaは上記8 code/test path + `FE-refactor.md`に限定。開始時から存在するbackend/docs/frontend fixture/WIPは編集・revert・stageしていない。allowlist比較はledger追記後の最終照合で実施し、並行sessionの外部deltaは帰属から除外する。`git diff --cached --name-only`は空。
- Failure Signature log: attempt 1 / RED harness — expected=local Vitest dependency resolution、actual=`Cannot find package 'jsdom'` / ephemeral npx Vitest 4.1.10、verification=RED command、error=3 worker startup errors、fix=依存installなしでfrontend container起動完了後に同じcommandを再実行、result=意図したREDへ到達。同一signature再発なし。pre-existing regressionはbaseline/final共通のためU6 failureとは分類しない。
- Runtime contract / Assumption deviation: allowlistに実際のcaller `use-medical-record-form.ts`が含まれないため変更せず、両hookの`canEdit?: boolean` explicit override + 既存`usePermission("medical-records")` fallbackで配線した。新規permission hookは作成していない。新規recordのcreate semanticsとbackend RBACは既存契約を維持し、U7〜U10の類似問題は未変更。
- Harness / orchestration: `tdd` + `sequential`を実運用。指定backing skillsを読了し、RED→GREEN→scoped regression→De-Sloppify→reconciliationを実施。Task/subagent primitiveはこのruntimeで利用できないため、main agentのfresh manual planning/security/React reviewへfallback。stop condition=全U6 checklist itemがPASSまたは既知baseline/ledger deviationとして記録済み。
- De-Sloppify: product behaviorに直結する5新規testのみ保持。helperはhookごとに1つ、console/log、catch-all、dialog/toast、raw color、any、drive-by cleanup、TreatmentsTab変更なし。post-cleanup scoped regression/ESLint/design-audit/diff-checkは上記結果。
- Independent Review: fresh manual security/React reviewでfail-open分岐、権限ありpayload/成功後処理、delete/fieldset、owner read-only control、route read-access遮断、helper重複、TreatmentsTab接触、whitespace、allowlistを再照合。CRITICAL/HIGH/MEDIUM actionable finding 0、Approve。subagent未使用理由は上記Harness欄と同じ。
- Coverage / full gates: `docs/ops/coverage-policy.md`のfrontend baselineは43.78%、tolerance 0.5pp ratchet。coverage artifactを生成する`test:coverage`と全体`docker compose exec frontend pnpm type-check`、同`pnpm test:run`、同`pnpm build`はprompt/CLAUDE禁止のため未実行し、CI/ユーザー手動gateとして残す。
- Prompt-defect / Harness Improvement Feedback: P2 — allowlistがhook callerを除外しているため、明示引数配線をcaller変更なしで成立させるfallback deviationが必要だった。今後はprompt生成時にcaller pathをallowlistへ含めるか、既存permission hook fallbackを明示する。その他none needed。
- Remaining scope / risk: `clinical-care-routes.tsx` route guard変更、U7〜U10、TreatmentsTab、backend/API/generated型、full type-check/test/build/coverage、stage/commit/stash/pushは未実施。backend RBACが最終防壁であり本unitはP1 defense-in-depth。
- **Mode 3 照合修正（2026-07-25・生成側セッション適用）**: executor実装は`handleVisitTypeChange`/`handleNextVisitDatePatch`のguardを先頭へ置いたため、「新規作成時(recordIdなし)はローカルstateのみ更新」の文書化済み既存契約(コメント明記)を破壊するHIGH regressionを含んでいた。照合側で2 handlerを「permission guard→ローカルstate更新→recordId無ならreturn」へ復元し、quick-patch-actions.test.tsへ新規作成経路の固定ケースを追補(修正後 medical-records suite 240/240・scoped ESLint 0・design-audit PASS)。付記: executor報告の「既存失敗1件(auto-create visit-date timestamp)」は日付が7/25へ変わった時点で消失=日付境界依存のflaky test(別途follow-up候補)。allowlist欠落(use-medical-record-form.ts)起因のdeviation(hook内usePermission fallback+明示override)は自己検査型component境界としてU3と整合するため受理。

### U7/F13+F18 実行 ledger（2026-07-25）

- Status: COMPLETE — hospitalization form save/delete と care plan / daily record / vital / care log / staff note の mutation 境界を positive permission match で fail-closed 化。TDD、scoped regression、scoped ESLint、design-audit、De-Sloppify、Independent Review、allowlist帰属を完了。git stage/commit/stash/pushは0。
- Product gate: 要件責任者=曽我、目的=入院診療の一次記録をUI guard迂回・権限剥奪raceから遮断する臨床安全defense-in-depth。新規permission hook、dialog/toast、backend/API/generated型、design token、U8〜U10/F14〜F17相当の変更は追加していない。
- Root cause / minimal patch: `use-hospitalization-form.ts`のcreate/update action、`HospitalizationForm.tsx`のdelete、`CarePlanTab.tsx`のcreate/update/delete、`DailyRecordsTab.tsx`の日次記録/vital/care log/staff noteがUI表示条件だけでmutationを再検査していなかった。formにはroute既存`canSubmit`を渡し、各componentには既存`usePermission("hospitalization")`値を使ったstrict positive match helper/refを追加した。既存validation・ID判定・ローカルstate/transition/payload順序は維持した。
- Changed files: `frontend/src/features/hospitalization/hooks/use-hospitalization-form.ts`、同`use-hospitalization-form.test.ts`、`frontend/src/features/hospitalization/routes/HospitalizationForm.tsx`、同`HospitalizationForm.permissions.test.tsx`、`frontend/src/features/hospitalization/components/CarePlanTab/CarePlanTab.tsx`、`frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx`、同`DailyRecordsTab.permissions.test.tsx`、本ledger。開始時から存在する5-session/backend/frontend medical-records WIPは変更していない。
- Saved Prompt Validation Gate: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-u7-hospitalization-mutation-guards.md` → `Prompt Craft Harness Validation: PASS`、`Profile: standard (declared-risk-tier)`、`Target: agent (detected)`、exit 0。SHA-256=`68d2dcd24096dc9632133b3f1a1b7c684412ed6dedf31a05d2e1b6c0ccea1363`。
- Binding pre-read: `AGENTS.md`、`~/.agents/codex/AGENTS.md`、`.codex/config.toml`、`.claude/CLAUDE.md`、`.claude/rules/claude-code-usage.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/refs/typescript-react.md`、`.claude/refs/accessibility-rules.md`、`frontend/CLAUDE.md`、`frontend/src/features/CLAUDE.md`、`docs/product-philosophy.md`、本節FE12-02 C6a/F13/F18/U7/U6 ledgerとMode 3照合修正、`git show 585e9eb20`、対象production/test files、`/Users/minoru/.agents/skills/tdd-workflow/SKILL.md`、`/Users/minoru/.agents/skills/react-testing/SKILL.md`、`/Users/minoru/.agents/skills/verification-loop/SKILL.md`、`/Users/minoru/.agents/skills/security-review/SKILL.md`。binding矛盾なし。
- Start baselines: `git status --porcelain | sort > "${TMPDIR:-/tmp}/fe12-u7-git-baseline.txt"` → ` D 5-session-agent.md`、` M backend/cmd/csv-import/main_test.go`、` M frontend/e2e/fixtures/ui-design-clinical.ts`、` M frontend/src/features/medical-records/components/TreatmentsTab/treatments-tab-model.ts`、`?? backend/cmd/migrate/sql_migrations_integration_test.go`、exit 0。`docker compose exec frontend npx vitest run src/features/hospitalization` → `Test Files  16 passed (16)`、`Tests  86 passed (86)`、exit 0。出力は`${TMPDIR:-/tmp}/fe12-u7-vitest-baseline.txt`に保存。
- TDD RED: `docker compose exec frontend npx vitest run src/features/hospitalization/hooks/use-hospitalization-form.test.ts src/features/hospitalization/routes/HospitalizationForm.permissions.test.tsx src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.permissions.test.tsx` → `Test Files  3 failed (3)`、`Tests  4 failed | 29 passed (33)`、exit 1。失敗はform permissionなし、delete revocation、child permission guardの未実装を示した。
- TDD GREEN / named regression: 同scoped command → `Test Files  3 passed (3)`、`Tests  35 passed (35)`、exit 0。`docker compose exec frontend npx vitest run src/features/hospitalization` → `Test Files  18 passed (18)`、`Tests  95 passed (95)`、exit 0。baseline比で新規失敗0、権限なし不発行・権限ありpayload・権限revocation・日次記録作成を被覆。
- Scoped ESLint: 初回 `docker compose exec frontend npx eslint <changed files> --max-warnings 0` は`Error: Cannot access refs during render` / `Cannot update ref during render`、2 errors、exit 1。Failure Signature attempt 1としてref同期をuseEffectへ変更。同じcontainer-relative changed-file commandの最終結果は diagnostics 0、`FE12_U7_FINAL_ESLINT_EXIT=0`。
- Design regression: `docker compose exec frontend pnpm design-audit` → `design-system-audit: C1 legacy accent — 0 件`、`C3 hex 直書き — 0 件`、`C5 非 brand colorVariant — 0 件`、`C6 rgba/hsla 直値 — 0 件`、`C7 maxWidth 生値 — 0 件`、`C8 PageLayout 未使用 routes — 0 件`、`C9 rounded 任意値 — 0 件`、`C10 shadow 既定/任意値 — 0 件`、`C11 font-size 任意値 — 0 件`、`C12 非仕様サイズ段(text-lg/2xl+) — 0 件`、`C13 ink 黒アルファ — 0 件`、`C14 tracking 直書き — 0 件`、`C15 route named color — 0 件`、`C16 非仕様 spacing(*-5) — 0 件`、`C17 CSS shadow 直書き — 0 件`、`C18 table cell override — 0 件`、`C18 raw legacy baseline — 10 件（non-gating ratchet）`、`C19 table row onClick — 0 件`、`design-system-audit: PASS — 違反 0 件`、exit 0。変更はguard/testのみでdesign surfaceを追加していない。
- Whitespace/diff: `git diff --check` → stdout 0行、exit 0。対象production各ファイルの`git diff --stat`と`git diff -w --stat`は一致し、`use-hospitalization-form.ts`=`1 file changed, 8 insertions(+), 1 deletion(-)`、`HospitalizationForm.tsx`=`1 file changed, 7 insertions(+), 3 deletions(-)`、`CarePlanTab.tsx`=`1 file changed, 18 insertions(+), 4 deletions(-)`、`DailyRecordsTab.tsx`=`1 file changed, 14 insertions(+), 5 deletions(-)`。
- Allowlist/index: pre-ledgerの必須コマンド`git status --porcelain | sort | diff "${TMPDIR:-/tmp}/fe12-u7-git-baseline.txt" -`は本unit追加差分を表示しexit 1。ただし表示された追加は上記allowlistの5 production/test pathsと許可された新規隣接test 2 pathsのみで、開始時WIPは不変・未編集。`git diff --cached --name-only`は出力なし、command exit 0。ledger追記後の最終allowlist照合で`FE-refactor.md`を含む本unit-owned pathsのみを再確認する。
- Harness / orchestration: `tdd` + `sequential`を実運用。指定4 skillsを読了し、RED→GREEN→scoped regression→De-Sloppify→reconciliationを実施。Task/subagent primitiveはこのruntimeで利用不可のため、main agentのfresh planning/security/React manual reviewへfallback。review観点はfail-open、既存分岐順序、第3象限、権限ありpayload、helper統一、whitespace、allowlistでCRITICAL/HIGH/MEDIUM actionable finding 0、Approve。
- Failure Signature log: attempt 1 / React refs lint — expected=changed files lint clean、actual=`Error: Cannot access refs during render` / `Cannot update ref during render`、verification=scoped ESLint、error=2 errors、fix=all permission refs synchronized in `useEffect` rather than during render、result=final scoped ESLint exit 0。同一signature再発なし。
- De-Sloppify: test files retain only product-facing permission denial, permission-granted payload, delete, and revocation cases; no console/log, `any`, broad catch, dialog/toast, style, or drive-by cleanup added. Final named Vitest, full hospitalization Vitest, ESLint, design-audit, and diff-check remained PASS.
- Coverage / full gates: `docs/ops/coverage-policy.md`のfrontend baselineは43.78%、tolerance 0.5pp ratchetであり本unitに固定80% fail thresholdはない。coverage artifactと禁止された全体`docker compose exec frontend pnpm type-check`、同`pnpm test:run`、同`pnpm build`は実行せず、CI/ユーザー手動follow-upとする。
- Prompt-defect / Harness Improvement Feedback: none needed。prompt validator、checklist、reconciliation、review、bounded retryが指定どおり機能した。
- Remaining scope / risk: U8〜U10、F14〜F17、backend RBAC、full type-check/test/build/coverage、stage/commit/stash/pushは未実施。backend RBACが最終防壁であり、本unitはP1 defense-in-depth。
- Final reconciliation: saved-prompt validator → `Prompt Craft Harness Validation: PASS` / exit 0; final design-audit → `C1/C3/C5〜C17/C18 table cell/C19 gating項目 0 件`、`C18 raw legacy baseline — 10 件（non-gating ratchet）`、`design-system-audit: PASS — 違反 0 件` / exit 0; required raw allowlist command → `FE12_U7_ALLOWLIST_RAW_DIFF_EXIT=1` because unit-owned additions are expected against the start snapshot; explicit unit delta → `FE12_U7_UNIT_DELTA_ALLOWLIST_EXIT=0`; cached path count → `FE12_U7_CACHED_PATH_COUNT=0`; `git diff --check` → `FE12_U7_DIFF_CHECK_EXIT=0`.

### U8/F14+F15 実行 ledger（2026-07-25）

- Status: COMPLETE — vaccination / examination の create・edit・delete mutation 境界を deny-all default と action 別 positive match で fail-closed 化。TDD、scoped regression、scoped ESLint、design-audit、De-Sloppify、Independent Review、allowlist 帰属を完了。git stage/commit/stash/push は0。
- Product gate: 要件責任者=曽我、目的=権限なし staff と権限剥奪 race から予防接種・検査一次記録の mutation を遮断する臨床安全 defense-in-depth。新規画面・入力・dialog/toast・permission hook・backend/API/generated 型・design token は追加していない。
- Root cause / minimal patch: 両 Form は UI の fieldset/button だけで権限を反映し、hook の create/update/delete は mutation 直前に権限を再検査していなかった。既存 `usePermission` の `canCreate/canEdit/canDelete` を各 hook へ渡し、既定値を全 false、ref を primitive dependency の `useEffect` で同期、`permissionsRef.current[action] === true` の単一 helper で action 別に guard した。検査 items は parent guard の後方に維持し、items 側へ重複 guard を追加していない。
- Changed files: `frontend/src/features/vaccinations/hooks/use-vaccination-form.ts`、同 `use-vaccination-form.test.ts`、`frontend/src/features/vaccinations/routes/VaccinationForm.tsx`、同 `VaccinationForm.permissions.test.tsx`、`frontend/src/features/examinations/hooks/use-examination-form.ts`、同 `use-examination-form.test.ts`、`frontend/src/features/examinations/routes/ExaminationForm.tsx`、同 `ExaminationForm.permissions.test.tsx`、本 ledger。`frontend/src/app/routes/clinical-care-routes.tsx` は変更なし。
- Saved Prompt Validation Gate: `node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-u8-vaccination-examination-mutation-guards.md` → `Prompt Craft Harness Validation: PASS`、exit 0。SHA-256=`e1e4441a4b036670e26adc8f12bdffad00e442c1b0781491a2b6ca3086b30730`。
- Binding pre-read: `AGENTS.md`、`~/.agents/codex/AGENTS.md`、`.codex/config.toml`、`.claude/CLAUDE.md`、`.claude/rules/claude-code-usage.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/refs/typescript-react.md`、`frontend/CLAUDE.md`、`frontend/src/features/CLAUDE.md`、`docs/product-philosophy.md`、`docs/ops/coverage-policy.md`、本節 FE12-02 C6a/F14/F15/U8/U6/U7/Mode 3、`git show 4cb08d286`、対象 hook/Form/test/route 全文、`tdd-workflow`、`react-testing`、`verification-loop`、`security-review`。binding 矛盾なし。
- Start baselines: 編集前 `git status --porcelain | sort > "${TMPDIR:-/tmp}/fe12-u8-git-baseline.txt"` は `5-session-agent.md`、backend CSV import test、frontend clinical fixture、medical-record TreatmentsTab model、`q&a.html`、`todo.md`、新規 migration integration test の7外部 WIP path。allowlist target は全て clean、index は空。`docker compose exec frontend npx vitest run src/features/vaccinations src/features/examinations` → `Test Files  13 passed (13)`、`Tests  180 passed (180)`、exit 0、出力は `${TMPDIR:-/tmp}/fe12-u8-vitest-baseline.txt`。開始時 `docker compose exec frontend pnpm design-audit` → gating 各0件、`C18 raw legacy baseline — 10 件（non-gating ratchet）`、`design-system-audit: PASS — 違反 0 件`、exit 0。
- TDD RED: denied create/edit/delete の初回6 case は `Test Files  2 failed (2)` / `Tests  6 failed | 62 passed (68)` / exit 1。権限剥奪後の captured action 2 case 追加後は `Test Files  2 failed (2)` / `Tests  8 failed | 62 passed (70)` / exit 1。旧実装が deny persona でも従来 payload の create/update/delete と検査 items を発行する F14/F15 を再現した。
- TDD GREEN / named regression: 両 hook spec は `Test Files  2 passed (2)` / `Tests  70 passed (70)` / exit 0。hook 2 spec + Form permission wiring 2 spec の最終 named gate は `Test Files  4 passed (4)` / `Tests  72 passed (72)` / exit 0。権限なし create/edit/delete、権限あり従来 payload、第3象限、captured action の true→false revocation、検査 parent-before-items 順序を被覆。
- Final scoped regression: `docker compose exec frontend npx vitest run src/features/vaccinations src/features/examinations` → `Test Files  15 passed (15)`、`Tests  190 passed (190)`、exit 0、baseline 比の新規 failure 0。検査 hook の `useActionState was called outside of a transition` と deliberate `Non-Axios Error: API error` stderr は開始 baseline と最終結果に共通する既知 test harness 出力で、production の `<form action={formAction}>` 経路の退行ではない。
- Scoped ESLint: `docker compose exec frontend npx eslint src/features/vaccinations/hooks/use-vaccination-form.ts src/features/vaccinations/hooks/use-vaccination-form.test.ts src/features/vaccinations/routes/VaccinationForm.tsx src/features/vaccinations/routes/VaccinationForm.permissions.test.tsx src/features/examinations/hooks/use-examination-form.ts src/features/examinations/hooks/use-examination-form.test.ts src/features/examinations/routes/ExaminationForm.tsx src/features/examinations/routes/ExaminationForm.permissions.test.tsx --max-warnings 0` → diagnostics 0、exit 0（Compose の未設定 DB env warning のみ）。
- Design regression: `docker compose exec frontend pnpm design-audit` → C1/C3/C5〜C17/C18 table cell/C19 gating 各0件、`C18 raw legacy baseline — 10 件（non-gating ratchet）`、`design-system-audit: PASS — 違反 0 件`、exit 0。開始時から全件数不変。
- Route guard decision: `clinical-care-routes.tsx:201-253` と `:257-309` は各 resource view guard 配下の単一 Form/detail route で、一覧 detail link も同じ `:id` へ遷移する。別 read-only detail route は実在しないため `action="edit"` guard を適用せず、閲覧権限 staff の read を維持した。mutation は hook 境界、最終認可は backend RBAC が担う。
- Third-quadrant / ordering contract: vaccination は `canCreate=true/canEdit=false/idなし` で exact legacy create payload、update 0回を固定。examination は同じ第3象限で exact parent payload と exact items payload、update parent 0回、`createMutate.mock.invocationCallOrder[0] < updateItemsMutate.mock.invocationCallOrder[0]` を固定。validation、pet/ID early return、parent→items の既存順序を保持した。
- Whitespace/diff: 対象 production 4 files の `git diff --stat` と `git diff -w --stat` はともに `4 files changed, 76 insertions(+), 11 deletions(-)`。`git diff --check` は stdout 0行、exit 0。追加 production diff の console / `any` / raw HTML / raw hex / inline style / TODO / FIXME / broad catch 検索は match 0。
- Allowlist/index: required raw comparison `git status --porcelain | sort | diff "${TMPDIR:-/tmp}/fe12-u8-git-baseline.txt" -` は unit-owned 8 code/test path と本 ledger の追加があるため expected exit 1。並行 owner が本 run 中に追加した `frontend/src/hooks/use-side-peek-dirty*`、`frontend/src/hooks/use-unsaved-changes*`、`3-session-agent.html` は baseline drift として表示されたが、本 unit の write target/log 外であり編集・revert・stage していない。本 unit owned delta は allowlist 9path のみ、変更上限10内。`git diff --cached --name-only` は空。
- Failure Signature log: attempt 1 / F14+F15 — expected=deny persona で mutation 0、actual=create/update/delete と検査 items が従来 payload で発行、verification=TDD RED、error=`8 failed | 62 passed`、fix=deny-all permission object + action 別 positive guard + ref/effect 最新値同期、result=GREEN。attempt 1 / Form wiring test — expected=Vaccination Form の mutable permission mock が edit mode で再評価、actual=同一 props の memo component を `rerender` したため旧 hook call が残存、verification=4 spec gate、error=`expected last called with "vaccination-1"`、fix=実 component semantics を変えず test を unmount→new render に修正、result=`4 files / 72 tests` PASS。同一 signature 再発なし。
- De-Sloppify: permission shape、deny constant、ref/effect、action helper を hook ごとに1組だけ保持。Form は既存 permission 値を中継するだけで UI 分岐は不変。test は product-facing deny/grant/revocation/第3象限だけを保持し、console、`any`、catch-all、style、dialog/toast、drive-by cleanup、items 重複 guard なし。primitive dependency 整理後の scoped ESLint、15/190 Vitest、design-audit、diff-check は全て PASS。
- Harness / orchestration: `tdd` + `sequential`（code/test 変更時 De-Sloppify overlay）を実運用し、指定4 backing skills を読了。planner と TDD guide の read-only pass を実装前に統合し、main agent が RED→GREEN→reconciliation、React/TypeScript/security の3 fresh reviewer が Independent Review を担当。planner の「Form/detail 単一路なので route 変更なし」「検査は parent 入口に1 guard」を accept。権限を submit/delete の2値へ圧縮する案は prompt の action 別権限要件に合わせて reject し、3値 shape を採用。stop condition=全 Acceptance Checklist PASS。
- Independent Review: React reviewer、TypeScript reviewer、security analyst の3 read-only pass が CRITICAL/HIGH/MEDIUM/LOW actionable finding 0、Approve。最新権限 ref、dependency、condition hook/a11y、deny-all/`=== true`、action 別配線、payload/第3象限、parent/items、route read 維持、backend endpoint RBAC、XSS/secret/log surface、allowlist を再照合。reviewer による file/git write なし。
- Coverage / full gates: executable 正本は frontend statements 43.78%、tolerance 0.5pp の CI ratchetで、固定80% unit threshold はない。coverage artifactを生成する full coverage と、禁止された全体 `type-check` / `test:run` / `build` / lint / install は未実行。CI またはユーザー手動で `docker compose exec frontend pnpm type-check`、`docker compose exec frontend pnpm test:run`、`docker compose exec frontend pnpm build` を実行する。
- Prompt-defect / Harness Improvement Feedback: none needed。validator、TDD、bounded retry、route deviation、review、reconciliation、allowlist が指定どおり機能した。
- Remaining scope / risk: U9〜U10、F16〜F17、backend RBAC の継続維持、full type-check/test/build/coverage は未実施。検査 parent/items は既存どおり2 HTTP operation のため parent success/items failure の部分成功 risk が残るが、本 unit の明示契約どおり parent-first/no duplicate guard を維持。stage/commit/stash/push 0。

### U9/F10 実行 ledger（2026-07-25）

- Status: COMPLETE — PermissionGroup の create/update/delete/rules mutation 境界を optional engage + action 別 positive match で fail-closed 化し、rule checkbox の accessible name と readOnly field 伝播を是正。TDD、scoped regression、scoped ESLint、design-audit、De-Sloppify、Independent Review、allowlist 帰属を完了。U10/F16/F9、backend/API/generated 型、他 master route へは進んでいない。git stage/commit/stash/push は0。
- Product gate: 要件責任者=曽我、目的=権限グループ定義を UI guard 迂回・権限剥奪 race・programmatic callback から保護する RBAC defense-in-depth。新規画面・入力・dialog/toast・permission hook・design token・backend変更は追加せず、backend RBACを最終防壁として維持した。
- Root cause / minimal patch: `useMasterSave` / `useMasterCRUD` は共有 mutation handler だが権限を受け取らず、PermissionGroup は表示上の hide/readOnly だけに依存していた。optional `permissions` を両hookへ追加し、PermissionGroupだけが engage。save は validation と `setValidationError(null)` の後・`crudStartSave` の前で create=`canCreate === true` / update=`canEdit === true`、delete は target確認後・mutate直前で `canDelete === true` を要求する。latest committed permission は `useLayoutEffect` 同期refで読み、absent consumerは従来挙動を維持。rules は親save `onSuccess` 内のままで独立guardを追加していない。
- Changed files: `frontend/src/features/master/hooks/use-master-save.ts`、同 `use-master-save.test.ts`、`frontend/src/features/master/hooks/use-master-crud.ts`、同 `use-master-crud.test.ts`、`frontend/src/features/master/routes/PermissionGroupSettings.tsx`、`frontend/src/features/master/components/PermissionRuleTableRow.tsx`、新規 `PermissionRuleTableRow.test.tsx`、`frontend/src/features/master/components/PermissionGroupSidePanel.tsx`、本ledger（9 files、上限10内）。
- Saved Prompt Validation Gate: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-u9-permission-group-mutation-guards.md` → `Prompt Craft Harness Validation: PASS`、`Profile: standard (declared-risk-tier)`、`Target: agent (detected)`、exit 0。
- Binding pre-read: `AGENTS.md`、`~/.agents/codex/AGENTS.md`、`.codex/config.toml`、`.claude/CLAUDE.md`、`.claude/rules/claude-code-usage.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/refs/typescript-react.md`、`.claude/refs/accessibility-rules.md`、`frontend/CLAUDE.md`、`frontend/src/features/CLAUDE.md`、`docs/product-philosophy.md`、`docs/ops/coverage-policy.md`、本節 FE12-02 C6a/F10/C-02/C-03/Mode 3/U6/U7/U8、`git show acebb4dca`、対象hook/route/component/test全文、`use-permission.ts`、`permission-rule-table-model.ts`、`MasterSidePanel.tsx`、`StatusToggleButton.tsx`、`tdd-workflow`、`react-testing`、`verification-loop`、`security-review`。binding矛盾なし。
- Start baselines: 編集前 `git status --porcelain | sort > "${TMPDIR:-/tmp}/fe12-u9-git-baseline.txt"` は空、indexも空。`docker compose exec frontend npx vitest run src/features/master` → `Test Files  26 passed (26)`、`Tests  181 passed (181)`、exit 0、出力は`${TMPDIR:-/tmp}/fe12-u9-vitest-baseline.txt`。開始時design-auditはgating各0件、`C18 raw legacy baseline — 10 件（non-gating ratchet）`、`design-system-audit: PASS — 違反 0 件`、exit 0。
- TDD RED: `docker compose exec frontend npx vitest run src/features/master/hooks/use-master-save.test.ts src/features/master/hooks/use-master-crud.test.ts src/features/master/components/PermissionRuleTableRow.test.tsx` → `Test Files  3 failed (3)`、`Tests  4 failed | 49 passed (53)`、exit 1。deny create/update/delete が mutation/transitionへ到達し、checkbox accessible name が空である旧F10を再現した。
- TDD GREEN / named regression: 同commandの最終結果は `Test Files  3 passed (3)`、`Tests  58 passed (58)`、exit 0。engaged deny/grant、permissions absent、create/edit交差、validation先行、grant→deny と absent→engaged deny のcaptured callback、delete pending target、resource+action accessible nameを固定した。
- Final scoped regression: `docker compose exec frontend npx vitest run src/features/master` → `Test Files  27 passed (27)`、`Tests  193 passed (193)`、exit 0。baseline比で新規failure 0、新規12 cases全PASS。開始/終了に共通する ReservationTypeSidePanel のnested form warningと occupations MSW未処理stderrは既存test harness出力で、本unitの新規failureではない。
- Scoped ESLint: `docker compose exec frontend npx eslint src/features/master/hooks/use-master-save.ts src/features/master/hooks/use-master-save.test.ts src/features/master/hooks/use-master-crud.ts src/features/master/hooks/use-master-crud.test.ts src/features/master/routes/PermissionGroupSettings.tsx src/features/master/components/PermissionRuleTableRow.tsx src/features/master/components/PermissionRuleTableRow.test.tsx src/features/master/components/PermissionGroupSidePanel.tsx --max-warnings 0` → diagnostics 0、exit 0（Compose DB env未設定warningのみ）。
- Design regression: `docker compose exec frontend pnpm design-audit` → C1/C3/C5〜C17/C18 table cell/C19 gating各0件、`C18 raw legacy baseline — 10 件（non-gating ratchet）`、`design-system-audit: PASS — 違反 0 件`、exit 0。開始時から全件数不変。
- Guard / rules / route contract: `PermissionGroupSettings.tsx` だけが `permissions: { canDelete }` と `{ canCreate, canEdit }` を渡す。reorderの既存 `if (!canEdit) return` は変更せず、`updateRulesMutation.mutateAsync` は親save `onSuccess` 内にだけ残した。denied save testは`onSuccess` 0回を確認するためrulesも推移的に0回。`MasterCRUDPage.tsx` / `MasterPageShell.tsx` と他master routeは未変更。
- readOnly / a11y: `PermissionGroupSidePanel` のtitle change handlerはreadOnly時returnし、status/description/color/rulesは`fieldset disabled={readOnly}`でnative disable。既存save/delete/readOnly UI gateも維持。rule checkboxは `RESOURCE_LABELS[resource] || resource` + `PERMISSION_ACTION_COLUMNS.label` の一意な`aria-label`を持ち、4 actionをrole/name queryで検証した。
- Consumer recensus / blast-radius follow-up: production route invocationは `useMasterCRUD` 20、`useMasterSave` 25。PermissionGroup以外の他19 master page/call siteはCRUD permissions未配線のまま、複合master内の追加saveを含めるとsave未配線は24 call sites。U9 scopeでは後方互換のため absent=従来発行を固定したが、同型の UI-gated / mutation-未再検査境界は残る。coordinating sessionは全master横展開を defense-in-depth 候補 finding（P1/P2）として裁定すること。
- Whitespace / diff: production 5 filesの通常 `git diff --stat` と `git diff -w --stat` はともに `5 files changed, 60 insertions(+), 5 deletions(-)`。`git diff --check` はstdout 0行、exit 0。console / raw HTML / secret / raw hex / TODO / FIXME / drive-by cleanupの追加なし。
- Allowlist / index: required raw comparisonはunit-owned code/test 8 pathsと本ledgerの追加によりexpected exit 1。開始baselineは空で、本unit deltaはallowlist 9 pathsのみ、変更上限10内。`git diff --cached --name-only` は空。
- Failure Signature log: attempt 1 / F10 RED — expected=engaged denyでmutation 0、actual=create/update/delete発行 + checkbox name空、verification=named RED、error=`4 failed | 49 passed`、fix=optional engage + action別positive-match guard + route配線 + aria-label、result=GREEN。attempt 1 / Independent Review — expected=権限剥奪commit後に旧allowなし、actual=passive effect同期にcommit→effect競合窓、verification=React/TypeScript/security fresh review、error=HIGH 1件、fix=permission ref同期を`useLayoutEffect`へ変更しabsent→engaged deny回帰2件追加、result=named 3/58・master 27/193・ESLint/design PASS。attempt 1 / whitespace gate — expected=通常/`-w` stat一致、actual=fieldset追加時の再indentで90/35対59/4、verification=production stat比較、fix=挙動を変えず既存indentへ復元、result=60/5で一致。同一signature再発なし。
- De-Sloppify: permission shape/ref/guardはhookごとに1組だけ、routeは既存permission値を中継するだけ。rules/reorder重複guard、toast/dialog、console、`any`、catch-all、style token、他page変更なし。testはproduct-facing deny/grant/absent/revocation/validation/payload/a11yだけを保持し、cleanup後のnamed/full Vitest、ESLint、design、diff-checkは全PASS。
- Harness / orchestration: `tdd` + `sequential`（code/test変更時De-Sloppify overlay）を実運用し、指定4 backing skillsを読了。plannerとTDD guideのread-only passを実装前に統合し、main agentがRED→GREEN→reconciliation、React/TypeScript/securityの3 fresh reviewerがIndependent Reviewを担当。plannerのexact consumer recensus、TDD guideのvalidation/revocation casesをaccept。transition内guard案、rules独立guard、他page default-deny、allowlist外route/readOnly test変更はreject。stop condition=全Acceptance Checklist PASS。
- Independent Review: 初回React/TypeScript/security fresh passはpermission refのpassive effect競合窓を共通HIGHとして検出。`useLayoutEffect`同期 + absent→engaged deny casesへ修正し、post-fix再審査を実施。最終CRITICAL/HIGH actionable finding 0、Approve。route wiring/readOnly専用test追加はallowlist外のためMEDIUM提案として採用せず、source構造check + master regressionで代替した。
- Coverage / full gates: executable正本はfrontend statements 43.78%、tolerance 0.5ppのCI ratchetで、一律80% thresholdはない。full coverage artifact生成と、prompt禁止の全体 `type-check` / `test:run` / `build` / lint / installは未実行。CIまたはユーザー手動で `docker compose exec frontend pnpm type-check`、`docker compose exec frontend pnpm test:run`、`docker compose exec frontend pnpm build` を実行する。
- Prompt-defect / Harness Improvement Feedback: P2 — promptの「両hook consumer各20」はcurrent codeでCRUD 20 / save 25だったため、current executable sourceを優先して補正。P2 — readOnly全fieldのsemantic regression testとroute wiring testは既存allowlistに含まれず、static source checkへフォールバックした。次回promptは対象test allowlistを明記する。eval captureは生成側session所掌。
- Remaining scope / risk: backend RBACを最終防壁として継続維持。他19 master page（save 24 / CRUD 19 call sites）の同型defense-in-depth横展開、U10、F16/F9、full type-check/test/build/coverageは未実施。stage/commit/stash/push 0。

### FE12-09 実行 ledger（2026-07-25）

- Status: COMPLETE — （2026-07-25 coordinating session が frontend service 復旧後に最終deterministic gateを実測）De-Sloppify後の必須gateを再実行し `docker compose exec frontend npx vitest run line-reserve/src/pages/ConfirmPage.test.tsx` → `Test Files 1 passed (1)` / `Tests 4 passed (4)` exit 0、`docker compose exec frontend npx eslint line-reserve/src/pages/ConfirmPage.tsx` → diagnostics 0 exit 0。全Acceptance Checklist項目PASS。coordinating sessionが実ファイル`ConfirmPage.tsx`を独立精査し useActionState(L143)/`<form action>`(L223)/`type="submit" disabled={isPending}`(L227)/LIFF await→onConfirm(L128-129)/409 clear(L132-137)/500文言(L139)を確認。executor帰属diffは`ConfirmPage.tsx`+本ledgerの2 allowlist pathのみ、`git diff --check` exit 0。FE12-03/05/06/07/08/10へは進まず、git stage/commit/stash/pushは0（commitはユーザー判断）。（下記の停止時点blocker記録は当時の履歴として保全）
- Product gate: 要件責任者=曽我、目的=84リーフルートをReact実装正本へ収束させるFE12計画のうち、LINE予約確定画面のmanual pending/error stateとtry/finallyを削除して既存React 19 form actionへ統合すること。新規画面・入力・確認dialog・自動化・shared abstractionは追加していない。
- Root cause / minimal patch: `ConfirmPage` は `useState(submitting/error)` と `useCallback` handlerでpending/errorを手動管理していた。typed `ConfirmFormState` とlocal `confirmAction`を `useActionState`へ渡し、footerを`<form action={formAction}>`、既存`PrimaryButton`を`type="submit"` / `disabled={isPending}`へ変更。payload、409 parse/default、500文言、`createReservation`→`await sendLiffMessage`→`onConfirm`順序は維持し、success/409は`{ error: null }`、その他errorは既存文言を返す。
- Changed files: `frontend/line-reserve/src/pages/ConfirmPage.tsx`、本ledger。`ConfirmPage.test.tsx`と`PrimaryButton.tsx`はHEADからbyte差分0。
- Saved Prompt Validation Gate: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-09-line-reserve-confirmpage-react19-form-action.md` → `Prompt Craft Harness Validation: PASS`、`Profile: standard (declared-risk-tier)`、`Target: agent (detected)`、`VALIDATOR_EXIT=0`。SHA-256=`73d7e3046b530559f64b7c7e898c1d4d5f310d93c32f90f53ffff54ab4238da7`。
- Binding pre-read: saved prompt、`AGENTS.md`、`.claude/CLAUDE.md`、`.claude/rules/claude-code-usage.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/refs/typescript-react.md`、`.claude/refs/error-handling.md`、`.claude/refs/accessibility-rules.md`、`frontend/CLAUDE.md`、`docs/product-philosophy.md`、`.codex/config.toml`、relevant `.codex/agents/*.toml`、FE12-09 row / FE12-04 / U9 ledger、`ConfirmPage.tsx`、`ConfirmPage.test.tsx`、`PrimaryButton.tsx`、`verification-loop`、`react-testing`。
- Start baseline: allowlist 3pathはclean、indexは空。global worktreeには開始前からA4/deploy/examinations/medical-records等の別owner WIPあり。`docker compose exec frontend npx vitest run line-reserve/src/pages/ConfirmPage.test.tsx` → `Test Files  1 passed (1)`、`Tests  4 passed (4)`、`FE12_09_BASELINE_VITEST_EXIT=0`。
- First implementation gates: 同Vitest → `Test Files  1 passed (1)`、`Tests  4 passed (4)`、`FE12_09_POST_EDIT_VITEST_EXIT=0`。`docker compose exec frontend npx eslint line-reserve/src/pages/ConfirmPage.tsx` → diagnostics 0、`FE12_09_SCOPED_ESLINT_EXIT=0`。test timing accommodationは不要でtest assertion/fileを変更していない。
- De-Sloppify: inline action初版のrequest payload再indentをlocal `confirmAction`へ整理し、production通常/`-w` statをともに`1 file changed, 22 insertions(+), 18 deletions(-)`へ収束。`git diff --check -- frontend/line-reserve/src/pages/ConfirmPage.tsx`はstdout 0行/exit 0。added `any` / console / raw HTML / eval / secret / raw hex / TODO / FIXMEはmatch 0。rows、format helper、`sendLiffMessage`本体、step progress、payload、className、公開props、全指定文言の意図しない変更なし。
- Final deterministic gate blocker: De-Sloppify後の `docker compose exec frontend npx vitest run line-reserve/src/pages/ConfirmPage.test.tsx` → `service "frontend" is not running`、`FE12_09_DESLOPPIFY_VITEST_EXIT=1`。`docker compose ps frontend`を2回確認し、両方ともheaderのみでrunning container 0。code failureとして扱わず、禁止された`docker compose up` / `restart`は実行していない。current codeへのfinal scoped ESLintも同じ外部state blockerのため未実行。
- Static reconciliation: manual-state grep `grep -nE "useState|setSubmitting|useCallback" ...` はstdout 0行/exit 1（expected no match）。action-pattern grepは`useActionState`、`<form action={formAction}>`を検出/exit 0。current source `ConfirmPage.tsx:128-139` は LIFF await→onConfirm、409→`onSlotTaken`+clear、500→exact error state、`:223-230` はform action + submit + pending disabled/textを保持。
- Allowlist/index: unit開始時allowlist 3pathはclean。本unit帰属は`ConfirmPage.tsx`と本ledgerの2pathのみ、testは変更なし、3path上限内、indexは空。literal global `git diff --name-only` は開始前WIPと本run中の並行driftを含むためsaved prompt記載のglobal subset gateをPASSにできない。他owner pathは編集・revert・stageしていない。
- Failure Signature log: attempt 1 / post-De-Sloppify Vitest — expected=current codeで1 file / 4 tests PASS、actual=`service "frontend" is not running`、verification=required scoped Vitest、error signature=frontend service absent、attempted fix=code変更なしで`docker compose ps frontend`を2回再確認、result=external state不変のためBLOCKED。required input=coordinating sessionがfrontend serviceを復旧後、scoped Vitestとscoped ESLintを再実行すること。
- Harness / orchestration: `tdd` + `sequential`（De-Sloppify overlay）を実運用し、`verification-loop`と`react-testing`を読了。planner / TDD guideのread-only passを実装前に統合し、React / TypeScript / general reviewerの3 fresh passを実施。local typed action、test無変更、source ordering案をaccept。retry中error非表示案と追加pending/ordering test案はpromptの最小scope・test編集制約によりreject。
- Independent Review: TypeScript reviewerとgeneral reviewerはcurrent codeにCRITICAL/HIGH/MEDIUM code finding 0でApprove。React reviewerはCRITICAL/HIGH 0、既存testがpending二重送信・LIFF実送信順序を直接証明しないMEDIUM evidence gapのみでWarning。静的にはhook/action/a11y/409/500/orderを満たす。reviewerによるfile/git/Docker writeは0。
- Coverage / full gates: prompt禁止の全体 `type-check` / `test:run` / `build` / lint / installとcoverage artifactは未実行。`tsc --noEmit`はtest除外のため本unitではdeferred。
- Prompt-defect / Harness Improvement Feedback: P1 — literal global allowlist gateは開始前からdirtyなshared worktreeで達成不能であり、baseline attribution gateを明記すべき。P2 — Objectiveはdouble-submit / LIFF ordering / 409 alert absenceを既存4 testが検証すると述べるが、現testはpending中disabled/API回数、LIFF実送信順序、409 alert absenceを直接assertしない一方、test編集をtiming accommodationに限定している。生成側eval corpusで整合させる。
- Remaining risk / resume condition: 【解消済 2026-07-25】frontend service復旧後にfinal scoped Vitest（`4 passed` exit 0）とESLint（diagnostics 0 exit 0）を実測しRun status=COMPLETEへ更新。backend RBACを最終防壁として継続。全体type-check/test/build、FE12-03/05/06/07/08/10は別scope。

### FE12-06 実行 ledger（2026-07-25）

- Status: COMPLETE — `frontend/src/features/estimates/utils/` を削除し、multi-consumer 2 moduleと隣接testをfeature-local `constants/` / `lib/`へbyte-identical relocation、single-consumer expiry helper/testを削除して`EstimateDetail`へinlineした。5 consumerのimportだけを更新し、全Acceptance Checklist項目PASS。FE12-03/05/07/08/10へは進まず、commit/stash/push/mergeは0。
- Root cause / minimal patch: feature内部だけで消費されるstatus optionとlocked-status predicateが禁止`utils/`に残り、expiry predicateは単一consumerに対する薄いwrapperだった。status option pairを`constants/`、locked-status pairを`lib/`へ移動し、`EstimateDetail`は`estimate.validUntil ? estimate.validUntil.slice(0, 10) < todayJSTISO() : false`を直接計算。exported symbol、constant value/label、locked-edit message、route assertion、`index.ts`は変更していない。
- Assumption: feature-type-dependent modules relocate feature-locally (layer-inversion lint forbids the plan row's literal `src/lib`/`src/constants` destinations); adjudicated by the generating session on 2026-07-25.
- Changed files (16-path allowlist only): delete=`frontend/src/features/estimates/utils/estimate-status-options.ts`、`estimate-status-options.test.ts`、`is-estimate-locked-status.ts`、`is-estimate-locked-status.test.ts`、`is-estimate-expired.ts`、`is-estimate-expired.test.ts`; add=`frontend/src/features/estimates/constants/estimate-status-options.ts`、`estimate-status-options.test.ts`、`frontend/src/features/estimates/lib/is-estimate-locked-status.ts`、`is-estimate-locked-status.test.ts`; edit=`frontend/src/features/estimates/hooks/use-estimate-form.ts`、`routes/EstimateForm.tsx`、`routes/EstimateForm.test.tsx`、`routes/EstimateDetail.tsx`、`routes/EstimateList.tsx`、本ledger。
- Saved Prompt Validation Gate: `node /Users/minoru/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-06-estimates-utils-relocation.md` → `Prompt Craft Harness Validation: PASS`、`Profile: standard (declared-risk-tier)`、`Target: agent (detected)`、`VALIDATOR_EXIT=0`。SHA-256=`8673cc8e54c6bc3afe7b972b3b095c0ca6e879cb501172825d7e0cbda4345773`。
- Binding pre-read: saved prompt、`AGENTS.md`、`.claude/CLAUDE.md`、`.claude/rules/claude-code-usage.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/refs/typescript-react.md`、`frontend/CLAUDE.md`、`frontend/src/features/CLAUDE.md`、`docs/product-philosophy.md`、`.codex/config.toml`、`.codex/agents/*`、FE12-06 row / FE12-04 / FE12-09 ledger、対象source/test、`~/.agents/skills/verification-loop/SKILL.md`、`~/.agents/skill-library/react-testing/SKILL.md`。
- Start baseline: `git status --porcelain | sort > "${TMPDIR:-/tmp}/fe12-06-baseline.txt"`でshared dirty-treeを固定。開始前から`FE-refactor.md`、A4/deploy、examinations/medical-records、backend staged rename等の別owner WIPあり。`docker compose exec frontend npx vitest run src/features/estimates` → `Test Files 9 passed (9)`、`Tests 82 passed (82)`、exit 0。feature file count=27、test file count=9。
- Final tests: 同一command → `Test Files 8 passed (8)`、`Tests 77 passed (77)`、exit 0。削除した`is-estimate-expired.test.ts`だけがtest-file countから減少し、relocated 2 unit tests、EstimateList / EstimateForm / EstimateDetail route testsを含む全8 filesがPASS。既存`useActionState` transition warningはbaseline/final共通で本差分由来ではない。
- Structural gates: `find frontend/src frontend/liff/src frontend/line-reserve/src -type d -name utils -print` → stdout 0行、`FIND_EXIT=0`。`rg -n "estimates/utils|utils/estimate-status-options|utils/is-estimate-locked-status|utils/is-estimate-expired" frontend/src frontend/liff/src frontend/line-reserve/src` → stdout 0行、`STALE_REF_EXIT=1`（expected no match）。`rg -n "isEstimateExpired" ...` → stdout 0行、`EXPIRED_SYMBOL_EXIT=1`。feature file count=25、test file count=8（net file count -2）。
- Lint / inline gates: `docker compose exec frontend npx eslint src/features/estimates` → diagnostics 0、`ESLINT_EXIT=0`。`grep -n "todayJSTISO\|isExpired" frontend/src/features/estimates/routes/EstimateDetail.tsx` → `2:import { todayJSTISO } from "@/lib/jst-date";`、`46:  const isExpired = estimate.validUntil ? estimate.validUntil.slice(0, 10) < todayJSTISO() : false;`、`117:                  {isExpired ? (`、`INLINE_GREP_EXIT=0`。green `EstimateDetail.test.tsx`は8 tests。
- Byte-identity gate: old HEAD → new working-tree SHA-256は4件すべて一致。status source=`75f350a8deaeaa24c0f29084ba5c1b03a53ba25969db33a52ed87d1b36049694`; status test=`59ee80a7419502814665e0bbcbede969edc6031e9161489a04e623d56facf72b`; locked source=`8342168f0929850489d86a5d792f80f54e7887082f1de45106e7cb80a2de5765`; locked test=`783cd882ca4a93d0d17775aa46d885d04d55ee7ffa3c5b69eb219c87c2fadd48`。`git diff --cached --stat`も4 files / 0 insertions / 0 deletions、各rename similarity 100%。
- Allowlist / staging: baseline statusとの差はexpiry旧2 path delete、5 consumer edit、4 rename recordsだけで、各old/new pathを展開すると本unit code 15 pathすべてallowlist内。本ledgerは開始前から`M`だったため別途`${TMPDIR:-/tmp}/fe12-06-ledger-baseline.md`を固定しappend-only比較。本runの`git mv`による4 renameだけstaged（R100）、5 consumer + expiry旧2 path + ledgerはunstaged。開始前WIPのstatus行は同一で、他owner pathをedit/revert/stage/stashしていない。`git diff --check -- frontend/src/features/estimates FE-refactor.md` → stdout 0行、exit 0。
- De-Sloppify: 新規helper/wrapper/abstraction、test assertion、console、commented-out code、catch-all、`any`、raw HTML/eval、TODO/FIXME、drive-by format変更なし。`EstimateDetail.test.tsx` / `EstimateList.test.tsx`はdiff 0、`EstimateForm.test.tsx`はimport path 1行だけ。cleanup patch不要のためfinal Vitest/ESLint結果を維持。
- Harness / orchestration: `tdd` + `sequential`（De-Sloppify overlay）を実運用し、`verification-loop`と`react-testing`を読了。plannerはexact 16-path disposition / append-only ledger注意、TDD guideはbaseline 9→final 8 test-file contractとDocker command trapsを提示し、main agentがaccept。general reviewerとTypeScript reviewerのfresh read-only passはともにCRITICAL/HIGH/MEDIUM 0、Approve。plan-row再裁定、assertion追加、他unit着手はreject。
- Failure Signature log: attempt 1 / relocation tool — expected=4 fileをcontent-preserving move、actual=`apply_patch verification failed: invalid hunk ... is empty`、verification=tool result、error signature=empty move-only hunk unsupported、fix=saved promptが許可するexplicit `git mv`へ切替、result=4件R100 / byte-identical PASS。attempt 1 / byte-hash script — expected=4 pair比較、actual=zshがspace-separated pairを分割せず`fatal: path ... does not exist` + false FAIL、verification=hash helper output、error signature=zsh pair splitting、fix=各old/newを明示引数にした`check_pair` functionへ変更、result=4件すべてSHA-256一致。attempt 1 / final file-count shell — expected=25、actual=zsh special parameter `path`をloop変数に使い同一shellの`PATH`が壊れて`find/wc/tr: command not found`、verification=final count helper output、error signature=zsh `path` collision、fix=loop変数を`target_path`へ変更したfresh command、result=`FEATURE_FILE_COUNT=25`。code/test failureなし、同一signature再発なし。
- Coverage / full gates: prompt禁止の全体`pnpm type-check` / `test:run` / `build` / lint / installとcoverage artifactは未実行。`tsc --noEmit`はtest除外のため本unitではdeferredし、scoped Vitest + ESLintを実行正本とした。
- Prompt-defect / Harness Improvement Feedback: P2 — Assumption文はglobal `src/lib` / `src/constants`の双方を「layer-inversion lint」で一括説明するが、current ESLintの直接的layer inversion対象は`src/lib`等で、feature-specific constantのglobal配置禁止はShared Constants ownership ruleが補完する。裁定destinationと実装には影響なし。P2 — product-philosophy上の個人責任者名はsaved promptにないが、behaviorを追加しない削除型FE12 maintenance unitのため非blocking。専用loop-monitor facilityはsequential single-passには不要。
- Remaining risk / follow-up: full type-check/test/build/coverageはCI / coordinating sessionへdefer。4 renameは`git mv`によりstaged、その他はunstagedで、commitはユーザー判断。backend/API/DB/clinic isolationと他FE12 unitは未変更。

### FE12-05 実行 ledger（2026-07-25）

- Status: COMPLETE —（2026-07-25 coordinating sessionがMode 3裁定）BLOCKEDの唯一の原因はprompt側の自己矛盾gate（保持必須のline 56 commentが`useMemo`文字列を含むため、negative substring grepが構造的に充足不能）であり、実装欠陥ではない。gateをcall-site/import-aware判定へ訂正し独立再検証: `grep -nE "useMemo\s*\("` no match / `grep -nE "^import.*useMemo"` no match / 全occurrenceはcomment 1行のみ / scoped Vitest `3 passed (3)` `6 passed (6)` exit 0 / scoped ESLint diagnostics 0 / 帰属2 allowlist pathのみ — 全て実測PASS。prompt defect（preserve指示対象とnegative grepの衝突）はeval corpus送りのHarness Improvement Feedbackとして記録。（下記の停止時点記録は当時の履歴として保全）BLOCKED時の記録: production patch自体はCOMPLETEで、`effectiveMax` / `initialDate` の2 `useMemo` call siteとunused importを削除し、final scoped Vitestは`Test Files 3 passed (3)` / `Tests 6 passed (6)`、scoped ESLintはdiagnostics 0、いずれもexit 0。だがsaved promptが保持を必須とする既存commentに`useMemo`が含まれる一方、同promptの必須gate `grep -n "useMemo" ...` はno match / exit 1を要求するため、literal Acceptance ChecklistをPASSにできない。commentを変更せず、必須grepをBLOCKEDとして記録した。FE12-03/07/08/10へ進まず、stage/commit/stash/pushは0。
- Root cause / minimal patch: string primitiveの単純導出に不要なdependency trackingが残っていた。React importから`useMemo` tokenだけを削除し、`effectiveMax`を既存ternary、`initialDate`を既存`clampDate` callのplain `const`へ置換。`memo()` wrapper、permission ref/effect、callbacks、transition、section component、line 56 comment、test filesは変更していない。
- Changed files: `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx`、本ledgerの2 allowlist pathのみ。3 test filesはHEAD差分0。
- Saved Prompt Validation Gate: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-05-daily-records-tab-usememo-removal.md` → `Prompt Craft Harness Validation: PASS`、`Profile: standard (declared-risk-tier)`、`Target: agent (detected)`、exit 0。
- Binding pre-read: saved prompt、`.claude/CLAUDE.md`、`.claude/rules/claude-code-usage.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/refs/typescript-react.md`、`frontend/CLAUDE.md`、`frontend/src/features/CLAUDE.md`、`docs/product-philosophy.md`、FE12-05 row / FE12-06 / FE12-09 ledger、`DailyRecordsTab.tsx`、隣接3 test file inventory、`~/.agents/skills/verification-loop/SKILL.md`、`~/.agents/skill-library/react-testing/SKILL.md`。
- Start baseline: `git status --porcelain | sort > "${TMPDIR:-/tmp}/fe12-05-baseline.txt"`でshared dirty-treeを固定。allowlist 2pathはclean、global worktreeには開始前からMakefile、A4/deploy、examinations、medical-records、scripts等の別owner WIPあり。`docker compose exec frontend npx vitest run src/features/hospitalization/components/DailyRecordsTab` → `Test Files 3 passed (3)` / `Tests 6 passed (6)` / `FE12_05_BASELINE_VITEST_EXIT=0`。
- Final tests / lint: 同directory-scoped Vitest → `Test Files 3 passed (3)` / `Tests 6 passed (6)` / `FE12_05_FINAL_VITEST_EXIT=0`。`docker compose exec frontend npx eslint src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx` → diagnostics 0 / `FE12_05_FINAL_ESLINT_EXIT=0`。baseline/final test countsは完全一致。
- Structural reconciliation: 必須`grep -n "useMemo" frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx` → `56:    // rerender-simple-expression-in-memo: string primitive は値比較のため useMemo 不要` / `FE12_05_REQUIRED_GREP_EXIT=0`（prompt内のkeep-comment制約との矛盾によりBLOCKED）。補助code-aware gate `rg -n '\buseMemo\s*\(' ...` → stdout 0行 / `FE12_05_CALLSITE_RG_EXIT=1`、import先頭inspection → stdout 0行 / `FE12_05_IMPORT_RG_EXIT=1`で、call site / importはともに0。
- Allowlist / staging: `git status --porcelain | sort | diff "${TMPDIR:-/tmp}/fe12-05-baseline.txt" -` → `0a1` / `>  M FE-refactor.md` / `6a8` / `>  M frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx` / `FE12_05_BASELINE_STATUS_DIFF_EXIT=1`（expected deltaあり）。追加されたstatus行は2 allowlist pathだけで、開始前の他owner WIP status行は同一。`git diff --cached --name-only -- <allowlist>`はstdout 0行、両pathともunstaged。test filesは`git diff --exit-code` / `FE12_05_TEST_FILES_UNCHANGED_EXIT=0`。
- De-Sloppify: production diffは`1 file changed, 3 insertions(+), 9 deletions(-)`。test diff 0、new helper/abstraction/guard、console、commented-out code、catch-all、`any`、raw HTML/eval、TODO/FIXME、format churn、他hook変更なし。`git diff --check -- DailyRecordsTab.tsx` → stdout 0行 / `FE12_05_TSX_DIFF_CHECK_EXIT=0`。cleanup patch不要後にfinal Vitest/ESLintを再実行しPASS。
- Harness / orchestration: `tdd` + `sequential`（De-Sloppify overlay）を実運用し、`verification-loop`と`react-testing`を読了。existing suiteをGREEN baselineとして固定し、literal minimal refactor後に同じ3 file / 6 test contractを再実測。REDはbehavior-preserving refactorのためnot applicable。TDD guideはbroader date logicへscope拡大しない条件でsuite十分と判定し、general / React / TypeScript reviewerはCRITICAL/HIGH/MEDIUM 0、全員Approve。stop conditionは3 checklist PASS + 1 prompt-contract BLOCKED。
- Failure Signature log: attempt 1 / baseline wrapper — expected=Vitest green + exit 0、actual=Vitest自体は3 files / 6 tests PASSだがzsh read-only parameter `status`への代入でwrapper exit 1、verification=baseline command wrapper、error signature=`zsh:1: read-only variable: status`、fix=wrapper変数を`result_code`へ変更、result=同suite 3 files / 6 tests PASS / exit 0。attempt 1 / required structural gate — expected=no match / exit 1、actual=保持必須comment line 56がmatch / exit 0、verification=prompt指定grep、error signature=keep-comment vs substring-grep contradiction、attempted fix=source変更なしでcall-site/import専用gateを補助実行、result=実code requirementは満たすがliteral required gateはBLOCKED。required input=saved promptのverification methodをcall-site/import-aware checkへ訂正すること。
- Coverage / full gates: prompt禁止のrepo-wide `pnpm type-check` / `pnpm test:run` / `pnpm build` / lint / installとcoverage artifactは未実行。scoped Vitest + ESLintを実行正本とした。
- Remaining risk / follow-up: runtime/code riskはreview・tests・lint上なし。残件はsaved promptのcomment保持制約とsubstring grepの自己矛盾のみ。backend/API/DB/clinic isolation、他FE12 unitは未変更。

### FE12-08 実行 ledger（2026-07-25）

- Status: COMPLETE — line-reserve 3画面のlocal date/time formatter 6関数を既存`@/shared-liff/jst-date`へ統合し、`formatTimeHHMM`をTDDで追加した。valid 4文字`HHMM`の出力、date padding、Confirm送信/Completeの`〜`、Confirm表示/TimeSelectの` 〜 `を維持。全Acceptance Checklist項目PASS。FE12-03/07/10へ進まず、git stage/commit/stash/pushは0。
- Product gate: behavior追加ではなく、重複6関数を削除して既存shared helperへ収束する②削除/③簡素化unit。新規画面・入力・確認dialog・自動化・helper treeは追加せず、valid入力のuser-facing outputは不変。saved promptに個人責任者名はないが、既存FE12計画のmaintenance unitであり非blocking。
- Root cause / minimal patch: ConfirmPage/TimeSelectPageに同一guard付き`formatTime`、ConfirmPage/CompletePageに`formatJapaneseDate` thin wrapper、CompletePageにlocal range formatterが残存していた。sharedにprompt逐語の`formatTimeHHMM` 1関数だけを追加し、3画面はshared helperを直接呼ぶよう置換。`formatJapaneseDate`、header、`extractConfirmationNumber`、range assembly、payload/action/UIは変更していない。
- Changed files: `frontend/src/shared-liff/jst-date.ts`、`frontend/src/shared-liff/jst-date.test.ts`、`frontend/line-reserve/src/pages/ConfirmPage.tsx`、`frontend/line-reserve/src/pages/TimeSelectPage.tsx`、`frontend/line-reserve/src/pages/CompletePage.tsx`、本ledgerの6 allowlist pathのみ。
- Saved Prompt Validation Gate: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-08-line-reserve-time-formatter-consolidation.md` → `Prompt Craft Harness Validation: PASS`、`Profile: standard (declared-risk-tier)`、`Target: agent (detected)`、exit 0。
- Binding pre-read: saved prompt、`AGENTS.md`、`.claude/CLAUDE.md`、`.claude/rules/claude-code-usage.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/refs/typescript-react.md`、`.claude/refs/accessibility-rules.md`、`frontend/CLAUDE.md`、`docs/product-philosophy.md`、`.codex/config.toml`、relevant `.codex/agents/*.toml`、FE12-08 row / FE12-05/06/09 ledger、shared target/test、3 page files、ConfirmPage/TimeSelectPage tests、`~/.agents/skills/verification-loop/SKILL.md`、`~/.agents/skill-library/react-testing/SKILL.md`。
- Start baseline: `git status --porcelain | sort > "${TMPDIR:-/tmp}/fe12-08-baseline.txt"`でshared dirty-treeを固定。6 allowlist pathはclean、indexは空。global worktreeには開始前からBE/A4/deploy/examinations/medical-records/scripts等の別owner WIPがあり、編集・revert・stage・stashしていない。
- TDD RED: test importと2 casesだけを先行追加し、`docker compose exec frontend npx vitest run src/shared-liff/jst-date.test.ts` → `TypeError: formatTimeHHMM is not a function`、`Test Files 1 failed (1)`、`Tests 2 failed | 10 passed (12)`、exit 1。production export欠落だけを再現した。
- TDD GREEN: prompt逐語bodyを追加後、同command → `Test Files 1 passed (1)`、`Tests 12 passed (12)`、exit 0。`"1000" → "10:00"`、`"" → ""`、`"123" → "123"`を固定した。
- Page regression baseline/final: page swap前の`docker compose exec frontend npx vitest run line-reserve/src/pages/ConfirmPage.test.tsx line-reserve/src/pages/TimeSelectPage.test.tsx` → `Test Files 2 passed (2)`、`Tests 10 passed (10)`、ConfirmPage `4 tests` / TimeSelectPage `6 tests`、exit 0。swap後の同commandも`Test Files 2 passed (2)`、`Tests 10 passed (10)`、ConfirmPage `4 tests` / TimeSelectPage `6 tests`、exit 0で件数一致。
- Structural / lint gates: prompt指定`grep -n "^function formatTime\|^function formatDate\|^function formatDatePadded" frontend/line-reserve/src/pages/ConfirmPage.tsx frontend/line-reserve/src/pages/TimeSelectPage.tsx frontend/line-reserve/src/pages/CompletePage.tsx` → stdout 0行 / exit 1（expected no match）。`docker compose exec frontend npx eslint src/shared-liff/jst-date.ts src/shared-liff/jst-date.test.ts line-reserve/src/pages/ConfirmPage.tsx line-reserve/src/pages/TimeSelectPage.tsx line-reserve/src/pages/CompletePage.tsx` → diagnostics 0 / exit 0（Composeの未設定DB env warningのみ）。`git diff --check` → stdout 0行 / exit 0。
- Assumption deviation: valid 4文字`HHMM`はbyte-identical。CompletePageのmalformed入力はprompt記載の4文字未満だけでなく、4文字超過でも旧`slice(2)`が末尾を保持する一方、shared `slice(2, 4)`は4文字目までに切るためsemantic deltaがある。ConfirmPage/TimeSelectPageは旧bodyと逐語一致のためmalformedを含め不変。out-of-contract差として記録し、scopeを拡大してvalidation/API/testを変更していない。
- De-Sloppify: testはproduct contract 2 casesだけ、productionはshared function 1個とdirect call swapだけ。console、`any`、raw HTML/eval、TODO/FIXME、commented-out code、catch-all、new abstraction、format churn、他behavior変更なし。cleanup patch不要。separator/import/manual diff inspectionとscoped gatesを維持。
- Harness / orchestration: `tdd` + `sequential`（De-Sloppify overlay）を実運用し、`verification-loop`と`react-testing`を読了。plannerはexact 6-path dispositionとmalformed長文字差、TDD guideはminimal 2-test RED/GREENを提示しmain agentがaccept。stop condition=全Acceptance Checklist PASS。
- Independent Review: general reviewer、TypeScript reviewer、React reviewerのfresh read-only 3 passはいずれもCRITICAL/HIGH/MEDIUM 0、Approve。shared body逐語一致、全call site、valid出力、`〜` / ` 〜 `、CompleteのASCII space、type/import/hook/a11y、scopeを再照合。reviewerによるfile/git writeは0。
- Failure Signature log: none。TDD REDは計画された期待失敗であり、unexpected checklist failure / retryは0。
- Coverage / full gates: prompt禁止の全体`pnpm type-check` / `pnpm test:run` / `pnpm build` / lint / installとcoverage artifactは未実行し、CI / coordinating sessionへdefer。CompletePageにはtest fileがないため、prompt指定の逐語replacement、shared unit test、lint、independent reviewでmitigate。
- Prompt-defect / Harness Improvement Feedback: P2 — Contextのmalformed-input Assumptionは4文字未満のguard差だけを記述し、4文字超過で旧`slice(2)`とshared `slice(2, 4)`が異なる点を欠落している。valid入力Success Criteriaには影響しないが、次回promptはmalformed全長域を明記する。専用loop monitorはsequential single-passのため不要。
- Remaining risk / follow-up: invalid 4文字超過入力のComplete表示差、CompletePage専用regression test不在、repo-wide type-check/test/build/coverage未実施。backend/API/DB/clinic isolationと他FE12 unitは未変更。変更はunstaged、commit/pushなし。
- Allowlist / index: `git status --porcelain | sort | diff "${TMPDIR:-/tmp}/fe12-08-baseline.txt" -` → `1c1` / `<  M BE-refactor.md` / `---` / `>  M FE-refactor.md` / `3a4,6` / `>  M frontend/line-reserve/src/pages/CompletePage.tsx` / `>  M frontend/line-reserve/src/pages/ConfirmPage.tsx` / `>  M frontend/line-reserve/src/pages/TimeSelectPage.tsx` / `11a15,16` / `>  M frontend/src/shared-liff/jst-date.test.ts` / `>  M frontend/src/shared-liff/jst-date.ts` / exit 1（expected deltaあり）。追加されたstatus行は本unitの6 allowlist pathだけ。開始前`BE-refactor.md` WIPの消失は本runのwrite target外で、並行ownerのexternal driftとして帰属。その他の開始前WIP status行は不変。`git diff --cached --name-only`はstdout 0行、indexは空。

### FE12-10 実行 ledger（2026-07-25）

- Status: COMPLETE — 14 feature sourceのhalfwidth `¥` + `.toLocaleString()` 24箇所を既存`formatCurrency`へ統合し、category (a)のnullish + `"-"` ternary 2件だけをcollapseした。category (b)/(c)のguard、U+002D `-`、U+2014 `—`、U+2212 `−`、日本語前後space、全角括弧を維持。全Acceptance Checklist項目PASS。FE12-03/07、FE12-02 U10、fullwidth `￥` clusterへは進まず、git stage/commit/stash/pushは0。
- Product gate / root cause / minimal patch: 新機能追加ではなく、`frontend/CLAUDE.md` Shared Helper配置MANDATORYへ既存表示を収束させる②削除/③簡素化unit。24 inline formatterと等価なnull guard 2件を削除し、既存helperを直接呼ぶ。追加は既存import 3行だけで、新helper・API・型・test・UI flow・権限/clinic境界変更は0。
- Changed files: `frontend/src/features/accounting-reports/components/DailyBreakdownTable.tsx`、`MonthlySummaryCards.tsx`、`frontend/src/features/accounting/components/AccountingDocument.tsx`、`AccountingItemRow.tsx`、`ItemListCard.tsx`、`OwnerAccountingHistoryParts.tsx`、`PaymentCard.tsx`、`RefundSection.tsx`、`frontend/src/features/accounting/hooks/use-accounting-settlement-actions.ts`、`frontend/src/features/cash-register/components/BillingDetailTable.tsx`、`UnifiedClosingSummaryTable.tsx`、`frontend/src/features/master/components/cage-side-panel-model.ts`、`TrimmingTabRows.tsx`、`frontend/src/features/medical-records/components/TreatmentsTab/TreatmentRow.tsx`、本ledger。15-path allowlist内のみ。
- Saved Prompt Validation Gate: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-10-currency-formatter-consolidation.md` → `Prompt Craft Harness Validation: PASS` / `Profile: standard (declared-risk-tier)` / `Target: agent (detected)` / `PROMPT_VALIDATOR_EXIT=0`。SHA-256=`7af1d159d5eb906ca2bebaf91c4cc429e1495db3cf58a65b6f504607aa530d1f`。
- Binding pre-read: saved prompt、`AGENTS.md`、`.claude/CLAUDE.md`、`.claude/rules/claude-code-usage.md`、`.claude/rules/go-gin-backend-guidelines.md`、`.claude/refs/typescript-react.md`、`.claude/refs/accessibility-rules.md`、`frontend/CLAUDE.md`、`frontend/src/features/CLAUDE.md`、`docs/product-philosophy.md`、`.codex/config.toml`、planner/reviewer/TypeScript reviewer role TOML、FE12-10 row / FE12-08 ledger、`frontend/src/lib/format/number.ts`、14 target source、`~/.agents/skills/verification-loop/SKILL.md`、`~/.agents/skill-library/react-testing/SKILL.md`。
- Baseline census: CENSUS block → `A5=24 FFE5=0 EM=10 MINUS=1 TLS=25 FCOD=2 DASH=10 FC=32`（required Before完全一致）。
- Final census: CENSUS block → `A5=0 FFE5=0 EM=10 MINUS=1 TLS=1 FCOD=2 DASH=8 FC=56`（required After完全一致）。
- Baseline scoped Vitest: prompt指定11 path → per-file `UnifiedClosingSummaryTable 2` / `MonthlySummaryCards 4` / `DailyBreakdownTable 5` / `AccountingItemRow 2` / `PaymentCard 13` / `TreatmentRow 4` / `RefundSection 4` / `ItemListCard 1` / `OwnerAccountingHistory 15` / `AccountingDocument 12` / `AccountingDetail 13 tests | 3 skipped`。summary=`Test Files 11 passed (11)` / `Tests 72 passed | 3 skipped (75)` / `BASELINE_VITEST_EXIT=0`。
- Final scoped Vitest: 同じ11 path・同じper-file test counts。summary=`Test Files 11 passed (11)` / `Tests 72 passed | 3 skipped (75)` / `FINAL_VITEST_EXIT=0`。baseline/final件数完全一致。
- Plan-named regression: `docker compose exec frontend npx vitest run src/features/accounting/routes/AccountingDetail.test.tsx src/features/cash-register/category-breakdown.test.ts src/features/hospitalization/components/HospitalizationCostSummary.test.tsx` → `Test Files 3 passed (3)` / `Tests 17 passed | 3 skipped (20)` / `PLAN_REGRESSION_EXIT=0`。fullwidth-yen hospitalization surfaceは変更なし。
- TLS audit: prompt指定Docker grep → `src/features/accounting/components/AccountingDocument.tsx:250:                    <span>{paymentInfo.insuranceAmount.toLocaleString()}</span>` / `TLS_GREP_EXIT=0`。除外対象1行だけ。
- Import audit: `1 DailyBreakdownTable.tsx` / `1 MonthlySummaryCards.tsx` / `1 AccountingDocument.tsx` / `1 AccountingItemRow.tsx` / `1 ItemListCard.tsx` / `1 OwnerAccountingHistoryParts.tsx` / `1 PaymentCard.tsx` / `1 RefundSection.tsx` / `1 use-accounting-settlement-actions.ts` / `1 BillingDetailTable.tsx` / `1 UnifiedClosingSummaryTable.tsx` / `1 cage-side-panel-model.ts` / `1 TrimmingTabRows.tsx` / `1 TreatmentRow.tsx`。14 pathすべてexactly 1。
- Scoped lint: prompt指定14 sourceの`docker compose exec frontend npx eslint ... > /tmp/fe12-10-eslint.txt 2>&1` → `ESLINT_EXIT=0`、ESLint diagnostics stdout 0（Composeの未設定DB env warning 3件のみ）。
- Guard classification: (a)=`cage-side-panel-model`、`TrimmingTabRows`の2件をwhole-callへcollapse。(b)=`UnifiedClosingSummaryTable` row/totals 2件で`!= null` + `—`を保持。(c)=`DailyBreakdownTable` 1、`MonthlySummaryCards` 1、`AccountingDocument` 5、`AccountingItemRow` 1、`ItemListCard` 3、`OwnerAccountingHistoryParts` 1、`PaymentCard` 2、`RefundSection` 3、`use-accounting-settlement-actions` 1、`BillingDetailTable` 1、`TreatmentRow` 1の20件でbusiness/no guardとsign/fallback/spacingを保持。
- Fullwidth exclusion / assumptions: hospitalization/medical-recordsのU+FFE5 `￥` clusterは`formatCurrency`のU+00A5と表示契約が異なるため裁定どおり変更0。ambient `.toLocaleString()`から`"ja-JP"`固定への置換はsaved promptの受入済みassumption。plannerが`OwnerAccountingHistoryParts`の`totalAmount`はlive codeで`?? 0`により常にnumberであり、promptのnull-divergence説明がstaleと確認したが、変換・Success Criteriaへの影響なし。
- De-Sloppify: production diffは24 formatter置換 + 3 importだけ。category (b)/(c) guard、`formatCurrencyOrDash` 2参照、test、comment、console/error handling、type/API、formatting、helper、他site変更0。cleanup patch不要。source census・同じ11-file Vitest・lintを最終状態で実測済み。
- Independent Review: general reviewerとTypeScript reviewerのfresh read-only 2 passはいずれもCRITICAL/HIGH/MEDIUM 0、Approve。全24 occurrence、3 import、glyph/sign/spacing/guard、scopeを再照合。`BillingDetailTable.tsx:60`は`detail.refundAmount > 0 ? \`-${formatCurrency(detail.refundAmount)}\` : "—"`でASCII sign、business guard、U+2014 fallbackを明示確認。TypeScript reviewerはfull type-check成功を主張していない。
- Harness / orchestration: `eval` + `sequential`（De-Sloppify overlay）を実運用。`verification-loop`と`react-testing`を読了し、project禁止のfull build/type/test/lint phaseはnarrow prompt gatesへoverride。planner subagentの24-site category planを実装前にacceptし、main agentがsingle edit pass + reconciliation、general/TypeScript reviewerがIndependent Reviewを担当。stop condition=全Acceptance Checklist PASS。
- Allowlist attribution: `comm -13 "${TMPDIR:-/tmp}/fe12-10-baseline.txt" "${TMPDIR:-/tmp}/fe12-10-final.txt"` → attempt 2で14 sourceの` M`行だけ（上記source全14 path）、全件allowlist内。`FE-refactor.md`は開始時から他owner WIPで` M`だったためadded status lineには現れないが、本unit所有変更はFE12-10 row + 本ledgerだけ。`comm -23 ...` → ` M BE-refactor.md` / ` M backend/internal/pet/chronic_condition_repository.go` / ` M backend/internal/pet/chronic_condition_repository_test.go` / `?? backend/internal/testdb/fixtures.go`が消失。並行ownerのexternal driftとして記録し、復元・編集していない。
- Failure Signature log: attempt 1 / allowlist external drift — expected=`comm -13`追加行が15-path allowlist内のみ、actual=並行sessionの` M todo.md`が一時的に追加、verification=final baseline-attribution、error signature=`todo.md` outside allowlist、attempted fix=source/doc変更・復元を行わずparallel WIP settle後に同一gateを再取得、result=attempt 2は14 source行だけでPASS。同一signature再発なし。
- Coverage / type / staging: behavior追加・test追加0のため新規coverage obligationなし。prompt禁止の全体`pnpm type-check` / `pnpm test:run` / `pnpm build` / lint / installとcoverage artifactは未実行し、type checkingはuser/CIへdefer。commit予定artifactなし、stage/commitなしのためtracked-or-not-ignored probeとstaged-path listingはnot applicable。
- Prompt-defect / Harness Improvement Feedback: P2 — `~/.agents/skills/code-review/`は存在するが`SKILL.md`がなく、stack-matching review skillとしてload不能。TypeScript reviewer role + main fresh diff reviewへfallback。P2 — `OwnerAccountingHistoryParts` null-divergence説明はlive codeの`totalAmount = ... ?? 0`とdrift。いずれもchecklist達成を阻害せず、eval captureは生成側session所掌。
- Remaining risk / follow-up: repo-wide type-check/build/full suite/coverageは未実施。ambient locale assumptionはsaved promptで受入済み。変更はunstaged・未commit、他FE12 unitとbackend/API/DB/clinic isolationは未変更。

<!-- FE12-02-REVIEW-END -->

### FE12-07 実行 ledger（2026-07-25）

- Status: COMPLETE — `backend/tygo.yaml` の `type_mappings` へ未マップ型4種を追加し `make codegen` を実行、`frontend/src/types/generated/models.ts` の `any` を **17→0** にした。生成物の手編集は0。consumer修正は0件。
- Product gate: 新機能追加ではなく `frontend/CLAUDE.md` Type Safety MANDATORY（`any` 禁止）への適合。②削除＝`any` 17件と、それを前提にした防御的cast。追加は mapping 4行×3ブロックのみで新helper・新型・新schemaは0。
- Changed files: `backend/tygo.yaml`、`frontend/src/types/generated/models.ts`（生成物）、本ledger + task表行。
- 追加した mapping（3ブロック共通・model/auth/trimming）:
  - `json.RawMessage: "unknown"`（6件）— 非構造化JSON。`any` のままだと型検査が嘘をつくため境界検証を消費側へ強制する
  - `datatypes.JSON: "unknown"`（4件）— 同上
  - `uuid.UUID: "string"`（6件）— wire上はただの文字列
  - `pq.Int64Array: "number[]"`（1件）— 既存 `pq.StringArray: "string[]"` の対
- 生成結果（`git diff` 実測）: `old_value`/`new_value`/`metadata`/`category_breakdown`/`customer_fields`/`dose_param_snapshot` → `unknown`、`options`/`error_log`/`tags`/`scenarios` → `unknown`、`job_id`/`id`/`csv_import_id` → `string`、`closed_weekdays` → `number[]`。合計17件。
- Any残存gate: `grep -nE ":\s*any\b|<any>|any\[\]" frontend/src/types/generated/*.ts` → 出力0行・exit 1（models.ts / auth-responses.ts / trimming-responses.ts の3生成物すべて）。
- Consumer影響の全数調査（10フィールド × 3アプリツリー grep・test除外）:
  - 消費者ゼロ = `old_value` / `new_value` / `metadata` / `error_log` / `scenarios` / `dose_param_snapshot` / `LstepFriendAttributeSnapshot.tags`
  - 既に `unknown` 受け＋型ガード済 = `customer_fields`（`reception/api/transforms.ts:19` と `lib/transforms/reservation.ts:26` がともに `(raw: unknown)` で `typeof`/`Array.isArray` 絞り込み）、`category_breakdown`（`lib/transforms/cash-register.ts:14` が既に `as unknown`）
  - **同名だが別型** = `options`（消費側は `checkups/api/get-checkup-type-fields.ts` の feature-local `CheckupTypeFieldApi` が正本と明記。生成型 `CheckupTypeField` は未参照）、`tags`（lstep API response型とcomponent propsであり生成型ではない）
  - 結論: `unknown` 化による破壊は0件。`any` は誰にも使われていない虚偽の型だった。
- **ポインタmappingは機能しない（発見）**: 初回は `*uuid.UUID` 等4件も追加したが、生成結果は `job_id?: string`（`string | null` ではない）で不一致。`grep -c "string | null" models.ts` → **0** であり、既存の `*string` / `*uint64` / `*bool` / `*time.Time` / `*float64` の5件も含めポインタ指定は1つも一致していない。tygo はポインタを剥がして `?:` optional へ変換する。追加した4件は証拠上deadのため削除済み（推測実装の撤回）。**既存5件×3ブロック=15行も同様にdead configであり、削除は別unit候補**（削除自体は出力不変だが `make codegen` 再実行での確認が要る）。
- **未コミットの付随ドリフト（本unitの変更ではない）**: 再生成により audit定数7件（`AuditActionAuthPasswordChange`/`Reset`/`AdminReplace`、`AuditActionTrimmingCreate`/`Update`/`Delete`、`AuditResourceAccount`/`AuditResourceTrimming`）と `TokenBlacklist` のdoc commentが models.ts へ追随した。**Go model変更後に `make codegen` が回されずgenerated型が陳腐化していた**ことを意味する。CI の `make codegen-check` は本来これを検出する。プロセス側の指摘であり本unitでは是正しない。
- Type-check gate（USER実行・禁止コマンドのため当方は未実行）: `docker compose exec frontend pnpm type-check` → `tsc --noEmit` 無出力・エラー0（2026-07-25 実行済）。consumer全数調査の「破壊0件」が実測で裏付けられた。
- **訂正（2026-07-25・FE12-13 実行中に判明）**: 本unitの列挙は**不完全だった**。`frontend/src/features/cash-register/routes/CashRegisterClosePage.tsx:230-233` に `¥{(...\n...\n).toLocaleString()}` の**複数行**inline formatterが残存しており、14ファイルのallowlistに入っていなかった。原因は着手時の列挙に使った grep が**同一行パターン**（`¥[^ ]*toLocaleString` 等）で、行を跨ぐ式を構造的に拾えなかったこと。上記 `A5=0` / 24箇所という数値は**allowlist内では正しいが、allowlist自体が1件取りこぼしていた**。census gateは自分が列挙した範囲しか検査しないため、列挙漏れは漏れごと通過する。是正は別unit（`/accounting/close` を `formatCurrency` へ収束）。FE12-13 の該当cellは `FE12-10` のまま保持されている。
- 上記訂正を除き全gate green。残unitは FE12-02 U10（F16決裁待ちでBLOCKED）と FE12-03（`pnpm build` chunk実測が前提）のみ。

### FE12-03 判定 ledger（2026-07-25・着手しない）

- Status: **不要と判定**。plan行が着手条件に据えていた「Viteのchunk実測」を `docker compose exec frontend pnpm build`（USER実行・vite 8.0.16 / rolldown）で取得した結果、前提が否定された。
- 実測: `dist/assets/vendor-icons-CGKOWK_u.js` = **28.00 kB / gzip 9.10 kB**。lucide-react は既に専用vendor chunkへ分離・tree-shake済みで、208ファイルのroot barrel importはバンドルを膨らませていない。
- 対比（同ビルドの上位chunk）: `manual` 522.69 kB / gzip 145.77 kB（500 kB警告の唯一の該当）、`vendor-charts` 376.30 kB / gzip 110.40 kB、`vendor-react` 274.12 kB / gzip 87.43 kB、`master` 238.63 kB / gzip 50.51 kB。**icon chunkは実バンドル問題の1/16以下**。
- 判定根拠: 「208ファイルのimport書き換え」は保守コスト・レビュー面積・regression表面積を確実に増やす一方、削減見込みは gzip 9.10 kB の一部にすぎない。①要件を疑う→存在すべきでない工程であり、②削除でなく③最適化を先に走らせる典型的な誤りにあたる。**着手しない**。
- New Work（本行の代替として起票候補・いずれも別unit）:
  - `manual` chunk 522.69 kB が唯一 500 kB 警告に該当。分割余地の調査が実バンドル改善の入口。
  - `[INEFFECTIVE_DYNAMIC_IMPORT]` 2件 — `src/features/auth/index.ts`（`app-routes.tsx`/`Sidebar.tsx` が動的importする一方 `src/app/router.tsx` が静的import）、`src/components/shared/MasterSelectModal/index.ts`（`TrimmingLazyModals.tsx` が動的・`TrimmingLeftColumn.tsx` が静的）。**lazy loadingを書いたのに分割されていない＝見せかけの最適化**で、②削除または静的import一本化の対象。

### FE12-11 実行 ledger（2026-07-25）

- Status: COMPLETE — line-reserve の webfont が一度も読み込まれていなかったライブ不具合を是正した。FE12-03 の chunk 実測中にビルド警告から発見。
- 発見の経緯: `pnpm build` が `Found 1 warning while optimizing generated CSS: @import rules must precede all rules aside from @charset and @layer statements` を出力。該当は `frontend/line-reserve/src/index.css:2` の Google Fonts `@import` で、1行目の `@import "tailwindcss"` が大量のルールへ展開されるため CSS 仕様条件を満たさず**ブラウザに無視されていた**。
- 実害の確定: `line-reserve/index.html` にはメインアプリ (`frontend/index.html:13-15`) のような `<link rel="stylesheet">` フォールバックが無く、`grep` 実測で 0 件。したがって line-reserve では Montserrat も Noto Sans JP も webfont として一度も読まれておらず、`index.css:20` のフォールバック連鎖（`-apple-system` / `BlinkMacSystemFont` / `Hiragino Sans`）でOS標準フォントに落ちていた。**LINE予約はモバイル専用の顧客向け画面**であり、iOS利用者は全員Hiraginoで見ていた。
- Montserrat の裁定 = **削除**（修正しない）: `DESIGN.md` および `docs/spec/design-system.md` に Montserrat の規定は **0件**（grep実測）。FE11 で「タイポグラフィ・形状・寸法の正本は DESIGN.md」と確定済みであり、仕様外フォント。利用は `TopPage.tsx:15` の h1 1箇所のみで、しかも一度もレンダリングされていない。壊れた `@import` を1行目へ移動して「読ませる」のは、存在すべきでないものを機能させる②違反にあたるため採らない。責任者不在の要件は要件ではない（①）。
- Noto Sans JP の裁定 = **実効化**: `index.css:20` が base `font-family` として宣言済みであり、メインアプリ `frontend/index.html:15` も `<link>` で読んでいる製品全体の日本語フォント。宣言と実装の乖離＝バグであり是正する。CSS `@import` ではなく `<link>` を採ったのは、①同一リポジトリで2方式が並立するドリフトを避ける（メインアプリの既存規約に合わせる）②CSS `@import` は直列取得でレンダリングを遅らせる、の2点。
- Changed files: `frontend/line-reserve/src/index.css`（`@import` 1行削除）、`frontend/line-reserve/src/pages/TopPage.tsx`（`style={{ fontFamily: ... }}` 削除）、`frontend/line-reserve/index.html`（preconnect 2行 + stylesheet link 1行）、本ledger + task表行。
- 削除/追加の収支: 削除=dead `@import` 1行・仕様外 inline style 1件。追加=`<link>` 3行。**追加だけの変更ではない**（②）。
- Gate: `grep -rn "Montserrat" frontend/src frontend/liff frontend/line-reserve` → 0行・exit 1（消滅確認）。`docker compose exec frontend npx eslint line-reserve/src/pages/TopPage.tsx` → `ESLINT_EXIT=0`。`TopPage.tsx` に併置テストは存在しない（`npx vitest run` が `No test files found`）。
- Build gate（USER実行・2026-07-25 実行済）: `docker compose exec frontend pnpm build` → **CSS警告 `@import rules must precede all rules` が消滅**。`dist/line-reserve/index.html` 1.23→1.81 kB（link 3行）、`line-reserve` JS 49.25→49.19 kB（inline style 削除）。
- **@import が死んでいたことの決定的証拠**: 修正前後で `dist/assets/line-reserve-Bn6pV4Pv.css` の**内容ハッシュが完全一致**（90.77 kB / gzip 17.06 kB も同値）。`@import` 行を削除しても出力CSSが1バイトも変わらない＝あの行はビルド出力に一切寄与していなかった。仮に有効だったなら `@import` は出力CSSへ残りハッシュが変わるはずである。
- 残gate: **line-reserve の実機目視**を推奨する。加えて **line-reserve の実機目視**を推奨する。本修正はレンダリングを「意図された状態」へ変える＝顧客向け画面の見た目が実際に変わるため、意図どおりかの確認は人間の目が要る。

### FE12-12 実行 ledger（2026-07-25）

- Status: **COMPLETE** — barrel経由の静的巻き込み2件を所有concrete moduleへのimportへ置換し、書かれているだけで実際には分割されていなかったlazy splitを実効化した。未コミット。
- Product philosophy: 既存のlazy intentと公開APIを増やさず、見せかけの最適化経路と不要な`Suspense`だけを削除した（②削除→③簡素化・最適化）。props、handler、ユーザー表示文字列は変更していない。
- Changed files: `frontend/src/components/shared/Layout/Sidebar.tsx`、`frontend/src/components/shared/Layout/Sidebar.test.tsx`、`frontend/src/features/trimming/components/TrimmingLeftColumn.tsx`、`FE-refactor.md`。
- Edit A before: `const ChangePasswordDialog = lazy(() => import("@/features/auth").then((m) => ({ default: m.ChangePasswordDialog })))`。after: `import { ChangePasswordDialog } from "@/components/shared/ChangePasswordDialog/ChangePasswordDialog";`。追随してtest mockも同じconcrete moduleへretargetした。
- Edit B before: `import { MasterSelectTrigger } from "@/components/shared/MasterSelectModal";`。after: `import { MasterSelectTrigger } from "@/components/shared/MasterSelectModal/MasterSelectTrigger";`。
- Suspense decision: **removed**。変更前の`<Suspense fallback={null}>`は`ChangePasswordDialog`だけをwrapしており、static import化後にsuspending childが残らないためdead boundary。
- Build before: `BUILD_BEFORE_EXIT=0`。`dist/assets/MasterSelectModal-Bj1zaZ-t.js  0.07 kB │ gzip: 0.08 kB`。warningは `src/features/auth/index.ts is dynamically imported by src/app/routes/app-routes.tsx, src/components/shared/Layout/Sidebar.tsx but also statically imported by src/app/router.tsx` と `src/components/shared/MasterSelectModal/index.ts is dynamically imported by src/features/trimming/routes/TrimmingLazyModals.tsx but also statically imported by src/features/trimming/components/TrimmingLeftColumn.tsx` の2件。
- Build after: `BUILD_AFTER_EXIT=0`。`dist/assets/MasterSelectModal-BlrCJxxL.js  2.47 kB │ gzip: 1.23 kB`。MasterSelectModal warningは消滅。残warningは `src/features/auth/index.ts is dynamically imported by src/app/routes/app-routes.tsx but also statically imported by src/app/router.tsx` のみで、Sidebarを含まない。
- Tests: before `Test Files 3 passed (3)` / `Tests 14 passed (14)`。Edit A+mock retarget後、Edit B後とも `Test Files 3 passed (3)` / `Tests 14 passed (14)`。test caseの追加・削除なし。
- Lint/reference gates: `ESLINT_EXIT=0`（source 2ファイル、diagnosticなし）。`grep -n "@/features" frontend/src/components/shared/Layout/Sidebar.tsx` → 0行・exit 1。`grep -rn 'import("@/features' frontend/src/components frontend/src/hooks frontend/src/lib --include="*.ts" --include="*.tsx"` → 0行・exit 1。`TrimmingLeftColumn.tsx:10`はconcrete `MasterSelectTrigger` importのみ。
- Assumption deviations / Failure Signature: none。coverage追加義務なし。full-project type-checkは規約どおり未実行。generated artifact、stage、commit、pushなし。
- Escalated follow-up（別unit）: `frontend/src/app/router.tsx:5`の`AuthProvider` static importが`@/features/auth` barrelをeagerに保つため、`app-routes.tsx`のLogin/ForgotPassword/ResetPassword lazy splitは引き続き実効化されない。feature deep-import禁止により単純なpath差替えは不可。`AuthProvider`の公開境界再配置はarchitecture decisionとして別unitで扱う。

### FE12-13 実行 ledger（2026-07-25）

- Status: **PARTIAL / BLOCKED** — current codeをroute単位で再検証し、`FE12-03` 57セルを`🔒`、適合を確認した`FE12-04` 41セル・`FE12-05` 1セル・`FE12-06` 4セル・`FE12-10` 7セルを`✅`へ置換した（計110セル）。`FE12-01` 8セルは必須`design-audit`がexit 1、`/accounting/close`の`FE12-10` 1セルはinline halfwidth yen formatter残存のためtask IDを保持した。source修正は本unitのscope外。
- Changed files / write set: `FE-refactor.md`のみ。本unitによるsource、test、index、HEAD変更はなく、stage/commit/stash/push/mergeは0。
- Saved Prompt Validation Gate: `node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-13-route-table-reconciliation.md` → `Prompt Craft Harness Validation: PASS`、`VALIDATOR_EXIT=0`。
- Distribution before:
  ```text
     8 axis1 FE12-01
    25 axis1 △ FE12-02
    57 axis2 FE12-03
     7 axis2 FE12-04
     1 axis2 FE12-05
    34 axis3 FE12-04
     4 axis3 FE12-06
     8 axis3 FE12-10
  ```
- Distribution after:
  ```text
     8 axis1 FE12-01
    25 axis1 △ FE12-02
     1 axis3 FE12-10
  ```
- Field count before: broad prompt command=`85 10`、`13 11`、`24 9`; marker限定route table=`85 10`。After: broad prompt command=`85 10`、`14 11`、`24 9`; marker限定route table=`85 10`。broad commandは必須FE12-13 backlog row追加により11-field rowが13→14へ増えるためbyte-identicalにはならない（prompt内gateと必須row追加の自己矛盾）。
- Design audit: `docker compose exec frontend pnpm design-audit` → `AUDIT_EXIT=1`、`design-system-audit: C18 table cell override — 9 件`、`design-system-audit: C18 raw legacy baseline — 10 件（non-gating ratchet）`、`design-system-audit: FAIL — 9 件の違反`。9件はすべて`frontend/src/features/examinations/components/ExamPivotTable.tsx:153-347`。必須exit 0を満たさないためFE12-01 8セルは未反映。
- **FE12-01 blockerの性質（2026-07-25 Mode 3 照合で確定）**: この9件は `bc4fe88cb feat(exams): #249 Phase 1` が追加した **別作業の新規ファイル**にあり、FE12-01 の8セルが指すroute（`/owners/:id/report`、`/accounting/new`、`/accounting/:id`、`/accounting/close`、`/accounting/close/history`、`/accounting/reports`、`/lstep/analytics`、`/settings/closing-time`）には examinations が1つも含まれない。したがって **8セルは「非準拠」ではなく「未判定」**である。原因は指示側のgate設計で、8 routeという狭い主張の証拠に**リポジトリ全体の機械gate**を課したため、無関係な箇所の違反で道連れにブロックされた。**gateの観測範囲は主張の範囲に一致させること**（本セッションで繰り返した同型ミス — 禁止パスの不可侵をadded-setで証明しようとした件と同根）。再判定はroute単位の証拠へ切り替えて行う。`ExamPivotTable.tsx` の C18違反9件の是正は #249 担当の領分。
- FE12-03 ruling: ledgerの`Status: **不要と判定**`、`vendor-icons` 28.00 kB / gzip 9.10 kBでtree-shake済み、`**着手しない**`を根拠に、complianceを意味する`✅`でなく裁定済み例外の`🔒`を57セルへ適用した。
- FE12-04 shared basis: `frontend/src/hooks/use-unsaved-changes.ts:15-26`が唯一の`beforeunload`購読、`frontend/src/hooks/use-side-peek-dirty.ts:25-26`が同hookを合成する。下表の34 routes / 41 cellsを個別確認し、推論のみの未確認routeは0。

| FE12-04 route | Axis | Current-code citation |
|---|---|---|
| `/inventory/new` | ③ | `frontend/src/features/inventory/routes/InventoryForm.tsx:51` |
| `/inventory/:id` | ③ | `frontend/src/features/inventory/routes/InventoryForm.tsx:51` |
| `/medical-records/new` | ③ | `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:118` |
| `/medical-records/:id` | ③ | `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:118` |
| `/hospitalization/new` | ③ | `frontend/src/features/hospitalization/routes/HospitalizationForm.tsx:73` |
| `/hospitalization/:id/edit` | ③ | `frontend/src/features/hospitalization/routes/HospitalizationForm.tsx:73` |
| `/trimming/new` | ③ | `frontend/src/features/trimming/routes/TrimmingForm.tsx:92` |
| `/trimming/:id` | ③ | `frontend/src/features/trimming/routes/TrimmingForm.tsx:92` |
| `/examinations/new` | ②/③ | `frontend/src/features/examinations/routes/ExaminationForm.tsx:71` |
| `/examinations/:id` | ②/③ | `frontend/src/features/examinations/routes/ExaminationForm.tsx:71` |
| `/vaccinations/new` | ③ | `frontend/src/features/vaccinations/routes/VaccinationForm.tsx:53` |
| `/vaccinations/:id` | ③ | `frontend/src/features/vaccinations/routes/VaccinationForm.tsx:53` |
| `/owners/new` | ③ | `frontend/src/app/pages/OwnerFormPage.tsx:89-93` → `frontend/src/features/owners/routes/OwnerForm.tsx:57` |
| `/owners/:id` | ③ | `frontend/src/app/pages/OwnerFormPage.tsx:89-93` → `frontend/src/features/owners/routes/OwnerForm.tsx:57` |
| `/settings/staff` | ③ | `frontend/src/features/master/routes/StaffSettings.tsx:106` |
| `/settings/treatment-items` | ②/③ | `frontend/src/features/master/routes/TreatmentPlanMaster.tsx:69` |
| `/settings/diagnosis` | ②/③ | `frontend/src/features/master/routes/DiagnosisSettings.tsx:54` |
| `/settings/animal-species` | ②/③ | `frontend/src/features/master/routes/AnimalSpeciesSettings.tsx:34` |
| `/settings/trimming` | ③ | `frontend/src/features/master/routes/TrimmingSettings.tsx:50` |
| `/settings/trimming-course-type` | ③ | `frontend/src/features/master/routes/TrimmingCourseTypeSettings.tsx:48` |
| `/settings/medicine` | ②/③ | `frontend/src/features/master/routes/MedicineSettings.tsx:58` |
| `/settings/reservation-type` | ③ | `frontend/src/features/master/routes/ReservationTypeSettings.tsx:69` |
| `/settings/hospitalization` | ③ | `frontend/src/features/master/routes/HospitalizationSettings.tsx:37` |
| `/settings/cage` | ③ | `frontend/src/features/master/routes/CageSettings.tsx:31` |
| `/settings/merchandise-items` | ③ | `frontend/src/features/master/routes/MerchandiseItemSettings.tsx:36` |
| `/settings/insurance` | ③ | `frontend/src/features/master/routes/InsuranceSettings.tsx:38` |
| `/settings/occupations` | ③ | `frontend/src/features/master/routes/OccupationSettings.tsx:40` |
| `/settings/permission-groups` | ③ | `frontend/src/features/master/routes/PermissionGroupSettings.tsx:54` |
| `/settings/inquiry-templates` | ③ | `frontend/src/features/master/routes/InterviewTemplateSettings.tsx:50` |
| `/settings/interview/chief-complaint` | ③ | `frontend/src/features/master/routes/ChiefComplaintSettings.tsx:41` |
| `/settings/interview/templates` | ③ | `frontend/src/app/routes/settings-routes.tsx:231-233` → `frontend/src/features/master/routes/InterviewTemplateSettings.tsx:50` |
| `/settings/shift-templates` | ②/③ | `frontend/src/features/shifts/routes/ShiftTemplateSettings.tsx:44` |
| `/settings/payment-methods` | ③ | `frontend/src/features/master/routes/PaymentMethodSettings.tsx:48` |
| `/settings/campaigns` | ③ | `frontend/src/features/master/routes/CampaignSettings.tsx:44` |

- FE12-05: `/hospitalization/:id`は`HospitalizationDetail`→`DailyRecordsTab`へ到達し、`frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx:58-59`が`effectiveMax`/`initialDate`をplain constで計算する。`rg '\buseMemo\s*\(|^import.*\buseMemo\b'`は0件（説明commentのtokenだけ残存）。1セルを`✅`。
- FE12-06:

| Route | Current-code citation | Forbidden result |
|---|---|---|
| `/estimates` | `frontend/src/features/estimates/routes/EstimateList.tsx:26,196` → `../lib/is-estimate-locked-status` | `estimates/utils` / old helper path 0件 |
| `/estimates/new` | `frontend/src/features/estimates/routes/EstimateForm.tsx:31-38,294` → `../constants` / `../lib` | 同上 |
| `/estimates/:id` | `frontend/src/features/estimates/routes/EstimateDetail.tsx:17,45-46`（expiry inline） | 同上 |
| `/estimates/:id/edit` | `frontend/src/features/estimates/routes/EstimateForm.tsx:31-38,294` | 同上 |

  `find frontend/src frontend/liff/src frontend/line-reserve/src -type d -name utils -print`と`rg 'estimates/utils|utils/estimate-status-options|utils/is-estimate-locked-status|utils/is-estimate-expired|isEstimateExpired'`はいずれもstdout 0行。4セルを`✅`。
- FE12-10:

| Route | Evidence / result | Applied |
|---|---|---|
| `/accounting` | `frontend/src/features/accounting/components/AccountingListTable.tsx:24,256`; forbidden pattern 0件 | `✅` |
| `/accounting/new` | `frontend/src/features/accounting/components/ItemListCard.tsx:17,311-314`; forbidden pattern 0件 | `✅` |
| `/accounting/:id` | 同上、`frontend/src/features/accounting/components/RefundSection.tsx:15,62-189`; forbidden pattern 0件 | `✅` |
| `/accounting/close` | `frontend/src/features/cash-register/routes/CashRegisterClosePage.tsx:230-233`にmultiline `¥{(...).toLocaleString()}`残存 | `FE12-10`保持 |
| `/accounting/close/history` | `frontend/src/features/cash-register/routes/CashRegisterHistoryPage.tsx:17,229-329`; literal yen + formatter 0件 | `✅` |
| `/accounting/reports` | `frontend/src/features/accounting-reports/components/DailyBreakdownTable.tsx:5,74-86`; forbidden pattern 0件 | `✅` |
| `/medical-records` | direct list routeにhalfwidth yen + formatter 0件。`TreatmentRow.tsx:11,304,401`はhelper使用 | `✅` |
| `/hospitalization/:id` | halfwidth yen + formatter 0件。fullwidth `￥` clusterはFE12-10 ledgerの明示的除外contract | `✅` |

- Individually checked vs inferred: FE12-04=41/41 cells current consumerを個別確認（同一componentを共有するrouteは同一citationを再利用）、FE12-05=1/1、FE12-06=4/4、FE12-10=8/8。FE12-01は8/8を一括する必須machine gateを実行したが非greenのため未反映。FE12-03は57/57をcell列挙でcountし、裁定ledgerを共通根拠として適用。未申告sampling・推論flipは0。
- `△ FE12-02`: before/afterとも`25 axis1 △ FE12-02`。axis cell差分で25セルがbyte-identicalであることを確認した。
- Failure Signature log:
  - attempt 1 / FE12-01 C18 gate — expected=`AUDIT_EXIT=0`、actual=`AUDIT_EXIT=1`, C18 override 9件、verification=`docker compose exec frontend pnpm design-audit`、error=`ExamPivotTable.tsx:153-347`、attempted fix=なし（source editはunit外）、result=8 cells保持 / BLOCKED。required external change=9 C18 violationの別unit是正後に再実行。
  - attempt 1 / FE12-10 scan — expected=8 routesでliteral halfwidth yen + `.toLocaleString()` 0件、actual=同一行regexは0件だが独立reviewが`CashRegisterClosePage.tsx:230-233`のmultiline残存を検出、verification=`rg -n -U -P '¥(?s:.{0,240}?)\.toLocaleString\(' .../CashRegisterClosePage.tsx`、attempted fix=一時flipを撤回しcellを`FE12-10`へ復元、result=虚偽`✅`を回避 / 1 cell保持。
- De-Sloppify: ledger proseは既存unit背景の再掲を避け、counts、current-code citations、FE12-03 ruling、blockerだけを保持。code/test変更0のためtest cleanup非該当。
- Harness / loop: `santa` + `verification-loop`を実際に読了し、sequential single passを採用。stop conditionは全checklist PASSまたはgenuine blocker。deterministic gate後のfresh dual reviewer B/Cは双方PASS、critical issue 0。Bは初回同一行currency grepが見逃したmultiline残存を独立source auditで再確認し、Cはroute非axis変更0・25 held cell不変・marker各1を再確認した。レビュー前の実装passが見逃し、adversarial evidence passで検出した事項は`/accounting/close`のmultiline formatter 1件で、一時flipを撤回済み。long-running loopではないためloop monitor非該当。
- Coverage/lint/test/build/type-check/security: docs-onlyでsource挙動変更0。prompt禁止のfull gateは未実行し、coverage thresholdを主張しない。機密情報・個人/臨床データは追加していない。generated artifactなし、stage/commit予定なしのためtracked-or-not-ignored / staged-path probeは非該当。
- Remaining blockers / follow-ups: FE12-01はC18 audit 9件の別unit修正後に8セルを再判定する。`/accounting/close`のinline currencyは別unitでhelperへ収束後に1セルを再判定する。held `FE12-02` U10 adjudicationと、escalated `AuthProvider` barrel decisionは未着手。`manual` chunk investigationも本unit外。

### FE12 closeout — C18 primitive・通貨統合・route 9セル再判定（2026-07-25）

- **Status**: COMPLETE。`ExamPivotTable.tsx` の raw cell opening 9件を指定どおり `TableHead` / `TableCell` へ移行し、`CashRegisterClosePage.tsx` の消費税合計1箇所を既存 `formatCurrency` へ統合した。route判定表はcurrent-code traceが成立した8個の軸①と1個の軸③だけを`✅`へ変更した。
- **Changed files / write set**: `frontend/src/features/examinations/components/ExamPivotTable.tsx`、`frontend/src/features/cash-register/routes/CashRegisterClosePage.tsx`、`FE-refactor.md`。隣接test 2本はbaseline/finalともpassし、意図した表示値を変更するassertionもなかったため変更なし。`frontend/scripts/design-system-audit.mjs`、index、HEAD、remoteは変更せず、stage / commit / stash / push / mergeは0。
- **Saved Prompt Validation Gate**:
  ```text
  $ node ~/.claude/scripts/prompt-craft-harness-validate.js /Users/minoru/.claude/prompt-craft-runs/agent-fe12-14-c18-closeout-and-route-readjudication.md
  Prompt Craft Harness Validation: PASS
  VALIDATOR_EXIT=0
  ```
- **Binding pre-read**:
  - `AGENTS.md`
  - `.claude/CLAUDE.md`
  - `.claude/rules/claude-code-usage.md`
  - `.claude/rules/go-gin-backend-guidelines.md`
  - `~/.agents/codex/AGENTS.md`
  - `.codex/config.toml`
  - `frontend/CLAUDE.md`
  - `frontend/src/features/CLAUDE.md`
  - `docs/spec/design-system.md`
  - `DESIGN.md`
  - `docs/spec/ui-design-compliance.md`
  - `docs/product-philosophy.md`
  - `FE-refactor.md`
  - `~/.agents/skills/tdd-workflow/SKILL.md`
  - `~/.agents/skills/react-testing/SKILL.md`
  - `~/.agents/skills/verification-loop/SKILL.md`
- **Step 0 scoped baseline**: 指定5 pathへの`git status --porcelain ... | sort`はstdout 0行。source 4 pathを含めdirty pathは0で、他owner差分への上書きなし。
- **RED — design audit**:
  ```text
  AUDIT_BEFORE_EXIT=1
  design-system-audit: C18 table cell override — 9 件
    src/features/examinations/components/ExamPivotTable.tsx:153: <td className={`min-w-24 border-l p-2 text-center ${C.borderMedium}`}>
    src/features/examinations/components/ExamPivotTable.tsx:167: <td
    src/features/examinations/components/ExamPivotTable.tsx:306: <th scope="col" className="min-w-36 p-2 text-left">
    src/features/examinations/components/ExamPivotTable.tsx:309: <th scope="col" className={`min-w-20 border-l p-2 ${C.borderMedium}`}>
    src/features/examinations/components/ExamPivotTable.tsx:312: <th scope="col" className={`min-w-24 border-l p-2 ${C.borderMedium}`}>
    src/features/examinations/components/ExamPivotTable.tsx:316: <th
    src/features/examinations/components/ExamPivotTable.tsx:333: <th scope="row" className="p-2 text-left font-medium">
    src/features/examinations/components/ExamPivotTable.tsx:344: <td className={`border-l p-2 text-center ${C.borderMedium} ${C.text60}`}>
    src/features/examinations/components/ExamPivotTable.tsx:347: <td className={`border-l p-2 text-center ${C.borderMedium} ${C.text60}`}>
  design-system-audit: C18 raw legacy baseline — 10 件（non-gating ratchet）
  design-system-audit: FAIL — 9 件の違反
  ```
- **RED — raw cell / inline currency**:
  ```text
  153:      <td className={`min-w-24 border-l p-2 text-center ${C.borderMedium}`}>
  167:    <td
  306:                <th scope="col" className="min-w-36 p-2 text-left">
  309:                <th scope="col" className={`min-w-20 border-l p-2 ${C.borderMedium}`}>
  312:                <th scope="col" className={`min-w-24 border-l p-2 ${C.borderMedium}`}>
  316:                  <th
  333:                  <th scope="row" className="p-2 text-left font-medium">
  344:                  <td className={`border-l p-2 text-center ${C.borderMedium} ${C.text60}`}>
  347:                  <td className={`border-l p-2 text-center ${C.borderMedium} ${C.text60}`}>
  RAW_CELL_CHECK=FAIL
  230:                      ¥{(
  231:                        preview.aggregate.taxBreakdown.standard.taxAmount +
  232:                        preview.aggregate.taxBreakdown.reduced.taxAmount
  233:                      ).toLocaleString()}
  ```
- **Scoped test baseline**:
  ```text
  Test Files  2 passed (2)
       Tests  12 passed (12)
  BASELINE_TESTS_EXIT=0
  ```
  内訳は`ExamPivotTable.test.tsx`=`10 tests`、`CashRegisterClosePage.test.tsx`=`2 tests`。
- **Exam-only GREEN**:
  ```text
  design-system-audit: C18 table cell override — 0 件
  design-system-audit: C18 raw legacy baseline — 10 件（non-gating ratchet）
  design-system-audit: PASS — 違反 0 件
  Test Files  1 passed (1)
       Tests  10 passed (10)
  ```
- **Final GREEN — audit / tests / raw / currency / lint**:
  ```text
  AUDIT_AFTER_EXIT=0
  design-system-audit: C18 table cell override — 0 件
  design-system-audit: C18 raw legacy baseline — 10 件（non-gating ratchet）
  design-system-audit: PASS — 違反 0 件
  TESTS_AFTER_EXIT=0
   ✓ src/features/cash-register/routes/CashRegisterClosePage.test.tsx (2 tests)
   ✓ src/features/examinations/components/ExamPivotTable.test.tsx (10 tests)
  Test Files  2 passed (2)
       Tests  12 passed (12)
  RAW_CELL_CHECK=PASS
  CURRENCY_SCAN_EXIT=1
  CURRENCY_SCAN_STDOUT=0 lines
  ESLINT_AFTER_EXIT=0
  ESLINT_STDOUT=0 lines
  ```
  ESLint対象は変更source 2本だけ。`git status --porcelain -- frontend/scripts/design-system-audit.mjs`もstdout 0行。
- **表示・臨床意味の保持**: 列内容・順序・幅、`min-w-*`、`border-l`、`data-status`、同日検査の`aria-label`、行ヘッダ`scope="row"`を保持した。異常高値`C.bgDanger8`・異常低値`C.bgStatusBlueLight`・通常`C.bgWhite`を保持し、primitiveの基底paddingを使うため全9箇所の`p-2`を削除した。309/312はbody列との中央揃えを保つ`text-center`を指定した。333の12px/600 muted化は承認済みの意図した変更で、`text-sm` / `font-medium`を戻していない。
- **通貨契約**: `formatCurrency`は`frontend/src/lib/format/number.ts:28-31`でnon-null numberを`¥${amount.toLocaleString("ja-JP")}`として返す。合計式と0の扱いを変えず、`formatCurrencyOrDash`は使っていない。既存test 2件は前後ともpassしたが消費税合計文字列を直接assertしないため、表示同値はhelper実装と変更差分のfresh reviewも併用して確認した。
- **Route判定表 before / after**:
  ```text
  BEFORE
     8 axis1 FE12-01
    25 axis1 △ FE12-02
     1 axis3 FE12-10
  AFTER
    25 axis1 △ FE12-02
  FIELD_COUNT_BEFORE
    85 10
  FIELD_COUNT_AFTER
    85 10
  ```
- **9セルのcurrent-code trace**:

| Route | Axis | Before | After | Route → raw cell holder | Classification |
|---|---:|---|---|---|---|
| `/owners/:id/report` | ① | `FE12-01` | `✅` | `OwnerReport` → `OwnerClinicalBriefing` → `OwnerClinicalHistoryPanel` → `ClinicalHistoryMatrix.tsx:133-201` | allowlist |
| `/accounting/new` | ① | `FE12-01` | `✅` | `AccountingDetail` → `AccountingDocument.tsx:189-214` | allowlist |
| `/accounting/:id` | ① | `FE12-01` | `✅` | `AccountingDetail` → `AccountingDocument.tsx:189-214` | allowlist |
| `/accounting/close` | ① | `FE12-01` | `✅` | `CashRegisterClosePage` → `ClosePrintArea.tsx:67-194` | PrintArea |
| `/accounting/close/history` | ① | `FE12-01` | `✅` | route tableは`TableHead` / `TableCell`、route固有raw cellなし | 該当なし |
| `/accounting/reports` | ① | `FE12-01` | `✅` | `AccountingReportsPage` → `MonthlyReportPrintArea.tsx:48-138` | PrintArea |
| `/lstep/analytics` | ① | `FE12-01` | `✅` | analytics 3 sectionは`TableHead` / `TableCell`、route固有raw cellなし | 該当なし |
| `/settings/closing-time` | ① | `FE12-01` | `✅` | `ClosingSettingsPage` → `StandardClosingTimeSection`は`TableHead` / `TableCell` | 該当なし |
| `/accounting/close` | ③ | `FE12-10` | `✅` | `CashRegisterClosePage.tsx:230-233`が`formatCurrency(a + b)`、multiline forbidden scan 0 hit | 該当なし |

- **`✅`と`🔒`の区別**: 実描画経路にallowlist対象または`PrintArea`名前規則のraw cellが残るrouteも、画面用tableはprimitive準拠済みで、残るraw cellは密度レポートまたはprint帳票として画面用デザイン体系の適用対象外であるため、routeの現行適合を表す`✅`とした。`✅`は「route内のraw opening tagが物理的に0件」という意味ではない。
- **未参照baseline component**: `CheckupHistorySection.tsx`と`ExaminationHistorySection.tsx`はproduction treeで定義以外の識別子参照0件、隣接testからのみ参照され、`owner-report/index.ts`もexportしていない。C18 baseline 10件は`/owners/:id/report`の実描画経路外。削除は本unit外であり、別unitのdead-code確認候補として保持する。
- **既知follow-up**: `FE12-01` / `FE12-10` / `FE12-13`のtask backlog行は本unitの9セル解決後に記述が古くなるが、9列表のfield数破損riskを避けるため本unitでは意図的にscope外。見落としではなく既知follow-upである。`△ FE12-02` 25セル、U10、F16/F9裁定、M-01〜M-05実機実測、owner-report baseline component削除、manual chunk、AuthProvider barrelには着手していない。
- **Assumption deviations**: なし。309/312は指定どおり`text-center`、消費税合計は`¥`と3桁カンマ区切りを維持した。
- **Failure Signature log**:
  - AC-12 independent reviewer、expected=review result、actual=reviewer / React reviewerの初回2 callが同じHTTP 503で結果未生成、verification=agent status、error=`503 Service Unavailable`、attempt 1/2、attempted fix=transient service recovery待ち後に汎用reviewerを1回だけ再実行、result=attempt 3でreview完了・CRITICAL/HIGH/MEDIUM/LOWすべて0。bounded retryの3-strikeで収束し、同じcallの追加反復なし。
  - 実装gateのFAILなし。REDは期待された着手前gateであり、実装後は同一gateが1回でGREENになった。
- **De-Sloppify**: 新規test・helper・abstract・comment・console出力なし。既存testの緩和なし。指定2 sourceの最小置換と9セル、ledger以外のdrive-by変更なし。
- **Independent review**: fresh汎用reviewerを使用。列構成・順序・幅・中央揃え、`data-status`、`aria-label`、`scope="row"`、異常高/低/通常背景、通貨合計・0契約、route marker内9セルのみの変更、9 route trace、audit/test/ESLint/raw/currency/field-count/allowlist attributionを再確認し、CRITICAL 0 / HIGH 0 / MEDIUM 0 / LOW 0。React reviewerの初回callは503で結果未生成だったが、成功した汎用reviewerが同じReact/a11y確認項目をfresh passで完遂した。
- **Coverage / full gates / generated artifact**: 新規分岐なし。unit固有promptに従いcoverage閾値は主張せず、full lint / test / build / type-check / installは未実行。生成物なし、indexに触れずcommit予定もないためtracked-or-not-ignored / staged-path probeは非該当。
- **Final allowlist attribution**:
  ```text
  FINAL_SCOPED_STATUS
   M FE-refactor.md
   M frontend/src/features/cash-register/routes/CashRegisterClosePage.tsx
   M frontend/src/features/examinations/components/ExamPivotTable.tsx
  COMM_ADDED
   M FE-refactor.md
   M frontend/src/features/cash-register/routes/CashRegisterClosePage.tsx
   M frontend/src/features/examinations/components/ExamPivotTable.tsx
  CACHED_PATHS
  DIFF_CHECK_EXIT=0
  ```
  書き込み対象の自己申告3 pathと一致し、allowlist内。test 2本とaudit scriptは変更0、cached pathは0。
