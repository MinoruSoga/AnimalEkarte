# 画面パターン規約 (Screen Patterns) — ナビ・一覧・フォーム・固定要素・空状態

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | spec.design-patterns／画面パターンの正本 |
| 決定状態 | **合意済み**（2026-10-20 STG 79 ルート監査で「既存正本に規定が無い」と判定した事項を規約化し、PR #733 で曽我が決裁。決裁事項: §1.4 セクション見出しは title 以下、§4.2 保存ボタンを未入力理由で disabled にしない、§7.1 曜日の青／赤を禁止、§7.2 受付ボード列の背景は範囲付き例外として許可）。既存正本（DESIGN.md / design-system.md / design-states.md）と重なる値は再定義せずリンクする |
| 実装状態 | **未実装が多数**。各規則の末尾に監査で観測した違反ページを記す（[監査記録](../work/design/stg-ui-audit-20261020.md)）。本書の掲載は実装済みを意味しない |
| 実機検証状態 | 監査は STG `stg.noah-karte.com`、2026-10-20、1280×900（代表 7 ページは 768×1024）、デモアカウント「林 文明（執行）」。UAT 用の長い合成データ（`UAT20261015_*`）が列幅問題を強調している可能性がある |
| 文書確認日・基準 | 2026-10-20／PR #733 |
| 読者 | フロントエンド実装者・レビュアー・全 AI エージェント |

> **位置づけ**: 色は [design-system.md](design-system.md)、寸法・タイポは [DESIGN.md](../../DESIGN.md)、状態は [design-states.md](design-states.md)、部品契約は [components/](components/README.md) が正本。本書は**それらを画面として組み合わせるときの規約**（画面シェル・一覧テーブル・フォーム・固定要素・空/読込状態・カレンダー/ボード）を定義する。個別画面の仕様は `screens/` / `domains/` が正本で、本書と矛盾する場合は画面仕様に決裁記録があるほうを優先し、無ければ本書に従う。

---

## 1. 画面シェルとナビゲーション

<a id="shell-current-location"></a>
### 1.1 現在地は初期表示で見える

- サイドバーの現在地（`aria-current="page"`）は、ページ表示直後に**スクロールなしで視認できる**こと。ナビが縦スクロールする場合は、マウント時と遷移時に現在地を `scrollIntoView({ block: "nearest" })` で可視範囲へ入れる。
- ナビが可視高さを超えるときは、上下端にスクロール可能であることを示す手掛かり（フェード or 影）を出す。アカウント行などの固定フッタの下に項目を隠さない。
- 観測: `/settings/*`・`/lstep/*`・`/line-reservation/*`・`/manual` など 10 ページ以上で nav `scrollHeight 2078 / clientHeight 748`、現在地 `y≈1190〜1700` が画面外。原因は 1 か所（nav の自動スクロール無し）。

<a id="shell-nav-label"></a>
### 1.2 ナビ項目は省略しない

- サイドバーの項目名は省略（`truncate`）しない。収まらない名前は**名前を短くする**（例「割引キャンペーンマスタ」→「キャンペーン」）か、ネストの字下げを減らす。どうしても省略する場合は §3.1 の全文到達手段が必須。
- 観測: 「権限グループマスタ」「物販・フードマスタ」「割引キャンペーンマスタ」「シフトテンプレート」「問診テンプレートマスタ」「主訴カテゴリマスタ」が `title` 無しで末尾切れ。

<a id="shell-page-title"></a>
### 1.3 ページタイトルは 1 つ・文脈を含む

- `h1` は画面ごとに 1 つ。ヘッダバーの `h1`（`{typography.title}` 20px/600）と本文先頭の大見出しを**重複させない**（観測: `/settings` で 20px と 26px の「マスタ設定」が二重）。
- `h1` 単体で機能が特定できる名前にする。「基本設定」「ページ編集」のような汎用名は、親機能名を含める（「LINE予約 基本設定」）か、パンくずを `h1` の直上に置く（観測: `/line-reservation/settings`・`/page-editor`）。
- ページ名は サイドバー／`h1`／`ui-design-compliance.md` §2 の表で**同一文字列**にする（観測: 「レジ締め履歴」と「締め履歴」の揺れ）。
- PageLayout / FormHeader のタイトルにはアイコンを付けない。セクション・パネル・ボタン・サイドバー内のアイコンは対象外（2026-10-20 に design-system §3.4 から移管）。
- 説明文付きヘッダ（`h1` + 補足 12px）を使う場合も、ヘッダの高さと上余白は 1 行ヘッダと同じ基準（`PageLayout` の既定）を守る。上端に貼り付けない。

