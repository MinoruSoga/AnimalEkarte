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

## 対応状況（2026-07-02 更新 — 対応済み項目は本書から削除済み）

対応済み・削除済み: H-1/H-2 (`a45da439`) / H-3 (`4774666a`) / H-4 (`a620bdfc`) / H-6 (`aa9c0a5d`) / M-1/M-2 (`9d3df80c`) / M-8 (`5b0ac22d`) / M-9 (`ba8cecea`) / M-12 (`3f836cda`)、C-2（park 原本6 Issue #89 #91 #97 #98 #99 #109 の再オープンで代替解決）、仕様未充足 Issue 11件の再オープン。詳細は各コミットと再オープン済み Issue のコメントを参照。**全コミット push 未実施**。

### PO 決定記録（2026-07-02・全7論点回答済み）

| 論点 | 決定 | 実施状況 |
|---|---|---|
| M-4 健診系統 | **Checkup 系（#211）を正とする** | ✅ seed 撤去 `406c6264`・ADR-004 `2e09391c`・#160/#211 コメント済み。残=**次回 STG デプロイ db_reset=true のみ** |
| M-1 訂正検証 | 再判断（2026-07-02）で **(c) 現状維持を最終仕様化**（単一 split API×合計不変条件により厳格化は機能死のため不採用） | ✅ 確定（`9d3df80c` の挙動が最終仕様・コード内コメント更新・#189 に決定記録。FE 警告等の残タスクは #189 で追跡） |
| M-3 越日 EMG | **実装する** | ✅ 実装済み（`9ab95845`・#215。migration 011 + resolvePeriodRange 越日化 + FE 越日表記。push 済み） |
| M-6 孤児 API | **削除する** | ✅ 削除完了（route/handler/service/repository/テスト一式。本書から削除） |
| M-5 #178 同居ペット | クライアント確認まで保留。PO 認識では現行の飼主レポート（PetSwitcher）で充足 | クローズ（追加実装なし。差し戻し時は #178 から再起票） |
| M-7 税率分類 | 起票＋即実装 | ✅ 実装完了（buildTaxBreakdown を病院マスタ exact-match に統一・回帰テスト付き。本書から削除） |
| M-10 .env.production | **追跡継続を明文化** | ✅ .gitignore 明文化 `830ee8e9`・#91 決定コメント済み。残=VITE_SHOW_DEMO_ACCOUNTS 本番 false 保証 |

| 残項目 | 状態 |
|---|---|
| C-1 ローテーション／#97 実値削除／.env.staging 追跡解除 | ⏳ **ユーザー実施要**（AWS/LINE/GitHub 操作）— 2026-07-02 ユーザー実施宣言済み |
| H-5 ECS secrets(valueFrom) 化 | ⏳ ユーザー実施要（SSM 登録＋IAM）— C-1 とセット |
| H-7 カバレッジ ratchet ゲート | ⏳ 未着手（BE-refactor R3-5 として計画済み・ベースラインは次回 CI 実測） |
| M-4 STG db_reset | ⏳ 次回 STG デプロイで db_reset=true 必須（migration 010/011 適用も同時） |
| M-10 残（SHOW_DEMO 本番 false 保証）／M-11 | ⏳ M-11 は GitHub Secrets 登録がユーザー側に必要 |
| #189 残（締め後訂正の可視化要否のみ） | ⏳ FE 警告は `0dad744e` で実装済み。締め詳細/月次への「締め後訂正あり」バッジの要否が PO 未決 |

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

---

## 🟠 HIGH

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

### M-4. [#160 vs #211] 健診の「Examination ⇔ Checkup」二重管理 — PO 決定済み・残タスクのみ

**PO 決定（2026-07-02）**: Checkup 系（#211 健診パッケージ）を正とする。

**実施済み**: #160 で投入した exam 系 seed（exam_types 12000-12003 / exam_type_fields 45-58）を `003_seed_demo.sql` から撤去（撤去箇所に墓標コメント・ID 再利用禁止を明記）。exam_results / exams に当該 ID への参照が無いこと・004 に波及が無いことを確認済み。`scripts/verify_seed.py` PASS。

**残タスク**
1. ✅ ADR-004 記録済み（`2e09391c`）・#160/#211 に決定コメント済み（2026-07-02）。
2. 適用済み seed の編集のため、**次回 STG デプロイは db_reset=true 必須**（migration 010 の適用も同時に行う）。

**検証方法**: fresh DB apply 後、健診記録の入力導線が Checkup 系1系統のみであることを UI で確認。

### M-10. [#91] frontend/.env.production が依然 git 追跡中・デモアカウント本番無効化なし

