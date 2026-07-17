# PO-008 ファクトシート — 曽我様宛仕様書 §7 確認事項の現行実装動作

- 作成日: 2026-07-16
- 目的: q&a.html PO-008 決裁に基づき、仕様書 §7 の確認事項6項目について「現在コードが実際にどう動いているか」を file:line 根拠付きで抽出し、クライアント確認用のファクトシートとしてまとめる。ここに記載する内容は**現行実装の動作の暫定正**であり、クライアントの意図した仕様と一致するとは限らない。

---

## 7-1 既存集計3項目の仕様

対象は顧客集計 API（`GET /api/v1/clinics/:clinic_id/owners/aggregations`、
`backend/internal/handler/aggregation_handler.go:106-123`, `128`）が返す集計値。
集計元 SQL は `backend/internal/repository/ltv_repository.go` の `FindOwnerLTV`。

### (a) 年間定義・合算・CSV

**年間定義は指標ごとに異なり、統一されていない。**

- `annual_visit_count`（年間来院回数）: **直近365日ローリング**固定。パラメータで変更不可。
  ```sql
  COUNT(DISTINCT CASE WHEN mr.date >= NOW() - INTERVAL '365 days' THEN mr.date END) AS annual_visit_count
  ```
  実装根拠: `backend/internal/repository/ltv_repository.go:157`

- `annual_amount`（年間売上）: **暦年（calendar_year）または明示的な期間指定に依存し、指定なしなら全期間**。ローリング365日ではない。
  優先順位は `From`/`To`（明示日付）→ `Year`（暦年 1/1〜12/31）→ `PeriodPreset`（`last_3_months`/`last_6_months`/`last_12_months`/`calendar_year`）→ 指定なし（全期間）。
  `last_12_months` プリセットのみが「直近12ヶ月ローリング」に相当する。
  実装根拠: `backend/internal/repository/ltv_repository.go:296-339`（`calculateDateRange`）

  → **「直近365日ローリング」はq&a回答文が想定する統一仕様としては不正確**。実際は `annual_visit_count` のみローリング365日固定で、`annual_amount` は別ロジック（暦年/期間指定/全期間）。この乖離自体をクライアントに確認する必要がある。

- **合算方法**: オーナー単位（`GROUP BY o.id`, `ltv_repository.go:185`）。オーナーに紐づく全ペットの `medical_records` を `owner_id` で JOIN し（`ltv_repository.go:180`）、来院回数は `COUNT(DISTINCT mr.date)` で**日付の重複を除去して集計**する。`medical_records.date` は `type:date`（時刻を持たない日付型）であるため、同日に複数ペットが来院しても1回とカウントされる。ペット単位の内訳は本APIのレスポンスに含まれない。
  実装根拠: `backend/internal/repository/ltv_repository.go:156-157, 180, 185`（集計SQL）、`backend/internal/model/medical_record.go:20`（`Date time.Time` に `gorm:"type:date"` — DBカラムが日付型であることの根拠）

- **CSV出力**: `ListOwnerAggregation` ハンドラは `c.JSON` のみを返し、CSVエンドポイントは存在しない（`backend/internal/handler/aggregation_handler.go:106-123`）。CSV出力機能を持つのは別機能（`accounting_report_handler.go:78-80` の会計レポート、`lstep_tag_summary_handler.go:45-50` のLステップタグサマリー）で、顧客集計（本項目）には未実装。
  実装根拠: `backend/internal/handler/aggregation_handler.go:106-129`（CSV関連コードなし）

### (b) 来院回数期間

3種類の来院回数フィールドが存在し、それぞれ期間定義が異なる:

| フィールド | 期間 | 実装根拠 |
|---|---|---|
| `total_visit_count` | 全期間（制限なし） | `ltv_repository.go:156` |
| `annual_visit_count` | 直近365日ローリング固定 | `ltv_repository.go:157` |
| `period_visit_count` | `PeriodPreset`/`Year`/`From`-`To` で指定した任意期間（未指定時は全期間扱い） | `ltv_repository.go:161-162`, `ltv_repository.go:296-339` |

いずれも `COUNT(DISTINCT mr.date)`（来院日ベースの重複除去カウント）。
実装根拠: `backend/internal/repository/ltv_repository.go:156-162`

### (c) 最終来院閾値

`last_visit_bucket` は SQL の `CASE` 式で以下のように分類される（`days_since_last_visit = EXTRACT(DAY FROM NOW() - MAX(mr.date))`）:

```sql
CASE
  WHEN MAX(mr.date) IS NULL THEN 'no_visit'
  WHEN EXTRACT(DAY FROM NOW() - MAX(mr.date)) < 90  THEN 'within_3m'
  WHEN EXTRACT(DAY FROM NOW() - MAX(mr.date)) < 180 THEN 'over_3m'
  WHEN EXTRACT(DAY FROM NOW() - MAX(mr.date)) < 365 THEN 'over_6m'
  ELSE 'over_1y'
END AS last_visit_bucket
```
実装根拠: `backend/internal/repository/ltv_repository.go:164-170`（バケット名定数 `ltv_repository.go:18-24`）

