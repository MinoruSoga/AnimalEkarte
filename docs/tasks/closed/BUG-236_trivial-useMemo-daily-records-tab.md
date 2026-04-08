# BUG-236: 不要な useMemo — DailyRecordsTab の getTodayStr()

## 概要
`DailyRecordsTab.tsx` で `useMemo(() => getTodayStr(), [])` を使って日付文字列をメモ化している。`getTodayStr()` は `new Date().toISOString().split("T")[0]` を返す軽量な純粋関数で、文字列（primitive）を返す。文字列は値で比較されるため、`useMemo` でメモ化しても `today` を依存配列に持つ下流の `useMemo` への影響は変わらない。`useMemo` 自体のオーバーヘッドが無駄になっている。

## 現状コード

### `features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx:45`
```typescript
// ❌ 軽量関数の戻り値（string primitive）を useMemo でメモ化
const today = useMemo(() => getTodayStr(), []);
// getTodayStr = () => new Date().toISOString().split("T")[0]

// today は以下 2つの useMemo の deps として使用
const effectiveMax = useMemo(
    () => (dischargeDate && dischargeDate < today ? dischargeDate : today),
    [dischargeDate, today]
);
const initialDate = useMemo(
    () => clampDate(today, admissionDate, effectiveMax),
    [today, admissionDate, effectiveMax]
);
```

## 分析

- `today` は string primitive — React の deps 比較は値比較（同じ日付なら同一）
- `useMemo(() => getTodayStr(), [])` を外しても `today` の文字列値は変わらない
- 下流の `useMemo` は string の値比較で正しく動作する
- `[]` deps で「マウント時一度だけ」は確かに意図的だが、date が変わらない限り毎回同じ値になる
- `useMemo` のメモ化コスト > 軽量文字列計算コスト

## 修正方針

```typescript
// ✅ 直接計算（または定数として定義）
const today = getTodayStr();

// 下流の useMemo はそのまま機能する（string は値比較のため）
const effectiveMax = useMemo(
    () => (dischargeDate && dischargeDate < today ? dischargeDate : today),
    [dischargeDate, today]
);
```

**補足**: もし「マウント後に日付が変わっても today を固定したい」という意図があるなら `useRef` パターンが正しい選択。ただし動物病院アプリのこのタブでは日付変更に追従すべきなので、直接呼び出しが適切。

## 準拠すべきプロジェクト規約

### `frontend/CODING_RULES.md` Section 12 — rerender-simple-expression-in-memo
> 軽量計算（boolean、単純プロパティアクセス、string 生成）には `useMemo` を使わない

### 関連チケット
- BUG-228: OwnerSearchModal / MedicineSettings の trivial useMemo（同パターン）

## 優先度
**Low** — 機能的影響なし。useMemo のオーバーヘッドを1件削除。修正は1行。

## 関連ファイル
- `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyRecordsTab.tsx:45`
