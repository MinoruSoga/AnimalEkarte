# BE-refactor.md — バックエンド リファクタリング計画書（第6期）

- **作成日**: 2026-07-13
- **基準コミット**: `e13a2987`(main)。本書の行番号はこの時点の実測値。ずれた場合は**シンボル名(関数名・型名)で再特定**する。行番号だけを頼りに編集しない。
- **性格**: 本書は実行計画の正本。実行者はこの計画書とコードのみを前提に作業する。判断が必要な箇所はすべて本書で決着済み。判断できない事態に遭遇したら**中断して報告**する。
- **別台帳**(本書と重複させない): PERF・SEED残 = `BE_todo.md` / 任意検証 = `BE-pending.md`

---

## 1. 現状理解(実行者への文脈共有)

### 1.1 何のシステムか

動物病院向け電子カルテ(マルチテナント SaaS)。Go 1.25 / Gin / GORM / PostgreSQL 18。`backend/` は軽量 3 層構成で、**すべてのデータは `clinic_id` でテナント分離**される(本計画で clinic_id 関連の挙動を変える項目はない)。

```
cmd/api/main.go ─ DI 組み立て + cron 3 本(no-show / dormant / delivery trigger、batch_scheduler.go の runScheduled)
internal/handler    リクエスト bind(*_request.go)→ service 呼出 → レスポンス変換(*_response.go)
internal/service    業務ロジック・トランザクション境界・監査ログ・slog
internal/repository GORM クエリ。clinicScope() で clinic_id 強制。raw SQL は集計系のみ
internal/model      構造体 + ステータス定数
internal/middleware Auth(JWT+clinic 切替)ほか
internal/infra      storage / crypto / line / lstep クライアント
internal/config, dbconn, logger, apicontract, seedbundle
cmd/{migrate,seed-export,stage-import,coverage-ratchet,lstep-migrate}  運用ツール(cmd/_archive/ はビルド対象外)
```

### 1.2 規約の要点(実行者が守るべきもの)

- **P1–P18** アーキテクチャ規約が各層 `CLAUDE.md`(`backend/internal/{handler,service,repository}/CLAUDE.md`)に定義済み。着手前に 3 ファイルとも読むこと。
- **機械強制ゲート**(変更で赤にしてはならない):
  - `internal/repository/preload_clinic_scope_lint_test.go` / `dbortx_inventory_lint_test.go` / `audit_tx_inventory_lint_test.go`
  - `internal/service/master_fk_write_inventory_lint_test.go`
  - `internal/handler/handler_routes_snapshot_test.go`(第5期新設 — ルートの method+path+handler 名 golden)
  - `internal/apicontract/`(ルート↔`docs/api.yaml` 突合)
  - CI の coverage-ratchet(baseline 89.9%、−0.5pp で fail)— **テスト関数を削除する項目は本計画に無い**。モック統合(Phase F)はモック定義のみ削除し、テスト関数は残す。
- **検証コマンド規約(Docker 必須・スコープ限定)**: フル `go test ./...`・`golangci-lint run ./...`・`gofmt -w ./...` は**実行禁止**。必ず `docker compose exec backend go test ./internal/<pkg>/ -run <Name> -count=1` の形式。コンパイル確認のみ `docker compose exec backend go build ./...` は可。変更 dir は `docker compose exec backend gofmt -l ./<dir>/` 無出力を確認してからコミット。
- **git**: main 直接作業(feature branch 禁止)。**push しない**。`Co-Authored-By` を**入れない**。メッセージ形式は `refactor(backend):` / `fix(backend):` / `test(backend):` / `docs(backend):`。

### 1.3 作業ツリーの重大な注意

基準時点で以下が**未コミット(別ワークストリームの作業中ファイル)**。**変更・コミット・stash 禁止**。`git add` は常にファイル指定(`git add -A` / `git add .` 禁止)。

```
BE_todo.md / FE-refactor.md / BE-pending.md / backend/docs/api.yaml
backend/internal/handler/liff_handler_test.go
backend/internal/handler/liff_response.go
backend/internal/service/liff_service_health_card.go
frontend/liff/src/api/liff-api.ts
frontend/liff/src/pages/PetHealthPage.tsx
```

例外は本計画が明示した 2 点のみ: C-4 は `backend/docs/api.yaml` を**読み取りのみ**、F-6 は `BE_todo.md` への**純追加**(手順は各項目に明記)。

### 1.4 健全度と本計画の由来(5 期分のリファクタリング済み)

第5期(2026-07-12〜13、45 コミット)で実バグ 4 件・死コード 19 シンボル・巨大関数の第1階層を解消済み。本計画は 4 軸の新監査(①第5期コミットの回帰レビュー ②並行処理・infra・cmd ③テスト層[初監査] ④第2階層関数・API規約)の結果である。特徴:

- **可用性の実バグが残っている**: バックグラウンド goroutine に panic recovery が 1 箇所も無い(`grep -rn "recover()" internal/ cmd/` 非テスト 0 件)。`gin.Recovery()` は HTTP スタックのみを守る。
- **第5期の回帰 3 件**を検出済み(A-2/A-5/C-1 で是正)。財務 SQL・認証・監査の抽出はバイト単位比較で忠実と確認済み — 再検証は不要。
- **テスト層は全層が単一パッケージ**(`package service` ×188 / `package handler` ×205 / `package repository` ×149)。モック・ヘルパーの統合に import 変更は一切不要という構造事実を Phase F で使う。

---

## 2. 項目 0 — 安全網の構築(最初に必ず実行)

### 0-1. 作業前スナップショット

```bash
cd /Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte
git log --oneline -1          # HEAD を記録(e13a2987 のはず。違えば先行作業ありとして記録だけして続行)
git status --short            # §1.3 のリスト以外に dirty が無いことを確認。あれば中断して報告
```

### 0-2. ベースラインテストの記録

新規の特性テストは不要(テスト 589 ファイルが既存)。**着手前の合否を記録**する:

```bash
docker compose exec backend go build ./...                                    # 期待: exit 0
docker compose exec backend go test ./internal/repository/ -count=1           # 結果を記録
docker compose exec backend go test ./internal/middleware/ -count=1           # 結果を記録
docker compose exec backend go test ./internal/handler/ -count=1              # 結果を記録
docker compose exec backend go test ./internal/service/ -count=1              # 結果を記録
docker compose exec backend go test ./cmd/... -count=1                        # 結果を記録
```

失敗がある場合: §1.3 の liff 系ファイル起因のものだけを「既知の失敗」として以後の完了条件から除外。それ以外の失敗は**中断して報告**。このステップではコミットしない。

---

## 3. 作業項目リスト(実行順)

> 共通ルール: 1 項目 = 1 コミット(例外は項目内に明記)。各項目は (a) 変更 → (b) 完了条件のコマンド全 PASS → (c) `gofmt -l` 無出力 → (d) 対象ファイルのみ `git add <files>` → commit。完了条件を満たせない場合は `git checkout -- <files>` で破棄して**中断・報告**。コミット後の問題発覚は `git revert <そのコミット>` で戻す。

---

### Phase A — 実バグ・可用性修正(`fix:` コミット)

#### A-1. バックグラウンド goroutine の panic recovery 導入(プロセス全滅の防止) ✅ DONE (33458602)

