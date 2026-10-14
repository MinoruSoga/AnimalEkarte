# C07 — XSS / render path review (出力・描画経路)

Campaign: emr-332-security-defense / Unit: U03-INPUT-OUTPUT / Attempt: att-u03-1
Template: `docs/work/emr/emr-332-security/baseline/case-verdict-template.md` (12 fields)
対象ツリー: `frontend/src/` + `frontend/liff/src/` + `frontend/line-reserve/src/` (3 app)。`frontend/src/shared-liff/` は liff/line-reserve が共有するコードを含む。

## 判定レコード

| # | Field | 記録 |
|---|-------|------|
| 1 | ケース ID | C07 |
| 2 | 対象リビジョン | worktree `feat/emr-332-security-defense-w03` — campaign base `de28e9765a29091339a872a85f1875a947e07e5b` (baseline/safety-record)。staging 実測なし |
| 3 | 実施方式 (method) | 行 1: `static` — rg による sink 全件棚卸し + 各 site の file:line 読取 / 行 2: `simulated` — memory render check は**未実施 (BLOCKED)**。node_modules が worktree に存在せず、`docker compose` frontend は共有リソースを触るため禁止 (prompt §Constraints)。task-isolated な依存 provisioned コンテナは利用不可 |
| 4 | 正常条件 | 期待: ユーザー由来テキストが HTML sink へ到達しない / 観測 (static): HTML 注入 sink は **0 件** — `dangerouslySetInnerHTML` 0 件、`innerHTML`/`outerHTML` 0 件、`document.write` 0 件、`eval(`/`new Function` 0 件、`srcDoc` 0 件、`insertAdjacentHTML`/`createContextualFragment`/`setHTMLUnsafe` 0 件、`<iframe>`/`<object>`/`<embed>` 0 件。`react-markdown` 使用 4 site は `rehype-raw` 非使用のため生 HTML ノードは描画されない。Markdown 内リンクは全て `getSafeMarkdownHref` (`src/lib/safe-markdown-href.ts`) 経由で scheme allowlist (http/https/mailto) + 制御文字除去 + `//` 拒否 |
| 5 | 拒否条件 | 期待: `javascript:`/`data:` scheme、`//host` protocol-relative、制御文字混入 URL が link href へ到達しない / 観測 (static): `getSafeMarkdownHref` が scheme を allowlist 判定し該当入力を `undefined` (無効化→`<span>` へ降格) にする — safe-markdown-href.ts:5-38。動的 href site は `tel:` 固定 prefix + 医院設定値 (line-reserve/src/pages/TopPage.tsx:40)、backend 生成 `plane_issue_url` (src/features/support/routes/BugReportsPage.tsx:83)、backend 生成画像 URL `img.src` (src/features/medical-records/components/ImageGalleryGroup.tsx:76) の 3 系のみ — いずれもサーバー供給値で、任意ユーザーテキスト由来ではない。`target="_blank"` は全 4 site で `rel="noopener noreferrer"` 付き (ManualContent.tsx:86-87, chat-markdown-components.tsx:34-35, BugReportsPage.tsx:84-85, ImageGalleryGroup.tsx:77-78)。`window.open` (src/lib/owner-report-window.ts:10) と `location.href` (src/lib/axios.ts:101) は内部パス + `parseInternalPath` allowlist (`src/lib/internal-navigation.ts:11-26` — `//`・制御文字・backslash・off-origin を拒否) |
| 6 | 防御判定 | **PASS (static の範囲)** — HTML 注入 sink 0 件・markdown 経路は sanitize 済み・動的 href はサーバー供給値のみ・`noopener noreferrer` 完備。render 実測は行 2 のとおり BLOCKED であり、static PASS を staging-observed/実行確認へ丸めていない |
| 7 | 副作用呼出回数 | **0** — render 非実行・静的読取のみ |
| 8 | 検出判定 | **N/A** — XSS 検出経路 (CSP 違反レポート等) は本ユニットの観測範囲外。`SecurityHeaders` middleware の CSP 設定存在は baseline/route-middleware-map §1 に記録済みだが本ケースで実測しない |
| 9 | 安全な証跡参照 | sink 棚卸し表は下記。「安全側」の根拠 site: `src/lib/safe-markdown-href.ts:5-38`, `src/features/manual/components/ManualContent.tsx:59-138` (MARKDOWN_COMPONENTS — a/img 上書き), `src/features/support/lib/chat-markdown-components.tsx:14-70` (CHAT_MARKDOWN_COMPONENTS), `src/lib/internal-navigation.ts:11-26`, `src/lib/axios.ts:96-104`, `src/lib/owner-report-window.ts:9-12` |
| 10 | 制約・未検証範囲 | 未検証: (a) **memory render check は BLOCKED** — React render→DOM 出力の実測 (vitest + jsdom/happy-dom) は node_modules 不在・compose 禁止のため実施不能。(b) `tel:`/`plane_issue_url`/`img.src` の 3 系はクライアント側で scheme 再検証していない — サーバー応答への信頼前提。backend が返す URL の健全性は別ケースの責務。(c) CSV export (`features/aggregation/lib/aggregation-csv.ts` の sanitizeCSVCell — formula injection 対策) はダウンロード物であり DOM sink ではないため本ケース外。(d) `console.log`/devtools 経路の DOM 操作は棚卸し対象外 |
| 11 | 修復案 | 該当なし — unsanitized sink 0 件のため FAIL/remediation candidate は発生しない。観測的 note (候補・非必須): `img.src`/`plane_issue_url`/`tel:` 値を client 側でも scheme allowlist 化すると defense-in-depth が上がる — 現状はサーバー生成値であり本ケースでは FAIL 扱いしない |
| 12 | 再検証結果 | 未実施 — 修復対象なし。render 実測は依存環境が供給され次第 `npx vitest run <path>` (task-isolated container 前提) で再検証要 |

