# UX-005: 権限マスタが実装されていない

## 概要

マスタ設定画面の「スタッフ・権限」セクションに、スタッフの権限を管理するマスタが存在しない。現在は職種（job_title）のみが表示されている。

## 症状

- マスタ設定の「スタッフ・権限」セクションに `staff` と `job_title` のみ存在
- 権限管理用マスタページがない
- スタッフに付与する権限を管理する手段がない

## 現在の状態

### Backend
- `accounts` テーブルに `user_type` 列（system_admin, clinic_admin, staff）
- `staffs` テーブルに `staff_role` enum 型（veterinarian, nurse, trimmer, reception, manager）
- 権限テーブル: **なし**（過去に permission_groups があったが削除済み）

### Frontend
- マスタ設定インデックス：`MASTER_SECTIONS` に権限マスタへの参照がない
- スタッフ編集時：職種のみ編集可能

## 必要な実装

### Option A: スタッフロールマスタテーブル化（推奨）

```sql
CREATE TABLE staff_roles (
  id BIGSERIAL PRIMARY KEY,
  clinic_id BIGINT NOT NULL,
  name VARCHAR(100) NOT NULL,          -- 「獣医師」「看護師」等
  description TEXT DEFAULT '',
  is_active BOOL DEFAULT true,
  sort_order INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  CONSTRAINT fk_staff_roles_clinic FOREIGN KEY (clinic_id) REFERENCES clinics(id),
  CONSTRAINT uk_staff_roles_clinic_name UNIQUE (clinic_id, name) WHERE deleted_at IS NULL
);

-- staffs.staff_role を BIGINT に変更 → staff_role_id
ALTER TABLE staffs ADD COLUMN staff_role_id BIGINT;
ALTER TABLE staffs ADD CONSTRAINT fk_staffs_staff_role FOREIGN KEY (staff_role_id) REFERENCES staff_roles(id);
```

### Option B: 権限グループマスタテーブル

```sql
CREATE TABLE permission_groups (
  id BIGSERIAL PRIMARY KEY,
  clinic_id BIGINT NOT NULL,
  name VARCHAR(100) NOT NULL,          -- 「管理者」「獣医師」等
  description TEXT DEFAULT '',
  permissions JSONB DEFAULT '[]',      -- ["read:medical_records", "write:billing", ...]
  is_active BOOL DEFAULT true,
  sort_order INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL
);

CREATE TABLE staff_permission_groups (
  id BIGSERIAL PRIMARY KEY,
  staff_id BIGINT NOT NULL,
  permission_group_id BIGINT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_staff_perm_staff FOREIGN KEY (staff_id) REFERENCES staffs(id),
  CONSTRAINT fk_staff_perm_group FOREIGN KEY (permission_group_id) REFERENCES permission_groups(id)
);
```

## 優先度

**HIGH** — 権限管理体系が不完全であり、スタッフの権限割り当てができない。

## 関連

- docs/ERD.md: スキーマ更新が必要
- backend/internal/model: StaffRole 型の再検討
- frontend/src/features/master: 権限マスタUIの追加

## 決定待ち

- [ ] Option A（スタッフロールマスタ化）vs Option B（権限グループマスタ）どちらを採用するか？
- [ ] スタッフの権限管理の粒度（機能レベル vs 操作レベル）
