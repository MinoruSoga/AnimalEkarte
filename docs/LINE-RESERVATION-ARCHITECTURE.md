# LINE予約システム アーキテクチャ全体像

> **作成日**: 2026-04-09
> **関連ドキュメント**: [LINE-SETUP.md](./LINE-SETUP.md) / [仕様書](./line-reseavation.md) / [タスク一覧](./tasks/open/reservation/00-OVERVIEW.md)

---

## 1. システム全体図

```
                          ┌─────────────────────────────────┐
                          │         LINE Platform           │
                          │                                 │
                          │  ┌───────────┐  ┌────────────┐ │
                          │  │ LIFF SDK  │  │ Messaging  │ │
                          │  │ (認証)    │  │ API (通知) │ │
                          │  └─────┬─────┘  └─────▲──────┘ │
                          └───────│───────────────│────────┘
                                  │               │
              ┌───────────────────┼───────────────┼──────────────────┐
              │                   │ ID Token      │ Push通知         │
              │    ┌──────────────▼──────────────┐ │                  │
              │    │      LIFF App (React)       │ │                  │
              │    │   reserve.noah-karte.com    │ │                  │
              │    │                             │ │                  │
              │    │  Top → Step1〜8 → Complete  │ │                  │
              │    │  マイ予約 → キャンセル       │ │                  │
              │    └──────────────┬──────────────┘ │                  │
              │                   │ /api/liff/     │                  │
              │    ┌──────────────▼──────────────────────────────┐   │
              │    │          Backend API (Go/Gin)               │   │
              │    │                                             │   │
              │    │  ┌─────────────┐   ┌─────────────────────┐ │   │
              │    │  │  LiffAuth   │   │  管理者API          │ │   │
              │    │  │MiddleWare   │   │ /v1/clinics/:id/    │ │   │
              │    │  │(IDToken検証)│   │ reservation-*       │ │   │
              │    │  └──────┬──────┘   └──────────▲──────────┘ │   │
              │    │         │                      │            │   │
              │    │  ┌──────▼──────────────────────┼─────────┐ │   │
              │    │  │         Service Layer        │         │ │   │
              │    │  │                              │         │ │   │
              │    │  │  LiffService ──► TimeslotEngine       │ │   │
              │    │  │       │         (時間枠計算)           │ │   │
              │    │  │       │                                │ │   │
              │    │  │       ▼                                │ │   │
              │    │  │  Notifier ──► LINE Push ───────────────┼─┘   │
              │    │  │       │                                │     │
              │    │  │       └──► Email (SMTP)                │     │
              │    │  └────────────────────────────────────────┘     │
              │    │         │                                       │
              │    └─────────┼───────────────────────────────────────┘
              │              │                                       
              │    ┌─────────▼──────────────┐                       
              │    │   PostgreSQL            │                       
              │    │                         │                       
              │    │  reservation_settings   │                       
              │    │  reservation_customers  │                       
              │    │  reservation_appointments│                      
              │    │  shift_entries + breaks  │                       
              │    │  service_types (拡張)    │                       
              │    │  staffs (拡張)           │                       
              │    └─────────────────────────┘                       
              │                                                      
              │    ┌─────────────────────────────────────────┐       
              │    │    管理画面 (既存 AnimalEkarte)          │       
              │    │    noah-karte.com/line-reservation       │       
              │    │                                         │       
              │    │  カレンダー / コース / スタッフ / 設定   │       
              │    │  スケジュール / ページ編集 / 顧客管理    │       
              │    └─────────────────────────────────────────┘       
              │                                                      
              │                    インターネット                     
              └──────────────────────────────────────────────────────┘
```

---

## 2. 予約フロー（顧客側 — LIFF App）

