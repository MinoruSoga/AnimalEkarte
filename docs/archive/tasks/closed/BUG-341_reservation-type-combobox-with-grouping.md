# BUG-341: 予約区分セレクトボックスをコンボボックス＋グループ階層表示に変更

## 概要
予約フォームの予約区分選択が単純な `<Select>` で実装されており、`reservation_type_groups` テーブルによるグルーピングが UI に反映されていない。コンボボックス（検索可能な選択UI）に変更し、グループを見出しとした階層表示にする必要がある。

## 再現手順
1. `admin@example.com` / `password` でログイン
2. 予約カレンダー → 「予約登録」ボタンをクリック
3. 予約区分フィールドを確認
4. **結果**: フラットな `<Select>` ドロップダウンで全区分が並列表示される。グループ見出しなし、検索不可。

## 期待する動作
- コンボボックス（入力による絞り込み検索が可能）に変更する
- `reservation_type_groups` のグループ名を見出し（`<SelectGroup>` + `<SelectLabel>`）として表示し、各グループ配下に属する予約区分を階層表示する
- グループに属さない予約区分は「その他」グループとしてまとめる

## 現状コード

### `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:66,191-205`
```tsx
// line 66: フラットなリストを取得
const reservationTypes = useMasterItems("reservationType");

// line 191-205: グルーピングなしの Select
<Select value={formData.type} onValueChange={(v) => onChange("type", v)}>
  <SelectTrigger>
    <SelectValue placeholder="予約区分を選択" />
  </SelectTrigger>
  <SelectContent>
    {reservationTypes.map((item) => (
      <SelectItem key={item.id} value={String(item.id)}>
        {item.name}
      </SelectItem>
    ))}
  </SelectContent>
</Select>
```

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// shadcn/ui の SelectGroup + SelectLabel でグルーピング表示
<SelectContent>
  {groups.map((group) => (
    <SelectGroup key={group.id}>
      <SelectLabel>{group.name}</SelectLabel>
      {group.types.map((type) => (
        <SelectItem key={type.id} value={String(type.id)}>
          {type.name}
        </SelectItem>
      ))}
    </SelectGroup>
  ))}
</SelectContent>
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:191-205` | 予約区分 Select → Combobox + GroupedSelect に変更 | 未修正 |
| `frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx` | 型・props 変更なし（formData.type は変わらない） | 影響なし |
| `frontend/src/features/master/api/get-master-items.ts` | グループ付き API エンドポイント追加が必要 | 要調査 |
| `backend/internal/handler/reservation_type_handler.go` | グループ情報を include したレスポンス追加 or 既存 list に group_id/group_name を含める | 要確認 |
| `backend/internal/model/reservation_type.go` | `GroupID *uint64`, `Group *ReservationTypeGroup` フィールド確認 | 実装済み |

## 修正方針

### 1. バックエンド: ReservationType の List レスポンスにグループ情報を含める
`backend/internal/handler/reservation_type_handler.go`
```go
// Preload("Group") を追加し、group_id / group_name をレスポンスに含める
db.WithContext(ctx).Preload("Group").Scopes(clinicScope(clinicID)).Find(&types)
```

### 2. フロントエンド: グループ別に分類してコンボボックス表示
```tsx
// ReservationFormFields.tsx
// 1) データをグループ別にグループ化
const groupedTypes = useMemo(() => {
  const map = new Map<string, { label: string; types: ReservationType[] }>();
  for (const t of reservationTypes) {
    const key = t.group_id ? String(t.group_id) : "__other__";
    const label = t.group?.name ?? "その他";
    if (!map.has(key)) map.set(key, { label, types: [] });
    map.get(key)!.types.push(t);
  }
  return [...map.values()];
}, [reservationTypes]);

// 2) Combobox（Command + Popover）または SelectGroup で階層表示
<SelectContent>
  {groupedTypes.map((g) => (
    <SelectGroup key={g.label}>
      <SelectLabel>{g.label}</SelectLabel>
      {g.types.map((t) => (
        <SelectItem key={t.id} value={String(t.id)}>{t.name}</SelectItem>
      ))}
    </SelectGroup>
  ))}
</SelectContent>
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Design Tokens / Conditional Render
> **Conditional Render**: 必ず `? (...) : null`（`&&` 禁止）

### `.claude/rules/code-style.md` — `js-cache-function-results`
> API 由来の JSX リスト生成は `useMemo([list])` でキャッシュ

グループ分けのロジックは `useMemo` で包む。

### プロジェクト内参照実装
- `frontend/src/features/master/routes/ReservationTypeSettings.tsx` — 予約区分マスタでのグループ表示パターン（参照）

## 優先度
**Medium** — UX 改善。件数が増えると選択が困難になるが、既存予約フローは機能している。

## 関連チケット
- なし

## 関連ファイル
- `frontend/src/components/shared/ReservationFormModal/ReservationFormFields.tsx:66,191-205` — 予約区分 Select
- `backend/internal/model/reservation_type.go` — ReservationType モデル（Group フィールド）
- `backend/internal/model/reservation_type_group.go` — ReservationTypeGroup モデル
- `backend/internal/handler/reservation_type_handler.go` — List エンドポイント
