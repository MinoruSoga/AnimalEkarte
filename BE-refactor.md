# BE-refactor 第8期（BE8）— backend/internal パッケージ構成の Go ベストプラクティス統一

> 起票: 2026-07-17（要件責任者: 曽我。フラット巨大パッケージの是正を Go/Gin ベストプラクティスへ統一する方針決定）
> **着手条件: Go-live（2026-07-18）完了後**。納品前の構造変更は禁止。
> 読者 = 着手するエージェント（Sonnet 5 想定）。本ファイルだけで作業に入れる粒度で書く。実測値は 2026-07-17 時点 — **着手時に §1 の計測コマンドで再実測してから始めること**。

---

## 0. 要約

`backend/internal/service`（404 files / 13.1万行）と `repository`（271 files / 5.2万行）は単一フラットパッケージであり、ドメイン間のカプセル化が型システムで効かず、コンパイル単位・名前空間が巨大化している。repository 側では既にドメインサブパッケージ分割が 9 個進行中だが**方針が未文書のまま 2 規約が混在**している。本計画は「層優先 × ドメインサブパッケージ」を正式規約とし、strangler 方式で段階統一する。**一斉移動は禁止**。最初の必須ゲートは BE8-0（自作 lint のサブディレクトリ盲点解消）— これを飛ばすと移動したファイルが臨床安全 lint の監視から静かに外れる。

---

## 1. 現状実測（2026-07-17）

```bash
# 再実測コマンド（着手時に必ず実行）
for d in repository service handler; do
  n=$(ls backend/internal/$d/*.go 2>/dev/null | grep -vc _test)
  t=$(ls backend/internal/$d/*.go 2>/dev/null | grep -c _test)
  echo "$d: impl $n + test $t / $(cat backend/internal/$d/*.go | wc -l) 行"
done
find backend/internal/repository backend/internal/service -mindepth 1 -type d
```

| 層 | 実測（2026-07-17） | 構成 |
|---|---|---|
| service | 202 + 202 test = 404 files / 131,093 行 | **完全フラット**（サブ dir 0） |
| repository | 107 + 164 test = 271 files / 52,158 行 | **混在** — フラット + サブパッケージ 9 個 |
| handler | 269 + 206 test = 475 files / 95,040 行 | フラット（今期スコープ外・§8） |

既存サブパッケージ（先行事例）: `repository/{paymentmethod, animalspecies, cage, closingspecialperiod, insurance, passwordreset, tokenblacklist, trimmingcoursetype}` + 共有ヘルパ `repository/repohelpers`（scope.go / tx.go）。`paymentmethod/repository.go` 冒頭コメントが分割の設計意図の先例（"thin domain split of the flat repository package. Shared clinic-scope helpers come from repository/repohelpers"）。

個々のファイルは最大 617 行で 800 行規約内。**問題はファイルサイズではなくパッケージ粒度**。

---

## 2. 調査結果（2026-07-17 実施）

