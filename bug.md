# クローズ済み Issue 仕様適合監査レポート

- **調査日**: 2026-07-02
- **対象**: 直近2週間（2026-06-18〜2026-07-02）にクローズされた GitHub Issue 45件（Search API と全件ローカルフィルタの2経路で突合し取りこぼしゼロを確認、全件 COMPLETED クローズ）
- **方法**: 各 Issue の本文・コメント・クローズ根拠を取得し、HEAD (main) の実コードと静的突合。CRITICAL/HIGH 全件＋主要 MEDIUM（M-4, M-6, M-8, M-9 および #181/#182/#188 の裏取り）は本体セッションで実コードを再検証済み。テスト実行・DB照会は未実施（静的検証のみ）。

## 判定サマリ

| 判定 | 件数 | 内訳 |
|---|---|---|
| MATCH（仕様充足） | 24 | #124 #125 #126 #127 #128 #151 #152 #153 #158 #159 #161 #179 #180 #182 #183 #184 #187 #188 #190 #191 #192 #198 #123 #193 |
| PARTIAL(一部未達) | 7 | #150 #155 #189 #194 #195 #196 #197 |
| MISMATCH(不一致) | 1 | #154 |
| CLOSED-AS-DECISION(実装なし・設計判断/リスク受容) | 13 | #156 #160 #178 #181 #185 #212 #89 #91 #96 #97 #98 #99 #109 |

MATCH 24件は受入条件を満たしている（一部 LOW 残課題あり、末尾参照）。以下、対応が必要な項目を深刻度順に「事象／根本原因／対応方法／検証方法」で記載する。

## 対応状況（2026-07-02 実施・main にコミット済み / push 未）

| 項目 | 状態 | コミット |
|---|---|---|
| H-1/H-2 締め履歴フィルタ＋ページネーション | ✅ 対応済み（FE を BE 契約に整合・AccountingDetail の同型バグも修正） | `a45da439` |
| H-3 現金集計 NULL 非対称 | ✅ 対応済み（短期案A: 後方互換復元＋回帰テスト3ケース復元） | `4774666a` |
| H-4 LINE 認証情報の暗号化 | ✅ 対応済み（AES-256-GCM・平文 fallback・機会的再暗号化） | `a620bdfc` |
| H-6 平文パスワードコメント除去 | ✅ 対応済み（**ローテーションは未実施** — C-1 参照） | `aa9c0a5d` |
| M-1/M-2 クレジット訂正（検証正直化＋監査 delta＋締め済み可視化） | ✅ 対応済み（総額変更は現行許容・厳格化は PO 判断待ち） | `9d3df80c` |
| M-8 auth ライフサイクルテスト | ✅ 対応済み（期限は service 層責務であることも確認） | `5b0ac22d` |
| M-9 actionlint バージョン固定 | ✅ 対応済み（v1.7.12） | `ba8cecea` |
| M-12 ピン記法ポリシー文書 | ✅ 対応済み（docs/ci-policy.md・ratchet 運用） | `3f836cda` |
| C-1 ローテーション／#97 実値削除／.env.staging 追跡解除 | ⏳ **ユーザー実施要**（AWS/LINE/GitHub 操作） | — |
| C-2 本番前ブロッカー追跡 | ✅ **代替解決**（2026-07-02）: park 原本6 Issue（#89 #91 #97 #98 #99 #109）を再オープンし残タスクをコメント — 追跡が原本 Issue 上に復元されたため新規傘 Issue は不要（起票は任意） | — |
| H-5 ECS secrets(valueFrom) 化 | ⏳ ユーザー実施要（SSM 登録＋IAM）— C-1 とセット | — |
| H-7 カバレッジ ratchet ゲート | ⏳ 未着手（BE-refactor R3-5 として計画済み・ベースラインは次回 CI 実測） | — |
| M-3/M-4/M-5/M-6/M-7 | ⏳ PO 判断待ち（M-3 のコード内誤参照コメントのみ `4774666a` で訂正済み。M-3=#150 は再オープン済み） | — |
| 仕様未充足 Issue の再オープン | ✅ 11件を再オープン＋残タスクコメント（2026-07-02）: #150 #189 #194 #196 #212 #89 #91 #97 #98 #99 #109 | — |
| M-10/M-11 | ⏳ 方針決定／GitHub Secrets 登録がユーザー側に必要 | — |

---

## 🔴 CRITICAL

### C-1. [#97/#89] 平文シークレットが現在も露出継続 — ローテーション未実施・「削除済み」前提が崩壊

**事象**
シークレット系7 Issue（#89 #91 #96 #97 #98 #99 #109）は2026-06-20に「STG限定運用のため park、本番移行前に再対応必須」として一括クローズされたが、実装による解決は1件もなく、かつ park の前提が現状と食い違っている。

- `.env.staging` は #89 本文で「4c24bd3e で削除済み」とされたが、その後 `ef9e32dd`（chore: track staging environment file）で**再追跡**され、平文の `DB_PASSWORD` / `JWT_SECRET` / `INTEGRATION_ENCRYPTION_KEY` を含んだまま `git ls-files` に現存する（本体セッションで確認済み）。
- Issue #97 の**本文自体に STG の JWT_SECRET / DB_PASSWORD / DB_HOST の実値が記載**されたまま GitHub 上で閲覧可能。git 履歴とは独立した露出経路。
- ローテーション実施の証跡はどこにもない。露出済みの値が現在も有効である可能性が高い。

**影響**: リポジトリ閲覧権限を持つ全員（および漏洩時は第三者）が STG の DB・JWT 署名鍵・統合暗号鍵に到達できる。JWT_SECRET 漏洩は任意ユーザーへのなりすましトークン偽造を許す。INTEGRATION_ENCRYPTION_KEY 漏洩は暗号化済み clinic_integrations の復号を許す。

