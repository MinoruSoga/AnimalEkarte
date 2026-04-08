# BUG-204: MedicalRecords.tsx・EstimateDetail.tsx の追加カラー違反

## 概要

`features/medical-records/routes/MedicalRecords.tsx:265,269` のスタッフ検証エラー表示に `text-red-500` がハードコードされており、`features/estimates/routes/EstimateDetail.tsx:75` の削除ボタンボーダーに `border-red-200` がハードコードされている。前者は BUG-185（MedicalRecordPrintView/VaccinationForm/TreatmentTable）、後者は BUG-198（EstimateDetail ローディング/エラー UI）と同一ファイルだが別箇所の未起票違反。

## 現状コード

### `frontend/src/features/medical-records/routes/MedicalRecords.tsx:265,269`
```tsx
// ❌ 無効スタッフの警告色ハードコード
// L265
<span className={!isValidStaff(r.doctor) ? "text-red-500 font-medium" : ""}>
  {r.doctor?.name}
</span>

// L269
<AlertTriangle className={`${ICON.xs} text-red-500`} />
```

### `frontend/src/features/estimates/routes/EstimateDetail.tsx:75`
```tsx
// ❌ 削除ボタンボーダーにハードコード（BUG-198 の loading/error 修正とは別箇所）
<button className="h-9 gap-1.5 text-sm border border-red-200 ...">
  削除
</button>
```

### 比較: 正しい実装
```tsx
import { C } from '@/lib/design-tokens';

// ✅ スタッフ検証エラー色
<span style={!isValidStaff(r.doctor) ? { color: C.bgDanger, fontWeight: 500 } : {}}>
  {r.doctor?.name}
</span>
<AlertTriangle style={{ color: C.bgDanger }} className={ICON.xs} />

// ✅ 削除ボタン（danger ghost ボタン）
<button
  style={{ borderColor: `${C.bgDanger}40` }}
  className="h-9 gap-1.5 text-sm border ..."
>
  削除
</button>
```

## 影響範囲

| 対象ファイル | 行番号 | 違反 | 状態 |
|---|---|---|---|
| `features/medical-records/routes/MedicalRecords.tsx` | 265 | text-red-500（スタッフ名無効表示） | 未修正 |
| `features/medical-records/routes/MedicalRecords.tsx` | 269 | text-red-500（警告アイコン） | 未修正 |
| `features/estimates/routes/EstimateDetail.tsx` | 75 | border-red-200（削除ボタン） | 未修正 |

## 修正方針

### `MedicalRecords.tsx:265,269`
```tsx
import { C } from '@/lib/design-tokens';

// L265
<span style={!isValidStaff(r.doctor) ? { color: C.bgDanger, fontWeight: 500 } : {}}>
  {r.doctor?.name}
</span>

// L269
<AlertTriangle style={{ color: C.bgDanger }} className={ICON.xs} />
```

### `EstimateDetail.tsx:75`
```tsx
import { C } from '@/lib/design-tokens';

// Before
<button className="... border border-red-200 ...">削除</button>

// After
<button
  style={{ borderColor: `${C.bgDanger}40` }}
  className="... border ..."
>
  削除
</button>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

## 優先度
**Low** — 機能的問題なし。BUG-185（MedicalRecords コンポーネント群）・BUG-198（EstimateDetail ローディング/エラー）の対応時に合わせて修正すること。

## 関連チケット
- BUG-185: medical-records 印刷ビュー・VaccinationForm の gray 系違反（同 feature）
- BUG-198: EstimateDetail のローディング/エラー UI 問題（同ファイル）

## 関連ファイル
- `frontend/src/features/medical-records/routes/MedicalRecords.tsx`
- `frontend/src/features/estimates/routes/EstimateDetail.tsx`
