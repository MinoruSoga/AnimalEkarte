# BE-refactor.md — バックエンド リファクタリング（残オープン項目）

- **更新日**: 2026-07-12
- **本書の規約**: 判断待ち・未実装フォローアップのみを記載する。対応済みの詳細・手順は git 履歴が正本（本書には残さない）。判断が出たら下記をそのまま実装単位にしてよい。

### 検証コマンド規約（Docker 必須・スコープ限定）

- 必ず `docker compose exec backend go test ./internal/<pkg>/ -run <Name> -count=1`。**フル `go test ./...`・`golangci-lint run ./...`・`gofmt -w ./...` は実行禁止**。
- 変更した Go ファイルは `docker compose exec backend gofmt -l ./internal/<dir>/` 無出力を確認してからコミット。
- `Co-Authored-By` なし。**push しない**（依頼があるまで）。

---

## 判断待ち項目（PO / オーナー決定後に着手）

判断が出るまで実装しない。決定後は該当案の手順どおり実行する。

> **判断材料の再実測（2026-07-12 / HEAD `4f9550c9`）**: 全項目のアンカー・主張を並列調査 5 レーン + 手動 grep で現 HEAD に対して再検証した。初版の記述に**事実誤認 4 件**（X-16「封筒化が必要」「handler 先例 = owner_request.go」/ F6+H-3「唯一の category 述語」/ last_visit_date「pin 解消が動機」）が見つかり、各項目内で訂正済み。各項目は **決定者 → 判断質問（そのまま転送できる一文）→ 推奨 → 判断材料 → 判断肢別の実装手順 → 検証** で構成する。

### X-15. 状態トグル系 DELETE 4 ルートの "edit" 権限（P6 逸脱） — ✅ DONE (`d89df5f4`)

PO 判断: death 系 2 本 (a) P6 例外明文化（挙動保存・`.claude/refs/gin-architecture-compliance.md` +
`handler/CLAUDE.md` に注記済み）、lstep-opt-out 系 2 本 (c) ルート削除（`DeleteOwnerLstepOptOut`
handler・専用テスト・api.yaml オペレーションごと削除、FE 呼び出し 0 件を確認済み）。詳細は git 履歴参照。

### X-16. 健診一覧のページネーション — PO 判断（②のみ残・①は DONE）

**① 健診期限アラート API 削除 — ✅ DONE (`e4a059ef`)**（消費者 0 件のため `GetCheckupAlerts` /
`GetAlerts` / `FindAlerts` / `toCheckupAlertsResponse` 系 + api.yaml オペレーション + 専用テストを削除済み。
詳細は git 履歴参照）。

**② 健診一覧の実サーバページング移行 — 次期送り確定（2026-07-12 / PO 推奨案採用）**:
- **決定者**: PO。**PO 判断**: 今期は着手しない。STG 実件数に基づき、次回 BE リファクタサイクルで再検討する。
- **決定背景**: STG 件数が数千行未満なら**次期送り**（自動生成が無く増加は手入力に線形 — 緊急性の根拠が薄い）。
- **判断材料**:
  - BE は既に退化封筒を返している（`checkup_handler.go:142` `newPaginatedResponse(responses, int64(len(responses)), 1, len(responses))` — total=全件数・page=1 固定の見せかけページング）。FE も封筒を読むが total を捨て**クライアント側ページング**（`get-checkups.ts:15-16` / `CheckupsList.tsx:118-120` の `usePagination`）。実体は「退化封筒 → 実サーバページング移行」であり、封筒キーは既に一致 = 漸進移行が可能で破壊的変更ではない。
  - handler 先例: **`parsePagination`（`query_helpers.go:17-41` — 既定 page=1 / limit=20 / 上限 100・per_page エイリアス対応）+ `vaccination_handler.go:13-46` + `newPaginatedResponse`（`handler.go:38-48`）**。
  - repo 現状: `FindByClinicID`（`checkup_repository.go:44-70` — Limit/Offset なし・Preload 5 連鎖・全件 Find）。service `ListByClinic`（`checkup_service.go:118-131`）は total 非対応。checkup の生成は**手動 1 経路のみ**（`checkup_service.go:170` が唯一の Create。`checkup_sync_service_create.go` は名前に反し Lstep タグ付与のみで INSERT しない）。
  - FE 改修必要 3 ファイル: `features/checkups/api/get-checkups.ts`（:8 戻り値型 / :15-16 total 破棄 / :19-26 hook に page/limit）・`api/types.ts`（`CheckupFilters` :17-22 に page/limit 追加）・`routes/CheckupsList.tsx`（:100 配列前提 / :103-115 クライアント検索・ソートの再設計要 / :118-120 usePagination 置換 / :240,279-290 総件数と Pagination 供給）。**`medical-records/api/checkups.ts` の同名 `useGetCheckups` はサブリソース用の別物 — 触るな**。
  - FE 先例: **medical-records が唯一の完全先例**（`get-medical-records.ts:40-75` の {data,total,page,limit} 消費 + `use-medical-records.ts:67-69`）。vaccinations FE は checkups と同じ退化パターン（`get-vaccinations.ts:19-23` が total 破棄）— 先例に使うな。
  - STG 実件数 SQL（人間実行・判断の入力）:
    ```sql
    SELECT clinic_id, COUNT(*) FROM checkups WHERE deleted_at IS NULL GROUP BY clinic_id;
    ```
