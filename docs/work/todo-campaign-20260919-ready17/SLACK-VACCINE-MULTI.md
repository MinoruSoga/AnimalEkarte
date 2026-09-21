# SLACK-VACCINE-MULTI: 登録失敗 vs 同日 2–3 件の順次保存（batch API を捏造しない）

状態: **経路分離 READY／自動回帰（F 失敗 + S 同日順次単件 POST）追加済／batch UX・実機受入は PO 残**。出典は [todo-issue.md](../../../todo-issue.md) 見出し `### SLACK-VACCINE-MULTI`（L142–146、索引 L427）。保持する現場条件:

- 「登録できない」と「初診/同日に 2–3 件入力」は **別ケース**
- 単件 POST の存在だけで「同日複数不可」や **batch API 必須**と決めない
- 各接種の実施日・lot・次回予定・金額が **行をまたいで混ざらない**こと
- [種別課題](../../../todo-issue.md#uat-q2-vaccine-species) の承認済み条件を保持する。接種間隔などの臨床判断は代行しない
- 一括入力が必要な操作数 / 失敗時の部分保存の扱いは **PO 裁定後**に設計する（本 unit は batch API を追加しない）

本票は [カルテ内フォーム](../../../frontend/src/features/medical-records/hooks/use-medical-record-vaccination-form.ts) と [独立フォーム](../../../frontend/src/features/vaccinations/hooks/use-vaccination-form.ts)、[作成 API 再export](../../../frontend/src/features/vaccinations/api/create-vaccination.ts)、[実 mutation](../../../frontend/src/hooks/use-create-vaccination.ts)、[接種 service](../../../backend/internal/medicalrecord/vaccination_service.go) を照合する。本 unit の最小実装は **owned テスト追加 + 本票更新**のみ（製品 Create 契約は単件 POST のまま。batch エンドポイントは追加しない）。

本ファイルは製品コードから import されない。キャンペーン unit `SLACK-VACCINE-MULTI` の owned path および人間が読む調査票である。製品モジュールからの呼び出し行は無い（sibling `SLACK-VITALS.md` と同じ）。既存 `docs/work/todo-campaign-20260918/` にも本 unit の票は無い。ledger `owned_paths` が本パス単体のため、他シートへの追記では unit 完了にならない。種別集計の再設計は [UAT-Q2-VACCINE-SPECIES](../todo-campaign-20260918/UAT-Q2-VACCINE-SPECIES.md) に残し、本票では繰り返さない。

## 医院事実（コード外・UNKNOWN）

数値・院内ルールをコードから捏造しない。未採取なら該当セルは **再現 BLOCKED**。

| 項目 | コードで分かること | 医院事実 |
| --- | --- | --- |
| 報告医院・端末・ブラウザ | なし | **UNKNOWN** |
| 「登録できない」がどの画面・どのエラーか | カルテ埋め込みと独立画面の失敗点が違う（後述） | **UNKNOWN**。採取前に片方へ決めない |
| 「初診/同日 2–3 件」がカルテタブか予防接種画面か | 両経路とも単件 POST | **UNKNOWN**。現場の入口を採取する |
| 2–3 件が別ワクチンか同一ワクチンの再接種か | DB に (pet, date, vaccine) UNIQUE は無い | **UNKNOWN**。臨床間隔は裁定しない |
| 失敗時に先行 1 件を残すか全部戻すか | 現行は 1 POST = 1 commit。補償トランザクションは無い | **UNKNOWN**。PO 裁定まで batch/補償を設計しない |
| 金額を接種行に持たせたいかマスタ単価で足りるか | 接種行に amount 列は無い（後述） | **UNKNOWN** |
| フロント/API revision | 本票作成時 worktree HEAD `aac697645` | 再現セッションの SHA は **UNKNOWN** |

## 混ぜてはいけないケース

現場の「ワクチンが登録できない / 同日に何本も打ちたい」は次のどれでも同じ言葉になる。単件失敗・順次成功・batch 未実装・種別課題・会計単価を混ぜない。

| ID | 何か | いまのコード | 混ぜてはいけない読み替え |
| --- | --- | --- | --- |
| F — 単件登録失敗 | クライアント必須チェック、未保存カルテ、権限、死亡ペット、日付、関係検証 | 「同日複数が API に無いから失敗した」と読まない |
| S — 同日 2–3 件の順次保存 | 成功後にフォームを開き直し、同じ `POST /v1/vaccinations` を N 回 | 1 画面に N 行あることと、1 HTTP で N 件書くことを同一視しない |
| B — batch API | route は `POST /api/v1/vaccinations` の単件 Create のみ（[routes_records.go](../../../backend/internal/medicalrecord/routes_records.go) L10–15、[routes_snapshot_test.go](../../../backend/internal/medicalrecord/routes_snapshot_test.go) L257） | 単件 POST があることを batch 契約の根拠にしない。無いことを「同日複数禁止」の根拠にもしない |
| P — 種別 | マスタ `vaccines.species` とペット種。新規候補は species クエリ無し | 猫に犬用が出る件は [UAT-Q2-VACCINE-SPECIES](../todo-campaign-20260918/UAT-Q2-VACCINE-SPECIES.md)。本票でマスタ種を補完・一括削除しない |
| A — 金額 | 接種行に price/amount は無い。単価は `vaccines.price`。会計は別 lock 経路 | 接種 POST 成功を「金額が接種行に保存された」と読まない |
| L — lot / 次回予定 | lot1–4 と next_date は **その POST の body** に載る | 2 件目入力中の値を 1 件目へ上書き保存したことにしない。失敗時に残る先行行と混ぜない |
| I — 接種間隔 | 次回予定の計算は UI ヘルパ。BE は「次回 > 接種日」だけ | 臨床上の再接種間隔をコードが保証していると書かない |

## 現行経路（保存は常に 1 接種 = 1 POST）

両フォームとも作成は `useCreateVaccination` → `axios.post("/v1/vaccinations", req)`（[use-create-vaccination.ts](../../../frontend/src/hooks/use-create-vaccination.ts) L72–77）。`features/vaccinations/api/create-vaccination.ts` は **再export だけ**（L1）。handler は JSON 1 件を bind して `service.Create`（[vaccination_handler.go](../../../backend/internal/medicalrecord/vaccination_handler.go) L85–109）。配列 body / `CreateMany` / `/batch` は route に無い。

成功時 invalidate は `queryKeys.vaccinations.all()` = `["vaccinations"]`（use-create-vaccination.ts L79–80）。カルテ履歴の `["vaccinations","pet",petId]`（[use-pet-vaccinations.ts](../../../frontend/src/hooks/use-pet-vaccinations.ts) L64–66, L70）も prefix 一致で再取得される。一覧は `limit: HISTORY_FETCH_LIMIT`（100、[fetch-limits.ts](../../../frontend/src/config/fetch-limits.ts) L3）。2–3 件が 20 件ページ窓の向こうに消える BUG-007 は、この limit を付けた状態が前提。

DB: [001_init.sql](../../../backend/migrations/001_init.sql) L1477–1496。`(pet_id, date, vaccine_id)` UNIQUE は無い。`uq_vaccinations_id_clinic` は `(id, clinic_id)`（L4042–4043）。同日・同一ペット・別（または同一）ワクチンの 2 行目は **一意制約では拒否されない**。

### C — カルテ埋め込み（初診カルテから打つ経路）

入口: [MedicalRecordServiceTabs](../../../frontend/src/features/medical-records/components/MedicalRecordServiceTabs.tsx) L25–32。新規 / `recordId` 無しは [MedicalRecordSaveRequired](../../../frontend/src/features/medical-records/components/MedicalRecordTabsShared.tsx) L33–36 が「カルテを保存してから使用できます」で **フォームをマウントしない**。

保存済みカルテ: [MedicalRecordVaccination](../../../frontend/src/features/medical-records/components/MedicalRecordVaccination.tsx) → `useMedicalRecordVaccinationForm(petId, medicalRecordId)`。

| 段階 | コード | 同日 2–3 件への意味 |
| --- | --- | --- |
| 必須 | vaccine 未選択 / `vaccineName==="0"` → `ワクチン種別を選択してください`。date 空 → `接種日を入力してください`。API は呼ばない（hook L129–140、BUG-015） | 単件失敗。複数入力不足ではない |
| 新規カルテ | `medicalRecordId` 欠落なら `カルテを保存してから接種を追加してください`（L142–145）。タブ側 SaveRequired と二重 | 「登録できない」の第一候補。同日複数の前にカルテ ID が要る |
| pet 欠落 | `if (!petId) return` で **トーストも fieldError も無い**（L141） | 無音失敗。複数件とは別 |
| POST | `pet_id` / `medical_record_id` / `vaccine_id` / `date`(YYYY-MM-DD) / lot1–4 / `next_date` / supplemental / next_schedule_type / remarks（L149–162）。**amount は送らない** | 1 回 = 1 行。`medical_record_id` 付き |
| 成功 | toast「接種記録を追加しました」。`resetForm()`。`isAdding=false`。実施日は JST 当日へ戻す（L163–165, L56–69、テスト L144–163） | 2 件目は「記録を追加」を再クリックして **別 POST** |
| 失敗 | catch は toast 済み（`onError: handleApiError(..., "ワクチン接種登録")`）。フォームは残る（テスト L166–186） | 先行成功行はサーバに残る。この POST だけ失敗 |

履歴: `useGetPetVaccinations(petId)`。価格表示は `v.vaccine?.price`（use-pet-vaccinations.ts L39）であり接種行の列ではない。

権限・死亡: この hook は `usePermission("vaccinations")` も死亡ガードも持たない。独立画面の拒否と混ぜない。BE Create も死亡ペットを明示拒否しない（後述）。

未来日: カルテ hook はクライアントで「今日以前」を見ていない。独立フォームは見る。BE は見る。

### I — 独立予防接種画面

入口: [VaccinationForm.tsx](../../../frontend/src/features/vaccinations/routes/VaccinationForm.tsx) L20–38。`usePermission("vaccinations")` の `canCreate` / `canEdit` / `canDelete`。死亡ペットは `canSubmit` 偽（L66–67）。

作成: [runVaccinationSave](../../../frontend/src/features/vaccinations/hooks/use-vaccination-form-helpers.ts) L261–314。

| 段階 | コード | 同日 2–3 件への意味 |
| --- | --- | --- |
| 必須 | [validateVaccinationForm](../../../frontend/src/features/vaccinations/hooks/use-vaccination-form-model.ts) L88–118。vaccine / date。未来日。次回が本日より前（新規のみ）。次回 ≤ 接種日 | 単件失敗 |
| pet | query `petId` または選択ペットが無ければ return（helpers L295–296）。未選択は selectPet へ（helpers L36–41） | 対象ペット未固定の登録失敗 |
| 権限 | `canCreate` 偽は toast「この操作を行う権限がありません」（L298–300） | カルテ埋め込み経路と失敗メッセージが違う |
| 死亡 | toast「死亡したペットの予防接種記録は保存できません」（L302–304） | BE は同メッセージを返さない。FE のみ |
| POST | [buildCreateVaccinationRequest](../../../frontend/src/features/vaccinations/hooks/use-vaccination-form-model.ts) L138–156。`medical_record_id: null`、`pet_id`、date は `jstDateStartISOString`。lot / next_date / remarks。**amount 無し** | カルテ紐付け無しの接種行。会計は後述のとおり `medical_record_id` 無しを請求できない |
| 成功 | toast「予防接種を登録しました」。画面は list へ navigate（VaccinationForm.tsx L57–61） | 2 件目は **一覧 → 新規** を繰り返す。1 画面に残って連打する UI では無い |
| 失敗 | `handleApiError(error, "保存")`。フォーム残留 | 先行成功行は残る |

更新は別 `PATCH /v1/vaccinations/:id`。同日 2–3 件の「追加」に更新を使わない。

### BE Create（両フォーム共通）

[Create](../../../backend/internal/medicalrecord/vaccination_service.go) L138–195。

1. `vaccine_id` 必須。nil input 拒否。
2. 接種日は JST 今日以前（L300–309、「接種日は今日以前の日付を入力してください」）。
3. `next_date` があるとき接種日より後（L312–323）。
4. transactor 必須。同一 TX で `validateRelations` → `repo.Create` → `FindByID`。読取失敗は成功に反転しない。
5. 関係検証（L362–398）: 医院内 vaccine。pet があれば医院内 pet。`medical_record_id` があればカルテを `LockByIDForUpdate` し、owner/pet を医院内に固定。**カルテの pet と body の pet が違えば NotFound**（L388–390）。doctor があれば医院内 staff。
6. **ペット種 vs `vaccines.species` は Create で見ていない。** 種別不一致を接種 POST の 4xx と読まない（UAT-Q2）。
7. タグ同期 `syncVaccineTag` は commit 後 best-effort（L193、L280–287）。タグ失敗を登録失敗と混ぜない。
8. bind: [createVaccinationRequest](../../../backend/internal/medicalrecord/vaccination_request.go) L64–79。`vaccine_id` / `date` required。lot1–4 / next_date / remarks。**price/amount フィールド無し**。[Vaccination モデル](../../../backend/internal/model/vaccination_record.go) L18–33 も同じ。

日付パーサは YYYY-MM-DD と RFC3339 の両方（[date.go](../../../backend/internal/httpapi/date.go) L32–33）。カルテの date-only と独立画面の RFC3339 を「形式が違うから片方だけ登録できない」と決めない。実 4xx は採取する。

## 単件失敗カタログ（F）と順次 2–3 件（S）

再現は **保存済みカルテ・対象ペット・必須項目・権限・エラー表示・species 候補を固定してから**、単件成功 → 同日別接種 2–3 件を順に保存 → 一覧再読込、の順にする。単件が赤い状態で N 件 UI を足さない。

### F — 登録失敗（1 POST が 2xx にならない）

| ID | 観察 | コード根拠 | 同日複数との関係 |
| --- | --- | --- | --- |
| F1 未保存カルテ | タブが SaveRequired、または hook が medicalRecordId 欠落 | ServiceTabs L21–32、hook L142–145 | 初診で「打てない」はこれが先。batch ではない |
| F2 種別未選択 | fieldError。API 未呼び出し | 両フォーム必須文言が一致（hook L129–132、model L97–99） | |
| F3 接種日空 / 未来日 | FE は独立画面のみ未来日。BE は両方 | model L100–108、service L145–147 | |
| F4 次回 ≤ 接種日 | FE+BE | model L115–117、service L148–150 | 臨床間隔の代行ではない |
| F5 権限 | 独立画面 toast。カルテ hook は未チェック | VaccinationForm L20, L67、helpers L298–300 | カルテ側で打てて独立で打てないなら権限差 |
| F6 死亡ペット | 独立 FE のみ拒否。BE Create に死亡ガード無し | helpers L302–304。repository の死亡除外は **一覧**テスト（vaccination_repository_test.go L416） | 登録失敗と一覧非表示を混ぜない |
| F7 関係 / 医院 | vaccine/pet/カルテ/doctor の clinic 不一致、カルテ pet 不一致 | validateRelations | クロスクリニックを「複数入力バグ」にしない |
| F8 無音 | カルテ hook の petId 欠落 | hook L141 | 「登録できない」報告の採取対象 |
| F9 bind | vaccine_id / date 欠落は 400 | handler L92–95 | |

実ステータス・文言は医院端末で未採取なら各行 **再現 BLOCKED**。

### S — 同日 2–3 件（順次単件。batch ではない）

コード上の最小手順（実装しない。受入手順の草案）:

1. 単件 F が無いことを 1 件目で確認する（C なら保存済み `recordId`、I なら `canCreate` と生存ペット）。
2. 1 件目: ワクチン A、実施日 D（JST 当日可）、lot/次回/備考を入れて保存。2xx と toast。履歴に A が 1 行。
3. 2 件目: **新しいフォーム状態**でワクチン B（または医院が同一ワクチン再接種と認めた場合は A）、同じ D、**別の lot/次回**。保存。1 件目の lot が変わっていないこと。
4. 3 件目も同様。
5. 再読込: カルテ履歴は pet query 再取得、独立画面は list。3 行が別 `id`、別 `vaccine_id`、別 lot/next_date。

カルテ経路の操作数: 成功のたびに `isAdding=false` になるため「記録を追加」× N。独立経路は成功のたびに list へ戻るため **新規画面を N 回開く**。操作数削減の一括 UI は PO 裁定後。現行契約は N 回の単件 POST。

2 件目が失敗したとき 1 件目は commit 済み。部分保存の取消・補償 API は無い。これを「batch が無いから壊れている」と書かない。扱いだけ PO。

## lot / 次回予定 / 金額が混ざる点

| フィールド | 永続先 | 混ざり方 | 分離 |
| --- | --- | --- | --- |
| 実施日 `date` | `vaccinations.date`（行ごと） | 2 件目入力中に 1 件目を開いて上書き保存すると変わる。順次 **Create** では 1 件目を触らない | 受入は Create 連続。誤って 1 件目 PATCH しない |
| lot1–4 | 接種行 | カルテの「複製」は lot/次回/備考をコピーし実施日は空（hook L72–81）。複製後に保存すると **新しい行**。1 件目 lot は残る | 複製を「同じ行の再登録」と読まない |
| 次回予定 | `next_date` / `next_schedule_type` | 実施日や種別変更で UI が再計算（カルテ hook L93–121、独立 `vaccinationOverridesOn*`） | 2 件目の計算結果を 1 件目へコピーしない。間隔の臨床妥当性は見ない |
| 金額 | **接種 POST に無い**。マスタ `vaccines.price`（[vaccine.go](../../../backend/internal/model/vaccine.go) L21）。FE 変換は `price ?? 0`（[treatment.ts](../../../frontend/src/lib/transforms/treatment.ts) L80） | 履歴に出る price はマスタ。未設定マスタは FE で 0 に見える。会計は `Price == nil` を fail-closed（[billing_item_repository_vaccination_lock.go](../../../backend/internal/billing/billing_item_repository_vaccination_lock.go) L74–88） | 「接種が登録できた」≠「請求単価が接種行に保存された」。0 表示と未設定マスタを混ぜない |
| 会計 | 請求は接種 `id` に紐づく billing item。`vaccinationRef.MedicalRecordID == nil` は請求拒否（同ファイル L43–45）。確認前も Conflict（L55–57） | 独立画面の `medical_record_id: null` 行は登録できても請求できない | 登録失敗と請求不可を混ぜない。MASTER 価格欠落は [UAT-R2-MASTER-PATH](../../../todo-issue.md#uat-r2-master-path) |

## 種別条件（本票では変更しない）

[UAT-Q2-VACCINE-SPECIES](../todo-campaign-20260918/UAT-Q2-VACCINE-SPECIES.md) を正本とする。圧縮:

- 現行 [useGetAllVaccinesMaster](../../../frontend/src/hooks/use-treatment-master.ts) L36–46 は `GET /v1/masters/vaccines` で **species を送らない**。両フォームの候補は `isActive` のみ（カルテ hook L51–54、独立 form L45–48）。
- repository の species フィルタは完全一致。`cat` に `both` は自動で含まれない（種票の将来条件であり、現行フォームの実装済みという意味ではない）。
- 履歴の犬用名称は誤参照の証拠ではない。名称から種を推測して補完しない。一括削除しない。
- 本 unit の 2–3 件再現でも、候補に出たマスタを種不一致として自動除外しない。種品質の集計は種票の読取条件が揃うまで BLOCKED。

## 自動回帰（本 unit で追加）

製品 Create 契約は変更せず、次を scoped テストで固定する。batch `/vaccinations/batch` や配列 body は追加しない。

| 面 | ファイル | カバー |
| --- | --- | --- |
| FE mutation | [use-create-vaccination.test.ts](../../../frontend/src/hooks/use-create-vaccination.test.ts) | 単件 POST のみ・amount 無し・4xx で `handleApiError`・同日 2–3 件順次 POST で vaccine/lot/next_date 非混在 |
| BE service | [vaccination_service_test.go](../../../backend/internal/medicalrecord/vaccination_service_test.go) | vaccine 関係欠落で Create 拒否、同日 3 件順次 Create で行分離 |
| BE handler | [vaccination_handler_test.go](../../../backend/internal/medicalrecord/vaccination_handler_test.go) | service invalid input → 400、vaccine NotFound → 404 |

種別条件は [UAT-Q2-VACCINE-SPECIES](../todo-campaign-20260918/UAT-Q2-VACCINE-SPECIES.md) 正本のまま。本票・本 unit テストは species フィルタやマスタ補完を変更しない。

## 停止条件

- 単件 POST があることだけを根拠に batch API を追加しない。
- 部分成功の自動 rollback / 全件取消を、PO 無しで入れない。
- 接種間隔・次回予定の臨床妥当性をコードが代行しない。
- 種別マスタの NULL 補完、履歴の一括削除、他医院マスタ結合をしない。
- 接種行へ amount 列を足して「金額保存」を満たしたことにしない。会計・マスタ単価と混ぜて設計しない。
- 実機未採取の失敗原因を 1 つに断定しない。

## 受入で先に採るもの（実機・PO。自動テストでは代替しない）

1. 画面（カルテタブ / 独立）。保存済みカルテか新規か。権限。ペット生存。
2. 単件: 必須を空にした失敗、正常 1 件の 2xx、履歴 1 行、lot/次回がその行だけ。
3. 同日 2 件目・3 件目を **別 Create**。各行の vaccine / date / lot / next_date。1 件目が変わっていないこと。
4. 2 件目だけ失敗させたときの 1 件目残存（PO が部分保存をどう扱うかの材料。自動補償はしない）。
5. 金額はマスタ単価と会計の別証拠。接種 GET に amount が無いことを確認する。

呼び出し行: **無い。**
