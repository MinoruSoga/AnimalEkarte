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

> **コード規約の恒久正本は `.claude/rules/go-package-conventions.md`**（本ファイルは対応後削除される計画書であり、規約は skill 側に残る。本節は作業用の写し）。

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
| 1 | **自作 lint の走査範囲は「1階層まで」** — `repoSourceFS` の embed は `//go:embed *.go */*.go`（`preload_clinic_scope_lint_test.go:35`）で、`repository/<domain>/*.go` は**既にカバー済み**。audit_tx / dbortx の 2 lint も同じ repoSourceFS を共有（各ファイル冒頭コメントで明記）。**残る盲点 = ①2階層目以降（`<domain>/<subdir>/*.go` は不可視）②service 側を走査する lint は存在しない**。migration_cascade は `os.ReadDir(migrationsDir)` で本件と無関係。※ファイル名は `test_schema_enum_parity_test.go`（`_lint_test` ではない）| **BE8-0 完了（2026-07-17・実測で訂正）**: 2階層目は実測で不可視のままと確認・`.claude/rules/go-package-conventions.md` の既存禁止規約で塞ぐ方針を維持（embed は変更しない）。**dbortx allowlist は baseFileName キーではない**（旧記述は誤り。実測訂正: §5-1 訂正参照）。恒久回帰テスト 3 本追加済み（`TestRepoSourceEmbed_ReachesOneLevelSubpackages` 他 2 本）。 |
| 1-2 | **【2026-07-17 実測訂正】dbortx allowlist のキー方式は preload/audit_tx と異なる** — `dbortx_inventory_lint_test.go` の `walkRepositoryForDBOrTx`（298-311行）は `keyFile := baseFileName(name); if strings.Contains(name, "/") { keyFile = name }` — **1階層サブパッケージのファイルは相対パス（例: `reservationtype/repository.go`）をキーにし、bare basename ではない**（既存 allowlist の `"reservationtype/repository.go\|repository.FindAll"` 等が実例）。一方 preload（`baseFileName(filename)`）と audit_tx（`keyName := baseFileName(name)`、273-274行）は**常に bare basename**（ディレクトリ情報を捨てる）。ライブ実証（`paymentmethod/` へ一時フィクスチャ配置 → 3 lint 同時実行）で確認: dbortx の違反キーは `paymentmethod/_be8_0_evidence_1level.go\|...`（パス付き）、preload/audit_tx の違反キーは `_be8_0_evidence_1level.go:11`／`_be8_0_evidence_1level.go\|...`（パスなし）と、3 lint 同一実行内で挙動が食い違うことを確認済み。**結論: dbortx allowlist は移動で「壊れない」のではなく「必ず壊れる」**（basename→相対パスへキーが変わるため、BE8-4 の各バッチで dbortx allowlist の該当エントリを手動更新する必要がある。移動を戻す必要はなく、キーを更新すればよい）。 |
| 1-3 | **✅ 解消済み（2026-07-17・BE8-0 補完 run）** 【当初発見時】preload/audit_tx の bare-basename キーは同名ファイルの衝突リスクを内包 — BE8-4 でドメイン分割が進むと `<domain>/repository.go` のような同名ファイルが複数ドメインに増える。preload/audit_tx の `baseFileName()` はディレクトリを無視するため、2 ドメインが同名ファイル（例: `repository.go`）を持つと、両者の違反が同一キーに集約され、site-exception 数量照合（`occurrences`）が誤ってマスクされうる（false negative = 臨床安全カバレッジの縮小）。現状は 14 サブパッケージのファイル名が全て非衝突（`repository.go`/`scope.go` の重複はあるが、audit_tx は Delete 呼び出しが無いドメインでは事実上無害）ため**潜在的（レイテント）で未発火**。**衝突条件の補足（clinic-isolation-auditor 監査で発見・2026-07-17）**: stutter 禁止規約（`.claude/rules/go-package-conventions.md`）によりサブパッケージのレシーバ型名は `repository`（ドメイン名を含まない）に統一されるため、preload/audit_tx のキー（`file|function|...`）は「ファイル名衝突」だけでなく「`repository.go\|repository.Method` という関数キーまで一致する」衝突になりうる — ドメイン識別子がキーのどこにも現れないため、想定より起きやすい。**【2026-07-17 補完 run で解消】** preload の `analyzeFilePreloads` と audit_tx の `analyzeFileForClinicalResultDeletes`/`walkRepositoryForClinicalResultDeletes` を `dbortx_inventory_lint_test.go:309-314` と同じ「root は basename・1階層サブパッケージは相対パス」分岐へ統一した。TDD で実証: ①修正前に basename 衝突フィクスチャ（`paymentmethod/repository.go` 向け例外が `vaccine/repository.go` の同一パターン違反を誤って免除する）を組み、両 lint で **FAIL を実測**（fail-open の再現） → ②キーイング修正 → ③同一テストが **PASS** に転じることを実測。既存フラット root ファイル向けの exception/allowlist エントリ（`staff_repository.go` 等）は basename のまま後方互換。恒久回帰テスト計4本追加（各lintに検出力テスト1本+衝突テスト1本）。3 lint 全て `git diff --name-only HEAD` で本番 `.go` 変更ゼロ・`_test.go` のみ。 |
| 2 | **サブパッケージからテスト基盤が使えない（構造的）** — `setupTestDB` は `repository/db_setup_test.go:130` の **`_test.go` 内定義**であり、テストファイルはパッケージ外から import 不能。先行 9 サブパッケージのローカルテストが 0 件なのはこれが原因 | grep 実測 | BE8-3: `db_setup_test.go` のヘルパを importable なパッケージ（案: `repohelpers/repotest`）へ抽出するのが必須の先行作業 |
| 3 | **DI 配線 = `cmd/api/main.go`**（`service.New*` を直接呼ぶ。他に `handler/auth_session.go`・`cmd/lstep-migrate/main.go` が service を参照）。移動バッチごとに main.go の import/呼び出しが変わる | grep 実測（初回調査の「main.go に無い」は BSD grep の `\b` 非対応による偽陰性 — 訂正済み） | 各バッチで main.go を必ず diff 確認。コンストラクタは `NewReservationServiceWithAvailabilityAndType` 等 stutter 命名 — 移動時はリネームしない（§3） |
| 4 | **同一パッケージ内のドメイン間参照は import に現れない** — 分割して初めて cycle がコンパイルエラー化する | Go 言語仕様 | BE8-2 の依存グラフ実測で葉から抽出。cycle は consumer 側 interface で切る — **in-repo 先例あり**: `reservation_service.go:125` の `typeRepo reservationTypeFinder`（小文字ローカル interface）。この形を標準とする |
| 5 | パス参照の追随対象: 各層 CLAUDE.md・`docs/architecture/overview.md`・`.claude/refs/gin-architecture-compliance.md`・scoped 検証規約。**ci.yml は `backend/**` 一括フィルタのため追随不要**（確認済み） | `ci.yml:46` | 各バッチのチェックリストに含める（BE8-4 手順テンプレ） |
| 6 | PR #186（main→staging）が open — 大規模 rename は PR を膨らませる | task.html | 着手は #186 マージ後 |

