# Phase 3: バックエンド公開API（LIFF用）

> **設計方針**: 予約データは既存 `reservation_appointments` に直接INSERT。
> LINE予約確定 → reservation_appointments にレコード作成 → カルテシステムに自動反映。
> `source = 'line'` で手動予約（`source = 'manual'`）と区別する。

## TASK-RES-020: LIFF認証ミドルウェア

**概要**: LINE LIFF ID Tokenを検証し、`reservation_customers` を特定するミドルウェア。

**対象ファイル**: `backend/internal/middleware/liff_auth.go`（新規）

**処理フロー**:
```
1. Authorization: Bearer {ID Token} を取得
2. LINE API (https://api.line.me/oauth2/v2.1/verify) でトークン検証
3. line_user_id を抽出
4. reservation_customers から検索（なければ新規作成）
5. context に customer_id, clinic_id をセット
```

**完了条件**:
- [ ] 有効なLIFF ID Tokenで認証成功
- [ ] 無効なトークンで401
- [ ] 新規ユーザー自動作成

---

## TASK-RES-021: 公開予約フローAPI

**エンドポイント**:
```
GET    /api/liff/settings                  # トップページ情報
GET    /api/liff/profile                   # 前回入力値の復元
GET    /api/liff/courses                   # コース一覧（is_internal=false）
GET    /api/liff/staffs?courseId=:id       # スタッフ一覧（コース絞り込み）
GET    /api/liff/available-dates?courseId=:id&staffId=:id
GET    /api/liff/available-times?courseId=:id&staffId=:id&date=2026-04-10
POST   /api/liff/reservations              # 予約確定
GET    /api/liff/my-reservations           # 自分の予約一覧
DELETE /api/liff/my-reservations/:id       # キャンセル
```

**対象ファイル（すべて新規）**:
- `backend/internal/handler/liff_handler.go`
- `backend/internal/handler/liff_request.go`
- `backend/internal/handler/liff_response.go`
- `backend/internal/service/liff_service.go`

※repositoryは Phase 2 のものを再利用。

**予約確定時の確認番号生成ルール**:
```
フォーマット: "R-" + YYYYMMDD + "-" + シーケンス4桁（日ごとリセット）
例: R-20260410-0001, R-20260410-0002
```

**LIFF URL パラメータ仕様**:
```
本番: https://reserve.noah-karte.com/{clinicId}
開発: http://localhost:3001/{clinicId}
LIFF認証ミドルウェアはURLパスからclinic_idを抽出する。
```

**プロフィール保存タイミング**:
POST `/api/liff/reservations`（予約確定）時に、STEP 1で入力された顧客情報を `reservation_customers` に自動保存。次回 GET `/api/liff/profile` で復元される。

**完了条件**:
- [ ] 全エンドポイント実装
- [ ] LIFF認証ミドルウェアで保護

---

## TASK-RES-022: 時間枠生成エンジン（★核心ロジック）

**概要**: 空き時間枠を計算するコアビジネスロジック。Phase 3 の最重要タスク。

**対象ファイル（新規）**:
- `backend/internal/service/timeslot_engine.go`
- `backend/internal/service/timeslot_engine_test.go`

**入力**:
```go
type TimeSlotInput struct {
    Date           time.Time
    CourseID       uint64
    StaffID        uint64              // 0 = 指名なし
    Settings       ReservationSetting
    StaffSchedule  *StaffScheduleOverride  // nilなら基本設定を使用
    Reservations   []Reservation           // 当日の既存予約
    CourseDuration int                     // 分
    AllCourses     []ReservationCourse     // 全コース（最短時間計算用）
}
```

**処理**:
```
1. 当日の勤務時間を決定（個人設定 > 基本設定）
2. 休憩時間を除外
3. 既存予約の時間帯を除外
4. 残りの空き時間からコース所要時間が収まる枠を列挙
5. time_slot_mode に応じて候補を生成:
   A) allow_gaps: 指定間隔で候補生成
   B) minimize_gaps: 最短コース時間を考慮し空きを最小化
```

