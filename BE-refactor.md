# BE-refactor 第8期（BE8）— backend/ 全体構成の Go ベストプラクティス統一

> 起票: 2026-07-17（要件責任者: 曽我。フラット巨大パッケージの是正を Go/Gin ベストプラクティスへ統一する方針決定）
> 2026-07-17 (2): ユーザー指示によりスコープを repository/service から **backend/ 全体**へ拡張。internal 全 15 パッケージ + cmd/ + worker/ + migrations/ + 直下ファイルを実測評価済み（§1・§1.5）。
> **着手条件: Go-live（2026-07-18）完了後**。納品前の構造変更は禁止。
> 読者 = 着手するエージェント（Sonnet 5 想定）。本ファイルだけで作業に入れる粒度で書く。実測値は 2026-07-17 時点 — **着手時に §1 の計測コマンドで再実測してから始めること**。

---

## 0. 要約

backend/ 全体を実測評価した結果、**問題は service / repository / handler の3層フラット肥大に集中**しており、他の 12 internal パッケージ・cmd/・worker/・migrations/・直下ファイルは健全（§1.5 で個別に判定根拠を明示 — 触らないこと自体が決定事項）。`service`（404 files / 13.1万行）と `handler`（475 files / 9.5万行）は単一フラットパッケージで、ドメイン間のカプセル化が型システムで効かず、コンパイル単位・名前空間が巨大化している。repository は分割進行中だが**方針が未文書のまま 2 規約が混在**し、しかも分割が lint 網羅性の検証（BE8-0）なしに進んでいる。本計画は「層優先 × ドメインサブパッケージ」を正式規約とし、strangler 方式で repository → service → handler の順に段階統一する。**一斉移動は禁止**。最初の必須ゲートは BE8-0 — これを飛ばすと移動したファイルが臨床安全 lint の監視から静かに外れる。

---

## 1. 現状実測（2026-07-17）

```bash
# 再実測コマンド（着手時に必ず実行）
cd backend
for d in internal/*/; do
  n=$(find "$d" -maxdepth 1 -name "*.go" | grep -vc _test); t=$(find "$d" -maxdepth 1 -name "*.go" | grep -c _test)
  echo "$d impl=$n test=$t subdirs=$(find "$d" -mindepth 1 -type d | wc -l) lines=$(find "$d" -name '*.go' | xargs cat | wc -l)"
done
```

**internal/ 全 15 パッケージ（2026-07-17 実測）:**

| パッケージ | impl + test | 行数 | 判定 |
|---|---|---|---|
| **service** | 202 + 202 | **131,093** | **是正対象（BE8-5）** — 完全フラット |
| **handler** | 269 + 206 | **95,040** | **是正対象（BE8-7）** — フラット（サブ dir は testdata のみ） |
| **repository** | 107 + 164 | **53,221** | **是正対象（BE8-4）** — 混在: ドメインサブパッケージ **14 個** + repohelpers + フラット残 |
| model | 85 + 18 | 5,751 | 現状維持（§8 — GORM モデルは FK 相互参照で分割すると cycle 不可避） |
| middleware | 9 + 8 | 2,343 | 健全 |
| infra | 7 + 2 | 1,450 | 健全（既にサブ dir 4 個で目標形） |
| apicontract | 1 + 2 | 1,232 | 健全（単一責務） |
| config | 2 + 2 | 525 | 健全 |
| errors | 1 + 1 | 447 | **任意是正（BE8-8）** — `package errors` が stdlib を遮蔽し全 import 側が `apperrors` alias を強制されている |
| csvimport / dbconn / logger / seedbundle / timeutil / authjwt | 各 1〜2 | 12〜257 | 健全（小さく単一責務。timeutil は用途限定名で `util` 禁止則に非抵触） |

repository サブパッケージ実勢（**計画起票日中に 9→14 へ増加 — strangler が BE8-0 ゲート未整備のまま進行中**であり、BE8-0/BE8-1 の緊急度が高い）: `animalspecies, cage, chiefcomplaint, closingspecialperiod, examtype, insurance, merchandiseitem, occupation, passwordreset, paymentmethod, reservationtype, tokenblacklist, trimmingcoursetype, vaccine` + `repohelpers`（scope.go / tx.go）。設計意図の先例 = `paymentmethod/repository.go` 冒頭コメント。