---

## 6. タスク分割（この順で実行）

### BE8-0: 自作 lint の網羅性固定【必須ゲート・他タスクの前提】— ✅ 完了（2026-07-17）

**実施内容（凍結下でも安全な Tier A のみ・本番コードパスは一切変更していない）**:
1. **RED 実証（一時フィクスチャ・削除済み・コミットなし）**: `repository/paymentmethod/_be8_0_evidence_1level.go` に既知違反 3 種（Preload 無述語 / `.Delete(&model.ExamResult{})` / `dbOrTx` 未登録メソッド）を配置し、`go test ./internal/repository/ -run 'RealRepositorySourceHasNoUnscopedMasterPreload|AllowlistMatchesRealSource|DBOrTxInventory_MatchesAllowlist'` を実行 → **3 lint 全て RED** を確認（1 回のテスト実行で同時確認）。直後に削除し GREEN 復帰を確認（`git status` でも残留なしを確認済み）。
2. **GREEN（不可視）実証（一時フィクスチャ・削除済み・コミットなし）**: 同時に `repository/paymentmethod/_be8_0_evidence_2level/fixture.go`（2 階層）へ同一パターンの違反を配置 → 上記と同一の実行結果に **2 階層側の違反は一切出現せず**（3 lint とも 2 階層フィクスチャに言及するエラーなし）。2 階層が 3 lint 全てから不可視であることを実証。対処は `.claude/rules/go-package-conventions.md` の既存禁止規約（「サブパッケージ内にさらにディレクトリを掘らない」）を維持する方針を採用 — 追加の規約明文化は不要（既に閉じている）。embed を `all:` へ変更する案は見送り（3 lint 共有部の変更でリスクが上がる・YAGNI）。
3. **dbortx allowlist の移動耐性実証**: 上記 RED 実証の実行結果そのものが実証を兼ねる。dbortx の違反は `paymentmethod/_be8_0_evidence_1level.go|be8_0_dbOrTxViolationRepo.Bar`（相対パスキー）として報告され、preload/audit_tx は `_be8_0_evidence_1level.go`（bare basename）として報告された — **3 lint 内でキー方式が食い違うことをライブ実証**。結論は §5-1（訂正行）に記録: **dbortx allowlist は「baseFileName キーで移動に耐える」という旧記述は誤りで、実際は「移動すると必ずキーが変わり fail する」**。BE8-4 の各バッチで dbortx allowlist エントリの手動更新が必須（手順テンプレへの申し送り）。
4. **恒久回帰テスト 3 本を追加**（embed メカニズムそのものを検証する設計 — advisor 助言により「違反フィクスチャを埋め込んだままにする」設計から変更。理由: 埋め込みっぱなしの違反フィクスチャは本番 gate 自体を fail させてしまう上、AST アナライザのロジックは既存の `*_Analyzer` テストが inline fixture で検証済みであり、真に守るべきリスクは「embed グロブが 1 階層サブパッケージに届かなくなる」こと自体）:
   - `preload_clinic_scope_lint_test.go`: `TestRepoSourceEmbed_ReachesOneLevelSubpackages`（`paymentmethod/repository.go` が embed に含まれること + 1 階層ファイル数の floor を pin）
   - `audit_tx_inventory_lint_test.go`: `TestClinicalResultAuditTxInventory_WalksAllEmbeddedFilesIncludingSubpackages`
   - `dbortx_inventory_lint_test.go`: `TestDBOrTxInventory_WalksAllEmbeddedFilesIncludingSubpackages`
   - **反証確認済み**: `//go:embed *.go */*.go` を一時的に `//go:embed *.go` に narrow → 上記 3 本全て FAIL することを確認 → 元に戻して GREEN 復帰確認（検出力のない vacuous green ではないことの証明）。
