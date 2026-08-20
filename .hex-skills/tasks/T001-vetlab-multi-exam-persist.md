# T001 — 検査機器受信: 1ジョブを exam_type ごとに複数 exam としてカルテへ保存する（VetLab 送信口）

**tracker_provider**: file (Linear MCP は本環境で利用不可 / tool-missing。スキル規約によりファイルモードで起票。)
**status**: In Progress
**created_at**: 2026-08-20
**task_type**: implementation

---

## Context (Why)

IDEXX VetLab は検査機器ではなく、複数の検査機器（CBC 機・生化学機など）の結果をまとめて送信する端末（送信口）である。
1回の受信フレームに複数の `exam_type` にまたがる項目が含まれる（例: CBC 結果 + 生化学結果）。

現状（ADR-007 §7 の実装）は「`exam_type_id` が2種以上なら保存拒否・`needs_review`」を行うが、これは VetLab の正常な送信を拒否してしまう。

判断の根拠:
- スタッフが機器を選ぶのではない。**送られてきた `device_item_code` → 検査機器項目マスタ → `exam_type_field` → `exam_type`** で検査を決める
- 同じ受信に検査 A の項目と検査 B の項目があれば、カルテに **A 検査と B 検査の 2 件**を作るのが正しい業務動作
- `lab_devices.exam_type_id` の複数化 schema は作らない。紐付けは既存の item master（code → field）が正

---

## Implementation Plan

### 1. Backend: `lab_device_exam_persist.go`

- `AssertSingleExamType` の呼び出しを削除
- 代わりに `UniqueMappedExamTypeIDs` で種別一覧を取得し、種別ごとに `PersistExam` を1回ずつ呼ぶ
- `PersistedCount = 合計 non-duplicate exams`, `DuplicateCount = 合計 duplicate exams`
- `markUsageTracking` には合計 `PersistedCount` を渡す

### 2. Backend: `lab_device_exam_persist_test.go`

- `TestLabDeviceExamPersister_MultipleExamTypesNeedsReview` を `TestLabDeviceExamPersister_MultipleExamTypesPersistsBoth` に変更
- 期待値: status = `persisted`、exam 2 件、detach で両方消える
- VetLab 専用テスト追加（`idexx_vetlab` source type で2種別）

### 3. Frontend: `lab-device-board-model.ts`

- `labDeviceNeedsReviewReason` の `lab_device_multiple_exam_types` 特有メッセージを除去（汎用メッセージへ fallthrough）
- 理由: 新ジョブでこのコードは設定されなくなる。旧ジョブが残った場合も「保存できません」は誤りなので汎用に変える

### 4. Frontend: `lab-device-board-model.test.ts`

- F-1 テストの `lab_device_multiple_exam_types` アサーションを更新

### 5. ADR-007 §7 最小追記

- 保存拒否条件の廃止と種別分割保存の追記

---

## Technical Approach

- `DetachDeviceJob` は `job_id` 由来の exams を全件取得・取り消す実装になっており、複数 exam にそのまま対応する（追加変更不要）
- `assertRevertSafe` は全 receipts・全 exams を走査するため複数 exam に対応済み
- `markUsageTracking` の `RowCount` は `PersistedCount` の合計値に変更する

---

## Acceptance Criteria

- [ ] VetLab（または任意 source）の1受信で、マップ済み項目が2つの exam_type に分かれるとき exams が2件作られる
- [ ] 1 exam_type だけなら exams 1件（回帰）
- [ ] 未マップのみなら既存どおり exam 0 / needs_review
- [ ] detach/undo でそのジョブの exam がすべて外れる
- [ ] clinic_id 分離を維持
- [ ] ボードに旧「検査種別が複数のため保存できません」の拒否が、正当な分割保存に置き換わる（未マップ混在の表示は残してよい）

---

## Affected Components

- `backend/internal/medicalrecord/lab_device_exam_persist.go`
- `backend/internal/medicalrecord/lab_device_exam_persist_test.go`
- `frontend/src/features/lab-device/lib/lab-device-board-model.ts`
- `frontend/src/features/lab-device/lib/lab-device-board-model.test.ts`
- `docs/architecture/adr/007-lab-device-receive-and-commit.md` (最小追記)

---

## Existing Code Impact

- `AssertSingleExamType` 関数と `LabDeviceMultipleExamTypesErrorCode` 定数は定義を残す（コンパイルエラー回避・既存テストの参照保持）
- `DetachDeviceJob` / `assertRevertSafe` は変更不要（既に複数 exam 対応済み）
- ADR-007 §7 の "保存拒否" 文言は最小追記で補足する

---

## Out of Scope

- グルーピング UI
- `lab_devices.exam_type_id` の複数化 schema
- `LAB_DEVICE_CONNECTIVITY.md` の編集
- 未知コードの catalog への推測追加

---

## Definition of Done

- [ ] Go テストが Docker でパス（scoped: `./internal/medicalrecord/...`）
- [ ] FE テストが Docker でパス（scoped: `lab-device-board-model.test.ts`）
- [ ] コミット未実施（ユーザー判断に委ねる）