<a id="shell-heading-scale"></a>
### 1.4 ページ内の見出し階層

- 見出しロールと実装クラスの正本は [design-system.md §3.4](design-system.md)（title 20px = PageLayout タイトル、heading-2 / heading-3 = ダイアログ・大型サイドパネル・文書型画面のみ）。本書は値を再定義しない。
- 配置の規則: `h1` は 1 画面に 1 つ（§1.3）。セクション見出しは `h2`、カード内の小見出しは `h3` と、DOM の見出しレベルを飛ばさない。
- KPI などの**数値の強調**は見出しではないので、`text-heading-*` ではなく数値専用のサイズを 1 画面で 1 段階に決めて揃える（観測: 20 / 26 / 40px が画面ごとに不揃い）。
- **本体画面のセクション見出しは title（20px/600）以下**（2026-10-20 決裁）。ページタイトルより大きいセクション見出しを置かない。heading-2 / heading-3 はダイアログタイトル・大型サイドパネル・文書型画面（`/manual/*`、単独ウィンドウ画面）に限定する（design-system §3.4 のロール表も同日改訂）。観測: `/hospitalization/:id/edit`・`/lstep/analytics`・`/lab-device`・`/manual` でセクション 22px > h1 20px の逆転。

<a id="shell-permission-badge"></a>
### 1.5 権限バッジは操作ボタンの列に置かない

- 「閲覧」「作成」「編集」などの権限表示は、ヘッダ右の操作ボタン群に混ぜない。表示する場合は `h1` の直下に `C.textInkMuted` の 12px テキスト（例「権限: 閲覧・編集」）で 1 行にまとめ、`title` で意味を補う。
- 出し分け基準（どの画面で出すか）を 1 つに決めて全画面で揃える。
- 観測: `/accounting/close`・`/accounting/reports`・`/lstep/checkup-sync`・`/lab-device`・`/settings/clinic`・`/settings/payment-methods`・`/settings/lab-device-item-masters` でバッジがボタンに隣接し、ボタンと誤認。同系のマスタ画面でも出る画面と出ない画面がある。

---

## 2. 一覧テーブル

> 基本様式（ヘッダ 12px/600、セル padding 12px/16px、行 onClick 禁止、セル上書き禁止）は [DESIGN.md Data Table](../../DESIGN.md) と [components/table.md](components/table.md)。本節は**列の設計**を規定する。

<a id="table-column-roles"></a>
### 2.1 列の役割と幅

| 列の役割 | 例 | 幅・折り返し |
|:--|:--|:--|
| **識別列（1 行固定）** | ID・番号・日付・金額・ステータス | `whitespace-nowrap` + 内容に合わせた `min-width`。省略しない |
| **識別列（氏名・ペット名）** | 飼主名・ペット名・ケージ名 | `nowrap` は使わない。`min-width` を全角 10 文字（約 150px）分確保し、それを超える分だけ語の区切りで折り返す（§3.2）。**1 文字ずつの折り返しを起こさない**。省略しない |
| **長文列** | 主訴・スタイル希望・備考・件名・説明 | `max-width` を決めて `truncate`。**必ず全文到達手段**（§3.1）。長文列が表全体を押し広げてはならない |
| **操作列** | 編集・⋯メニュー・選択 | 右端固定（`sticky right-0`）。横スクロール時も常に見える |
| **補助列** | 体重・環境・前回来院など | 余った幅を使う。識別列を圧迫するなら `md`/`lg` で非表示にする |

- 列幅は**識別列 → 操作列 → 長文列 → 補助列**の順に確保する。補助列に余白があるのに識別列が折り返している状態を禁止する。
- 観測: `/owners`・`/medical-records`・`/trimming`・`/checkups`・`/estimates`（見積番号が 2 行）・`/accounting`（768px で支払方法が 3 行）・`/settings/staff`（768px で職種が縦書き）など約 15 ページ。`/checkups` は長文の健診種別が折り返されず表を 1305px に広げ、担当医・操作が画面外。

