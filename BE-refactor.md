# バックエンド リファクタリング計画

- **作成日**: 2026-07-02
- **対象**: `backend/`（Go 1.25 / Gin / GORM / handler→service→repository 軽量レイヤリング）
- **性格**: 全項目 **behavior-preserving（挙動保存）** を原則とする負債返済計画。振る舞いを変える修正（bug.md の H-1/H-3/M-1/M-2 等）は**バグ修正トラックであり本計画の対象外**（相互参照のみ）。
- **根拠**: 2026-06-30 の系統リファクタ Phase 0 で「コードベースは既に良好・全面リファクタ不要」と判定済み。本計画はその判定を本日の実測で再確認した上で、**残存している構造的負債に的を絞る**。推測ベースの投機的リファクタは行わない（プロジェクト方針）。

---

## 1. 現状評価（2026-07-02 実測）

### 健全な点（リファクタ不要と判断する根拠）

| 観点 | 実測値 | 評価 |
|---|---|---|
| 最大ファイルサイズ | `service/treatment_service.go` 645行（上限800行内） | 分割不要。上位15ファイル全て800行未満 |
| TODO/FIXME/HACK | 2件のみ（`liff_service_availability.go:62,73` の documented N+1） | 放置負債なし |
| レイヤリング | handler(257src)→service(195src)→repository(104src) 一方向 | P1-P18 準拠。アーキ変更不要 |
| ガードレール | preload clinic_id lint（go/ast）・audit-tx inventory lint・clinic分離テスト群が CI 常時実行 | 3監査クラスタの再発防止が機械化済み |
| クロステナント | read IDOR 13repo / write FK 6+5service 修正済み（6/29-30 監査完遂） | 対応済み |

### 残存する構造的負債（本計画の対象）

| # | 負債 | 証拠（現HEADで確認済み） | リスク |
|---|---|---|---|
| D1 | 監査ログ書込が tx 外 best-effort のまま残る書込経路 6箇所 | `grep auditSvc.LogEntry(` service層 6件（LogEntryTx 移行済みは refund / checkup_field_result のみ） | 本体成功・監査失敗で証跡欠落（医療記録系で特に問題） |
| D2 | WithTx 内から呼ばれる読取が tx 非参加（TOCTOU） | `refund_repository.go:50-75` SumByBillingID / SumByBillingIDAndPaymentMethod、`accounting_repository.go:173` LockAndFindByID — いずれも `r.db.WithContext(ctx)` 直参照で dbOrTx 未使用 | FOR UPDATE ロックと集計読取が別コネクションになり、返金上限チェック等の read-modify-write が並行時に破れる |
| D3 | 日付 tz 表現割れの残り3フィールド | FirstVisit / inventory ExpiryDate / cash_register PaidAt（6/30 の localTimePtr 統一時の明示的積み残し） | 同一 API 内で `Z` と `+09:00` が混在し FE 側の日付解釈がフィールド毎に分岐 |
| D4 | 孤児コード: `GET /owners/:id/medication-history` | `handler.go:149` + service + repository + テスト一式が FE から呼び出しゼロ（bug.md M-6） | デッドコードの保守コスト・監査ノイズ |
| D5 | permissionGroup.UpdateRules の二重化 | 6/29 write 監査時の積み残し（P1） | 同一処理の重複実装ドリフト |
| D6 | repository 層のテスト希薄 | src 104 に対し test 31（handler 257/151・service 195/115 と比べ最弱） | #212 でも repository 11.8% 起点。回帰検出力が最も低い層が DB 直結層 |
| D7 | auth ライフサイクルテスト不在 | password_reset_token / token_blacklist の repository テスト 0件（bug.md M-8・確認済み） | 認証セキュリティ不変条件が未検証 |
| D8 | LIFF 空き日付一覧の N+1 | `liff_service_availability.go:62,73`（documented TODO） | 日数×capacity クエリ。LINE 導線の応答劣化 |
| D9 | ガードレール未整備分 | CASCADE 検出 lint（P0）・preload lint の新 master 自動登録（P1）・openapi format↔シリアライズ突合 CI（P1）・repo テスト pool 枯渇（P1） | 過去監査で確立した不変条件の一部が人力レビュー頼み |
| D10 | fail-open 残存疑い（F-2: 休憩時間チェック） | 6/30 second-lens 監査の未対応報告。今回の grep では発生箇所を特定できず**要再調査** | エラー時に検証スキップで予約が通る可能性 |
| D11 | golangci-lint の issue 数 cap が設定に残存 | `backend/.golangci.yml:140-141` `max-issues-per-linter: 50` / `max-same-issues: 10`（現HEADで確認済み）。6/30 に「cap で11件目以降が隠れる」事故を経験し教訓化したが**設定自体は未修正** | 同種 lint issue が閾値を超えると CI green のまま隠蔽される |
| D12 | count/junction クエリの clinic スコープ未統一（6/29 read 監査の MEDIUM 残・report-only） | junction 3件: `permission_group.FindAllGroupIDsByStaffID` / `reservation_staff.FindAllExcludedReservationTypes` / `...ByStaffIDs`。count 3件: `medical_record.CountEstimatesByMedicalRecordID` / `reservation.CountMedicalRecordsByReservationID` / `estimate.CountItemsByEstimateID` | 実害は低い（不透明ID/整数のみ・handler 側 `verifyStaffClinicMembership` 等の上位防御あり）が、「repository は clinic スコープ必須」規約の例外が無追跡で残る |
| D13 | 臨床結果テーブルの DB レベル複合 FK 不在（#211 follow-up） | exam_results / checkup_field_results の clinic 整合はアプリ層検証のみで、DB レベルの複合 FK（clinic_id 込み）が無い | アプリ層のバグ・直接 SQL 操作でクロステナント FK が物理的に入り得る（defense-in-depth の欠落） |