これとは別に、Lステップ配信（dormant/VISIT系タグ付与）で使う休眠判定閾値は**クリニックごとに設定可能な4段階**（`DormantThresholds{Stage180, Stage210, Stage240, Stage365}`）で、DB未設定時は既定値 **180日/210日/240日/365日**（q&a回答文の「C-1確定値」に一致）。
実装根拠:
- 既定値: `backend/internal/model/dormant_thresholds.go:21-34`
- クリニック設定からの取得: `backend/internal/service/lstep_settings_thresholds.go:27-41`
- タグ判定ロジック（`dormant_180d`/`210d`/`240d`/`365d` の付与条件、240日以上で `cpm_dormant` も付与）: `backend/internal/service/lstep_tag_sync_visit_dormant.go:31-45`
- 別系統の `VISIT_*` タグ閾値（120/180/220/240日、重複付与可）: `backend/internal/service/lstep_tag_sync_visit_dormant.go:94-112`

→ 集計API側の `last_visit_bucket`（90/180/365日、3分類+未来院）と、Lステップ配信側の dormant 閾値（180/210/240/365日、4段階）は**別ロジック・別目的**であり、日数の境界値も一致しない。クライアントには両者が別物である旨の確認が必要。

---

## 7-3 Lステップ API仕様3項目

対象実装: `backend/internal/infra/lstep/tag.go`, `backend/internal/infra/lstep/user.go`（Lステップ HTTP クライアント）、
`backend/internal/service/lstep_tag_sync_api.go`, `lstep_tag_service.go`, `lstep_lifecycle_service.go`, `checkup_sync_service_create.go`（呼び出し側サービス）。

### (a) タグ付与解除

**重要な前提: Lステップ Write系 API は 2026-05-15 頃から一時停止（no-op化）されており、本ファクトシート作成時点（2026-07-16）でも停止継続中。**

- `AddTag` / `RemoveTag`（`backend/internal/infra/lstep/tag.go:18-25, 27-38`）、`SetProperty`（`backend/internal/infra/lstep/user.go:61-72`）は全て `[DISABLED]` コメント付きで実HTTP呼び出しが削除され、`lineUserID` が空文字の場合のみ `ErrUserNotFound` を返し、それ以外は常に `nil`（成功扱い）を返す no-op 実装になっている。
  `AddTagBulk`（`backend/internal/infra/lstep/tag.go:71-79`）も同様に `[DISABLED]` で実HTTP呼び出しが削除されているが、他3関数とはシグネチャ・挙動が異なる: 単一 `lineUserID` パラメータを取らず（複数件を一括処理する関数のため）、空文字チェックのコードも存在しない。**入力に関わらず無条件に `nil` を返す**（`ErrUserNotFound` を返す経路は無い）。関数コメントには「空の lineUserIDs は即座にnilを返す（APIを呼ばない）」とあるが、これは未実装のコメントのみで、実際は空でない入力でも常に `nil` を返す。
  実装根拠:
  - `AddTag`: `backend/internal/infra/lstep/tag.go:18-25`
  - `RemoveTag`: `backend/internal/infra/lstep/tag.go:27-38`
  - `AddTagBulk`: `backend/internal/infra/lstep/tag.go:71-79`
  - `SetProperty`: `backend/internal/infra/lstep/user.go:61-72`
- **Read系（`GetUserTags` / `GetUser`）は稼働継続中**で実際に HTTP GET を呼ぶ。
  実装根拠: `backend/internal/infra/lstep/tag.go:40-69`, `backend/internal/infra/lstep/user.go:25-59`
- 呼び出し側サービス（タグ同期・オーナー手動タグ操作等）は no-op を成功と区別できないため、**実際には Lステップ上のタグは一切変化していないにもかかわらず、ローカルのタグキャッシュ（`lstep_tag_cache` 等）は「付与/解除済み」として更新され続ける**。つまり現在の「タグ付与解除」機能は、UI上は成功して見えても、LINE公式アカウント側の友だちタグには反映されない状態。

### (b) エラー処理

q&a回答文の「add=fatal/remove=non-fatal」は**コード上、文脈依存で一様ではない**ため、そのまま正としない。

実際の挙動は2系統に分かれる:

1. **通常のタグ同期パス（sync/manual操作）— add・remove 双方とも fatal**:
   - `applyTagState`（desired=true で AddTag、desired=false で RemoveTag）: いずれの失敗も `notifyAPIFailure` 呼び出し後、ラップしたエラーを呼び出し元に返す（fatal）。タグキャッシュ更新（`UpsertTag`/`DeleteTag`）の失敗のみ non-fatal（ログのみ）。
     実装根拠: `backend/internal/service/lstep_tag_sync_api.go:108-133`（コメントで add/remove とも「失敗時 notifyAPIFailure + wrap されたエラーを返す」と明記）
   - `AddOwnerTag`（手動タグ付与）: AddTag 失敗は fatal。
     実装根拠: `backend/internal/service/lstep_tag_service.go:209-212`
   - `RemoveOwnerTag`（手動タグ解除）: RemoveTag 失敗は fatal（ただし `ErrUserNotFound` の場合のみ冪等成功として nil を返す）。
     実装根拠: `backend/internal/service/lstep_tag_service.go:253-259`
   - `applyCheckupTag`（健診タグ一括付与）: AddTag 失敗のみが判定対象で fatal。
     実装根拠: `backend/internal/service/checkup_sync_service_create.go:144-156`

2. **ライフサイクル系クリーンアップパス（オプトアウト・オーナー削除・ペット死亡）— remove は non-fatal（best-effort）**:
   - `removeAllTagsFromLstep`（オプトアウト時・削除時の全タグ解除）: RemoveTag 失敗はログのみで処理を継続する。
     実装根拠: `backend/internal/service/lstep_lifecycle_service.go:261-286`
   - `HandleOwnerDeletion`: 「タグ解除失敗は削除フローを止めない」と明記。
     実装根拠: `backend/internal/service/lstep_lifecycle_service.go:241-259`（コメント: 255行目）
   - `removePetDerivedTagsFromLstep`（死亡ペット由来タグ解除）: RemoveTag 失敗はログのみ（best-effort と明記）。
     実装根拠: `backend/internal/service/lstep_lifecycle_service.go:288-313`

3. **エラーカウンター補助パス（notifyAPIFailure / notifyAPISuccess）— add・remove とも non-fatal（付随処理）**:
   - `notifyAPIFailure`（連続失敗閾値到達時に `EXCL_カルテ連携エラー` タグを付与）の AddTag 失敗はログのみで呼び出し元に伝播しない。
     実装根拠: `backend/internal/service/lstep_tag_sync_api.go:135-157`
   - `notifyAPISuccess`（成功時にエラーカウンターをリセットし `EXCL_` タグを解除）の RemoveTag 失敗もログのみ。
     実装根拠: `backend/internal/service/lstep_tag_sync_api.go:163-191`

→ 「add=fatal」は主経路（1系統）でおおむね成立するが、「remove=non-fatal」は**クリーンアップ系パス（2系統）とエラーカウンター補助パス（3系統）に限られ**、通常の同期パス（1系統）では remove も fatal。q&a回答文の一様な一般化は不正確。

なお (a) の通り Write API 自体が現在 no-op のため、上記の fatal/non-fatal 分岐は**現状すべて到達不能**（AddTag/RemoveTagが常に `nil` を返すため失敗パスが発火しない）。Write API 再有効化後にのみ意味を持つ。

### (c) LINE ID紐付け

`LinkLineUserID`（`PATCH /owners/:id/line-user-id`, `backend/internal/handler/owner_handler.go:133-153`）の実装:

- **重複判定**: 同一クリニック内で同じ `line_user_id` が既に別オーナーに紐付いている場合、`FindByLineUserID` で検出し `apperrors.WrapConflict` を返す。
  実装根拠: `backend/internal/service/owner_service_line.go:17-23`
- **ステータスコード**: `WrapConflict` は `resolveErrorResponse` で `http.StatusConflict`（409）にマッピングされ、`RespondError` 経由でクライアントに返る。
  実装根拠: `backend/internal/handler/response.go:52-54`（マッピング定義）、`backend/internal/handler/owner_handler.go:148-151`（本エンドポイントが `RespondError` を使用）
- **監査ログ**: 紐付け成功時（重複なく `UpdateLineUserID` が成功した後）、`auditSvc.LogLstepOperation` で `action = "owner.line_id.link"`（解除時は `"owner.line_id.unlink"`）を記録する。監査ログ書き込みの失敗は **best-effort**（警告ログのみ、リンク操作自体はロールバックしない）。
  実装根拠: `backend/internal/service/owner_service_line.go:28-37`
- 参考: LINE ID確認操作（`ConfirmLineID`）も同様に `owner.line_id.confirm` を best-effort で監査記録する。
  実装根拠: `backend/internal/service/owner_service_line.go:40-67`

→ q&a回答文の「LINE ID紐付け=409+監査」は本項目についてはコードと一致することを確認済み。

---

## 注記

本ファクトシートは 2026-07-16 時点のリポジトリ（`main` ブランチ）に対する file:line 単位のコード実測に基づく**暫定正**である。クライアント確認の結果、意図する仕様と食い違いが判明した場合は、要件を再確認の上、必要に応じて実装（またはこのファクトシートの記載内容）を見直すこと。特に 7-1(a) の「年間」定義がフィールドごとに不統一である点、7-3(a) のLステップ Write API 停止状態、7-3(b) の fatal/non-fatal 挙動がパスごとに異なる点は、クライアントとの認識合わせが必須。
