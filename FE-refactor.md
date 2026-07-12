# FE-refactor.md — 残バックログ（第 4 期以降）

- **更新日**: 2026-07-12（HEAD `a7e4b7ce` + `44e35b3b` に対し全項目をコード実測で再調査し、意思決定パック化）
- **完了済み**: FE4-1〜18 / FE5-1〜4 / FE6-1〜9 / FE7-1〜3 / FE8-1〜4 / LinkOwner cross-clinic 検証（`44e35b3b`）。詳細は `git log --grep='FE4-\|FE5-\|FE6-\|FE7-\|FE8-'` が正本
- **クローズ・PO 決定の記録**: 本書には残さない。正本はメモリ `fe_backlog_decision_pack_20260711.md`（ラベル分岐 Q3-A 現状維持 / iso-date・design-tokens YAGNI / XSS 該当なし / tygo widen 上流解決 等）
- **本書の規約**: 行動可能な未対応タスクのみを記載する。各項目は「判断が必要な点」を明示し、判断が下れば追加調査なしで着手できる粒度にする

---

## 残件 2: medical-records `Treatment` の #201 未追随（#201 FE UI の一部として対応）

### 現状（2026-07-12 実測）

- **BE は実装済み**: `model/treatment.go:52-56` に `dose_weight_kg` / `dose_weight_source` / `dose_amount_mg` / `dose_amount_unit` / `dose_param_snapshot`（全て nullable + omitempty）。保存時再検証・逸脱監査（`AuditActionTreatmentDoseDeviation`）も稼働中。
- **生成型は追随済み**: `types/generated/models.ts:2962-2966` に dose 系フィールドが出力されている。`Medicine.dose_params?: MedicineDoseParam[]`・`MedicineDoseBasis` union も生成済み。
- **取り残されているのは手書き型のみ**: `features/medical-records/types/index.ts:27` の `Treatment` / `CreateTreatmentInput` / `UpdateTreatmentInput` に dose 系が無い。手書き型が存在する理由は **ID 表現の乖離**（FE は `id: string`、生成型は `id: number`）であり、生成型への単純置換はできない。

### 対応内容（Issue #201 OPEN・FE スコープは issue 本文 L153-215 に確定済み）

1. 手書き `Treatment` 3 型に dose 系 5 フィールドを追加（nullable・生成型と同名）。
2. `lib/medicine-dose.ts` 純粋関数を新設（BE と同一計算仕様・table-driven test 必須）。
3. `TreatmentsTab/`: 薬剤追加時に体重 + species からプリフィル、**丸め前/丸め後 mg を併記**、安全域逸脱の警告 + hard gate、手動上書き round-trip。
4. `MedicineSidePanel*`: 計算パラメータ（dose_per_kg 等）編集欄を追加（現状は name/price/tax 系のみで**欄が存在しない**）。

### 判断が必要な点

- 本リファクタ単独で着手しない方針は維持（型追加だけ先行しても UI が無ければ死にフィールド）。**#201 の FE 実装をいつやるかのスケジューリング判断のみ**が残っている。型追加 → UI の順で #201 チケット内で一括実施せよ。

---

## 残件 3: 低優先（各individually 着手可能）

### 3-3. `NextRecommendedVisitDate` 常時 null — **PO 判断待ち（選択肢確定済み）**

- **実測**: `liff_service_health_card.go:23` で宣言のみ・未代入。FE（`liff/src/pages/PetHealthPage.tsx:99`）は「次回推奨来院日: なし」を常時表示。一方、**同じ関数がワクチン接種記録の `NextDueAt`（次回接種予定日）を既に集約しており**、健康手帳のワクチンカードには次回予定が別途表示されている。
- **PO に提示すべき選択肢**:
  - (A) **フィールド削除** — ワクチン次回予定と表示が重複するなら、この欄自体が二重表示。response + FE 表示行を消す（PRODUCT_PHILOSOPHY ②削除。最小工数）。
  - (B) **ワクチン由来で算出** — `min(未来の NextDueAt)` を次回推奨来院日とする。データは同関数内に既にあり実装は 10 行程度。ただし「ワクチン以外（フィラリア・健診）を含まない」ことを仕様として明記する必要がある。
  - (C) **完全版** — フィラリア/ノミダニ予防スケジュール・健診周期まで統合。lstep 側に deadline タグ計算資産はあるが、健康手帳への転用は新規設計。工数大。
- **推奨**: PO への質問は 1 つ — 「この欄は何を約束する欄か」。回答が出るまで現状（null 表示）維持。捏造値の表示だけは絶対にしない。
