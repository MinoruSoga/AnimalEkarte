# Lステップ連携 実装仕様

Ekarte からLステップへの患者情報タグ連携・セグメントDM配信基盤の技術仕様。

---

## 1. 概要

| 項目 | 内容 |
|------|------|
| 連携方向（基本） | Ekarte → Lステップ（一方向同期） |
| 逆流（限定） | 友だち追加・ブロック・同意状態のみ Lステップ → Ekarte (Webhook) |
| 連携キー | `line_customers.line_user_id`（主キー）、Lステップ顧客ID（補助キー） |
| 同期方式 | イベント駆動（即時 + 再試行付き非同期キュー） |

---

## 2. スコープ

### In Scope
- **①カルテ連携**: Ekarteに登録された患者情報（飼い主・ペット）をLステップのタグへ紐づける
- **②セグメントDM配信**: タグ条件で対象を絞り込んだ配信対象制御・配信停止/同意フラグ管理

### Out of Scope
- Lステップ上のシナリオ設計・配信文面の作成（院内運用）
- Lステップ管理画面の操作・設定代行
- 配信結果レポート・分析基盤

---

## 3. システム構成

```
┌─────────────────────────────────────────────────────────┐
│                     Ekarte Backend                       │
│                                                          │
│  Event Handler ──► LstepSyncQueue ──► LstepAPIClient   │
│       │                  │                    │          │
│  (予約/来院/ワクチン/健診) │           LステップAPI       │
│                    line_sync_logs                        │
└──────────────────────────┬──────────────────────────────┘
                           │ Webhook
                    ┌──────▼──────┐
                    │  Lステップ   │
                    │  タグ管理    │
                    │  顧客属性    │
                    └─────────────┘
```

---

## 4. DB設計

### 4.1 既存テーブルへの追加カラム

```sql
-- line_customers に追加
ALTER TABLE line_customers
    ADD COLUMN lstep_customer_id VARCHAR(100),     -- Lステップ補助キー
    ADD COLUMN last_synced_at    TIMESTAMPTZ,      -- 最終同期日時
    ADD COLUMN sync_error        TEXT;             -- 直近エラー内容（NULL=正常）
```

### 4.2 新規テーブル

```sql
CREATE TYPE lstep_sync_event AS ENUM (
    'tag_add',
    'tag_remove',
    'attr_update'
);

CREATE TYPE lstep_sync_status AS ENUM (
    'pending',
    'success',
    'failed'
);

CREATE TABLE line_sync_logs (
    id             BIGSERIAL PRIMARY KEY,
    clinic_id      BIGINT      NOT NULL REFERENCES clinics(id),
    line_user_id   VARCHAR(100) NOT NULL,
    event_type     lstep_sync_event NOT NULL,
    payload        JSONB       NOT NULL,
    status         lstep_sync_status NOT NULL DEFAULT 'pending',
    retry_count    INT         NOT NULL DEFAULT 0,
    error_message  TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_line_sync_logs_clinic_status
    ON line_sync_logs(clinic_id, status)
    WHERE status IN ('pending', 'failed');

CREATE INDEX idx_line_sync_logs_line_user
    ON line_sync_logs(line_user_id, created_at DESC);
```

---

## 5. タグ設計

> **前提**: Lステップ管理画面でのタグ作成は完了済み。タグ一覧はLステップAPI（`GET /v1/tags`）から取得し、Ekarteイベントと紐づける。手動共有は不要。

### 5.1 タグ一覧（イベント駆動）

タグはイベント発生時に即時付与・解除する。

| イベント | 付与するタグ | 解除するタグ |
|---------|------------|------------|
| 予約作成（初回） | `予約あり` `来院予定` `初診` | — |
| 予約作成（2回目以降） | `予約あり` `来院予定` `再診` | `初診` |
| 予約キャンセル | `キャンセル歴あり` | `予約あり` `来院予定` |
| 来院完了（受付終了） | `来院済` `フォロー対象` `再診候補` | `来院予定` |
| ワクチン接種記録 | `ワクチン接種済` `次回案内対象` | — |
| 健診実施記録 | `健診済` `定期案内対象` | — |
| 次回案内期限切れ | — | `次回案内対象` `定期案内対象` |
| 友だち追加 (Webhook) | `LINE連携済` | — |
| 配信停止 | `配信停止中` | すべての配信対象タグ |

### 5.2 属性タグ（患者情報から静的付与）

| 条件 | タグ |
|------|------|
| `pets.species = '犬'` | `犬` |
| `pets.species = '猫'` | `猫` |
| `pets.species = その他` | `その他` |
| 年齢 < 1歳 | `子犬・子猫` |
| 1歳 ≤ 年齢 < 7歳 | `成犬・成猫` |
| 年齢 ≥ 7歳 | `シニア（7歳以上）` |
| `pets.gender = 'male'` | `オス` |
| `pets.gender = 'female'` | `メス` |
| ワクチン記録あり | `ワクチン管理対象` |
| 健診記録あり | `健診管理対象` |