- **対象**: ① `cmd/api/batch_scheduler.go:23` — `runScheduled` 内の `task(ctx)` 呼出(cron 3 本の実行点)② fire-and-forget 4 箇所: `internal/service/appointment_notification_service.go:71,120` / `internal/service/password_reset_service.go:113` / `internal/service/checkup_service.go:178`
- **問題**: 非テストコードに `recover()` が 0 件。`gin.Recovery()` は HTTP ハンドラスタックのみを守るため、cron バッチや通知 goroutine 内の 1 件の nil 逆参照で **API プロセス全体が落ちる**。
- **変更**:
  1. `internal/service/go_safe.go` を新設:
     ```go
     // goSafe は panic を回復してログする goroutine 起動ヘルパー。
     // バックグラウンド goroutine の panic はプロセスを落とすため、直接 go fn() を書かない。
     func goSafe(name string, fn func()) {
         go func() {
             defer func() {
                 if r := recover(); r != nil {
                     slog.Error("background goroutine panicked", "name", name, "panic", r, "stack", string(debug.Stack()))
                 }
             }()
             fn()
         }()
     }
     ```
  2. 上記 4 箇所の `go func() {...}()` を `goSafe("<用途名>", func() {...})` に置換(クロージャ本体は不変。用途名は "reservation notify created" 等、既存ログ文言から採る)。
  3. `runScheduled`(`batch_scheduler.go`)の `task(ctx)` を per-iteration クロージャで包み、`defer recover()` + `logger.Error(name+" panicked", "panic", r, "stack", string(debug.Stack()))` を追加(panic してもループは次の発火時刻まで継続する)。
- **テスト追加**: `cmd/api/batch_scheduler_test.go` に「task が panic してもループが継続し、次周期で再実行される」ケースを 1 件追加。
- **完了条件**: `docker compose exec backend go test ./cmd/api/ -count=1` PASS / `docker compose exec backend go test ./internal/service/ -run 'TestReservationNotification|TestPasswordReset|TestCheckup' -count=1` ベースライン比新規失敗なし。
- **リスク**: 低(防御の追加のみ・既存パスの挙動不変)。**戻し方**: revert。
- **依存**: 0。

#### A-2. `RespondError` の reflection フォールバック退行を是正(第5期回帰 #1) ✅ DONE (bb3aab93)

- **対象**: `internal/handler/response.go:80-91`(`resolveErrorResponse` の `extractCodeMessage` reflection 分岐)
- **問題**: 第5期 533e6c05 のマップ統合で、旧 `RespondError` には無かった reflection フォールバックが `RespondError` 経路にも効くようになった。exported な `Code`/`Message` string フィールドを持つ非 AppError エラー(サードパーティ SDK エラー等)が **500「internal server error」ではなく 409 + エラー自身のメッセージ**で返る — 旧実装が防いでいた情報露出クラスの退行。reflection 分岐の本来の対象は `service.ReservationLimitError` のみ(唯一の利用元 `liff_handler.go` の `RespondErrorWithExtras`)。
- **変更**(テスト先行): reflection による `extractCodeMessage` を明示の型チェックに置換:
  ```go
  // 変更前(reflection で任意の Code/Message フィールドを拾う)
  if code, msg, ok := extractCodeMessage(err); ok { ... }
  // 変更後(対象を ReservationLimitError に限定)
  var rle *service.ReservationLimitError
  if errors.As(err, &rle) { ... rle.Code / rle.Message を使用 ... }
  ```
  `extractCodeMessage` と reflection 用コードは削除(他利用者ゼロを `grep -rn "extractCodeMessage" internal/` で確認)。handler は既に `service` を import 済み。`ReservationLimitError` の実定義(ポインタ/値レシーバ)を確認して `errors.As` のターゲット型を合わせる。
- **テスト追加**: `response_test.go` に「exported Code/Message フィールドを持つ無名 struct エラー → 500 + "internal server error"」の RED ケースを先に追加。既存の `ReservationLimitError` ケース(liff の予約上限応答)が PASS のままであることを確認。
- **完了条件**: `docker compose exec backend go test ./internal/handler/ -run 'TestRespondError' -count=1` PASS + `docker compose exec backend go test ./internal/handler/ -count=1` 新規失敗なし。
- **リスク**: 低中(エラー応答経路)。**戻し方**: revert。
- **依存**: 0。

#### A-3. SMTP 実装の統合(通知側の 15 秒タイムアウトが no-op の是正) ✅ DONE (280b50b7)

- **対象**: `internal/service/appointment_notification_service.go:300`(`sendEmail(_ context.Context, ...)` — ctx を捨てて `smtp.SendMail` を :319 で呼ぶ)と `internal/service/password_reset_service.go:244-271`(ctx 対応の手動 SMTP クライアント: DialContext + deadline + validateLine)
- **問題**: 予約通知の `context.WithTimeout(…, 15*time.Second)`(:72/:121)は SMTP のダイヤル・書込に一切届かない(`net/smtp.SendMail` は期限なしの `net.Dial`)。SMTP サーバがハングすると通知 goroutine が無期限に滞留する(予約イベント毎に leak)。一方 password_reset には正しい ctx 対応実装が既にあり、**同一パッケージ内で SMTP 実装が 2 系統に分岐**している。
- **変更**: `internal/service/smtp_sender.go` を新設し、`password_reset_service.go:244-271` の ctx 対応送信ロジックを**移動**(コピーではない — password_reset 側も新ヘルパーを呼ぶよう書き換える)。シグネチャは password_reset 側の現行実装を正として決める(例: `sendSMTPMail(ctx context.Context, cfg smtpConfig, to []string, msg []byte) error`)。`appointment_notification_service.go` の `sendEmail` を新ヘルパー呼出に置換し、第 1 引数の `_` を `ctx` に戻す。**エラーメッセージ・ログは両呼出元とも現行文字列を維持**。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestPasswordReset|TestReservationNotification' -count=1` PASS(password_reset のタイムアウト実測テストが移動後も通ることが等価性の証明)。
- **リスク**: 中(メール送信経路の統合)。**戻し方**: revert。
- **依存**: A-1(同ファイルを触るため後に実施)。

#### A-4. `Validate()` に `S3_SHARED_BUCKET` チェックを追加(fail-fast の欠落) ✅ DONE (d5a84cff)

- **対象**: `internal/config/config.go:107-109`(`S3Bucket`/`S3Region` のみ検査)と `cmd/api/main.go:83-90`(`STORAGE_TYPE=s3` で `cfg.S3SharedBucket` から `NewS3FileStorage` を構築)
- **問題**: `STORAGE_TYPE=s3` かつ `S3_SHARED_BUCKET` 未設定でも起動に成功し、最初の共有ファイルアップロードで初めて失敗する — main.go の G9-2 コメントが謳う fail-fast 意図に反する。
- **変更**: `Validate()` に `if c.StorageType == "s3" && c.S3SharedBucket == "" { return fmt.Errorf("S3_SHARED_BUCKET is required when STORAGE_TYPE=s3") }` を追加(文言は既存の同系メッセージの文体に合わせる)。`config_validate_test.go` に 1 ケース追加。
- **完了条件**: `docker compose exec backend go test ./internal/config/ -count=1` PASS。
- **リスク**: 極小。**戻し方**: revert。
- **依存**: 0。

#### A-5. `lockDraftMedicalRecord` の nil fail-open 契約を硬化(第5期回帰 #3) ✅ DONE (25c2a195)

