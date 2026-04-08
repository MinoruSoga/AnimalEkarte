# BUG-157: 薬剤マスタのインライン「新しい薬剤を追加...」ボタンが create=F で表示される

## 概要
`/settings/medicine` でヘッダーの「新規登録」ボタンは BUG-156 修正で非表示になったが、
テーブル下部のインラインの「新しい薬剤を追加...」ボタンが create=F でも表示されている。
API は 403 でブロック済み。

## 再現手順
1. 全リソース view-only の権限グループでログイン
2. `/settings/medicine` にアクセス
3. テーブル最下部にスクロール
4. **結果**: 「新しい薬剤を追加...」ボタンが表示される

## 修正方針
`MedicineSettings` のインラインフォームボタンを `canCreate` で条件表示。

## 優先度
**Low** — API で 403。UI のみ。

## 関連ファイル
- `frontend/src/features/master/routes/MedicineSettings.tsx`
