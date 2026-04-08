# BUG-159: 権限グループ登録・編集ページでメインコンテンツ側が縦スクロールする

## 概要
権限グループマスタ（`/settings/permission-groups`）で行をクリックしてサイドパネルを開くと、
左側のメインコンテンツ（一覧エリア）が独立して縦スクロール可能になる。
一覧はわずか 2〜3 件なのでスクロールは不要であり、UX として不自然。

## 再現手順
1. `/settings/permission-groups` にアクセス
2. 「執行」行をクリック → サイドパネルが開く
3. 左側の一覧エリアをスクロールしようとすると、**スクロールできてしまう**

## 原因
`main` 要素に `overflow: auto` が設定されており、
サイドパネル（権限テーブル 23 行）の高さが main の scrollHeight を押し上げている。

```
main.scrollHeight = 1912px
main.clientHeight = 991px
main.overflow = auto
→ スクロール可能
```

## 期待する動作
- 左側の一覧は固定表示（スクロール不要）
- サイドパネル（右側）のみがスクロール可能
- または SidePeek レイアウトが左右独立スクロールになるべき

## 修正方針
SidePeek パネルが開いた状態で、左側のコンテンツ部分にスクロールが発生しないように:
- サイドパネルの高さを `overflow-y: auto` で独立スクロールにする
- または main の overflow を `hidden` にして、サイドパネルは absolute/fixed で配置

## 優先度
**Low** — 機能的な問題なし。UX の不自然さ。

## 関連ファイル
- `frontend/src/components/shared/SidePeek/SidePeekPanel.tsx` — レイアウト
- `frontend/src/features/master/routes/PermissionGroupSettings.tsx` — 権限グループページ
- `frontend/src/components/shared/Layout/Layout.tsx` — main 要素の overflow 設定
