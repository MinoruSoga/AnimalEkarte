# BE-049: 検査管理一覧 500エラー — exams.clinic_id カラム不一致

## 問題

`/examinations` ページで500エラーが発生。

```
ERROR: column exams.clinic_id does not exist (SQLSTATE 42703)
examination_repository.go:61
```

## 根本原因

`examination_repository.go` の `FindAll()` で `medical_records` テーブルを JOIN し `medical_records.clinic_id` でフィルタしているが、GORM が `Count()` と `Find()` で別々のクエリを発行する際に、テーブルスコープが失われ `exams.clinic_id` を参照しようとしている。

`exams` テーブル自体には `clinic_id` カラムが存在する（001_init.sql:727）が、GORM の内部エイリアス生成が `Preload` と `Joins` の組み合わせで不安定になっている。

## 修正方針

`medical_records` 経由の JOIN フィルタを廃止し、`exams.clinic_id` を直接 WHERE で指定する。

```go
// 修正前
q := r.db.WithContext(ctx).Model(&model.Examination{}).
    Joins("JOIN medical_records ON medical_records.id = exams.medical_record_id").
    Where("medical_records.clinic_id = ?", clinicID)

// 修正後
q := r.db.WithContext(ctx).Model(&model.Examination{}).
    Where("clinic_id = ?", clinicID)
```

## 対象ファイル

- `backend/internal/repository/examination_repository.go` — FindAll メソッド

## テスト

```bash
docker compose exec backend go test ./internal/repository/... -v -run Examination
curl -s http://localhost:8080/api/v1/examinations | jq .
```
