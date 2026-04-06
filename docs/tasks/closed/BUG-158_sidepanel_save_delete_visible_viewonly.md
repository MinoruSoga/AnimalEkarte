# BUG-158: view-only でもサイドパネルに「保存」「削除」ボタンが表示され編集可能な状態になる

## 概要
全リソース view-only（create=F, edit=F, delete=F）の権限でマスタ一覧のレコードをクリックすると、
サイドパネルが**編集モード**で開き、「保存」「キャンセル」ボタンと削除アイコン（🗑️）が表示される。
名前フィールドも入力可能な状態。

API は 403 でブロックするため実データは変更されないが、ユーザーが編集→保存→エラーという不快な UX になる。

## 脆弱性分類
- **UX 問題**（セキュリティ実害なし — API で 403）
- **影響**: ユーザーが編集できると思って操作し、保存時にエラーになる

## 再現手順
1. 全リソース view-only の権限グループでログイン
2. `/settings/animal-species` にアクセス
3. 「犬」行をクリック
4. **結果**: サイドパネルが「編集」モードで開く。「保存」ボタン（青）、削除アイコン（🗑️）、編集可能なフォームフィールドが表示される

## スクリーンショットで確認済み
- ヘッダー: 「編集」テキスト
- フォーム: 名前フィールドが入力可能
- フッター: 「キャンセル」「保存」ボタン
- 右上: 削除アイコン（🗑️）

## 影響範囲

### マスタページ（サイドパネル）

| ページ | サイドパネルに「保存」表示 | 判定 |
|--------|------------------------|------|
| 動物種類 | ✅ 表示 | ❌ |
| スタッフ | ✅ 表示 | ❌ |
| 保険 | ✅ 表示 | ❌ |
| 職種 | ✅ 表示 | ❌ |
| 診療サービス | ✅ 表示 | ❌ |
| 権限グループ | 非表示（パネル開かない） | ✅ |

### 診療系ページ（詳細フォーム）

| ページ | フォーム状態 | 判定 |
|--------|------------|------|
| 飼主詳細 `/owners/:id` | 全17フィールド中16個が編集可能。ヘッダー「飼主・ペット 編集」 | ❌ |

## 期待する動作

### edit=F の場合
- サイドパネルは**閲覧モード**で開く
- フォームフィールドは読み取り専用（disabled or readonly）
- 「保存」ボタン非表示
- 「キャンセル」→「閉じる」に変更

### delete=F の場合
- 削除アイコン（🗑️）非表示

### edit=T, delete=F の場合
- 「保存」ボタン表示、削除アイコン非表示

## 修正方針

`MasterSidePanel` または `MasterCRUDPage` で `canEdit` / `canDelete` を受け取り、
サイドパネルの表示を制御:

```typescript
// MasterSidePanel.tsx
const { canEdit, canDelete } = usePermission(resource);

// フォームフィールド
<Input value={name} disabled={!canEdit} />

// フッター
{canEdit ? (
  <>
    <Button variant="outline" onClick={onCancel}>キャンセル</Button>
    <SubmitButton>保存</SubmitButton>
  </>
) : (
  <Button variant="outline" onClick={onClose}>閉じる</Button>
)}

// 削除アイコン
{canDelete ? <DeleteButton onClick={onDelete} /> : null}
```

## 準拠すべきプロジェクト規約・ベストプラクティス

### `.claude/CLAUDE.md` — Conditional Render
> 必ず `? (...) : null`（`&&` 禁止）

### `.claude/rules/security.md`
> "Validate on both client and server"

UI 層でも編集不可状態を正しく表現すべき。

## 優先度
**Medium** — ユーザーが編集操作を試みて保存時にエラーになる。体験が悪い。

## 関連チケット
- BUG-124/132/152/156/157（修正済み）: create ボタンの権限制御

## 関連ファイル
- `frontend/src/components/shared/SidePeek/MasterSidePanel.tsx`
- `frontend/src/features/master/hooks/use-master-crud.ts`
- `frontend/src/features/master/components/MasterCRUDPage.tsx`
