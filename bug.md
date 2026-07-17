# AnimalEkarte — バグ台帳（bug.md）

> **本書の役割**: 未修正バグの正本台帳。バグではないが対応すべきもの（USER アクション・実装タスク・改善）は `todo.md`。
> **運用**: 受け入れシナリオ・レビュー・実機で発見した不具合はここへ BUG-XXX で起票する（BUG-4xx = 受け入れシナリオ由来）。修正完了したら本書から削除し、修正コミットと発見経緯は git 履歴・実行レポート（`docs/ops/testing/scenarios/reports/`）を正とする。
> **粒度**: 次に着手するエージェントが本書だけで調査に入れること（症状・再現手順・調査済みの根因・修正方向）。

## Open

### BUG-415:【MEDIUM・healthcare-reviewer指摘】generic PATCH /pets/:id が status を deceased_at・監査ログと無結合のまま書込可能（BUG-409 backend follow-up）

- **経緯**: BUG-409（生死ステータスの二重管理・修正済み 74652f72）の独立監査（healthcare-reviewer、2026-07-18、APPROVE WITH FOLLOW-UP REQUIRED）で指摘。BUG-409 の修正は FE の生死ラジオ（唯一のUI攻撃面）を disabled 化して閉じたが、**二重管理の構造自体は backend API 契約層に残っている**。
- **症状**: `frontend/src/lib/transforms/pet.ts:201`（`transformUpdatePetRequest`）は `status` を無条件に常時送信する。ペット編集の「更新」保存は毎回 generic `PATCH /pets/:id` に到達し、`backend/internal/service/pet_service.go:109-111`（`buildPetUpdate`）が `status` を `deceased_at` と無結合・監査ログなしで書き込む。現行 FE UI からは divergent な値を生成できない（disabled ラジオのため）が、任意の API クライアント・将来の FE 面・transform のバグにより `status=死亡, deceased_at=NULL, 監査なし`（またはその逆）を生成できる**構造**が生きたままである。死亡登録の監査証跡はカルテの法的・臨床記録要件。
- **修正方針（healthcare-reviewer 提案）**: generic 書込経路から `status` を削除する。
  - FE: `transformUpdatePetRequest`（`frontend/src/lib/transforms/pet.ts`）から `status` フィールドを除去。
  - BE: `buildPetUpdate`（`backend/internal/service/pet_service.go:109-111`）の `status` 分岐を除去。
  - 結果、`HandlePetDeath`/`HandlePetRevival`（`lstep_lifecycle_service.go`、status+deceased_at+監査を同一tx・fail-closed で原子更新）が `status` の唯一の書込元になる。status は alive/deceased の2値のみ・Create 時に設定・遷移は死亡登録/取消のみ、と既に確定しているため安全に削除可能と判定済み。
  - 副次効果: `use-pet-form-list-state.ts` の `isPetRevival`（generic PATCH 後に revoke を呼ぶ冗長経路）が自然に不要化・簡素化される。
- **重要度**: MEDIUM（現行 UI から divergent 値は作れないため即時 hotfix 級ではないが、Go-live 前の backend ハードニング対象）。
- **発見**: 2026-07-18（healthcare-reviewer による BUG-409 修正の独立監査、APPROVE WITH FOLLOW-UP REQUIRED）。

### BUG-416:【LOW〜MEDIUM・healthcare-reviewer指摘・現状は潜在的で未発火】カルテ診断(diagnosis1/2)保存の非対称送信・バリデーション欠如・楽観ロック欠如（BUG-410 backend/UI follow-up、3件）