- **実装手順**: (1) repo を `vaccination_repository.go:35-70` の buildBase closure 同型へ（戻り値 `([]model.Checkup, int64, error)`）(2) service input/戻り値に page/limit/total (3) handler を `parsePagination` + 実 total の `newPaginatedResponse` へ (4) api.yaml 更新 (5) FE 3 ファイルを medical-records 先例で改修（クライアント検索・ソートをサーバ側に寄せるかは FE 設計判断が別途必要）。
- 検証: `docker compose exec backend go test ./internal/repository/ -run TestCheckupRepository -count=1` + `./internal/service/ -run TestCheckupService -count=1` + `./internal/apicontract/ -count=1` + FE `npx vitest run src/features/checkups`

### F6 + H-3. Lstep 系死にコード 4 メソッドの keep/delete — ✅ DONE (`0d192642`)

オーナー判断: (a) 全削除。`BulkAddOwnerTag` / `SyncPetSpeciesTags` / `SyncSeniorTag` /
`FindOwnersByCategoryPurchaseDate` と専用型・実装ファイル（`lstep_tag_sync_pet_species.go` /
`_senior.go` は丸ごと削除）・mock・専用テストを削除済み。`SyncChronicConditionTags` /
`SyncOwnerAnimalClassificationTags` / `HasFoodPurchaseByOwnerSince` と共有ヘルパーは無変更。
復元は本コミットの revert で可能。詳細は git 履歴参照。

### M2. MeResponse.AvatarURL の三方向 drift — 解消済み（コード変更不要）

2026-07-12 再調査時点で、api.yaml（MeResponse）・FE（`transforms.ts` / `types/auth.ts` / fixture）
のいずれにも `avatar_url` / `avatarUrl` 宣言は現存しない（grep 0 件、pnpm キャッシュのみヒット）。
本項目作成時点の記述が既に解消済みの状態を反映していなかったための誤登録。対応不要。

### last_visit_date の wire 形式 — ✅ DONE (`f74c9242`)

LIFF 健康手帳 API の `last_visit_date` を RFC3339 datetime から date-only 文字列
（`In(time.Local).Format(time.DateOnly)`）へ変更済み。api.yaml の format も `date` に統一、
stale だった drift pin も削除。詳細は git 履歴参照。

### CI 構成: Backend ジョブの Lint 直列ゲート — ワークフローオーナー判断（緊急度低）

- **決定者**: ワークフローオーナー。**判断質問（転送用）**: 「Backend CI で Lint が失敗しても Test を最後まで走らせる変更（continue-on-error + 最終集約ゲート、約 6 行）を承認しますか？ Lint 失敗が Test の赤を隠す構造（TestMedicalRecordTenantScope が 2 日隠れた実例）の恒久解消です」
- **推奨**: 肢(i)（continue-on-error + 集約 step）。最小差分・ガードレール非衝突・fail-fast マスクの恒久解消。
- **判断材料（2026-07-12 実測）**:
  - main は現在**緑**（`70f4c298` 2026-07-11 12:17 success）。直近失敗 run（`0676efa7`）の Backend 赤は **Test step**（Lint は通過）+ OpenAPI Date-Format Drift Gate で、同日中に解消済み。※7/11 時点の「Backend/Lint 赤」というメモは実測では Test 赤 — 訂正。
  - backend ジョブ（ci.yml:340-484）: Build(:394) → Lint(:398, golangci-lint-action@v9) → Test(:404) 直列。Lint に continue-on-error 無し = fail-fast マスク構造は残存。スキップ条件 :343 は `needs.changes.outputs.backend == 'true' && (github.event_name != 'pull_request' || github.base_ref != 'main')` — **PR→main では backend ジョブごとスキップ**（静的ゲート 6 種を無条件の独立ジョブにした理由。:193-195 コメント）。
  - ガードレール `scripts/check-ci-step-order.sh`（ci.yml:109-120 から呼出）は**ステップ名ベースの順序のみ**検証（gate 名完全一致 `Lint`/`Build`/`Type check` の最大 index < risk 名部分一致 `test`/`ratchet` の最小 index）。continue-on-error は検知しない → **両肢とも衝突なし**。
