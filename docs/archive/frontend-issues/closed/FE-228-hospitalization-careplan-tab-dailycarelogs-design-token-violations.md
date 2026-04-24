# FE-228: 入院 CarePlanTab・DailyCareLogsSection のデザイントークン違反

## 概要

FE-202 で未対応の hospitalization コンポーネント2ファイルに
直接 Tailwind カラークラスが使用されている。

## 違反ファイル

### `frontend/src/features/hospitalization/components/CarePlanTab/CarePlanTab.tsx`

| 行 | 違反コード | 用途 | 修正方針 |
|----|-----------|------|---------|
| 62 | `bg-blue-100 text-blue-700` | アクティブステータスバッジ | デザイントークンへ |
| 69 | `bg-green-100 text-green-700` | 完了ステータスバッジ | デザイントークンへ |
| 75 | `bg-gray-100 text-gray-500` | キャンセルステータスバッジ | `C.bgLight`, `C.text50` |
| 89 | `bg-indigo-100 text-indigo-700` | アクティブタイミングバッジ | デザイントークンへ |
| 90 | `bg-gray-50 text-gray-300` | 非アクティブタイミングバッジ | `C.bgLight`, `C.text30` |
| 126 | `bg-blue-50/50 border-blue-100` | 編集行のハイライト背景 | デザイントークンへ |

### `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyCareLogsSection.tsx`

| 行 | 違反コード | 用途 | 修正方針 |
|----|-----------|------|---------|
| 68 | `bg-orange-50 border-orange-100 text-orange-700` | 食事ケアログ | `design-tokens.ts` に `C.careType.food` 相当を追加 |
| 70 | `bg-teal-50 border-teal-100 text-teal-700` | 排泄ケアログ | 同上 |
| 72 | `bg-purple-50 border-purple-100 text-purple-700` | 投薬ケアログ | 同上 |
| 74 | `bg-red-50 border-red-100 text-red-700` | 処置ケアログ | 同上 |
| 76 | `bg-gray-50 border-gray-100 text-gray-700` | その他ケアログ | `C.bgLight`, `C.borderLight`, `C.text70` |

## 備考

FE-202 は以下のファイルを対象としており本チケットとは**重複しない**：
- `DischargeAlertDialog.tsx`
- `CarePlanItemRow.tsx`
- `DailyRecordTimeline.tsx`
- `TimingSection.tsx`
- `HospitalizationExpandedView.tsx`

ケアログタイプのカラーマップは `DailyRecordTimeline.tsx` の医療タイプバッジ（FE-202 対象）と
同様のパターンのため、`design-tokens.ts` への `C.careType` オブジェクト追加時に一括対応すると効率的。

## 準拠すべきプロジェクト規約

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリングで `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。
> **PROHIBITED**: 直接 Tailwind カラークラスの指定は厳禁。

## 優先度
**Medium** — 機能的障害なし。FE-202 の修正と同時に対応推奨。

## 関連ファイル
- `frontend/src/features/hospitalization/components/CarePlanTab/CarePlanTab.tsx`
- `frontend/src/features/hospitalization/components/DailyRecordsTab/DailyCareLogsSection.tsx`
- `frontend/src/lib/design-tokens.ts`
- 関連: FE-202（同 feature の先行デザイントークン修正チケット）
