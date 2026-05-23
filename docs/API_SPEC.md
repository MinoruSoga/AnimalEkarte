# API 仕様書 (API Specification)

> **Animal Ekarte**: 高信頼な動物病院バックエンド Go API
> **バージョン**: v2.3 | **最新更新**: 2026-05-21 | **ステータス**: Production Ready

---

## 1. 共通仕様

### 1.1 エンドポイント
- **Base URL**: `https://api.noah-karte.com/api/v1`
- **データ形式**: JSON (Request/Response)
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
- `GET /reports/monthly` — 月次売上レポート。
- `GET /clinics/:clinic_id/owners/aggregations` — 累計売上・来院頻度ランキング。

### 2.4 LINE/Lステップ連携 (CRM)
- `GET/PATCH /lstep-settings` — 連携状態・判定閾値の管理。
- `GET /lstep/tags/summary` — タグ分布統計。
- `GET /clinics/:clinic_id/lstep/delivery-monitor/logs` — 自動配信実行ログ。
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