**出力**:
```go
type TimeSlot struct {
    StartTime string // "0900"
    EndTime   string // "0915"
}
```

**テストケース（必須）**:
- [ ] 基本: 営業時間内で枠生成
- [ ] 休憩時間をまたぐ枠が除外される
- [ ] 既存予約と重複する枠が除外される
- [ ] 個人設定で営業時間が変更された場合
- [ ] 個人設定で休日の場合（空リスト）
- [ ] allow_gaps モード: 指定間隔で生成
- [ ] minimize_gaps モード: 最短コース考慮
- [ ] 指名なし: 全有効スタッフの空き時間を統合
- [ ] 60分コース（手術）の枠生成
- [ ] 15分コース（一般診察）の枠生成

---

## TASK-RES-023: 空き日付計算

**対象ファイル**: `backend/internal/service/available_dates.go`（新規）

**処理**:
```
1. booking_window_min_days 〜 booking_window_max_days の範囲を計算
2. 各日について:
   a. 休業曜日チェック
   b. 休業日チェック
   c. 祝日チェック（national_holiday_closed時）
   d. コースの曜日オプションチェック
   e. スタッフ個人設定の休日チェック
   f. 時間枠が1つ以上存在するかチェック
3. 予約可能な日付のみ返す
```

**祝日データの取得方法**:
```
Go標準ライブラリ or OSSの祝日計算パッケージを使用。
推奨: github.com/holiday-jp/holiday_jp-go
日本の祝日（移動祝日含む）を年単位で計算。外部API依存なし。
ハードコードではなくライブラリに委譲し、法改正時はライブラリ更新で対応。
```

**完了条件**:
- [ ] 予約受付期間外の日付が除外される
- [ ] 休業日・祝日が除外される（祝日ライブラリで判定）
- [ ] 曜日オプション（土曜限定等）が機能する
- [ ] スタッフ個人休日が反映される

---

## TASK-RES-024: 予約制限チェック

**対象ファイル**: `backend/internal/service/reservation_validators.go`（新規）

**チェック項目**:
```
1. 同日予約制限: 同一customer_idの当日予約数 < daily_limit
2. 同月予約制限: 同一customer_idの当月予約数 < monthly_limit
3. 時間枠の空き: 選択した枠が確定時にまだ空いているか（楽観ロック）
4. 稼働状態: settings.status == "running"
```

**楽観ロック実装方式**:
```
方式: SELECT FOR UPDATE（PostgreSQL行ロック）
処理:
1. トランザクション開始
2. SELECT ... FROM reservations WHERE staff_id=? AND date=? AND deleted_at IS NULL FOR UPDATE
3. 空き時間枠を再計算
4. 選択された枠がまだ空いていればINSERT、埋まっていれば409 Conflict
5. トランザクション終了

※version列は使わない。短時間のトランザクションロックで十分。
※409時のフロントエンド挙動: エラーメッセージ表示 → STEP 5（時間選択）に戻す。
```

**完了条件**:
- [ ] 各制限でエラーメッセージが返る
- [ ] SELECT FOR UPDATEによる楽観ロック
- [ ] 409 Conflict時にフロントエンドが時間選択に戻せるレスポンス形式

---

## TASK-RES-025: 指名なし委譲ロジック

**対象ファイル**: `backend/internal/service/reservation_service.go` 内

**処理**:
```
first_available モード:
  表示順上位から順に、該当時間枠が空いているスタッフを検索
  → 最初に見つかったスタッフに割り当て

top_priority モード:
  常に表示順1位のスタッフに割り当て
```

**完了条件**:
- [ ] first_available: 空きスタッフに自動割当
- [ ] top_priority: 最上位スタッフに固定割当
- [ ] `is_staff_delegated = true` がセットされる