**対応方法（この順で実施）**
1. **ローテーション**（最優先・順序重要）:
   - RDS: STG DB パスワードを AWS コンソール/CLI で変更 → 新値を配布経路（現状は `.env.staging`、恒久対応は H-5 参照）に反映 → ECS サービス再デプロイ。
   - `JWT_SECRET`: 新値に変更してデプロイ。**全ユーザーの既存セッションが無効化される**ため、STG 利用者に事前周知。
   - `INTEGRATION_ENCRYPTION_KEY`: 変更すると既存の暗号化済みデータ（clinic_integrations の Lステップ認証情報）が**復号不能になる**。手順: 旧鍵で全レコード復号 → 新鍵で再暗号化する一時スクリプトを実行してから鍵を切替（または STG は再入力運用でも可）。
   - LINE channel secret / access token: LINE Developers コンソールで再発行（#96 の平文保存対象でもある）。
2. **Issue #97 本文の実値を削除**: `gh issue edit 97 --body ...` で値をマスクした本文に差し替え。編集履歴にも残るため、ローテーション完了後に実施すれば実害はない（旧値は失効済みになる）。
3. **`.env.staging` の追跡解除**: `git rm --cached .env.staging` + `.gitignore` へ追加。ただし現状 CI（backend-deploy.yml）が `.env.staging` をパースしてタスク定義を生成しているため、**H-5 の対応（SSM/Secrets Manager 化）とセットで行う**こと。単独で追跡解除すると STG デプロイが壊れる。
4. **再発防止**: pre-commit hook または CI に gitleaks / trufflehog 等のシークレットスキャンを追加。`.claude/hooks/` に既存の pre-edit ガードがあるためそこへの追加も可。
5. git 履歴の書き換え（filter-repo）は、ローテーション完了後は必須ではない（失効済みの値しか残らない）。実施するなら全開発者の再クローンが必要になるため、コスト対効果で判断。

**検証方法**: `git ls-files | grep -E '\.env'` で `.env.staging` が消えること。旧 DB パスワード / 旧 JWT でのアクセスが拒否されること。CI シークレットスキャンが green であること。

### C-2. シークレット系 park の追跡担保が消失 — 本番デプロイ時に素通りする構造

**事象**: クローズコメントは再対応の担保先を「#123系の本番前チェックリスト」としたが、#123 は **STG リリース前**チェックリスト（本番用ではない）で既に CLOSED（2026-06-27）、本文に secret/rotation/暗号化への言及はゼロ。open Issue 全件を横断してもシークレット追跡 Issue は存在しない。記録はローカルのメモリノート1件のみでチーム可視ではない。

**影響**: 本番移行の意思決定時に、これら CRITICAL 群が「クローズ済み＝解決済み」として素通りする構造的リスク。

**対応方法**
1. 「🔴 本番デプロイ前ブロッカー: シークレット管理残タスク」として open Issue を新規起票し、以下をチェックリスト化する:
   - [ ] C-1 のローテーション一式（実施済みなら証跡リンク）
   - [ ] #96: line_reservation_settings の暗号化（→ H-4）
   - [ ] #99: ECS タスク定義の secrets(valueFrom) 化（→ H-5）
   - [ ] #98: stg-db-tunnel.sh の平文コメント除去（→ H-6）
   - [ ] #109: CI テスト認証情報の Secrets 化（→ M-11）
   - [ ] #91: .env.production の扱い決定（→ M-10）
   - [ ] 本番用シークレットは STG と別値で新規発行（使い回し禁止）
2. ラベル `production-blocker` を作成して付与。本番デプロイ手順書（または本番用チェックリスト Issue）から相互リンク。
3. クローズ済みの #89/#91/#96/#97/#98/#99/#109 それぞれに新 Issue へのリンクコメントを1行追加し、追跡先を明示。

**検証方法**: `gh issue list --label production-blocker` で残タスクが列挙されること。

---

## 🟠 HIGH

### H-1. [#154] 締め履歴一覧の年月フィルタが完全に機能していない（MISMATCH）

**事象**: 締め履歴ページ（/accounting/close/history）で年・月をどう変えても表示が変わらず、常に「最新20件（全期間）」が表示される。受入条件「Date range filter works correctly」未達。さらにページネーション UI が無いため、**最新20件より古い締め記録には UI 上一切到達できない**。運用が進むほど過去記録が不可視化する。

**根本原因**: FE↔BE のクエリ契約不整合。
- FE は `year`/`month` を送信: `frontend/src/features/cash-register/api/get-cash-register-closes.ts:11-14`（`GetCashRegisterClosesParams { year?, month?, page?, limit? }`）
- BE は `start_date`/`end_date` しかパースしない: `backend/internal/handler/cash_register_request.go:14-24`（`newListCashRegisterClosesQuery` は `values.Get("start_date")`/`values.Get("end_date")` のみ。year/month は無視）
- 結果 `filters.StartDate=nil, EndDate=nil` → `cash_register_close_repository.go` の FindAll が日付条件を適用せず、`backend/internal/handler/query_helpers.go:26-30` のデフォルト `limit=20, page=1` が常に効く
- FE 側クライアントフィルタは period のみ: `frontend/src/features/cash-register/routes/CashRegisterHistoryPage.tsx:63-65`

**対応方法（推奨: 案B — FE を既存 BE 契約に合わせる。BE 変更不要で影響範囲最小）**
1. `CashRegisterHistoryPage.tsx` で選択中の年月から `start_date`（月初 YYYY-MM-01）と `end_date`（月末）を導出し、`useGetCashRegisterCloses({ start_date, end_date, page, limit })` に変更する。
2. `get-cash-register-closes.ts` の `GetCashRegisterClosesParams` を `{ start_date?: string; end_date?: string; page?: number; limit?: number }` に改める（year/month を廃止）。
3. **ページネーション追加**: BE は既に `page`/`limit`/`total` を返す（query_helpers 既存実装）。FE でページャ UI を追加し `total` を消費する。月別フィルタが効けば1ヶ月分は通常20件以内（AM/PM/EMG×日数）だが、月をまたぐ閲覧・将来の複数レジ対応を考慮してページャは付けておく。
4. 代替案A（BE で year/month→月初/月末変換を追加）は、契約が二重になり同じ乖離を再生産しうるため非推奨。
5. **回帰防止**: handler テストに「start_date/end_date 指定で範囲外レコードが返らない」ケースを追加。FE は `CashRegisterHistoryPage` のクエリパラメータ生成をユニットテスト化（year/month→日付文字列の境界: 12月→翌年、うるう年2月）。

