# LINE予約 飼主側フロー 仕様書 (LINE Reserve Owner Flow)

> スタッフ管理側（設定・予約枠・文言編集）は [28-line-reservation.md](./28-line-reservation.md) を参照。本書は飼主が LINE アプリ内で操作する予約フロー（`frontend/line-reserve` 別エントリアプリ）のみを扱う。

## 概要
- **画面の目的**: 飼主が LINE（LIFF）から診療・トリミングの予約作成、予約確認、キャンセルを完結させる。
- **URLパターン**: `/line-reserve/{clinicId}`（pathname から医院 ID を解決。単一ページアプリで内部ステップを切替え、URL は遷移しない）
- **アクセス権限**: 院内権限は不要。LINE IDトークン認証（`Authorization: Bearer` ヘッダ）を `LiffAuth` ミドルウェアが検証する。`/settings` のみ認証不要（トップ表示用・30回/分の IP レートリミット。`/my-reservations` も 30回/分）。

---

## 画面構成

### フロー全体図

```
トップ ─┬─ step1 お客様情報 → step2 コース選択 ─┬─（一般）──────────────┐
        │                                        └─（トリミング）step2b コース → step2c オプション ┘
        │        → step3 スタッフ → step4 日付 → step5 時間 → step6 ご要望 → step7 確認 → step8 完了
        └─ マイ予約（予約確認・キャンセル）←──────────────────── step8 からも遷移可
```

ステップ画面には `ProgressDots`（現在位置ドット・`aria-label` 付き）と `BackButton` が付き、戻っても入力値は `useReservationFlow` の状態に保持される。トリミング分岐（step2b/step2c 経由）では総ステップ数が一般フローより 2 つ多くなるため、`getStepProgress`（`lib/step-progress.ts`）が分岐の有無に応じた一貫した `current`/`total` を算出する（step3 以降の共有ページも含め、分岐前後で番号が後退しない）。step3 スタッフ選択の「戻る」は分岐時のみ step2c（トリミングオプション選択）へ戻る。

### 各画面の役割と入力項目

| 画面 | コンポーネント | 役割・入力項目 |
|:---|:---|:---|
| トップ | `TopPage` | ヘッダー文言（設定の header_text）表示。「新規予約」「予約確認・キャンセル」の 2 導線。電話番号設定時は tel リンクをフッター表示 |
| step1 お客様情報 | `CustomerInfoPage` | お名前・電話番号（必須）、飼い主名（任意）、ペット選択。初期値は 戻り操作の入力 → 前回予約時の入力（additional_fields）→ 紐付け済みオーナー情報 の優先順（BUG-387）。紐付け済みオーナーの登録ペットはチェックボックス選択（1頭のみなら自動選択）、新規ペットは名前+種類の自由入力で複数追加可 |
| step2 コース選択 | `CourseSelectPage` | 予約区分一覧（所要時間・紹介文・画像付き）。category が trimming の区分を選ぶと step2b へ分岐 |
| step2b トリミングコース | `TrimmingCourseSelectPage` | トリミングコース一覧（料金・説明付き）から 1 件選択 |
| step2c オプション | `TrimmingOptionSelectPage` | トリミングオプションの複数選択（任意・チェックボックス） |
| step3 スタッフ選択 | `StaffSelectPage` | コース対応スタッフ一覧。設定 show_no_staff_option 有効時は「指名なし」（staff_id=0、サーバー側で no_staff_mode に従い自動割当）を先頭表示 |
| step4 日付選択 | `DateSelectPage` | 月めくり `Calendar`。過去日・booking_window（受付期間）超過日・空きなし日は選択不可。不可日は理由（休診日/祝日休診/スタッフ不在/満席）をツールチップと `aria-label` で提示。日付未選択時は「次へ」を無効化 |
| step5 時間選択 | `TimeSelectPage` | 指定日の空き時間枠一覧。0 件時は「この日の空き時間はありません」 |
| step6 ご要望 | `RequestPage` | 自由記述テキストエリア（任意）。プレースホルダに設定のリクエスト例を表示 |
| step7 確認 | `ConfirmPage` | 入力内容の一覧表示、「まだ予約は完了していません」警告、予約時注意事項・キャンセルポリシー・プライバシーポリシー（設定文言）を表示して確定 |
| step8 完了 | `CompletePage` | 予約番号（notes 中の R-YYYYMMDD-NNNN を抽出）・日時・コース・担当を表示。マイ予約/新規予約への導線 |
| マイ予約 | `MyReservationsPage` | 自分の予約一覧（新しい順）とキャンセル操作 |
| メンテナンス | `MaintenancePage` | 受付停止中の案内（静的表示） |

---

## 主要な機能

### 1. 予約確定までのバリデーション（臨床安全・データ保護）
- **フロント**: step1 で氏名・電話番号の必須チェック（エラーは `role="alert"` で通知）。step4 は日付未選択時に進行不可。送信ボディは axios interceptor で NULL バイト除去（`sanitizeNullBytes`・R-F20）。API レスポンスは全て Zod スキーマで実行時検証（FE5-18）。
- **バックエンド**（`ValidateAndCreate`・`reservation_validators.go`）: 稼働状態 → 営業時間・受付期間・休業日（BUG-LINE-008: API 直叩き対策）→ 予約区分/トリミングマスタの医院所有権検証（クロステナント防止・hard fail）→ 医院単位 advisory lock（`AcquireBookingLock`）下で枠競合・区分定員・同日/同月件数制限をチェックし、通過時のみ INSERT。
- **入力サイズ制限**（`liff_validation.go`）: request_text 1,000 文字以内、customer_fields 全体 10KB・キー 20 個・各文字列値 500 文字以内（DoS・DB 肥大化対策）。

