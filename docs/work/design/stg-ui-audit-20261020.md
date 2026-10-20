# STG UI 監査記録（2026-10-20・88 ルート中 79 ルート）

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | work.design.stg-ui-audit-20261020／実施記録 |
| 対象 | `https://stg.noah-karte.com`（staging）。`ui-design-compliance.md` §2 の 88 製品ルートのうち 79 ルートを目視 + DOM 計測。1280×900、代表 7 ページ（`/`, `/owners`, `/medical-records`, `/reservations`, `/accounting`, `/medical-records/new`, `/settings/staff`）は 768×1024 も確認 |
| 実施者 | Claude（巡回は Sonnet サブエージェント、重要度 H の再現確認と切り分けは Fable）。依頼者: 曽我 |
| アカウント | デモ「林 文明（執行）」。デモアカウントのパスワードはログイン画面が自動入力 |
| 前提・制約 | STG データは UAT 用合成データ（`UAT20261015_*` の長い名称）を多く含み、列幅・省略の問題を実運用より強調している可能性がある。監査中に staging の Backend Deploy が migration で失敗し（run 39356679343）、復旧まで約 40 分中断。セッションが 2 回切れ再ログインした |
| 副作用 | `/medical-records/select-pet` の「選択」を 1 回押下した結果、**カルテ 1 件（作成中）が自動作成され、受付ボードに「診療中」カード 1 件が残った**。削除は監査方針で禁止のため未処置。STG 側での処分が必要。対象のカルテ番号・飼主・ペットの識別子は本書に記載せず、依頼者へ別経路で伝達済み（staging であっても診療記録の識別子をリポジトリに置かない）。※「選択」= 即レコード作成という挙動自体は product-philosophy の不可逆操作の観点で要確認 |
| 未確認 | `/reset-password` の入力フォーム（有効トークンが必要）、`/identity-links`（`/` へリダイレクト。権限か未実装かは不明）、`/medical-records/new` のフォーム本体（select-pet へリダイレクト）、作成・編集・送信の操作系、モバイル幅 375px、`:id` ルートは各 1 例のみ |
| 成果物 | 本書（要約と違反一覧）、[design-patterns.md](../../spec/design-patterns.md)（規定が無かった事項の規約化）、`DESIGN.md` / `design-states.md` / `ui-design-guide.md` への追記。生ログ（79 ページ分の所見）は PR #733 のセッションに保持し、リポジトリには要約のみ置く |

## 1. 方法

1. 各ルートを開き、DOM 計測スクリプトで次を取得: 横はみ出し（`scrollWidth > clientWidth` の要素）、44px 未満のタッチ面、幅 90px 未満で 2 行以上に折り返した要素、`alt` 無し画像、名前の無いボタン、ラベル無し入力、inline の色指定、使用 font-size / shadow / radius、`th` の computed style、行 `onClick`、空状態文言。
2. スクリーンショットを目視し、折り返し・省略・整列・色・テーブル・フォーム・空状態・アイコンボタン・ナビの 9 観点で記録。
3. 重要度 H（読めない・押せない・見切れ）の項目は Fable が別途 `getBoundingClientRect` / `elementFromPoint` / computed color で再現確認した。
4. 所見を「実装の違反（既存正本に規定があり、守られていない）」「ガイドに規定が無い」「判断不能（データ起因・要件側）」に切り分けた。

## 2. 結論

- **「デザインガイドの問題か」への答え: 両方。** 違反 54 件相当は既存正本（DESIGN.md / design-system.md / design-states.md / compliance C1〜C24）で既に禁止されているものの実装漏れ。48 件相当は**正本のどこにも規定が無い**領域（ナビの現在地、列幅の優先度、省略の全文到達、固定要素の重なり、フォームの幅・検証・部品統一、空状態・読込の最小要件、カレンダー列数・曜日色、見出し階層、権限バッジ）で、これを [design-patterns.md](../../spec/design-patterns.md) に規約化した。
- 既存の機械検査（C1〜C24）は「禁止クラスを使っていないか」を見るもので、今回の崩れの大半（列幅配分・重なり・省略・現在地）は**レイアウトの結果**であり静的検査では拾えない。design-patterns.md §9 に静的検査 5 件と e2e 2 件の追加候補を置いた。

## 3. 重要度 H（13 系統・Fable 再現済みは ✔）

