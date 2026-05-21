# Migrations

## 命名規則

```
{連番}_{description}.sql
例: 005_add_trimming_course.sql
```

連番は既存の最大番号 + 1。説明は snake_case。

## 必須チェック

- **clinic_id スコープ**: 新テーブルにクリニック間分離が必要な場合は `clinic_id NOT NULL` を付ける
- **ソフトデリート対象**: 業務データは `deleted_at TIMESTAMPTZ` を追加する
- **CASCADE DELETE 禁止**: 動物患者・医療記録テーブルへの `ON DELETE CASCADE` は絶対禁止
- **インデックス**: `clinic_id` を含む複合インデックスを追加する

## 実行禁止コマンド（自動実行禁止）

```bash
make db            # DB リセット（高い副作用）
docker compose exec db psql ...  # 直接 SQL 実行
```

マイグレーション適用はユーザーが手動で実施する。

## migration 統合後のローカル復旧

`001_init.sql` の checksum mismatch が出たローカル環境は、`docs/infra/deploy/LOCAL_DB_RESET.md` の手順で DB volume を再構築する。