**検証方法**: `docker compose exec backend go test ./internal/handler/... -run CashRegister`（scoped）。FE は `npx vitest run src/features/cash-register`。ブラウザで年月変更→一覧が切り替わること、過去月の締めに到達できること。

### H-2. [#155] 月次レポート→締めドリルダウンが過去日で空振り（H-1 の連鎖障害）

**事象**: 月次集計の日次行クリック → `/accounting/close/history?date=YYYY-MM-DD` 遷移とハイライト機構自体は実装済み（`frontend/src/features/accounting-reports/routes/AccountingReportsPage.tsx:59`、`CashRegisterHistoryPage.tsx:20-24,79-84`）。しかし H-1 により BE が常に最新20件しか返さないため、そこに含まれない日付へのドリルダウンは対象行が一覧に存在せずハイライトが空振りする。当月でも締め件数が20件を超えると発生。

**根本原因**: H-1 と同一（`?date=` から年月 state を初期化しても、その年月がクエリに反映されない）。

**対応方法**: H-1 の修正で自動解消する。追加で以下を確認:
1. `?date=` で指定された日付の月の start_date/end_date がクエリに載ること。
2. H-1 のページネーション導入後、`?date=` 指定時は該当日を含むページを初期表示する（または該当日で `start_date=end_date=date` の1日絞りにする方が単純で確実）。
3. E2E またはコンポーネントテストで「過去月の日付でドリルダウン→該当行がハイライトされる」ケースを追加。

**検証方法**: H-1 と同じ。加えて月次レポートから2ヶ月前の日付をクリックして該当締め行が表示・ハイライトされること。

### H-3. [#197] レジ締め理論現金と月次レポートで「現金」の定義が食い違う回帰

**事象**: 同一データに対して、レジ締めの理論現金と月次レポートの現金集計が食い違う。`payment_method_id = NULL` の現金 split（#128 修正以前に作成されたレガシー行・seed 行）が、月次レポートでは「現金」に計上されるのに、レジ締め理論現金からは除外される。締め画面の現金差額（実際現金−理論現金）が実態より大きく表示され、月次との突合も合わない。

**根本原因**: #197（system_key 化）の際に `calcTheoreticalCash` のシグネチャを `*uint64`→`uint64` に変更し、#128 が確立した「NULL は現金として加算」の後方互換を削除した。月次側は残存しているため非対称になった。
- レジ締め（NULL除外）: `backend/internal/service/cash_register_service.go:444-452` — `if r.PaymentMethodID != nil && *r.PaymentMethodID == cashMethodID` のみ加算（本体セッションで確認済み）
- 月次レポート（NULL=現金）: `backend/internal/service/accounting_report_service.go:496-500` — `resolvePaymentMethodName` が `id == nil` → "現金"（本体セッションで確認済み）
- 発現データ実在: `backend/migrations/003_seed_demo.sql:3624` の payment_splits `(1, 20, 'cash', NULL, 5000, ...)`
- テスト退行: `backend/internal/service/cash_register_service_test.go` の `TestCalcTheoreticalCash` から「NULL行も現金として加算」「NULL+現金マスタid混在で両方加算」「cashMethodID nil 時は NULL のみ現金」の回帰防止3ケースが削除済み

**対応方法（推奨: 短期A＋恒久B の2段階）**
- **短期（案A・即日可能）**: `calcTheoreticalCash` に NULL=現金の後方互換を復帰する。
  - シグネチャを元に戻すか、`if r.PaymentMethodID == nil || *r.PaymentMethodID == cashMethodID` とする（NULL 行は #128 以前の書込みで method='cash' の場合のみ存在するため NULL=現金は安全。ただし念のため集計クエリ側で `method='cash'` も条件に含まれているか `PaymentAggregateRow` の生成元を確認すること）。
  - **削除された回帰テスト3ケースを復元**する。これが本回帰の再発防止の本体。
- **恒久（案B・別Issue起票）**: additive migration でレガシー/seed の `payment_splits.payment_method_id` を backfill する。
  - 方針: `payment_splits.method`（ENUM）→ 同一クリニックの `payment_methods.system_key` 一致行の id を設定。クリニック特定は payment_splits→payments→billings.clinic_id の JOIN 経由（実カラム名は 001_init.sql で要確認）。
  - seed（003/004）の NULL 行も同時修正。**適用済み migration の編集は checksum mismatch を起こすため、既存ファイル編集ではなく新規 migration として追加し、STG は db_reset か新規適用で対応**（プロジェクト運用ルールどおり）。
  - backfill 完了後も、防御として NULL ハンドリング（案Aのコード）は残してよい（デッドコード化するが安全側）。
- 月次側 `resolvePaymentMethodName` の NULL→現金は据え置き（backfill 後は到達しなくなる）。

**検証方法**: `docker compose exec backend go test ./internal/service/... -run 'TheoreticalCash|CashRegister'`。seed 適用環境で「レジ締め理論現金」と「月次レポート現金合計」が同日で一致すること。

### H-4. [#96] line_reservation_settings のシークレットが依然として平文保存

**事象**: LINE予約連携の `line_channel_secret` / `line_access_token` が DB に平文で保存されている。`json:"-"` は API 応答に出さないだけで、DB バックアップ・ログ・SQL アクセスからは平文で読める。同水準の秘匿情報である clinic_integrations（Lステップ）は AES-256-GCM で暗号化済みであり、保護レベルが非対称。

**根本原因**: 暗号化未実装のまま park クローズ。
- 平文モデル: `backend/internal/model/line_reservation_setting.go:33,35`（`string` + `json:"-"`）
- 対照（暗号化実装の参照先）: `backend/internal/service/lstep_settings_update.go:31`（encrypt）/ `lstep_settings_credentials.go:37`（decrypt）
- DDL 平文 text 列: `backend/migrations/001_init.sql:2609,2611`、seed に平文値: `003_seed_demo.sql:2385-2396`