**方針**: D1/D2 が金銭・監査証跡に直結するため最優先。D3-D5・D12 は表現統一と削減。D6-D9・D11・D13 は再発防止の機械化。D10 は調査から。

---

## 2. フェーズ計画

規模: S=半日以内 / M=1日 / L=2-3日。各項目は独立コミット。着手順はフェーズ内で上から。

### Phase 1: トランザクション整合性の構造統一（安全性・最優先）

#### R1-1. 読取系 repository メソッドの dbOrTx 統一（D2）— 規模 M

- **現状**: `WithTx` 内から呼ばれる `LockAndFindByID`（accounting_repository.go:173、reservation_repository.go:230）、`SumByBillingID` / `SumByBillingIDAndPaymentMethod`（refund_repository.go:50,62）が `r.db.WithContext(ctx)` 直参照。呼び出し元（refund_service.go:48,60,84 等）は txCtx を渡しているのに、読取が tx 外の別コネクションで実行される。**FOR UPDATE が tx を保護しない**。
- **あるべき姿**: 書込系で確立済みの `dbOrTx(ctx, r.db)` パターンに読取系も統一。ambient tx があれば参加、なければ従来どおり。
- **手順**:
  1. `grep -rn "r.db.WithContext" backend/internal/repository/` で全読取メソッドを棚卸しし、「WithTx 内から呼ばれるもの」を呼び出し元遡及で特定する（Lock系・Sum系・check系が対象。一覧系 FindAll 等は tx 外呼び出しのみなら対象外）。
  2. 対象メソッドを `dbOrTx(ctx, r.db).WithContext(ctx)` へ置換。シグネチャ変更なし（ctx 経由で tx 伝播する既存機構をそのまま使う）。
  3. **挙動保存の証明**: 既存テスト GREEN 維持。加えて #211 refund で確立した DB-backed atomicity テストパターンを流用し、「WithTx 内の SumByBillingID が tx 内の未コミット挿入を読める」ことを検証するテストを追加（これが tx 参加の直接証明になる）。
- **検証**: `docker compose exec backend go test ./internal/repository/... ./internal/service/... -run 'Refund|LockAndFind'`
- **注意**: `dbOrTx` 化により Lock系がネスト tx で呼ばれた場合の挙動（GORM の SavePoint）に差が出ないか、reservation 側の呼び出し元も確認する。

#### R1-2. 監査ログ書込の LogEntryTx 横断展開（D1）— 規模 M

