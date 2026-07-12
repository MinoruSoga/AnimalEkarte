# FE-refactor.md — 残バックログ（第 4 期以降）

- **更新日**: 2026-07-12（HEAD `a7e4b7ce` + `44e35b3b` に対し全項目をコード実測で再調査し、意思決定パック化）
- **完了済み**: FE4-1〜18 / FE5-1〜4 / FE6-1〜9 / FE7-1〜3 / FE8-1〜4 / LinkOwner cross-clinic 検証（`44e35b3b`）。詳細は `git log --grep='FE4-\|FE5-\|FE6-\|FE7-\|FE8-'` が正本
- **クローズ・PO 決定の記録**: 本書には残さない。正本はメモリ `fe_backlog_decision_pack_20260711.md`（ラベル分岐 Q3-A 現状維持 / iso-date・design-tokens YAGNI / XSS 該当なし / tygo widen 上流解決 等）
- **本書の規約**: 行動可能な未対応タスクのみを記載する。各項目は「判断が必要な点」を明示し、判断が下れば追加調査なしで着手できる粒度にする

---

## 残件 1: query-keys registry 全面採用 + clinic-id ルートキー化（長期・設計判断待ち）

### 現状（2026-07-12 実測）

- registry（`frontend/src/lib/query-keys.ts`）の収録は 3 エンティティのみ: `accountings.all/detail`・`masters.category`・`ME_QUERY_KEY`。
- inline `queryKey` は **187 ファイル**、`invalidateQueries` 呼び出しは **249 箇所**。feature 別の分布（上位）: master 30 / 共有 hooks 24 / medical-records 21 / accounting 14 / owners 12 / lstep 12 / shifts 10 / hospitalization 7 / vaccinations 6 / examinations 6。
- ad-hoc 文字列キーが registry 管理キーと**同一ファイル内で混在**する例が既にある（`["accounting-refunds", accountingId]` と `queryKeys.accountings.all()` が並ぶ）。キー命名の typo・階層不一致による invalidate 漏れはコンパイルで検知できない。

### 現在の安全性（なぜ今すぐ壊れないか）

clinic 切替は 3 層で守られている: ① `switchClinic` が reload 前に `queryClient.clear()`（FE5-3）、② マルチタブは storage イベント検知で reload（FE5-2）、③ 切替自体が full reload 前提（`clinic_context_architecture` 設計）。**キーに clinicId が無くてもキャッシュは切替時に全滅するため、漏れは発生しない**。つまり本件は「将来 SPA 切替（reload 廃止）に移行する場合の前提工事」であり、単独では緊急性がない。

### 判断が必要な点（これが決まれば着手可能）

1. **SPA 切替を将来やるのか** — やらない（reload 恒久）なら clinicId キー化は不要で、本件は「registry 統一による invalidate 信頼性向上」だけに縮小する（工数は約 1/3 になる）。**最初にこれを決めよ**。
2. キー形状 — `[clinicId, "entity", ...]`（ルート prefix）か `["entity", clinicId, ...]` か。ルート prefix は `queryClient.removeQueries({ queryKey: [oldClinicId] })` の一発無効化が可能で優位。
3. clinicId の注入方法 — key factory を純粋関数のまま保ち呼び出し側が渡すか（テスト容易・推奨）、factory 内部で auth store を読むか（利便性は高いが hooks 外から呼べず React 依存が漏れる）。
4. 再発防止 — ESLint `no-restricted-syntax` で「registry 外の inline queryKey 配列リテラル」を禁止するか。**やらないなら 187 ファイル移行後も新規逸脱が再蓄積する**ため、移行と同時導入すべきだ。

### 実行方針（判断後）

- feature 単位の段階移行（249 invalidate 箇所の対応キーを突合しながら）。挙動保存 — キー文字列が変わるため deploy 直後の初回 fetch は全 miss になるが、staleTime 内の再取得が走るだけで実害なし。
- 検証: feature ごとに scoped vitest + 当該画面の CRUD→リスト反映を手動確認。ESLint ルール導入時は既存違反 0 を移行完了の定義とする。

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