### 2. 枠競合時の挙動（確定時 409）
確定 POST が 409（SLOT_TAKEN / DAILY_LIMIT / MONTHLY_LIMIT / MAINTENANCE）を返すと、`ConfirmPage` がレスポンスの error 文言と redirect_step を読み取り、指定ステップ（4=日付選択、5=時間選択。不明時は 4）へ戻す。通知はブロッキングダイアログではなく画面上部のインライン赤バナー（`role="alert"`・×で閉じる）で行う（FE5-21）。

### 3. 予約完了時の連携
- 確定成功後、LINE トーク画面へ予約番号・日時・コース・ペットを含む確認メッセージを送信（`sendLiffMessage`。LINE アプリ外・失敗時は無視される best-effort）。
- サーバー側でも予約確定/キャンセル通知（飼主 LINE + 病院メール）を fire-and-forget 送信。
- 入力した氏名・電話・ペットは顧客の `additional_fields` に自動保存され、次回予約時に復元される。**氏名+電話番号による `line_customers.owner_id` 自動紐付けは行わない**（SEC-CS2-F02 / `liff_service_reservations.go`）。既に link-token またはスタッフ操作で owner が紐付いている顧客のみ、予約へ `owner_id` / `pet_id` を best-effort で反映する。

### 4. マイ予約のキャンセル可否ルール
- 一覧は本人（LINE 顧客 ID）の予約のみ。ステータスバッジ（確定/確認中/受付済/診察中/会計中/完了/未来院/キャンセル済）を色分け表示。
- **キャンセルボタンはステータスが「確定」（confirmed）の予約にのみ表示**。来院後ステータス（受付済以降）はこの画面からキャンセル不可。日時による締切はなく、当日でもキャンセル可能。
- 押下するとカード内インラインで「本当にキャンセルしますか？」の確認を挟み、確定するとバックエンドの `CancelByID` が本人・医院スコープでステータスをキャンセル済+ソフトデリートに更新（スタッフ側の予約管理から消える・Q7 仕様）。成功時は一覧上のバッジを即時「キャンセル済」に更新し、失敗時は `role="alert"` のエラーを表示する。

### 5. メンテナンス・エラー時の挙動
- 起動時に取得した設定の status が running 以外なら `MaintenancePage` を表示（フロー開始不可）。フロー途中に停止された場合は確定時に 409 MAINTENANCE で検出される。
- 医院 ID 欠落・LIFF 初期化失敗・設定取得失敗は共有 `ErrorPage` を表示。未ログイン時は LIFF ログインへ自動リダイレクト（`useLiff`）。
- 各一覧取得の失敗はページ内にステータス別メッセージ + 再試行ボタンを表示（`useFetchState` / `resolveFetchError`。401 はトークン失効のため再試行不可とし LINE アプリの再起動を案内）。

---

## 技術仕様

### 使用コンポーネント・状態管理
- **`App.tsx`（line-reserve）**: ページ状態機械。設定取得 → LIFF 初期化（`useLiff`・liff/line-reserve 共有）→ プロフィール取得 → トップ表示の 2 段初期化。
- **`useReservationFlow`**（`use-reservation-flow.ts`）: 予約フロー横断の入力状態（顧客情報・コース・スタッフ・日時・要望・トリミング選択）を単一オブジェクトで保持。完了後・新規予約開始時にリセット。URL やストレージには永続化しない。
- **`liffApi`**（`liff-api.ts`）: axios クライアント。全レスポンスを Zod 検証、書込系ボディを NULL バイトサニタイズ。
- **共有 UI（`frontend/src/shared-liff/`）**: `Spinner`・`ErrorPage`・`useFetchState`・`resolveFetchError`・JST 日付ユーティリティを LIFF 2 アプリ（liff / line-reserve）で共用。
- **`ProgressDots` / `Calendar` / `ListItem` / `PrimaryButton` / `BackButton`**: line-reserve 専用のプレゼンテーション部品。

### API連携
認証は `/settings` を除き全て LINE IDトークン（`Authorization: Bearer`）。院内 RBAC 権限は適用されない（`RegisterLiffRoutes`・`internal/reservation/routes.go`）。

| メソッド | エンドポイント | 用途 |
|:---|:---|:---|
| GET | `/api/liff/:clinicId/settings` | 公開設定の取得（認証不要・30回/分レートリミット） |
| GET | `/api/liff/:clinicId/profile` | LINE プロフィール・前回入力・紐付けオーナー/ペットの取得 |
| GET | `/api/liff/:clinicId/courses` | 予約区分（コース）一覧 |
| GET | `/api/liff/:clinicId/trimming-courses` | トリミングコース一覧 |
| GET | `/api/liff/:clinicId/trimming-options` | トリミングオプション一覧 |
| GET | `/api/liff/:clinicId/staffs` | コース対応スタッフ一覧（courseId 指定） |
| GET | `/api/liff/:clinicId/available-dates` | 予約可能日一覧（不可理由・受付期間付き） |
| GET | `/api/liff/:clinicId/available-times` | 指定日の空き時間枠一覧 |
| POST | `/api/liff/:clinicId/reservations` | 予約確定（201 + 予約番号。制限違反は 409 + redirect_step） |
| GET | `/api/liff/:clinicId/my-reservations` | 自分の予約一覧（30回/分レートリミット） |
| DELETE | `/api/liff/:clinicId/my-reservations/:id` | 予約キャンセル（204） |

---
