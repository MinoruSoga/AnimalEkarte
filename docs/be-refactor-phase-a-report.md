# Phase A 検証結果 — Corrected Report（ローカル作業ID版）

> 本ファイルは `docs/be-refactor-followup-status.md`（正本・詳細根拠）のPhase Aセクションを、
> 架空Issue番号を排したローカル作業IDで要約した「読みやすいエグゼクティブサマリー」。
> 詳細な実測手順・証拠・訂正経緯は必ず正本側を参照すること（本ファイルは要約であり正本ではない）。

## 前提の再確認（2026-07-03）

- `gh issue view 216/219/221/223/225` は全て `GraphQL: Could not resolve to an issue or pull request` を返した。
- `gh issue list --state all` 直近確認の最大番号は **#215**。
- 依頼元プロンプトが `#216`/`#219`/`#221`/`#223`/`#225a` として参照していた5項目は、**実在しない番号**であり、
  過去計画上の仮IDに過ぎない。以降 **Phase A-1〜A-5** のローカル作業IDでのみ管理する。

## ステータス一覧

| ローカル作業ID | 内容 | ステータス | 備考 |
|---|---|---|---|
| Phase A-1 | date-only FE 影響インベントリ | 🟡 調査完了・実装BLOCKED | FE/PO確認要。3系統が個別対応必要（下記） |
| Phase A-2a | RLS role privilege 実測 | ✅ 完了 | `ekarte_user`は`rolsuper=true rolbypassrls=true`実測確認 |
| Phase A-2b | DBトランザクション開始点inventory | ✅ 完了（訂正込み） | 当初「Begin/BeginTx=0件」は誤り→**4件**存在。tx開始点合計**27箇所** |
| Phase A-2c | 接続プーラー構成 | ✅ 完了 | 外部プーラー無し。`SET LOCAL`漏洩の既知の罠は非該当 |
| Phase A-3 | 非ownerロールRLSローカル実証 | ✅ 完了（未コミット注意） | `rls_effectiveness_test.go`で6 subtest PASS。**git管理外**、次回コミット時に含める必要あり |
| Phase A-4 | batch/bypass経路監査 | ✅ 完了（訂正込み） | 当初「bypass機構0件」は誤り→`system_admin`という審査済みapp層バイパスが実在。MEDIUM指摘2件は未対応follow-up |
| Phase A-5 | BE-refactor.md陳腐化記述の状態確認 | 🔵 BLOCKED（意図的） | followup-status.mdへの集約は完了。BE-refactor.md本体への反映は並行編集解消待ち |

## 実行済み検証（executable evidence）

- Phase A-2a: `SELECT rolname, rolsuper, rolbypassrls FROM pg_roles WHERE rolname = current_user`（ローカルdocker db実測）
- Phase A-2b: `.Begin(`/`.BeginTx(`/`.Transaction(` の全site grep + 手動分類（2026-07-03再検証で`trimming_repository.go:85`の誤分類を訂正）
- Phase A-3: `docker compose exec backend go test ./internal/repository/... -run 'TestDBConnectionRolePrivileges_LocalMeasurement|TestRLSPolicyEffectiveness_ForcedRLSIsolatesByClinicGUC' -v` → 2テスト（6 subtest含む）全PASS
- Phase A-4: `permission_middleware.go:10,23` / `context_helpers.go:104-191` の直接コード確認（日本語表記によるregex false negativeの実例）

## 修正済み不整合（followup-status.md記載の再検証・矛盾なし）

前回検証時点の記述と今回のRead結果を突合し、**差分なし**（矛盾レコードなし）。

| 項目 | 当初記述 | 訂正後（現在の正本記述） |
|---|---|---|
| 生Begin/BeginTx件数 | 0件（`internal/`限定の誤った一般化） | **4件**（`cmd/migrate`3箇所 + `cmd/stage-import`1箇所。いずれもリクエストパス外の独立バイナリ） |
| tx開始点合計 | 22箇所（ignore-ambient）と誤カウント | **27箇所**（ambient起点1 + dbOrTx参加可能4 + ignore-ambient21 + service1）。`trimming_repository.go:85`は「dbOrTx参加可能」側に属する（当初「新規tx起点」に誤分類） |