1. **Go 公式**（[go.dev/doc/modules/layout](https://go.dev/doc/modules/layout)）: サーバプロジェクトはロジックを `internal/` 配下に置き、公式例は `internal/auth/`, `internal/metrics/`, `internal/model/` の**ドメイン名パッケージ**。cmd/ にバイナリ。→ 本リポジトリは internal/ + cmd/ は準拠済み。巨大フラットパッケージは公式例の姿ではない。
2. **Google Go スタイルガイド**（[best-practices](https://google.github.io/styleguide/go/best-practices)）:
   - パッケージ名は機能を表すドメイン名。`util`/`helper`/`common` は不可。
   - **識別子でパッケージ名を繰り返さない**（stutter 禁止: `paymentmethod.NewRepository` であって `paymentmethod.NewPaymentMethodRepository` ではない）。
   - 分割基準 = 概念的に独立した機能は小さな専用パッケージへ。逆に「両方 import しないと使えない」なら統合が正。
3. **Gin/コミュニティ実勢（2025-2026）**: `internal/{handler,service,repository,domain,middleware}` の層構成 + 小さく焦点の合ったパッケージ・consumer 側 interface 定義・非循環依存が合意。参照: [Go Project Structure 2026 (reintech)](https://reintech.io/blog/go-project-structure-2026-clean-architecture-best-practices)、[oneuptime: How to Structure Go Projects](https://oneuptime.com/blog/post/2026-01-07-go-project-structure/view)、[go-gin-clean-starter](https://github.com/Caknoooo/go-gin-clean-starter)、[golang-gin-clean-architecture](https://github.com/harmannkibue/golang-gin-clean-architecture)、[awesome-go-education: project layout](https://mehdihadeli.github.io/awesome-go-education/project-layout-structure/)。

---

## 3. 目標構成（決定）

**採用 = Option B: 層優先を維持し、各層内をドメインサブパッケージ化**

```
backend/internal/
  repository/
    repohelpers/        # 共有: clinicScope・DBOrTx（既存）
    reservation/        # ドメイン単位。package reservation
    accounting/
    ...
  service/
    servicehelpers/     # 必要になった時点で新設（先行して作らない — YAGNI）
    reservation/        # package reservation（service 側）
    ...
  handler/              # 今期は現状維持（§8）
```

**不採用 = Option A（ドメイン優先: `internal/reservation/{handler,service,repository}`）**。理由: ①層別 CLAUDE.md・P1-P18 lint 体系・scoped 検証規約がすべて層パスを前提としており波及が桁違い ②repository の先行 9 分割が Option B 形であり方向転換は二重の手戻り。

**命名規約（BE8-1 で CLAUDE.md へ明文化する内容）**:
- パッケージ名 = 単数形・全小文字・アンダースコアなしのドメイン名（`paymentmethod` 先例に従う）。`util`/`common`/`helpers` 単独名は禁止（`repohelpers` は既存例外として存置）。
- 新規型は stutter 禁止（`reservation.Repository` / `reservation.NewRepository`）。**既存型の公開リネームは移動と同時にやらない**（呼び出し側全変更で diff が爆発する）— 移動時は型名維持、リネームは別コミットで機械的に。
- service ↔ service のドメイン間参照は **consumer 側で interface を定義**して受ける（import cycle の構造的回避。調査結果 3 の合意事項）。

---

## 4. 移行方式 = strangler（一斉移動禁止）

- **新規ドメインは必ずサブパッケージで作る**（即日発効・BE8-1 で規約化）。
- **既存フラットファイルは、そのドメインに実装変更が入るときに移す**。移動だけの巨大 PR を作らない。ただし BE8-4/5 の計画バッチ（葉ドメインから 5〜10 ファイル単位）は例外として許可。
- 完了の定義 = フラット直下に .go が残っていない状態。期限は設けない（混在は BE8-1 の規約明文化があれば管理された状態になる）。

---

## 5. 制約・地雷（着手前に必読）

| # | 地雷 | 実測根拠 | 対処 |
|---|------|---------|------|
| 1 | **自作 lint のサブディレクトリ盲点** — `preload_clinic_scope_lint_test.go` は `go:embed` + `fs.WalkDir` でソースを走査する。embed パターンがフラット `*.go` のみだと、サブパッケージへ移した瞬間にそのファイルが**臨床安全 lint の監視から静かに外れる**（サイレント緑） | `backend/internal/repository/preload_clinic_scope_lint_test.go:32,365` | BE8-0 が必須ゲート。5 本全 lint（preload_clinic_scope / audit_tx_inventory / dbortx_inventory / migration_cascade / test_schema_enum_parity）で確認 |
| 2 | **先行 9 サブパッケージはパッケージローカルテスト無しで出荷済み** — `paymentmethod/*_test.go` は 0 件 | `ls repository/paymentmethod/` | BE8-3 でテスト基盤（setupTestDB の共有方法）を確立してから残りを移す |
| 3 | **DI 配線が cmd/api/main.go に無い**（`Service` 文字列 0 件）— 配線箇所が未特定 | grep 実測 | BE8-2 で配線ファイルを特定してから移動計画を確定 |
| 4 | **同一パッケージ内のドメイン間参照は import に現れない** — 分割して初めて cycle がコンパイルエラー化する | Go 言語仕様 | BE8-2 の依存グラフ実測で葉から抽出。cycle は §3 の consumer 側 interface で切る |
| 5 | パス参照の追随対象: `.github/workflows/ci.yml`（paths フィルタ）・各層 CLAUDE.md・`docs/architecture/overview.md`・`.claude/refs/gin-architecture-compliance.md`・scoped 検証規約 | grep 実測 | 各バッチのチェックリストに含める（BE8-4 手順テンプレ） |
| 6 | PR #186（main→staging）が open — 大規模 rename は PR を膨らませる | task.html | 着手は #186 マージ後 |

---

## 6. タスク分割（この順で実行）

### BE8-0: 自作 lint のサブディレクトリ網羅化【必須ゲート・他タスクの前提】
- **作業**: 5 本の lint テストが `internal/repository/` 配下を**再帰的に**走査することを検証し、フラット限定なら修正する。検証は temp-revert RED 方式 — 既知違反コード片を `repository/paymentmethod/` 配下に一時ファイルとして置き、lint が **RED になることを確認**してから削除する（GREEN のままなら盲点が実在する）。
- **対象**: `backend/internal/repository/{preload_clinic_scope,audit_tx_inventory,dbortx_inventory,migration_cascade,test_schema_enum_parity}_lint_test.go`
- **検証**: `docker compose exec backend go test ./internal/repository/ -run 'Lint|Inventory|Parity' -count=1`
- **完了条件**: 5 本すべてに「サブディレクトリ内の違反を検出する」回帰テストケースが追加されている。
- 注意: lint テスト自体の移動は不要。service 側へ分割を進める際（BE8-5）に service を走査する lint があるかを同方式で再確認する。

### BE8-1: 規約の明文化（即日可・コード変更なし）
- **作業**: §3 の決定（Option B・命名規約・strangler・consumer 側 interface）を以下へ追記: `backend/internal/repository/CLAUDE.md`・`backend/internal/service/CLAUDE.md`（各 3〜5 行 + 本ファイルへのリンク）、`backend/CLAUDE.md`（1 行）。
- **完了条件**: 新規ドメイン実装時にエージェントが迷わず「サブパッケージで作る」を選べる記述になっている。

### BE8-2: 依存グラフ実測と抽出順リストの確定
- **作業**: ① service 202 ファイルのドメイン間参照を機械集計する（同一パッケージ内のため import では見えない — go/ast で「他ドメインファイル定義の識別子参照」を数える使い捨てスクリプトを scratchpad に書く。ドメイン境界はファイル名プレフィックスで近似: `reservation_*.go` 等）② DI 配線ファイルを特定する（`grep -rn "NewReservationService(" backend/` 起点）③ 出力 = 被参照ゼロの葉ドメインから並べた抽出順リストを**本ファイル §9 として追記**。
- **完了条件**: 抽出順リスト（ドメイン名・ファイル数・被参照元）が本ファイルに追記され、最初の 3 バッチが確定している。

### BE8-3: サブパッケージ用テスト基盤の確立
- **作業**: フラット repository パッケージ内の `setupTestDB` 系ヘルパをサブパッケージから利用可能にする（候補: `repohelpers` へ export / `internal/testutil` 新設 — 既存コードの実態を見て低摩擦な方を選ぶ）。先行 9 サブパッケージのうち 1 個（paymentmethod 推奨）にパッケージローカルテストの雛形を実装し、以後のバッチのテンプレとする。
- **検証**: `docker compose exec backend go test ./internal/repository/paymentmethod/ -count=1`
- **完了条件**: サブパッケージ内で DB 統合テストが書ける状態 + 雛形 1 本が green。

### BE8-4: repository 残り約 107 ファイルの段階分割
- **作業**: BE8-2 の抽出順に従い、1 バッチ = 1 ドメイン（5〜10 ファイル）で: ① `git mv` で `repository/<domain>/` へ（テストも同時） ② package 宣言変更 ③ import 更新（呼び出し側含む） ④ §5-5 のパス参照追随 ⑤ scoped test。型リネームはしない（§3）。
- **検証（バッチごと）**: `docker compose exec backend go test ./internal/repository/<domain>/ -count=1` + 呼び出し側パッケージの scoped test + `gofmt -l` 無出力。
- **完了条件**: `ls backend/internal/repository/*.go` が lint テスト・repohelpers 関連のみになる。

### BE8-5: service の段階分割（BE8-4 完了後）
- BE8-4 と同じ手順テンプレ。追加事項: ドメイン間参照は consumer 側 interface（§3）で切ってから移動する。cycle が出たら **移動を戻すのではなく interface 抽出で解決**する。service を走査する自作 lint の有無を BE8-0 方式で先に確認。
- **完了条件**: 同上（service フラット直下が空になる）。

### BE8-6: ドキュメント最終同期
- `docs/architecture/overview.md`・`gin-architecture-compliance.md`・`docs/spec/screens/` の該当 doc のパッケージパス記述を一括更新。`scripts/check-docs-symbol-drift.sh` green を確認。

---

## 7. 開始トリガ・凍結条件

- **凍結**: Go-live（2026-07-18）完了まで着手禁止。BE8-1（文書のみ）だけは即日可。
- **開始トリガ**: 納品完了 + PR #186 マージ + main CI green の 3 条件成立後、BE8-0 から。

## 8. やらないこと（決定済み）

- **Option A（ドメイン優先の全面転換）** — §3 の理由により不採用。再評価しない。
- **pkg/ ディレクトリ新設** — self-contained server binary であり公式ガイダンス上 `internal/` で完結（§2-1）。
- **handler 層の分割** — service 完了後に実測データを持って再評価。今期スコープに含めない。
- **移動と同時の公開型リネーム** — diff 爆発防止。リネームは別コミット。
