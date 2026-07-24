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
| 認証/共通 | 飼主カルテレポート | /owners/:id/report | OwnerReport | FE12-01 | ✅ | ✅ | C18 raw: Checkup/Examination history |
| 認証/共通 | 404 Not Found | * | inline element | 🔒 | 🔒 | 🔒 | 対象外1件: lazy pageを持たないinline fallback |
| 会計 | 会計一覧 | /accounting | AccountingList | ✅ | FE12-03 | FE12-10 | `AccountingList.tsx:17`; 金額表示共通化 |
| 会計 | 会計 - ペット選択 | /accounting/select-pet | AccountingPetSelection | ✅ | ✅ | ✅ | 追加指摘なし |
| 会計 | 会計登録 | /accounting/new | AccountingDetail | FE12-01 | FE12-03 | FE12-10 | ItemList/Refund raw cell・通貨表現 |
| 会計 | 会計詳細 | /accounting/:id | AccountingDetail | FE12-01 | FE12-03 | FE12-10 | 同上 |
| 会計 | レジ締め | /accounting/close | CashRegisterClosePage | FE12-01 | FE12-03 | FE12-10 | Billing/summary raw cell |
| 会計 | レジ締め履歴 | /accounting/close/history | CashRegisterHistoryPage | FE12-01 | FE12-03 | FE12-10 | route raw cell `CashRegisterHistoryPage.tsx` |
| 会計 | 月次集計レポート | /accounting/reports | AccountingReportsPage | FE12-01 | FE12-03 | FE12-10 | DailyBreakdown raw cell |
| 在庫/見積/シフト | 在庫一覧 | /inventory | InventoryList | ✅ | FE12-03 | ✅ | `InventoryList.tsx:10` |
| 在庫/見積/シフト | 在庫登録 | /inventory/new | InventoryForm | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 在庫/見積/シフト | 在庫編集 | /inventory/:id | InventoryForm | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 在庫/見積/シフト | 見積一覧 | /estimates | EstimateList | ✅ | FE12-03 | FE12-06 | 禁止`utils/`をfeature indexから参照 |
| 在庫/見積/シフト | 見積作成 | /estimates/new | EstimateForm | ✅ | FE12-03 | FE12-06 | 併発: FE12-04 |
| 在庫/見積/シフト | 見積詳細 | /estimates/:id | EstimateDetail | ✅ | FE12-03 | FE12-06 | `EstimateDetail.tsx:5` |
| 在庫/見積/シフト | 見積編集 | /estimates/:id/edit | EstimateForm | ✅ | FE12-03 | FE12-06 | 併発: FE12-04 |
| 在庫/見積/シフト | シフトカレンダー | /shifts | ShiftCalendarPage | ✅ | ✅ | ✅ | 追加指摘なし |
| カルテ | カルテ一覧 | /medical-records | MedicalRecords | △ FE12-02 | FE12-03 | FE12-10 | C6a danger/異常/RBACレビュー |
| カルテ | カルテ作成 - ペット選択 | /medical-records/select-pet | MedicalRecordPetSelection | △ FE12-02 | ✅ | ✅ | C6a死亡/危険ペット選択 |
| カルテ | カルテ作成 | /medical-records/new | MedicalRecordForm | △ FE12-02 | FE12-03 | FE12-04 | 併発: C18 FE12-01、通貨 FE12-10 |
| カルテ | カルテ編集 | /medical-records/:id | MedicalRecordForm | △ FE12-02 | FE12-03 | FE12-04 | 併発: C18 FE12-01、通貨 FE12-10 |
| 入院/ホテル | 入院・ホテル一覧 | /hospitalization | HospitalizationList | △ FE12-02 | FE12-03 | ✅ | C6a死亡表示・操作抑止 |
| 入院/ホテル | 入院・ホテル登録 - ペット選択 | /hospitalization/select-pet | HospitalizationPetSelection | △ FE12-02 | ✅ | ✅ | C6a死亡/危険ペット選択 |
| 入院/ホテル | 入院・ホテル登録 | /hospitalization/new | HospitalizationForm | △ FE12-02 | FE12-03 | FE12-04 | 併発: C18 FE12-01、通貨 FE12-10 |
| 入院/ホテル | 入院・ホテル詳細 | /hospitalization/:id | HospitalizationDetail | △ FE12-02 | FE12-05 | FE12-10 | 併発: C18 FE12-01 |
| 入院/ホテル | 入院・ホテル編集 | /hospitalization/:id/edit | HospitalizationForm | △ FE12-02 | FE12-03 | FE12-04 | 併発: C18 FE12-01、通貨 FE12-10 |
| トリミング | トリミング一覧 | /trimming | TrimmingList | ✅ | FE12-03 | ✅ | `TrimmingList.tsx:12` |
| トリミング | トリミング登録 - ペット選択 | /trimming/select-pet | TrimmingPetSelection | ✅ | ✅ | ✅ | 追加指摘なし |
| トリミング | トリミング登録 | /trimming/new | TrimmingForm | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| トリミング | トリミング編集 | /trimming/:id | TrimmingForm | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 検査 | 検査一覧 | /examinations | ExaminationsList | △ FE12-02 | FE12-03 | ✅ | C6a異常値/RBAC |
| 検査 | 検査登録 - ペット選択 | /examinations/select-pet | ExaminationPetSelection | △ FE12-02 | ✅ | ✅ | C6a死亡/危険ペット選択 |
| 検査 | 検査登録 | /examinations/new | ExaminationForm | △ FE12-02 | FE12-04 | FE12-04 | global listener重複 |
| 検査 | 検査編集 | /examinations/:id | ExaminationForm | △ FE12-02 | FE12-04 | FE12-04 | 同上 |
| ワクチン | ワクチン一覧 | /vaccinations | VaccinationList | △ FE12-02 | FE12-03 | ✅ | C6a期限超過/死亡/RBAC |
| ワクチン | ワクチン接種 - ペット選択 | /vaccinations/select-pet | VaccinationPetSelection | △ FE12-02 | ✅ | ✅ | C6a死亡/危険ペット選択 |
| ワクチン | ワクチン登録 | /vaccinations/new | VaccinationForm | △ FE12-02 | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| ワクチン | ワクチン編集 | /vaccinations/:id | VaccinationForm | △ FE12-02 | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 定期健診 | 定期健診一覧 | /checkups | CheckupsList | △ FE12-02 | FE12-03 | ✅ | C6a要フォロー表示 |
| 定期健診 | 定期健診登録 - ペット選択 | /checkups/select-pet | CheckupPetSelection | △ FE12-02 | ✅ | ✅ | C6a死亡/危険ペット選択 |
| 定期健診 | 定期健診登録 | /checkups/new | CheckupForm | △ FE12-02 | FE12-03 | ✅ | C6a臨床status |
| 受付/飼主/予約 | 受付 | / | Reception | △ FE12-02 | ✅ | ✅ | C6a危険/死亡ペットの受付表示 |
| 受付/飼主/予約 | 飼主一覧 | /owners | OwnersList | △ FE12-02 | FE12-03 | ✅ | C6a danger/deceased filter、併発: C18 FE12-01 |
| 受付/飼主/予約 | 飼主登録 | /owners/new | OwnerForm | △ FE12-02 | FE12-03 | FE12-04 | OwnerSearchModal C18併発 |
| 受付/飼主/予約 | 飼主編集 | /owners/:id | OwnerForm | △ FE12-02 | FE12-03 | FE12-04 | OwnerSearchModal C18併発 |
| 受付/飼主/予約 | 集計ダッシュボード | /aggregation | AggregationDashboardPage | ✅ | FE12-03 | ✅ | `AggregationDashboardPage.tsx:3` |
| 受付/飼主/予約 | 予約管理 | /reservations | ReservationManagement | ✅ | FE12-03 | ✅ | `ReservationManagement.tsx:7` |
| 運用/Lステップ | Lステップ健診連携 | /lstep/checkup-sync | CheckupSyncPage | ✅ | FE12-03 | ✅ | `CheckupSyncPage.tsx:4` |
| 運用/Lステップ | Lステップ配信モニター | /lstep/delivery-monitor | LstepDeliveryMonitorPage | ✅ | FE12-03 | ✅ | `LstepDeliveryMonitorPage.tsx:2` |
| 運用/Lステップ | Lステップ分析 | /lstep/analytics | LstepAnalyticsPage | FE12-01 | FE12-03 | ✅ | C18 raw 26件（3 component） |
| 運用/LINE予約 | LINE予約設定(index) | /line-reservation | LineReservationSettings | ✅ | ✅ | ✅ | 本体route。別app axis③はFE12-08 / FE12-09 |
| 運用/LINE予約 | LINE予約設定 | /line-reservation/settings | LineReservationSettings | ✅ | ✅ | ✅ | 同上 |
| 運用/LINE予約 | LINE予約ページエディタ | /line-reservation/page-editor | LineReservationPageEditor | ✅ | ✅ | ✅ | 同上 |
| 運用/LINE予約 | LINE予約枠設定 | /line-reservation/slots | LineReservationSlotsSettings | ✅ | FE12-03 | ✅ | `LineReservationSlotsSettings.tsx:3` |
| 運用/医院 | 医院マスタ設定 | /settings/clinic | ClinicMasterList | ✅ | ✅ | ✅ | 追加指摘なし |
| 運用/マニュアル | マニュアル | /manual | ManualPage | ✅ | FE12-03 | ✅ | 独自shellは裁定済み、icon importのみ対象 |
| 運用/マニュアル | マニュアル記事 | /manual/:category/:slug | ManualPage | ✅ | FE12-03 | ✅ | 同上 |
| 設定/マスタ | 設定トップ | /settings | MasterSettingsIndex | ✅ | FE12-03 | ✅ | `MasterSettingsIndex.tsx:7` |
| 設定/マスタ | 職員マスタ | /settings/staff | StaffSettings | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 設定/マスタ | 診療項目マスタ | /settings/treatment-items | TreatmentPlanMaster | ✅ | FE12-04 | FE12-04 | global listener重複 |
| 設定/マスタ | 診断名マスタ | /settings/diagnosis | DiagnosisSettings | ✅ | FE12-04 | FE12-04 | 同上 |
| 設定/マスタ | 動物種マスタ | /settings/animal-species | AnimalSpeciesSettings | ✅ | FE12-04 | FE12-04 | 同上 |
| 設定/マスタ | トリミングマスタ | /settings/trimming | TrimmingSettings | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 設定/マスタ | トリミングコース種別マスタ | /settings/trimming-course-type | TrimmingCourseTypeSettings | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 設定/マスタ | 薬剤マスタ | /settings/medicine | MedicineSettings | ✅ | FE12-04 | FE12-04 | global listener重複 |
| 設定/マスタ | 予約種別マスタ | /settings/reservation-type | ReservationTypeSettings | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 設定/マスタ | 入院・ホテルマスタ | /settings/hospitalization | HospitalizationSettings | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 設定/マスタ | ケージマスタ | /settings/cage | CageSettings | ✅ | FE12-03 | FE12-04 | 併発: 通貨 FE12-10 |
| 設定/マスタ | 物販品マスタ | /settings/merchandise-items | MerchandiseItemSettings | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 設定/マスタ | 保険マスタ | /settings/insurance | InsuranceSettings | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 設定/マスタ | 職種マスタ | /settings/occupations | OccupationSettings | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 設定/マスタ | 権限グループマスタ | /settings/permission-groups | PermissionGroupSettings | △ FE12-02 | FE12-03 | FE12-04 | 併発: C18 FE12-01、RBAC review |
| 設定/マスタ | 問診テンプレートマスタ | /settings/inquiry-templates | InterviewTemplateSettings | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 設定/マスタ | 主訴マスタ | /settings/interview/chief-complaint | ChiefComplaintSettings | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 設定/マスタ | 問診テンプレート(interview) | /settings/interview/templates | InterviewTemplateSettings | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 設定/マスタ | シフトテンプレートマスタ | /settings/shift-templates | ShiftTemplateSettings | ✅ | FE12-04 | FE12-04 | global listener重複 |
| 設定/マスタ | 締め時間設定 | /settings/closing-time | ClosingSettingsPage | FE12-01 | FE12-03 | ✅ | StandardClosingTime raw cell |
| 設定/マスタ | 支払方法マスタ | /settings/payment-methods | PaymentMethodSettings | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 設定/マスタ | 割引キャンペーンマスタ | /settings/campaigns | CampaignSettings | ✅ | FE12-03 | FE12-04 | 併発: axis② FE12-04 |
| 設定/マスタ | Lステップ連携設定 | /settings/integrations/lstep | LstepSettingsPage | ✅ | ✅ | ✅ | 追加指摘なし |
| 設定/マスタ | Lステップタグ管理 | /settings/lstep/tags | LstepTagManagementPage | ✅ | FE12-03 | ✅ | `LstepTagManagementPage.tsx:2` |
<!-- FE12-ROUTE-TABLE-END -->