```
  顧客が LINE リッチメニューの「予約する」をタップ
    │
    ▼
┌─────────────────────────────────────────────────────────┐
│  LIFF App 起動                                          │
│                                                         │
│  1. liff.init(liffId)                                   │
│  2. 未ログイン → liff.login() → LINE認証画面            │
│  3. ログイン済み → liff.getIDToken() → JWT取得          │
│  4. GET /api/liff/:clinicId/settings → 営業状態確認      │
│     ├─ status="stopped" → メンテナンス画面              │
│     └─ status="running" → トップページ表示              │
└────────────────────┬────────────────────────────────────┘
                     │ 「新規予約」タップ
                     ▼
┌─────────────────────────────────────────────────────────┐
│                                                         │
│  STEP 1: お客様情報入力                                  │
│  ┌───────────────────────────────────┐                  │
│  │ お名前 / 電話番号 / 飼い主名      │                  │
│  │ ペットの名前と種類 / 診察内容     │                  │
│  └───────────────────────────────────┘                  │
│  ※ 2回目以降: GET /profile でプリフィル                  │
│                      │                                  │
│  STEP 2: コース選択   ▼                                  │
│  ┌───────────────────────────────────┐                  │
│  │ GET /courses → is_internal=false のみ表示            │
│  └───────────────────────────────────┘                  │
│                      │                                  │
│  STEP 3: スタッフ選択 ▼                                  │
│  ┌───────────────────────────────────┐                  │
│  │ GET /staffs?courseId=X            │                  │
│  │ 「指名なしで予約」オプションあり   │                  │
│  └───────────────────────────────────┘                  │
│                      │                                  │
│  STEP 4: 日付選択     ▼                                  │
│  ┌───────────────────────────────────┐                  │
│  │ GET /available-dates              │                  │
│  │ 月間カレンダー（不可日グレーアウト）│                  │
│  └───────────────────────────────────┘                  │
│                      │                                  │
│  STEP 5: 時間選択     ▼                                  │
│  ┌───────────────────────────────────┐                  │
│  │ GET /available-times?date=X       │                  │
│  │ 空き時間枠リスト表示              │                  │
│  └───────────────────────────────────┘                  │
│                      │                                  │
│  STEP 6: 要望入力     ▼                                  │
│  ┌───────────────────────────────────┐                  │
│  │ テキストエリア（任意入力）         │                  │
│  └───────────────────────────────────┘                  │
│                      │                                  │
│  STEP 7: 確認画面     ▼                                  │
│  ┌───────────────────────────────────┐                  │
│  │ 全入力内容のサマリー表示          │                  │
│  │ 「まだ予約は完了していません」警告  │                  │
│  │ [確認して予約する]ボタン           │                  │
│  └───────────────────────────────────┘                  │
│                      │ POST /reservations               │
│  STEP 8: 完了画面     ▼                                  │
│  ┌───────────────────────────────────┐                  │
│  │ 「ご予約を承りました」            │                  │
│  │ 予約番号 / 日時 / コース表示      │                  │
│  └───────────────────────────────────┘                  │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

---

## 3. 認証フロー

```
┌──────────┐    ┌──────────────┐    ┌─────────────────┐    ┌──────────┐
│ LIFF App │    │ Backend API  │    │ LINE Platform   │    │    DB    │
└────┬─────┘    └──────┬───────┘    └───────┬─────────┘    └────┬─────┘
     │                 │                    │                    │
     │ liff.init()     │                    │                    │
     │────────────────────────────────────►│                    │
     │                 │     ID Token       │                    │
     │◄────────────────────────────────────│                    │
     │                 │                    │                    │
     │ Authorization:  │                    │                    │
     │ Bearer {token}  │                    │                    │
     │────────────────►│                    │                    │
     │                 │                    │                    │
     │                 │  POST /oauth2/     │                    │
     │                 │  v2.1/verify       │                    │
     │                 │───────────────────►│                    │
     │                 │  { sub, name }     │                    │
     │                 │◄───────────────────│                    │
     │                 │                    │                    │
     │                 │  FindOrCreate      │                    │
     │                 │  ByLineUserID      │                    │
     │                 │───────────────────────────────────────►│
     │                 │  reservation_customer                  │
     │                 │◄───────────────────────────────────────│
     │                 │                    │                    │
     │  200 OK + data  │                    │                    │
     │◄────────────────│                    │                    │
     │                 │                    │                    │

認証結果がGinコンテキストに設定される:
  - liff_customer_id   （reservation_customers.id）
  - liff_clinic_id     （clinicId パスパラメータ）
  - liff_line_user_id  （LINE の sub = ユーザー固有ID）
  - liff_display_name  （LINE 表示名）
