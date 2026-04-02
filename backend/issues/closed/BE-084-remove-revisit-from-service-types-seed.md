# BE-084: service_types マスタから「再診」を削除し、関連予約データを修正

## 背景

予約フォームにおいて「診療サービス」(旧: 予約区分) は `service_types` テーブルを参照する。
しかし現行シードデータに `name = '再診'` のサービス種別 (id=8) が存在しており、
`visit_type`（初診/再診）と概念が重複し混乱を招く。

FE 側は「診療サービス」フィールドと「予約区分」(初診/再診) を明確に分離した。
BE 側では seed データの整合性を取る必要がある。

## 変更内容

### 1. `backend/migrations/002_seed_master.sql`

`service_types` へのINSERT から `再診`（id=8 相当の行）を削除する。

```sql
-- 削除対象
INSERT INTO service_types (...) VALUES (..., '再診', ...);
```

### 2. 同ファイル内の予約シードデータ

`service_type_id = 8` (再診) を参照している予約レコードを、適切な service_type_id に変更する。
（例: `一般診療` (id=1) など、実態に合うものを選ぶ）

```sql
-- 変更対象例（再診 service_type を持つ reservation）
INSERT INTO reservations (..., service_type_id, ...) VALUES (..., 8, ...);
-- → service_type_id = 1 (一般診療) などに変更
```

### 3. DB リセット & 再起動

```bash
make down
docker compose up -d --build
make migrate
```

## 確認方法

```sql
SELECT id, name FROM service_types ORDER BY sort_order;
-- 「再診」が存在しないこと

SELECT id, service_type_id FROM reservations WHERE service_type_id = 8;
-- 0件であること
```

## 影響範囲

- `service_types` マスタ (1件削除)
- `reservations` テーブル (service_type_id=8 参照レコードを変更)
- カレンダー表示の色分け（再診の色が使われなくなる）

## 優先度

中（機能上は動作するが、概念的な混乱を解消するため）