- **対象**: `internal/service/medical_record_lock.go:26-36`
- **問題**: 第5期 E-5 の統合で `parent != nil &&` ガードが全サイトに広がった。本番 repo(`medical_record_repository.go:342-352`)は `(nil, nil)` を返さないため今日は不活性だが、ヘルパーの契約が「record が nil でもエラーなしなら draft とみなして子書込を続行」という fail-open になっている。
- **変更**: nil ガードを fail-closed に反転:
  ```go
  if parent == nil {
      return nil, apperrors.WrapNotFound("medical_record", fmt.Sprintf("%d", recordID))
  }
  ```
  `(nil, nil)` を返す 2 つの prescription モック(`prescription_service_test.go` 内 — `LockByIDForUpdate` を grep で特定)を実 repo と同じ契約(NotFound エラー返却)に修正する。影響テストの期待値が変わる場合、それは fail-open 依存だった証拠なので NotFound に更新してよい(本項目の意図された修正)。
- **テスト追加**: `(nil, nil)` を返すモック locker で NotFound になるユニットテストを 1 件追加。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestPrescription|TestTreatment|TestExamination|TestVital|TestCheckupFieldResult|LockDraft' -count=1` PASS。
- **リスク**: 低(本番挙動は不変)。**戻し方**: revert。
- **依存**: 0。

---

### Phase B — シャットダウン・並行基盤

#### B-1. 通知/フォローアップ goroutine の shutdown ドレイン ✅ DONE (7c545cf8)

- **対象**: `cmd/api/main.go:199-217`(`svcs.PasswordReset.Wait()` :216 が唯一のドレイン)、`internal/service/appointment_notification_service.go:71,120`、`internal/service/checkup_service.go:178`
- **問題**: プロセス終了時、送信途中の LINE/メール通知 goroutine と checkup フォローアップが途中で殺される(PasswordReset だけが PERF-FOLLOWUP-05 の WaitGroup パターンを持つ)。
- **変更**: ① `reservationNotificationService` に `wg sync.WaitGroup` と `Wait()` を追加(`passwordResetService` と同型。起動箇所は `s.wg.Add(1); goSafe(name, func(){ defer s.wg.Done(); ... })`)② checkup フォローアップ(:178)も同様に `checkupService` へ ③ `main.go` の shutdown シーケンスに `Wait()` 呼出を `PasswordReset.Wait()` の並びで追加(interface への `Wait()` 追加要否は実定義を確認)。cron ループ(`runScheduled`)の join は**やらない**(バッチはクリニック単位逐次・次回発火で再実行 — 中断許容と判断済み)。
- **完了条件**: `docker compose exec backend go build ./...` exit 0 / `docker compose exec backend go test ./internal/service/ -run 'TestReservationNotification|TestCheckup|TestPasswordReset' -count=1` 新規失敗なし。
- **リスク**: 中(シャットダウン順序)。Wait は 30 秒の shutdown ctx に律速される構成を維持。**戻し方**: revert。
- **依存**: A-1、A-3。

#### B-2. `DetachTx` ヘルパー新設と checkup フォローアップの潜在 tx 再利用封じ ✅ DONE (7dd0be89)