- **現状**: `AuditTxLogger` / `LogEntryTx`（audit_service.go）は導入済みで、refund と checkup_field_result のみ移行完了。service 層に非 Tx の `auditSvc.LogEntry(` が **6箇所**残存。うち tx を持つ書込経路（本体書込と監査が原子であるべきもの）が移行対象。
- **あるべき姿**: 「臨床・会計データの書込 tx 内で監査も書く（fail-closed）」を全書込経路に適用。tx を持たない読取系操作のログ（best-effort で許容されるもの）は現状維持し、audit-tx inventory lint の allowlist にその根拠を明記。
- **手順**:
  1. 6箇所を列挙し、「tx あり（移行必須）/ tx なし（allowlist 根拠明記）」に分類する。#211 inventory lint の分類（exam_results=pending-migration）に従い exam_results を最優先で移行。
  2. 移行は refund（6f432912）・checkup_field_result（c9028a18 系）の確立パターンを踏襲: `WithTx` 内で `LogEntryTx(txCtx, ...)`、失敗時は tx ごと rollback（fail-closed）。
  3. 各移行に temp-revert RED（監査書込失敗注入で本体も rollback されること）→ GREEN の実証を付ける（#211 で確立した検証手法）。
  4. inventory lint の allowlist から移行済み項目を削除し、lint が「未移行 0件（または根拠付き allowlist のみ）」を強制する状態に収束させる。
- **検証**: `docker compose exec backend go test ./internal/service/... -run 'Audit|Tx'` + inventory lint（CI unconditional job）
- **対象外**: care_plan_items は既存の別 Issue 管轄。

#### R1-3. fail-open 経路の再調査と fail-closed 化（D10）— 規模 S(調査)+S(修正)

- **現状**: 6/30 監査で F-2「休憩時間チェックの fail-open」が未対応と報告されたが、今回の grep では該当箇所を特定できず。修正済みの可能性もある。
- **手順**:
  1. **調査から入る**（憶測で直さない）: `reservation_validators.go:225` の validateBusinessRules と休憩時間取得経路（shift/schedule repository）を読み、「取得エラー時に検証をスキップして予約を通す」分岐が残っているか確定させる。
  2. 残っていれば F-3 で確立した fail-closed パターン（エラー時は拒否＋失敗注入 RED→GREEN）で是正。**これは挙動変更**（エラー時に予約が通らなくなる）なので、bug 修正としてコミットメッセージを fix: にする。
  3. 併せて横断 grep: エラー swallow 型の検証スキップ（`if err != nil` で握りつぶして検証を通す形）を service 層で棚卸しし、同型があれば同一 PR で列挙（修正は個別判断）。
- **検証**: 失敗注入テスト＋`go test ./internal/service/... -run 'Reservation|Validate'`

### Phase 2: 表現統一とデッドコード削減

#### R2-1. 日付 tz 表現の localTimePtr 統一・残り3フィールド（D3）— 規模 S

- **現状**: 6/30 に pet の birth_date/neutered_date/last_visit は `localTimePtr`（ParseInLocation JST canonical）へ統一済み（c0e32421/5c5d43f4）。**FirstVisit / inventory ExpiryDate / cash_register PaidAt が未統一**のまま明示的に積み残されている。
- **手順**: 各 response builder で raw time → `localTimePtr` 化。既存の pet 対応コミットをテンプレートにする。openapi の format 宣言（date-time）は整合済み（6/30 調査で 0/76 ミスマッチ確認）なので API 定義変更は不要。FE 側で `Z` / `+09:00` の差に依存した処理がないことを grep で確認してから入れる。
- **検証**: 対象 handler の response テスト（シリアライズ形式 pin）＋ `npx vitest run` の該当 feature（FE で日付表示に使う画面のみ）
- **完了条件**: API 応答の日付系フィールドが全て同一 tz 表現。format↔シリアライズ突合 CI（R3-3）が入ればここで pin される。

#### R2-2. 孤児エンドポイント medication-history の処分（D4）— 規模 S

- **現状**: `GET /owners/:id/medication-history`（handler.go:149・service medical_record_report.go・repository FindOwnerMedicationHistory・テスト一式）が FE 呼び出しゼロ（bug.md M-6・確認済み）。
- **手順**: **削除を推奨**（YAGNI。飼主横断投薬歴ビューの構想が生きているかを先に PO へ1問確認し、回答が No または不明なら削除）。ルート・handler・service・repository・テスト・openapi 定義を一括除去し `make codegen` 同期。残す判断になった場合はコメントで将来用途を明記して本項クローズ。
- **検証**: `grep -rn medication-history backend/` が空 + scoped go test + FE ビルドに影響なし（呼び出しゼロなので型生成差分のみ）。