| # | 現象 | ページ | 分類 | 規約 |
|--:|:--|:--|:--|:--|
| 1 | ✔ サイドバーの現在地が画面外（nav `scrollHeight 2078 / clientHeight 748`、現在地 `y=1683`、自動スクロール無し） | `/settings/*`, `/lstep/*`, `/line-reservation/*`, `/manual`, `/lab-device` 等 10+ | 規定無し→新設 | patterns §1.1 |
| 2 | ✔ サイドバー項目の省略（`title` 無し）: 権限グループマスタ 143/137、物販・フードマスタ 144/137、割引キャンペーンマスタ 175/137、シフトテンプレート 144/137、問診テンプレートマスタ 176/113、主訴カテゴリマスタ 144/113 | `/settings/*` 全般 | 規定無し→新設 | patterns §1.2, §3.1 |
| 3 | ✔ 浮遊ヘルプボタン（fixed, z-index 9999, 48×48, x1216–1264/y836–884）が「最後のページ」ボタン（44×44, y856–900）の上半分を覆う。`elementFromPoint` 上半分 = FAB | `/accounting`（ページネーション付き一覧全般で再現見込み） | 規定無し→新設 | patterns §5.1 |
| 4 | ✔ 飼主編集のペット行「受付」アイコンボタン: 背景 `rgb(2,112,120)` に対しアイコン `rgb(97,93,89)`（ink-muted）。塗りボタンにアイコン色トークンが継承されていない | `/owners/:id` 全行 | **実装の違反**（design-states §1: on-primary 文字 / コントラスト 3:1） | design-states §2 |
| 5 | 会計明細表: 左カード 610px に 11 列（sw 896〜1215）。sticky「金額」が「数量」に被さり、ヘッダが「数」「税分」と途中で切れる。数量・割引が初期表示で見えない | `/accounting/new`, `/accounting/:id` | 規定無し→新設 | patterns §2.2, §2.3 |
| 6 | 見積番号（識別子）が 85〜90px の列で 2 行に折り返し（`310862-001-` / `00000001`） | `/estimates` 全行 | **実装の違反**（EMR-379 系）+ 規定補強 | patterns §2.1 |
| 7 | シフト表: 固定スタッフ列 119px で氏名欠け（173/119, 178/119, 300/119、`title` 空）。日付セル 47px で「全日 09:00〜18:00」が 3 行 | `/shifts` | 規定無し→新設 | patterns §7.1, §3.1 |
| 8 | 主訴列 200px `truncate`、`title` 無し（sw 456/438/335）。臨床情報に一覧から到達できない。「会計」バッジが「会/計」に折り返し。操作列が初期表示外（1090/1034） | `/medical-records` | **実装の違反**（truncate）+ 規定無し（全文到達） | patterns §3.1, §2.4, §2.2 |
| 9 | ケージカードのケージ名が省略記号なしで clip、または「UAT-ZB ケ／ージ1」と語中で折り返し。「詳細」ボタン 38px 幅 | `/hospitalization` | 規定無し→新設 | patterns §7.2, §3.2 |
| 10 | 入院タイプのラジオ 16×16、ケージ連結選択チェックボックス 16×16（ラベルを含むヒット領域なし） | `/hospitalization/:id/edit`, `/settings/cage` | **実装の違反**（design-states §1.1 44px）+ 規定補強（ヒット領域の定義） | patterns §4.4 |
| 11 | 表全幅が横スクロールし、ステータス・操作列が画面外（`/trimming` 1369/1034、`/checkups` 1305/1034 — 長文の健診種別が折り返されず表を広げる） | `/trimming`, `/checkups` | 規定無し→新設 | patterns §2.1, §2.2 |
| 12 | 名前の無いトグル 7 個（`button.peer`、テキスト・`aria-label` 無し）、定休曜日チェック 16px | `/line-reservation/settings` | **実装の違反**（accessibility-rules / C 系）+ 規定補強 | patterns §4.4 |
| 13 | 768px: select-pet でペット名が「選択」ボタンの下に潜る。`/accounting` 支払方法 81px で「クレジットカード」3 行。`/settings/staff` 職種 59px で縦書き | 代表 3 ページ | **実装の違反**（列幅） | patterns §2.1, §2.2 |

## 4. 実装の違反（既存正本で禁止済み・M 以上の主なもの）