- **対象**: `internal/service/checkup_service.go:177-184`(`context.WithoutCancel(ctx)`)と `internal/repository/transactor.go`(txKey の定義場所)
- **問題**: `context.WithoutCancel` は ctx の**値を全て保持**するため、ambient tx(ctx-txKey)も goroutine に持ち越される。現在 `checkupService.Create` は `WithTx` 非使用のため実害はないが、将来の監査 tx 化(#211 系の方向)が入った瞬間、goroutine 内の repo 呼出がコミット済み tx に参加してランタイムエラーになる。
- **変更**: `transactor.go` に `func DetachTx(ctx context.Context) context.Context` を追加 — `context.WithoutCancel(ctx)` の結果から txKey を外した ctx を返す(txKey は非公開のため repository パッケージ内に置く)。`checkup_service.go:177` の `context.WithoutCancel(ctx)` を `repository.DetachTx(ctx)` に置換。doc コメントに「goroutine 境界を跨ぐ WithoutCancel は tx を剥がすこと」を 1 行明記。
- **完了条件**: `docker compose exec backend go build ./...` exit 0 / `docker compose exec backend go test ./internal/repository/ -run 'TestTransactor|TestDBOrTxInventory' -count=1` PASS / `docker compose exec backend go test ./internal/service/ -run 'TestCheckup' -count=1` 新規失敗なし。
- **リスク**: 低(現挙動は等価)。**戻し方**: revert。
- **依存**: 0。

#### B-3. lstep / LINE クライアントの `http.Client` 共有化 ✅ DONE (cce3e41e)

- **対象**: `internal/infra/lstep/client.go:46-52`(`NewClient` が毎回新 Transport を生成。呼出は操作毎: `checkup_sync_service.go:160` / `lstep_delivery_trigger_client.go:51` / `lstep_tag_sync_service.go:274` / `lstep_tag_service.go:125` / `lstep_lifecycle_service.go:78`)/ `internal/infra/line/client.go:38-43`(`appointment_notification_service.go:89,138` から毎回生成)
- **問題**: 通知・同期のたびに新 `http.Client`+Transport を作るため TCP/TLS 接続が再利用されない(毎回フルハンドシェイク)。資格情報はリクエストヘッダ渡しなので、クリニック毎にクライアントを分ける必要はない。
- **変更**: 両パッケージにパッケージレベルの共有 `http.Client`(タイムアウトは現行の 15s / 10s を維持 — 差は意図的の可能性があるため揃えない)を 1 個ずつ定義し、`NewClient(apiKey, baseURL)` は**シグネチャ不変**のまま共有クライアントを参照。呼出元は無変更。
- **完了条件**: `docker compose exec backend go test ./internal/infra/... -count=1` PASS / `docker compose exec backend go test ./internal/service/ -run 'TestLstep|TestLineSend|TestReservationNotification' -count=1` 新規失敗なし。
- **リスク**: 低。**戻し方**: revert。
- **依存**: 0。

#### B-4. `logger.Default()` の遅延初期化レースを解消 ✅ DONE (aa3295ac)

- **対象**: `internal/logger/logger.go:48-53` — パッケージ変数 `defaultLogger` の非同期 read/write(`if defaultLogger == nil { Init(...) }`)
- **変更**: フォールバック初期化を `sync.Once` で包む(全エントリポイントは main で `Init` 済みのため実害は理論値だが、`-race` 検出クラスを残さない)。
- **完了条件**: `docker compose exec backend go build ./...` exit 0 / `docker compose exec backend go test ./internal/logger/ -count=1 -race` PASS(テストが無ければ Once 化の 1 ケースを追加)。
- **リスク**: なし。**戻し方**: revert。
- **依存**: 0。

---

### Phase C — 共有化・取りこぼし完了(機械的)

#### C-1. ローカル TZ RFC3339 ヘルパーの共有パッケージ化(第5期 C-6 の完了) ✅ DONE (8c1a8f2d)

- **対象**: `internal/handler/time_response.go:23-25`(`localTimeRFC3339` — handler パッケージ在住)と、変換不能で残った service 層のインライン 6 箇所: `internal/service/lab_report_query_service.go:97,134,135` / `internal/service/cash_register_service.go:281,300,301`
- **問題**: 第5期 C-6 はヘルパーを handler に置いたため service 層のクローンを変換できず、ドリフト源(TZ 直列化ポリシー変更時の見落とし)が再生した。
- **変更**: ① `internal/timeutil/format.go` を新設し `func LocalRFC3339(t time.Time) string` を定義 ② handler の `localTimeRFC3339` は `timeutil.LocalRFC3339` を呼ぶ 1 行ラッパに書き換え(handler 30 箇所の再置換はしない)③ service の 6 箇所を `timeutil.LocalRFC3339(...)` に置換 ④ 完了後に `grep -rn "In(time.Local).Format(time.RFC3339)" internal/ cmd/ | grep -v _test | grep -v timeutil` が 0 件であることを確認(残があれば同様に置換)。
- **完了条件**: 上記 grep 条件 + `docker compose exec backend go test ./internal/service/ -run 'TestLabReport|TestCashRegister' -count=1` PASS / `docker compose exec backend go test ./internal/handler/ -count=1` 新規失敗なし。
- **リスク**: 極小(出力同一)。**戻し方**: revert。
- **依存**: 0。

#### C-2. line / lstep クライアントの retry/newRequest クローン統合 ✅ DONE (ffe51a68)

- **対象**: `internal/infra/line/client.go:57-81` と `internal/infra/lstep/client.go:55-80`(同一の retry/backoff/429-drain ロジック。差分は wrap prefix のみ)、`newRequest`(line:46-54 / lstep:83-91)
- **変更**: `internal/infra/httpx/retry.go` を新設し `func DoWithRetry(ctx context.Context, client *http.Client, maxRetries int, initialWait time.Duration, do func() (*http.Response, error)) (*http.Response, error)` と bearer リクエストビルダーを抽出。両クライアントは wrap prefix・タイムアウト定数(15s/10s)を自パッケージに残したまま httpx を呼ぶ。**429 ボディの drain+close 順序は現行実装を厳密に維持**。両 `errors.go`(各10行・型が違う)は統合しない。
- **テスト追加**: `httpx` にリトライ回数・backoff・429 drain のテーブルテストを新設。
- **完了条件**: `docker compose exec backend go test ./internal/infra/... -count=1` PASS。
- **リスク**: 低。**戻し方**: revert。
- **依存**: B-3(同ファイル群 — 共有クライアント化の後に実施)。

#### C-3. `cmd/lstep-migrate` の DI 重複を共有コンストラクタで解消 ✅ DONE (3c6ba2bf)

- **対象**: `cmd/lstep-migrate/main.go:90-106`(`NewLstepTagSyncService` 14 引数の手組み)と `internal/service/service.go` 内の同一構築部
- **問題**: `LstepTagSyncService` の依存追加のたびに 2 箇所の同期が必要。同型引数の順序ミスはコンパイラで検出できない。
- **変更**: `internal/service` に `func NewLstepTagSyncFromRepos(repos *repository.Repositories, settings LstepSettingsService) LstepTagSyncService` を新設し、`NewServices` 内と `cmd/lstep-migrate/main.go` の両方から呼ぶ(引数リストは現在 `NewServices` が渡しているものを正として移す)。
- **完了条件**: `docker compose exec backend go build ./...` exit 0 / `docker compose exec backend go test ./internal/service/ -run 'TestLstepTagSync|TestSync' -count=1` 新規失敗なし。
- **リスク**: 低(コンパイル検証可能)。**戻し方**: revert。
- **依存**: 0。

#### C-4. レスポンス側 `omitempty` 4 フィールドの API 契約突合(調査 + 記録のみ)

- **対象**: レスポンス構造体の非ポインタ string + `json:",omitempty"` は全層で 4 件のみ(実測済み): `internal/handler/pet_response.go:167` `Status` / `line_customer_response.go:18` `OwnerName` / `shift_response.go:22` `StaffName` / `line_send_response.go:8` `TagAdded`
- **問題**: FE 側で「常に存在する」前提の型になっていると、空文字時のキー欠落で undefined アクセスが起きるクラス(FE 過去バグ「me omitempty crash」と同型)。
- **変更**: 4 フィールドそれぞれについて `backend/docs/api.yaml` の該当スキーマを**読み**、当該プロパティが `required` に入っていないこと(= optional 宣言)を確認する。結果を本項目の直下に表(`ファイル / フィールド / api.yaml で optional? / 判定`)で追記する。**4 件とも optional なら何も変更しない**(コミットは本書の表追記のみ、`docs(backend):`)。required に入っているものがあれば**その場で直さず中断して報告**(wire 変更 or api.yaml 修正の判断は計画外)。
- **完了条件**: 本書に 4 行の表が追記されている。コード・api.yaml は無変更。
- **リスク**: なし。**戻し方**: revert。
- **依存**: 0。

**突合結果**(2026-07-13 実施):

| ファイル | フィールド | api.yaml で optional? | 判定 |
|---|---|---|---|
| `internal/handler/pet_response.go:167` | `Status`(`status`) | `PetSummary` スキーマに `status` プロパティ有り・`required` ブロック無し | optional(一致・変更不要) |
| `internal/handler/line_customer_response.go:18` | `OwnerName`(`owner_name`) | `LineCustomer` スキーマに `owner_name` プロパティ自体が**未文書化**・`required` ブロック無し | required では無い(= 突合上は optional 扱いで問題無いが、スキーマ未記載というドリフトあり。C-4 のスコープ外のため修正しない) |
| `internal/handler/shift_response.go:22` | `StaffName`(`staff_name`) | `/shifts` の応答スキーマ `ShiftEntry` に `staff_name` プロパティ自体が**未文書化**(`ShiftEntry` は整数ID・`breaks`/`staff_name` フィールド無しで、実ワイヤ形と乖離)・`required` ブロック無し | required では無い(= 突合上は optional 扱いで問題無いが、`ShiftEntry` スキーマ自体が実レスポンス形と大きく乖離しているドリフトあり。C-4 のスコープ外のため修正しない) |
| `internal/handler/line_send_response.go:8` | `TagAdded`(`tag_added`) | `LineSendResponse` スキーマに `tag_added` プロパティ有り・`required` ブロック無し | optional(一致・変更不要) |

**結論**: 4 件とも `required` 指定なし → 本項目の完了条件どおりコード・api.yaml は無変更。ただし `owner_name`/`staff_name`（特に `ShiftEntry` 全体）は api.yaml 側にプロパティ自体が欠落しているドリフトを検出した。修正は計画外のため次期監査の候補として引き継ぐ（§5 参照)。

---

### Phase D — legacy コンストラクタ梯子の削除(第5期からの持越し・全て検証済み)

> 共通事実: 削除対象は**全て本番呼出ゼロ**(本番は `internal/service/service.go:183,235,257,292,301,317,339` で `*WithAudit` / `*WithCampaign` / `*WithAvailabilityAndType` / `*WithType` / `newPermissionGroupServiceImpl` を配線)。本番系コンストラクタの追加引数は全て nil 安全(各構造体コメントで確認済み: `medicine_service.go:172`「nil 可・後方互換」/ `treatment_service.go:157-161` / `billing_item_service.go:115-116,139,153` / `reservation_service.go:196` + `reservation_validators.go:213`)。手順は毎回同じ: レガシーコンストラクタを削除 → テスト呼出箇所を本番系 + `nil` 引数に置換 → コンパイル + scoped テスト。**テスト関数自体は削除しない**。

#### D-1. `NewMedicineService` の削除 ✅ DONE (d82db5d7)

- **対象**: `internal/service/medicine_service.go:175`。テスト呼出 6 箇所: `medicine_service_test.go:71,352` / `cross_tenant_master_fk_write_test.go:1891,1934,1971,2011`
- **変更**: 呼出を `NewMedicineServiceWithAudit(<既存引数>, nil)` に置換(実シグネチャを確認して合わせる)。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestMedicine|TestCrossTenantMasterFKWrite' -count=1` PASS。
- **リスク**: 極小。**依存**: 0。

#### D-2. `NewTreatmentService` の削除 ✅ DONE (b8aae251)

