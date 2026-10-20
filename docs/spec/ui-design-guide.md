# UI デザインガイド — エージェント共通入口 (UI Design Guide for Agents)

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | spec.ui-design-guide／入口（routing）文書 |
| 決定状態 | 本書は新しい設計値・規約を決定しない。各節は既存正本へのリンクと要約のみ。要約と正本が食い違う場合は正本が勝つ |
| 実装状態 | 該当なし（文書のみ） |
| 実機検証状態 | 該当なし |
| 文書確認日・基準 | 2026-10-20／`main` の `DESIGN.md`・`docs/spec/design-*.md`・`frontend/scripts/design-system-audit.mjs` |
| 読者 | Claude Code・Codex・opencode・Devin など、`frontend/` の UI を実装・変更・レビューする全エージェントと開発者 |

> **目的**: UI 作業に入るエージェントが、ランタイムを問わず同じ順序で同じ正本を読み、同じ検査で完了判定できるようにする。
> Claude Code は `frontend/CLAUDE.md` を自動ロードするが、他のエージェントは自動ロードしない。本書はその差を埋める**唯一の入口**であり、第二の正本ではない。

---

## 1. 読む順（UI 作業の必須ステップ）

| 順 | 文書 | 何の正本か | いつ読むか |
|:--|:--|:--|:--|
| 1 | [design-system.md](design-system.md) | **色**（brand / primary / ink / 臨床・業務ステータス色）と SSOT の分担表 | すべての UI 作業の最初 |
| 2 | [`DESIGN.md`](../../DESIGN.md)（ルート・Animal Ekarte Design Language） | **タイポ / 形状 / 余白 / エレベーション / 動き（Motion の 4 段とトースト時間）/ コンポーネント寸法**（字義で従う）。色の値は持たない | 見た目の値・アニメーションを決める・変えるとき |
| 3 | [design-states.md](design-states.md) | hover / focus-visible / disabled / readonly / loading / invalid の**状態仕様**と WCAG 2.2 AA コントラスト | インタラクティブ要素を触るとき |
| 3b | [design-patterns.md](design-patterns.md) | **画面の組み立て**: ナビ現在地、一覧の列幅優先度・横スクロール、省略の全文到達、フォームの幅/検証/部品、固定要素の重なり、空状態/読込、カレンダー/ボード/グラフ、**表記**（日付・金額・空値 `—` は §10） | 一覧・フォーム・画面シェル・グラフを作る/直すとき、値を画面に出すとき |
| 4 | [design-token-layers.md](design-token-layers.md) | トークンの**配置と参照方向**（Primitive → Semantic → Component） | トークン追加・色変更・新規プリミティブ |
| 5 | [components/README.md](components/README.md) | `components/ui/` 7 プリミティブの**契約**（Anatomy・状態表・既知ギャップ） | `components/ui/*` を変更する前 |
| 6 | [ui-design-compliance.md](ui-design-compliance.md) | 準拠チェック **C1〜C20** の定義と全 88 ルートの正本対応表。`pnpm design-audit` が strict fail するのは **C1〜C29**。C21〜C24（`ui/` の stories 必須・`PALETTE` 層の純度・`BADGE` コントラスト 4.5:1・`shared/` の stories ラチェット）の定義は [design-token-layers.md §6](design-token-layers.md)、C25〜C29（省略の全文保持・ナビ省略・z-index・native 入力・バリデーション disabled）は compliance §1 | 新規ルート追加・監査失敗の調査・トークン / 共有 UI の変更 |
| 7 | [shared/ui-contracts.md](shared/ui-contracts.md) | 共通 UI・患者選択の**受入観点**（FR-common-ui-*） | 画面間で共有する部品・患者文脈を扱うとき |
| 実装 | `frontend/src/lib/design-tokens.ts` | 上記決定を実装する `PALETTE` / `C` / `STYLE` / `BADGE` / `LAYOUT` | コードを書くとき。**独自判断で値を足さない** |
| 補助 | [`.claude/refs/accessibility-rules.md`](../../.claude/refs/accessibility-rules.md) | semantic HTML / label / ARIA / キーボード操作 | フォーム・操作 UI を書くとき |

タスクに無関係な文書までまとめ読みしない（[AGENTS.md](../../AGENTS.md) Required First Step と同じ方針）。

## 2. ハードルール要約（違反は `design-audit` で strict fail、または必須レビュー指摘）

各行の根拠は右列の正本。詳細条件・例外は正本を読む。

