# BUG-356: マイグレーションにデッド ENUM 型とデッドコメントが残存

## 概要

`001_init.sql` に以下のデッドコードが残存している:

### 1. デッド ENUM 型 `account_status`（L17）

`CREATE TYPE account_status AS ENUM ('active', 'inactive', 'locked')` が定義されているが、
どのテーブルのカラム定義でも使用されておらず、Go コードにも対応する型定義がない。
`accounts` テーブルは `is_active BOOLEAN` で管理。

### 2. 旧テーブル名コメント `user_clinic_memberships`（L602）

```sql
-- 27. user_clinic_memberships（ユーザー所属クリニック）
```
テーブルは `staff_clinic_assignments` にリネーム済み。古い番号とテーブル名のコメントが残存。

### 3. `-- Deleted:` コメント（6箇所: L1323, L1330, L1343, L1528, L1550, L1599）

削除済みインデックス/テーブルの記録コメント。設計変更履歴としては有用だが、本番マイグレーションとしては不要。

## 修正内容

```diff
- CREATE TYPE account_status AS ENUM ('active', 'inactive', 'locked');
```

L602 のコメントを修正:
```diff
-- ------------------------------------
-- 27. user_clinic_memberships（ユーザー所属クリニック）
+-- 6a. staff_clinic_assignments（スタッフ-クリニック中間テーブル）
-- ------------------------------------
- -- 6a. staff_clinic_assignments（スタッフ-クリニック中間テーブル）
```

`-- Deleted:` コメント 6 箇所を削除。

## 優先度

**LOW** — 機能影響なし。マイグレーションファイルの可読性改善。
