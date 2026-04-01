# BUG-114: 治療プランのステータス変更が 400 エラー（日本語 → 英語 enum 不一致）

## 概要

カルテ編集 Tab2「診察/治療プラン」の治療プラン行のステータスコンボボックスで
「完了」を選択すると PATCH リクエストが 400 Bad Request になる。

フロントエンドが日本語ラベル（`"完了"`, `"未完了"`, `"-"`）をそのまま API に送信しているが、
バックエンドは英語 enum 値（`"pending"`, `"completed"`, `"not_applicable"`）のみを受け付ける。

## 症状

- カルテ編集 → Tab2「診察/治療プラン」→ 治療プラン行のステータスコンボボックスで「完了」を選択
- PATCH /api/v1/medical-records/1/treatments/2 → **400 Bad Request**
- リクエストボディ: `{"status":"完了"}`
- レスポンス: `{"error":"invalid status: 完了"}`
- トースト: 「invalid status: 完了」

## 期待する動作

- 「完了」選択 → `{"status":"completed"}` を送信 → 200/204 成功
- 「未完了」選択 → `{"status":"pending"}` を送信
- 「-」選択 → `{"status":"not_applicable"}` を送信

## 根本原因

`frontend/src/components/shared/TreatmentSearchDialog/` または
治療プランのステータス選択コンポーネントで、日本語ラベル → 英語 enum のマッピングが実装されていない。

バックエンドの有効値（`model/treatment.go`）:
```go
TreatmentStatusPending       TreatmentStatus = "pending"
TreatmentStatusCompleted     TreatmentStatus = "completed"
TreatmentStatusNotApplicable TreatmentStatus = "not_applicable"
```

フロントエンドが送信している値:
- `"完了"` → 正しくは `"completed"`
- `"未完了"` → 正しくは `"pending"`
- `"-"` → 正しくは `"not_applicable"`

## 修正方針

フロントエンドのステータスコンボボックスで日本語ラベル → 英語 enum 値のマッピングを追加する。

```typescript
// ステータスの選択肢
const STATUS_OPTIONS = [
  { label: "未完了", value: "pending" },
  { label: "完了", value: "completed" },
  { label: "-", value: "not_applicable" },
];
```

## 影響ファイル

- `frontend/src/features/medical-records/` 内の治療プランコンポーネント（ステータス選択箇所）

## 優先度

High（治療プランの実施状態を更新不能 → 診療フロー阻害）

## 関連

- Tab2「診察/治療プラン」テスト（Section 4.10）
- テスト確認日: 2026-04-01（ローカル環境）
- PATCH http://localhost:8080/api/v1/medical-records/1/treatments/2 [400] 確認
