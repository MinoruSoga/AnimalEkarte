# BUG-161: スタッフ設定のサイドパネルで保存ボタンが削除ダイアログを誤起動する

## 概要
`/settings/staffs` のスタッフ編集サイドパネルで「保存」ボタンをクリックした際に、
保存処理が実行されず、代わりに削除確認ダイアログが表示される（または予期しない動作が起きる）。

## 再現手順
1. `admin@example.com` でログイン
2. `/settings/staffs` にアクセス
3. スタッフ行をクリック → サイドパネルが開く
4. 権限グループを変更する
5. 「保存」ボタンをクリック
6. **結果**: 保存が実行されず、削除ダイアログが表示される（または React エラー）

## 期待する動作
- 「保存」ボタンクリックで `PUT /api/v1/masters/staffs/:id/permission-groups` が呼ばれ保存完了
- 削除ダイアログは Trash アイコンクリック時のみ表示される

## 現状コード（要調査）

### `frontend/src/features/master/routes/StaffSettings.tsx:162-174`
```tsx
const handleSave = useCallback(() => {
  if (!f.name.trim()) { ... }
  onSave(f);        // ← parent の onSave を呼ぶ
  if (!isNew && staffId) {
    onSaveGroups(staffId, groupIds);   // ← permission-groups PUT
    onSaveClinics(staffId, clinicIds); // ← clinics PUT
  }
}, [...]);
```

### `frontend/src/features/master/routes/StaffSettings.tsx:215-219`
```tsx
<MasterSidePanel
  action={handleSave}  // ← form action として渡している
  onDelete={item !== null ? () => onDeleteRequest(item) : undefined}
```

### 疑問点
- `action={handleSave}` と `MasterSidePanel` の `form action` パターンの組み合わせで
  `SubmitButton` の `onClick` と form submit が二重起動する可能性がある
- `SidePeekFooter` の `SubmitButton` に `onClick={onSave}` が設定されており、
  form 内に配置されると submit も同時に発火する

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/features/master/routes/StaffSettings.tsx` | スタッフ設定ページ | 要調査 |
| `frontend/src/components/shared/SidePeek/MasterSidePanel.tsx:111-114` | form action パターン | 要調査 |
| `frontend/src/components/shared/SidePeek/SidePeekFooter.tsx:33-39` | SubmitButton の onClick | 要調査 |

## 修正方針
未調査。以下を確認する:
1. `action` prop と `onSave` prop の両方が設定された場合の二重発火
2. SubmitButton が form 内に配置されたときの onClick + submit の二重呼び出し
3. NavigationBlocker ダイアログが「削除ダイアログ」と誤認された可能性も確認

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — React 19 Action パターン
> **React 19 Action**: 原則 `useActionState` と `<form action={formAction}>` を使用
> **Submit Button**: 送信ボタンは必ず `SubmitButton` を使用

form action と onClick の二重設定は避けるべき。

## 優先度
**Medium** — 保存操作が機能しない場合は UX 上の重大問題。ただし再現性の確認が必要。

## 暫定対処
`vet@example.com`（staff ID 9）の権限グループ変更は DB 直接更新で完了済み:
```sql
UPDATE staff_permission_groups SET group_id = 2 WHERE staff_id = 9 AND group_id = 7;
```

## 関連チケット
- BUG-158: view-only サイドパネル対応（同コンポーネントの変更）

## 関連ファイル
- `frontend/src/features/master/routes/StaffSettings.tsx:162-174` — handleSave
- `frontend/src/components/shared/SidePeek/MasterSidePanel.tsx:111-114` — form action
- `frontend/src/components/shared/SidePeek/SidePeekFooter.tsx:33-39` — SubmitButton
