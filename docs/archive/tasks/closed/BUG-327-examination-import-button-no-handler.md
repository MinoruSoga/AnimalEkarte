# BUG-327: カルテ「検査」タブの「検査取り込み」ボタンが未実装

**Status**: CLOSED  
**Priority**: High  
**Discovery**: 機能テスト Section 4.3 検査タブ (2026-04-12)

## 概要

カルテ編集の「検査」タブにある「検査取り込み」ボタンをクリックしても何も起きない。`ExaminationFilter` コンポーネントのボタンに `onClick` ハンドラが存在しないため、検査記録をカルテに紐付けるダイアログが開かない。

## 再現手順

1. `/medical-records/21` を開き「検査」タブに切り替える
2. 「検査取り込み」ボタンをクリック
3. **結果**: 何も起きない（ダイアログなし・APIリクエストなし）
4. **期待**: 検査記録選択ダイアログが開き、既存の検査結果をカルテに紐付けられる

## 現状コード

### `frontend/src/features/medical-records/components/ExaminationFilter.tsx:37-44`
```tsx
// ❌ onClick ハンドラなし — ボタンを押しても何も起きない
<Button
  size="sm"
  className={`${C.bgAccent} ${C.bgAccentHover} ${C.textWhite} gap-2 h-10 text-sm shadow-sm border-transparent px-4`}
>
  <FileText className={ICON.action} />
  検査取り込み
</Button>
```

### `frontend/src/features/medical-records/components/MedicalRecordExamination.tsx`
```tsx
// ❌ ExaminationFilter に onImport prop を渡していない
<ExaminationFilter
  searchTerm={deferredSearch}
  onSearchChange={setSearchTerm}
  dateStart={dateStart}
  onDateStartChange={setDateStart}
  dateEnd={dateEnd}
  onDateEndChange={setDateEnd}
  // ← onImport prop なし
/>
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/features/medical-records/components/ExaminationFilter.tsx:37-44` | 検査取り込みボタンに onClick なし | ❌ 未実装 |
| `frontend/src/features/medical-records/components/MedicalRecordExamination.tsx` | onImport ハンドラを ExaminationFilter に渡していない | ❌ 未実装 |
| 検査取り込みダイアログ | 未作成 | ❌ 未実装 |
| `POST /v1/medical-records/:id/examinations` 等 | 検査紐付けAPIの呼び出し | ❌ 未呼び出し |

## 修正方針

### 1. `ExaminationFilter.tsx` — onImport prop 追加

```tsx
interface ExaminationFilterProps {
  searchTerm: string;
  onSearchChange: (value: string) => void;
  dateStart: string;
  onDateStartChange: (value: string) => void;
  dateEnd: string;
  onDateEndChange: (value: string) => void;
  onImport?: () => void;  // ← 追加
}

// ボタンに onClick を追加
<Button
  size="sm"
  onClick={onImport}
  className={`...`}
>
  <FileText className={ICON.action} />
  検査取り込み
</Button>
```

### 2. `MedicalRecordExamination.tsx` — ダイアログ state + handler 追加

```tsx
const [isImportDialogOpen, setIsImportDialogOpen] = useState(false);

const handleImport = useCallback(() => {
  setIsImportDialogOpen(true);
}, []);

return (
  <div ...>
    <ExaminationFilter
      ...
      onImport={handleImport}
    />
    {/* 検査取り込みダイアログ */}
    {isImportDialogOpen ? (
      <ExaminationImportDialog
        petId={petId}
        onClose={() => setIsImportDialogOpen(false)}
      />
    ) : null}
    ...
  </div>
);
```

### 3. `ExaminationImportDialog` — 新規コンポーネント作成

既存の検査記録一覧（`GET /v1/examinations?pet_id={petId}`）を表示し、選択した検査をカルテに取り込む機能。
バックエンドAPIの仕様確認が必要。

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — React 19 Patterns
> データ更新（Mutation）は **React 19 Action** パターンを標準とする

ただしネストフォーム内のため `useTransition` + `onClick` パターンを使用すること（BUG-326 の修正方針と同様）。

### `.claude/rules/typescript-react.md` — 条件レンダリング
> 条件レンダーは必ず `? (...) : null`（`&&` 禁止）

### プロジェクト内参照実装
- `frontend/src/components/shared/TreatmentSearchDialog/` — 類似の検索・選択ダイアログの実装参照
- `frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx` — カルテ内でのダイアログ使用パターン参照

## 優先度

**High** — カルテ画面から検査記録の取り込みが一切できない機能不全

## 関連チケット
- BUG-326: 予防接種タブ保存ボタン未実装（同様パターン）
