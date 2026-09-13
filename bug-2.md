# 一時バグ／障害メモ（全ページ CRUD UAT · 2026-09-13）

> ユーザー依頼: 全ページの CRUD／動作確認で発見したバグ・コンソールエラーを本ファイルに記載。  
> `bug.md`（スタッフ／予約系）とは分離。製品 FAIL 正本は引き続き `todo.md#product-bugs`。

更新日: 2026-09-13  
実施完了: 2026-09-13 03:46 JST

修正計画追記: 2026-09-13。各項目の方針・対象ファイル・検証・完了条件は索引から参照。以下の `OPEN` は元の UAT 記録を保持しており、計画追記による完了扱いはしない。

## 索引

| ID | status | area | severity | 種別 | 修正・検証プラン |
|:---|:---|:---|:---|:---|:---|
| BUG2-MR-ENTERED-BY-CLINIC | FIXED | medical-records | High | **バグ断定**（entered_by と選択医院の複合FK） | [本人 ID を保持し、所属認可・FK・履歴読取を整合](#plan-bug2-mr) |
| BUG2-PAYMETHOD-CREATE-FORBIDDEN | OPEN | payment-methods / authz | Medium | **権限／UX**（一覧可・作成 403） | [既存の権限制御を再検証し、付与方針を PO 判断](#plan-bug2-paymethod) |
| BUG2-RES-DIALOG-A11Y | OPEN | reservations / a11y | Low | **コンソール**（`bug.md` の BUG-RES-DIALOG-A11Y-CONSOLE と同系） | [既存修正を紐付け、同じ操作で再検証](#plan-bug2-a11y) |
| NOTE2-SWEEP-COVERAGE | — | uat | — | カバレッジ記録 | [未確認の詳細画面・入院・検査・健診を補完](#plan-note2-coverage) |

## 進捗

- 証拠: `reports/uat-2026-09-13/all-pages/`
  - ルート巡回: `all-pages-20260913-033603.json`（**82** ページ、clinic_id=2）
  - CRUD深掘り: `crud-deep-*.json` / `crud-retry-*.json` / `crud-final-*.json` / `crud-verify.json`
- 方針: HydrateFallback は既知ノイズとして非起票（`bug.md` / BUG-20260906-002 参照）
- 動的詳細で ID 未取得のためスキップ: `['/accounting/:id', '/hospitalization/:id', '/hospitalization/:id/edit', '/inventory/:id']`

## ページ結果サマリ

- **ルート巡回**: 82 / 展開可能ルートすべて **画面到達 PASS**（認証破綻なし）
- **コンソール（ノイズ除外）**: `/reservations` で DialogContent Description 欠落警告のみ（新規ダイアログオープン時）
- **HTTP≥400（ページ sweep）**: 実質なし（login の一時 `/me` は除外）

### CRUD 成立したもの（城東 clinic_id=2・執行デモ）

- 職種 / ケージ / 保険 / 予約区分 / 薬品 / 物販 / 問診テンプレ / 権限グループ / 診断タイプ / 健診タイプ / キャンペーン / シフトテンプレ
- 飼主・ペット CUD、見積 CRUD、ワクチン CD、トリミング Create、在庫 CRUD
- スタッフマスタ／予約は前回 `bug.md` 範囲のため本ファイルでは再掲最小限

### CRUD で権限・契約上ブロック（製品欠陥としない／要PO）

- **動物種 create**: 403（system admin 限定の既存仕様）
- **支払方法 create**: 403 → 下記 BUG2-PAYMETHOD-CREATE-FORBIDDEN（view 可・create 不可）
- **健診 POST `/api/v1/checkups`**: ルートなし（作成はカルテ経由 `create-checkup-medical-record` → `/v1/medical-records`）

## バグ断定

### BUG2-MR-ENTERED-BY-CLINIC: カルテ作成が `entered_by`×医院の複合FKで「参照先が存在しません」になる

- **status**: **FIXED**（2026-09-13 · attempt `att-bug2-mr-entered-by-20260913-001`）
- **現象**: 城東（clinic_id=2）選択中に `POST /api/v1/medical-records` すると **400** `参照先が存在しません`。既存 STG 飼主・ペットでも再現。
- **実測（2026-09-13）**:
  - ログイン: 八王子メインの執行デモ（staff_id=**10000021**）、`X-Clinic-ID: 2`
  - body 例: `pet_id`/`owner_id` を string、`visit_date=2026-09-13`、`visit_type=revisit`
  - backend log:  
    `violates foreign key constraint "fk_medical_records_entered_by_clinic" (SQLSTATE 23503)`  
    INSERT の `entered_by=10000021`, `clinic_id=2`
- **原因**: `entered_by` にログインスタッフ ID を入れるが、複合 FK `(entered_by, clinic_id) → staffs(id, clinic_id)` に対し、**10000021 は城東行を持たない**（八王子所属）。医院切替だけでは `entered_by` の所属が満たされない。
- **修正内容**:
  1. 認証 actor の staff ID を `entered_by` のまま保持（identity-link 置換なし）
  2. 新規 migration `backend/migrations/002_medical_records_entered_by_staff_fk.sql` で複合 FK を `entered_by → staffs(id)` に置換（`001_init.sql` 未編集）
  3. create tx 内で有効所属ロック（system admin は DB 検証付き例外、auto-create は `EnteredBy=nil` の内部経路）
  4. list/detail の親 guard から entered_by 医院相関を外し、記録者 preload を履歴表示向けに整合（doctor/owner/pet 隔離は維持）
- **検証**: `docker compose exec backend go test ./internal/medicalrecord` exit 0（EnteredBy_* real-DDL 含む）
- **ユーザー作業**: **`make migrate`** を実行して 002 を適用（エージェントは未適用）
- **UAT 証拠（旧）**: docker `animalekarte-backend-1` ログ（2026-09-13 03:46 JST 付近）· `reports/uat-2026-09-13/all-pages/crud-verify.json`

### BUG2-PAYMETHOD-CREATE-FORBIDDEN: 執行ロールで支払方法マスタが「見れるが作れない」

- **現象**: `/settings/payment-methods` は開ける。`GET /api/v1/payment-methods` は 200。`POST` は **403 forbidden**。
- **実測**: `/api/v1/me` の permissions で `master-payment-method`: `view=true, create=false, edit=true, delete=false`
- **判断**: コード欠陥というより **権限マトリクス**。執行で create=false かつ edit=true は運用として不自然な可能性があるため起票。
- **次アクション**: PO に「執行は支払方法を作成できるべきか」を確認。UI は create 不可なら新規ボタンを出さない方がよい。

### BUG2-RES-DIALOG-A11Y: 予約ページの DialogContent Description 欠落コンソール警告

- **現象**: `/reservations` で新規予約を開くと  
  `Warning: Missing Description or aria-describedby={undefined} for {DialogContent}.`
- **注**: `bug.md` の `BUG-RES-DIALOG-A11Y-CONSOLE` と同系。全ページ sweep でも再確認。

## コンソール／HTTP 所見（全ページ）

| 場所 | 内容 | 扱い |
|:---|:---|:---|
| `/reservations` | DialogContent Description 欠落 warning | BUG2-RES-DIALOG-A11Y |
| 他 81 ページ | HydrateFallback 以外の興味イベントなし | 非起票 |

## NOTE2-SWEEP-COVERAGE

- 対象医院: **城東センター病院 (2)**（デモ執行・メインは八王子だが X-Clinic-ID/localStorage で 2）
- 画面到達: paths.ts 静的＋取得できた詳細 ID を展開（82）
- API CRUD: 主要マスタ／飼主ペット／見積／ワクチン／トリミング／在庫まで確認
- カルテ新規は上記 BUG2 でブロック（UI 新規も同経路の可能性大）
- 入院は `cage_id` 必須など契約が厳しく、ケージ指定付きの追加確認は未完了（契約バリデーションとしては正常応答）
- 検査 create は examination-types 取得後も追加フィールド不足で 400 — 追加切り分け余地あり（本ラウンドでは製品断定せず）

---

## 修正・再検証プラン（2026-09-13）

調査基準はローカル `main` の `f759492cf1052c6ce4d22989d646eba5d470c0d5`。本追記はコードと既存 UAT 記録に基づく計画であり、今回の実装・テスト・DB 適用・ブラウザ再検証は未実施。Linear は読み取り検索で対応 Issue を特定できておらず、実行ステータスは **UNKNOWN**。実装着手時に `todo.md#product-bugs` と Linear の対応先を確認する。

### 着手順・依存関係

| 順序 | 対象 | 最初の作業 | 次に進む条件 |
|:---|:---|:---|:---|
| 1 | BUG2-MR-ENTERED-BY-CLINIC | 記録者と医院所属の契約を確定し、実 DDL で失敗を再現するテストを準備 | 記録者・system actor・履歴表示の設計確認後にアプリと FK を修正 |
| 並行可 | BUG2-PAYMETHOD-CREATE-FORBIDDEN | 現行 UI の権限ガードを再検証し、PO が執行の create 可否を決定 | 不許可なら期待値を訂正、許可するなら対象医院・権限グループを限定して変更 |
| 並行可 | BUG2-RES-DIALOG-A11Y | 既存修正 `cde5d4b1e` の対象ビルドを特定して再検証 | 同一操作で警告なし・説明とフォーカス正常を確認 |
| 2 | NOTE2-SWEEP-COVERAGE | 未確認操作と必要なテストデータを一覧化 | カルテ修正後に健診を再確認。入院・検査等の独立した確認は先行可 |

実装時は各 ID の claim と担当ファイルを確保する。共通コードに触れる作業は同時編集せず、別 worktree の変更を順に統合する。

<a id="plan-bug2-mr"></a>

### BUG2-MR-ENTERED-BY-CLINIC — 本人の記録者 ID と選択医院の認可を両立する

**確認できたこと**

- [medical_record_handler.go](backend/internal/medicalrecord/medical_record_handler.go) と [medical_record_request.go](backend/internal/medicalrecord/medical_record_request.go) は、認証コンテキストの staff ID を `EnteredBy` に設定する。リクエスト本文から記録者を選ぶ仕組みではない。
- [001_init.sql](backend/migrations/001_init.sql) の `fk_medical_records_entered_by_clinic` は `(entered_by, clinic_id) → staffs(id, clinic_id)` を要求する。一方、[認証仕様](docs/architecture/auth.md) は、一般スタッフの選択医院への所属と医院別権限を `staff_clinic_assignments` 等から評価する。スタッフの主所属医院と、操作を許可された医院が一致しない場合がある。
- [medical_record_repository.go](backend/internal/medicalrecord/medical_record_repository.go) の親レコード絞り込みは主所属または assignment を考慮するが、記録者の一覧用 preload と詳細用 preload は有効所属・退職の扱いが異なる。FK だけの修正では、保存成功後に一覧・詳細から消える可能性が残る。
- [medical_record_auto_create.go](backend/internal/medicalrecord/medical_record_auto_create.go) の予約からの自動作成は `EnteredBy=nil` で共通 create 処理を呼ぶ。通常の HTTP 作成と区別が必要。

**推奨方針と実装前の決定事項**

記録者は「記録した本人」を表すものとし、認証された元の staff ID を保持する。`entered_by` に限って staff 単体への FK で存在を保証し、選択医院での作成可否は認可と保存時の検証で保証する案を推奨する。飼主・ペット・担当医・予約の医院境界は維持する。

実装担当と設計レビュー担当は、次の契約をコード変更前に確定する。

| 論点 | 計画上の扱い |
|:---|:---|
| 通常の HTTP 作成 | 本人 actor 必須。一般スタッフは選択医院への有効所属と当該医院の create 権限が必要。他院での権限を流用しない |
| system admin | 既存仕様の active clinic への操作権限を維持。通常 assignment がないことだけを理由に拒否しない。作成者の本人性は保持 |
| 予約自動作成など内部経路 | 既存の system actor 契約を明示した内部経路として扱う。HTTP の actor 欠落を同じ扱いに落とさない。共通 create に通常ユーザーの create 権限を無条件で要求しない |
| 所属解除・退職後の記録 | 閲覧者自身の医院・カルテ閲覧権限で過去記録を取得でき、記録者 ID が変わらないことを目標とする。無効化後の記録者名の表示方針と最小限の返却項目を決める |
| 不正な記録者参照への防御 | 保存時の本人性・認可検証と既存データ点検を根拠に、`entered_by` 専用の読取契約を定める。親 guard を一律削除したり、他院 Staff 全情報を無条件 preload したりしない |

元の候補 1 の staff identity-link は現行の飼主・ペット向け identity-link で代用できないため採用しない。本人 ID の別スタッフへの置換、主所属医院の書換え、HTTP の actor を NULL/0 にする回避も行わない。候補 3 のエラー明確化は無権限時の補助であり、正当に所属・権限を持つ本人の作成成功が完了条件となる。

**作業手順・主な対象**

1. **再現テストを先に追加する。** 主所属 A・有効所属 B・B の作成権限ありの本人で、B の飼主・ペットを使う POST を再現する。最初は担当医未指定として原因を actor に限定する。実際の migration の複合 FK を使い、旧コードで失敗することを確認する。[realdb_records_isolation_test.go](backend/internal/medicalrecord/realdb_records_isolation_test.go) 等の AutoMigrate ベースのテストだけでは、当該手書き DDL の再現証拠としない。
2. **保存前の actor 検証を追加する。** [medical_record_crud.go](backend/internal/medicalrecord/medical_record_crud.go) の作成 transaction で、関連レコードの更新より先に actor 契約を検証する。HTTP／内部作成を明示的に区別し、認証外の入力で admin や system actor を指定できない形にする。所属確認には既存の [staff_clinic_assignment_repository.go](backend/internal/staff/staff_clinic_assignment_repository.go) の transaction 対応・ロック機構を優先して利用する。
3. **競合時の保証を明示する。** [current_access_staff_reader.go](backend/internal/auth/current_access_staff_reader.go) の resolver を transaction 内で呼ぶだけでは同一 transaction のロックにならない。所属解除や staff/account/clinic 無効化に関わる行、更新側のロック順（staff → assignment）、権限変更側の仕組みを確認し、必要な検証を ambient transaction に参加させる。権限剥奪は認証仕様の「変更コミット後の次リクエストで拒否」を最低条件とし、受付済みリクエストとの競合をどちらの順で成立させるかを定義して独立 transaction のテストで確認する。
4. **FK と読取処理を一緒に整合させる。** `backend/migrations/` に着手時点の次番号の migration を追加し、`entered_by` のみ単列 FK へ置換する。既存 `001_init.sql` は書き換えない。`medical_record_repository.go` と [medical_record_repository_list.go](backend/internal/medicalrecord/medical_record_repository_list.go) の親 guard・一覧・詳細・count を確認し、記録者専用の最小投影で履歴を表示する。Doctor と共有する staff helper を一括で緩和しない。既存の foreign entered-by 拒否テストは、新しい actor 契約を根拠に保存拒否・読取項目制限のテストへ対応づけ、他の FK 破損防御は残す。
5. **監査と失敗時の状態を確認する。** create の監査は現在 post-commit の best-effort。actor が実際の本人 ID のまま記録されることを確認する。この監査だけを履歴の所属正当性の証拠にしない。監査を必須・同一 transaction に変える場合は明示的な契約変更として扱う。認可拒否時はカルテや予約関連更新を残さず、FK 名や他院情報を返さない。
6. **移行と適用順を検証する。** 既存データの参照切れ・不整合を点検し、黙って本人 ID を補正しない。新規構築と既存 schema からの増分適用を、承認された検証 DB で確認する。FK を緩める前に全稼働 writer の actor 検証を用意し、混在バージョンで無防備な期間を作らない。新たな院外主所属 actor の記録ができた後は旧複合 FK をそのまま復元できないため、適用前に停止・復旧手順も決める。DB 作成・適用・破棄は別途承認された運用で行い、エージェントは migration を自動適用しない。pull 後の適用担当には `make migrate` を案内する。

**回帰テスト・完了条件**

| ケース | 必須の結果 |
|:---|:---|
| 主所属 A・所属 B・B の create 権限あり | B に 201 で作成され、`entered_by` は本人 ID。再取得・一覧・詳細・件数が整合 |
| 主所属と選択医院が同一 | 既存の通常作成が成功 |
| 非所属／B の create 権限なし／A の権限だけあり | 既定の認可エラーで拒否、カルテ・関連更新なし |
| 本文の別人 actor、actor 欠落、無効な本人 | 別人への付替えや system actor への逃げ道がなく、認証由来の本人性を強制 |
| assignment のない system admin | 有効な対象医院で既存認可仕様に従い作成・再取得できる |
| 所属解除・無効化・権限変更との競合 | 定義した処理順と次リクエストでの失効が成立。所属変更後も権限のある閲覧者が過去記録を取得可能 |
| 他院の owner/pet/doctor/appointment を混入 | 各既存境界で拒否。記録者の例外が患者・診療対象の医院分離に波及しない |
| 予約自動作成・健診からのカルテ作成 | 内部 system actor 経路と HTTP 本人経路がそれぞれ成功し、二重作成など既存防御も維持 |
| 既存 DB の増分移行 | 既存カルテと記録者を保持。実 DDL で旧 FK の再現から修正後の成功を確認 |

実装後は `docker compose exec backend go test ./internal/medicalrecord` を基本に、変更した auth・staff・migration の対象テストを追加実行する。DB を使う検証は承認済み接続先・schema・実行方法を先に確認する。`TestPreloadClinicScope` の `EnteredByStaff` 免除や Master FK inventory の Staff 免除があるため、既存静的チェックの PASS だけでは本件を完了にしない。最後に元の医院切替 → 新規カルテ保存 → 再読込を UI で確認し、証拠を紐付ける。

<a id="plan-bug2-paymethod"></a>

### BUG2-PAYMETHOD-CREATE-FORBIDDEN — 作成権限の期待値を確定する

**確認できたこと**

- [billing/routes.go](backend/internal/billing/routes.go) は GET に view、POST に create、PATCH に edit、DELETE に delete を要求する。`create=false` の POST 403 はこの契約と一致する。
- [clinic_service.go](backend/internal/clinic/clinic_service.go) の既定権限表でも、支払方法の執行権限は view/edit のみ。今回の値だけを根拠に権限データ破損とは判断できない。
- [PaymentMethodSettings.tsx](frontend/src/features/master/routes/PaymentMethodSettings.tsx) は既に `usePermission` を利用する。[MasterPageShell.tsx](frontend/src/features/master/components/MasterPageShell.tsx) は create 不可時に新規ボタンを隠し、[MasterCRUDPage.tsx](frontend/src/features/master/components/MasterCRUDPage.tsx) は作成パネルを readOnly にし、[use-master-save.ts](frontend/src/features/master/hooks/use-master-save.ts) は送信境界でも権限を確認する。

**作業手順**

1. 元の 403 が画面操作由来か、UAT の直接 POST 由来かを証拠で分ける。同じ選択医院の `/me`、画面、送信を対応づけ、最新コードで新規ボタンや送信の漏れが本当にあるか確認する。漏れが再現した箇所だけ、先に失敗テストを置いて修正する。
2. **PO 決定待ち:** 執行に `master-payment-method:create` を付与するか、対象医院・権限グループを含めて決める。決定前は既定権限を維持し、画面再検証とテスト準備を進める。
3. **付与しない場合:** 403 と新規操作不可を正常な期待値として UAT に記録する。コードが既に一致すれば追加実装せず、権限方針確認事項として終了する。
4. **付与する場合:** 既存の権限管理機能で対象を限定して変更する計画を作る。既存医院の権限変更と、新規医院向け `clinic_service.go` の既定値変更は別に判断する。既定値や seed の変更だけで既存医院にも反映されたとは扱わない。[支払方法仕様](docs/spec/screens/settings/payment-methods.md) と権限テストを更新し、実環境の権限変更は承認された担当者が行う。
5. 医院切替・権限再取得・編集中の権限剥奪も確認する。UI のロール名固定判定や view からの create 推測を増やさず、現在の医院別 action 権限を利用する。バックエンド 403 時は入力を保持し、成功通知を出さない。

**検証・完了条件**

- PO の付与可否と対象を記録し、`create=false` では新規ボタン非表示・作成送信なし・直接 POST は 403 を確認する。edit が許可される既存行は、行の制約に従って編集できる。
- 付与する場合のみ、正当な権限ありの custom 支払方法作成が 201、権限なし・別医院では拒否されることを route middleware を含めて検証する。
- system_key を持つ標準行の操作制限・医院分離・[ADR-003](docs/architecture/adr/003-payment-method-identity-and-consistency.md) の識別・会計整合を維持する。create の変更から delete まで広げない。
- フロントは既存の `MasterCRUDPage.test.tsx`、`use-master-save.test.ts`、`payment-method-settings-model.test.ts` を優先して回帰を追加し、Docker の `vitest run` で変更対象ファイルを指定する。バックエンドを変更した場合は該当 billing／clinic／auth テストを Docker で実行する。

**終了判断:** 現行契約どおりなら「仕様どおり／UAT 期待値訂正」、実際の UI 漏れや権限仕様を変更したなら、その差分と検証証拠を添えて終了する。無条件の POST 200/201 化を目標にしない。

<a id="plan-bug2-a11y"></a>

### BUG2-RES-DIALOG-A11Y — 既存修正を再利用して UAT を閉じる

**現状:** `bug.md` の `BUG-RES-DIALOG-A11Y-CONSOLE` には修正記録があり、現在の HEAD に `cde5d4b1e`（`fix(a11y): restore Radix dialog description wiring on reservations`）が含まれる。本件は **既存修正あり・今回の再検証待ち** として扱う。

[ReservationFormModalPanels.tsx](frontend/src/components/shared/ReservationFormModal/ReservationFormModalPanels.tsx) は `DialogDescription` の ID を Radix に任せている。既存修正は独自 ID と Radix の説明 ID の不一致への対応で、[ReservationFormModal.dialog-a11y.test.tsx](frontend/src/components/shared/ReservationFormModal/ReservationFormModal.dialog-a11y.test.tsx) には実 Dialog を使う回帰テストもある。共通 `CommandDialog` 等を今回の原因として追加修正する根拠はない。

**作業手順・検証**

1. UAT を行ったビルドと修正コミットの包含関係、現在の Docker／ブラウザが読むビルドを確認する。古い成果物での再現と現行コードの回帰を区別する。
2. 既存テストを実行する: `docker compose exec frontend npx vitest run src/components/shared/ReservationFormModal/ReservationFormModal.dialog-a11y.test.tsx`。
3. `/reservations` の新規予約を開く・閉じる・再度開く操作と、入れ子の予約区分ダイアログを確認する。説明参照先の存在、accessible name、キーボード操作、閉じた後のフォーカス復帰、および Missing Description 警告が出ないことを記録する。
4. 現行ビルドでも再現した場合のみ、該当する DialogContent と説明の組を特定し、失敗する回帰ケースを追加して最小修正する。コンソール抑制や警告回避だけの `aria-describedby={undefined}` は対策にしない。

**完了条件:** 上記テストと元の UI 操作が成功し、ビルド・日時・証拠を記録したうえで `bug.md` の同一修正に紐付けて解消／重複を整理する。コミットや説明要素の存在だけで元の UAT を PASS に置き換えない。

<a id="plan-note2-coverage"></a>

### NOTE2-SWEEP-COVERAGE — 未確認の操作を埋める

これは追加の製品バグではなく検証の残作業。**82 ページ到達 PASS と、全ページの CRUD 完了は分けて記録する。**

| 未確認対象 | 準備・確認方法 | 依存／判定 |
|:---|:---|:---|
| `/accounting/:id` | 対象医院の参照可能な会計 ID をテストデータから取得し、詳細表示と権限別操作を確認 | ID 未取得のまま PASS にしない |
| `/hospitalization/:id`、`/hospitalization/:id/edit` | 有効なケージ・飼主・ペット等を [入院画面仕様](docs/spec/screens/09-hospitalization-form.md) と request DTO に合わせて準備し、作成→詳細→編集→再読込を確認 | 必須項目不足の 400 は入力契約確認。正当な入力でも失敗すれば新規不具合として切り分け |
| `/inventory/:id` | 既存の CRUD 成功記録と作成 ID を照合し、対象医院の詳細画面を確認 | API 成功を詳細 UI の検証に代用しない |
| 検査 create | [検査画面仕様](docs/spec/screens/13-examinations-form.md)・request DTO・画面 payload を照合し、不足項目を特定して有効な入力で再実行 | 400 の原因が契約・UI 送信漏れ・API 不具合のどれかを確定して記録 |
| 新規カルテ・健診 | [create-checkup-medical-record.ts](frontend/src/features/checkups/api/create-checkup-medical-record.ts) のカルテ作成→カルテ配下の健診操作を UI から確認 | BUG2-MR-ENTERED-BY-CLINIC の解消に依存。存在しない汎用 `/checkups` 作成 API を追加しない |

**実行・完了条件**

1. `reports/uat-2026-09-13/all-pages/` の既存証拠を保持し、新しい run の対象ビルド・医院・権限・前提データ・route/action・期待結果・実結果・証拠を記録する。対象一覧を母数として、未実行は PENDING、前提不足は BLOCKED、対象外は理由付き SKIP とする。
2. テストデータは承認された検証環境で準備し、他の利用者の診療データを変更しない。更新・削除の確認は仕様上許可される操作を対象とし、削除不可の臨床記録を物理削除して CRUD を埋めない。
3. 上表の未確認操作に結果と証拠が付き、製品 FAIL があれば正本へ紐付ける。未解消の FAIL／BLOCKED が残る場合は対象と次の作業を残し、「全 CRUD 完了」と報告しない。

### この計画追記の検証範囲

既存 UAT 本文・ID・status を保持し、索引と 4 項目の計画の対応、参照ファイル、差分を確認する。変更対象は `bug-2.md` のみ。文書のみの変更のため runtime 検証は不要で、上記のテスト・DB 移行・UAT は将来の実装／再検証時の実行項目である。