<a id="table-overflow"></a>
### 2.2 横スクロールの条件

- 表がコンテナを超えるときは、**識別列（先頭 1〜2 列）とステータス・操作列を sticky で残し**、中間の補助列だけがスクロールする。初期表示で操作列が見えない状態を禁止する。
- sticky 列は背景を不透明（`C.bgWhite`）にし、スクロール側との境界に hairline + 影（`shadow-level1` 相当の片側影）を出す。sticky 列が他の列の**上に被さって見出しを隠す**ことを禁止する。
- 横スクロールがあることは、右端のフェードまたはスクロールバーの常時表示で示す。「横にスクロールしてください」という注意文で代替しない。
- 観測: `/trimming`（1369/1034）・`/checkups`（1305/1034）・`/shifts`（1732/1036）・`/medical-records`（1090/1034、操作列が初期表示外）・select-pet 共通（1213/1034）。`/accounting/new`・`/accounting/:id` の明細表は 610px の左カードに 11 列を入れ、sticky「金額」が「数量」に被さる。

<a id="table-side-by-side"></a>
### 2.3 2 カラム内の表

- 2 カラムレイアウトの片側に列数の多い表を置く場合、**表側のカラムを広く**する（最低 2/3）。それでも収まらない列は「展開行」か「詳細ダイアログ」へ逃がし、横スクロールで核心列（数量・割引など業務上必須のもの）を隠さない。
- 観測: 会計明細（左 610px に 11 列、右カラムは余白が大きい）。

<a id="table-badges"></a>
### 2.4 セル内バッジ・補助要素

- バッジは `whitespace-nowrap` + `min-w` を持ち、**2 行に折り返さない**（観測: `/medical-records` の「会計」が「会/計」、`/settings/bug-reports` の「対応済み」）。
- バッジの有無で同じ列の値の開始位置を変えない。バッジは専用の小列か、値の後ろ（右側）に置く（観測: `/vaccinations` の「期限切れ」で次回予定日が +55px ずれる）。
- 1 つの列に「バッジ／ボタン／リンク／テキスト」を混在させない（観測: `/settings/bug-reports` の状態列に 5 種類）。

<a id="table-empty-value"></a>
### 2.5 空値・件数・ページサイズの表記

- 空値の字形と色は §10（`—` + `C.textInkMuted`）。`-`・`–`・空白・「未設定」を画面ごとに混ぜない（観測: `/settings/clinic`・`/settings/diagnosis`・`/estimates` で 3 種類）。
- 件数は上部に**総件数**「1,234件」（表記は §10）、ページネーション側に「1–20件 / 1,234件」。同じ画面で「20件」（ページ内件数）と「14,690件中」を混ぜない（観測: `/examinations`、`/settings/*` の「2 件」「55 件」）。
- 取得上限（100 件など）で打ち切る一覧は、**上限に達していることを件数の隣に明示**する（観測: `/trimming`・`/vaccinations`）。
- ページサイズは既定 20。変える場合は画面仕様に理由を書く（観測: 10 / 20 / 50 が混在）。

<a id="table-reorder"></a>
### 2.6 並べ替え操作は 1 方式

- 手動並べ替えを持つマスタ一覧の**可視の操作は 1 方式**（ドラッグハンドルか上下ボタン）。同じ行にハンドルと↑↓ボタンを並べない。50 件を超える一覧ではドラッグを使わない（観測: `/settings/merchandise-items` が 685 件でハンドル＋↑↓の 2 方式）。
- ドラッグを採用する場合は、WCAG 2.2 SC 2.5.7（Dragging Movements）により**ドラッグを使わない代替**を必ず持つ: ハンドルに focus して矢印キーで移動、または行の ⋯ メニューの「上へ／下へ」。代替は可視の第二操作ではなく、キーボードとシングルタップで到達できる手段として用意する（[accessibility-rules.md §9](../../.claude/refs/accessibility-rules.md)）。
- 並べ替え可能な一覧は、ハンドルに `aria-label` と `title`（「ドラッグして並べ替え」）を付ける。

<a id="table-header-padding"></a>
### 2.7 ヘッダ余白