**対応方法**
1. `line_reservation_setting_service.go` の保存経路で、clinic_integrations と同じ暗号化ヘルパー（INTEGRATION_ENCRYPTION_KEY / AES-256-GCM）を用いて `line_channel_secret` / `line_access_token` を暗号化してから永続化する。読出し経路（LINE webhook 署名検証・Messaging API 呼び出し）で復号する。カラム型は text のままで可（暗号文格納）。
2. 既存の平文レコードの移行: 起動時 migration スクリプトまたは一時管理コマンドで「平文→暗号化」変換を1回実行する。暗号文かどうかの判定はプレフィックス（clinic_integrations の既存実装の形式に合わせる）で行い、冪等にする。STG はデータ量が少ないため再入力運用でも可。
3. seed（003/004）の平文値はダミー値である旨をコメント明記するか、seed 投入後に暗号化変換が走る形にする（seed に本物の認証情報を置かないことが本質）。
4. C-1 のとおり、現在平文で入っている実認証情報は LINE Developers 側で再発行する。
5. 回帰防止: 「line_reservation_settings に保存される secret/token が平文で SELECT できない」ことを検証するテストを追加（暗号化プレフィックス検証で十分）。

**検証方法**: scoped テスト＋STG で LINE 予約フロー（webhook 署名検証含む）の疎通確認。DB 直読みで暗号文であること。

### H-5. [#99] backend-deploy.yml が secrets を ECS environment に平文展開（意図的継続）

**事象**: デプロイワークフローが `.env.staging` を丸ごとパースし、ECS タスク定義の `environment`（平文）に全キーを展開、`secrets` を明示的に `[]` にしている。JWT_SECRET / DB_PASSWORD 等がタスク定義メタデータ・ECS コンソール・CloudTrail・`describe-task-definition` 応答に平文露出する。

**根本原因**: `.github/workflows/backend-deploy.yml:123-137`（.env パース）、`:151`（`.containerDefinitions[0].environment = $envvars`）、`:152`（`.containerDefinitions[0].secrets = []`）、コメント `:140`「env 全置換・secrets 撤廃」— 仕様（valueFrom 方式）の逆を明示採用。migrate タスク（`:190`）も同様。

**対応方法**
1. シークレット類（DB_PASSWORD / JWT_SECRET / INTEGRATION_ENCRYPTION_KEY / LINE系 / Lステップ系）を SSM Parameter Store（SecureString）に登録する。命名例: `/animal-ekarte/stg/DB_PASSWORD`。
2. タスク定義の `secrets` に `{"name": "DB_PASSWORD", "valueFrom": "arn:aws:ssm:...:parameter/animal-ekarte/stg/DB_PASSWORD"}` 形式で移す。非シークレット設定（PORT、フラグ類）のみ `environment` に残す。
3. ECS の **taskExecutionRole** に `ssm:GetParameters`（+ KMS 復号）権限を追加。
4. `backend-deploy.yml` の jq 加工を「非シークレット env の置換」と「secrets 配列の維持」に分離。`.env.staging` からシークレット行を撤去（→ C-1 の追跡解除とセットで完了）。migrate タスク定義も同様に修正。
5. Terraform 管理下なら IaC 側で定義（STG P2 で Terraform 使用実績あり）。
6. 本番環境は最初からこの方式で構築し、平文 environment 方式を持ち込まない。

**検証方法**: デプロイ後 `aws ecs describe-task-definition` の出力に平文シークレットが含まれないこと。アプリ起動・DB接続・JWT 発行が正常であること。

### H-6. [#98] scripts/stg-db-tunnel.sh に平文 DB パスワードコメントが現存

**事象/根本原因**: `scripts/stg-db-tunnel.sh:4` に DB パスワードの平文コメントが残置。初回コミット以降一度も修正されていない。#97 の DB_PASSWORD と同一値の可能性が高い。

**対応方法**: (1) C-1 のローテーション後にコメント行を削除（ローテーション前に消しても値は失効しないため、必ずローテーションとセット）。(2) 代替として「パスワードは 1Password / SSM から取得」の参照コメントに置換。(3) git 履歴からの除去は C-1-5 と同じ判断（ローテーション済みなら必須ではない）。

**検証方法**: `grep -i password scripts/stg-db-tunnel.sh` で実値が出ないこと。

### H-7. [#212] 「全パッケージ90%」目標を主要3層未着手のままクローズ — カバレッジゲートも不在

**事象**: 達成は config 92% / errors 100% / middleware 92.9% の基礎3パッケージのみ。handler（57.9%起点）/ service（59.1%起点）/ repository（11.8%起点）は「今後段階的に拡充」と明示先送りでクローズ。加えて「Phase 3 ゲート値クリア＋CI マージ基準化」も未達 — `.github/workflows/ci.yml:243` は非ゲート維持で、**カバレッジが下がっても誰も気付けない**。

**根本原因**: 受入基準（全パッケージ90%・CIゲート化）に対して部分達成でクローズする運用判断。先送り自体は透明に記録されているが、追跡 Issue と自動ゲートの両方が無い。

**対応方法**
1. **先にゲートを入れる**（テスト追加より優先度が高い。ゲートが無い限り既存カバレッジも守れない）: `docs/coverage-policy.md:53-88` の Phase 1（warn）→ Phase 2（現状値を下回ったら fail する ratchet 方式）を CI に実装。90% という遠い目標より「現状から下げない」ゲートが実効的。
2. `docs/coverage-policy.md:49` のベースライン数値プレースホルダを実測値で埋める（M-13/LOW #194 と同一作業）。
3. handler / service / repository の拡充は「全パッケージ90%」ではなく、パッケージ別の現実的な中間目標（例: service 75% → 85%）を刻んだ追跡 Issue を起票する。repository はDB結合テストが必要でコストが高いため、clinic_id 隔離テスト群（#196 完了分）を土台に CRUD 正常系から積む。
4. クローズ済み #212 に追跡 Issue へのリンクコメントを追加。