## FE12 task backlog

同じ根因をページ数だけ複製しない。`対象route` は修正影響範囲であり、ページ表の主IDまたは「併発」注記から参照する。FE12-07〜09は3ツリー横断監査で検出したglobal taskのため、本体84ページの適合判定を機械的に❌へ変えない。

<!-- FE12-TASK-TABLE-START -->
| ID | Priority | 軸 | 対象route | 証跡(file:line) | 根拠(category/rule or MANDATORY/duplication) | 修正方針 | ②削除判定 | 将来のscoped検証 |
|---|---|---|---|---|---|---|---|---|
| FE12-01 | P1 | ① | owner report、会計new/detail/close/history/reports、medical-record form、hospitalization form/detail、Lstep analytics、permission-groups、closing-time | `frontend/scripts/design-system-audit.mjs:135-161`、`frontend/src/features/hospitalization/components/HospitalizationTreatmentTable.tsx:84` | C18 / DESIGN.md `ex-data-table-cell`; 22ファイル204 raw cellのnon-gating ratchet | 22ファイルを表primitive移行可能、帳票/特殊matrix、構造cellの3群へ再分類し、標準表だけを`TableHead`/`TableCell`へ置換する。1バッチごとにbaseline件数を減らし、allowlistは拡大しない。**Batch 1 COMPLETE (2026-07-24)**: 標準data table 17ファイルを移行し、C18 baselineを204→44へ削減。**Batch 2 COMPLETE (2026-07-24)**: form内table 3ファイル・34 raw cellを移行し、baselineを44→10へ削減。下記実行記録を参照 | **可**。raw `th`/`td`と重複classを削除し既存primitiveへ統合。追加だけは禁止 | `docker compose exec frontend pnpm design-audit` |
| FE12-02 | P0 | ① | owners/reception、medical-records、hospitalization、examinations、vaccinations、checkups、permission-groups | `docs/spec/ui-design-compliance.md:14`、`frontend/src/features/medical-records/routes/MedicalRecords.tsx:261-270`、`frontend/src/features/hospitalization/components/HospitalizationListView.tsx:78` | C6a / design-system.md §2.4・§9。危険/死亡/RBAC非活性は静的網羅不能で臨床安全が最優先 | danger・死亡・異常値・期限超過・権限なしの各sentinelをコードレビューし、色だけでなく文言/操作抑止/accessible nameを確認する。不一致時だけ誤った装飾を既存semantic tokenへ置換し、正常表示の追加はしない | **条件付き可**。適合ならコード追加0。不適合なら誤った色・操作・重複分岐を削除/置換 | `docker compose exec frontend npx vitest run src/features/hospitalization/components/HospitalizationListView.test.tsx src/features/medical-records/routes/MedicalRecords.test.tsx src/features/owners/routes/OwnerForm.permissions.test.tsx src/features/master/routes/MasterReorderPermissionGuards.test.tsx` |
| FE12-03 | P1 | ② | page表で参照した本体route（production 208ファイル） | `frontend/src/features/accounting/routes/AccountingList.tsx:17`、`frontend/src/features/estimates/routes/EstimateList.tsx:10`、`frontend/src/features/master/routes/StaffSettings.tsx:2` | Bundle Size Optimization / `bundle-barrel-imports`。ただしproject Feature Indexing barrelはMANDATORYのため対象外、第三者`lucide-react` root importだけを対象 | Viteのchunk実測を先に取り、対応可能ならicon単位subpath importまたは既存Vite最適化へ置換する。feature indexは維持し、独自icon wrapperは新設しない | **可**。208箇所のroot barrel importを削除/置換し、新規wrapper追加0 | `rg -n "from .lucide-react." frontend/src`、`docker compose exec frontend npx vitest run src/features/accounting/routes/AccountingRouteGuards.test.tsx src/features/estimates/routes/EstimateList.test.tsx src/features/master/routes/StaffSettings.test.tsx` |
| FE12-04 | P1 | ②/③ | dirty form群・shift/master side peek群 | `frontend/src/hooks/use-side-peek-dirty.ts:31-40`、`frontend/src/hooks/use-unsaved-changes.ts:14-25` | Client-Side Data Fetching / `client-event-listeners`; `frontend/CLAUDE.md` Shared Helper配置MANDATORY / duplication | 共通beforeunload購読を`src/hooks`の既存hookへ統合し、dirty状態APIの差だけをadapterなしで吸収する。2重listener実装と同一handlerを削除する | **可**。2実装を1実装へ統合してlistener/effect重複を削除。追加だけは禁止 | `docker compose exec frontend npx vitest run src/hooks/use-unsaved-changes.test.tsx src/features/estimates/routes/EstimateForm.test.tsx src/features/master/routes/MasterReorderPermissionGuards.test.tsx` |
| FE12-05 | P2 | ② | `/hospitalization/:id` | `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx:51-60` | Re-render Optimization / `rerender-simple-expression-in-memo` | primitive文字列比較と`clampDate`の単純導出をinline計算へ置換し、依存配列と不要importを削除する。挙動・初期dateは維持 | **可**。2つの`useMemo`と不要importを削除し、追加0 | `docker compose exec frontend npx vitest run src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.test.tsx` |
| FE12-06 | P1 | ③ | `/estimates`、`/estimates/new`、`/estimates/:id`、`/estimates/:id/edit` | `frontend/src/features/estimates/utils/estimate-status-options.ts:1-20`、`frontend/src/features/estimates/utils/is-estimate-expired.ts:1-9`、`frontend/src/features/estimates/utils/is-estimate-locked-status.ts:1-13` | `frontend/CLAUDE.md` Shared Helper配置・Shared Constants配置・Feature Indexing MANDATORY | status定数を`src/constants`、再利用helperを`src/lib`の既存責務へ統合し、feature index経由の公開境界を維持する。単一consumerならconsumerへinlineしてファイル自体を消す | **可**。禁止`utils/` directoryと不要な薄いhelperを削除/統合。移動だけでファイル増加しない | `find frontend/src frontend/liff/src frontend/line-reserve/src -type d -name utils -print`、`docker compose exec frontend npx vitest run src/features/estimates/routes/EstimateList.test.tsx src/features/estimates/routes/EstimateForm.test.tsx src/features/estimates/routes/EstimateDetail.test.tsx` |
| FE12-07 | P1 | ③ | 3ツリーglobal（generated model consumers） | `frontend/src/types/generated/models.ts:210-216`、`frontend/src/types/generated/models.ts:497`、`frontend/src/types/generated/models.ts:1114`、`frontend/src/types/generated/models.ts:1816-1822` | `frontend/CLAUDE.md` Type Safety MANDATORY。production generated modelに`any` 17件 | 生成元mappingを修正し、JSONは`unknown`+境界schema、UUIDは`string`へ置換する。生成物の手編集はせず、consumerの不要cast/防御分岐も型確定後に削除する | **可**。17個の`any`と派生castを削除/置換。schemaは既存境界へ統合し追加だけにしない | `rg -n "\bany\b" frontend/src/types/generated/models.ts`、`docker compose exec frontend npx vitest run src/features/cash-register/category-breakdown.test.ts src/features/estimates/api/transforms.test.ts` |
| FE12-08 | P2 | ③ | line-reserve 3画面（本体84ページ別デザイン監査外） | `frontend/line-reserve/src/pages/ConfirmPage.tsx:32-42`、`frontend/line-reserve/src/pages/TimeSelectPage.tsx:19-22`、`frontend/line-reserve/src/pages/CompletePage.tsx:19-25` | `frontend/CLAUDE.md` Shared Helper配置MANDATORY / 3-tree duplication | 既存`@/shared-liff/jst-date`へ契約一致分を統合し、padding有無を明示引数で維持する。時刻formatterも1箇所へ統合して画面local wrapperを削除する | **可**。date/time formatter 6関数を既存shared helperへ統合・削除し、新規helper treeは作らない | `docker compose exec frontend npx vitest run line-reserve/src/pages/ConfirmPage.test.tsx line-reserve/src/pages/TimeSelectPage.test.tsx` |
| FE12-09 | P1 | ③ | line-reserve ConfirmPage（本体84ページ別デザイン監査外） | `frontend/line-reserve/src/pages/ConfirmPage.tsx:88-139`、`frontend/line-reserve/src/pages/ConfirmPage.tsx:223-225` | `frontend/CLAUDE.md` React 19 Patterns MANDATORY（`useActionState` + `<form action>` + SubmitButton） | manual `submitting` state/try-finally actionをReact 19 form actionへ置換し、既存error/409分岐をaction stateに統合する。二重送信防止とLIFF message順序を維持 | **可**。`submitting` state、setter、manual pending/finallyを削除し、既存form patternへ統合 | `docker compose exec frontend npx vitest run line-reserve/src/pages/ConfirmPage.test.tsx` |
| FE12-10 | P2 | ③ | accounting/reports/cash-register、medical-records、hospitalization、master金額表示 | `frontend/src/lib/format/number.ts:28-40`、`frontend/src/features/accounting/components/ItemListCard.tsx:311-314`、`frontend/src/features/hospitalization/components/HospitalizationCostSummary.tsx:38-100` | `frontend/CLAUDE.md` Shared Helper配置MANDATORY / feature間duplication | `¥`/`￥`、符号、0の扱いが既存契約と一致する箇所だけ`formatCurrency`/`formatCurrencyOrDash`へ置換する。帳票固有・差額固有は別契約として残し、無理な一律化をしない | **可**。一致するinline `toLocaleString`と重複分岐を削除し、既存helperへ統合。新helper追加0 | `docker compose exec frontend npx vitest run src/features/accounting/routes/AccountingDetail.test.tsx src/features/cash-register/category-breakdown.test.ts src/features/hospitalization/components/HospitalizationCostSummary.test.tsx` |
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

<!-- FE12-02-REVIEW-END -->
