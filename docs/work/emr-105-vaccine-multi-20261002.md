# EMR-105: 同日複数ワクチン登録（保存済みカルテ内経路）— 検証記録

状態: **検証済み・FAIL なし・プロダクトコード変更なし**（テスト追加のみ）。

本票は **カルテ内（in-chart）ワクチン登録経路** の検証記録。単独フォーム経路は
2026-09-23 に実保存証拠済み（`docs/ops/testing/scenarios/S29-same-day-multiple-vaccinations.md`、
UAT 証跡 `reports/uat-2026-09-23/S29`）であり、本票では再実行しない。

## 2 経路の違い（検証対象の特定）

| 経路 | 入口 | medical_record_id | 既検証状況 |
|---|---|---|---|
| 単独フォーム | `features/vaccinations/...`（ワクチン管理の独立フォーム） | 送らない（null で保存） | 2026-09-23 S29 実保存証拠済み |
| **カルテ内フォーム（本票）** | `features/medical-records/components/MedicalRecordVaccination.tsx` — 保存済みカルテ内の「記録を追加」→ `VaccinationForm` | **必須**（`useMedicalRecordVaccinationForm(petId, medicalRecordId)` が両引数を要求し `medical_record_id: Number(medicalRecordId)` を常時送信） | 本票で検証 |

- カルテ内フォームは親カルテフォーム内に描画されるため nested `<form>` を避け、
  `SubmitButton formAction={formAction}`（`useActionState`）で送信する。
- 1 回の送信 = 1 件の `POST /v1/vaccinations`。バッチ登録・部分成功は仕様外。
- 成功時 `resetForm` で全フィールドを初期化し、React Query の `vaccinations` invalidation で
  `GET /v1/vaccinations?pet_id=...` を再取得 → `VaccinationHistory` が行を再描画する。

## 操作列・ペイロード・保存行・再表示

### バックエンド（実 DB）

`backend/internal/medicalrecord/vaccination_same_day_multi_db_test.go`
`TestVaccinationService_SameDayMultiVaccineInChartRealDB`

- fixture: clinic=1 / owner / pet / ワクチン A(¥3000)・B(¥5500)・C(¥8000)・単独経路(¥1200) /
  保存済み draft カルテ `EMR105-SAMEDAY`
- 操作列: 同一日（JST 当日）に `svc.Create` を逐次 4 回（3 回は `MedicalRecordID=&record.ID`
  のカルテ内経路、1 回は `MedicalRecordID=nil` の単独経路対照）
- 各入力（同日 `date=today`）:

| # | 経路 | vaccine_id | lot1 | next_date | remarks |
|---|---|---|---|---|---|
| 1 | カルテ内 | A | LOT-A | today+28d | 1回目 |
| 2 | カルテ内 | B | LOT-B | today+35d | 2回目 |
| 3 | カルテ内 | C | LOT-C | today+365d | 3回目 |
| 4 | 単独（対照） | 単独経路 | LOT-S | today+14d | 単独フォーム対照 |

- 保存行: `vaccination_id=1..4` の独立 4 行（サービス INFO ログで確認）
- 再取得: `svc.List(clinic, pet, start=end=today)` → total=4 行、各行にて
  `vaccine.id == row.vaccine_id`（preload 解決）、`lot1`、`remarks`、`vaccine.price`、
  `next_date`（暦日比較）、`pet_id` が行自身の値。カルテ内 3 行は
  `medical_record_id == record.ID`、対照行は `medical_record_id IS NULL`。
- 詳細再取得: `svc.GetByID` ×3 行でも `lot1` / `vaccine_id` / `medical_record_id` / `next_date` が一致。

### フロントエンド（axios モック + 実フォーム / 実フック）

`MedicalRecordVaccination.real-form.test.tsx`（EMR-105 describe、同日 2 件）:

