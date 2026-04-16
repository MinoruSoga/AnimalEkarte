# BUG-346: ファイル行数 500行超過（3ファイル）

## 概要

`backend/CLAUDE.md` に「パッケージ1ファイル < 500行」の規約がある。非テストファイルで 3 ファイルが超過している。

## 対象ファイル

| ファイル | 行数 | 超過 | 分割候補 |
|---------|------|------|---------|
| `backend/internal/handler/auth_handler.go` | 617 | +117 | 認証フロー / ヘルパー関数 / レスポンス構築 |
| `backend/internal/handler/staff_handler.go` | 609 | +109 | CRUD / クリニック配属 / パーミッショングループ / 予約区分 |
| `backend/internal/service/liff_service.go` | 555 | +55 | 予約可能日時計算 / スタッフ解決 / 自動リンク |

## 現状コード

### `backend/internal/handler/auth_handler.go`（617行）

関数一覧:
```
:27  buildMeResponse()            — MeResponse 構築ヘルパー（71行）
:98  authenticateUser()           — 認証ロジック（35行）
:133 resolveClinicInfo()          — クリニック情報解決（16行）
:149 issueAuthCookies()           — Cookie 発行（65行）
:214 Login()                      — ログインハンドラ（57行）
:271 Logout()                     — ログアウトハンドラ（63行）
:334 RefreshToken()               — トークンリフレッシュ（121行）
:455 ChangeMyPassword()           — パスワード変更（59行）
:514 GetMe()                      — 現在ユーザー取得（69行）
:583 buildAllPermissions()        — 全権限構築（11行）
:594 calculateEffectivePermissions() — 実効権限計算（23行）
```

### `backend/internal/handler/staff_handler.go`（609行）

関数一覧:
```
:20  ListStaffs()                     — スタッフ一覧
:37  CreateStaff()                    — 作成（88行）
:125 UpdateStaff()                    — 更新（74行）
:199 GetStaff()                       — 取得
:220 DeleteStaff()                    — 削除
:239 GetStaffPermissionGroups()       — パーミッショングループ取得
:261 SetStaffPermissionGroups()       — パーミッショングループ設定
:292 GetStaffClinicAssignments()      — クリニック配属取得
:318 SetStaffClinicAssignments()      — クリニック配属設定
:351 GetStaffExcludedReservationTypes()  — 除外予約区分取得
:373 SetStaffExcludedReservationTypes()  — 除外予約区分設定
:403 ReorderStaffs()                  — 並び替え
:421 RegisterMasterRoutes()           — ルート登録（長大）
```

### `backend/internal/service/liff_service.go`（555行）

関数一覧:
```
:67  GetSettings()             — 設定取得
:76  GetProfile()              — プロファイル取得
:85  GetCourses()              — コース一覧
:100 GetStaffs()               — スタッフ一覧
:123 GetAvailableDates()       — 予約可能日取得（51行）
:174 GetAvailableTimes()       — 予約可能時間取得（69行）
:243 CreateReservation()       — 予約作成（50行）
:293 GetMyReservations()       — 予約一覧
:302 CancelReservation()       — キャンセル（27行）
:329 resolveTargetStaffs()     — スタッフ解決（33行）
:362 buildStaffSlotInputs()    — スロット入力構築（50行）
:412 parseBusinessHoursForDate() — 営業時間解析（26行）
:438 delegateStaff()           — スタッフ委任（35行）
:473 isStaffAvailable()        — スタッフ空き確認（19行）
:492 isExcluded()              — 除外確認（12行）
:504 tryAutoLinkOwner()        — オーナー自動リンク（長大）
```

## 修正方針

各ファイルを論理的な責務単位で分割する。

### 1. `auth_handler.go` → 2ファイルに分割

```
auth_handler.go          ← Login / Logout / RefreshToken / ChangeMyPassword / GetMe（公開ハンドラのみ）
auth_handler_helpers.go  ← buildMeResponse / authenticateUser / resolveClinicInfo / 
                           issueAuthCookies / buildAllPermissions / calculateEffectivePermissions
```

### 2. `staff_handler.go` → 2ファイルに分割

```
staff_handler.go              ← ListStaffs / CreateStaff / UpdateStaff / GetStaff / 
                                 DeleteStaff / ReorderStaffs
staff_assignment_handler.go   ← GetStaffPermissionGroups / SetStaffPermissionGroups /
                                 GetStaffClinicAssignments / SetStaffClinicAssignments /
                                 GetStaffExcludedReservationTypes / SetStaffExcludedReservationTypes +
                                 RegisterMasterRoutes の配属系ルート
```

### 3. `liff_service.go` → 2ファイルに分割

```
liff_service.go              ← GetSettings / GetProfile / GetCourses / GetStaffs / 
                                GetMyReservations / CancelReservation / CreateReservation
liff_availability_service.go ← GetAvailableDates / GetAvailableTimes / resolveTargetStaffs /
                                buildStaffSlotInputs / parseBusinessHoursForDate / 
                                delegateStaff / isStaffAvailable / isExcluded
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `auth_handler.go` | ヘルパー関数の移動のみ。外部 API に変更なし | 未対応 |
| `staff_handler.go` | ハンドラメソッドの移動のみ。同一パッケージ内なのでシグネチャ変更不要 | 未対応 |
| `liff_service.go` | `LiffService` インターフェース実装の分割。`NewLiffService` で全フィールドを共有するため struct は維持 | 未対応 |

## 準拠すべきプロジェクト規約

### `backend/CLAUDE.md` — ファイルサイズ規約
> パッケージ1ファイル < 500行

### プロジェクト内参照実装
- `backend/internal/handler/diagnosis_handler.go`（295行）— 適切なサイズの参照例
- `backend/internal/service/medical_record_service.go`（337行）— 適切なサイズの参照例

## 優先度

**Low** — 機能的な問題はない。コードの読みやすさ・保守性の改善。他の Critical/High が全て解消されているため次の余裕があるタイミングで対応。

## 関連チケット

- BUG-345: 第8回 Backend Go Convention Audit（親チケット）

## 関連ファイル

- `backend/internal/handler/auth_handler.go:1-617` — 617行
- `backend/internal/handler/staff_handler.go:1-609` — 609行
- `backend/internal/service/liff_service.go:1-555` — 555行
