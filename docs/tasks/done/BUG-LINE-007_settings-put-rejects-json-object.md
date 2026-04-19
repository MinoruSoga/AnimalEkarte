# BUG-LINE-007: LINE予約設定の保存が常に 400 エラー（`[]byte` vs `json.RawMessage`）

## 概要

`PUT /api/v1/clinics/:clinicId/line-reservation-settings` が **JSON オブジェクトフィールド全てで 400 エラー** を返す。原因は Backend のリクエスト型が `[]byte` を使っており、Go の標準 JSON デコーダが `[]byte` を base64 エンコード文字列として解釈するため。

## 影響（CRITICAL）

管理画面 UI の以下のページで **全ての保存が失敗する**:

- 🚫 LINE予約管理 > 基本設定 (`/line-reservation/settings`)
- 🚫 LINE予約管理 > ページ編集 (`/line-reservation/page-editor`)

稼働状態・営業時間・休憩時間・定休曜日・LIFF ID など全ての設定が変更不能。

## 再現手順

1. 管理画面にログイン
2. 「LINE予約管理 > ページ編集」に移動
3. 任意のテキスト変更
4. 「変更を保存」をクリック
5. トースト「business_hours: 正しい形式で入力してください」 / Network: 400

## 該当コード

**Backend** (`backend/internal/handler/line_reservation_setting_request.go:35`):
```go
// jsonRawOrEmpty は JSON フィールドを []byte として保持するためのエイリアス
type jsonRawOrEmpty = []byte
```

**リクエスト型**（同ファイル L12-L27）:
```go
BusinessHours           jsonRawOrEmpty `json:"business_hours"`
BusinessHoursByWeekday  jsonRawOrEmpty `json:"business_hours_by_weekday"`
BreakHours              jsonRawOrEmpty `json:"break_hours"`
ClosedWeekdays          jsonRawOrEmpty `json:"closed_weekdays"`
ClosedDates             jsonRawOrEmpty `json:"closed_dates"`
AdditionalFields        jsonRawOrEmpty `json:"additional_fields"`
```

## 根本原因

Go 標準の `encoding/json` では、`[]byte` フィールドは **base64 エンコード文字列** として扱われる（RFC 7159 の慣習）。Frontend が JSON オブジェクトを送ると:

```json
"business_hours": {"start": "0900", "end": "1900"}
```

Go は `{...}` を base64 文字列として解釈しようとし、`json.UnmarshalTypeError` で失敗 → `response.go:120` が「正しい形式で入力してください」を生成。

Response 型 (`line_reservation_setting_response.go`) では `json.RawMessage` を使っているので、**GET は正常動作**するが PUT だけ失敗する。

## 修正案

```go
import "encoding/json"

// Before
type jsonRawOrEmpty = []byte

// After
type jsonRawOrEmpty = json.RawMessage
```

`json.RawMessage` は `[]byte` と同じ実体だが、`MarshalJSON`/`UnmarshalJSON` が実装されているため任意の JSON を保持できる。

サービス層は `[]byte` を期待しているが、`json.RawMessage` は `[]byte` の型エイリアスなので互換性あり（`service.UpsertLineReservationSettingInput.BusinessHours []byte` にそのまま渡せる）。

## テスト観点

- 既存テストがこのバグを検出できていない（テストは多分モックで service 層を直接叩いている）
- ハンドラ統合テストで実際の JSON リクエストを PUT する必要がある
- response 型との対称性を確認する（GET で返すものを PUT に戻せるか）

## 優先度

**CRITICAL** — 管理画面 UI の主要な保存機能が全滅。ステージング運用の前提が崩れる。

## 確認環境

- staging: `https://api.stg.noah-karte.com/api/v1/clinics/3/line-reservation-settings`
- リクエスト: `PUT` + GET レスポンスをそのまま PUT に送信 → 400
- テスト実施日: 2026-04-14