- `th` の padding は DESIGN.md `ex-data-table-cell` の `12px 16px` のみ。`12px 0px` は禁止（観測: `/settings/treatment-items`・`diagnosis`・`animal-species`・`medicine`・`cage`・`merchandise-items`・`permission-groups`・`shift-templates` で 0px。見出しが本文より約 12px 左にずれる）。compliance C18 の raw `th` allowlist を見直し、設定マスタの表も `TableHead` 経由にする。

---

## 3. 省略・折り返し・全文到達

<a id="truncate-rule"></a>
### 3.1 省略（…）の条件

- 省略してよいのは**長文列・補足テキスト・ナビ以外のラベル**だけ。識別子（ID・番号・氏名・ペット名・ケージ名・金額・日付）は省略しない。
- 省略した要素には**全文へ到達する手段を必ず 1 つ**付ける: focus とタッチの両方で開くツールチップ（トリガーは focusable な要素）、または展開（「…もっと見る」）。`title` 属性は hover 専用でキーボード・タッチでは出ないため、**補助にはなるが単独では不可**。スクリーンリーダー向けに `aria-label` か `sr-only` で全文も提供する（これも視覚ユーザーへの到達手段にはならない）。
- 固定幅（`max-w-[80px]` など）で省略する場合、その幅は内容が判別できる文字数を基準に決める（長文列なら先頭 20 文字程度が読める幅）。識別子は §2.1 のとおり省略しないので、この規則の対象外。
- 観測: `/` 受付カードの「担当医未設定」「予約内容」、`/medical-records` 主訴（200px、`title` 無し）、`/trimming` スタイル希望、`/shifts` スタッフ名（`title` 空）、各編集画面ヘッダの飼主名（64〜140px で `…`）、`/hospitalization` ケージ名が省略記号なしで clip。

<a id="wrap-rule"></a>
### 3.2 折り返し

- 日本語の固有名詞・複合語は `word-break: keep-all` 相当で**語の途中で折らない**（観測: 「UATペッ／ト」「UAT-ZB ケ／ージ1」「ALAN WAIN／WRIGHT」）。
- 折り返しで行の高さが変わる列は、行高を揃えるために `min-height` を使わず、**列幅を見直す**（§2.1）。

---

## 4. フォーム

<a id="form-width"></a>
### 4.1 幅とグリッド

- 新規／編集フォームのレイアウトは 2 種類だけ: **(A) 単一カラム中央**（マスタ・設定）、**(B) 全幅 12 列グリッド**（飼主・カルテなど業務記録）。既存は 530 / 580 / 2/3 幅 / 全幅が混在しており、新規実装は A か B を選ぶ。幅の値は本書では決めない: 最大幅は `LAYOUT.pageContentMaxWidth` の既存段（`form` / `formMid` / `formNarrow`）から選び、より狭い段が必要なら [DESIGN.md Grid & Container](../../DESIGN.md) に決裁を記録してからトークンへ追加する。
- グリッドでは**1 行のラベル高さを揃える**（長いラベルは短くするか、2 行を全列で許容）。同じ行で入力欄の上端がずれる状態を禁止する（観測: `/owners/new`・`/lstep/checkup-sync`）。
- セグメントコントロール（会員区分など）は**選択肢が 1 行に収まる幅を確保**する。収まらなければ `Select` にする。選択肢の一部だけ 2 行目に落ちる状態を禁止する（観測: `/owners/new`・`/owners/:id` の「他診/準」、`/line-reservation/settings` の定休曜日「日」）。
- 40 項目を超える設定フォームは、セクションごとにカードまたはアコーディオンで分節し、各セクションに見出しを付ける（観測: `/settings/integrations/lstep`）。

<a id="form-validation"></a>
### 4.2 検証と保存ボタン

- 保存ボタンの `disabled` 条件（未入力理由の disabled 禁止、押下→欄直下にエラー→先頭エラー欄へフォーカス、`disabled` は送信中・権限なし・ワークフロー状態ゲートのみ）の正本は [design-states.md §1.1](design-states.md)（2026-10-20 決裁）。本書は画面側の追加規則だけを置く。
- 既存データが現在の必須条件を満たさない場合（レガシーデータ）は、開いた時点で該当欄にエラー文を表示し、何を直せば保存できるかを示す。
- 観測: `/estimates/new`・`/estimates/:id/edit` が理由表示なしで保存 disabled（他の新規画面は押下→エラー表示で、UX が 2 流儀）。

<a id="form-required"></a>
### 4.3 必須・任意の表記

