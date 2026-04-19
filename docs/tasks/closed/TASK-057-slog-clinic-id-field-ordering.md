# TASK-057: slog clinic_id フィールド順序不統一 — 複数ドメイン

## 優先度

LOW

---

## 概要

参照実装 `medicine_service.go` では slog の**最初のフィールドが常に `clinic_id`**。  
ログ集約基盤（Datadog / CloudWatch 等）でテナント単位の絞り込みを行う際、`clinic_id` が先頭にあることで検索・フィルタが容易になる。

多数のサービスで `clinic_id` が 2 番目以降に配置されており、参照実装との一貫性が失われている。

---

## 参照実装（medicine_service.go）

```go
// ✅ clinic_id が常に先頭
slog.InfoContext(ctx, "medicine created",
    slog.Uint64("clinic_id", clinicID),   // ← 1番目
    slog.Uint64("medicine_id", id))
```

---

## 違反箇所一覧

| サービス | 操作 | 現状（entity_id → clinic_id） |
|---------|------|------------------------------|
| occupation_service | Create, Update, Delete | occupation_id → clinic_id |
| inquiry_template_service | Create, Update, Delete | template_id → clinic_id |
| chief_complaint_type_service | Create, Update, Delete | category_id → clinic_id |
| exam_type_service | Create, Delete | exam_type_id → clinic_id |
| procedure_service | Create, Delete | procedure_id → clinic_id |
| cage_service | Create, Delete | cage_id → clinic_id |
| diagnosis_service (Type) | Create, Update, Delete | type_id → clinic_id |
| diagnosis_service (Name) | Create, Update, Delete | name_id → clinic_id |
| checkup_type_service | Create, Update, Delete | checkup_type_id → clinic_id |
| hospitalization_plan_service | Create, Delete | plan_id → clinic_id |
| trimming_master_service | Course Create, Course Update, Option Create | course_id/option_id → clinic_id |
| insurance_service | Update, Delete | insurance_id → clinic_id |
| vaccine_service | Delete | vaccine_id → clinic_id |
| merchandise_item_service | Update, Delete | merchandise_item_id → clinic_id |
| reservation_type_service | Delete, CreateUnavailableTime, LinkOccupation | entity_id → clinic_id |
| reservation_type_group_service | Update | reservation_type_group_id → clinic_id |

---

## 修正方針

各 slog 呼び出しの引数順序を `clinic_id` 先頭に並べ替える。ロジック変更なし・フィールド入れ替えのみ。

```go
// ❌ 修正前
slog.InfoContext(ctx, "occupation created",
    slog.Uint64("occupation_id", id),
    slog.Uint64("clinic_id", clinicID))

// ✅ 修正後
slog.InfoContext(ctx, "occupation created",
    slog.Uint64("clinic_id", clinicID),
    slog.Uint64("occupation_id", id))
```

---

## 備考

- **Reorder 系**はすでに `clinic_id` が先頭で統一されている（✅）。
- 優先度は LOW。他の機能タスクの合間に一括対応すること。
- 1 コミットで全ドメインをまとめて修正してよい。
