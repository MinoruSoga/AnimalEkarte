# UX-002: スタッフマスタの職種列が空白表示 [RESOLVED]

## 概要

スタッフマスタページ (`/settings/staff`) の一覧テーブルにおいて、**職種列がすべて空白**で表示されると報告されていた。

**2026-04-05 再検証結果**: ✅ **既に解決済み** — 職種データが正常に表示されている

**確認内容 (2026-04-05):**
- スタッフ10件がすべて職種データ付きで表示 ✅
  - 山田太郎: 獣医師 ✅
  - 田中太郎: 獣医師 ✅
  - 佐藤美咲: 看護師 ✅
  - 鈴木一郎: 受付 ✅
  - 高橋さくら: トリマー ✅
  - 他: 運営管理者 ✅
- API レスポンス: `GET /api/v1/masters/staffs` → 200 ✅
- バックエンド実装: `staff_response.go:52` で `StaffRole` を JSON 出力 ✅
- フロントエンド実装: `staffs.ts:67` で `data.staff_role` を変換 ✅

## 影響範囲

- **ページ**: `/settings/staff` (スタッフマスタ)
- **UI**: テーブルの "職種" カラム
- **ユーザー影響**: 一覧からスタッフの職種を確認できない

## 問題の詳細

### 現象
```
氏名           | 職種 | ステータス | 操作
山田 太郎     |      | 無効      | [編集]
田中 太郎     |      | 無効      | [編集]
```

職種列が空白のままで、職種データが表示されない。

### 根本原因の可能性
1. **バックエンド API**: `GET /v1/masters/staffs` が `staff_role` を返していない（null または undefined）
2. **型変換エラー**: Frontend の `transformStaff()` で `data.staff_role` が undefined に変換されている
3. **DBデータ不足**: staffs テーブルの `staff_role` カラムが NULL である

### コード分析

**Frontend (StaffSettings.tsx:293)**
```typescript
<TableCell className={`text-base ${C.text}`}>{STAFF_ROLE_LABELS[item.staffRole]}</TableCell>
```

**STAFF_ROLE_LABELS定義 (staffs.ts:44-50)**
```typescript
export const STAFF_ROLE_LABELS: Record<StaffRoleValue, string> = {
  veterinarian: "獣医師",
  nurse: "看護師",
  trimmer: "トリマー",
  reception: "受付",
  manager: "運営管理者",
};
```

**Transform関数 (staffs.ts:56-74)**
```typescript
staffRole: data.staff_role as StaffRoleValue,
```

→ `data.staff_role` が undefined または null の場合、表示されない

## 調査ステップ

### 1. API応答確認
- Network DevTools で `/v1/masters/staffs` の応答を確認
- `staff_role` フィールドが返されているか確認

### 2. バックエンド確認
- `backend/internal/handler/staff_handler.go` の GetStaffs メソッドを確認
- JSON マーシャリングで `staff_role` が含まれているか確認
- `backend/internal/repository/staff_repository.go` の `FindAll()` で staff_role が取得されているか確認

### 3. モデル確認
- `backend/internal/model/staff.go` の Staff 構造体で `StaffRole` フィールドが正しく定義されているか確認

## 修正方針

- [ ] API応答で `staff_role` が正しく返されているか確認
- [ ] DBで staffs テーブルの `staff_role` カラムにデータが存在するか確認
- [ ] Frontend の型定義と Backend の JSON マーシャリングのマッピングが一致しているか確認

## 関連ファイル

**Backend:**
- `backend/internal/model/staff.go`
- `backend/internal/handler/staff_handler.go`
- `backend/internal/repository/staff_repository.go`
- `backend/internal/service/staff_service.go`

**Frontend:**
- `frontend/src/features/master/routes/StaffSettings.tsx`
- `frontend/src/features/master/api/staffs.ts`

## テスト環境

- ブラウザ: Chrome
- ローカルURL: `http://localhost:3003/settings/staff`
- テスト日時: 2026-04-05
