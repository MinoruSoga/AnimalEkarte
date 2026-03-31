---
status: closed
---

# マスタ reorder エンドポイント 未実装6件 + api.yaml 未反映10件

## 背景

全マスタに `sort_order` カラムが存在し、フロントエンドの D&D 並び替えを想定している。
しかし以下の問題が2点ある。

1. **バックエンド未実装**: 6マスタに reorder エンドポイントが存在しない
2. **api.yaml 未反映**: 実装済み10マスタが api.yaml に記載されていない

---

## 問題1: バックエンド未実装（6マスタ）

### 未実装一覧

| マスタ | エンドポイント |
|--------|--------------|
| スタッフ | `PATCH /v1/masters/staffs/reorder` |
| 保険 | `PATCH /v1/masters/insurances/reorder` |
| 入院プラン | `PATCH /v1/masters/hospitalization-plans/reorder` |
| トリミングコース | `PATCH /v1/masters/trimming-courses/reorder` |
| トリミングオプション | `PATCH /v1/masters/trimming-options/reorder` |
| 役職 | `PATCH /v1/masters/job-titles/reorder` |

### 実装方針

`exam_type_handler.go` の `ReorderExaminationTypes` と完全に同パターン。

**リクエスト**

```json
{ "ids": [3, 1, 2] }
```

- `ids`: `uint64[]`。配列の順序が新しい `sort_order` になる（index 0 → sort_order 1）

**レスポンス**

```json
{ "message": "reordered" }
```

**各マスタの実装ファイル（参照先）**

| マスタ | handler | service | repository |
|--------|---------|---------|-----------|
| staffs | `staff_handler.go` | `staff_service.go` | `staff_repository.go` |
| insurances | `insurance_handler.go` | `insurance_service.go` | `insurance_repository.go` |
| hospitalization-plans | `hospitalization_plan_handler.go` | `hospitalization_plan_service.go` | `hospitalization_plan_repository.go` |
| trimming-courses | `trimming_course_handler.go` | `trimming_course_service.go` | `trimming_course_repository.go` |
| trimming-options | `trimming_option_handler.go` | `trimming_option_service.go` | `trimming_option_repository.go` |
| job-titles | `job_title_handler.go` | `job_title_service.go` | `job_title_repository.go` |

**注意事項**

- `clinic_id` フィルタ必須（マルチテナント）
- トランザクション内で一括更新
- `ids` に存在しない ID が含まれる場合は 400 Bad Request
- ルーター登録は `staff_handler.go` の `RegisterMasterRoutes()` 内、静的パス `/masters/xxx/reorder` を `/:id` より**前**に登録すること

---

## 問題2: api.yaml 未反映（10マスタ）

以下のエンドポイントはバックエンドに実装済みだが `backend/docs/api.yaml` に記載がない。

| エンドポイント | operationId |
|--------------|------------|
| `PATCH /masters/cages/reorder` | `reorderCages` |
| `PATCH /masters/medicines/reorder` | `reorderMedicines` |
| `PATCH /masters/vaccines/reorder` | `reorderVaccines` |
| `PATCH /masters/service-types/reorder` | `reorderServiceTypes` |
| `PATCH /masters/consultations/reorder` | `reorderConsultations` |
| `PATCH /masters/procedures/reorder` | `reorderProcedures` |
| `PATCH /masters/examination-types/reorder` | `reorderExaminationTypes` |
| `PATCH /masters/checkup-types/reorder` | `reorderCheckupTypes` |
| `PATCH /masters/diagnosis-categories/reorder` | ✅ 記載済み |
| `PATCH /masters/diagnosis-names/reorder` | ✅ 記載済み |

### api.yaml 追記フォーマット（既存の reorder 定義を参照）

```yaml
  /masters/cages/reorder:
    patch:
      operationId: reorderCages
      summary: ケージマスタ一括並び替え
      description: |
        指定した順序で `sort_order` を一括更新する。
        リクエスト内の `ids` の並び順が表示順に対応する（`ids[0]` が `sort_order=1`）。
        存在しない ID または他クリニックの ID を含む場合は 400 を返す。
      tags: [Masters]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/MasterReorderRequest'
            example:
              ids: [3, 1, 2]
      responses:
        '200':
          description: 並び替え成功
          content:
            application/json:
              schema:
                type: object
                properties:
                  message:
                    type: string
                    example: reordered
        '400':
          $ref: '#/components/responses/BadRequest'
        '401':
          $ref: '#/components/responses/Unauthorized'
        '500':
          $ref: '#/components/responses/InternalError'
```

共通スキーマ `MasterReorderRequest` を components/schemas に追加することで重複排除できる。

---

## 完了条件

- [ ] `PATCH /v1/masters/staffs/reorder` 実装
- [ ] `PATCH /v1/masters/insurances/reorder` 実装
- [ ] `PATCH /v1/masters/hospitalization-plans/reorder` 実装
- [ ] `PATCH /v1/masters/trimming-courses/reorder` 実装
- [ ] `PATCH /v1/masters/trimming-options/reorder` 実装
- [ ] `PATCH /v1/masters/job-titles/reorder` 実装
- [ ] api.yaml に未記載8エンドポイントを追記
- [ ] api.yaml に `MasterReorderRequest` 共通スキーマを追加（`DiagnosisCategoriesReorderRequest` / `DiagnosisNamesReorderRequest` も統合）