```

---

## 4. 時間枠計算エンジン

予約システムの**核心ロジック**。DBアクセスを持たない純粋な計算関数。

```
                    入力データ
                       │
      ┌────────────────┼────────────────┐
      │                │                │
      ▼                ▼                ▼
 ┌──────────┐  ┌──────────────┐  ┌──────────────────┐
 │ 営業時間  │  │ スタッフ設定  │  │ 既存予約          │
 │ 休憩時間  │  │ 個人シフト    │  │ (重複除外用)      │
 │ 休業日    │  │ 休日設定      │  │                  │
 └─────┬────┘  └──────┬───────┘  └────────┬─────────┘
       │               │                   │
       └───────────────┼───────────────────┘
                       │
                       ▼
            ┌─────────────────────┐
            │  GenerateTimeSlots  │
            │  (timeslot_engine)  │
            │                     │
            │  1. 営業区間を生成   │
            │  2. 休憩時間を除外   │
            │  3. 個人設定を反映   │
            │  4. 既存予約を除外   │
            │  5. コース時間で分割  │
            │  6. minimize_gaps    │
            │     モード適用       │
            └──────────┬──────────┘
                       │
                       ▼
              ┌─────────────────┐
              │  TimeSlot[]     │
              │                 │
              │  09:00 - 09:15  │
              │  09:15 - 09:30  │
              │  10:00 - 10:15  │
              │  ...            │
              └─────────────────┘

  呼び出し階層:
  ┌──────────────────────────────────────────┐
  │  LiffService                             │
  │    └─► CalcAvailableDates                │
  │          ├─► 営業日判定（祝日含む）        │
  │          └─► GenerateTimeSlots (per day)  │
  │                └─► TimeSlot[] 返却        │
  └──────────────────────────────────────────┘
```

---

## 5. 通知フロー

```
  POST /api/liff/:clinicId/reservations
    │
    ▼
  LiffService.CreateReservation()
    │
    ├─ 1. バリデーション（予約制限チェック）
    │     ├─ 同日予約制限
    │     ├─ 同月予約制限
    │     ├─ 時間枠重複チェック（SELECT FOR UPDATE）
    │     └─ NG → 409 Conflict + エラーコード
    │
    ├─ 2. reservation_appointments INSERT
    │     └─ source = "line"
    │
    └─ 3. notifier.NotifyCreated() ← fire-and-forget（別goroutine）
          │
          ├──────────────────────┐
          │                      │
          ▼                      ▼
   LINE Push通知            メール通知
   ┌──────────────┐    ┌──────────────┐
   │ POST /v2/bot │    │ SMTP送信     │
   │ /message/push│    │              │
   │              │    │ To: 病院     │
   │ To: 顧客LINE │    │ notification │
   │              │    │ _email       │
   │ ┌──────────┐ │    │              │
   │ │予約確認  │ │    │ 予約詳細     │
   │ │R-000123  │ │    │ + 顧客情報   │
   │ │4/10 9:00 │ │    │ + ペット情報 │
   │ │一般診察  │ │    │              │
   │ └──────────┘ │    └──────────────┘
   └──────────────┘
   
   ※ 通知失敗しても予約自体は成功する（fire-and-forget）
   ※ channelToken 未設定時は LINE Push をスキップ
```

---

## 6. DBテーブル関連図

```
  ┌─────────────┐
  │   clinics    │
  └──────┬──────┘
         │ 1:1
         ▼
  ┌──────────────────────┐        ┌──────────────────────────────┐
  │ reservation_settings │        │     service_types (拡張)     │
  ├──────────────────────┤        ├──────────────────────────────┤
  │ clinic_id (UNIQUE)   │        │ + reservation_visible (BOOL) │
  │ status               │        │ + reservation_comment (TEXT)  │
  │ business_hours       │        │ + reservation_image_url      │
  │ break_hours          │        │ + is_internal (BOOL)         │
  │ closed_weekdays      │        │ + reservation_display_order  │
  │ booking_window_*     │        └──────────┬───────────────────┘
  │ line_channel_id      │                   │
  │ line_channel_secret  │                   │ M:N
  │ liff_id              │                   │
  │ line_access_token    │     ┌─────────────┴─────────────┐
  └──────────────────────┘     │ staff_excluded_service_    │
                               │ types                      │
  ┌──────────────────────┐     ├────────────────────────────┤
  │   staffs (拡張)      │     │ staff_id                   │
  ├──────────────────────┤     │ service_type_id            │
  │ + reservation_visible│◄────┘ (UNIQUE)                   │
  │ + staff_type         │     └────────────────────────────┘
  │ + reservation_       │
  │   display_order      │
  └──────────┬───────────┘
             │ 1:N
             ▼
  ┌──────────────────────┐     ┌──────────────────────┐
  │    shift_entries      │────►│  shift_entry_breaks  │
  ├──────────────────────┤ 1:N ├──────────────────────┤
  │ staff_id             │     │ shift_entry_id       │
  │ date                 │     │ break_start          │
  │ start_time           │     │ break_end            │
  │ end_time             │     └──────────────────────┘
  │ is_holiday           │
  └──────────────────────┘

  ┌──────────────────────────────────────────────┐
  │        reservation_appointments (拡張)        │
  ├──────────────────────────────────────────────┤
  │ clinic_id                                    │
  │ + source ("manual" | "line")  ← LINE予約識別 │
  │ + line_customer_id ──────────────────┐       │
  │ + customer_fields (JSONB)            │       │
  │ + is_staff_delegated (BOOL)          │       │
  └──────────────────────────────────────┼───────┘
                                         │ FK
                                         ▼
                          ┌──────────────────────────┐
                          │  reservation_customers    │
                          ├──────────────────────────┤
                          │ clinic_id + line_user_id  │ UNIQUE
                          │ display_name (LINE名)     │
                          │ real_name                 │
                          │ additional_fields (JSONB) │
                          │ owner_id ─────────┐      │
                          └───────────────────┼──────┘
                                              │ FK (nullable)
                                              ▼
                                    ┌──────────────┐
                                    │    owners     │
                                    │  (既存テーブル) │
                                    └──────────────┘