- 保存済みカルテ（petId=1, medicalRecordId=99）で「記録を追加」→ 実 `VaccinationForm`。
- 1 件目: 予防接種名=7 / LOT1=`LOT-A` / 次回=`2026-09-26`（接種日は JST 当日 default `2026-08-29`）
  → 成功 → フォームが閉じ一覧に戻る（resetForm 実証）。
- 2 件目: 開き直し → 予防接種名=8 / `LOT-B` / `2026-10-10`。
- axios モックは POST body を **そのままエコー** して store に積むため、
  再取得された行は「各登録が送った実値」を反映する。

`use-medical-record-vaccination-form.test.ts`（EMR-105 describe）:

- `useMedicalRecordVaccinationForm("1","99")` で 2 回連続 submit。
- 1 回目成功後 `isAdding=false` / `lot1=""`（state が次登録へ漏れない）を確認。

## 判定

| # | チェック | 判定 | 証跡 |
|---|---|---|---|
| A1-1 | カルテ内経路の同一日 3 登録が独立行として永続化 | PASS | `vaccination_id=1..3` INFO ログ、`assert.Len(createdIDs,4)` + ID 相異 |
| A1-2 | 行ごとの vaccine 紐付け / lot / remarks / 金額(vaccine.price) / next_date | PASS | `assertRow` + `wantNextByVaccine` 暦日比較（List 再表示） |
| A1-3 | カルテ内行は `medical_record_id` を保持、対照行は NULL | PASS | List/GetByID 双方で `record.ID` / `Nil` |
| A1-4 | 後日フェッチ（pet スコープ再表示）で混在なし | PASS | `svc.List` total=4、`seenIDs` 全件ヒット |
| A2-1 | in-chart フォームは保存済み medical_record_id 必須 | PASS | 全 POST が `medical_record_id: 99` を含む（hook は引数必須） |
| A2-2 | 各送信は単一 create payload（逐次・バッチなし） | PASS | `mockPost` 2 回・各 `objectContaining` で個別値を照合 |
| A2-3 | 同日連続登録で state/payload が混ざらない | PASS | POST#1 `{vaccine_id:7,lot1:"LOT-A",next_date:"2026-09-26"}` / POST#2 `{vaccine_id:8,lot1:"LOT-B",next_date:"2026-10-10"}`、date は両方 `2026-08-29` |
| A2-4 | 成功時 reset → 次登録へ state 漏洩なし | PASS | `isAdding=false`/`lot1=""`（hook）＋一覧復帰→再オープン（実フォーム） |
| A2-5 | 履歴再表示が行を独立描画 | PASS | `既存ワクチン`/`新混合ワクチン`/`狂犬病ワクチン`、次予定 `26/9/26` と `26/10/10` が各行 |
| A3 | 本証跡ドキュメント | PASS | 本ファイル |

## 検証コマンド receipt（この worktree で実走）

### バックエンド（広範囲パッケージ、`-p 1` 直列・`--entrypoint go` 必須）

```text
$ docker run --rm --network animalekarte_ekarte-network \
    --env-file /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte/.env.local \
    -e APP_ENV=test -e DB_HOST=db -e DB_PORT=5432 -e TZ=Asia/Tokyo -e GOMAXPROCS=4 \
    -v /private/tmp/ae-emr-105-r4/backend:/app \
    -v /private/tmp/ae-emr-105-r4/frontend:/frontend:ro \
    -v /private/tmp/ae-emr-105-r4/docs:/docs:ro \
    -v /private/tmp/ae-emr-105-r4/scripts:/scripts:ro \
    -v /private/tmp/ae-emr-105-r4/Makefile:/Makefile:ro \
    -v ekarte-go-mod-cache:/go/pkg/mod -v ekarte-go-build-cache:/root/.cache/go-build \
    -w /app --entrypoint go animalekarte-backend \
    test -count=1 -p 1 ./internal/medicalrecord/...

ok  	github.com/animal-ekarte/backend/internal/medicalrecord	101.529s
```

