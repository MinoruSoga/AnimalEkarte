# BUG-075: ケアプランダイアログで type 変更後も旧フィールドの値が残留・送信される

## 種類
バグ（フロントエンド — フォームステート管理ミス）

## 重要度
中

## 発見日
2026-03-29

## 再現手順

1. ケアプランダイアログ（`CarePlanDialog.tsx`）を開く
2. type を選択し、`unitPrice`・`masterId`・`category` を入力する
3. type を別の値に変更する
4. 「更新」ボタンをクリックして保存する

## 期待動作

- type 変更時に、新しい type と無関係なフィールド（`unitPrice`/`masterId`/`category`）がクリアされる
- 「マスタ連動中」バッジが type 変更に応じて消える
- 保存時は現在の type に関連するフィールドのみが送信される

## 実際の動作

- `CarePlanDialog.tsx` L128: type を変更しても `unitPrice`/`masterId`/`category` が `formData` に残留
- 「マスタ連動中」バッジが消えない
- L63-71: `onUpdate(id, formData)` 送信時に、type 変更後も旧条件の `unitPrice`/`masterId` が含まれる
- 旧 type のデータが意図せず保存される

## 影響範囲

- ケアプランダイアログの type 変更フロー全体
- 不正なフィールド組み合わせのデータが DB に保存される可能性

## 修正方針

`CarePlanDialog.tsx` の type 変更ハンドラで、type に依存するフィールドを null/undefined にリセットする処理を追加する。
また、送信前に type に関連するフィールドのみを payload に含める絞り込みロジックを追加する。

## 優先度
中（type 変更後に不正なデータが保存される）

## 派生イシュー

| イシュー | 領域 | 内容 |
|---------|------|------|
| BUG-075（FE） | Frontend | CarePlanDialog type 変更時に関連フィールドをリセット・送信 payload を現 type に限定 |