```

---

## 7. 管理画面（AnimalEkarte 内）

```
  /line-reservation
    │
    ├── /calendar (デフォルト)
    │     予約カレンダー
    │     ├─ 月表示: 日ごとの予約件数
    │     ├─ 日表示: 30分グリッド × スタッフ列
    │     ├─ 手動予約入力
    │     └─ 予約キャンセル
    │
    ├── /courses
    │     コース設定（CRUD）
    │     ├─ コース名 / 時間 / コメント
    │     ├─ is_internal（管理者専用 = LIFF非表示）
    │     └─ 並び順変更（ドラッグ）
    │
    ├── /staffs
    │     スタッフ設定（CRUD）
    │     ├─ 名前 / 肩書 / 施設名
    │     ├─ 非対応コース設定
    │     └─ staff_type (veterinarian / groomer / ...)
    │
    ├── /schedule
    │     個人スケジュール管理
    │     ├─ ガントチャート表示
    │     ├─ 日単位の営業時間上書き
    │     ├─ 休日設定
    │     └─ 個人休憩時間設定
    │
    ├── /settings
    │     基本設定
    │     ├─ 営業状態（Running / Stopped）
    │     ├─ 営業時間 / 休憩時間 / 休業曜日
    │     ├─ 予約受付期間（最小〜最大日数）
    │     ├─ 同日・同月予約制限
    │     ├─ 時間枠モード（minimize_gaps / allow_gaps）
    │     ├─ 指名なしモード（first_available / top_priority）
    │     └─ LINE連携設定（Channel ID / Secret / LIFF ID / Token）
    │
    ├── /page-editor
    │     ページ編集
    │     ├─ トップページ ヘッダーテキスト
    │     ├─ 予約時の注意事項
    │     ├─ キャンセルポリシー
    │     └─ プライバシーポリシー
    │
    └── /customers
          LINE顧客管理
          ├─ LINE顧客一覧
          ├─ 既存オーナーとの紐付け
          └─ 予約履歴閲覧