#### R2-3. permissionGroup.UpdateRules 二重化解消（D5）— 規模 S

- **現状**: 6/29 write 監査（72e8887c）時に「UpdateRules の二重化」が P1 積み残しとして記録されている。
- **手順**: 現状確認から（該当コードを読み、二重実装が本当に残っているか確定）→ 残っていれば単一の実装へ統合し、呼び出し元を寄せる。権限系は F-3（自己参照 bypass）の前歴があるため、統合後に既存の権限テストが GREEN であることに加え、fail-closed 挙動（権限取得失敗時に拒否）が保存されていることを明示的に確認する。
- **検証**: `docker compose exec backend go test ./internal/... -run 'Permission'`

#### R2-4. LIFF 空き日付一覧の N+1 解消（D8）— 規模 M

- **現状**: `liff_service_availability.go:62,73` — 日付一覧取得時の capacity チェックが日付ごとにクエリ発行（documented TODO）。
- **手順**: 対象期間の予約数/capacity を日付範囲でバッチ取得する repository メソッド（GROUP BY date）を追加し、ループ内クエリを置換。**応答の内容は不変**（並び・件数・判定結果を table-driven テストで pin してから置換）。clinic_id スコープは既存の preload lint / isolation テスト規約に従う。
- **検証**: 既存 LIFF テスト GREEN + 追加した pin テスト + （可能なら）クエリ数の before/after をテストログで確認。

#### R2-5. count/junction クエリの clinic スコープ統一（D12）— 規模 S

- **現状**: 6/29 read IDOR 監査で report-only とされた6箇所（D12 の表参照）が、clinic_id 述語なしのまま残存。いずれも上位防御あり・漏洩内容は不透明ID/整数のみで実害は低いが、「repository の読取は clinic スコープ必須」という規約の無追跡例外になっている。
- **手順**: 6箇所に `clinicScope(clinicID)` または clinic_id 述語を追加（呼び出し元は全て clinicID を保持済みのはず — シグネチャ変更が要る場合はその箇所だけ個別判断）。正当データでは結果不変（挙動保存）を既存テスト GREEN で確認し、越境データでの遮断を isolation テスト雛形で1本ずつ追加。
- **検証**: `docker compose exec backend go test ./internal/repository/... -run 'Isolation|Count'`
- **備考**: `FindAllByCategory` の Doctor preload は accept-and-document 済み（6/29 の設計判断・退職スタッフの既往担当医名を消す回帰を避ける）のため対象外。

### Phase 3: ガードレール整備（再発防止の機械化）

#### R3-1. CASCADE 検出 lint（D9・P0）— 規模 M

- **現状**: #211 の CASCADE 安全監査で「設計は安全（SET NULL が患者結果値を保護）」と確認済みだが、**新規 migration が CASCADE DELETE を持ち込むことを検出する lint が無い**（P0 指定）。migrations/CLAUDE.md の「CASCADE DELETE 禁止」が人力レビュー頼み。
- **手順**: 既存の go:embed + go/ast lint（preload lint）と同じ形式で、`backend/migrations/*.sql` を走査し `ON DELETE CASCADE` を検出したら fail する lint テストを追加（許容例外は allowlist に根拠付きで登録。純粋従属の子テーブル等）。#211 で実 DDL から挙動テストを抽出した手法があるため、静的検出＋既存の cascade 挙動テストの二層になる。
- **検証**: 意図的に CASCADE を含む一時 migration で lint が RED になること（temp-revert 実証）。

#### R3-2. preload lint の新 master 自動登録（D9・P1）— 規模 M

- **現状**: `preload_clinic_scope_lint_test.go` は clinic-scoped master テーブルのリストが手動管理。新規 master 追加時に登録漏れすると lint が素通りする。
- **手順**: model 層の走査（clinic_id カラムを持つ master 系 struct の自動抽出）でリストを生成する方式へ変更、または「model に存在するのに lint リストに無い」ことを検出する突合チェックを追加（後者が小さく確実。#211 inventory lint の双方向突合と同型）。
- **検証**: ダミー master を一時追加して lint が検出すること。