**検証方法**: カバレッジを意図的に下げる PR で CI が warn/fail すること。

---

## 🟡 MEDIUM

### M-1. [#189] クレジット訂正の整合検証が恒真化 — 訂正額の妥当性を実質検証していない

**事象**: 確定後クレジット訂正 API（POST /:id/credit-correction）で、訂正後の支払内訳合計と請求額の整合チェックが**常に成立してしまい、機能していない**。どんな金額を入れても総額検証を素通りし、`billing_amount` が「入力値の合計」で silent に再定義される。クレジット打ち間違い「防止」機能なのに、訂正入力自体の打ち間違いは検出できない。

**根本原因**: `backend/internal/service/accounting_service_correction.go:96-103` — `validatePaymentSplits(toValidationInputs(corrected), &newBillingAmount)` に渡す `newBillingAmount` が corrected 内訳の合計そのもの（:97-99 で `for ... newBillingAmount += corrected[i].Amount`）。検証内部の総額一致チェック（`accounting_service_builders.go:106-108` `total != *billingAmount`）が `sum != sum` となり常に偽（本体セッションで確認済み）。完了時（`accounting_service_core.go:107`）は「本来の請求額（保険・割引反映後）」に照合しており、訂正経路だけこの不変条件を破っている。

**対応方法**
1. **最小修正**: 訂正前の `payment.BillingAmount`（本来の請求額）を検証に渡す:
   `validatePaymentSplits(toValidationInputs(corrected), &payment.BillingAmount)` とし、訂正後合計が本来の請求額と一致することを要求する。これで「カード端末に打つ金額を間違えた（内訳間の配分ミス）」の訂正は通り、「合計が変わる訂正」は 400 で弾かれる。
2. **仕様確認が必要な論点（PO判断）**: 「実際にカード端末で誤った金額を決済してしまい、受領総額自体が請求額と異なる」ケースを訂正で表現したいのか。もし必要なら、`allow_total_change: true` のような明示フラグ＋変更前後の差額を監査ログ（`:151-175` の before/after に加えて delta）へ記録する設計にし、silent な総額変更は禁止する。デフォルトは案1の厳格検証。
3. 回帰テスト: 「訂正後合計 ≠ 請求額 → エラー」「配分変更のみ（合計不変）→ 成功」の2ケースを `accounting_service_correction` のテストに追加。

**検証方法**: `docker compose exec backend go test ./internal/service/... -run CreditCorrection`

### M-2. [#189] クレジット訂正に締め済み期間への波及ガードが無い

**事象**: 日次・月次売上集計は `SUM(payment_splits.amount) WHERE status=completed`（`accounting_repository_reports_daily.go:28,41` / `_monthly.go:42`）で算出される。訂正は `SavePaymentSplits`（`accounting_service_correction.go:113`）で金額を直接書き換えるため、**締め確定済み期間の売上を、再集計も警告もなしに silent に改変できる**。締め時点の帳票と事後の月次が食い違い、原因追跡は監査ログを掘るしかない。

**根本原因**: PATCH 編集経路には締め済み判定がある（`accounting_handler.go:160-165` の isClosed 分岐）が、credit-correction 経路には一切ない非対称。受入条件「訂正が月次集計・締め整合と矛盾しないことを検証」がコードで担保されず運用へ先送りされた。

**対応方法**
1. credit-correction のサービス層に PATCH 経路と同じ締め済み判定を追加する（対象 billing の日付＋区分が確定済み `cash_register_closes` に含まれるかを照会）。権限は既に post-close-edit を要求しているため**拒否ではなく明示化**が目的:
   - 締め済み期間への訂正の場合、監査ログに `post_close: true` と対象締めID を記録する。
   - FE の `CreditCorrectionDialog` に「この会計は締め確定済み期間です。訂正すると締め時点の帳票と差異が生じます」の警告を表示する。
2. 締め済み期間の訂正発生を検出可能にする: 締め詳細画面に「締め後訂正あり」バッジ、または月次レポートに訂正件数を出す（#115 締め後編集と同じ扱いに揃える）。
3. 再締め（締めのやり直し）を許すかは PO 判断。少なくとも差異の可視化までは実装すべき。

**検証方法**: 締め確定済み日の会計に訂正 → 監査ログに post_close フラグと締めIDが記録され、FE に警告が出ること。

### M-3. [#150/#151] 越日 EMG（18:30〜翌8:59）が未実装のまま追跡 Issue が消失（孤児化）

**事象**: EMG 区分の enum・検証・UI は全層実装済みだが、仕様要件「EMG=18:30〜翌8:59:59（日をまたぐ）」は未実装。`cash_register_service.go:401-407` は EMG を pmEnd〜当日24:00（同日内）で集計するため、**深夜0:00〜8:59 の緊急会計は翌日の AM 区分に流入する**。同様に #151 の「AM は 9:00 開始フロア」も未実装（現状 AM は 0:00 開始）。

**根本原因**: コードと #151 本文は「AM開始時刻フィールドは #156 依存」と記載するが、#156 の実スコープは「クリニック別スコープ検証」であり該当機能を扱わずクローズ済み。誤参照により、未達機能を追跡する Issue が存在しなくなった。

**対応方法**
1. **PO に越日 EMG の要否を確認**した上で、必要なら新規 Issue を起票する。実装スコープ:
   - `clinic_settings` に AM 開始時刻（am_start、既定 9:00）を追加（additive migration）
   - `resolvePeriodRange` の EMG を「当日 pmEnd 〜 翌日 am_start−1秒」の2区間 or 越日レンジで集計（集計クエリの日付境界処理に注意）
   - AM を「am_start〜boundary」に変更（0:00〜am_start は前日 EMG に帰属）
   - `closing-time-ranges.ts` のプレビュー表示を越日表記（「18:30〜翌8:59:59」）に対応
   - 既存の締め記録は再計算しない（過去データ非破壊）ことを明記