- **対象**: `internal/service/treatment_service.go:153`。テスト呼出 16 箇所(`treatment_service_test.go` ×12、`cross_tenant_master_fk_write_test.go:211,268,304,361`)
- **変更**: `NewTreatmentServiceWithAudit(<既存引数>, nil)` に置換。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestTreatment|TestCrossTenantMasterFKWrite' -count=1` PASS。
- **リスク**: 極小。**依存**: 0。

#### D-3. `NewBillingItemService` の削除 ✅ DONE (a9680376)

- **対象**: `internal/service/billing_item_service.go:128`。テスト呼出 21 箇所(`billing_item_service_test.go` ×20、`cross_tenant_master_fk_write_test.go:947`)
- **変更**: `NewBillingItemServiceWithCampaign(<既存引数>, nil, nil)` に置換(campaignRepo/ownerRepo は nil ガード済み)。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestBillingItem|TestCrossTenantMasterFKWrite' -count=1` PASS。
- **リスク**: 極小。**依存**: 0。

#### D-4. `NewReservationService` + `NewReservationServiceWithAvailability` の削除 ✅ DONE (b9e67ee2)

- **対象**: `internal/service/reservation_service.go:116,124`。テスト呼出: 旧形 18 箇所 + WithAvailability 4 箇所(`appointment_service_test.go:548,591,814,966`)。**注意**: 旧形は可変長 `staffRepo ...` で、:508,:888 は第 3 引数を渡している。
- **変更**: すべて `NewReservationServiceWithAvailabilityAndType(repo, <availabilityOrNil>, tx, <staffRepoOrNil>, nil)` に置換(引数順は実シグネチャを確認)。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestReservation|TestAppointment' -count=1` PASS。
- **リスク**: 低(22 箇所の機械置換)。**依存**: 0。

#### D-5. `NewReservationAdminService` + `WithAvailability` の削除 ✅ DONE (90dbdd22)

- **対象**: `internal/service/appointment_admin_service.go:49,57`。テスト呼出 6 箇所(`appointment_admin_service_test.go:109,148,237,321,360,408`)
- **変更**: `NewReservationAdminServiceWithAvailabilityAndType(<既存引数>, nil)` 形に置換(実シグネチャ確認)。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestReservationAdmin|TestAppointmentAdmin' -count=1` PASS。
- **リスク**: 極小。**依存**: 0。

#### D-6. `NewLiffService` の削除 ✅ DONE (4025666b)

- **対象**: `internal/service/liff_service.go:49`。テスト呼出 1 箇所(`liff_service_test.go:975` — コンストラクタ自体のテスト)
- **変更**: コンストラクタを削除し、`TestNewLiffService` を `NewLiffServiceWithType(<既存引数>, nil)` に対するテストへ書き換え(検証内容は維持)。**注意**: `liff_service_health_card.go` は §1.3 の dirty ファイル — 触らない。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestLiff' -count=1` ベースライン比新規失敗なし。
- **リスク**: 低。**依存**: 0。

#### D-7. `NewPermissionGroupService` の削除 ✅ DONE (3780399c)

- **対象**: `internal/service/permission_group_service.go:93`。テスト呼出 12 箇所(`permission_group_service_test.go`)
- **変更**: テストは同一パッケージなので `newPermissionGroupServiceImpl(repo)` 直呼びに置換し、exported コンストラクタを削除。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestPermissionGroup' -count=1` PASS。
- **リスク**: 極小。**依存**: 0。

---

### Phase E — 第2階層関数の抽出(挙動保存・純粋抽出)

> 共通ルール: 抽出前後で入出力・エラーメッセージ・ログキー・監査内容を bit 単位で維持。日本語メッセージは一字も変えない。tx 本体(原子性コメント付き)は分割しない — 各項目に「残すもの」を明記してある。

#### E-1. lstep_settings `GetSettings` の閾値マッピング抽出(105 行 → 約 45 行) ✅ DONE (23cb77fb)

- **対象**: `internal/service/lstep_settings_service.go:167-271` — うち :203-268 は `ClinicSettings` → レスポンスの純粋なフィールドマッピング(CPM v1/v2/dormant/health 閾値)
- **変更**: `func applyClinicSettingsToLstepResponse(resp *LstepSettingsResponse, cs *model.ClinicSettings)`(自由関数)として抽出。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestLstepSettings' -count=1` PASS。
- **リスク**: 極小。**依存**: 0。

#### E-2. medicine `Update` の用量設定再計算を抽出(104 行 → 約 70 行) ✅ DONE (da5019ed)

- **対象**: `internal/service/medicine_service.go:345-448` — :358-391 の eff\* マージ + `ValidateMedicineDoseConfig` 呼出(`doseFieldsChanged` ゲート含む)
- **変更**: `func validateDoseConfigAfterUpdate(existing *model.Medicine, input *UpdateMedicineInput) error` として抽出(純関数)。tx 本体(:420-439 — fields + 在庫名同期 + per_weight 監査、R1-2 原子性)は**残す**。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestMedicine' -count=1` PASS(`medicine_dose_config_test.go` 含む)。
- **リスク**: 低。**依存**: D-1(同ファイル)。

#### E-3. treatment `Update` の eff\* マージを抽出(120 行 → 約 100 行) ✅ DONE (c93bd881)

- **対象**: `internal/service/treatment_service.go:308-427` — tx 内 :378-389 の existing↔input マージ(`effItemType/effMedicineID/effQty`)
- **変更**: `func effectiveDoseInputs(existing *model.Treatment, input *UpdateTreatmentInput) (model.TreatmentItemType, *uint64, float64)` として抽出(純関数・tx 内から呼ぶ — 原子性に影響なし)。:394-401(snapshot apply/clear)の `applyDoseSnapshotFields` 抽出は**同コミット内で任意**。tx 本体(lockDraftMedicalRecord + fail-closed 監査 :372-410)は**残す**。`Create`(:195-306)は触らない(§4-7 参照)。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestTreatment' -count=1` PASS(`treatment_dose_save_test.go` 含む)。
- **リスク**: 低。**依存**: D-2(同ファイル)、A-5。

#### E-4. accounting `Update` の支払解決を抽出(128 行 → 約 100 行) ✅ DONE (c4a69c39)

