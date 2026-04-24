# FE-188: スタッフマスタ SidePanel で readOnly が渡されず保存ボタンが常時表示される

## 概要

`StaffSettings.tsx` の `renderSidePanel` callback で `readOnly` を destructure していないため、
`canEdit=false` のユーザーがスタッフ詳細 SidePanel を開いても保存ボタンが表示され、全フォームフィールドが編集可能な状態になる。

## 根本原因

`MasterCRUDPage` は `renderSidePanel({ ..., readOnly: !canEdit })` を正しく渡しているが、
`StaffSettings.tsx:548` の callback 引数で `readOnly` を受け取っていない。

```tsx
// StaffSettings.tsx:548 — readOnly を受け取っていない ❌
renderSidePanel={({ item, onClose, onSave, onDeleteRequest }) => (
  <StaffSidePanel
    key={item?.id ?? "new"}
    item={item}
    onClose={onClose}
    onSave={onSave}
    onDeleteRequest={onDeleteRequest}
    allOccupations={allOccupations}
    allGroups={allGroups}
    onSaveGroups={handleSaveGroups}
    allClinics={allClinics}
    onSaveClinics={handleSaveClinics}
    // readOnly={readOnly} ← 渡していない
  />
)}
```

`StaffSidePanel` は `readOnly?: boolean` props を受け取る実装になっているが（行 73, 91, 226）、
`renderSidePanel` callback の引数から取り出せていないため常に `undefined`（= false 扱い）になる。

## 影響

- `canEdit=false` ユーザーがスタッフ行をクリックすると SidePanel が「編集」モードで開く
- 氏名（textbox）、職種（combobox）、資格番号、パスワード、所属医院 checkbox、権限グループ checkbox がすべて操作可能
- 「保存」ボタンが disabled=false で表示され、クリックで保存 API が呼ばれてしまう可能性

## 修正方法

```tsx
// StaffSettings.tsx
renderSidePanel={({ item, onClose, onSave, onDeleteRequest, readOnly }) => (
  <StaffSidePanel
    key={item?.id ?? "new"}
    item={item}
    onClose={onClose}
    onSave={onSave}
    onDeleteRequest={onDeleteRequest}
    readOnly={readOnly}           // ← 追加
    allOccupations={allOccupations}
    allGroups={allGroups}
    onSaveGroups={handleSaveGroups}
    allClinics={allClinics}
    onSaveClinics={handleSaveClinics}
  />
)}
```

## 優先度

**HIGH** — 閲覧のみユーザーがスタッフのパスワードや権限グループを変更・保存できる状態。

## 関連ファイル

- `frontend/src/features/master/routes/StaffSettings.tsx` — 行 548: renderSidePanel callback
- `frontend/src/features/master/components/MasterCRUDPage.tsx` — 行 127: `readOnly: !canEdit` を渡している
- 発見日: 2026-04-08（RBAC Phase 3 テスト中）
- 関連: FE-161〜FE-176（マスタ設定 readOnly 修正）の修正漏れ
