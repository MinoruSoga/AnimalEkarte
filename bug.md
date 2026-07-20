# AnimalEkarte — バグ台帳（bug.md）

> **本書の役割**: 未修正バグの正本台帳。バグではないが対応すべきもの（USER アクション・実装タスク・改善）は `todo.md`。
> **運用**: 受け入れシナリオ・レビュー・実機で発見した不具合はここへ BUG-XXX で起票する（BUG-4xx = 受け入れシナリオ由来）。修正完了したら本書から削除し、修正コミットと発見経緯は git 履歴・実行レポート（`docs/ops/testing/scenarios/reports/`）を正とする。
> **粒度**: 次に着手するエージェントが本書だけで調査に入れること（症状・再現手順・調査済みの根因・修正方向）。
> **BE9 注記（2026-07-19〜）**: backend は domain package へ移行中（`BE-refactor.md` BE9 / ADR-006）。本書が参照する backend パスは移行で順次移動する — 着手時に `docs/architecture/be9-2a-classification-manifest.csv` で現在地を確認し、修正が backend 新規実装を含む場合は `internal/handler|service|repository` へ追加せず target domain package へ実装する（todo.md「BE 実装規約」参照）。

## Open

### BUG-417:【LOW・潜在・BE9-2A監査で検出】billing_item_repository.go の Update/Delete が clinic 分離を実質担保していない（defense-in-depth 不全・現状は生きた漏洩ではない）

- **症状**: `internal/repository/billing_item_repository.go`（BE9 target=billing）の Update/Delete が `.Joins("JOIN billings ON ...billings.clinic_id=?...")` を `.Updates()`/`.Delete()` へ連結する形式だが、**GORM の `Joins()` は UPDATE/DELETE SQL へ伝播しない**ため、repository 層の clinic 述語は実質 no-op。Treatment/ClinicalPlan 等が subquery 形式で正しく回避している同型の罠に、このファイルだけが該当（billing_confirmation/estimate は検証済みで正しい）。
- **現状の安全性**: `billing_item_service.go` の UpdateItem/DeleteItem が事前に clinic-scoped `FindByID` で gate しているため**現時点で生きた漏洩ではない**。ただし事前 check を経由しない新規経路（admin 経路・background job 等）が repository method へ直接到達すると silent なクロステナント書き込み/削除が発生し得る。クロステナント分離 test も現状ゼロ。
- **修正方向**: subquery 形式（`WHERE id IN (SELECT ... JOIN billings ... WHERE billings.clinic_id=?)`）への是正＋クロステナント分離 test 追加。**詳細の正本 = `docs/architecture/be9-2a-boundary-map.md` §7.4／ADR-006 未解決論点#6**。
- **修正タイミング**: BE9-2C/2D の billing domain 着手時の**必須前提**（ADR-006 で着手前ゲート化済み）。ただし BE9 と無関係にこのファイルへ触れる修正が先に発生した場合も、その場での是正を必須とする。
- **発見**: 2026-07-19（BE9-2A santa dual-review round 2・clinic-isolation-auditor。BE9-2A は measurement-only のため未修正のまま記録）。

### BUG-416:【LOW・healthcare-reviewer指摘】カルテ診断(diagnosis1/2)保存の残存リスク（BUG-410 backend/UI follow-up）— 残るのは①④のみ（②FE病名バリデーション欠如=修正済み 08c82490／③clinical_plan楽観ロック欠如=修正済み 797f4d2d）

