# API 仕様書 (API Specification)

> **Animal Ekarte**: 高信頼な動物病院バックエンド Go API
> **バージョン**: v3.1.0 | **最新更新**: 2026-05-27 | **ステータス**: Production Ready

## ⚠️ OpenAPI の Single Source of Truth (SSOT) 運用方針

本プロジェクトには `docs/openapi.yaml` と `backend/docs/api.yaml` の 2 つの OpenAPI 仕様書が存在します。以後の開発・運用において矛盾が発生しないよう、以下の運用方針を徹底します：

1. **真実の源泉 (SSOT)**: **`docs/openapi.yaml`** を唯一の SSOT とします。
   - `docker-compose.swagger.yml` の Swagger UI / Redoc コンテナは `docs/openapi.yaml` を読み込みます。
   - 認証方式（`access_token` / `refresh_token` の dual-token 構成）の最新化もこちらが正となります。
2. **同期（内部コピー）**: **`backend/docs/api.yaml`** は、詳細なAPIリクエスト/レスポンススキーマを保持するバックエンド内部向けの互換コピーです。
   - エンドポイントの追加・変更時には、必ず `docs/openapi.yaml` を正として更新し、必要に応じて最小差分で `backend/docs/api.yaml` に同期させてください。

---

## 1. 共通仕様

### 1.1 エンドポイント
- **Base URL**: `https://api.noah-karte.com/api/v1`
- **データ形式**: JSON (Request/Response)
- **文字コード**: UTF-8
- **日付形式**: ISO 8601 (UTC) `YYYY-MM-DDTHH:mm:ssZ`
- **金額単位**: 最小単位（円）

### 1.2 認証と認可
- **方式**: dual-token (Access/Refresh Token) 方式。
- **認可**: 全てのエンドポイントで RBAC 権限チェックが適用されます。
- **テナント分離**: `X-Clinic-ID` ヘッダーまたはトークン内の `clinic_id` に基づき、データは物理的に隔離されます。

---

## 2. 主要リソース別エンドポイント (Verified)

### 2.1 診療記録 (Medical Records)
カルテ本体と、それに紐付く多種多様な臨床データの管理。

- `GET /medical-records` — カルテ一覧（検索・ページング対応）
- `POST /medical-records` — カルテ新規作成（ID発番）
- `GET/PATCH /medical-records/:id` — 詳細取得・基本情報更新

#### ── サブリソース ──
- `GET/PUT /medical-records/:id/treatments` — 処置・処方明細。一括更新（PUT）対応。
- `GET/POST/DELETE /medical-records/:id/exams` — 検査記録と数値結果。
- `GET/POST/DELETE /medical-records/:id/vitals` — バイタル測定値（グラフ用）。
- `GET/POST/DELETE /medical-records/:id/images` — 患部写真・資料 PDF。
- `GET/POST/PATCH/DELETE /medical-records/:id/checkups` — 定期健診記録。
- `GET/PATCH /medical-records/:id/clinical-plan` — 身体検査所見・診断名。
- `POST /medical-records/:id/billing-confirmation/confirm` — 会計を医師確認済みにする。

### 2.2 入院管理 (Hospitalization)
- `GET /hospitalizations` — 入院中・予約患者一覧。
- `POST /hospitalizations` — 入院・ホテル預かり登録。
- `GET/POST /hospitalizations/:id/daily-records` — 日次ケア記録（朝・昼・夜）。
- `GET/POST/PATCH/DELETE /hospitalizations/:id/care-plan-items` — 入院中の指示計画。

### 2.3 会計・経営 (Accounting)
- `GET /accountings/unpaid` — 売掛金・未納者一覧。
- `GET /cash-register/preview` — レジ締め前の売上集計。
- `POST /cash-register/closes` — レジ締め確定保存。
- `GET /cash-register/closes` — レジ締め履歴一覧取得。
- `GET /cash-register/closes/:id` — レジ締め履歴詳細取得。
- `GET /reports/monthly` — 月次売上レポート。
- `GET /reports/monthly/csv` — 月次売上レポートの CSV エクスポート。
- `GET /clinics/:clinic_id/owners/aggregations` — 累計売上・来院頻度ランキング。

### 2.4 シフト設定 (Shifts)
- `GET/POST /clinic-holidays` — クリニック休診日一覧取得・登録。
- `DELETE /clinic-holidays/:date` — 指定日の休診日削除。

### 2.5 LINE/Lステップ連携 (CRM)
- `GET/PATCH /lstep-settings` — 連携状態・判定閾値の管理。
- `GET /lstep/tag-summary` — タグ別飼い主数集計（タグ分布統計）。
- `GET /lstep/owners` — Lステップタグ別の飼い主一覧（絞り込み・CSVダウンロード）。
- `GET /clinics/:clinic_id/lstep/delivery-monitor/logs` — 自動配信実行ログ。
- `GET /clinics/:clinic_id/lstep/delivery-monitor/summary` — 自動配信ステータス集計。
- `GET /clinics/:clinic_id/lstep/checkup-sync/preview` — 健診対象者とタグ同期プレビュー。
- `POST /clinics/:clinic_id/lstep/checkup-sync` — 健診タグ一括付与。

---

## 3. エラーレスポンス

システム共通の `RespondError` 形式を採用しています。

```json
{
  "error": "具体的なエラー理由",
  "code": "ERROR_CODE",
  "request_id": "uuid-v4"
}
```

| HTTP Status | 意味 |
|:---|:---|
| **400** | 不正な入力（型エラー、必須欠落、バリデーション失敗）。 |
| **401/403** | 認証エラー、またはリソースに対する操作権限不足。 |
| **404** | リソースが存在しない、または他テナントへのアクセス。 |
| **409** | データの衝突（重複登録、使用中のマスタ削除制限）。 |

---