5. **新規発見（当初は BE8-4 申し送り事項・後述 6 で本 run 内に解消）**: preload/audit_tx は bare basename キー、dbortx は相対パスキーと**3 lint 間でキー方式が不統一**。ドメイン分割が進むと同名ファイル（`repository.go` 等）の衝突で preload/audit_tx 側が false negative を起こしうる。§5-1-3 に記録。
6. **✅ BE8-0 補完（2026-07-17・別 run）**: 独立検証の結果、上記4本の恒久回帰テストは「embed が入れ子に届くこと（配送機構）」のみを証明しており、**「実際に違反を検出すること」は一度も証明していなかった**と判明（audit_tx 側は `walkRepositoryForClinicalResultDeletes` の戻り値を捨てて呼ぶだけの smoke テストに留まっていた）。これを受けて追加 run で: ①合成ソース + 入れ子パス名をアナライザへ直接投入する検出力テストを2本追加（preload・audit_tx 各1本） ②§5-1-3 の basename 衝突を実際に FAIL するテストとして再現（TDD RED 実測） ③preload/audit_tx のキーイングを dbortx と同じ相対パス方式へ統一 ④同一テストが PASS に転じることを実測（GREEN） ⑤既存 22 テスト全 green・本番 `.go` 変更ゼロを確認。詳細は §5-1-3 訂正行。

**検証**: `docker compose exec backend go test ./internal/repository/ -run 'TestRepoSourceEmbed|TestClinicalResultAuditTxInventory|TestDBOrTxInventory|TestPreloadClinicScope' -count=1` → PASS（16 テスト）。`gofmt -l` 対象 3 ファイル無出力。
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

### BE8-3: サブパッケージ用テスト基盤の確立 — ✅ 完了（2026-07-18・commit aa0dd6804）
- **実施内容**: `repotest/repohelpers` ではなく **`repository/repotest`（深さ1）** へ抽出（`repohelpers/repotest` は2階層になり「サブパッケージ内にさらにディレクトリを掘らない」規約に抵触するため実測で不採用に変更 — 抽出対象のsetupTestDB系ヘルパーはPreload/audit_tx/dbOrTxパターンを含まないため、深さ1への配置で3lintへの影響なし）。`SetupTestDB`/`SetupIsolatedTestDB`/`EnsureClinicSettingsTable`/`MakeTestOwner`/`EnsureAutoMigrated`/`MarkAutoMigrated`/`CloseSharedTestDB`/`EnumType`/`SharedTestSchemaEnumTypes`/`EnumValueRe` を export。`paymentmethod/repository_test.go` に雛形実装済み。
- **既存165テストファイルへの影響**: 164ファイルは無変更。**唯一の例外** = `test_schema_enum_parity_test.go`（型がパッケージ境界を跨ぐため `.name`→`.Name`/`.create`→`.Create`/`enumValueRe`→`repotest.EnumValueRe` の機械的フィールドアクセス変更が必須。Go の非export フィールド越境不可のため回避不能）。
- **検証済み**: `go test -p 1 ./internal/repository/... ` green（flat 165 + repotest + paymentmethod）。gofmt clean。