- **対象**: `internal/service/accounting_service_core.go:108-235` — :134-161 の支払方法マスタ ID 解決(tx 外・読み取りのみ)
- **変更**: `func (s *accountingService) resolvePaymentWrites(ctx context.Context, input *UpdateAccountingInput) (*model.Payment, []model.PaymentSplit, error)` として抽出(`s.loadPaymentMethodSystemKeyToID` を使用。#128 コメントはヘルパーへ移す)。tx 本体(:163-217 — R1-2/X-12 コメント付きの fields+payment+監査+来院完了の原子ブロック)は**分割禁止**。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestAccounting' -count=1` PASS + `docker compose exec backend go test ./internal/repository/ -run 'TxAtomicity' -count=1` 新規失敗なし。
- **リスク**: 低中(会計)。**依存**: 0。

#### E-5. `ReplaceForCheckup` の結果構築ループを抽出(130 行 → 約 85 行) ✅ DONE (51f814d1)

- **対象**: `internal/service/checkup_field_result_service.go:115-244` — :144-188 の fieldByID map 構築 + per-input 検証 + フィールド型 switch
- **変更**: `func buildCheckupFieldResults(clinicID, checkupID uint64, fields []model.CheckupTypeField, inputs []UpsertCheckupFieldResultInput) ([]model.CheckupFieldResult, error)` として抽出(純関数。「該当列のみ書込」コメントを持たせる)。tx 本体(:197-234 — lock → snapshot → replace → fail-closed 監査、#211)は**残す**。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestCheckupFieldResult' -count=1` PASS。
- **リスク**: 低。**依存**: A-5。

#### E-6. CSV インポートの owner マッチングを抽出(126 行 → 約 85 行) ✅ DONE (13ae1157)

- **対象**: `internal/service/lstep_csv_import_service.go:63-188` — stage 6-7(:120-158)の `line_user_id` map 構築 + 行ループ → snapshots/errEntries
- **変更**: `func (s *lstepCsvImportService) matchRowsToOwners(ctx context.Context, clinicID uint64, importID uuid.UUID, dataRows [][]string, colIdx map[string]int, now time.Time) ([]*model.LstepFriendAttributeSnapshot, []csvImportErrorEntry, error)` として抽出(repo エラーはそのまま返し、`markImportFailed` は呼出元に残す)。番号付き stage コメント構造は 1 stage = 1 呼出として維持。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestLstepCsvImport|CsvImport' -count=1` PASS(integration テスト含む)。
- **リスク**: 低。**依存**: 0。

#### E-7. 健診/予防バッチの 4 クローン mapping キャッシュを統合(113 行 → 約 95 行) ✅ DONE (4fa7a45b)

- **対象**: `internal/service/lstep_health_tag_sync_batch.go:14-126` — :29-51 の fetch-mappings-warn-nil-fallback ×4(HlthHealthcheckDoneTag / PrevFilariaTag / PrevFleaTickTag / LtvFoodPurchaseTag)
- **変更**: `func (s *lstepTagSyncService) cachedTagMappings(ctx context.Context, clinicID uint64, tagName, label string) []model.LstepTagCodeMapping` として抽出(エラー時は warn ログ + nil 返却 = per-owner フォールバック維持)。`syncOwners` クロージャとカーソルループ(PERF-FOLLOWUP-02)は**残す**。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestSyncHealthPrevention|TestLstepHealthTag' -count=1` PASS。
- **リスク**: 低。**依存**: 0。

#### E-8. line_send `Send` の分割(106 行 → 約 60 行) ✅ DONE (315916bb)

- **対象**: `internal/service/line_send_service.go:80-185` — 送信・ログ・監査・タグの 4 責務
- **変更**: 純粋抽出 2 本: ① :101-130 のメッセージ型 switch → `dispatchLineMessage(ctx, lineClient <実型>, clinicID uint64, lineUserID string, input *SendLineMessageInput) (contentSummary string, validationErr, sendErr error)`(pdf/image 分岐の検証エラーは fail-fast の別経路 — 現行の分岐構造を厳密に保つため戻り値を 2 系統に分ける)② :166-182 の purpose タグ upsert → `applySendPurposeTag(ctx, clinicID, ownerID uint64, purpose string, sentAt time.Time) string`(付与タグ名を返す。"" = なし)。best-effort パスの挙動は全て維持。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestLineSend' -count=1` PASS。
- **リスク**: 低。**依存**: 0。

#### E-9. lab_result `Commit` の結果集計を抽出(124 行 → 約 90 行) ✅ DONE (ef4a051a)

- **対象**: `internal/service/lab_result_import_service.go:74-197` — :140-175 の persisted/duplicate/failed カウント + 終端ステータス決定 + `TransitionCounts` 構築
- **変更**: `func summarizeLabBatchResults(inputCount int, batchResults []LabExamPersistResult) (model.LabImportJobStatus, TransitionCounts)` として抽出(:153-157 の終端ステータス真理値表コメントをヘルパーに移し、テーブルテスト可能にする)。ジョブ遷移シーケンス(:61-73 のフローコメント)は本体に**残す**。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestLabResultImport|TestLabImport' -count=1` PASS。
- **リスク**: 低。**依存**: 0。

#### E-10. liff availability の重複クロージャ統合 + 単日閉業判定の抽出 ✅ DONE (b0236b10)

- **対象**: `internal/service/liff_service_availability.go` — ① `GetAvailableDates`(:18-122): :76-91 と :93-102 の `filterSlotsByCapacity`-with-warn クロージャがほぼ同一 ② `GetAvailableTimes`(:168-276): :193-212 の JST 変換 + 単日の閉業曜日/日付/祝日判定
- **変更**: ① capacity フィルタを 1 回だけ構築(例: `capacityFilter := s.buildCapacityFilterFn(ctx, clinicID, typeID, course)`)して両分岐で使用 — 重複 warn-and-fallback ブロックを 1 本化 ② `func isDateClosed(settings AvailableDatesSettings, dateJST time.Time) bool` を同ファイルに抽出(単日用の線形走査。**`CalcAvailableDates` の prebuilt set 方式とは統合しない** — ループ内 set は意図的設計)。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestLiff|TestGetAvailable' -count=1` ベースライン比新規失敗なし(実テスト名は `liff_service_availability_*_test.go` を grep して確定)。
- **リスク**: 低。**依存**: 0(§1.3 の `liff_service_health_card.go` には触らない)。

---

### Phase F — テスト層の負債(初監査分)

#### F-1. 死んだモックメソッド約 30 定義の削除 ✅ DONE (88e7eb0f)

- **対象**: 本番から削除済みのメソッドを今もスタブする定義群(検証済みの代表): `FindLastVisitDateByOwner` / `FindOwnersByLTVRange` / `FindOwnersByPetAgeRange`(`internal/service/examination_service_test.go:154` 付近 — この mock は 22 メソッド持つが実 interface は 19)、`SyncFilariaTag`/`SyncDormantTags`/`SyncCPMStageTagV2`/`SyncFleaTickTag`/`SyncVaccineDeadlineTag`/`SyncHealthcheckTags`/`SyncFoodPurchaseTag`/`SyncAnnual4CheckupTag`(`internal/handler/medical_record_handler_test.go:185` 付近と `internal/service/lstep_lifecycle_service_test.go:130` 付近 — 第5期 B-3 で削除された旧 API)、`BulkCreate`(`lstep_analytics_service_test.go:83`)、`UpsertPayment`(`cash_register_service_test.go:93`)、`CountByTag`(3 ファイル)、`TriggerSuppRefillReminder` / `GetDeliveryHistoryByOwner` / `HasReservationByOwnerInRange`
- **判定手順(機械的)**: 各モックメソッド名について `grep -rn "<Name>(" internal/ --include="*.go" | grep -v _test` が 0 件(= 本番に実体なし)であることを削除前に確認。0 件でないものは**削除しない**。
- **変更**: 該当メソッド定義(と、それだけを使うテスト内フィールド)を削除。テスト関数は削除しない。
- **完了条件**: `docker compose exec backend go build ./...` exit 0 / `docker compose exec backend go test ./internal/service/ -count=1` と `./internal/handler/ -count=1` ベースライン比新規失敗なし。
- **リスク**: 極小(コンパイルが正)。コミットは `test(backend):`。**依存**: 0。

#### F-2. DB テスト基盤を `ltv_repository_test.go` から専用ファイルへ移動

- **対象**: `internal/repository/ltv_repository_test.go`(1,738 行)に埋まっているパッケージ全体の基盤: `TestMain`(:1312)、`setupTestDB`(:1324)、`setupIsolatedTestDB`(:1358)、`setupSharedTestSchema`(:1604)、`sharedTestSchemaEnumTypes`(54 エントリの ENUM スキーマ複製)
- **問題**: 148 の他ファイルが依存する基盤が LTV 機能のテストファイルに同居している。
- **変更**: 新規 `internal/repository/db_setup_test.go` へ**純粋移動**(同一パッケージ・挙動不変)。prepared-statement キャッシュの根拠コメントと G12-2 パリティゲートの配線コメントは**全て一緒に移す**。
- **完了条件**: `docker compose exec backend go test ./internal/repository/ -run 'TestPreloadClinicScope|TestLtv|TestOwnerLTV' -count=1` PASS + ENUM パリティテスト PASS(実テスト名は `test_schema_enum_parity_test.go` を grep)。コミットは `test(backend):`。
- **リスク**: 極小。**依存**: 0。

#### F-3. 最大クラスタのモック統合(MedicalRecordRepository ×6 → 1)

- **対象**: `repository.MedicalRecordRepository`(19 メソッド)の 6 重複モック: `mockMedicalRecordRepository`(`medical_record_service_test.go:17` — fn-field 式・nil デフォルト)/ `mockMedicalRecordRepoForTreatment`(`treatment_service_test.go:18`)/ `mockMedicalRecordRepositoryForExam`(`examination_service_test.go:79`)/ `mockMedicalRecordRepositoryForImage`(`medical_record_image_service_test.go:24`)/ `mockMedRecordRepoForDelivery`(`lstep_delivery_trigger_service_test.go:169`)/ `mockMedRecordRepoForLstepVisit`(`lstep_tag_sync_visit_test.go:23`)
- **変更**: 新規 `internal/service/mocks_medical_record_test.go` に fn-field 式モック 1 本(`medical_record_service_test.go:17` の既存パターンを移動して正とする)を置き、他 5 変種を削除して呼出側を fn フック設定に置換。**各変種の挙動デフォルトは呼出側で明示フックに変換**(例: ForTreatment の `FindByID` は draft レコードを返すデフォルト → `treatment_service_test.go` と `treatment_dose_save_test.go:155` で明示設定)。A-5 で禁止した `(nil, nil)` 契約を再導入しない。テスト関数は削除しない。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestMedicalRecord|TestTreatment|TestExamination|TestLstepDeliveryTrigger|TestLstepTagSync' -count=1` PASS + `docker compose exec backend go test ./internal/service/ -count=1` ベースライン比新規失敗なし。コミットは `test(backend):`。
- **リスク**: 中(6 ファイルのテスト書き換え)。本項目で手法を確立する。**依存**: F-1、D-2、E-3、A-5。

#### F-4. 第2クラスタのモック統合(AccountingRepository ×4 + mockAuditService ×3)

- **対象**: ① `repository.AccountingRepository`(21 メソッド)の 4 重複(定義位置は `grep -n "type mock.*AccountingRepo" internal/service/*_test.go` で列挙)② handler 層の `mockAuditService` 3 重複: `auth_session_test.go:73` / `permission_group_handler_test.go:97` / `manual_article_handler_test.go:53`
- **変更**: F-3 と同じ手法。① は `internal/service/mocks_accounting_test.go`、② は `internal/handler/mocks_audit_test.go` に統合。**2 コミットに分ける**(service 側 / handler 側、いずれも `test(backend):`)。
- **完了条件**: ① `docker compose exec backend go test ./internal/service/ -run 'TestAccounting|TestCashRegister|TestRefund' -count=1` PASS ② `docker compose exec backend go test ./internal/handler/ -run 'TestAuth|TestPermissionGroup|TestManualArticle' -count=1` PASS。それぞれ全パッケージでベースライン比新規失敗なし。
- **リスク**: 中。**依存**: F-3(手法確立後)。

#### F-5. 通知テストの sleep-無-assert を同期 + 検証に置換

- **対象**: `internal/service/appointment_notification_service_test.go:61,79,105,123` — `NotifyCreated/NotifyCancelled` 呼出後に `time.Sleep(50ms)` して終了(**assert ゼロ** — 「panic しなかった」ことしか確認していない。goroutine がテスト終了後まで生存するレースも内包)
- **変更**: モック(`mockLineSettingRepo` 等)の fn フックで channel を close し、`select { case <-done: case <-time.After(2 * time.Second): t.Fatal(...) }` で到達を待ってから期待分岐(送信された/されなかった)を assert する。sleep は全廃。B-1 の `Wait()` を使ってもよい(どちらでも可 — 到達 assert は必須)。
- **完了条件**: `docker compose exec backend go test ./internal/service/ -run 'TestReservationNotification' -count=1 -race` PASS。コミットは `test(backend):`。
- **リスク**: 低(テストのみ)。**依存**: A-1、A-3、B-1。

#### F-6. 恒久 skip された既知バグテスト 17 件の台帳化(コード変更なし)

- **対象**: `t.Skip("known production bug — …")` の 17 箇所(代表: `trimming_repository_test.go:177,192,202` / `hospitalization_repository_test.go:221,296` / `medicine_repository_test.go:301,308` / `clinic_repository_test.go:284,337` / `clinic_settings_repository_test.go:41` / `pet_chronic_condition_repository_test.go:209` / `hospitalization_plan_repository_test.go:229` / `payment_method_master_repository_test.go:258`。全数は `grep -rn 't.Skip("known production bug' internal/` で列挙)
- **問題**: skip されたテストはバグが直っても・悪化しても永遠に fail しない。issue への紐付けも無く、実行可能なバグ仕様が静かに腐る。
- **変更**: 全件を grep で列挙し、`BE_todo.md` の末尾に新セクション「### 既知バグの skip テスト台帳(2026-07-13 棚卸し)」として `ファイル:行 / テスト名 / skip メッセージ要約` の表を追記する。**テストコード自体は変更しない**(skip 解除・expected-failure 化は各バグ修正時の判断)。**注意**: `BE_todo.md` は §1.3 の dirty ファイル — 追記前に `git diff BE_todo.md` で既存の未コミット差分を確認し、自分の追記だけを含むコミットが作れない(既存差分と分離不能な)場合は**中断して報告**。
- **完了条件**: `BE_todo.md` に表が追記されている(件数は grep 実測に従う)。コミットは `docs(backend):`。
- **リスク**: 低。**依存**: 0。

---

## 4. やらないことリスト(善意の逸脱の先回り禁止)

1. **機能追加・仕様変更・API 変更禁止**: ルート増減・`docs/api.yaml` 変更・レスポンス JSON のキー/値/形式の変更は禁止。例外は本計画に明示した A-2(未分類カスタムエラー 409→500 — 情報露出退行の是正)のみ。
2. **依存ライブラリ更新禁止**: `go.mod` / `go.sum` に触らない。
3. **PO 見送り済み事項に着手しない**(2026-07-12 決定・git 履歴の第5期計画 §4 が正本):
   - `SharedFileService.CleanupExpired` / `TokenBlacklistService.DeleteExpired` の cron 配線(PO「7日は短すぎ・削除しなくていい」)
   - lstep バッチ 3 本(`RunLTVTopPercentSyncAllClinics` 等)の cron 配線または削除(Lstep Write API 再開判断とセット)
   - どちらも「10 行で書けてしまう」が**書くな**。
4. **`BE-pending.md` の項目に着手しない**: STG クロステナント監査 SQL は人間実行のみ。
5. **migration / seed / DB スキーマに触らない**: `backend/migrations/` 配下は一切変更しない。
6. **§1.3 の dirty ファイルに触らない**(liff 系 3 ファイル + api.yaml + 台帳 md)。例外は C-4 の api.yaml **読み取り**と F-6 の BE_todo.md **純追加**のみ。
7. **検討済み・意図的に不採択(再提案・実施禁止)**:
   - `NewServices` / `NewRepositories` / `RegisterMasterRoutes` の分割・table 化(第5期決定の維持)
   - count+find ページネーション抽象化・手書き複合 `Where("clinic_id = ?")` の clinicScope 変換(受容済み・repository CLAUDE.md に文書化済み)
   - `CalcAvailableDates`(112 行)の分割 — フラットな early-continue ガード連鎖で構造的に正当
   - `SyncLTVTopPercent`(108 行)の分割 — PERF-FOLLOWUP-08 直後・go-reviewer 承認済み・根拠コメント密度が高い
   - `GetMonthlyReportByPeriod`(143 行)の再分割 — 第5期抽出後の残りは SQL リテラル + 線形組み立てで正当
   - treatment `Create` の分割 — tx 本体が原子性コメント密。リテラル抽出は E-3 実施時の任意事項に留める
   - `cmd/migrate/main.go`(538 行)の分割 — 内部構造は検証済みで健全(per-migration tx・checksum・advisory lock)。長さは正当
   - setupXxxTestDB 118 本の自動 TRUNCATE 導出への統合 — TRUNCATE リストはテスト間隔離の契約。一括変換は隔離前提を壊すリスク。機会的移行のみ(今期対象外)
   - fixture maker 135 本の共有 fixtures_test.go への big-bang 統合 — 暗黙デフォルト共有は実証済みの脆弱化クラス(CPM テスト)。機会的移行のみ
   - handler 401/400/500 ボイラープレート約 75 箇所の一括ヘルパー化 — 価値が薄い。機会的で足りる
   - `t.Parallel()` の導入 — 共有 DB プール + TRUNCATE 隔離では正当性バグになる(repository コメントに明記)
   - テーブルテスト命名(`tests/tt` vs `cases/tc` 40 箇所)の正規化 — 価値なし
   - permission_group / manual_article の監査ログ処理の service 層移動(引き続き次期以降・監査 tx 設計との整合が前提)
   - line/lstep の `errors.go` 2 枚の統合(各10行・型が違う)
   - 第5期 E-3 で入った refresh 時の JTI 失効順序変更 — 実害なしと判定済み。戻さない
   - lstep/line クライアントのタイムアウト統一(15s/10s)— 差は意図的の可能性。揃えない
8. **テストの期待値を変えて通さない**。例外は A-5 に明記した 1 点(fail-open 依存だったテストの NotFound 化)のみ。
9. **リファクタついでの「改善」禁止**: 明示箇所以外のログ追加・エラーメッセージ文言統一・コメント書き直し・変数名変更は禁止。計画にある変更だけを行う。
10. **push しない・PR を作らない**。すべてローカル main へのコミットのみ。
11. **coverage-ratchet の baseline(`backend/.coverage-baseline`)を変更しない**(CI 実測でのみ更新する運用)。

---

## 5. 次期監査への引き継ぎ(実行者は無視してよい)

今回の監査で**未検証のまま残った領域**(次回の監査レンズ候補): ① service interface のオーファンメソッド全数スイープ(第6期の削除後に再走査)② slog レベルの誤用(Error を期待条件に使う箇所)③ helper 経由の退化ページネーション封筒(構造体リテラル直書きは 0 件と確認済みだが helper 経由は未走査)④ `backend/worker/`(Cloudflare Workers エントリ・TypeScript 358 行・テスト付き)— 第5期・第6期とも専任監査を当てていない。小規模かつ 2026-07-10〜12 の infra 作業で直近整備済みのため優先度は低いが、Go 3 層の監査スコープ外であることをここに明記する。

---

## 6. 実行者への指示文(このままコピペして渡す)

```
あなたは AnimalEkarte リポジトリ(/Users/minoru/Dev/Case/AnimalHospital/AnimalEkarte)の
バックエンドリファクタリング実行者です。BE-refactor.md が唯一の作業指示書です。

手順:
1. まず BE-refactor.md を全文読む。次に backend/CLAUDE.md と
   backend/internal/{handler,service,repository}/CLAUDE.md を読む。
2. 計画書 §2(項目 0)を最初に実行し、ベースラインを記録する。
   §1.3 のリスト外の dirty ファイルがあれば中断して報告。
3. 作業項目を A-1 から F-6 まで記載順に、1 項目ずつ実施する。
   各項目の「依存」に記載された項目が完了済みであることを着手前に確認する。
4. 1 項目 = 1 コミット(F-4 のみ 2 コミットと明記)。コミット前に必ず:
   - その項目の「完了条件」のコマンドをすべて実行し、期待結果を得る
   - docker compose exec backend gofmt -l ./<変更したdir>/ が無出力
   - git add はファイル指定のみ(git add -A / git add . 禁止)
   - コミットメッセージは refactor(backend):/fix(backend):/test(backend):/docs(backend): 形式、
     Co-Authored-By を含めない
5. 完了条件を満たせない場合: 変更を git checkout -- <files> で破棄し、
   何をしてどう失敗したかを記録して中断・報告する。勝手に別解を試みない。
6. 計画書の行番号は基準コミット e13a2987 時点の値。ずれていたらシンボル名
   (関数名・型名)で特定する。シンボルが見つからない場合は中断して報告。
7. 禁止事項: §4「やらないことリスト」を厳守。特に PO 見送り済みの cron 配線 2 件は
   絶対に書かない。push 禁止。フル go test ./... / golangci-lint run ./... /
   gofmt -w ./... 禁止。テストは常に docker compose exec backend go test
   ./internal/<pkg>/ 形式(cmd 配下は ./cmd/<name>/)で実行する。
8. 全項目完了後: git log --oneline で全コミットを列挙し、
   実施した項目 ID・スキップ/中断した項目 ID とその理由を報告する。
```

---

## 付録: 実行順トレース検証(作成時に実施済み)

- **appointment_notification_service.go**: A-1(goSafe 置換)→ A-3(sendEmail 統合)→ B-1(wg 追加)→ F-5(テスト同期化)の直列順。各段はシンボル特定で行番号ずれを吸収する。
- **checkup_service.go**: A-1(goSafe)→ B-1(wg)→ B-2(DetachTx)。B-2 は `TriggerCheckupFollowUp` 呼出部をシンボルで特定。
- **medical_record_lock.go と prescription/treatment 系**: A-5(nil 硬化 + prescription モック修正)→ E-3(treatment 抽出)→ F-3(MedicalRecord モック統合)。F-3 は A-5 の契約((nil,nil) 禁止)を前提に統合 — 依存に明記済み。
- **medicine / treatment / billing / reservation 系**: D-1〜D-5(コンストラクタ削除)→ E-2/E-3(同ファイルの関数抽出)→ F-3/F-4(モック統合)。テストファイルを複数回触る設計だが、各段の完了条件が独立に green を要求するため前提破壊はない。
- **infra/lstep・infra/line**: B-3(共有 client)→ C-2(httpx 抽出)。C-2 は B-3 後の姿に対して実施 — 依存に明記済み。
- **response.go**: A-2 のみが触る。ルート変更なしのため第5期 golden(ルートスナップショット)に影響しない。
- **F-1(死にモック削除)→ F-3/F-4(モック統合)**: 統合対象の変種に死にメソッドが残っていると統合時に混入するため F-1 が先。
- **C-4 / F-6(dirty ファイル例外)**: 読み取りのみ(C-4)と純追加 + 分離不能時中断(F-6)に限定 — 手順を項目内に明記済み。
- **coverage-ratchet**: テスト関数を削除する項目が無いことを全項目で確認済み(F-1/F-3/F-4 はモック定義のみ削除)。
