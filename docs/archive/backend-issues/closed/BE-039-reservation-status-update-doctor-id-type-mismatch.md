# BE-039: 予約ステータス更新 - doctor_id 型の不一致

## 概要

予約管理ページの `ReservationDetailModal` からステータスを更新しようとすると、PATCH リクエストが HTTP 400 で失敗する問題。

## 問題の詳細

フロントエンドから送信される PATCH `/api/v1/reservations/{id}` リクエストで、`doctor_id` が **文字列（医師名）** として送信されているが、バックエンドは **数値（staff_id）** を期待しているため型エラーが発生。

### エラーメッセージ
```
json: cannot unmarshal string into Go struct field updateReservationRequest.doctor_id of type uint64
```

### 実際のリクエストボディ（不正）
```json
{
  "start_time": "2026-03-15T02:00:00.000Z",
  "end_time": "2026-03-15T02:30:00.000Z",
  "visit_type": "first",
  "service_type": "ワクチン接種",
  "doctor_id": "山田 太郎",
  "is_designated": false,
  "status": "in_consultation",
  "notes": "初回ワクチン接種"
}
```

### 期待値
```json
{
  "start_time": "2026-03-15T02:00:00.000Z",
  "end_time": "2026-03-15T02:30:00.000Z",
  "visit_type": "first",
  "service_type": "ワクチン接種",
  "doctor_id": 1,
  "is_designated": false,
  "status": "in_consultation",
  "notes": "初回ワクチン接種"
}
```

## 影響範囲

- 予約管理ページ内での予約ステータス変更機能が使用不可
- その他の予約 CRUD（作成・削除・編集）は正常に動作

## 根本原因

### フロントエンド側
`frontend/src/features/reservations/components/ReservationDetailModal.tsx` で、予約情報を取得する際に医師名（`doctor` フィールド）を使用しており、API 更新時に同じ値をそのまま送信している。

### バックエンド側
`backend/internal/handler/reservation_handler.go` の updateReservationRequest では `doctor_id` を `uint64` として定義しており、これは正しい型定義。

## 修正方法

### オプション A: フロントエンド修正（推奨）

予約情報をロードする際、医師IDを一緒に保持して更新リクエストで使用する：

```typescript
// ReservationDetailModal.tsx
const [appointment, setAppointment] = useState({
  id: string,
  doctor_id: number,  // ← 追加
  doctor: string,
  // ...
});

// ステータス変更時
const handleStatusChange = async (newStatus: string) => {
  await updateReservation({
    doctor_id: appointment.doctor_id,  // ← 医師ID使用
    status: newStatus,
    // ...
  });
};
```

### オプション B: バックエンド修正（次善案）

`doctor_id` フィールドを文字列で受け取り、内部で変換する：

```go
type UpdateReservationRequest struct {
    DoctorID string `json:"doctor_id" binding:"required"`  // 文字列で受け取る
    Status   string `json:"status"`
    // ...
}

// ハンドラー内で変換
docID, err := strconv.ParseUint(req.DoctorID, 10, 64)
if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "invalid doctor_id"})
    return
}
```

## テスト方法

1. 予約管理ページを開く
2. ステータス変更用ドロップダウンをクリック
3. 別のステータスを選択
4. ブラウザ DevTools > Network タブで PATCH リクエストを確認
5. リクエストボディの `doctor_id` が数値であることを確認
6. レスポンスが 200 OK であることを確認

## 優先度

- **中** — ステータス更新は重要だがブロッカー機能ではない（他の CRUD は動作）

## テスト証拠

ブラウザテスト実施日: 2026-03-16
- リクエスト: `PATCH http://localhost:8080/api/v1/reservations/9`
- ステータスコード: 400
- エラーボディ: `{"error":"json: cannot unmarshal string into Go struct field updateReservationRequest.doctor_id of type uint64"}`
