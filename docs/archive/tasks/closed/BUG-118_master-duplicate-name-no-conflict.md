# BUG-118: 全マスタ共通 — 重複名称登録エラーなし (201)

## 概要
複数のマスタ登録エンドポイントで、既存と同名のアイテムを POST しても
409 Conflict を返さずに 201 Created で登録成功してしまう。
同名のマスタ項目が複数存在する状態になる。

## 再現手順
1. 任意のマスタページ（例: 動物種類マスタ `/settings/animal-species`）を開く
2. 「新規登録」→ サイドパネルで既存と同じ名称（例: 「犬」）を入力
3. 「保存」をクリック
4. → 「登録しました」トーストが表示され、一覧に「犬」が2件並ぶ

## 実確認 (ローカル確認: 2026-04-01)
| エンドポイント | 重複 POST のステータス | 状態 |
|--------------|----------------------|------|
| `POST /api/v1/masters/chief-complaint-categories` | **201** | ❌ バグあり |
| `POST /api/v1/masters/animal-species` | **201** | ❌ バグあり |
| `POST /api/v1/masters/medicines` | **201** | ❌ バグあり |
| `POST /api/v1/masters/job-titles` | **201** | ❌ バグあり |
| `POST /api/v1/masters/trimming-courses` | **201** | ❌ バグあり |
| `POST /api/v1/masters/hospitalization-plans` | **201** | ❌ バグあり |
| `POST /api/v1/masters/insurances` | **201** | ❌ バグあり |
| `POST /api/v1/masters/consultations` | **201** | ❌ バグあり |
| `POST /api/v1/masters/service-types` | **409** | ✅ 正常 |
| `POST /api/v1/masters/diagnosis-categories` | **409** | ✅ 正常 |

## 期待動作
- 同一 `clinic_id` 内で同名のマスタ項目が存在する場合 → **409 Conflict**
- エラーメッセージ: 「同じ名称が既に登録されています」

## 修正方針
各 service 層の `Create` メソッドに重複チェックを追加する。
または DB の `(clinic_id, name)` に UNIQUE 制約 + `ON CONFLICT` ハンドリングを追加する。

### 修正が必要なファイル (backend)
- `internal/service/master_service.go` の各 Create 系メソッド
- または `backend/migrations/` にインデックス追加マイグレーション

## テストデータ汚染
この BUG 確認テストにより以下が重複登録されているため、DB 修正が必要:
- `animal_species`: 「犬」が2件
- `chief_complaint_categories`: 「嘔吐・下痢」が2件（うち1件は食欲不振削除後に追加したもの）
