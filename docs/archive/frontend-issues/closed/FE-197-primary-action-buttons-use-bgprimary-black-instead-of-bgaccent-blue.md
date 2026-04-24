# FE-197: 主要アクションボタンが C.bgAccent（青）でなく C.bgPrimary（黒）を使用

## 概要

`EstimateForm.tsx` と `ReservationDetailModal.tsx` の主要アクションボタンが、
デザインシステム規約に反して `C.bgPrimary`（黒 `#37352F`）を使用している。
主要アクション（作成・更新・カルテ作成）ボタンは `C.bgAccent`（青 `#2383E2`）で統一すべきである。

## 再現手順

1. `/estimates/new` — 見積書作成フォームを開く
2. ヘッダ右上の「作成」ボタン → **黒色**（期待：青）
3. 予約詳細モーダルを開く
4. フッタの「カルテ作成」ボタン → **黒色**（期待：青）

## 期待する動作

- 作成・更新などの主要アクションボタンは `C.bgAccent`（青）+ `C.bgAccentHover` を使用

## 現状コード

### `frontend/src/features/estimates/routes/EstimateForm.tsx:301-307`
```tsx
<SubmitButton
  size="sm"
  disabled={!form.title.trim()}
  className={`h-9 ${C.bgPrimary} ${C.hoverBgPrimaryDark} text-white text-sm`}
>
  {isEdit ? '更新' : '作成'}
</SubmitButton>
```

### `frontend/src/features/reservations/components/ReservationDetailModal.tsx:224-231`
```tsx
<Button
  size="sm"
  className={`${C.bgPrimary} text-white ${C.hoverBgPrimaryDark} h-9 text-sm gap-1.5 shadow-sm`}
  onClick={() => onCreateRecord(appointment)}
>
  <actionConfig.Icon className={ICON.action} />
  {actionConfig.label}
</Button>
```

### 比較: 正しい実装（VaccinationForm.tsx 等）
```tsx
<SubmitButton className={`${C.bgAccent} ${C.bgAccentHover} text-white`}>
  保存
</SubmitButton>
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/features/estimates/routes/EstimateForm.tsx:304` | 作成/更新ボタン | 要修正 |
| `frontend/src/features/reservations/components/ReservationDetailModal.tsx:226` | カルテ作成ボタン | 要修正 |

## 修正方針

### 1. `EstimateForm.tsx:304`
```tsx
// Before
className={`h-9 ${C.bgPrimary} ${C.hoverBgPrimaryDark} text-white text-sm`}

// After
className={`h-9 ${C.bgAccent} ${C.bgAccentHover} text-white text-sm`}
```

### 2. `ReservationDetailModal.tsx:226`
```tsx
// Before
className={`${C.bgPrimary} text-white ${C.hoverBgPrimaryDark} h-9 text-sm gap-1.5 shadow-sm`}

// After
className={`${C.bgAccent} text-white ${C.bgAccentHover} h-9 text-sm gap-1.5 shadow-sm`}
```

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — Design Tokens
> 色やスタイルは必ず `C`, `STYLE` 定数を使用。  
> 主要アクション（保存・作成・検索）ボタン: `C.bgAccent`（青）

### プロジェクト内参照実装
- `frontend/src/features/vaccinations/routes/VaccinationForm.tsx:168` — `C.bgAccent` + `C.bgAccentHover` で正しく実装

## 優先度
**Medium** — 機能的障害はないが、UI の一貫性が損なわれる。

## 関連ファイル
- `frontend/src/features/estimates/routes/EstimateForm.tsx:304`
- `frontend/src/features/reservations/components/ReservationDetailModal.tsx:226`
- `frontend/src/lib/design-tokens.ts` — C.bgAccent 定義
