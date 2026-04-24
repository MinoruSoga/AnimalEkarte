# FE-231: medical-records コンポーネント群のデザイントークン違反

## 概要

`frontend/src/features/medical-records/components/` 配下の複数コンポーネントで
直接 Tailwind カラークラスが使用されている。

## 違反ファイル一覧

### `ExaminationGroup.tsx`

| 行 | 違反コード | 用途 |
|----|-----------|------|
| 87 | `text-red-600` | 検査値 HIGH（異常高値）のテキスト色 |
| 89 | `text-blue-600` | 検査値 LOW（異常低値）のテキスト色 |
| 105 | `bg-red-500 hover:bg-red-600` | HIGH バッジ背景 |
| 113 | `text-blue-600 border-blue-600 bg-blue-50` | LOW バッジ |
| 119 | `text-green-500/50` | 正常値チェックアイコン |

### `ImageGalleryGroup.tsx`

| 行 | 違反コード | 用途 |
|----|-----------|------|
| 96 | `border border-red-200 hover:bg-red-50` | 削除ボタンのボーダー・hover |
| 98 | `text-red-500` | 削除アイコン色 |
| 109 | `group-hover:text-blue-600` | 画像 hover テキスト色 |

### `TreatmentTable.tsx`

| 行 | 違反コード | 用途 |
|----|-----------|------|
| 127 | `hover:bg-gray-50` | 行 hover 背景 |

### `MedicalRecordPrintView.tsx` ⚠️ 多数

| 行 | 違反コード | 用途 |
|----|-----------|------|
| 69, 92 | `border border-gray-400` | 患者情報ボックスのボーダー |
| 74, 79, 85 | `text-gray-600` | 印刷ビューラベル |
| 101, 109, 117, 125, 133 | `border-b border-gray-300` | セクションヘッダー下線 |
| 156 | `border-t border-gray-300 text-gray-500` | フッター |

### `VaccinationForm.tsx`（medical-records コンポーネント内）

| 行 | 違反コード | 用途 |
|----|-----------|------|
| 164, 177, 190, 203 | `border-gray-400` | RadioGroup アイテムのボーダー（4箇所） |

## 備考

`MedicalRecordPrintView.tsx` は印刷用コンポーネントのため、
`@media print` スコープ内の色指定は別途検討が必要な可能性がある。
ただし現状は通常の Tailwind クラスを使用しているため規約違反として対応する。

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: 直接 Tailwind カラークラスの指定は厳禁。

## 優先度
**Medium** — `ExaminationGroup` の HIGH/LOW バッジは診断情報の視認性に影響するため優先度高め。

## 関連ファイル
- `frontend/src/features/medical-records/components/ExaminationGroup.tsx`
- `frontend/src/features/medical-records/components/ImageGalleryGroup.tsx`
- `frontend/src/features/medical-records/components/TreatmentTable.tsx`
- `frontend/src/features/medical-records/components/MedicalRecordPrintView.tsx`
- `frontend/src/features/medical-records/components/VaccinationForm.tsx`
- `frontend/src/lib/design-tokens.ts`