- **経緯**: BUG-410（構造化診断 hydrate 欠落・修正済み 1407a39a）の独立監査（healthcare-reviewer、2026-07-18、APPROVE・3件とも非ブロッキング）で指摘。いずれも本コミットが新規導入したものではなく、修正後も残る既存の潜在リスク。
- **① save-action の diagnosis1/diagnosis2 送信非対称（クリアUI追加時の前提条件）**: `use-medical-record-save-action.ts` は diagnosis1 を `?? undefined`（未送信=BE 側で「更新しない」扱い）、diagnosis2 は state の値をそのまま送信（`null` なら明示クリア）という非対称な契約になっている。現行 UI（`SearchableSelect`）には選択解除操作が無いため両方とも現状は発火しないが、**将来 diagnosis1/2 いずれかにクリアボタンを追加する場合、diagnosis1 側は `?? undefined` のためクリア操作がサイレントに no-op する**（「保存しました」トーストは出るが DB は変わらない）。クリアUI追加時はこの非対称の是正が前提条件。
- **② diagnosis2 の FE 病名バリデーション欠如**: diagnosis1 には `if (diagnosis1CategoryId && !diagnosis1NameId)` の FE バリデーションがあるが、diagnosis2 には同等のチェックが無い。backend `clinical_plan_service.go` の `assertNameBelongsToType`（AUD-007）も `nameID == nil` の場合は許可を返すため、「diagnosis2 のカテゴリだけ変更→病名未選択のまま保存」が 400 で拒否されず、type あり/name NULL の不完全状態が永続化されうる。データ喪失ではないが臨床データの不整合。
- **③ clinical_plan PATCH に楽観ロック（version）が無い**: `updateTreatmentPlanMutation` の payload に `version` が含まれない（同じフックの後続 `updateMutation`＝次回来院日更新は version を送る）。2名の獣医が同一カルテの診断を同時編集すると last-write-wins で lost update の理論的余地がある。
- **④ レコード切替時の hydrate guard 再利用リスク（調査済み・現状未到達と判定）**: `useApplyMedicalRecord` の hydrate は `existingRecord.diagnosisXxx != null` の場合のみ setter を呼ぶ。理論上、同一 `MedicalRecordForm` インスタンスが保持されたまま record A（diagnosis2=4,9）→ record B（diagnosis2=null）へ切り替わると、B 用の setter が一度も呼ばれず A の値が state に残存し、B の保存時に A の診断が誤って書き込まれうる（データ喪失より悪いクロスペイシェント汚染）。**実コード調査で現状は再現不可と判定**: ルート定義（`frontend/src/app/routes/clinical-care-routes.tsx` の `path: ":id"`）に `key` 指定なし かつ `/medical-records/<id1>` → `/medical-records/<id2>` へ直接 `navigate()` する呼び出しはリポジトリ全体で0件（`medicalRecords.detail.getHref` の全呼び出し元＝会計一覧/健診一覧/カルテ一覧/新規作成auto-createはいずれも別ルートを経由してから `:id` に遷移するため React Router がコンポーネントを再マウントする）。`MedicalRecordForm.tsx` 内の来院履歴パネル（`InterviewHistory.tsx` 等）も展開/折りたたみのみで他レコードへの遷移リンクを持たない。**将来「次の来院/前の来院」等、同一画面内で record ID だけを差し替えるナビゲーションUIを追加する場合は、hydrate 全体（この guard パターンを共有する chiefComplaintTypeId/plan/assessment/notes 等も含む）を record 切替時に明示的リセットする設計が前提条件になる**。
- **重要度**: ①③④ は LOW（現状 UI からは未到達）、② は MEDIUM（到達可能な不完全データ永続化）。
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
- **残存する設計判断（未決・本エントリが追跡するのはこれのみ）**: マスタ項目を選択ペットの動物種でどう絞り込むか。姉妹フォームも動物種フィルタを持たないため、上記修正は parity に留め新規リスクを増やしていないが、犬患者に猫用ワクチン（またはその逆）が選択可能な状態そのものは残っている。選択ペットのデータ型に `species` が含まれておらず型拡張が必要になる可能性がある。
- **調査の起点**: `usePetSelection` が返す pet オブジェクトの型定義、`VaccineItem.species` フィールド（`frontend/src/lib/transforms/treatment.ts`）との突合。
- **発見**: 2026-07-17（BUG-401 調査中）。vaccine_id 誤保存自体は同日中に修正済み。