> **ペット複数頭の場合**: 該当する全ペットのタグを飼い主アカウントへ付与する（要クライアント確認）

---

## 6. API連携仕様

### 6.1 Lステップ API（Ekarte → Lステップ）

| 操作 | メソッド・パス | タイミング |
|------|------------|---------|
| **タグ一覧取得** | `GET /v1/tags` | 起動時・定期同期（タグマスタ取得） |
| タグ付与 | `POST /v1/contacts/{userId}/tags` | イベント発生時（即時） |
| タグ解除 | `DELETE /v1/contacts/{userId}/tags` | 条件変化時 |
| 顧客属性更新 | `PUT /v1/contacts/{userId}` | 飼い主・ペット情報更新時 |
| 顧客検索 | `GET /v1/contacts?line_user_id={id}` | 初回紐づけ時 |

### 6.2 Go実装パターン

```go
// internal/service/lstep_sync_service.go

type LstepSyncService interface {
    SyncTagsOnAppointmentCreated(ctx context.Context, appointmentID uint64) error
    SyncTagsOnVisitCompleted(ctx context.Context, appointmentID uint64) error
    SyncTagsOnVaccineRecorded(ctx context.Context, vaccinationID uint64) error
    SyncTagsOnCheckupRecorded(ctx context.Context, checkupID uint64) error
    SyncPatientAttributes(ctx context.Context, ownerID uint64) error
    RetryFailedSyncs(ctx context.Context, clinicID uint64) error
}

// internal/infrastructure/lstep/client.go

type LstepClient interface {
    AddTags(ctx context.Context, lineUserID string, tags []string) error
    RemoveTags(ctx context.Context, lineUserID string, tags []string) error
    UpdateContact(ctx context.Context, lineUserID string, attrs map[string]any) error
    FindContactByLineUserID(ctx context.Context, lineUserID string) (*LstepContact, error)
}
```

### 6.3 Webhook受信（Lステップ → Ekarte）

```
POST /api/webhooks/lstep
```

| イベント | Ekarte での処理 |
|---------|---------------|
| `follow` | `line_customers` レコード作成・`LINE連携済`タグ付与 |
| `unfollow` | `is_blocked = true` 更新 |
| `consent_agreed` | `is_consent = true` 更新 |
| `consent_revoked` | `is_consent = false`・配信停止フラグ更新 |

---

## 7. エラーハンドリング・再試行

### フロー

```
API呼び出し失敗
    ↓
line_sync_logs に status='failed', retry_count++ で記録
    ↓
retry_count < 3 → 指数バックオフで再試行（1分 / 5分 / 30分）
    ↓
retry_count = 3 → status='failed' で停止
    ↓
管理画面から手動再同期 or バッチで再試行
```

### 再試行バッチ

```go
// 失敗レコードを定期的にリトライ（cronジョブ）
func (s *LstepSyncService) RetryFailedSyncs(ctx context.Context, clinicID uint64) error {
    logs, err := s.repo.FindFailedSyncLogs(ctx, clinicID, maxRetry)
    // ...
}
```

---

## 8. セグメントDM配信

Lステップ側のセグメント配信機能を使用する。Ekarte側はタグを正確に管理することで配信対象を間接制御する。

### セグメント例

| 配信目的 | タグ条件 |
|---------|---------|
| 猫ワクチン案内 | `猫` ＋ `次回案内対象` |
| 高齢犬の健診推奨 | `犬` ＋ `シニア（7歳以上）` ＋ `定期案内対象` |
| 初回来院フォロー | `初診` ＋ `来院済` |
| 再来院促進 | `再診候補` |
| 一斉配信（全体） | `LINE連携済` ＋ `配信停止中` 除外 |

### 配信除外条件（必須）

配信時は必ず以下を除外すること：
- `is_consent = false`（同意未取得）
- `is_blocked = true`（配信停止）
- `配信停止中` タグ付き

---

## 9. 実装フェーズ

| フェーズ | 内容 | 優先度 |
|---------|------|-------|
| Phase 1 | Lステップ APIクライアント実装・認証・疎通確認 | 高 |
| Phase 2 | イベント起点タグ付与・解除（予約/来院/ワクチン/健診） | 高 |
| Phase 3 | 飼い主・ペット属性同期・Webhook受信実装 | 中 |
| Phase 4 | 再試行キュー・同期ログ・手動再同期API | 中 |
| Phase 5 | 年齢帯タグの定期バッチ更新（誕生日起点） | 低 |

---

## 10. 未確認事項（クライアント確認待ち）

- [x] ~~タグ名称の確定~~ → Lステップ管理画面で作成済み。`GET /v1/tags` で取得してマッピングに使用する
- [ ] セグメント配信の具体的なユースケースと条件
- [ ] 配信停止・同意管理の正本（Ekarte vs Lステップ）
- [ ] ペット複数頭の場合のタグ付与方針
- [ ] 同期失敗時の通知先・運用担当者
- [ ] 自動配信と手動配信の境界
- [ ] 配信頻度・時間帯制限の有無
- [ ] No.23スコープ（タグ基盤のみ or シナリオ設計支援まで）
