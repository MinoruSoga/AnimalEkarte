# scripts/old-db/sql/ — companion SQL

`scripts/` 直下の Python ランナー（`apply-*.py` / `link-*.py`）が実行する SQL。
単体で実行するものではなく、ランナー経由で TEMP テーブルへのマップロード＋適用を行う。

## ファイル

| ファイル | 呼び出し元 | 用途 |
|------|------|------|
| `apply-old-db-staff-activity-map.sql` | `apply-old-db-staff-activity-map.py` / `staff-activity-map-apply.sh` | old_db `staff-activity-map-manifest-v1` の TEMP `tmp_staff_activity_map` 経由で移行スタッフを有効化 |
| `link-old-db-staff-identity-map.sql` | `link-old-db-staff-identity-map.py` / `staff-identity-map-link.sh` | old_db `staff-identity-map-v1` の TEMP `tmp_staff_identity_map` 経由で cross-clinic 同一人物紐付け |
| `link-old-db-cross-clinic-staff-accounts.sql` | `link-old-db-cross-clinic-staff-accounts.py` | 同名スタッフを accounts 共有 + `staff_clinic_assignments` で紐付け（`staffs.id` は書き換えない） |

## 規約

- TEMP テーブル方式: ランナーがマニフェスト/CSV を TEMP にロード → SQL が参照して適用。マニフェストの `schemaVersion` と SQL は対で管理される
- PHI を含む old_db データを扱う — PHI quarantine ルール（`stage-old-db-handoff.sh` / `check-old-db-handoff.sh`）に従う
- 変更時は対応する Python ランナー側の schemaVersion / 列名と整合を確認する