### BE8-4: repository 残り約 107 ファイルの段階分割 — **8バッチ実施済み（2026-07-18・commit ebe845de9/8ef9f80d0/2b304cd66/71b0c209/79f451ef/1ffbd50f/5e9cfb60/bda8f6e9）、残多数**
- **実施済みバッチ**: ①`reservationtypegroup`②`reservationtypeavailableslot`+`reservationtypeunavailabletime`③`manualarticle`（以上前run）④`inquiry_template`⑤`checkup_type`⑥`daily_record`⑦`prescription`⑧`refund`（本run・reservation/medical_record/accounting クラスタと audit を除く1ファイル低結合の葉ドメインを investigation subagent 4体の並行調査で選定）。
- **手順テンプレ確定**: ① 新規サブパッケージへ実装+テストを新規作成（`Repository`/`repository`/`New` の非stutter命名）② repohelpers.X 直接呼び出し（cage/vaccine等の主流規約、paymentmethod のローカル scope.go alias 規約とは非統一のまま — 別途 BE8-1 追補が必要）③ **旧flatファイルは型名維持のfacade化**（`type XxxRepository = <domain>.Repository`）で service/handler呼び出し側を無変更に保つ ④ 3lint該当エントリの手動更新（該当ありはdbortxのみ・本runで3ドメイン=daily_record/prescription/refund計8メソッド更新、新規免除純増0）⑤ `go test -p 1 ./internal/repository/... `（**`-p 1` 必須**）。
- **本runの追加知見**: ①**tx-atomicity/sum-tx-participation系のテストが `repository.Transactor`/`NewTransactor`/フラット共有fixtureヘルパー（`makeBillingWith`等）に依存している場合、そのままではimport cycle**（facadeでrepositoryがサブパッケージをimportするため）。解決テンプレ: `repohelpers.WithTxValue` 直結の `withTx(ctx, db, fn)` ヘルパーをテストファイル内にローカル定義し `NewTransactor(db)`+`tx.WithTx` を置換、共有fixtureヘルパーはモデル直接構築で最小複製する。②**investigation subagentの報告は鵜呑みにしない**: 並行調査で「daily_record/prescription/refund等の該当キーはaudit_tx_inventory_lint_test.goにある」という報告を受けたが実測では全て`dbortx_inventory_lint_test.go`だった（ファイル取り違え）。適用前に必ず`grep`で実ファイルを確認すること。
- **残**: 5〜10ファイル/バッチ換算で20バッチ超が未着手（`accounting`クラスタ8ファイル・`reservation`/`medical_record`/`accounting`/`LSTEP`各クラスタはCLAUDE.mdの「Forbidden in drive-by tasks」により境界マップ確定まで着手不可・`appointment_admin`(392行テストファイルが不釣合い)と`lab_import`(3コンストラクタ)は次候補から意図的除外・要個別検討）。
- **検証**: `docker compose exec backend go test ./internal/repository/<domain>/ -count=1` + `go test -p 1 ./internal/repository/... ` + 3lint実行数確認（`-run` は関数名prefix、`Lint`は不可）。

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
- **2026-07-17 実測（凍結ゲート判定）**: PR #186 = MERGED（`mergedAt: 2026-07-17T07:32:57Z`）。main CI = green（HEAD `e842e45ec`, 全 job success）。**Go-live（2026-07-18）は実行時点で未完了**（凍結解除の 3 条件のうち 1 つが未成立）。よって**フラット repository/service/handler への構造変更（BE8-3/BE8-4/BE8-8 等・本番コードパスに触れるタスク）は本 run では着手せず BLOCKED**。一方 BE8-0（lint 回帰テスト追加）と BE8-2（依存グラフ実測・ドキュメント更新）は本番バイナリのコードパスに一切触れないため凍結の趣旨に抵触しないと判断し、2026-07-17 に実施した（詳細は各タスク節）。
- **2026-07-18 再実測（凍結解除）**: PR #186 = MERGED（変わらず）。main CI = HEAD `85cc02b7b`（着手時点）で CI success / Security Scan success 実測（プロンプト内の古い実測値「未green」を実測で上書き）。**Go-live完了をユーザーへ明示確認し「完了している」の回答を取得**（Issue #257 は OPEN のまま・受け入れ条件未チェックだったため、推測せず確認を挟んだ）。3条件成立によりBE8-3/BE8-4のTier B作業を実施（詳細は各タスク節）。

