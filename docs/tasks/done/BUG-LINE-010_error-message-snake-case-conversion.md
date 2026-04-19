# BUG-LINE-010: バリデーションエラーメッセージで `type_id` が `type_i_d` に変換される

## 概要

LIFF 予約作成で `type_id` フィールドが欠落した際のエラーメッセージが:

```json
{"error":"type_i_d は必須です"}
```

`type_i_d` は `TypeID` を snake_case 化する際に `_I_D` となってしまっている誤変換。

## 再現

```javascript
await fetch('/api/liff/3/reservations', {
  method: 'POST', headers: authHeaders,
  body: JSON.stringify({ staff_id: 33, date: '2026-04-25', ... })  // type_id omitted
});
// → 400 {"error":"type_i_d は必須です"}
//   期待: {"error":"type_id は必須です"}  または "course_id は必須です"
```

## 原因推定

`parseBindError` (`response.go` 付近) が Go struct の `TypeID` フィールドを snake_case 化する際に、大文字の連続を適切に扱えていない。

修正には `strings` 系の単純な変換ではなく、Go のタグ (`json:"type_id"`) から取得するか、適切な変換関数を使う必要がある。

## 影響

- ユーザーに表示されるエラーメッセージが **存在しないフィールド名** を示す
- Frontend のエラー表示が `type_i_d` フィールドを赤く表示しようとしても該当フィールドがなく、動作が壊れる可能性

## 優先度

**MEDIUM** — 表示品質問題だが、バリデーション時にユーザーを混乱させる