- **経緯**: BUG-410（構造化診断 hydrate 欠落・修正済み 1407a39a）の独立監査（healthcare-reviewer、2026-07-18、APPROVE・3件とも非ブロッキング）で指摘。②は 2026-07-18 commit `08c82490`、③は 2026-07-19 commit `797f4d2d`（clinical_plan_request.go に Version フィールド追加・clinical_plan_repository.go の UPDATE に `version = ?` 述語追加・FE の `use-medical-record-save-action.ts` も version 送信に対応、medical_record と parity化）でそれぞれ修正済み。詳細は git 履歴を正とする。残る①④はいずれも「将来 UI 追加時のみ発火する前提条件」で現状未到達（2026-07-19 再監査で確認）。
- **① save-action の diagnosis1/diagnosis2 送信非対称（クリアUI追加時の前提条件）**: `use-medical-record-save-action.ts` は diagnosis1 を `?? undefined`（未送信=BE 側で「更新しない」扱い）、diagnosis2 は state の値をそのまま送信（`null` なら明示クリア）という非対称な契約になっている。現行 UI（`SearchableSelect`）には選択解除操作が無いため両方とも現状は発火しないが、**将来 diagnosis1/2 いずれかにクリアボタンを追加する場合、diagnosis1 側は `?? undefined` のためクリア操作がサイレントに no-op する**（「保存しました」トーストは出るが DB は変わらない）。クリアUI追加時はこの非対称の是正が前提条件。
- **④ レコード切替時の hydrate guard 再利用リスク（調査済み・現状未到達と判定）**: `useApplyMedicalRecord` の hydrate は `existingRecord.diagnosisXxx != null` の場合のみ setter を呼ぶ。理論上、同一 `MedicalRecordForm` インスタンスが保持されたまま record A（diagnosis2=4,9）→ record B（diagnosis2=null）へ切り替わると、B 用の setter が一度も呼ばれず A の値が state に残存し、B の保存時に A の診断が誤って書き込まれうる（データ喪失より悪いクロスペイシェント汚染）。**実コード調査で現状は再現不可と判定**: ルート定義（`frontend/src/app/routes/clinical-care-routes.tsx` の `path: ":id"`）に `key` 指定なし かつ `/medical-records/<id1>` → `/medical-records/<id2>` へ直接 `navigate()` する呼び出しはリポジトリ全体で0件（`medicalRecords.detail.getHref` の全呼び出し元＝会計一覧/健診一覧/カルテ一覧/新規作成auto-createはいずれも別ルートを経由してから `:id` に遷移するため React Router がコンポーネントを再マウントする）。`MedicalRecordForm.tsx` 内の来院履歴パネル（`InterviewHistory.tsx` 等）も展開/折りたたみのみで他レコードへの遷移リンクを持たない。**将来「次の来院/前の来院」等、同一画面内で record ID だけを差し替えるナビゲーションUIを追加する場合は、hydrate 全体（この guard パターンを共有する chiefComplaintTypeId/plan/assessment/notes 等も含む）を record 切替時に明示的リセットする設計が前提条件になる**。
- **重要度**: LOW（①④とも将来UI追加時のみ発火する前提条件で現状未到達）。
- **発見**: 2026-07-18（healthcare-reviewer による BUG-410 修正の独立監査、APPROVE・3件とも別チケット化推奨。④は同監査後の react-reviewer 観点フォローで発見・コード調査により現状未到達と確定）。

### BUG-413:【要監視・#250後に判定】予防接種・トリミング一覧が同型のページネーション不可視化リスクを抱える（現状 seed 0件で無症状）

