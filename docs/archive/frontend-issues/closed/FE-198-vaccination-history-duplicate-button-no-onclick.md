# FE-198: VaccinationHistory「複製」ボタンに onClick がなく機能しない

## 概要

カルテの予防接種タブにある `VaccinationHistory` コンポーネントの「複製」ボタンに
`onClick` ハンドラが存在しない。クリックしても何も起きない死ボタンになっている。

## 再現手順

1. 任意のカルテを開く → 「予防接種」タブ
2. 過去の接種履歴一覧の右端に「複製」ボタンが表示される
3. 「複製」ボタンをクリック
4. **結果**: 何も起きない（コンソールエラーもなし）

## 期待する動作

- 「複製」ボタンをクリックすると、その行の予防接種データを元にした新規登録フォームが開く
  （または確認ダイアログを経て複製が実行される）

## 現状コード

### `frontend/src/features/medical-records/components/VaccinationHistory.tsx:134-142`
```tsx
<div className={`w-[70px] px-2 flex justify-center border-l ${C.borderMedium}`}>
  <Button
    variant="outline"
    size="sm"
    className={`h-10 w-[50px] text-sm ${C.bgPrimary} text-white ${C.hoverBgPrimaryDark} hover:text-white border-transparent px-0`}
  >
    複製
  </Button>
</div>
```

**`onClick` なし** — ボタンがクリック不能なデッドコード状態。

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/features/medical-records/components/VaccinationHistory.tsx:134-142` | 複製ボタン | 要修正 |
| 予防接種 API (`create-vaccination.ts`) | 複製時の POST エンドポイント | 要確認 |

## 修正方針

### 1. props に `onDuplicate` コールバックを追加

`VaccinationHistory` コンポーネントの props に複製ハンドラを追加し、
呼び出し元（`MedicalRecordVaccinations.tsx` 等）から実装を注入する。

```tsx
// VaccinationHistory.tsx の props に追加
interface VaccinationHistoryProps {
  // ... 既存 props
  onDuplicate?: (item: VaccinationHistoryItem) => void;
  canCreate?: boolean;
}

// ボタンに onClick + canCreate ガードを追加
{canCreate ? (
  <Button
    variant="outline"
    size="sm"
    className={`h-10 w-[50px] text-sm ${C.bgAccent} text-white ${C.bgAccentHover} hover:text-white border-transparent px-0`}
    onClick={() => onDuplicate?.(item)}
  >
    複製
  </Button>
) : null}
```

### 2. 呼び出し元で複製ロジックを実装

```tsx
// 呼び出し元
const handleDuplicate = useCallback((item: VaccinationHistoryItem) => {
  // 複製データを prefill して新規作成フォームを開く
  openCreateForm({ prefilledData: item });
}, [openCreateForm]);

<VaccinationHistory
  onDuplicate={handleDuplicate}
  canCreate={canCreate}
/>
```

**注意**: ボタン色も `C.bgPrimary`（黒）→ `C.bgAccent`（青）に修正すること。

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — RBAC
> `canCreate` がない場合は `複製` ボタンを非表示にすること。

### プロジェクト内参照実装
- `frontend/src/features/vaccinations/routes/VaccinationList.tsx` — CRUD ボタンの参照実装

## 優先度
**High** — ユーザーに表示されるボタンが機能しない。UX 上の重大な問題。

## 関連ファイル
- `frontend/src/features/medical-records/components/VaccinationHistory.tsx:134-142` — 要修正
- `frontend/src/features/medical-records/components/MedicalRecordVaccinations.tsx` — 呼び出し元（要確認）