```

---

## 8. API エンドポイント一覧

### 公開API（LIFF App 用）— `/api/liff/:clinicId/`

| Method | Path | 認証 | 説明 |
|--------|------|------|------|
| GET | `/settings` | なし | 営業状態・LIFF ID取得 |
| GET | `/profile` | LiffAuth | 顧客プロフィル（プリフィル用） |
| GET | `/courses` | LiffAuth | 公開コース一覧 |
| GET | `/staffs` | LiffAuth | スタッフ一覧（コース絞り込み） |
| GET | `/available-dates` | LiffAuth | 予約可能日一覧 |
| GET | `/available-times` | LiffAuth | 空き時間枠一覧 |
| POST | `/reservations` | LiffAuth | 予約確定 |
| GET | `/my-reservations` | LiffAuth | 自分の予約一覧 |
| DELETE | `/my-reservations/:id` | LiffAuth | 予約キャンセル |

### 管理者API — `/v1/clinics/:clinicId/`

| Method | Path | 説明 |
|--------|------|------|
| GET/PUT | `reservation-settings` | 基本設定 |
| GET/POST | `reservation-courses` | コース一覧/作成 |
| GET/PUT/DELETE | `reservation-courses/:id` | コース詳細/更新/削除 |
| PUT | `reservation-courses/order` | コース並び順変更 |
| GET/POST | `reservation-staffs` | スタッフ一覧/作成 |
| GET/PUT/DELETE | `reservation-staffs/:id` | スタッフ詳細/更新/削除 |
| PUT | `reservation-staffs/order` | スタッフ並び順変更 |
| GET/PUT/DELETE | `reservation-staffs/:id/schedules` | 個人スケジュール |
| GET | `line-reservations` | 予約一覧（月/日） |
| POST | `line-reservations` | 手動予約入力 |
| DELETE | `line-reservations/:id` | 予約キャンセル |
| GET | `reservation-customers` | LINE顧客一覧 |
| PATCH | `reservation-customers/:id/link-owner` | オーナー紐付け |

---

## 9. 技術スタック

| レイヤー | 技術 | 備考 |
|---------|------|------|
| LIFF App | React 19 + TypeScript + Vite + Tailwind | 独立プロジェクト (`liff-app/`) |
| LIFF SDK | `@line/liff` | 認証 + プロフィル取得 |
| 管理画面 | React 19 (既存 AnimalEkarte) | `features/line-reservation/` |
| Backend | Go 1.25 + Gin + GORM | 既存バックエンドに統合 |
| 認証 (LIFF) | LINE ID Token + `/oauth2/v2.1/verify` | LiffAuth ミドルウェア |
| 認証 (管理) | 既存セッション認証 + RBAC | `reservation:read/write` |
| 通知 | LINE Messaging API + SMTP | fire-and-forget |
| DB | PostgreSQL 18 | 既存テーブル拡張 + 新規4テーブル |
| 祝日判定 | `holiday_jp-go` | 外部API依存なし |
| 時間枠計算 | 純粋関数 (`timeslot_engine.go`) | DBアクセスなし |

---

## 10. デプロイ構成

```
  ┌─────────────────────────────────────────────────┐
  │                  本番環境                        │
  │                                                 │
  │  noah-karte.com          reserve.noah-karte.com │
  │  ┌──────────────┐       ┌──────────────┐       │
  │  │  管理画面     │       │  LIFF App    │       │
  │  │  (Vercel)    │       │  (Vercel)    │       │
  │  └──────┬───────┘       └──────┬───────┘       │
  │         │                      │                │
  │         │     API Gateway      │                │
  │         └──────────┬───────────┘                │
  │                    ▼                            │
  │           ┌──────────────┐                      │
  │           │ Backend API  │                      │
  │           │ (ECS/Fargate)│                      │
  │           └──────┬───────┘                      │
  │                  │                              │
  │           ┌──────▼───────┐                      │
  │           │ PostgreSQL   │                      │
  │           │ (RDS)        │                      │
  │           └──────────────┘                      │
  └─────────────────────────────────────────────────┘

  LINE Platform ◄──► Backend API
    - LIFF認証: ID Token 検証
    - Push通知: Messaging API
```

---

## 11. 実装ステータス

| Phase | 内容 | 状態 |
|-------|------|------|
| Phase 1 | DB・モデル基盤 | 完了 |
| Phase 2 | 管理者API（14エンドポイント） | 完了 |
| Phase 3 | LIFF公開API + 時間枠エンジン | 完了 |
| Phase 4 | 管理画面フロントエンド（7画面） | 完了 |
| Phase 5 | LIFF App（12ページ） | 完了 |
| Phase 6 | LINE Push通知 + メール通知 | 完了 |
| **Phase 7** | **CORS + Docker + 結合テスト + デプロイ** | **未着手** |

### Phase 7 残タスク

1. CORS設定（`reserve.noah-karte.com` 許可追加）
2. Docker Compose に `liff-app` コンテナ追加
3. 結合テスト（LINE実機での全フロー通し確認）
4. staging デプロイ
5. リッチメニュー作成（[LINE-SETUP.md](./LINE-SETUP.md) 手順6）
6. 管理画面へのクレデンシャル入力（[LINE-SETUP.md](./LINE-SETUP.md) 手順7）