- **経緯**: BUG-412（在庫一覧の偽ページネーション）の派生監査（2026-07-17）で発見。BUG-412 は 2026-07-17 に修正済み（`InventoryList`/`use-inventory.ts`/`frontend/src/features/inventory/api/inventory.ts` をサーバサイド page/limit 転送 + 実 total 消費に変更。backend `inventory_repository.go` は元々 clinic-scoped 実 COUNT を WHERE 適用後に返しており変更不要だった）。本エントリは BUG-412 調査で判明した**同一パターンの潜在リスク**のうち、今回は修正しないと判断したものを引き続き追跡する。
- **VaccinationList.tsx（予防接種一覧）**: `use-vaccinations.ts` → `get-vaccinations.ts` の `getVaccinations` は `start_date`/`end_date` を日付フィルタ使用時のみ送る任意パラメータで、未指定時は無条件で backend `parsePagination` の既定 `limit=20` に落ちる（date-scoped ではない）。`backend/migrations/seeds/003_demo/vaccinations.csv` は現状 0 件（ヘッダのみ）で無症状・検証不能。
- **TrimmingList.tsx（トリミング一覧）**: `use-trimming-records.ts` → `get-trimmings.ts` の `getTrimmings` は明示的に `page:1, limit:HISTORY_FETCH_LIMIT`（`frontend/src/config/fetch-limits.ts` で `100` = backend `defaultMaxPaginationLimit`）を送信しており `limit=20` の欠陥ではないが、1リクエストの上限100件を超えた場合の継続取得導線が無い。トリミング実績は `reservation_type.category='trimming'` の予約（`appointments.csv`）であり、現状 0 件で無症状・検証不能。
- **今回修正しない理由**: 両画面とも seed が 0 件のため、ページネーション化しても page=2 が別レコードを返すことを実測で反証できず、「直った」ことを検証できない。加えて両フックとも医師/ステータス/種別等のクライアントサイドフィルタを持ち、BUG-411/412 と同様「フィルタがページ内スコープに退行しないか」の再設計が必要（Trimming はさらに limit=100→カーソル/オフセット継続取得という追加拡張が要る）。検証不能な状態での臨床データ一覧の書き換えは、修正漏れより悪い結果（未検証の新規回帰）を生みうるため見送った。
- **要監視**: #250（本番データ移行）でこの2画面の実件数が判明した時点で、実件数が limit（Vaccination=20 / Trimming=100）を超えるか再確認し、超える場合は本エントリを起票根拠に BUG-411/412 と同型の修正を行う。#250 の受け入れ条件にこの再確認を明示的に含めること。
- **発見**: 2026-07-17（BUG-412 対応中の派生調査、コーディネーターが claim を再検証済み）。

### BUG-408:【設計判断待ち】予防接種フォームのワクチン選択に動物種(species)フィルタが無い

- **経緯**: 旧 BUG-408/401（予防接種フォームが `VACCINE_TYPE_ITEMS` というハードコード2択で、ワクチンマスタを一切クエリせず、誤った vaccine_id を永続化していたデータ破損）は **2026-07-17 に修正済み**。`VaccinationFormPanels.tsx` が姉妹フォーム（`MedicalRecordVaccination.tsx`）と同じ `useGetAllVaccinesMaster()` パターンで実マスタを `isActive` フィルタ付きでクエリするよう変更し、`use-vaccination-form.ts` の `VACCINE_SCHEDULE_MAP`（"1"/"2"固定キー）も選択ワクチンの `interval` フィールドから次回予定を導出する方式に置き換えた（回帰テスト: `use-vaccination-form.test.ts` の「BUG-401」節、`VaccinationFormPanels.test.tsx`）。
- **残存する設計判断（未決・本エントリが追跡するのはこれのみ）**: マスタ項目を選択ペットの動物種でどう絞り込むか。姉妹フォームも動物種フィルタを持たないため、上記修正は parity に留め新規リスクを増やしていないが、犬患者に猫用ワクチン（またはその逆）が選択可能な状態そのものは残っている。選択ペットの `species`（`AnimalSpecies.name` 由来のフリーテキスト、`frontend/src/lib/transforms/pet.ts:84`）は既に取得済みで型拡張は不要（2026-07-19 再監査で訂正）。実際の未決点は、`Vaccine.species` が固定3値enum（`dog`/`cat`/`both`、`frontend/src/types/generated/models.ts:3130-3133`）である一方 `AnimalSpecies.name` はマスタ管理のフリーテキストで固定語彙が無いため、両者をどう対応付けるか（またはマスタ側にenum相当のフィールドを新設するか）が未決という点。
- **調査の起点**: `Vaccine.species` 固定3値enumと `AnimalSpecies.name` フリーテキストのマッピング設計、`VaccineItem.species` フィールド（`frontend/src/lib/transforms/treatment.ts`）との突合。
- **発見**: 2026-07-17（BUG-401 調査中）。vaccine_id 誤保存自体は同日中に修正済み。

## 直近クローズ（次回整理で削除）

- **BUG-404**（入院デイリー記録 GET/ケアログ POST 全 500）: **修正済み 2026-07-17**（commit 58c653df）。根因 = TIME 列を `time.Time` で Scan（書込成功・読取全滅）+ 永続テスト DB のスキーマドリフトがテストを素通りさせていた（自己修復 ALTER 追加済み）。次回シナリオ再実行で最終確認したら本行を削除。