2. 実装しない判断なら、#150/#151 の本文の「#156 依存」記述を「越日 EMG は未実装・現行は同日内集計」と訂正し、仕様として確定させる。
3. いずれの場合も `cash_register_service.go:399-401` 付近のコメント（#156 参照）を実態に合わせて修正。

**検証方法**: 実装時は「23:00 の会計が当日 EMG に、翌朝 7:00 の会計が前日 EMG に計上される」table-driven テスト。

### M-4. [#160 vs #211] 健診の「Examination ⇔ Checkup」二重管理がアーキテクチャ矛盾として現実化

**事象**: #160 は案B（Checkup 拡張）を「検査機能と機能重複・データ二重管理になる」と明示却下し、既存 Examination へのマッピング（`003_seed_demo.sql:4650-4698` の exam_types 12000-12003 ＋ fields 45-58）でクローズした。しかし後続の #211（健診パッケージ）が `backend/migrations/010_add_checkup_packages.sql` で `checkup_type_fields` / `checkup_field_results`（exam_type_fields と同型の型付き子フィールド機構、本体セッションで確認済み）を追加した。**#160 が回避したはずの二重管理が現実に発生**しており、同じ「歯科健診の歯石付着度」が exam 系（field id 45-58）と checkup 系（010 の seed）の両方に存在する。

**根本原因**: #160 のクローズ判断（案A採用）と #211 の実装（実質案B）が正面衝突。product 決定が未収束のまま両方が本流に入った。

**対応方法**
1. **PO 再確認が最優先**。論点: 健診の記録・集計は Examination 系（検査タブ）と Checkup 系（健診パッケージ）のどちらを正とするか。
2. 決定後の片付け:
   - Checkup 系を正にする場合 → #160 で投入した exam_types 12000-12003 / exam_type_fields 45-58 の seed を撤去（または非表示化）し、#160 のクローズコメントに決定変更を追記。
   - Examination 系を正にする場合 → #211 の checkup_type_fields 機構の利用を停止（migration 010 は未適用なら取り下げ可能。既適用なら機能フラグで封印）。
3. 決定を ADR（docs/adr/）として記録し、#160 と #211 の両方からリンクする。二重入力が始まってからの統合はデータ移行を伴い高コストになるため、**運用開始前に決めること**。

**検証方法**: 決定後、健診記録の入力導線が1系統のみであることを UI で確認。

### M-5. [#178] クライアント原文要求が未実装のまま代替案でクローズ — 差し戻しリスク

**事象**: クライアント要求は「(2) 同居の子も**カルテ上で**確認でき、クリックで**その子のカルテへ遷移**」。現実装は (a) 別窓の飼主レポートで、(b) PetSwitcher は**レポート内の表示切替でありカルテには遷移しない**。カルテ本体（`frontend/src/features/medical-records/`）に同居ペット表示は存在しない（grep 0件）。#178 本文自身が「この代替は要望を満たさない可能性が高い」と警告していた案でクローズされている（PO 明示判断のためクローズ自体は正当）。

**対応方法**
1. 次回クライアント確認の場で「飼主レポートで代替した」ことを明示し、原文要求（カルテ上表示＋カルテ間遷移）の要否を確定させる。**確認前に実装しない**（推測実装禁止の方針どおり）。
2. 原文要求が生きていた場合の実装スコープ（見積もり用）:
   - カルテ画面のヘッダー/サイドに同居ペット一覧（`useGetPets(ownerId)` 再利用）を表示
   - クリックで「そのペットの最新カルテ or カルテ一覧（petId フィルタ）」へ遷移（新規カルテ作成に飛ばすかは要確認）
   - 遷移先の権限・clinic_id 隔離は既存の medical-records ルートのガードをそのまま通る設計にする
3. 確認結果を #178 にコメントで追記（差し戻しなら新規 Issue 起票）。

### M-6. [#158] GET /owners/:id/medication-history が孤児コード

**事象**: BE の投薬歴エンドポイント（`backend/internal/handler/handler.go:149` ルート、service `medical_record_report.go:27-54`、repository `FindOwnerMedicationHistory`、テスト完備）は、仕様再定義（飼主レポートは per-pet の `TreatmentHistorySection(filter="medicine")` 表示に変更）により**フロントから一切呼ばれていない**（`grep -rn "medication-history" frontend/src` → 0件、本体セッションで確認済み）。動作・IDOR防御・clinic 隔離は正しいが、保守対象として残り続ける。

**対応方法（どちらかを明示的に選ぶ）**
- **削除する場合**: ルート（handler.go:149）、handler メソッド、service、repository メソッド、対応テストを一括削除。openapi 定義があれば同期（`make codegen` 対象確認）。
- **残す場合**: 「飼主横断の投薬歴ビュー（全ペット横断1テーブル）を将来提供する際の API」として handler にコメントを付し、#158 のクローズコメントに「BE のみ先行実装・FE 未配線」と追記する。
- 判断基準: 飼主レポートに「全ペット横断の投薬歴タブ」を追加する構想が生きているなら残す。無いなら削除（YAGNI）。

**検証方法**: 削除時は `docker compose exec backend go test ./internal/...`（scoped）と `grep -rn medication-history backend/` が空になること。

### M-7. [#191] 締めレジ確定パスの税率分類が固定閾値 `>8` のまま — 全帳票一貫適用が未達

**事象**: 月次レポート経路は病院マスタ税率基準（exact-match、`accounting_report_service.go:82-98`）へ移行済みだが、`buildTaxBreakdown`（`accounting_report_service.go:250-262`、**締めレジ確定パスと共有**）は `TaxRate > 8` の固定閾値のまま意図的に据え置かれた。親 #179 ②の「病院マスタ税率を更新するだけで全帳票へ一貫適用」は締めレジサマリーで未達。軽減税率が 8% 超（例 9%）へ改定された場合、締め集計で軽減を標準へ誤分類する。

**根本原因**: #191 が締めレジ挙動の不変性を優先して明示的に据え置き決定（判断自体は妥当）。ただし残債の追跡 Issue が無い。