個々のファイルは最大 617 行で 800 行規約内。**問題はファイルサイズではなくパッケージ粒度**。

### 1.5 対象外と評価した領域（触らないことも決定事項）

| 領域 | 実測 | 判定理由 |
|---|---|---|
| `cmd/`（api, migrate, lstep-migrate, seed-export, stage-import, coverage-ratchet） | 6 バイナリ + `_archive`（underscore prefix で Go ビルド対象外） | 公式レイアウト準拠。変更不要 |
| `worker/`（index.ts ほか） | TypeScript の Cloudflare Worker ラッパ | Go スコープ外。配置は妥当（backend デプロイ単位に同梱） |
| `migrations/` | 001_init.sql + seeds | 独自規約あり（migrations/CLAUDE.md）。本計画のスコープ外 |
| backend/ 直下のバイナリ・成果物（`api`, `migrate`, `lstep-migrate`, `seed-old-db`, `stage-import`, `coverage.out`, `tmp/`） | **全て gitignored を確認済み**（git ls-files / check-ignore 実測） | 衛生問題なし。`seed-old-db` は対応する cmd/ が既に無い stale ローカルバイナリ — 見つけたら手元で消してよい（git 影響なし） |
| `backend/docs/`（api.yaml） | OpenAPI 正本 | 変更不要 |
| Dockerfile.dev / .production / entrypoint.sh / tygo.yaml / CODING_RULES.md | 設定・規約ファイル | 公式レイアウト上、非 Go ファイルのルート配置は正当 |

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

> **コード規約の恒久正本は `.claude/skills/go-package-conventions/SKILL.md`**（本ファイルは対応後削除される計画書であり、規約は skill 側に残る。本節は作業用の写し）。

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
  handler/              # BE8-7: service 完了後に同方式で分割（testdata/ は共有のまま）
  model/                # 現状維持（§8 — 分割しない決定）
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

| # | 地雷 | 実測根拠（2026-07-17 検証済み） | 対処 |
|---|------|---------|------|
| 1 | **自作 lint の走査範囲は「1階層まで」** — `repoSourceFS` の embed は `//go:embed *.go */*.go`（`preload_clinic_scope_lint_test.go:35`）で、`repository/<domain>/*.go` は**既にカバー済み**。audit_tx / dbortx の 2 lint も同じ repoSourceFS を共有（各ファイル冒頭コメントで明記）。**残る盲点 = ①2階層目以降（`<domain>/<subdir>/*.go` は不可視）②service 側を走査する lint は存在しない**。migration_cascade は `os.ReadDir(migrationsDir)` で本件と無関係。※ファイル名は `test_schema_enum_parity_test.go`（`_lint_test` ではない）| BE8-0: 「2階層目の違反が RED になる」回帰テスト追加（不可なら 2 階層構成を規約で禁止）。dbortx の allowlist は `baseFileName` キーのため移動で壊れない（要 1 件実証） |
| 2 | **サブパッケージからテスト基盤が使えない（構造的）** — `setupTestDB` は `repository/db_setup_test.go:130` の **`_test.go` 内定義**であり、テストファイルはパッケージ外から import 不能。先行 9 サブパッケージのローカルテストが 0 件なのはこれが原因 | grep 実測 | BE8-3: `db_setup_test.go` のヘルパを importable なパッケージ（案: `repohelpers/repotest`）へ抽出するのが必須の先行作業 |
| 3 | **DI 配線 = `cmd/api/main.go`**（`service.New*` を直接呼ぶ。他に `handler/auth_session.go`・`cmd/lstep-migrate/main.go` が service を参照）。移動バッチごとに main.go の import/呼び出しが変わる | grep 実測（初回調査の「main.go に無い」は BSD grep の `\b` 非対応による偽陰性 — 訂正済み） | 各バッチで main.go を必ず diff 確認。コンストラクタは `NewReservationServiceWithAvailabilityAndType` 等 stutter 命名 — 移動時はリネームしない（§3） |
| 4 | **同一パッケージ内のドメイン間参照は import に現れない** — 分割して初めて cycle がコンパイルエラー化する | Go 言語仕様 | BE8-2 の依存グラフ実測で葉から抽出。cycle は consumer 側 interface で切る — **in-repo 先例あり**: `reservation_service.go:125` の `typeRepo reservationTypeFinder`（小文字ローカル interface）。この形を標準とする |
| 5 | パス参照の追随対象: 各層 CLAUDE.md・`docs/architecture/overview.md`・`.claude/refs/gin-architecture-compliance.md`・scoped 検証規約。**ci.yml は `backend/**` 一括フィルタのため追随不要**（確認済み） | `ci.yml:46` | 各バッチのチェックリストに含める（BE8-4 手順テンプレ） |
| 6 | PR #186（main→staging）が open — 大規模 rename は PR を膨らませる | task.html | 着手は #186 マージ後 |