**事象**: `frontend/.env.production` は追跡継続（本体セッションで確認済み）。中身は `VITE_API_URL`（STG ALB endpoint）と `VITE_SHOW_DEMO_ACCOUNTS`。ALB endpoint の秘匿性は低いが Issue の除去要件は未達。`.gitignore:1-4` には「.env.production は git track する」と Issue と**逆方向のコメント**があり、方針が不整合のまま。

**PO 決定（2026-07-02）**: 案 (a) 追跡継続を明文化する。あわせて `VITE_SHOW_DEMO_ACCOUNTS` の本番 false 保証（下記 2）を実装する。

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

---

## 🟢 LOW（記録のみ・対応は任意）

| Issue | 内容 | 推奨対応 |
|---|---|---|
| #194 | 受入基準「現行ベースライン数値の記録」未達。`docs/coverage-policy.md:49` が「(CIにて測定予定)」プレースホルダのまま | 次回 CI 実行の coverage artifact から実数を転記（H-7-2 と同一作業） |
| #182 | クローズコメント本文「本Issueはクローズしない（OPEN維持）」と実際の CLOSED 状態が矛盾 | #182 に「子 #188/#189 完了に伴いクローズ」の訂正コメントを1行追加 |
| #123 | P2-4「change_amount 必須化」は実装が `binding:"min=0"` 非負強制のみで必須ではない | 現仕様として受容済み（チェックリストコメントに明示あり）。対応不要 |
| #187 | 「印刷面はちょうど1件」は既定値依存で構造保証ではない。全ポータル `active={false}` 時に白紙化しうる | JSDoc 文書化済み・呼び出し側責務。新規 PrintPortal 追加時のレビュー観点として維持 |
| #179 | クローズコメントの migration 番号（010/008）が統合後の実体（001_init.sql）と乖離 | 文書ドリフトのみ。必要なら #179 にコメント1行追記 |
| #196 | pet CRUD テストが R/D 止まり（Create/Update なし）。medical_record_image_repository 未カバー | 未対応（M-8 対応 `5b0ac22d` には同梱せず）。任意で追加 |
| #151 | PM終了 18:29:29 を「−1秒ルール」で :59 として実装（mockup 誤記と判断・コメント明記済み） | 妥当。対応不要 |
| #153 | 統合締めテーブルの件数（billingDetails）と金額（categories）がサーバ別系統値でクライアント突合なし（コード内明記・整合はサーバ責務） | 妥当。対応不要 |
| #153 | 参照PDFの視覚レイアウトは Issue 添付欠落のため未検証（列は既存2表の和集合から導出） | 実運用で帳票様式の指摘があれば対応 |
| #191 | 月次 exact-match 化により 0%（非課税）行が旧「軽減」→「標準」バケットへ移動する挙動変化 | 解消（2026-07-02）: M-7 対応で締めレジ経路も同一規則に統一し、0%→標準を回帰テストで固定（`accounting_report_tax_breakdown_test.go`） |
| #184 | 実ブラウザ印刷の視覚検証は JSDOM 制約で未実施 | Playwright 導入時に印刷ビューのビジュアルリグレッションを追加 |

---

## 本監査の対象外とした既知の横断事項

- **監査ログ書込がコードベース全体で tx 外の best-effort**（`auditRepository.Create` が dbOrTx 非使用）: #189 等の個別 Issue の受入条件とは独立した横断的既知事項のため、各 Issue の判定には含めていない。#211 系の follow-up 作業（audit-tx 原子化: refund/checkup_field_results は対応済み、残りは横断展開中）で別途追跡されている。
- **検証手法の限界**: 静的コードリーディングのみ。テスト実行・DB照会・実ブラウザ検証は未実施。「テストが存在する」ことは確認済みだが「テストが通る」ことは未確認。#212 のカバレッジ実数も実測していない。

---

## 推奨アクション（優先順・依存関係つき）

| # | 対応 | 対象 | 依存 |
|---|---|---|---|
| 1 | シークレット即時ローテーション＋Issue #97 実値削除 | C-1 | なし（最優先・ユーザー実施宣言済み） |
| 2 | ECS secrets(valueFrom) 化 → .env.staging 追跡解除 | H-5 → C-1-3 | 1 の後（新値で登録） |
| 3 | 次回 STG デプロイで db_reset=true（migration 010/011 同時適用） | M-4/#215 | デプロイタイミング |
| 4 | カバレッジ ratchet ゲート導入＋ベースライン記録 | H-7/LOW#194 | なし |
| 5 | CI 衛生残（SHOW_DEMO 本番 false 保証・M-11 Secrets 登録）＋#189 残（締め後訂正の可視化要否） | M-10/M-11/#189 | なし |