## 8. やらないこと（決定済み）

- **Option A（ドメイン優先の全面転換）** — §3 の理由により不採用。再評価しない。
- **pkg/ ディレクトリ新設** — self-contained server binary であり公式ガイダンス上 `internal/` で完結（§2-1）。
- **model の分割** — GORM モデル 85 files は FK・Preload で相互参照しており、ドメイン分割すると model 間 import cycle が不可避。単一 `model` パッケージは go.dev 公式例（`internal/model/`）とも整合。5,751 行と軽量で実害なし。
- **§1.5 の健全領域への変更**（cmd/・worker/・migrations/・小規模 12 パッケージ）— 触らないことが決定事項。改善提案が出たら①要件から検証する。
- **移動と同時の公開型リネーム** — diff 爆発防止。リネームは別コミット。
- ~~handler 層の分割はやらない~~ → スコープ全体化（2026-07-17 (2)）に伴い **BE8-7 として計画内へ移動**（BE8-5 完了後・判断ゲート付き）。

---

## 9. service ドメイン依存グラフと抽出順（2026-07-17 go/ast 実測・BE8-2）

### 9.1 実測手法

- 使い捨て go/ast スクリプト（scratchpad のみ・**リポジトリ非コミット**）で `internal/service/*.go`（実装 202 files）を 2 パス解析。
  - **Pass 1**: 全ファイルのトップレベル識別子（`Recv==nil` の func・type・var/const）→ 定義ドメインのマップを構築。**メソッドは名前衝突（`Create`/`Update` 等）で汚染するため除外**。
  - **Pass 2**: 各ファイルの USE 参照（`ast.Ident`）を走査し、宣言サイト（FuncDecl/TypeSpec/Field/ValueSpec 名）と `SelectorExpr.Sel`（`x.Foo` の `.Foo`）を除外。マップに存在し **かつ定義ドメイン ≠ 自ファイルドメイン** の参照のみをドメイン間エッジとして計上。
- ドメイン境界 = ファイル名 prefix。`reservation_type`/`reservation_staff`/`checkup_sync` は first-token だと親（reservation/checkup）に吸収されるため明示分離。残りは第 1 underscore トークン。全 69 ドメイン・773 エッジ。

### 9.2 計測の限界（正直な明記）

- **メソッド経由・interface 経由の依存は検出されない**（`f.svc.DoThing()` や §5 の `reservationTypeFinder` ローカル interface）。これは欠陥ではなく設計通り — 測っているのは「`git mv` でサブパッケージ化したとき**コンパイルを壊す構文的識別子結合**」であり、実行時配線ではない。**`out-dom=0` のドメインは、他ドメインへの残依存があってもそれは既に interface 化されている ⟹ 抽出しても cycle にならない**（false-positive は edge を増やす方向にしか出ないため、実測 `out-dom=0` は真に `out-dom=0`）。
- 逆に **incoming（in-ref）はローカル変数名の衝突で過大計上され得る**（`audit`/`timeslot` 等の高 in-ref は概算）。incoming は cycle 安全性ではなく「抽出時に import 追随が必要なファイル数の目安（churn）」として使う。

### 9.3 コンポジションルートの発見（抽出順の主軸が変わる根拠）

`service.go`（`Service` 集約 struct + `NewService`）は **59 ドメインの `NewXxxService` を参照する out-dom=59 の合成ルート**。このため実ドメインは全て「root からの incoming エッジ 1 本」を持ち、**被参照ゼロの純粋な葉は `service` 自身しか存在しない**（当初想定の「in=0 を葉とする」は不成立）。表中の `in-dom=1` の大半はこの root 配線であり、ドメイン間結合ではない。

→ 抽出安全性の主軸は **outgoing（依存の向き）**。抽出後 `service.go` は必ず `service/<D>` を import する（配線エッジ）。**D が親パッケージ残留の識別子を参照する（out-dom>0）と D→親 の逆 import が生じ cycle 化**する。従って **`out-dom=0` のシンクから逆トポロジカル順に抽出する**（§7 の「臨床安全コアは最後尾」とも整合）。

