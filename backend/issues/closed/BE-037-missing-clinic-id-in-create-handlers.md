# BE-037: Create系ハンドラで extractClinicID が抜けている（10ハンドラ）

## 重大度
**High** — 対象ハンドラの全 POST 操作が 500 になる（clinic_id=0 で FK制約違反）

## 症状

```
POST /api/v1/masters/cages → 500 {"error":"internal server error"}
（以下、全10ハンドラ同様）
```

## 根本原因

以下の Create ハンドラが `extractClinicID(c)` を呼び出していない。

結果として `model.Xxx.ClinicID` がゼロ値（0）のままリポジトリに渡される。

DB の `clinic_id` カラムは `NOT NULL` かつ `clinics` テーブルへの外部キー制約があるため、
`clinic_id=0` はFK違反 → PostgreSQL エラー → 500 Internal Server Error。

## 影響ハンドラ一覧（9件）

| ハンドラ | ファイル |
|---------|---------|
| CreateCage | cage_handler.go |
| CreateCarePlanItem | care_plan_item_handler.go |
| CreateCheckup | checkup_handler.go |
| CreateDailyRecord | daily_record_handler.go |
| CreateExamination | examination_handler.go |
| CreateRecordImage | record_image_handler.go |
| CreateTreatmentPlanForMedicalRecord | treatment_plan_handler.go |
| CreateTreatmentPlanForHospitalization | treatment_plan_handler.go |
| CreateVital | vital_handler.go |

## ※ CreateClinic は対象外（仕様確定）

`CreateClinic` はクリニック自体を新規作成するハンドラであり、`extractClinicID` も `clinic_id` パラメータも不要。修正対象から除外する。

## 再現確認済み

- **CreateCage**: POST `/api/v1/masters/cages` → 500
  - リクエスト: `{"name":"テストケージ_全項目入力","cage_type":"icu","cage_size":"large","price":9000,...}`
  - レスポンス: `{"error":"internal server error"}`
  - ネットワークログ: reqid=1741（2026-03-16）

## 正常ハンドラとの比較

```go
// ✅ 正常例: CreateVaccination (vaccination_handler.go)
func (h *Handler) CreateVaccination(c *gin.Context) {
    clinicID, ok := extractClinicID(c)  // ← あり
    if !ok {
        return
    }
    vaccination := &model.Vaccination{
        ClinicID: clinicID,  // ← 設定済み
        // ...
    }
}

// ❌ バグ例: CreateCage (cage_handler.go)
func (h *Handler) CreateCage(c *gin.Context) {
    // extractClinicID(c) を呼ばない
    cage := &model.Cage{
        Name:     input.Name,
        ClinicID: 0,  // ← ゼロ値のまま → FK違反 → 500
    }
}
```

## 修正方針

各ハンドラに以下を追加:

```go
func (h *Handler) CreateXxx(c *gin.Context) {
    clinicID, ok := extractClinicID(c)  // ← 追加
    if !ok {
        return
    }
    // ...
    xxx := &model.Xxx{
        ClinicID: clinicID,  // ← 追加
        // ...
    }
}
```

## 関連

- BE-036: Update系ハンドラ同様の問題（14件）
- 発見: マスタ設定ページ全ページ 登録テスト中に ChromeMCP で発見（2026-03-16）
- `grep` による全ハンドラスキャンで10ハンドラに同一バグが存在することを確認