- **判断肢と実装コスト**:
  - **(i) continue-on-error + 集約 step（~6 行・推奨）**: Lint step（:398）に `id: lint` + `continue-on-error: true` を追加し、ジョブ末尾（Schema drift check :475 の後）に `if: always()` + `steps.lint.outcome != 'success'` で fail する集約 step を新設。**集約 step 名は完全一致 `Lint`/`Build` と部分一致 `test`/`ratchet` を避ける**（例: "Verify lint outcome" — ガードレールの gate/risk 判定に乗せないため）。
  - **(ii) Lint 独立ジョブ化（~20 行）**: 二次論点が発生する — 既存静的ゲートは needs/if 無しの全イベント無条件のため、同形にすると **PR→main でも Lint が走る挙動拡大**になる（現在はスキップ）。paths 条件を付けるか否かの追加判断が必要で (i) より重い。
- ワークフロー変更のためオーナー承認後に**別チケット・別コミット**で実施（本ファイルのコミット規約外）。

---

## オープン・フォローアップ（実装待ち / 人間作業）

### FOLLOWUP-X14A. MedicineID → DecreaseStock フォールバック（P1）

- X-14a は `InventoryID` 直接指定の越境を塞いだが、`InventoryID == nil` 時に `MedicineID` を inventory id として `DecreaseStock`（clinic 非参加）へ渡す経路が残る。
- 草案: `docs/tasks/open/FOLLOWUP-X14A-decrease-stock-medicine-fallback.md`（`.gitignore` 対象の場合あり）
- 対象: `treatment_service.go` Create 内在庫減算 / `inventory_repository.go` `DecreaseStock`

### STG クロステナント監査 SQL（人間実行・自動実行禁止）

ガード導入前から存在する越境データの有無確認。ヒット時は今後操作不能/不可視になり得る。

```sql
-- 1) treatments.inventory_id（X-14a）
SELECT t.id FROM treatments t JOIN inventory_items i ON i.id = t.inventory_id
WHERE t.inventory_id IS NOT NULL AND i.clinic_id <> t.clinic_id;

-- 2) trimming_courses.course_type_id（X-14b）
SELECT c.id FROM trimming_courses c JOIN trimming_course_types ct ON ct.id = c.course_type_id
WHERE c.course_type_id IS NOT NULL AND ct.clinic_id <> c.clinic_id;

-- 3) appointment_trimming_details.course_id（X-14c）
SELECT d.id FROM appointment_trimming_details d JOIN trimming_courses c ON c.id = d.course_id
WHERE d.course_id IS NOT NULL AND c.clinic_id <> d.clinic_id;

-- 4) appointment_trimming_options.option_id（X-14c、options 側に clinic_id 列なし → appointment 経由）
SELECT o.id FROM appointment_trimming_options o
JOIN appointment_trimming_details d ON d.appointment_id = o.appointment_id
JOIN trimming_options t ON t.id = o.option_id
WHERE t.clinic_id <> d.clinic_id;

-- 5) staff_permission_groups（H-1）
SELECT spg.staff_id, spg.group_id, pg.clinic_id AS group_clinic_id
FROM staff_permission_groups spg
JOIN permission_groups pg ON pg.id = spg.group_id
WHERE NOT EXISTS (
  SELECT 1 FROM staff_clinic_assignments sca
  WHERE sca.staff_id = spg.staff_id AND sca.clinic_id = pg.clinic_id AND sca.deleted_at IS NULL
);

-- 6) staff_reservation_exclusions（H-2）
SELECT sre.staff_id, sre.reservation_type_id, rt.clinic_id AS type_clinic_id
FROM staff_reservation_exclusions sre
JOIN reservation_types rt ON rt.id = sre.reservation_type_id
WHERE NOT EXISTS (
  SELECT 1 FROM staff_clinic_assignments sca
  WHERE sca.staff_id = sre.staff_id AND sca.clinic_id = rt.clinic_id AND sca.deleted_at IS NULL
);
```

### 別台帳

- `BE_todo.md` PERF-FOLLOWUP-05（password_reset goroutine の shutdown drain）— X-18 完了後の補完タスク。本ファイルのスコープ外。
