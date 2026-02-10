# AnimalEkarte Database Schema Reference

## テーブル一覧 (22 tables)

### コアテーブル
| # | テーブル名 | 説明 | 主要カラム |
|---|-----------|------|-----------|
| 1 | `owners` | 飼い主情報 | name, phone, email, address |
| 2 | `patients` | 患者（動物）情報 | name, species, breed, birth_date, owner_id |
| 3 | `medical_records` | 診療記録（SOAPS） | patient_id, veterinarian_id, visit_date |
| 4 | `soaps_entries` | SOAPS各項目 | record_id, type(S/O/A/P/S), content |

### 薬剤・処方
| # | テーブル名 | 説明 | 主要カラム |
|---|-----------|------|-----------|
| 5 | `medications` | 薬剤マスタ | name, unit, category |
| 6 | `prescriptions` | 処方情報 | record_id, medication_id, dosage, frequency |
| 7 | `medication_stocks` | 在庫管理 | medication_id, quantity, lot_number |

### 検査
| # | テーブル名 | 説明 | 主要カラム |
|---|-----------|------|-----------|
| 8 | `exam_types` | 検査種別マスタ | name, category, unit |
| 9 | `exam_results` | 検査結果 | record_id, exam_type_id, value, reference_range |
| 10 | `lab_orders` | 検査オーダー | record_id, exam_type_id, status |

### 予約・スケジュール
| # | テーブル名 | 説明 | 主要カラム |
|---|-----------|------|-----------|
| 11 | `appointments` | 予約情報 | patient_id, veterinarian_id, scheduled_at, status |
| 12 | `schedule_slots` | 診療枠 | veterinarian_id, date, start_time, end_time |

### ユーザー・権限
| # | テーブル名 | 説明 | 主要カラム |
|---|-----------|------|-----------|
| 13 | `users` | ユーザー | email, password_hash, role |
| 14 | `veterinarians` | 獣医師情報 | user_id, license_number, specialties |
| 15 | `roles` | ロールマスタ | name (veterinarian/nurse/receptionist) |
| 16 | `user_roles` | ユーザーロール | user_id, role_id |

### 入院・手術
| # | テーブル名 | 説明 | 主要カラム |
|---|-----------|------|-----------|
| 17 | `hospitalizations` | 入院記録 | patient_id, admitted_at, discharged_at, cage_number |
| 18 | `surgeries` | 手術記録 | record_id, surgery_type, duration, notes |

### 会計
| # | テーブル名 | 説明 | 主要カラム |
|---|-----------|------|-----------|
| 19 | `invoices` | 請求書 | record_id, owner_id, total_amount, status |
| 20 | `invoice_items` | 請求明細 | invoice_id, description, amount, quantity |
| 21 | `payments` | 支払記録 | invoice_id, amount, method, paid_at |

### システム
| # | テーブル名 | 説明 | 主要カラム |
|---|-----------|------|-----------|
| 22 | `audit_logs` | 監査ログ | user_id, action, table_name, record_id, changes |

## リレーション図 (主要)

```
owners 1──N patients 1──N medical_records 1──N soaps_entries
                                  │
                          ┌───────┼───────┐
                          │       │       │
                     prescriptions  exam_results  surgeries
                          │
                     medications
```

## GORMモデル確認コマンド

```bash
# モデルファイル一覧
find backend/internal/models -name "*.go" -type f

# 特定モデルの構造確認
grep -A 20 "type Patient struct" backend/internal/models/patient.go

# マイグレーションファイル
find backend -path "*/migrations/*.go" -type f

# AutoMigrate呼び出し確認
grep -rn "AutoMigrate" backend/
```

## PostgreSQL接続情報

```
Host: localhost
Port: 5432
Database: ekarte_db
User: ekarte_user
Password: ekarte_password
```

MCP PostgreSQLサーバー経由でクエリ実行可能。
