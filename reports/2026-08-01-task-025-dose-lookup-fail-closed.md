# TASK-025: dose-lookup technical failure fail-closed

**Date:** 2026-08-02  
**Issue:** #201 (technical failure slice; DEC-48)  
**Claim:** `claim/TASK-025`  
**Scope:** FE only. `backend/**` 無変更。missing data 挙動は TASK-033 まで HOLD。

## 欠陥（修正前）

### 1. TreatmentsTab prefill bare catch

`handleSelectFromMaster` が `fetchQuery` 失敗を bare `catch {}` で握り潰し、`quantity = 1` + `doseCalcInput = null` のまま `createTreatmentFn` に到達していた。  
`computeDoseGate(null, …)` はブロックしないため、ネットワーク/HTTP/parse 障害でも treatment が作成されていた。

### 2. TreatmentRow query error discard

`useMedicineDoseParams` から `data` のみを取り、`isError` を無視。query 失敗時 `doseParams` は `undefined` → `buildDoseCalcInput` が null → gate 非ブロック → `onUpdate` 実行。

### 共通の根

technical failure と missing data がどちらも `null` に潰されていた。

## 是正（修正後）

### 型で 3 分岐

| 層 | 型 | technical failure | missing data | ok |
|----|-----|-------------------|--------------|-----|
| lookup | `DoseParamsAuthority` (`medicine-dose-lookup.ts:45-48`) | `status: "failed"` | `success` + empty/ineligible via `buildDoseCalcInput` null | `success` + params |
| gate | `DoseGateSource` (`treatment-row-dose-gate.ts:20-23`) | `kind: "technical_failure"` | `kind: "missing"` | `kind: "ready"` |

固定文言: `DOSE_PARAMS_LOOKUP_FAILED_MESSAGE`（upstream body を含めない）。

### 修正箇所 2 経路（before → after）

| 経路 | before | after |
|------|--------|-------|
| Master create | bare catch → create | catch → `setMasterDoseBlockReason(固定文言)` + pending item + `return`（`TreatmentsTab.tsx:281-286`）。retry ボタンで再取得（`resetQueries` + re-enter） |
| Row update | `data` only → null gate open | `toDoseParamsAuthority` + `resolveDoseGateSource` → `technical_failure` で `isBlocked`、`onUpdate` 前に return（`TreatmentRow.tsx:92-120, 202-209`）。retry = `refetch()` |

## 追加テスト

### TreatmentsTab.test.tsx

- `dose-params が 500 のとき create せず technical failure を role=alert で表示する`（alert + createCount=0 + no upstream body）
- `dose-params technical failure 中は POST /treatments が 0 回のままである`
- `retry で dose-params が成功したら通常保存経路が再開する`
- missing data 3 件: 体重未記録 / petSpecies 未設定 / dose-params `[]`

### TreatmentRow.test.tsx

- `dose-params 取得失敗時に visible error を role=alert で表示し onUpdate を呼ばない`（`onUpdate` not called: L288）
- `retry 後に dose-params が成功したら安全域内 quantity の onUpdate が復元する`（L331）
- missing data 3 件: 体重欠落 / species 欠落 / 空配列

### treatment-row-dose-gate.test.ts

- `technical_failure` 固定文言・isBlocked
- `resolveDoseGateSource` が failed と null を同一視しない

## RED / GREEN（逐語要約）

コマンド:

```bash
docker compose exec -T frontend npx vitest run \
  src/features/medical-records/components/TreatmentsTab/TreatmentsTab.test.tsx \
  src/features/medical-records/components/TreatmentsTab/TreatmentRow.test.tsx \
  src/features/medical-records/components/TreatmentsTab/treatment-row-dose-gate.test.ts
```

**RED（実装前・型とテストのみ）:**

```
 Test Files  2 failed | 1 passed (3)
      Tests  5 failed | 21 passed (26)
```

Failed (5):

- dose-params が 500 のとき create せず technical failure を role=alert で表示する
- dose-params technical failure 中は POST /treatments が 0 回のままである
- retry で dose-params が成功したら通常保存経路が再開する
- dose-params 取得失敗時に visible error を role=alert で表示し onUpdate を呼ばない
- retry 後に dose-params が成功したら安全域内 quantity の onUpdate が復元する