- 必須マーカーは共有 `RequiredMark` を使い、間隔は親ラベルの `gap-2`（8px）に統一する。マーカーは装飾として `aria-hidden`（2026-10-20 に design-system §3.4 から移管）。必須は `*` のみで示し、「（任意）」「（必須）」の文字併記を同じフォームで混ぜない（観測: `/lstep/checkup-sync`）。
- 必須欄が空のまま存在する既存データは §4.2 に従う。

<a id="form-controls"></a>
### 4.4 入力部品

- 業務フォーム・設定フォームの日付・数値・金額は共有 `DatePicker` / `NumberInput` / `MoneyInput`、選択は共有 `Select` を使う。ネイティブ入力（`<input type="date">`・素の `<select>`）を許容するのは**検索・一覧フィルター・表の行内編集・時刻入力**だけ（2026-10-20 に design-system §3.4 から移管。機械検査は compliance C28）。
  - 観測の再分類: `/lstep/checkup-sync`・`/lstep/delivery-monitor` のネイティブ date は検索フィルターなので許容範囲。`/accounting/close` の対象日と `/settings/closing-time` の曜日チェックは設定・業務フォームなので共有部品が必要。
- チェックボックス・ラジオ・トグルは共有 `ui/`（`checkbox` / `radio-group` / `switch`）を使い、ヒット領域は [design-states.md §1.1](design-states.md) に従う（視覚 16〜20px 可、クリック面はラベル込みで 44×44）。トグルには可視ラベルと `aria-labelledby`（観測: `/hospitalization/:id/edit` の入院タイプ 16px、`/settings/cage` の連結選択 16px、`/line-reservation/settings` で名前の無いトグル 7 個・曜日チェック 16px）。
- `Select` の表示幅は**最長の選択肢が切れない幅**にする（観測: 「以外（手動）」が 2 画面で欠ける）。
- 入力欄の高さは 44px（`h-11`）。`h-9` は design-system §3.4 の compact 制御（フィルター・行内。ラベルは body-sm）に限り、業務フォームの本文入力には使わない（観測: `/medical-records/:id` の履歴検索 36px、`/trimming/:id` の日付ボタン 38px）。
- 単位（kg・℃・円・%）は**入力欄内の右端サフィックス**に統一する。別ボックスの単位セレクトは単位が切り替わる項目だけ。

<a id="form-destructive"></a>
### 4.5 破壊操作ボタン

- 登録・編集フォームの操作はヘッダに `[削除] キャンセル 保存` の順で置き、削除は左端で主要操作から離す。固定アクションバーでも同じ順。削除対象外の画面では該当操作を省く。権限・状態（送信中・確定済み）の既存ゲートは維持し、未入力理由の `disabled` は [design-states.md §1.1](design-states.md) に従う（2026-10-20 に design-system §3.4 から移管）。
- 本書の追加規則: 削除・取消処理など不可逆操作のボタンは `Button variant="destructive"` に統一する。この variant は `button-variants.ts` の定義どおり**塗り**（`bg-destructive` + 白文字）であり、赤アウトラインの variant は存在しない（`outline` + 赤文字の自作も禁止）。赤テキストのみ（`ghost-danger`）は一覧の行内アクション専用（観測: `/examinations/:id`・`/vaccinations/:id` が赤テキスト、`/trimming/:id`・`/hospitalization/:id/edit` が自作の赤アウトライン。いずれも `destructive` へ寄せる）。
- 不可逆操作の安全性はサーバー側の検証・Undo で担保し、確認ダイアログは補助（[product-philosophy](../product-philosophy.md)）。

<a id="form-save-position"></a>
### 4.6 保存ボタンの位置

- スクロールが発生するフォームでは、保存・キャンセルを**ヘッダ右または固定アクションバー**に置き、常に到達できるようにする。フォーム末尾だけに置かない（観測: `/line-reservation/settings`・`/settings/integrations/lstep`）。
- 1 ページに複数の保存単位（カードごとに保存）がある場合は、各カードに「未保存」の状態表示を出す（観測: `/settings/closing-time`）。

---

## 5. 固定要素・重なり・トースト

<a id="fixed-elements"></a>
### 5.1 浮遊ボタン（FAB）・固定アクションバー