## sink 棚卸し表 (3 tree 全件 — `src/` `liff/src/` `line-reserve/src/`)

### HTML 注入系 sink — 0 件 (全 tree・非テストファイル)

| sink パターン | 件数 |
|---------------|------|
| `dangerouslySetInnerHTML` | 0 |
| `.innerHTML` / `.outerHTML` | 0 |
| `document.write` / `eval(` / `new Function` | 0 |
| `srcDoc` / `insertAdjacentHTML` / `createContextualFragment` / `setHTMLUnsafe` | 0 |
| `<iframe` / `<object` / `<embed` | 0 |

### Markdown 描画系 — 4 site、全て sanitize 済み

| site | renderer | sanitize 状況 |
|------|----------|----------------|
| `src/features/manual/components/ManualContent.tsx:198` | react-markdown + remark-gfm | `rehype-raw` 非使用 → 生 HTML 不描画。リンクは `a` override → `getSafeMarkdownHref` (scheme allowlist)。画像は `img` override → `resolveImageSrc` (static bundle glob) |
| `src/features/support/components/HelpChat.tsx` (CHAT_MARKDOWN_COMPONENTS 使用) | react-markdown | 同上 — `a` override → `getSafeMarkdownHref` (chat-markdown-components.tsx:26-41) |
| `src/features/support/routes/ChatHistoryPage.tsx` (同上) | react-markdown | 同上 |
| `src/features/support/lib/chat-markdown-components.tsx` | (components 定義) | 同上 |

### 動的 href / URL 系 — 3 系、全てサーバー供給値または内部パス

| site | ソース | 判定 |
|------|--------|------|
| `line-reserve/src/pages/TopPage.tsx:40` `tel:${settings.phone_number}` | `/api/liff/:clinicId/settings` 応答 (医院設定 — staff 設定値) | server-supplied; `tel:` prefix 固定。scheme 再検証なし → note |
| `src/features/support/routes/BugReportsPage.tsx:83` `href={report.plane_issue_url}` | backend が Plane API 応答から生成して保存 (`internal/support/repository.go:226`) | server-generated; scheme 再検証なし → note |
| `src/features/medical-records/components/ImageGalleryGroup.tsx:52,76` `src/href={img.src}` | backend の画像 URL (uploads/S3 signed) | server-generated; scheme 再検証なし → note |
| `src/lib/owner-report-window.ts:10` `window.open(href,...)` | `paths.*` 内部パス + `encodeURIComponent` | safe |
| `src/lib/axios.ts:101` `location.href = login?from=...` | `parseInternalPath` allowlist (`internal-navigation.ts:11-26`) | safe |
| `src/features/auth/components/AuthProvider.tsx:397-400` `?from=${encodeURIComponent(sanitized)}` | `parseInternalPath` 経由 | safe |

## 証跡コマンド (再現用)

```sh
# HTML 注入 sink 全件 (3 tree)
grep -rn "dangerouslySetInnerHTML\|innerHTML\|outerHTML\|document\.write\|eval(\|new Function\|srcDoc\|insertAdjacentHTML\|createContextualFragment\|setHTMLUnsafe" \
  src/ liff/src/ line-reserve/src/
# → 0 件 (exit 1)
grep -rn "<iframe\|<object\|<embed" src/ liff/src/ line-reserve/src/   # → 0 件
# rehype-raw / sanitize 系依存
grep -E '"rehype|"remark|"marked|"dompurify' package.json liff/package.json line-reserve/package.json
# → package.json: remark-gfm のみ (rehype-raw なし)
# target=_blank の rel 欠落確認
grep -rn 'target="_blank"' src/ liff/src/ line-reserve/src/ | grep -v "noopener"
# → 0 件 (全 site rel="noopener noreferrer" 付き)
```

*Marker 文字列・攻撃構文・実ペイロード・資格情報・PII は本ファイルに含まない (A5)。XSS vector の説明は sink/ scheme 名の列挙に留め、実行可能な攻撃文字列は記述しない。*