## 残BLOCKED項目（未完了・過大申告防止のため明記）

- **RLS full実効化**: 非ownerアプリロール新設 + 全clinicテーブルへの`FORCE ROW LEVEL SECURITY` + 全27 tx開始点への`SET LOCAL app.current_clinic_ids`配線が必要。all-or-nothingで高リスク。**未着手**（architect/PO判断待ち）。
- **date-only wire契約変更**: 22箇所のresponse drift解消（datetime→date-only統一）は本体未着手。R3-3 gate（CI）がdrift可視化のみ実施中。**FE/PO確認完了までは着手しない**。
- **BE-refactor.md本体への反映**: 並行セッション編集中につき見送り継続。統合計画は`docs/be-refactor-integration-plan.md`参照。

## 独立再検証ログ（2026-07-03 追加パス）

以下の主張を本セッションで独立に再実行・再確認した（前回パスの記述を鵜呑みにせず executable evidence を優先）。

| 主張 | 再検証方法 | 結果 |
|---|---|---|
| #216/#219/#221/#223/#225a は非実在、直近最大は#215 | `gh issue view 216/219/221/223/225`（全て "Could not resolve"）+ `gh issue list --state all --limit 5` | ✅ 一致（最大 #215） |
| 生Begin()/BeginTx() = 4件 | `grep -rn '\.Begin(\|\.BeginTx(' backend/ \| grep -v _test.go` | ✅ 一致（`cmd/migrate/main.go:177,316,351` + `cmd/stage-import/apply.go:155`） |
| tx開始点合計27箇所（ambient1+dbOrTx参加可能4+ignore-ambient21+service1） | `.Transaction(` 全site grep + 手動分類（コメント行除外） | ✅ 一致。21件のignore-ambientリストも1件ずつ完全一致 |
| `trimming_repository.go:85` は dbOrTx参加可能（ignore-ambientではない） | 該当行を直接Read | ✅ `dbOrTx(ctx, r.db).Transaction(...)` を確認、分類正しい |
| BE-refactor.md diffは3行追加/1行削除で並行編集が継続中 | `git diff --stat -- BE-refactor.md` + `git diff -- BE-refactor.md` | ✅ 一致。内容も整合訂正のみで本タスクの結論と矛盾なし |
| rls_effectiveness_test.go 6 subtest GREEN | `docker compose exec backend go test ... -run TestRLSPolicyEffectiveness_ForcedRLSIsolatesByClinicGUC` | ✅ 6/6 PASS 再確認 |
| rls_role_privilege_test.go: rolsuper=true rolbypassrls=true | `docker compose exec backend go test ... -run TestDBConnectionRolePrivileges_LocalMeasurement` | ✅ 一致（ログ文言まで同一） |

**発見して修正した不整合（本パスのみ）**:
- `docs/be-refactor-commit-prep.md` の並行変更ファイル数が「421件」で固定表記されていたが、実測は**453件**（32件増加=並行セッション継続中の証拠）。固定値をやめ「commit直前に再計測」の指示に書き換えた。
- 同ファイルの見出しチェック用grepコマンド（`^#\{1,6\} .*#21[6-9]...`）が、安全に限定された「旧ラベル」表記（例: 表ヘッダーで非実在を宣言した上での「旧 #219 相当」という行内表記）まで誤検知する設計だった。単一行正規表現では表ヘッダーの文脈や複数行にまたがる限定表現を認識できないため、これは構造的に解消不能な誤検知であり、目視での前後文脈確認を必須とする注記を追加してコマンドを差し替えた。
- 上記2点以外は前回パスの記述と完全に一致（矛盾レコードなし）。

## 参照

- 正本: `docs/be-refactor-followup-status.md`
- Issue起票ドラフト: `docs/be-refactor-issue-drafts.md`
- commit準備: `docs/be-refactor-commit-prep.md`
- BE-refactor.md統合計画: `docs/be-refactor-integration-plan.md`