- 画面右下の浮遊ヘルプボタン（48×48、fixed）など、固定要素は**操作可能な要素の上に重ならない**。スクロールコンテナの末尾に固定要素の高さ + `{spacing.md}` の余白（`pb-20` 相当）を確保し、ページネーションや最下行の操作列と干渉させない。
- 固定アクションバーが本文に被さる場合も同様に、本文末尾に同じ余白を取る。
- 固定要素の `z-index` は `Z` トークンから選ぶ。`9999` の直書きを禁止する。
- 観測: `/accounting` で FAB が「最後のページ」ボタンの上半分を覆う（`elementFromPoint` が FAB を返す）。`/settings/staff`・`/settings/merchandise-items` の最下行操作、`/medical-records/:id` の固定バー + FAB で「保存」の右が混み合う。

<a id="toast"></a>
### 5.2 トースト

- トーストはヘッダの操作ボタン・入力欄の上に重ねない（右下または右上でヘッダ領域を避ける位置）。
- 閉じるボタンのヒット領域は [design-states.md §1.1](design-states.md) に従う（アイコン 16〜20px、クリック面 44×44）。
- エラートーストは「何が取れなかったか」と「次にできること」を含める（観測: `/hospitalization/:id` の「データ取得対象が見つかりません。」が担当医ボタンの上に出て、閉じるボタン 20px）。

---

## 6. 空状態・読込状態

<a id="empty-state"></a>
### 6.1 空状態の最小要件

- 空状態は **(1) 何が無いか (2) なぜ／条件 (3) 次の操作** の 3 点を含む（DESIGN.md `ex-empty-state-card`）。「○○がありません」の 1 行だけを禁止する。
- 作成できる対象なら、空状態の中に主要操作ボタン（新規登録）を置く。ヘッダ右にもあってよいが、空状態側を省略しない。
- フィルタ結果が 0 件のときは、絞り込み条件を示し「条件を変える」導線を出す（観測: `/accounting/close/history`・`/lstep/delivery-monitor` が当日 1 日のフィルタで 0 件表示のみ）。
- 観測: `/settings/hospitalization`・`insurance`・`campaigns`・`inquiry-templates`・`lstep/tags`、`/examinations/:id` 履歴、`/vaccinations/:id` 履歴、受付ボードの各列。良い例: `/accounting/close`・`/settings/lab-device-item-masters`・`/owners/new` のペット情報。
- 空のまま列ヘッダだけを出す場合、カードの高さを最小限（空状態ブロック 1 つ分）にする。

<a id="loading-state"></a>
### 6.2 読込状態

- 300ms を超える読込は**スケルトン**（表なら行のプレースホルダ、KPI ならカード枠）で示す。「読み込み中...」のプレーンテキストだけで 3 秒以上放置しない。
- 読込中もページシェル（ヘッダ・検索行・列ヘッダ）は描画し、レイアウトシフトを起こさない。
- 観測: `/aggregation` が約 10 秒間テキストのみ。

---

## 7. カレンダー・ボード・ダッシュボード

<a id="calendar"></a>
### 7.1 週・月グリッド

- 「5 日表示」を選んだら列も 5 本にする。表示日数とグリッド列数を一致させ、見えない列を横スクロールの先に置かない（観測: `/reservations` が 5 日表示で 7 列、金・土が画面外）。
- 固定列（スタッフ名など）は識別に必要な幅（氏名で 150px 以上）を確保し、§3.1 に従う。日付セルの最小幅は、その中の最短表記（「全日 09:00–18:00」）が 2 行以内に収まる幅（観測: `/shifts` が 47px セルで 3 行折り返し、スタッフ列 119px で氏名欠け）。
- 曜日・休診日の色は [design-system.md §2.5](design-system.md)（土日の青／赤禁止、曜日は文字 + ink-muted、休診日は `C.bgPage` + 「休診」。2026-10-20 決裁）。本書は再定義しない（観測: `/shifts`・`/line-reservation/slots`・`/reservations` 768px）。
- 凡例は色ドットだけで区別しない。ドット + ラベルを必ず対にする。

<a id="board"></a>
### 7.2 カンバン・カード一覧

