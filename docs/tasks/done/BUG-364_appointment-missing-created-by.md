# BUG-364: appointments テーブルに予約作成者（created_by）が記録されない

## 優先度: MEDIUM

## 概要

`appointments` テーブルに予約を作成したスタッフIDが記録されない。
手動予約時の入力スタッフが追跡不能。

## 現状

- `source` カラム（manual / line）で予約元は区別可能
- `doctor_id` は担当医であり、予約作成者ではない
- LINE 予約の場合 created_by は NULL（顧客が作成）
- 手動予約の場合 created_by にスタッフIDをセットすべき

## 修正内容

### DB スキーマ
```sql
created_by bigint REFERENCES staffs(id)  -- nullable（LINE予約は NULL）
```
+ FK インデックス追加

### 全レイヤー
- Model: `CreatedBy *uint64` + `CreatedByStaff *Staff` リレーション
- Service: Create に staffID 追加
- Handler: extractStaffID → Service に渡す（管理API側のみ。LIFF APIは NULL）
- Response DTO / Frontend / api.yaml 更新