| 現象 | ページ | 根拠 |
|:--|:--|:--|
| `th` padding `12px 0px`（規定 `12px 16px`）。見出しが本文より約 12px 左にずれる | `/settings/treatment-items`, `diagnosis`, `animal-species`, `medicine`, `cage`, `merchandise-items`, `permission-groups`, `shift-templates` | DESIGN.md `ex-data-table-cell`、C18（raw th allowlist の見直し） |
| `/settings/reservation-type` のグループ見出し `th` が 16px/700/padding 0 | `/settings/reservation-type` | 同上 |
| 入力高さ 36px（`h-9` 検索）・38px（日付ボタン）が 44px の入力と混在 | `/medical-records/:id`, `/trimming/:id` | design-states §1.1、design-system §3.4 compact 制御 |
| 業務・設定フォームでのネイティブ date / 16px チェック（`/accounting/close` 対象日、`/settings/closing-time` 曜日）。※ `/lstep/*` のネイティブ date は検索フィルターなので design-system §3.4 の許容範囲 | `/accounting/close`, `/settings/closing-time` | design-system §3.4「日付・数値入力」 |
| 削除ボタンのスタイル不統一（赤アウトライン vs 赤テキスト）、保存の隣に配置 | `/examinations/:id`, `/vaccinations/:id`（赤テキスト）／`/trimming/:id`, `/hospitalization/:id/edit`（赤アウトライン） | components/button.md variants、patterns §4.5 |
| ~~hex 直書きクラス（`text-[#615D59]` 等）~~ → **再分類: 違反ではない**。ブラウザ DOM で観測した `text-[#615D59]` / `#E6E6E6` / `#F6F5F4` / `#C0392B` / `#FFF3CD` は `design-tokens.ts` の `C` / `STYLE` が展開した正規クラスであり、該当ルートのソースに直書きは無い（2026-10-20 に `frontend/src` を grep して確認。`design-tokens.ts` と test 以外に一致なし）。C3 は `design-tokens.ts` を除外したソース検査なので、DOM 上のクラス名は C3 違反の証拠にならない | `/owners/:id`, `/inventory`, `/settings/bug-reports` | — （C3 は PASS。DOM 観測から C3 を起票しない） |
| 見出し階層の逆転（セクション 22px > h1 20px、`/lab-device` は 40/22/20 混在） | `/hospitalization/:id/edit`, `/lstep/analytics`, `/lab-device`, `/manual`, `/settings`（h1 二重） | design-system §3.4 ロール表、patterns §1.4 |
| 空状態が 1 行のみ（次の操作なし） | `/settings/hospitalization`, `insurance`, `campaigns`, `inquiry-templates`, `lstep/tags`, `/accounting/close/history`, `/lstep/delivery-monitor`, `/lstep/analytics`×3, 各編集画面の履歴カード | DESIGN.md `ex-empty-state-card`、patterns §6.1 |
| 読込中がプレーンテキストのみ約 10 秒 | `/aggregation` | design-states §1 loading、patterns §6.2 |
| 「期限切れ」バッジの有無で日付列の開始位置がずれる | `/vaccinations` | patterns §2.4 |
| 「対応済み」バッジ 58px で 2 行、状態列に 5 種類の部品混在、Plane リンク 20px 高 | `/settings/bug-reports` | patterns §2.4、design-states §1.1 |
| トーストの閉じるボタン 20×20、ヘッダ操作の上に重なる | `/hospitalization/:id` | design-states §1.1、patterns §5.2 |
| 参照リンク高さ 16px | `/settings/chat-history` | design-states §1.1 |
| 「今月誕生日」バッジ兼ボタン 28×20 | select-pet 共通 | design-states §1.1、components/badge.md（バッジに操作を持たせない） |
| 保存ボタンが理由表示なしで disabled | `/estimates/new`, `/estimates/:id/edit` | patterns §4.2（新設） |
| 「画面イメージ」見出しの下が空 | `/manual/*` | patterns §8 |
| 部門別合計の項目名が内部キー（`procedure`, `trimming`）、「飼主様様 - ペットちゃん」のプレースホルダ | `/accounting/reports`, `/accounting/new` | 文言バグ |

## 5. 判断不能・要件側で確認するもの

- select-pet の「選択」押下で即カルテ作成 + 受付「診療中」遷移（キャンセル時の破棄手段が見えない）。
- `/estimates/:id` の集計表示（小計 ¥250,206 / 消費税 ¥227,460 / 合計 ¥227,460）と、`/estimates/new` で小計・消費税・合計が手入力欄になっている点（二重入力の疑い）。
- ~~受付ボード列のパステル背景~~ → 2026-10-20 決裁: 受付ボードの列面に限った例外として許可（design-system §2.5 例外表）。
- `/settings/inquiry-templates` と `/settings/interview/templates` が同一画面（ルート重複）。`/identity-links` が `/` へリダイレクト。
- セッションが約 20〜40 分で切れる（refresh の有効期限か、STG の再起動起因かは未確認）。
- `/accounting/reports` の集計サマリーが日次明細 31 行の下にある（結論→詳細の順でない）。情報設計の判断。
- 検査一覧「結果概要」列が全行同じ値（検査名）で情報量が無い。

## 6. 次のアクション（提案）

1. ~~PR #733 のレビューで design-patterns.md の各規則を決裁~~ → 2026-10-20 決裁済み（§1.4 見出し上限、§4.2 保存 disabled 禁止、§7.1 曜日色禁止、§7.2 受付ボード列背景は例外許可）。実装側の修正対象: `/estimates/new`・`/estimates/:id/edit`（保存 disabled）、`/shifts`・`/line-reservation/slots`・`/reservations`（曜日色）、見出し逆転 4 画面。
2. 実装の修正は Plane で起票（本書は起票しない）。優先: H-1 ナビ現在地（1 箇所の修正で 10 ページ以上に効く）→ H-3 FAB 余白 → H-4 アイコン色 → H-5 会計明細 → 一覧の列幅（EMR-379 の拡張）。
3. design-patterns.md §9 の C25〜C27 を `design-system-audit.mjs` に追加し、e2e `ui-design-compliance-readonly.spec.ts` に「現在地が viewport 内」「FAB と操作要素の非重なり」を追加する。
4. 監査中に作成された STG のカルテ 1 件と受付カード 1 件の処分（識別子は依頼者へ別経路で伝達済み）。