---

## 6. タスク分割（この順で実行）

### BE8-0: 自作 lint の網羅性固定【必須ゲート・他タスクの前提】
- **前提事実（検証済み・§5-1）**: 1 階層サブパッケージは `go:embed *.go */*.go` で既にカバーされている。よって本タスクは「盲点調査」ではなく**「カバー範囲を回帰テストで固定し、範囲外を規約で塞ぐ」**作業。
- **作業**:
  1. temp-revert RED 実証: 既知違反コード片を `repository/paymentmethod/` 配下（1階層）に一時配置 → preload/audit_tx/dbortx の 3 lint が **RED になること**を確認 → 削除。これを恒久回帰テスト（embed 済みフィクスチャ）として 3 lint に追加。
  2. 2 階層（`repository/<domain>/<subdir>/`）に同じ違反片を置き、**GREEN のまま＝不可視**であることを実証。対処は「`all:` embed 化して検出」or「サブパッケージ内のさらなるディレクトリ分割を規約で禁止（repository/CLAUDE.md へ 1 行）」— 後者を推奨（YAGNI・embed 変更は 3 lint 共有部の変更でリスクが上がる）。
  3. dbortx allowlist が `baseFileName` キーで移動に耐えることを、allowlist 記載済みファイル 1 件を仮移動して実証。
- **対象**: `backend/internal/repository/{preload_clinic_scope,audit_tx_inventory,dbortx_inventory}_lint_test.go`（migration_cascade は `os.ReadDir(migrations)` のため対象外。`test_schema_enum_parity_test.go` は走査方式を着手時に確認）
- **検証**: `docker compose exec backend go test ./internal/repository/ -run 'Lint|Inventory|Parity' -count=1`
- **完了条件**: ①1階層違反検出の回帰テスト 3 本追加 ②2階層の扱いが規約化 ③allowlist 移動耐性の実証記録。
- 注意: **service 側を走査する lint は存在しない**（§5-1）。BE8-5 開始時に「service にも同種 lint が必要か」を判断事項として q&a.html に起票する。

### BE8-1: 規約の明文化（即日可・コード変更なし）— **✅ 完了（2026-07-17）**
- **実施済み**: §3 の決定を `backend/internal/repository/CLAUDE.md`・`backend/internal/service/CLAUDE.md`（各「パッケージ分割規約」節）・`backend/CLAUDE.md`（Architecture 節 1 行）・`.claude/refs/go-language.md` §8（+Checklist 1 項目）へ追記。他エージェント向けミラー `.agents/` はルート AGENTS.md → `.claude/CLAUDE.md` 正本参照 + sync-agents-skills.sh 再生成で追随。
- 完了条件充足: 新規ドメイン実装時に、どの層の CLAUDE.md を読んでも「サブパッケージで作る」へ誘導される。

