# GLM セッション バグ／疑義メモ（ユーザー依頼 2026-09-26）

> GLM 系エージェントセッションが検出したバグ・疑義の台帳。2026-09-26 に `bug.md` から新設・移管。
> 製品 FAIL の正本は引き続き `todo.md#product-bugs`。確定済みのみをそちらへ起票し、本ファイルは **切り分け前の疑義・調査中項目** を保持する。
> 確定した項目は `todo.md#product-bugs` へ昇格し、本ファイルでは昇格記録を残す。環境起因と確定した項目は「バグではない（記録のみ）」へ移動する。

更新日: 2026-09-26

## 索引

| ID | status | area | severity | 種別 | 詳細 |
|:---|:---|:---|:---|:---|:---|
| BUG-MR-DRAFT-AUTOPOST-FAILED | FE修正済み(PR)・STGデータ投入待ち | medical-record | High(確定) | **バグ断定**（seed に general 予約区分が無く、カルテ自動作成が再試行不能のデッドエンドになる） | [下記](#bug-mr-draft-autopost-failed) |

---

<a id="bug-mr-draft-autopost-failed"></a>

## UAT 2026-09-25 · STG ゼロベース走行 検出分（未断定・調査中）

> この節は STG（`https://stg.noah-karte.com`、clinic 1・合成アカウント）でのゼロベースUAT走行（run 記録: `reports/uat-2026-09-25/zerobase-run.md`）で検出した項目。製品 FAIL の正本は引き続き `todo.md#product-bugs` であり、本項は **切り分け未了のため未起票**。local 再現で製品起因と確定した時点で確定 FAIL へ昇格する。

### BUG-MR-DRAFT-AUTOPOST-FAILED（S06・未確定・暫定 High）

- **現象**: カルテ新規画面 `/medical-records/new?petId=` を開くとカルテ入力フォーム（タブ一式）は表示されるが、「**予約の作成に失敗しました。再試行する**」バナーが出て draft の自動 POST が失敗する。URL は `/medical-records/:id` に昇格せず、S06 の手順 1（表示と同時に draft 自動 POST → 昇格）が成立しない。
- **実証（2026-09-25、STG clinic 1・合成 owner/pet 1000000005・執行アカウント）**:
  - `/medical-records/new?petId=1000000005` を表示 → バナー表示、URL 不変。
  - 「再試行する」を 2 回実行 → いずれも同バナーのまま失敗（再現 3 回）。
  - 同ペットでは検査登録（`/examinations/new`）・入院登録（`/hospitalization/new`）・予防接種登録（`/vaccinations/new`）は成功しており、ペット/clinic スコープの障害ではない。
  - 自動化（ZCode in-app browser、合成イベント）で実施。ただし本フォームは locator fill・ページ内 click が有効な画面であり、失敗はサーバー応答由来のエラーバナー（イベント不達の無音失敗ではない）。
- **根因（未確定）**: 候補 (a) STG clinic 1 のマスタ/設定欠落に起因する環境障害（例: 診療種別「入院/外来」等のマスタ紐付け欠落） (b) draft 作成経路の製品欠陥。エラーバナーの語彙が「**予約の作成**に失敗」で文脈（カルテ作成）と不一致である点も要確認（別経路の appointment 作成を内部で試みて失敗している可能性）。
- **影響**: カルテが 1 件も生成できないため、S06（確定 Lock・監査証跡）に加え、カルテ詳細を前提とする S13/S16〜S19/S24〜S27/S29 など一連のシナリオが STG で検証不能。業務上は「外来カルテの新規作成が失敗する」事象に相当し、確定なら High。
- **次アクション**: local（disposable clinic）で同一手順を再現 → 製品起因なら `todo.md#product-bugs` へ確定起票+Linear 化、環境起因なら本項を「バグではない（記録のみ）」へ移動し STG fixture 手順に反映。
- **証拠**: `reports/uat-2026-09-25/zerobase-run.md`（S06 行・製品FAIL疑義セクション）、Plane 起票ドラフト `reports/uat-2026-09-25/tickets/BUG-MR-DRAFT-AUTOPOST-FAILED.md`

### 切り分け結果（2026-09-26・コード診断 + STG API 実測）— **製品欠陥として確定**

- **STG 実測**: `stg-staff-10000021`（執行）で login 200 → `GET /api/v1/masters/reservation-types` は **1 件のみ**（`id=1 / トリミング / category=trimming`）。`category=general` は **0 件**。
- **根因（確定）**: `backend/migrations/seeds/002_master/reservation_types.csv` には `category=trimming` の行しかなく、**seed は一般区分を 1 件も投入しない**。一方 FE `findGeneralReservationType`（`use-medical-record-form-model.ts`）は `category === "general"` の区分のみをカルテ自動作成の前提とし、無ければ前提欠落 failure になる。結果、**002_master だけで構成された clinic（STG 全医院・fresh DB 全体）ではカルテ自動作成が必ず失敗し、再試行は 30 分 staleTime の reservation-types キャッシュ下で再解決されないため永久デッドエンド**。
- **既存のデータ修復手順**: `backend/migrations/seeds/live_insert_standard_reservation_types.sql`（EMR-193。診察/お手入れ/ワクチン/健診を全医院へ冪等投入する手動手順）が存在するが **STG 未適用**。適用は runbook 経由で USER 実施（エージェントは direct DB op をしない）。fresh DB への恒久反映は seed-export 再生成（USER レーン）。

### 対応（fix/bug-mr-draft-autopost-failed・PR）

- **FE**: `MedicalRecordAutoCreateFailurePhase` に `appointment-master-missing` を新設し、前提欠落と API 失敗を区別。master 欠落時は「予約区分マスタに診察系の予約区分が登録されていません。マスタ設定 → 予約区分 で診察区分を追加した後、ページを再読み込みしてください。」を表示し、**再試行ボタンを非表示**（再試行ではキャッシュが再解決されず失敗が続くため）。
- **テスト**: `MedicalRecordAutoCreateFailure.test.tsx`（新 phase の表示+ボタン非表示）、`use-medical-record-form.auto-create-new.test.ts`（BUG-503 の trimming-only ケースを新 phase に更新）。vitest 3 ファイル 21 tests PASS / type-check PASS / scoped eslint PASS。
- **残件（USER レーン）**: ①STG へ `live_insert_standard_reservation_types.sql` を適用（適用後、カルテ作成は一般区分で成功する）②seed-export による fresh DB 恒久反映 ③Plane チケット起票（ドラフト済み・アクセス待ち）