#### R3-3. openapi format↔シリアライズ突合 CI（D9・P1）— 規模 M

- **現状**: 6/30 調査で「format 宣言と Go 実装は現状整合（0/76）」と確認済みだが、**将来の drift を検出する CI が無い**。R2-1 の統一が完了してもそれを守る仕組みがない。
- **手順**: openapi の date/date-time 宣言と、対応する response builder の `.Format()` verb（wire format は型でなく Format verb で決まる — 6/30 の教訓）を突合するテストを追加。全 76 フィールドの手書き対応表ではなく、response 構造体の time 系フィールドを ast 走査 → openapi 定義と機械照合する形式にする。
- **検証**: 一時的に 1 フィールドの Format を崩して RED になること。
- **状態（2026-07-02 検証・完了）**: `backend/internal/apicontract/openapi_date_format_drift_test.go` に実装済み。`docs/api.yaml` の `format: date` 宣言 ↔ handler `*_response.go` の `time.Time`/`*time.Time`（datetime wire）を go/ast で突合。独立 package のため handler test package の compile 状態に非依存＝BLOCKED 事由「handler 大規模改修中」を回避（実測: 変更中の handler `*.go` 100 件は全て `_test.go`・response builder は不変）。既存 drift 22 箇所/18 キーを allowlist 固定し新規・数変化・stale で fail。**20:31 時点の版**で scoped test 4 件 GREEN・allowlist 1 件除去で RED を実証（drift 検出が機能することを確認）。dedicated CI job `openapi-date-format-drift`（全イベント無条件）配線済み。⚠️ 6/30「format↔実装 0/76 整合」は現 HEAD では不成立（22 drift 既存）— 解消（openapi date-time 化 or handler date-only 文字列化）は FE contract の follow-up。**本領域は並行セッションが同時実装・改良中**（当該版はその後 go-reviewer 指摘反映で `gopkg.in/yaml.v3` ベースへ改良され本セッションでは未再実行。handler test 群・coverage-ratchet も並行進行）。最終 GREEN は当該セッションの CI に委ねる。実装・コミット所有は並行セッション側。

#### R3-4. repository 層テスト補強（D6/D7）— 規模 L

- **現状**: repository は src 104 / test 31 で最弱層。特に auth ライフサイクル（password_reset_token の期限/単回使用・token_blacklist の失効）はテスト 0件（bug.md M-8）。pet CRUD は R/D のみ、medical_record_image 未カバー。6/29 監査で「repo テスト pool 枯渇」（P1）も報告済み。
- **手順**:
  1. **先に pool 枯渇を解消**（これを直さないとテスト追加が不安定化する）: setupTestDB のコネクション管理を確認し、テスト毎の close 漏れ / MaxOpenConns を是正。既知の DROP 順序罠は #196 のノウハウに従う。
  2. auth ライフサイクルテスト追加（M-8 の 3+2 ケース: 期限内取得/期限切れ失敗/単回使用/失効拒否/クリーンアップ）。
  3. pet Create/Update の clinic スコープテスト、medical_record_image の隔離テストを既存 table-driven 雛形で追加。
  4. それ以上の網羅（#212 の 90% 目標）は本計画の範囲外 — カバレッジ追跡 Issue（bug.md H-7-3）へ委譲。
- **検証**: `docker compose exec backend go test ./internal/repository/...` が安定 GREEN（3回連続実行で flake なし）。

#### R3-5. カバレッジ ratchet ゲート（bug.md H-7 連動）— 規模 S

- **現状**: CI はカバレッジ計測のみで非ゲート。下がっても検出されない。
- **手順**: bug.md H-7 の対応方針どおり（coverage-policy.md Phase 1-2 の実装＋ベースライン実数記録）。本計画では「Phase 1〜3 のリファクタでカバレッジを下げていないこと」の証明装置として位置付ける。**Phase 1 着手前に入れるのが理想**（リファクタの安全網になる）。
- **検証**: カバレッジを下げる一時変更で CI が warn/fail。

#### R3-6. golangci-lint の issue 数 cap 解除（D11）— 規模 S

