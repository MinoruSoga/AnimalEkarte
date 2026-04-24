# BUG-LINE-012: customer_fields が無制限にフィールドを受け入れる（入力検証欠落）

> **状態 (2026-04-14 再テスト)**: 部分修正済み・追加対応推奨
> - ✅ 長さ制限 (500 文字) 実装済み — `'a'.repeat(10000)` は 400 で拒否
> - ❌ 未定義フィールド（例: `unknown_field`）が依然として保存される
> - ❌ HTML/JS タグ (`<script>`) がサニタイズされず DB に格納される
>
> 現状 React 自動エスケープで XSS は発火しないが、以下の経路で実害リスクあり:
> - CSV/PDF エクスポート
> - LINE Push メッセージ表示
> - 第三者ツール連携
>
> Follow-up: `additional_fields` 定義に基づくスキーマ検証を追加

## 概要

LIFF 予約作成の `customer_fields` は任意のキー・値を受け入れる。`line_reservation_settings.additional_fields` で定義されたフィールド以外も無制限に保存される。また HTML/JS ペイロードもサニタイズされない。

## 再現

```javascript
await fetch('/api/liff/3/reservations', {
  method: 'POST',
  body: JSON.stringify({
    type_id: 1, staff_id: 33, date: '2026-05-14',
    start_time: '1000', end_time: '1015',
    customer_fields: {
      customer_name: '<script>alert("XSS")</script>',   // サニタイズなし
      phone: '09000000000',
      malicious: '<img src=x onerror=alert(1)>',        // 定義外フィールドが保存される
      huge_field: 'a'.repeat(1000000)                    // サイズ制限なし
    },
    request_text: '<script>alert("req")</script>'
  })
});
// → 201 作成成功、DB にそのまま格納
```

## 影響

### 現状（Admin UI 経由）
- React の自動エスケープで XSS は発火しない（`dangerouslySetInnerHTML` 不使用を確認済）
- ただし以下の経路で漏洩・実行リスク:
  - **将来の CSV/PDF エクスポート** — HTML injection
  - **LINE Push 通知メッセージ** — プレーンテキストなので XSS にはならないが、`<script>` が表示される見苦しさ
  - **第三者ツール・API 連携** — エスケープしない外部ツールでは発火

### データ肥大化
- customer_fields に任意のキー名で MB 級のテキストが保存可能 → DB 肥大化・DoS の余地

## 修正案

1. **スキーマ検証**: `line_reservation_settings.additional_fields` に定義されたキーのみ受け入れる
2. **値の長さ制限**: 各フィールドに適切な max_length（例: 500 文字）
3. **HTML エスケープ**: 保存前に `html.EscapeString()` または明示的に HTML タグを拒否
4. **request_text サイズ制限**: max 1000 文字程度

```go
// Service 層で入力検証
for key, value := range customerFields {
    def, ok := settingDef.AdditionalFields[key]
    if !ok {
        return apperrors.WrapInvalidInput(fmt.Sprintf("未定義のフィールド: %s", key))
    }
    if len(value) > def.MaxLength {
        return apperrors.WrapInvalidInput(fmt.Sprintf("%s は %d 文字以内", key, def.MaxLength))
    }
}
```

## 優先度

**MEDIUM** — 現状の Admin UI では発火しないが、防御深層化の原則に反する。将来的な機能追加（エクスポート等）で実害が発生する。

## 確認環境

- staging: id=4 (曽我　稔 Minoru) の `additional_fields` に生 `<script>` 格納確認
- テスト実施日: 2026-04-14
