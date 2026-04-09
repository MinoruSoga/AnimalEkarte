# BUG-184: medical-records 内 ExaminationGroup・ImageGalleryGroup・VitalsTab のハードコードカラー違反

## 概要

`features/medical-records/components/` 配下の `ExaminationGroup.tsx`・`ImageGalleryGroup.tsx`・`VitalsTab/VitalsTab.tsx` で Tailwind プリセットカラーを使用している。特に検査値の HIGH/LOW/NORMAL 状態表示に `text-red-600`・`text-blue-600`・`bg-red-500` がハードコードされており、デザイントークン体系外で色が決まっている。

## 再現手順

1. 電子カルテ画面の「検査」タブを開く
2. 高値（HIGH）・低値（LOW）を持つ検査結果を表示する
3. **結果**: HIGH = 赤・LOW = 青・NORMAL = 緑 で表示されるが、すべて Tailwind プリセットで定義されている
4. 同じく「バイタル」タブの追加・ホバー時のグレー背景も Tailwind プリセット

## 期待する動作

- HIGH/LOW/NORMAL の検査値色はデザイントークンまたは `status-helpers.ts` の関数で定義する
- バイタルタブのグレー背景は `C.bgHover` / `C.bgPage` 系トークンを使用する

## 現状コード

### `frontend/src/features/medical-records/components/ExaminationGroup.tsx:87-89`
```tsx
// ❌ 検査値 HIGH/LOW テキスト色ハードコード
<span className={result.flag === "H" ? "text-red-600" : result.flag === "L" ? "text-blue-600" : ""}>
  {result.value}
</span>
```

### `frontend/src/features/medical-records/components/ExaminationGroup.tsx:105-106`
```tsx
// ❌ HIGH バッジ — bg-red-500 ハードコード
<span className="bg-red-500 hover:bg-red-600 text-white text-xs px-1 rounded">
  HIGH
</span>
```

### `frontend/src/features/medical-records/components/ExaminationGroup.tsx:113`
```tsx
// ❌ LOW バッジ — blue ハードコード
<span className="text-blue-600 border-blue-600 bg-blue-50 text-xs px-1 rounded border">
  LOW
</span>
```

### `frontend/src/features/medical-records/components/ExaminationGroup.tsx:119`
```tsx
// ❌ NORMAL アイコン色ハードコード
<CheckCircle className="text-green-500/50" />
```

### `frontend/src/features/medical-records/components/ImageGalleryGroup.tsx:96,98`
```tsx
// ❌ 削除ボタン hover 色ハードコード
<button className="border border-red-200 hover:bg-red-50 ...">
  <Trash2 className="text-red-500" />
</button>
```

### `frontend/src/features/medical-records/components/ImageGalleryGroup.tsx:109`
```tsx
// ❌ ダウンロードアイコン hover 色ハードコード
<Download className="group-hover:text-blue-600" />
```

### `frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx:158-159,173-174`
```tsx
// ❌ バイタル追加行ホバー・背景ハードコード
<tr className="bg-gray-50 hover:bg-gray-100 cursor-pointer">
<td className="text-gray-400">未記録</td>
```

### 比較: 正しい実装
```tsx
import { C, BADGE, STYLE } from '@/lib/design-tokens';

// ✅ 検査値フラグ色
const getExamFlagStyle = (flag: string): React.CSSProperties => {
  if (flag === 'H') return { color: C.bgDanger };
  if (flag === 'L') return { color: C.bgAccent };
  return {};
};

// ✅ HIGH/LOW バッジ
<span style={flag === 'H' ? BADGE.red : BADGE.blue}>
  {flag === 'H' ? 'HIGH' : 'LOW'}
</span>

// ✅ 削除ボタン
<button style={{ borderColor: `${C.bgDanger}40` }} className={STYLE.btnDangerGhost}>
  <Trash2 style={{ color: C.bgDanger }} />
</button>

// ✅ グレー背景
<tr style={{ backgroundColor: C.bgHover }}>
<td style={{ color: C.textSecondary }}>未記録</td>
```

## 影響範囲

| 対象ファイル | 違反箇所数 | 状態 |
|---|---|---|
| `features/medical-records/components/ExaminationGroup.tsx` | 4箇所 (L87-89, L105-106, L113, L119) | 未修正 |
| `features/medical-records/components/ImageGalleryGroup.tsx` | 3箇所 (L96, L98, L109) | 未修正 |
| `features/medical-records/components/VitalsTab/VitalsTab.tsx` | 4箇所 (L158, L159, L173, L174) | 未修正 |

## 修正方針

### 1. `ExaminationGroup.tsx` — 検査値フラグ色を関数定義してトークン使用
```tsx
import { C, BADGE } from '@/lib/design-tokens';

// 検査値フラグスタイル関数
const getExamValueStyle = (flag?: string): React.CSSProperties => {
  if (flag === 'H') return { color: C.bgDanger };
  if (flag === 'L') return { color: C.bgAccent };
  return {};
};

// テキスト
<span style={getExamValueStyle(result.flag)}>{result.value}</span>

// バッジ
{result.flag === 'H' && <span style={BADGE.red} className="text-xs px-1 rounded">HIGH</span>}
{result.flag === 'L' && <span style={BADGE.blue} className="text-xs px-1 rounded border">LOW</span>}
{result.flag === 'N' && <CheckCircle style={{ color: `${C.bgStatusGreenDot}80` }} />}
```

### 2. `ImageGalleryGroup.tsx` — 削除・ダウンロードボタン色をトークンに
```tsx
// 削除ボタン
<button
  style={{ borderColor: `${C.bgDanger}30` }}
  className="hover:bg-[#FFE2DD] rounded p-1"
>
  <Trash2 style={{ color: C.bgDanger }} />
</button>

// ダウンロードアイコン hover
<Download className="group-hover:text-[#2383E2]" />
// または style で
style={{ color: undefined }}
onMouseEnter={e => e.currentTarget.style.color = C.bgAccent}
```

### 3. `VitalsTab.tsx` — グレー系をトークンに
```tsx
import { C } from '@/lib/design-tokens';

<tr style={{ backgroundColor: C.bgHover }} className="cursor-pointer">
<td style={{ color: C.textSecondary }}>未記録</td>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/rules/code-style.md` — Styling & Design Tokens
> **MANDATORY**: すべてのスタイリング（Tailwind 4, Inline styles）で `src/lib/design-tokens.ts` の定数 (`C`, `STYLE`) を使用する。

検査値フラグ（HIGH/LOW/NORMAL）は業務上意味のある状態分類であり、デザイントークンとして明示的に定義することで将来の仕様変更にも対応しやすくなる。

### プロジェクト内参照実装
- `utils/status-helpers.ts` — フラグ → スタイルマッピングの参照実装パターン
- `features/medical-records/routes/MedicalRecordForm.tsx` — 正しい削除ボタン実装

## 優先度
**Medium** — 電子カルテの検査値表示は臨床判断に使われる重要情報であり、色の一貫性は信頼性に関わる。デザイントークンへの統一で将来の保守が容易になる。

## 関連チケット
- BUG-162: medical-records 他のハードコード違反
- BUG-170: Hospitalization の同様のカラーマッピング関数違反

## 関連ファイル
- `frontend/src/features/medical-records/components/ExaminationGroup.tsx`
- `frontend/src/features/medical-records/components/ImageGalleryGroup.tsx`
- `frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx`
- `frontend/src/lib/design-tokens.ts`
