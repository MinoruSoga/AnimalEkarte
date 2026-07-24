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
| FE12-01 | P1 | ① | owner report、会計new/detail/close/history/reports、medical-record form、hospitalization form/detail、Lstep analytics、permission-groups、closing-time | `frontend/scripts/design-system-audit.mjs:135-161`、`frontend/src/features/hospitalization/components/HospitalizationTreatmentTable.tsx:84` | C18 / DESIGN.md `ex-data-table-cell`; 22ファイル204 raw cellのnon-gating ratchet | 22ファイルを表primitive移行可能、帳票/特殊matrix、構造cellの3群へ再分類し、標準表だけを`TableHead`/`TableCell`へ置換する。1バッチごとにbaseline件数を減らし、allowlistは拡大しない。**Batch 1 COMPLETE (2026-07-24)**: 標準data table 17ファイルを移行し、C18 baselineを204→44へ削減。下記実行記録を参照 | **可**。raw `th`/`td`と重複classを削除し既存primitiveへ統合。追加だけは禁止 | `docker compose exec frontend pnpm design-audit` |
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