- **現状**: `backend/.golangci.yml:140-141` に `max-issues-per-linter: 50` / `max-same-issues: 10` が残存。6/30 に「max-same-issues:10 で11件目が隠れ、修正済みと誤認する」事故を実際に経験し（27件一括解消で解決）、`--max-same-issues 0 --max-issues-per-linter 0` が正本 gate と教訓化されたが、**設定ファイルは cap されたまま**。
- **手順**: 両値を `0`（無制限）に変更。6/30 時点で隠れ分は全解消済みのため、現状 clean なら差分は設定2行のみ。もし新たに露出する issue があれば同一 PR で解消（それ自体が cap の実害の証明になる）。
- **検証**: scoped lint（`docker compose run --no-deps --entrypoint golangci-lint backend run ./internal/...` — キャッシュフレッシュ化に注意）で issue 0 件。CI の Backend job green（Lint だけでなく Test step まで確認 — 順次 step マスキングの既知罠）。

#### R3-7. 臨床結果テーブルの DB レベル複合 FK 追加（D13）— 規模 M

- **現状**: exam_results / checkup_field_results のクロステナント FK 整合はアプリ層検証（FindByID(clinicID) ガード・#124 同型対策）のみ。DB レベルでは通常の単純 FK しかなく、clinic_id を含む複合 FK による物理防御が無い（#211 follow-up の明示的積み残し）。
- **手順**:
  1. **既存データの違反有無を先に検証**（違反行があると FK 追加が失敗する）: 対象テーブルの親子 clinic_id 突合クエリを verify_seed.py 系または一時検証で実行。
  2. 親テーブル側に `(id, clinic_id)` の UNIQUE 制約（既存 PK + clinic_id）、子テーブル側に複合 FK を張る **additive migration** を新規追加。ON DELETE は既存設計（SET NULL ★患者結果値保護）を踏襲し、CASCADE は持ち込まない（R3-1 の lint と整合）。
  3. #211 で確立した「migration から実 DDL を抽出して挙動テスト」の手法で、複合 FK が越境 INSERT を拒否することをテスト化。
- **検証**: cascade/FK 挙動テスト GREEN + STG 適用は db_reset 運用ルールに従う。
- **備考**: 制約追加は「不正データを物理的に拒否する」変更であり、正当データに対しては挙動保存。適用前検証（手順1）が必須条件。
- **状態（2026-07-02 検証）**: checkup_field_results = **完了**（`migrations/012_add_clinical_result_composite_fk.sql` ＋ `repository/checkup_field_composite_fk_test.go`：越境 INSERT 拒否/同一 clinic 許可/列指定 SET NULL 保護）。挙動テストの live 実行は並行セッションの container 飽和（同時 `go test` 10 本）で本セッションでは未取得だが、DDL・テスト・schema を直接確認済みで CI backend job（`go test ./...`）が clean runner で実行する。**exam_results = 正当 BLOCKED（別タスク）**: exam_results も参照先 `exam_type_fields` も `clinic_id` 列を持たず（clinic は `exam_type_fields→exam_types→clinics` と 2 段先。`001_init.sql` 実測）、`(id, clinic_id)` 複合 FK が構造的に張れない。同等防御は `clinic_id` 列追加＋backfill という**非 additive** スキーマ拡張を要し、実施ルール 5（additive 追記のみ）・behavior-preserving の範囲外。現行防御はアプリ層 `FindByID(clinicID)` ガード（#124 同型）で維持。

---

## 3. 非対象（明示的にやらないこと）