**対応方法**
1. 追跡 Issue を起票: 「buildTaxBreakdown を clinicTaxRates 基準へ移行（締めレジ経路）」。
2. 実装方針: `buildTaxBreakdown` に `clinicTaxRates` を引数注入し、月次経路と同じ exact-match 分類へ統一。締めレジ経路の呼び出し元（`cash_register_service.go`）で clinic 設定を取得して渡す。
3. **過去の締め記録は再計算しない**（保存済みスナップショットは不変）ことを仕様に明記。切替後の新規締めのみ新分類。
4. 回帰テスト: 標準10%/軽減8%の既定クリニックで新旧分類が一致すること（挙動不変の証明）＋ 軽減9%設定クリニックで正しく分類されること。

**検証方法**: `docker compose exec backend go test ./internal/service/... -run 'TaxBreakdown|CashRegister'`

### M-8. [#196] auth 永続化テスト欠落 — クローズ理由がカテゴリ誤認

**事象**: 受入基準「password_reset_token_repository（期限/単回使用）と token_blacklist_repository（失効）のテスト追加」が未達。`backend/internal/repository/` に両 repository の実装ファイルは存在するが**テストファイルは0件**（本体セッションで確認済み）。クローズコメントは「clinicScope 除外リスト該当のため対象外」と正当化するが、P3 要件は**認証トークンのライフサイクル検証であり clinic_id 隔離とは別カテゴリ**。パスワードリセットの期限切れ・使い回し、失効トークンの拒否という認証セキュリティ不変条件が未検証。

**対応方法**
1. `password_reset_token_repository_test.go` を追加。最低限のケース:
   - 有効期限内トークンの取得成功 / 期限切れトークンの取得失敗
   - 使用済みマーク後の再利用が失敗（単回使用）
   - 存在しないトークンで失敗
2. `token_blacklist_repository_test.go` を追加:
   - ブラックリスト登録済みトークンの照会が「失効」を返す
   - 期限切れエントリのクリーンアップ（実装にあれば）
3. 既存の repository テスト基盤（setupTestDB、DROP 順序罠に注意 — #196 の既知ノウハウ）をそのまま使う。
4. 可能なら service/middleware 層の統合ケース（失効トークンでのリクエストが 401）も1本。

**検証方法**: `docker compose exec backend go test ./internal/repository/... -run 'PasswordReset|TokenBlacklist'`

### M-9. [#193] actionlint が main ブランチから未ピン fetch（サプライチェーンリスク）

**事象/根本原因**: `.github/workflows/actionlint.yml:20` が `bash <(curl -Ls https://raw.githubusercontent.com/rhysd/actionlint/main/scripts/download-actionlint.bash)` を実行（本体セッションで確認済み）。上流 main が改竄・破壊されると CI で任意コード実行になり、再現性も無い。

**対応方法（いずれか）**
1. **バージョン固定（最小変更）**: download スクリプトはバージョン引数を受け付ける。タグ付きの raw URL＋バージョン指定に変更する:
   `bash <(curl -Ls https://raw.githubusercontent.com/rhysd/actionlint/v1.7.7/scripts/download-actionlint.bash) 1.7.7`（バージョンは最新安定を確認して固定）。
2. **より堅くするなら**: リリースバイナリを直接 DL し sha256 検証してから実行。またはメンテされている action（SHA ピン）を利用。
3. Renovate/Dependabot の監視対象にならない形式のため、四半期ごとのバージョン見直しを M-12 のピンポリシー文書に含める。