### BE8-2: 依存グラフ実測と抽出順リストの確定
- **作業**: ① service 202 ファイルのドメイン間参照を機械集計する（同一パッケージ内のため import では見えない — go/ast で「他ドメインファイル定義の識別子参照」を数える使い捨てスクリプトを scratchpad に書く。ドメイン境界は §9 の prefix 近似を初期値とし、集計結果で補正）② 出力 = 被参照ゼロの葉ドメインから並べた抽出順リストで **§9 の表を置き換える**。
- **完了条件**: 抽出順リスト（ドメイン名・ファイル数・被参照元）が §9 に反映され、最初の 3 バッチが確定している。DI 配線は特定済み（§5-3: `cmd/api/main.go`）。

### BE8-3: サブパッケージ用テスト基盤の確立
- **作業**: `repository/db_setup_test.go:130` の `setupTestDB` は `_test.go` 内定義のためサブパッケージから **import 不能**（§5-2 — 先行 9 分割のテスト 0 件の構造的原因）。これを importable なパッケージへ抽出する（案: `repository/repohelpers/repotest` — 通常 .go ファイルとして定義し `//go:build` タグ不要。setupTestDB が参照する DB 名生成・DROP 順序等の付随ヘルパも一括で）。抽出後、先行 9 サブパッケージのうち 1 個（paymentmethod 推奨）にパッケージローカルテストの雛形を実装し、以後のバッチのテンプレとする。既存フラット側の呼び出し元 164 テストファイルは薄い alias で無変更に保つ。
- **検証**: `docker compose exec backend go test ./internal/repository/paymentmethod/ -count=1`
- **完了条件**: サブパッケージ内で DB 統合テストが書ける状態 + 雛形 1 本が green。

### BE8-4: repository 残り約 107 ファイルの段階分割
- **作業**: BE8-2 の抽出順に従い、1 バッチ = 1 ドメイン（5〜10 ファイル）で: ① `git mv` で `repository/<domain>/` へ（テストも同時） ② package 宣言変更 ③ import 更新（呼び出し側含む） ④ §5-5 のパス参照追随 ⑤ scoped test。型リネームはしない（§3）。
- **検証（バッチごと）**: `docker compose exec backend go test ./internal/repository/<domain>/ -count=1` + 呼び出し側パッケージの scoped test + `gofmt -l` 無出力。
- **完了条件**: `ls backend/internal/repository/*.go` が lint テスト・repohelpers 関連のみになる。

### BE8-5: service の段階分割（BE8-4 完了後）
- BE8-4 と同じ手順テンプレ。追加事項: ドメイン間参照は consumer 側 interface（§3）で切ってから移動する。cycle が出たら **移動を戻すのではなく interface 抽出で解決**する。service を走査する自作 lint の有無を BE8-0 方式で先に確認。
- **完了条件**: 同上（service フラット直下が空になる）。

### BE8-6: ドキュメント同期（各フェーズ末に反復）
- `docs/architecture/overview.md`・`gin-architecture-compliance.md`・`docs/spec/screens/` の該当 doc のパッケージパス記述を更新。`scripts/check-docs-symbol-drift.sh` green を確認。BE8-4/5/7 の各完了時に実施する（最後に一括ではなく）。

### BE8-7: handler 層の分割（BE8-5 完了後）
- **規模**: 269 + 206 test = 475 files / 9.5万行（3層で最多ファイル数）。サブ dir は `testdata/` のみ — 分割後も `testdata/` は共有位置に残す（各サブパッケージからの相対参照を確認）。
- **作業**: BE8-4/5 と同じ手順テンプレ。handler 固有の追加確認: ①ルート登録（`handler.go`・`master_routes.go`・`reservation_line_routes.go`）が全ハンドラを参照するため、バッチごとに登録側の import 更新が必須 ②handler 内 lint（`medical_record_image_handler_test.go`・`lab_report_handler_test.go` 等の allowlist 型テスト）の走査範囲を BE8-0 方式で先に確認 ③P5（RequirePermission 必須）等の handler 系 P ルールの検査機構がパッケージ分割に耐えるか確認。
- **完了条件**: handler フラット直下が ルート登録ファイル・lint テスト・testdata のみになる。
- **判断ゲート**: BE8-5 完了時点の実測（ビルド時間・見つけにくさの実感）で「handler は分割せず現状維持」へ倒すことも許可する。その場合は §8 へ理由付きで移す。