- 受付ボードの列背景にステータス色を使うことの可否と範囲は [design-system.md §2.5 の例外表](design-system.md)（受付ボードの列面に限り許可・2026-10-20 決裁）。本書の追加規則は 1 つ: 列の区別は背景色だけに頼らず、見出しのドット + 文字を必ず併置する。
- カード内の識別子（ペット名・飼主名・ケージ名）は §3.1 に従い省略しない。カードの最小幅はそれが入る幅（ケージカード 130px は不足。観測: `/hospitalization`）。
- カードの主要操作（「詳細」）はカード全幅のボタンまたは右下に置く。幅 38px の孤立ボタンにしない（寸法は design-states §1.1）。

<a id="dashboard-density"></a>
### 7.3 単独ウィンドウ・ダッシュボード型画面

- `/owners/:id/report` のような単独ウィンドウ画面も、本体画面と同じ typography ロール（本文 15〜16px、見出しは §1.4）に従う。12〜14px を本文に使ってよいのは**印刷プレビュー**だけ。
- パネル内の表は末尾行が clip されない高さにする（観測: 「今日の来院」の最終行欠け）。横スクロールが 10,000px を超えるマトリクスは、期間フィルタか列の折りたたみを前提に設計し直す。

<a id="chart"></a>
### 7.4 グラフ（Recharts）

- 系列色・罫線・軸文字の値は [design-system.md §2.5](design-system.md) のグラフ行（`PALETTE.chart*`）。業務ステータスを集計した棒・円は当該ステータスの `STATUS_COLORS` を使い、画面のバッジと同じ色にする。それ以外の色（Tailwind パレット直指定・hex）をグラフに持ち込まない（compliance C3 の対象）。
- 系列は色だけで区別しない。凡例は **ドット + ラベル（単位付き）** を対にし、線グラフは系列ごとにマーカー（`dot`）を付ける。1 グラフの系列は 4 本まで。超えるなら系列の表示切替（`VitalsGraph` の metric toggle）を付ける。
- 軸: ラベル文字は `PALETTE.chartAxisText`、罫線は `PALETTE.chartGrid` の `3 3` 破線。単位は軸ラベルか凡例に 1 回だけ書き、tick ごとに繰り返さない。日付 tick は §10 の表示形式に従う。
- ツールチップ: hover と focus（キーボード）で出し、値は §10 の数値・日付表記で描く。ツールチップ内に色だけの凡例を置かない（系列名を書く）。
- 空・読込: データ 0 件は §6.1 の空状態（軸だけの空グラフを出さない）。読込中は同じ高さのスケルトン（§6.2）で高さを確保し、レイアウトシフトを起こさない。
- グラフは `aria-label` で「何のグラフか」を 1 文で与え、数表（テーブル）か CSV への導線を近くに置く（色覚・スクリーンリーダー対策）。
- 参照実装: `features/medical-records/components/VitalsTab/VitalsGraph.tsx`（系列色・凡例・toggle）、`features/lstep/components/LstepDeliveryStatsSection.tsx`（ステータス積み上げ棒）。

---

## 8. 本体画面以外

- `/manual/*` のような文書型画面は、アプリのサイドバーと文書用サイドバーを**同時に全幅で出さない**（本文が 700px 以下になる）。文書側を開いたらアプリ側はアイコン幅に畳む。
- 文書内の表の必須表記は、本体の「赤 `*`」と同じ語彙にする。絵文字（✅）だけで必須を表さない。
- 「画面イメージ」など画像前提の見出しは、画像が無い間は見出しごと出さない。

---

---

## 9. 機械化の候補（design-audit への追加提案）