新規テスト単体:

```text
=== RUN   TestVaccinationService_SameDayMultiVaccineInChartRealDB
2026/10/02 15:14:22 INFO vaccination created vaccination_id=1 clinic_id=1
2026/10/02 15:14:22 INFO vaccination created vaccination_id=2 clinic_id=1
2026/10/02 15:14:22 INFO vaccination created vaccination_id=3 clinic_id=1
2026/10/02 15:14:22 INFO vaccination created vaccination_id=4 clinic_id=1
--- PASS: TestVaccinationService_SameDayMultiVaccineInChartRealDB (0.49s)
PASS
ok  	github.com/animal-ekarte/backend/internal/medicalrecord	0.493s
```

### フロントエンド（スコープ vitest / eslint）

```text
$ docker run --rm -v /private/tmp/ae-emr-105-r4/frontend:/app \
    -v ekarte-frontend-node-modules:/app/node_modules -w /app animalekarte-frontend \
    npx vitest run src/features/medical-records

 ✓ src/features/medical-records/components/MedicalRecordVaccination.real-form.test.tsx (5 tests)
 ✓ src/features/medical-records/hooks/use-medical-record-vaccination-form.test.ts (8 tests)
 ✓ src/features/medical-records/components/VaccinationHistory.test.tsx (5 tests)
 ✓ src/features/medical-records/components/VaccinationForm.test.tsx (2 tests)
 ✓ src/features/medical-records/components/MedicalRecordVaccination.test.tsx (11 tests)
 ...
 Test Files  77 passed (77)
      Tests  635 passed (635)
   Duration  21.51s
```

```text
$ docker run --rm -v /private/tmp/ae-emr-105-r4/frontend:/app \
    -v ekarte-frontend-node-modules:/app/node_modules -w /app animalekarte-frontend \
    npx eslint src/features/medical-records
(exit 0 — 出力なし)
```

### ドキュメント検証

```text
$ F=/private/tmp/ae-emr-105-r4/docs/work/emr-105-vaccine-multi-20261002.md
$ test -s "$F" && grep -q "in-chart" "$F" && grep -Eq "PASS|FAIL" "$F"
(exit 0)
```

## 変更ファイル（テスト・証跡のみ。プロダクトコード差分なし）

- 追加: `backend/internal/medicalrecord/vaccination_same_day_multi_db_test.go`
- 変更: `frontend/src/features/medical-records/hooks/use-medical-record-vaccination-form.test.ts`（EMR-105 describe 追加）
- 変更: `frontend/src/features/medical-records/components/MedicalRecordVaccination.real-form.test.tsx`（master 3 択化・POST エコー store 化・EMR-105 describe 追加）
- 追加: 本ファイル

## Failure Signature

検証中にテスト側の時刻比較不備を 1 件検出・修正（製品 FAIL ではない）:

- expected: `next_date` が行自身の暦日と一致
- actual: `time.Equal` による instant 比較が失敗（`date` 列は UTC 0:00 で scan され、入力は JST 0:00）
- check: `go test -run TestVaccinationService_SameDayMultiVaccineInChartRealDB`
- signature: `next_due must be the row's own (vaccine=57 got 2026-10-30 00:00:00 +0000 UTC)`
- fix: `Format(time.DateOnly)` の暦日比較に変更（既存 repository テストと同一手法）
- result: PASS（製品コード差分なし）

## 残リスク / フォローアップ

- 実ブラウザでの 3 件以上 UI 連続入力は本票対象外（Vitest で 2 件・実 DB で 3 件をピン済み）。
- `amount` はワクチン行の金額（`vaccine.price`）を指す。接種行に数量・単価の別カラムは現行モデルに存在しない。
- `next_schedule_type` は行ごとに永続化されるが、本票の暦日・ロット・紐付け検証では補助フィールドとして扱った。