### BE8-8: [任意・独立] `internal/errors` → `internal/apperrors` リネーム
- **背景（実測）**: `package errors` が標準ライブラリ `errors` を遮蔽するため、**全 import 側が `apperrors "..."` の alias を強制**されている（service 層で確認）。Google スタイルの「呼び出し側から見た名前」原則に反する。
- **作業**: ディレクトリ・package 宣言を `apperrors` へ変更し、既存の alias 付き import を素の import に機械置換（alias 名と新パッケージ名が一致するため、`apperrors "..."` → `"..."` の削除だけで動く低リスク変換）。
- **検証**: 変更パッケージの scoped test + `gofmt -l` 無出力。**優先度: 低**。どのフェーズとも独立して実行可。

---

## 7. 開始トリガ・凍結条件

- **凍結**: Go-live（2026-07-18）完了まで着手禁止。BE8-1（文書のみ）だけは即日可。
- **開始トリガ**: 納品完了 + PR #186 マージ + main CI green の 3 条件成立後、BE8-0 から。

## 8. やらないこと（決定済み）

- **Option A（ドメイン優先の全面転換）** — §3 の理由により不採用。再評価しない。
- **pkg/ ディレクトリ新設** — self-contained server binary であり公式ガイダンス上 `internal/` で完結（§2-1）。
- **model の分割** — GORM モデル 85 files は FK・Preload で相互参照しており、ドメイン分割すると model 間 import cycle が不可避。単一 `model` パッケージは go.dev 公式例（`internal/model/`）とも整合。5,751 行と軽量で実害なし。
- **§1.5 の健全領域への変更**（cmd/・worker/・migrations/・小規模 12 パッケージ）— 触らないことが決定事項。改善提案が出たら①要件から検証する。
- **移動と同時の公開型リネーム** — diff 爆発防止。リネームは別コミット。
- ~~handler 層の分割はやらない~~ → スコープ全体化（2026-07-17 (2)）に伴い **BE8-7 として計画内へ移動**（BE8-5 完了後・判断ゲート付き）。

---

## 9. ドメイン別ファイル数（2026-07-17 prefix 近似・BE8-2 の初期入力）

> 集計コマンド: `ls backend/internal/service/*.go | grep -v _test | xargs -n1 basename | <prefix抽出> | sort | uniq -c | sort -rn`
> 境界はファイル名 prefix の近似。BE8-2 の go/ast 集計で置き換える。

**service（実装 202 files）の主クラスタ:**

| クラスタ | files（近似） | 備考 |
|---------|------|------|
| lstep_*（tag/health/delivery/settings/batch/csv） | **約 40** | 最大クラスタ。Write API 停止中（PO-001）のため参照断面が安定しており分割好機。ただし delivery trigger は実 LINE 配信に繋がる — 移動時 scoped test 必須 |
| liff_* | 10 | 飼主向け。他ドメインからの被参照が少ない葉候補 |
| reservation_type / reservation_staff | 11 | reservation 本体と分離可能か BE8-2 で判定 |
| medical_record | 9 | finalized ガード（142f5ebe）を含む臨床安全コア — 移動は最後尾に回す |
| validators_*（pet/owner/name/master/contact/auth/accounting） | 8 | ドメイン横断の入力検証。`validators` サブパッケージ 1 個に束ねる（util 名は不可・§3） |
| staff / owner / accounting / checkup_sync | 各 4 | 中規模 |
| 残り | 1〜2 files × 多数 | ロングテール。機械的に移せる葉 |

**repository（フラット実装 107 files）:** `accounting` 8 files が唯一のクラスタで、**残りはほぼ 1 ドメイン = 1 ファイル**（reservation_type 系 6 種・trimming 系 4 種・staff/shift 系 4 種など）。分割は service より機械的で、BE8-4 を先行させる根拠。