| 項目 | 理由 |
|---|---|
| bug.md の挙動修正（H-1 締め履歴フィルタ / H-3 現金集計 / M-1・M-2 クレジット訂正等） | バグ修正トラック。behavior-preserving でないため本計画に混ぜない（混ぜると「リファクタで挙動が変わった」の切り分けが不能になる）。ただし H-3 の**回帰テスト復元**は R1 と同じテスト資産を触るため、同時期実施が効率的 |
| `cmd/seed-old-db/transform.go`（962行）の分割 | seed 移行 CLI。6/30 監査で高凝集 keep 判断済み。プロダクション経路外 |
| `treatment_service.go`（645行）等の大ファイル分割 | 全ファイル 800 行上限内。分割の実益なし（YAGNI） |
| レイヤ追加・DDD 化・パッケージ再編 | 現行の軽量レイヤリングで問題が出ていない。アーキ変更は課題駆動でのみ行う |
| goroutine 周りの見直し | 6/30 監査で全箇所 documented・問題なしと確認済み |
| シークレット管理（bug.md C-1/H-4/H-5） | セキュリティ修正トラック。リファクタではない |
| B4（締め後編集の権限判定位置） | 対応済みを確認: 認可は handler に残す設計判断がコメントで明文化され、post_close_reason 検証は service へ移設済み（accounting_handler.go:157-161） |
| FindAllByCategory の Doctor preload（Staff 履歴表示） | accept-and-document 済みの設計判断（6/29）: EXISTS スコープ化は退職スタッフの既往担当医名を消す回帰を生む。漏洩は staff 名のみ低 severity |
| **pwreset が効かない・30s タイムアウト（6/30 監査の未対応バグ）** | 挙動修正でありリファクタ対象外。追跡先は `docs/tasks/open/PERF-FOLLOWUP-05-pwreset-30s-timeout-unreliable.md`（2026-07-02 確認 — 当初「孤児」と記載したが docs/tasks で追跡済みだったため訂正）。R1-3 の fail-open 調査（同 FOLLOWUP-06）と同時期の着手が効率的 |

---

## 4. 実施ルール

1. **挙動保存の証明を各コミットに含める**: 既存テスト GREEN 維持が最低条件。テストが無い箇所を触る場合は**先に現挙動を pin するテストを書いてから**リファクタする（R2-4 の手順が典型）。
2. **唯一の挙動変更例外は fail-closed 化（R1-3）**: これは fix: として明示し、失敗注入 RED→GREEN で変更を実証する。
3. **検証は scoped で自走**: `docker compose exec backend go test ./internal/<pkg>/...` を各コミット前に実行。全体 test / lint / build はプロジェクトルールに従いユーザー手動（完了報告時にコマンドを提示）。
4. **コミット粒度は 1 項目 1 コミット**（R 番号をメッセージに含める）。並行セッション対策として commit 前に HEAD 確認・パス限定 stage。
5. **migration は新規追記のみ**（適用済みファイルの編集は checksum mismatch）。本計画で migration が必要になるのは R3-1 の検証用一時ファイルのみ（コミットしない）。
6. **subagent の指摘は直接再検証してから採用**（6/30 の教訓: 過大報告の前例あり）。
7. **順序**: R3-5（安全網）→ Phase 1 → Phase 2 → Phase 3 残り、が理想。ただし各項目は独立しており、優先度変更があっても単独で着手可能。

## 5. 全体見積もりと完了条件

| フェーズ | 項目数 | 規模合計 |
|---|---|---|
| Phase 1（tx 整合性） | 3 | M+M+S〜M ≒ 3日 |
| Phase 2（表現統一・削減） | 5 | S+S+S+M+S ≒ 2.5日 |
| Phase 3（ガードレール） | 7 | M+M+M+L+S+S+M ≒ 6.5日 |

**完了条件**:
- audit-tx inventory lint の未移行分類（pending-migration）が 0 件
- WithTx 内呼び出しの repository メソッドが全て dbOrTx 参加（tx 参加の直接証明テストあり）
- API 日付表現が単一形式で、drift 検出 CI が存在する
- `grep -rn medication-history backend/` が空（または保留根拠コメントあり）
- repository の読取クエリ（count/junction 含む）が clinic スコープ規約に例外なく準拠（既決の accept-and-document を除く）
- repository テストが安定 GREEN（pool 枯渇解消）＋ auth ライフサイクル検証あり
- カバレッジ ratchet が CI で有効・golangci-lint の cap が 0（隠蔽なし）
- checkup_field_results に DB レベル複合 FK（越境 INSERT 拒否のテスト付き）＝**完了**（migration 012）。exam_results は `clinic_id` 列不在で additive 複合 FK 不可＝**正当 BLOCKED**（別タスク: 非 additive な列追加＋backfill、要 PO/architect 判断）。R3-7 の状態を参照
- pwreset タイムアウトバグの追跡 Issue が起票済み（本計画外への引き継ぎ完了）