**GREEN（コンポーネント配線後）:**

```
 Test Files  3 passed (3)
      Tests  26 passed (26)
```

## 回帰

```bash
docker compose exec -T frontend npx vitest run src/features/medical-records/components/TreatmentsTab/
```

| | Test Files | Tests |
|--|------------|-------|
| baseline | 4 passed | 20 passed |
| after | 4 passed | 34 passed |

新規 FAIL: 0。baseline FAIL 集合: 空。after FAIL 集合: 空 ⊆ baseline。

## BE 無変更

```bash
git diff --name-only HEAD -- backend/
```

出力: 空。

## silent fallback 残存調査（allowlist 内）

| ヒット | 判定 |
|--------|------|
| `TreatmentsTab.tsx` `catch { … return }` | **意図的 fail-closed**。create しない。fixed message only。 |
| `TreatmentRow.tsx` `parseFloat(localQuantity) \|\| 1` | 数量パース既定。lookup failure 経路では gate が先に return。 |
| `error.message` / `response.data` 表示 | **無し** |

## 未実測境界

- 実機ブラウザ / E2E / runtime 環境での手動確認は未実施。
- 臨床上限値・warning 帯の妥当性は本 unit の対象外（何も主張しない）。
- BE 再検証契約は読むだけで変更していない。

---

## Follow-up 是正（2026-08-02・reconciliation で検出）

初回実装 `eaa608b6a` の独立照合で 3 件を検出し、同一 unit 内で是正した。

### 1. `pending` を `success` として符号化していた（型の不正直）

`toDoseParamsAuthority` は取得中／未取得を `{ status: "success", params: [] }` として返していた。
挙動は missing に落ちるため既存と同じだが、**「取得成功したが当該 species の param が無い」と区別不能**になる。
`authority.status === "success"` を権威データの根拠に使う実装が将来必ず壊れる。型を正直にするために導入した
分岐が、逆に嘘をついていた。

修正: `{ status: "pending" }` を独立の分岐として追加。gate 上は missing 扱いのままで挙動は不変。
`resolveDoseGateSource(null, { status: "pending" })` → `missing`、
`resolveDoseGateSource(input, { status: "pending" })` → `ready` を test で固定した。

### 2. `computeDoseGate` が `DoseCalcInput | null` を受け付けたまま

`normalizeDoseGateSource` により `null` → `missing` の後方互換経路が残っていた。
production call site を全数確認したところ 3 つとも `DoseGateSource` へ移行済みで、
**旧シグネチャの使用者はテストのみ**だった。`null` 許容は「技術障害を欠落と同一視して保存を通す」経路が
型の上で復活していることを意味する。

修正: シグネチャを `DoseGateSource` のみへ narrowing、`normalizeDoseGateSource` を削除。
gate test 5 箇所を `{ kind: "missing" }` / `{ kind: "ready", input }` へ更新。

### 3. `resetQueries` の無条件実行（本 unit が持ち込んだ回帰）

失敗キャッシュ対策の `queryClient.resetQueries` が、**成功経路も含めた毎回の薬剤選択前**に実行されていた。

- (a) `QUERY_STALE_TIMES.STATIC` が無効化され、同一薬剤の再追加でも毎回ネットワークへ行く。
- (b) より重大 — 共有 queryKey をリセットするため、同一薬剤を購読中の `TreatmentRow` が一往復のあいだ
  `data: undefined` に落ち、上記 1 の写像により missing 扱いとなって**行の投与量ゲートが一時的に開く**。
  窓は狭いが、本タスクが潰しにいったバグクラスそのもので、修正前には存在しなかった。

修正: `handleRetryMasterDoseLookup` の内部へ移動。通常選択経路では実行しない。

### 是正後の検証

```
docker compose exec -T frontend npx vitest run src/features/medical-records/components/TreatmentsTab
  Test Files  4 passed (4)
       Tests  35 passed (35)      # 34 → 35（pending 分岐の test を追加）
```

`git diff --name-only HEAD -- backend/` は引き続き空。missing-data 挙動（体重未記録 / species 解決不能 /
param 未設定）は 3 ケースとも保存継続のまま不変で、TASK-033 の cutover 境界を侵していない。