| 検査 | 状態 | 検出方法 | 対応する規約 |
|:--|:--|:--|:--|
| C25 | **実装済み**（ラチェット） | `truncate` / `text-ellipsis` / `line-clamp-N` を持つ opening tag に `title` / `aria-label` / `aria-labelledby` が無い。focus/タッチで開くツールチップの有無は静的に判定できないため、機械化は「全文がどこにも無い」最低線のみ。既存 57 ファイルは `C25_TRUNCATE_ALLOWLIST` | §3.1 |
| C26 | **実装済み**（ラチェット） | `<nav>` 内と `Layout/Sidebar*.tsx` の `truncate`。既存 3 ファイルは allowlist | §1.2 |
| C27 | **実装済み**（ラチェット） | `z-[N]` / `zIndex: N` の直書き（`Z` / `Z_CLASS` 参照は許可）。既存 2 ファイルは allowlist | §5.1 |
| C28 | **実装済み**（ラチェット） | `src/features/**` の raw `<select>` / `<input type="date">`。§4.4 の許容文脈（フィルター等）は理由付き allowlist、既存違反 2 ファイルはラチェット allowlist | §4.4 |
| C29 | **実装済み**（0 件） | `disabled={…isValid / isFormValid / errors.length…}`。権限 `can*` と送信中は対象外。`/estimates` の `canSubmit` は RBAC 由来で検出対象外（監査で観測した disabled の原因は要再確認） | §4.2 |
| e2e | **実装済み（検知モード）**: `ui-design-compliance-readonly.spec.ts` が各ルート × 4 viewport で `nav [aria-current=page]` の viewport 内在と、`position: fixed` の操作要素と他の操作要素の非重なりを採取する（判定は `e2e/helpers/ui-design-patterns.ts`）。既定は test annotation への記録のみで、`UI_DESIGN_PATTERNS_STRICT=1` で hard fail。実装修正が揃ったら strict を既定にする | §1.1, §5.1 |

C25〜C29 の定義は [ui-design-compliance.md §1](ui-design-compliance.md)、実装は `frontend/scripts/design-system-audit.mjs`（`pnpm design-audit`。allowlist は削減方向のみ）。C21〜C24 は既存（stories / `PALETTE` 純度 / `BADGE` コントラスト / `shared/` stories）。新しい検査を足すときは同 §1 と audit スクリプトへ同一コミットで登録し、既存番号と重複させない。

---

<a id="notation"></a>
## 10. 表記（日付・時刻・数値・金額・空値）

> 値の整形は `frontend/src/lib/format/`（`date.ts` / `number.ts`）と `lib/jst-date.ts` の共有関数だけを使う。画面側で `toLocaleString` / `toFixed` / テンプレート結合を直書きしない。新しい表記が必要なら共有関数を足してから使う（二重管理の禁止）。

| 対象 | 表記 | 共有関数 | 備考 |
|:--|:--|:--|:--|
| 日付（一覧・フォーム・詳細） | `yyyy/MM/dd`（0 埋め） | `formatDate` | PO 決定 2026-07-12（`DISPLAY_DATE_FORMAT`）。wire 形式 `yyyy-MM-dd` を画面に出さない |
| 日付 + 曜日（予約・日送りナビ） | `yyyy/MM/dd (土)` | `formatDateWithWeekday` | 曜日は全角 1 文字、半角括弧 |
| 見出し用の長い日付（受付ボード・入院日次） | `2026年10月20日（火）` | `formatJapaneseDate` | 0 埋めなし・全角括弧。本文・表には使わない |
| 時刻 | `HH:mm`（24 時間・0 埋め） | `formatJSTTime` | 秒は出さない。表示は常に JST |
| 期間・範囲 | `09:00–18:00`、`10/01–10/31` | — | 区切りは EN DASH `–`（U+2013）、前後に空白なし |
| 金額 | `¥1,234`、負値は `-¥1,234` | `formatCurrency` | 3 桁区切り、小数なし、`円` を付けない。0 や未設定を空値にするかは `formatCurrencyOrDash` / `formatCurrencyIfNonzero` の契約に従う |
| 体重 | `26.5kg`、整数は `26kg` | `formatWeight` | 小数 1 桁、単位との間に空白なし |
| 体温・その他の計測値 | `38.5℃` | 画面仕様 | 単位は入力欄内の右端サフィックス（§4.4）。共有関数が無い値は `number.ts` に追加する |
| 件数 | `1,234件`、`1–20件 / 1,234件` | — | 数字と「件」の間に空白なし |
| 空値 | **`—`（U+2014 EM DASH）1 文字**、`C.textInkMuted` | — | 2026-10-20 決裁。ink-faint は 2.44:1 で不可（design-states §2.2）。共有関数は `EMPTY_VALUE`（`—`、`lib/format/number.ts`）を返す。フォーマッタ側の整合は PR #739 |
| 数字の等幅 | 表の数値列・金額列は `tabular-nums` | — | DESIGN.md Typography Principles |

- 和文の中の英数字は半角、単位記号は `kg` / `℃` / `%` / `¥` を上表の位置に固定する。
- 観測: `/examinations` の「20件」と「14,690件中」の混在、`/settings/*` の「2 件」「55 件」（空白あり）、空値 3 種類（§2.5）。
