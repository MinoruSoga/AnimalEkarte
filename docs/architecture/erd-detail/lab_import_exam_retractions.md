<!-- GENERATED: scripts/tools/generate-erd-detail.py; do not edit -->
# `lab_import_exam_retractions`

[詳細ERD一覧](index.md) · [概要ERD](../erd.md)

このページは、マイグレーションをファイル名順に静的解析した最終定義です。稼働DBの状態は未確認です。

## カラム

型・DEFAULTはDDLの宣言表記です。`bigserial` は採番宣言であり、DBカタログ上の型名ではありません。

| カラム | 型（DDL） | NULL | DEFAULT（DDL） | キー構成列 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `id` | `BIGSERIAL` | NOT NULL | — | PK, UK | [001_init.sql:5448](../../../backend/migrations/001_init.sql#L5448) |
| `clinic_id` | `BIGINT` | NOT NULL | — | FK, UK | [001_init.sql:5449](../../../backend/migrations/001_init.sql#L5449) |
| `job_id` | `UUID` | NOT NULL | — | FK, UK | [001_init.sql:5450](../../../backend/migrations/001_init.sql#L5450) |
| `exam_id` | `BIGINT` | NOT NULL | — | FK, UK | [001_init.sql:5451](../../../backend/migrations/001_init.sql#L5451) |
| `actor_id` | `BIGINT` | NULL | — | — | [001_init.sql:5452](../../../backend/migrations/001_init.sql#L5452) |
| `reason` | `TEXT` | NOT NULL | — | — | [001_init.sql:5453](../../../backend/migrations/001_init.sql#L5453) |
| `parent_snapshot` | `JSONB` | NOT NULL | — | — | [001_init.sql:5454](../../../backend/migrations/001_init.sql#L5454) |
| `created_at` | `TIMESTAMPTZ` | NOT NULL | `now()` | — | [001_init.sql:5455](../../../backend/migrations/001_init.sql#L5455) |

PK/FK/UKはキーの構成列を示します。複合キーの各列が単独で一意という意味ではありません。部分uniqueは下記indexの述語を参照してください。

## ERD

このテーブルの全カラムと、すべての参照先を表示します。参照先はFKの対象列だけを表示し、全カラムは各テーブルのページに置きます。

親側の任意参照はFK列のNULL可否、子側の最大1件は無条件のPK/UNIQUEから導出します。部分uniqueによる条件付きの最大件数は図に反映せず、index一覧に記録します。実線はFK全列が子のPKに含まれる識別関係、破線はそれ以外です。

```mermaid
erDiagram
    lab_import_exam_retractions {
        BIGSERIAL id PK,UK
        BIGINT clinic_id FK,UK
        UUID job_id FK,UK
        BIGINT exam_id FK,UK
        BIGINT actor_id
        TEXT reason
        JSONB parent_snapshot
        TIMESTAMPTZ created_at
    }
    exams {
        BIGSERIAL id PK,FK,UK
        bigint clinic_id FK,UK
        uuid job_id FK,UK
    }
    lab_import_jobs {
        uuid id PK,UK
        bigint clinic_id FK,UK
    }
    lab_import_jobs ||..o{ lab_import_exam_retractions : "clinic_id, job_id"
    exams ||..o{ lab_import_exam_retractions : "clinic_id, exam_id, job_id"
```

## 外部キー（全件）

| 制約 | 子の列（順序あり） | 参照先 | 親の列（順序あり） | 削除時の動作 | 定義元 |
|:---|:---|:---|:---|:---|:---|
| `fk_lab_import_exam_retractions_clinic_job` | `clinic_id, job_id` | [lab_import_jobs](lab_import_jobs.md) | `clinic_id, id` | `ON DELETE RESTRICT` | [001_init.sql:5469](../../../backend/migrations/001_init.sql#L5469) |
| `fk_lab_import_exam_retractions_clinic_exam_job` | `clinic_id, exam_id, job_id` | [exams](exams.md) | `clinic_id, id, job_id` | `ON DELETE RESTRICT` | [001_init.sql:5475](../../../backend/migrations/001_init.sql#L5475) |

正確なON UPDATE・MATCH・DEFERRABLE等を含む制約DDL:

- `fk_lab_import_exam_retractions_clinic_job`: `CONSTRAINT fk_lab_import_exam_retractions_clinic_job     FOREIGN KEY (clinic_id, job_id)     REFERENCES lab_import_jobs (clinic_id, id)     ON DELETE RESTRICT`
- `fk_lab_import_exam_retractions_clinic_exam_job`: `CONSTRAINT fk_lab_import_exam_retractions_clinic_exam_job     FOREIGN KEY (clinic_id, exam_id, job_id)     REFERENCES exams (clinic_id, id, job_id)     ON DELETE RESTRICT`

## その他の制約

| 制約 | 種別 | DDL | 定義元 |
|:---|:---|:---|:---|
| `lab_import_exam_retractions_pkey` | primary | `PRIMARY KEY` | [001_init.sql:5448](../../../backend/migrations/001_init.sql#L5448) |
| `uq_lab_import_exam_retractions_clinic_id_job_exam` | unique | `CONSTRAINT uq_lab_import_exam_retractions_clinic_id_job_exam         UNIQUE (clinic_id, id, job_id, exam_id)` | [001_init.sql:5456](../../../backend/migrations/001_init.sql#L5456) |

## インデックス

| 名前 | UNIQUE | 述語 | DDL | 定義元 |
|:---|:---|:---|:---|:---|
| `idx_lab_import_exam_retractions_clinic_job` | NO | — | `CREATE INDEX idx_lab_import_exam_retractions_clinic_job     ON lab_import_exam_retractions (clinic_id, job_id, exam_id, id)` | [001_init.sql:5461](../../../backend/migrations/001_init.sql#L5461) |
| `idx_lab_import_exam_retractions_clinic_exam_job` | NO | — | `CREATE INDEX idx_lab_import_exam_retractions_clinic_exam_job     ON lab_import_exam_retractions (clinic_id, exam_id, job_id)` | [001_init.sql:5465](../../../backend/migrations/001_init.sql#L5465) |
