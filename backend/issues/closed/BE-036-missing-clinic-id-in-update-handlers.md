# BE-036: Update系ハンドラで extractClinicID が抜けている（14ハンドラ）

## 重大度
**High** — 対象ハンドラの全 PATCH/UPDATE 操作が 404 になる（clinic_id=0 でレコードが見つからない）

## 症状

```
PATCH /api/v1/masters/consultations/:id      → 404 {"error":"resource not found"}
PATCH /api/v1/masters/examination-types/:id  → 404 {"error":"resource not found"}
PATCH /api/v1/masters/procedures/:id         → 404 {"error":"resource not found"}
（以下、全14ハンドラ同様）
```

## 根本原因

以下の Update ハンドラが `extractClinicID(c)` を呼び出していない。

結果として モデルの `ClinicID` がゼロ値（0）のままリポジトリに渡される。

リポジトリの Update は `WHERE id = ? AND clinic_id = ?` でフィルタするため、
clinic_id=0 では一致するレコードが存在せず `RowsAffected == 0` → `ErrNotFound` → 404 を返す。

## 影響ハンドラ一覧（14件）

| ハンドラ | ファイル | 行 |
|---------|---------|-----|
| UpdateCage | cage_handler.go | 70 |
| UpdateCarePlanItem | care_plan_item_handler.go | 75 |
| UpdateCheckup | checkup_handler.go | 59 |
| UpdateCheckupType | checkup_type_handler.go | 73 |
| UpdateClinic | clinic_handler.go | 78 |
| UpdateClinicalPlan | clinical_plan_handler.go | 31 |
| UpdateCompany | company_handler.go | 22 |
| UpdateConsultation | consultation_handler.go | 73 |
| UpdateExaminationType | exam_type_handler.go | 71 |
| UpdateProcedure | procedure_handler.go | 75 |
| UpdateTreatmentPlan | treatment_plan_handler.go | 109 |
| UpdateUser | user_account_handler.go | 101 |
| UpdateVaccine | vaccine_handler.go | 80 |
| UpdateVital | vital_handler.go | 70 |

## 正常ハンドラとの比較

```go
// ✅ 正常例: UpdateExamination (examination_handler.go:106)
func (h *Handler) UpdateExamination(c *gin.Context) {
    clinicID, ok := extractClinicID(c)  // ← あり
    if !ok {
        return
    }
    // ...
    exam := &model.Examination{
        ID: id,
        // clinicID は examination_handler ではリポジトリ内部で使用するため不要
    }
}

// ❌ バグ例: UpdateConsultation (consultation_handler.go:73)
func (h *Handler) UpdateConsultation(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    // extractClinicID(c) を呼ばない
    consultation := &model.Consultation{
        ID:       id,
        ClinicID: 0,  // ← ゼロ値のままリポジトリへ渡る
    }
}
```

## 修正方針

各ハンドラに以下を追加:

```go
func (h *Handler) UpdateXxx(c *gin.Context) {
    clinicID, ok := extractClinicID(c)  // ← 追加
    if !ok {
        return
    }
    id, err := strconv.ParseUint(c.Param("id"), 10, 64)
    // ...
    xxx := &model.Xxx{
        ID:       id,
        ClinicID: clinicID,  // ← 追加（モデルに ClinicID フィールドがある場合）
    }
}
```

## 注意: clinic_id を直接モデルに設定しないケース

`UpdateExamination` など、リポジトリが `clinic_id` を context から取得するパターンでは
モデルに `ClinicID` を設定しない実装もある。各リポジトリの WHERE 条件を確認してから修正すること。

## 再現手順

1. マスタ設定 → 治療プランマスタ → 検査タブで項目を新規登録
2. 登録した項目をクリック → 編集フォームを開く
3. 名称を変更して保存 → 「更新に失敗しました」（404）

## 発見経緯

マスタ設定ページ全ページ 登録・編集テスト中に ChromeMCP で発見（2026-03-16）。
`grep` による全ハンドラスキャンで14ハンドラに同一バグが存在することを確認。