### 9.4 抽出段階（reverse-topological・sinks-first）

| 段階 | ドメイン | files | in(dom/ref) | out(dom/ref) | 根拠 |
|------|---------|:---:|:---:|:---:|------|
| **①雛形** | **daily** | 1 | 1 / 2 | **0 / 0** | 純粋シンク・in は root のみ・churn ゼロ。テンプレ実証用 |
| **①雛形** | **inventory** | 1 | 1 / 2 | **0 / 0** | 同上 |
| **①雛形** | **manual** | 1 | 1 / 2 | **0 / 0** | 同上 |
| ②共有カーネル | audit | 2 | 16 / 86 | **0 / 0** | out=0 で常時 cycle-safe。高 fan-in — 先に抜くと多数の依存元がシンク化し②以降を解錠 |
| ②共有カーネル | update (update_fields) | 1 | 7 / 8 | **0 / 0** | 同上（共有 update ヘルパ） |
| ②共有カーネル | timeslot | 1 | 4 / 64 | **0 / 0** | 同上（in-ref は概算） |
| ②共有カーネル | dose | 3 | 2 / 19 | **0 / 0** | 用量計算の純粋シンク |
| ②その他シンク | reservation_staff, token, species, account, shared, smtp, go | 各1〜2 | 低 | **0 / 0** | いずれも out=0・機械的に移せる葉 |
| ③解錠後の準シンク | validators (out-dom=2), company / chronic / clinical / care / refund / vaccination / prescription ほか out-dom=1 群 | 各1〜2 | 低〜中 | 1〜2 / 低 | ②抽出で残依存が消えるとシンク化。**validators は in-dom=37 の最大 fan-in — 残 2 依存を先に解決 or interface 化してから** |
| ④結合コア（最後尾） | lstep | 48 | 12 / 62 | 3 / 15 | out-dom=3 の準シンクだが **48 files — 単一バッチ不可。tag_sync / health_tag / delivery / settings / batch / csv でサブバッチ必須** |
| ④結合コア | liff | 11 | 3 / 4 | 5 / **80** | 高 outgoing（80 refs）。依存先を先に抽出 |
| ④結合コア | accounting(6) / reservation_type(9) / treatment(3) / trimming(4) / medicine(2) / estimate / appointment(3) | — | 中 | 4〜6 / 中 | 相互結合。依存解決後に |
| ④結合コア（**厳守最後尾**） | reservation | 4 | 9 / 34 | 5 / 15 | 高結合。in/out 双方大 |
| ④結合コア（**厳守最後尾**） | medical_record | 9 | 8 / 24 | 4 / 23 | finalized ガード（142f5ebe）の臨床安全コア（§7）。最後に移す |

> **注記**: `service`（`service.go`, in-dom=0）は合成ルートであり**親パッケージに残す — 抽出対象ではない**。表の `in-dom=1` は全て root 配線（依存結合ではない）。

### 9.5 最初の 3 バッチ（BE8-5 着手順・確定）

1. **daily**（`daily_record_service.go`）— out=0・被参照は root のみ。サブパッケージ雛形（BE8-3 の repotest 基盤）を service 側で実証。
2. **inventory**（`inventory_service.go`）— 同型の純粋シンク。手順テンプレの反復確認。
3. **manual**（`manual_article_service.go`）— 同上。3 本で「1 ドメイン = git mv + package 宣言 + service.go の import 1 行 + scoped test」の型を固める。

3 本完了後、**②共有カーネル（audit → update → timeslot → dose）** を抜いて多数の準シンクを解錠 → ③ → ④結合コアの順。lstep(48f) と medical_record(臨床安全) は最後尾。

> **被参照ゼロの葉の確認**: 純粋な in-dom=0 は合成ルート `service` のみ（§9.3 の理由で他は存在しない）。実運用上の葉 = **out-dom=0 のシンク 15 ドメイン**（daily/inventory/manual/audit/update/timeslot/dose/reservation_staff/token/species/account/shared/smtp/go ほか）を抽出起点とする。

**repository（フラット実装 107 files）:** `accounting` 8 files が唯一のクラスタで、**残りはほぼ 1 ドメイン = 1 ファイル**（reservation_type 系 6 種・trimming 系 4 種・staff/shift 系 4 種など）。分割は service より機械的で、BE8-4 を先行させる根拠。