## 直近クローズ（次回整理で削除）

- **BUG-410**（カルテ編集保存時に構造化診断（diagnosis1/2）が再投入されず上書きされる）: **修正済み 2026-07-18**（commit 1407a39a、payload実証テスト 9f03e0eb、react-reviewer Block是正 0b5b4b8b）。再現テストで確認: `transformMedicalRecord` が `clinical_plan.diagnosis_type_id`/`diagnosis_name_id`/`diagnosis_2_type_id`/`diagnosis_2_name_id` を一切マップせず、`useApplyMedicalRecord` にも対応する hydrate setter が存在しなかったため diagnosis1/2 の state は常に null。診断2 (`diagnosis2CategoryId`/`diagnosis2NameId`) は保存アクションでそのまま `null` 送信され、backend `clinical_plan_request.go` の `nullableUint64RequestField` 契約（null=NULLクリア）により、診察/治療プランタブで診断以外を編集して保存するだけで保存済み diagnosis2 が無言クリアされることを実測確認（diagnosis1 は `?? undefined` 送信のため偶然保護されていた）。修正は BUG-406 の hydrate パターン（0213e4c9f）と同型で transform + hydrate setter を追加。診断入力体験自体は変更していない。**独立監査（react-reviewer）で二次欠陥を検出・是正済み**: 初回修正は `useState(existingRecord)` によるレンダー中比較（render-phase setState）を維持したままだったため、TanStack Query のウォームキャッシュ（`staleTime`=5分）で同一カルテを短時間内に再訪問すると hydrate が一度も発火せず、BUG-410 の症状（診断の無言クリア）が再現条件次第で残存していた（RED で実測確認）。`useEffect` 化（先例: 10f69364 の会計側修正）で解消。同一機構を共有する chiefComplaint/plan/assessment/notes 等の既存フィールドの潜在バグも副次的に解消。レコード切替時の state 漏れ（cross-record leak）は調査の結果、現行ルーティングでは到達不能と判定（BUG-416④）。
- **BUG-409**（ペット死亡ステータスの二重管理が外側フォーム経由で再発しうる）: **修正済み 2026-07-18**（commit 74652f72、follow-up ae615051/15b386ba）。再現テストで確認: `PetCareSection.tsx` の生死ラジオが `deceased_at`/監査ログと独立に `status` のみを書き換え可能で、①生存ペットで死亡ラジオ→保存 = status=死亡・deceased_at=null・監査ログ無し、②ダイアログ死亡登録後に生存ラジオ→保存 = status=生存・deceased_at残存、の2経路で不整合を実際に生成できた（`transformUpdatePetRequest`/backend `buildPetUpdate` いずれも deceased_at に触れない）。修正はラジオを現在値表示専用(disabled)にし、生死変更を監査付き `PetDeceasedRecordButton` 経由に一本化（transform への deceased_at 追加は無監査の第二死亡経路を新設するため不採用）。独立監査（typescript-reviewer + react-reviewer）で disabled 直接検証テスト・UXヒント文言・無効な `readOnly` 属性削除を追加反映済み。**残課題（非ブロッキング・未対応）**: disabled ラジオは恒久表示専用 UI としては a11y 的に弱いパターン（フォーカス対象外・スクリーンリーダーで「無効」と読まれる）という指摘があり、将来的にプレーンテキスト/バッジ表示への置換を検討余地あり（両レビュアーともブロック理由にはしていない）。backend API 契約層の残存二重管理構造は BUG-415 として別起票済み。
- **BUG-404**（入院デイリー記録 GET/ケアログ POST 全 500）: **修正済み 2026-07-17**（commit 58c653df）。根因 = TIME 列を `time.Time` で Scan（書込成功・読取全滅）+ 永続テスト DB のスキーマドリフトがテストを素通りさせていた（自己修復 ALTER 追加済み）。次回シナリオ再実行で最終確認したら本行を削除。