**検証方法**: workflows/** を触る PR で actionlint ジョブが従来どおり lint 結果を返すこと。

### M-10. [#91] frontend/.env.production が依然 git 追跡中・デモアカウント本番無効化なし

**事象**: `frontend/.env.production` は追跡継続（本体セッションで確認済み）。中身は `VITE_API_URL`（STG ALB endpoint）と `VITE_SHOW_DEMO_ACCOUNTS`。ALB endpoint の秘匿性は低いが Issue の除去要件は未達。`.gitignore:1-4` には「.env.production は git track する」と Issue と**逆方向のコメント**があり、方針が不整合のまま。

**対応方法**
1. **方針を明文化して二択を確定**:
   - (a) 「VITE_* はビルド時公開値でありシークレットではない」として追跡を継続する → `.gitignore` のコメントを根拠付きで残し、#91 に決定コメントを追記してクローズ状態を正当化。**シークレットは絶対に置かない**ルールを CI スキャン（C-1-4）で担保。
   - (b) 追跡を外し、デプロイワークフローの env 注入（frontend-deploy.yml）へ移す。
   - 推奨は (a)＋スキャン担保。VITE_* はバンドルに埋め込まれる公開値であり、隠す意味が薄い。
2. `VITE_SHOW_DEMO_ACCOUNTS` は**本番ビルドで必ず false** になる仕組みを入れる: 本番用 env ファイル分離（`.env.production` は本番、STG は `--mode staging` で `.env.staging` を使う Vite 標準方式へ整理）または CI の本番デプロイジョブで明示上書き。現状「production という名前のファイルが STG 設定」になっている命名の歪みも同時に解消する。

**検証方法**: 本番ビルド成果物で `grep -r SHOW_DEMO` した際に false であること。デモアカウント欄が本番で非表示なこと。

### M-11. [#109] CI ワークフローの平文テスト認証情報が残存

**事象/根本原因**: `.github/workflows/performance-tests.yml:89-90,99-100` に `TEST_EMAIL: admin@example.com` / `TEST_CRED` がハードコード残存。仕様が要求した `${{ secrets.CI_TEST_EMAIL }}` / `${{ secrets.CI_TEST_PASSWORD }}` への置換は未実施。

**対応方法**
1. リポジトリ Secrets に `CI_TEST_EMAIL` / `CI_TEST_PASSWORD` を登録し、workflow の env を `${{ secrets.* }}` 参照へ置換（2箇所）。
2. **このテストアカウントのパスワードが STG のデモ/実アカウントと同一でないことを確認**する（同一なら C-1 に含めてローテーション）。テスト専用アカウントは最小権限にする。
3. seed のデモアカウント認証情報と CI の認証情報を意図的に分離し、seed 側はデモ専用と明記。

**検証方法**: performance-tests ワークフローの手動 dispatch が green のこと。workflow ファイルに平文credが無いこと。

### M-12. [#195] Actions ピン記法ポリシーが未決定・文書なし

**事象**: 主目的（setup-node@v6 統一・同一 action の単一バージョン収束）は達成済み。しかし受入基準「ピン記法ポリシー決定＋全 uses: 準拠＋文書化」が未消化で、完全 semver（`@v7.0.0`）/ メジャータグ（`@v6`）/ SHA ピン（`security-scan.yml:41`）が混在。

**対応方法**
1. ポリシーを決定して `docs/ci-policy.md`（新規・10行程度で十分）に明文化する。推奨基準:
   - GitHub 公式 actions（actions/*）: メジャータグ（`@v6`）— 公式の改竄リスクは低く、パッチ追従の利便を優先
   - サードパーティ actions: **コミット SHA ピン**＋コメントでバージョン明記（サプライチェーン対策）
   - シェルからの外部スクリプト取得（M-9 の actionlint 等): バージョンタグ固定必須
2. 全8ワークフローの `uses:` を決定基準に合わせて一括修正（`grep -rn "uses:" .github/workflows/` で棚卸し）。
3. actionlint ジョブ（#193）がある限り記法の破壊は検出されるため、ポリシー逸脱の検出は四半期レビューで運用。

**検証方法**: actionlint green＋全ワークフローの手動確認。

---

## 🟢 LOW（記録のみ・対応は任意）

| Issue | 内容 | 推奨対応 |
|---|---|---|
| #194 | 受入基準「現行ベースライン数値の記録」未達。`docs/coverage-policy.md:49` が「(CIにて測定予定)」プレースホルダのまま | 次回 CI 実行の coverage artifact から実数を転記（H-7-2 と同一作業） |
| #182 | クローズコメント本文「本Issueはクローズしない（OPEN維持）」と実際の CLOSED 状態が矛盾 | #182 に「子 #188/#189 完了に伴いクローズ」の訂正コメントを1行追加 |
| #123 | P2-4「change_amount 必須化」は実装が `binding:"min=0"` 非負強制のみで必須ではない | 現仕様として受容済み（チェックリストコメントに明示あり）。対応不要 |
| #187 | 「印刷面はちょうど1件」は既定値依存で構造保証ではない。全ポータル `active={false}` 時に白紙化しうる | JSDoc 文書化済み・呼び出し側責務。新規 PrintPortal 追加時のレビュー観点として維持 |
| #179 | クローズコメントの migration 番号（010/008）が統合後の実体（001_init.sql）と乖離 | 文書ドリフトのみ。必要なら #179 にコメント1行追記 |
| #196 | pet CRUD テストが R/D 止まり（Create/Update なし）。medical_record_image_repository 未カバー | M-8 のテスト追加時に同梱すると効率的 |
| #151 | PM終了 18:29:29 を「−1秒ルール」で :59 として実装（mockup 誤記と判断・コメント明記済み） | 妥当。対応不要 |
| #153 | 統合締めテーブルの件数（billingDetails）と金額（categories）がサーバ別系統値でクライアント突合なし（コード内明記・整合はサーバ責務） | 妥当。対応不要 |
| #153 | 参照PDFの視覚レイアウトは Issue 添付欠落のため未検証（列は既存2表の和集合から導出） | 実運用で帳票様式の指摘があれば対応 |
| #191 | 月次 exact-match 化により 0%（非課税）行が旧「軽減」→「標準」バケットへ移動する挙動変化 | 非課税を別バケットにするかは M-7 の Issue 内で合わせて判断 |
| #184 | 実ブラウザ印刷の視覚検証は JSDOM 制約で未実施 | Playwright 導入時に印刷ビューのビジュアルリグレッションを追加 |

---

## 本監査の対象外とした既知の横断事項

- **監査ログ書込がコードベース全体で tx 外の best-effort**（`auditRepository.Create` が dbOrTx 非使用）: #189 等の個別 Issue の受入条件とは独立した横断的既知事項のため、各 Issue の判定には含めていない。#211 系の follow-up 作業（audit-tx 原子化: refund/checkup_field_results は対応済み、残りは横断展開中）で別途追跡されている。
- **検証手法の限界**: 静的コードリーディングのみ。テスト実行・DB照会・実ブラウザ検証は未実施。「テストが存在する」ことは確認済みだが「テストが通る」ことは未確認。#212 のカバレッジ実数も実測していない。

---

## 推奨アクション（優先順・依存関係つき）

| # | 対応 | 対象 | 依存 |
|---|---|---|---|
| 1 | シークレット即時ローテーション＋Issue #97 実値削除 | C-1 | なし（最優先） |
| 2 | 本番前ブロッカー追跡 Issue 起票 | C-2 | なし |
| 3 | 締め履歴フィルタ修正＋ページネーション | H-1/H-2 | なし（FE のみで完結可能） |
| 4 | 現金集計非対称の解消＋回帰テスト復元 | H-3 | なし（短期案Aは即日可） |
| 5 | ECS secrets(valueFrom) 化 → .env.staging 追跡解除 | H-5 → C-1-3 | 1 の後（新値で登録） |
| 6 | line_reservation_settings 暗号化 | H-4 | 1（LINE credentials 再発行）と同時が効率的 |
| 7 | クレジット訂正の請求額照合＋締め済みガード | M-1/M-2 | 検証仕様は PO 確認1点あり |
| 8 | 健診二重管理の PO 決定 | M-4 | PO（運用開始前に必須） |
| 9 | カバレッジ ratchet ゲート導入＋ベースライン記録 | H-7/LOW#194 | なし |
| 10 | 孤児化残課題の Issue 起票 or 本文訂正（越日EMG・税率閾値・medication-history） | M-3/M-7/M-6 | PO 確認 |
| 11 | CI 衛生（actionlint ピン・テスト認証情報 Secrets 化・ピンポリシー文書・auth テスト） | M-8〜M-12 | なし |