| ルール | 根拠 |
|:--|:--|
| hex / rgba / 名前付き white・black の直書き禁止。色は `design-tokens.ts` 経由 | compliance C3 / C6b / C15 |
| Tailwind が走査できない実行時クラス合成（`hover:${C.x}`、`text-[${v}]`）禁止 | compliance C20、token-layers §3 |
| `text-[Npx]`・`rounded-[Npx]`・`shadow-[...]`・Tailwind 既定影・`tracking-*`・20px 系 `*-5` の任意値禁止 | compliance C9〜C14 / C16 / C17、DESIGN.md のスケール |
| ページは `PageLayout` または `Master*Page` shell。canvas は暖色 canvas-soft（純白ページ禁止） | compliance C2 / C4 / C8、design-system §2.2 / §4 |
| 新規リーフルートは **同一コミット**で ui-design-compliance §2 の表を更新。`PageLayout` / `Master*Page` を使う通常ルートは `design-system-audit.mjs` の C8 allowlist に**載せない**（allowlist に入ると shell の有無を検査しなくなり、後の shell 欠落を検出できない）。allowlist へ追記するのは正当な非 shell ページとルート helper だけ | compliance C8、`frontend/CLAUDE.md` |
| `focus-visible` インジケータの除去禁止。`outline-none` / `ring-0` を書いたら代替を書く | design-states §1.1（EMR-207 回帰防止） |
| 臨床 sentinel（死亡・危険・期限超過）とステータスは**色だけで示さない**（バッジ文字・アイコン併置） | design-states §1.1、`frontend/CLAUDE.md` 臨床安全境界 |
| readonly は disabled と区別（ink を落とさない）。loading はスピナー + `aria-busy` + 操作抑止 | design-states §1 |
| タッチターゲット 44×44px（dense 行アクション例外は design-system §7.2） | design-states §1.1 |
| 汎用 CTA は `colorVariant="primary"`（既定）。`brand` は認証・製品識別のみ | compliance C5、design-system §7.2 |
| `TableHead` / `TableCell` の typography・padding 上書き禁止。行全体 `onClick` 禁止（cell 内 native link / button を使う） | compliance C18 / C19、components/table.md |
| `components/ui/` の新規プリミティブは colocated `.stories.tsx` 必須（C21）。md 仕様書は「stories を読んでも分からない契約」がある場合だけ | compliance C21、design-system FE-DS2 決裁、components/README.md カバレッジ表 |
| story の axe 違反は失敗扱い（`.storybook/preview.ts` の `a11y.test: "error"`）。正当な例外は story 側の `parameters.a11y.config.rules` に理由付きで記録し、黙って `"off"` にしない | design-states §2、accessibility-rules.md |
| 省略（`truncate` / `max-w`）した要素には、focus とタッチの両方で開くツールチップか展開（「…もっと見る」）で**全文到達手段**を必ず付ける。`title` 属性は hover 専用なので補助にしかならず、単独では不可。識別子（ID・氏名・ペット名・金額・日付）とナビ項目は省略しない | patterns §3.1, §1.2、compliance C25 / C26（静的段は `title` / `aria-*` の有無のみ） |
| 一覧は識別列を折り返さず、長文列は `max-width` + 省略、操作列は右端 sticky。表が溢れたら識別列と操作列を残して中間列だけスクロールする。空値は `—`、件数は「1,234件」 | patterns §2 |
| サイドバーの現在地（`aria-current="page"`）は初期表示で見える位置へスクロールする。ページ `h1` は 1 つで、ページ内の見出しを `h1` より大きくしない | patterns §1 |
| 固定要素（浮遊ボタン・固定アクションバー・トースト）は操作要素に重ねない。スクロール末尾に固定要素分の余白を取り、`z-index` は `Z` トークンから選ぶ | patterns §5、compliance C27 |
| 業務・設定フォームの入力部品は共有 `ui/`（`DatePicker` / `NumberInput` / `MoneyInput` / `checkbox` / `switch`）。ネイティブ入力の許容範囲（検索・一覧フィルター・行内編集・時刻）は patterns §4.4。破壊操作は `Button variant="destructive"`（塗り）で `[削除] キャンセル 保存` の順 | patterns §4、design-states §1.1、compliance C28 |
| 保存ボタンを未入力理由で `disabled` にしない（押下→欄直下にエラー→先頭エラー欄へフォーカス）。`disabled` は送信中・権限なし・ワークフロー状態ゲート（確定済み等）だけ | patterns §4.2、design-states §1.1（2026-10-20 決裁）、compliance C29 |
| 本体画面のセクション見出しはページタイトル（title 20px）以下。heading-2 / heading-3 はダイアログ・大型サイドパネル・文書型画面のみ | design-system §3.4、patterns §1.4（2026-10-20 決裁） |
| カレンダー・シフト表の土日を青／赤で塗らない（赤は `C.danger` と衝突）。曜日は文字 + ink-muted、休診日は `C.bgPage` + 「休診」 | patterns §7.1、design-system §2.5（2026-10-20 決裁） |
| ステータス色を構造の面に使わない。唯一の例外は受付ボードの列面（design-system §2.5 の例外表）で、他画面へ広げない | design-system §2.5、patterns §7.2 |
| 空状態は「何が無いか／条件／次の操作」の 3 点。読込は 300ms 超でスケルトン | patterns §6 |
| 日付・時刻・金額・体重・件数・空値は `lib/format` / `lib/jst-date` の共有関数で整形する（`toLocaleString` / `toFixed` の直書き禁止）。日付 `yyyy/MM/dd`、金額 `¥1,234`、空値 `—` | patterns §10 |
| グラフの系列色は `PALETTE.chart*`、ステータス集計は `STATUS_COLORS`。凡例はドット + ラベル、系列は 4 本まで。動きは DESIGN.md Motion の 4 段のみ、位置の動きは reduced-motion で止める | design-system §2.5、patterns §7.4、DESIGN.md Motion |
| 新しい色・値を決めるときは design-system.md（色）/ DESIGN.md（寸法）に決裁（日付・決定者）を記録してからトークンへ流す。本書や他文書に第二正本を作らない | token-layers §2 |

