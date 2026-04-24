# BE-MEDI-010: Treatment に admin_route (投与方法) カラムを追加

## 概要
薬品 (item_type=medicine) の治療明細に投与方法フィールドを永続化するため、
`treatments` テーブルに `admin_route` カラムを追加する。

## 背景
フロントエンドで薬品追加時に投与方法（経口・注射・外用など）を表示する UI が必要。
現状は `memo` フィールドで代替しているが、専用カラムに昇格させる必要がある。

## 実装内容

### migration
```sql
ALTER TABLE treatments ADD COLUMN admin_route VARCHAR(50) DEFAULT '';
```

### model (`backend/internal/model/medical_record.go`)
```go
AdminRoute string `gorm:"column:admin_route;type:varchar(50);default:''" json:"admin_route"`
```

### API
- `POST /v1/medical-records/:id/treatments` の request body に `admin_route` を追加
- `PATCH /v1/medical-records/:id/treatments/:tid` で更新可能に

## 優先度
Low

## 関連
- フロントエンドタスク: `docs/tasks/open/ux/BUG-MEDI-010_treatment_admin_route_missing.md`
- テスト確認日: 2026-03-31
