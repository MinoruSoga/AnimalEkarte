# BE-029: clinic_id フィルタ欠落 — マスタリポジトリのマルチテナント修正

## 問題
以下のリポジトリがclinic_idフィルタなしにデータを返しており、
クリニック間でのデータ漏洩リスクがある。

## 影響ファイル・箇所
- `backend/internal/repository/trimming_master_repository.go:34,42,122,130`
  — TrimmingCourse / TrimmingOption の Find/First にclinic_idなし
- `backend/internal/repository/chief_complaint_category_repository.go:44`
  — clinic_idなし
- `backend/internal/repository/consultation_repository.go:34,42`
  — Find/First にclinic_idなし
- `backend/internal/repository/insurance_repository.go:32,40`
  — Find/First にclinic_idなし
- `backend/internal/repository/hospitalization_plan_repository.go:34,42`
  — Find/First にclinic_idなし

## 確認事項
各テーブルにclinic_idカラムが存在するか確認の上、存在する場合は
全クエリに `.Where("clinic_id = ?", clinicID)` を追加する。

## 修正方針
1. 各モデルのスキーマ（migration/001_init.sql）でclinic_idの有無を確認
2. clinic_idが存在するテーブルは全repository関数にclinic_idフィルタを追加
3. サービス層のインターフェースにclinicIDパラメータを追加（必要な場合）

## 優先度
CRITICAL（セキュリティ・データ分離）