## 3. 完了前チェック（Definition of Done・UI 変更を含む PR）

1. **機械検査**: `docker compose exec frontend pnpm design-audit` を実行し違反 0 件（`make ci` にも配線済み。スコープ限定の許可コマンドであり自動実行してよい）。
2. **スコープ限定テスト**: `docker compose exec frontend npx vitest run <変更した *.test.ts(x) のパス>`。位置引数はテストファイルのフィルタなので、本番ソース（`src/components/ui/button.tsx` 等）のパスを渡すと `No test files found` で exit 1 になる。本番ファイルから隣接テストを選びたいときは `docker compose exec frontend npx vitest related <src のパス> --run`。`pnpm test:run -- <path>` は全件実行になる罠（`frontend/CLAUDE.md` 参照）。
3. **状態の目視**: 変更した要素で hover / focus-visible（キーボード Tab）/ disabled / readonly / loading / invalid を実ブラウザまたは Storybook で確認。静的検査では C6a（臨床安全 UI）と状態の見え方は判定できない。
   - **見た目の回帰**: `frontend/**` を触る PR では `.github/workflows/chromatic.yml` が Storybook を Chromatic へ publish し visual / accessibility tests を実行する（`CHROMATIC_PROJECT_TOKEN` 未設定時は notice を出して skip する。skip された PR は Chromatic による保証がないので、その旨を報告する）。
   - **実機の準拠監査**: ルートの shell・canvas・テーブル様式・臨床 sentinel 表示を実ブラウザで検査する `frontend/e2e/ui-design-compliance-readonly.spec.ts`（4 viewport）がある。`./scripts/run-e2e.sh e2e/ui-design-compliance-readonly.spec.ts`（frontend ディレクトリ・ローカル Docker スタック起動済みが前提・ユーザー実行）。レイアウト・テーブル・PageLayout に触れた PR では実行する。同 spec は design-patterns §1.1（現在地が viewport 内）と §5.1（固定要素の非重なり）も採取し、既定では annotation に記録するだけ、`UI_DESIGN_PATTERNS_STRICT=1` で fail させる（HTML レポートの annotations を確認する）。この spec の route inventory は 88 ルートを全て含まない（`/examinations/new`・`/settings/bug-reports`・`/settings/chat-history` が未登載 — [docs/ops/testing/README.md](../ops/testing/README.md)）ため、結果を `ui-design-compliance.md` に記録するときは「実行した spec の対象ルート数」を明記し、未登載ルートが残る間は最終 runtime 監査日を**全 88 ルートの監査として更新しない**（部分監査として記録する）。未登載ルートを spec へ追加した PR だけが全件の監査日を更新できる。
4. **文書同期**: 新規ルート → compliance §2（C8 allowlist は非 shell ページだけ）。新規 `ui/` プリミティブ → stories（必要なら `components/_TEMPLATE.md` から md）。新しい色・値 → 正本へ決裁記録。
5. **実データ幅での確認**: 一覧・フォームを触った PR は、氏名 10 文字・ID 16 桁・長文 100 文字の値を入れて、識別列の折り返し・省略の全文到達・操作列の可視・現在地の可視・固定要素の非重なり（[design-patterns.md](design-patterns.md) §1〜§5）を 1280px と 768px で確認する。STG の合成データ（`UAT20261015_*`）で崩れる場合は実運用値でも再現するかを見て報告する。
6. **報告**: 実行した検査コマンドと結果、目視した状態、更新した文書を PR 本文に書く。実行していない検査を PASS と書かない。

## 4. ランタイム別の読み込み方

| ランタイム | 本書への到達経路 |
|:--|:--|
| Claude Code | `.claude/CLAUDE.md` Context Loading 表 → 本書。`frontend/CLAUDE.md` は自動ロード |
| Codex | [`.codex/CODEX.md`](../../.codex/CODEX.md) → [AGENTS.md](../../AGENTS.md) Frontend / UI Work → 本書 |
| opencode | `AGENTS.md`（自動読込）→ 本書。`opencode.json` の instructions には **追加しない**（UI 以外のタスクで 400 行超の規約を常時ロードしないため） |
| Devin / その他 | リポジトリ直下の `AGENTS.md` を最初に読み、Frontend / UI Work 節から本書へ |

いずれのランタイムでも、本書のリンク先を**実際に開いて読む**こと。本書の要約だけで実装しない。
