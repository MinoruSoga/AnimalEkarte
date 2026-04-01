# BUG-115: 見積書「行を追加」ボタンがフォームを送信 → Tab1（問診）に強制遷移

## 概要

カルテ編集 見積書タブの「行を追加」ボタンが `type="button"` を持たないため、
`<form action={formAction}>` のデフォルト submit として実行される。
結果、chiefComplaint バリデーションが走り「必須項目が未入力です」トーストが 2 回表示され、
画面が Tab1（問診）に強制遷移する。

## 症状

1. カルテ編集 → 「見積書」タブをクリック
2. 「行を追加」ボタンをクリック
3. 期待: 見積書テーブルに空行が追加される
4. 実際:
   - トースト「必須項目が未入力です」が 2 回表示される
   - ページが Tab1（問診）に強制遷移する
   - 見積書に行は追加されない

## 根本原因

見積書コンポーネント内の「行を追加」ボタンに `type="button"` が指定されていない。
React 19 の `<form action={formAction}>` 内ではデフォルトが `type="submit"` のため、
クリック時にフォームが submit される。

```tsx
// ❌ 現状（フォーム送信が起きる）
<button onClick={handleAddRow}>行を追加</button>

// ✅ 修正後
<button type="button" onClick={handleAddRow}>行を追加</button>
```

同種の問題が BUG-101 でも報告されている（主訴タブの保存フロー）。

## 影響ファイル

- `frontend/src/features/medical-records/` 内の見積書コンポーネント（行追加ボタン箇所）

## 優先度

High（見積書が全く使用不能）

## 関連

- BUG-101: chiefComplaint バリデーション未除去（同根パターン）
- テスト確認日: 2026-04-01（ローカル環境）
- カルテ編集 /medical-records/1 の見積書タブで確認
