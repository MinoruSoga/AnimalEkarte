# SLACK-OWNER-HEIGHT: OwnersListTable vs OwnerSearchModal のスクロール責務

> Task migrated to Plane `EMR-187`. This file remains supporting acceptance/evidence material; use Plane for current status.

状態: **ローカル再現 DONE（1366×625 CSS px・2026-09-23、下記「再現記録」）／実機採取 未実行（端末 ID・CSS viewport UNKNOWN）**。出典は [todo-issue.md](../../../todo-issue.md) 見出し `### SLACK-OWNER-HEIGHT`（L107–110）および Slack 736–763（`1789351591.207249`、L472）。保持する現場条件:

- 最大化済みでも飼主**検索結果**をスクロールできない
- Windows 8 / Chrome、申告 1366×625、15.6インチ
- **マスタ候補の高さとは別**（[UAT-R2-MASTER-LIST-HEIGHT](../../../todo-verification.md#uat-r2-master-list-height) / `TreatmentSearchDialog` の `max-h-[calc(80vh-12rem)]` を再実装しない）

本票は「どちらが結果スクロールを持つか」をコードから絞る。ウィンドウを再度最大化する案内で閉じない。物理インチ/申告解像度を CSS viewport と同一視しない。

本ファイルは製品コードから import されない。キャンペーン unit `SLACK-OWNER-HEIGHT` の owned path および人間が読む調査票である。製品モジュールからの呼び出し行は無い（sibling `SLACK-STAFF-SELECT.md` と同じ）。既存 `docs/work/todo-campaign-20260918/` にも本 unit の票は無い。ledger `owned_paths` が本パス単体のため、他シートへの追記では unit 完了にならない。

## 医院事実（コード外・UNKNOWN）

CHART-FIT と **同じ UNKNOWN 残り**を共有する。数値を計算で捏造しない。

| 項目 | コードで分かること | 医院事実 |
| --- | --- | --- |
| 申告 1366×625 が物理画面かブラウザ表示領域か | なし | **UNKNOWN**。[CHART-FIT票](../todo-campaign-20260918/UAT-R2-CHART-FIT.md) L9–11: ローカル基準は 1366×625 CSS px。実機は `window.innerWidth` / `window.innerHeight` を別に採取 |
| 実機 CSS viewport | レイアウト計算は CSS px | **UNKNOWN**（CHART-FIT L37）。`1366/1.25` や `deviceScaleFactor` を 125% の代替にしない（CHART-FIT L59） |
| OS 表示スケール | なし | **UNKNOWN**。ブラウザズームと別に記録（CHART-FIT L11） |
| Chrome 版（Win8 最終は 109） | 現行 FE は Tailwind v4 / Vite `esnext` | **UNKNOWN**。最新 Chromium の PASS を旧 Chrome の PASS にしない（todo-issue.md L74–75） |
| 最大化状態での実測 innerHeight | Dialog `max-h-[80vh]` は **CSS viewport の 80%** | 最大化済みでも失敗、という報告は「まだ最大化しろ」では解けない。採取表の「ウィンドウ最大化状態」は観察欄であり修正手段ではない（CHART-FIT L75–76） |
| 報告画面がカルテ飼主付け替えかペット編集か飼主一覧か | 呼出元は複数 | **UNKNOWN**。下記 3 surface を混ぜて再現しない |

## 結論: 報告症状のスクロール owner は OwnerSearchModal

| Surface | 何か | 縦スクロールの実際の owner（コード契約） | 報告「飼主検索結果をスクロールできない」との一致 |
| --- | --- | --- | --- |
| **OwnerSearchModal** | カルテ/ペット編集の飼主付け替え検索。結果は最大 100 件の生 `<table>` | **意図:** 結果 `div`（`flex-1 overflow-auto min-h-[200px]`）。**実体:** Dialog が `max-h-[80vh]` のみで高さ不定のため、多数件では内側スクロールが立たず、body は Radix が scroll-lock し得る | **一致。** 検索欄・結果テーブル・選択ボタン。最大化後も 80vh は viewport 依存のまま |
| **OwnersListTable** | 飼主・ペット**一覧ページ**。サーバ 20 行/ページ + Pagination | **PageLayout 本文**（`overflow-y-auto`）。DataTable の `overflow-auto` は高さ未拘束のため縦スクロールに関与しない | **不一致。** PropertyFilter 検索はあるが OwnerSearchModal ではない。ページ全体はスクロールできる設計 |
| **TreatmentSearchDialog**（対照・本票外） | 治療プラン検索 | 一覧に **天井** `max-h-[calc(80vh-12rem)]` + `overflow-y-auto`（テスト済み） | マスタ高さ。OWNER-HEIGHT の再現対象にしない |

絞り込み根拠（報告文言 + 呼出元）:

1. 症状は「飼主**検索結果**」。OwnerSearchModal のタイトルは「飼主検索」、結果領域は検索実行後のテーブルである（[OwnerSearchModal.tsx](../../../frontend/src/components/shared/OwnerSearchModal/OwnerSearchModal.tsx) L154–247）。
2. OwnersListTable は [OwnersList.tsx](../../../frontend/src/features/owners/routes/OwnersList.tsx) L322–388 の PageLayout 子であり、OwnerSearchModal を import しない。
3. OwnerSearchModal の本番マウントは 2 箇所のみ: [MedicalRecordFormModals.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordFormModals.tsx) L92–100（カルテ編集時の飼主付け替え）、[PetEditModal.tsx](../../../frontend/src/features/owners/components/PetEditModal.tsx) L253–261（ペット編集の飼主変更）。どちらも結果クラスを上書きしない。
4. マスタ側は既に一覧天井がある（[TreatmentSearchDialog.tsx](../../../frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx) L175, L200–202）。本票でそれを飼主検索に「流用済み」としない。

## 混ぜてはいけないケース

現場の「スクロールできない」は次のどれでも同じ言葉になる。

| ID | Surface | 多数件 | 0件 | キーボード | いまのコード |
| --- | --- | --- | --- | --- | --- |
| M1 | OwnerSearchModal | `GET /v1/owners?page=1&limit=100`。`total > length` なら「先頭100件」バナー。選択ボタン `min-h-11` × 最大 100 行 | 検索前「検索してください」。0件/API 失敗は「該当する飼主が見つかりません」（失敗も空と同一表示） | 検索 Input `autoFocus`。Enter で検索。行クリックでは選択しない（テスト L119–150）。Tab は各行「選択」ボタン。結果領域に `tabIndex` も listbox もない | 結果 `div` に天井なし。`min-h-[200px]` は床 |
| L1 | OwnersListTable | 1 ページ 20 ペット行（[loaders.ts](../../../frontend/src/features/owners/loaders.ts) L79–81）。総数が多ければ Pagination | DataTable `emptyMessage`「データが見つかりません」 | ページの overflow がスクロール owner。テーブル専用キーボードリストは無い | `flex-1 min-h-0` は外側のみ。FilteringIndicator が flex 連鎖を切る |
| H1 | TreatmentSearchDialog | マスタ一覧 `max-h-[calc(80vh-12rem)]` | EmptyState | 検索/閉じる到達はマスタ受入の残作業 | OWNER-HEIGHT と別 ID |

M1 の多数件で内側スクロールが立たないとき、ダイアログは画面中央固定（`top-[50%] translate-y-[-50%]`）のまま行が枠外へ描画され、最大化しても 80vh の参照先が同じ CSS viewport なら変わらない。

## 現行経路（OwnerSearchModal = 報告対象）

1. **Dialog chrome:** 既定 [dialog.tsx](../../../frontend/src/components/ui/dialog.tsx) L60–62 は `grid`・`gap-4`・`p-6`・中央固定。`max-h` / `overflow` / `min-h-0` なし。OwnerSearchModal L154 が `sm:max-w-2xl max-h-[80vh] flex flex-col` を足す。[cn](../../../frontend/src/lib/utils.ts) は `twMerge` なので **`flex` が `grid` に勝つ**。
2. **結果領域（意図上の scroll owner）:** L190 `flex-1 overflow-auto min-h-[200px]`。DataTable は使わない。生 `<table>`（L202–236）。`min-h-0` も `max-h-[calc(80vh-…)]` も無い。
3. **高さ不定で内側 overflow が失敗し得る理由:** 親が `max-h` だけで `h-*` / `overflow-hidden` が無いと、flex 子の `min-height: auto` が内容高さになり、`flex-1 overflow-auto` が内容に合わせて伸び、スクロールバーが付かない。対照: [ExaminationImportDialog.tsx](../../../frontend/src/features/medical-records/components/ExaminationImportDialog.tsx) L116 は同じ `max-h-[80vh] flex flex-col` でも子が `flex-1 overflow-y-auto min-h-0`。TreatmentSearchDialog は子に明示天井。
4. **625 CSS px での算術（ローカル基準。実機 innerHeight は UNKNOWN）:** `80vh` = 500px。`p-6`+`gap-4`+ヘッダ+検索行を引くと残りはおおよそ 300px 前後。`min-h-[200px]` の床は 625 基準では主因になりにくい。主因は天井欠如と高さ不定。実機 innerHeight が 625 未満（ブラウザ UI・125%・OS スケール）なら床が余り高さを食う可能性は残る → 実測まで **UNKNOWN**。
5. **入れ子:** PetEditModal は親 `<Dialog>`（L169、`DialogContent` は `LAYOUT.modal.xl overflow-y-auto` = `max-h-[90vh]`）の**内側**に OwnerSearchModal をマウントする（L253–261）。OwnerSearchModal 自身が別 `Dialog.Root` を持つため **入れ子 Root**（兄弟 Root ではない）。カルテ側はフラグメント直下の独立 Dialog。どちらも結果クラスは同じ。

## 現行経路（OwnersListTable = 別 surface）

1. PageLayout L45–56: ルート `h-full` + `STYLE.page`（`overflow-hidden`）。本文 `flex-1 overflow-y-auto`。内側ラッパは `flex-1 flex flex-col` で **`min-h-0` なし**。
2. OwnersListTable L101: 外側だけ `flex-1 min-h-0`。PropertyFilter は shrink しない。FilteringIndicator は opacity のみ（flex/overflow なし）→ DataTable への高さ連鎖が切れる。
3. DataTable L82–86: `STYLE.tableContainer`（`flex flex-col flex-1 min-h-0`）+ 内側 `overflow-auto`。さらに [table.tsx](../../../frontend/src/components/ui/table.tsx) L12 が別の `overflow-auto`。どちらも**高さ未拘束**なので縦は PageLayout が owner。sticky thead はテーブル内部 overflow 基準であり、ページスクロールでは張り付かない。
4. 20×`h-16`（64px）+ header `h-11` は 625 本文残りより大きい。フィルタと Pagination は行と一緒にページスクロールで画面外へ出る。これは一覧ページの既知契約であり、モーダル検索結果の欠落スクロールとは別件。

## 1366×625 ケース表（OwnerSearchModal）

ローカル再現の viewport 指定は **1366×625 CSS px**。実機 innerWidth×innerHeight・ズーム・OS スケールは未採取。最大化の有無は観察列に書く。

| ID | 条件 | 期待（受入） | 現行（コード） | 最大化で閉じるか |
| --- | --- | --- | --- | --- |
| C0 未検索 | モーダル open、未入力/未検索 | 検索欄と閉じる（X / ESC）が画面内。本文は「検索してください」 | L243–245。スクロール不要 | いいえ。未検索は報告症状の対象外 |
| C1 0件 | ヒット 0 | 空メッセージ到達。検索欄・閉じる残置 | L238–241 EmptyState。API 500 も同じ空文（L93–97, テスト L153–160） | いいえ |
| C2 少数件 | 1〜数行（テーブルが残り高さより低い） | 検索・先頭行・末行「選択」・閉じるがスクロール無しで到達 | 内容が `min-h-[200px]` 内なら見える | いいえ。少数件成功を多数件 PASS にしない |
| C3 多数件 | limit=100 相当、または残り高さを超える行 | **結果領域だけ**が縦スクロール。ヘッダ・検索・閉じるは残る。末行「選択」と確認ダイアログへ到達 | 天井なし。内側 overflow が立たないと末行が枠外。body lock でページも送れない | **いいえ。** 報告は最大化済みでも失敗 |
| C4 打ち切り | `total > 100` | バナー可視。先頭 100 件の末行まで C3 と同じ | L197–201。ページング UI は無い | いいえ |
| C5 キーボード | Tab / Enter / ESC | 検索 Enter → 結果。Tab で可視な「選択」へ。確認後閉じる。フォーカスが枠外行に消えない | Enter 検索あり。矢印キーによるリスト移動なし。枠外行は scrollport が無いと `scrollIntoView` 先が無い | いいえ |
| C6 入れ子 | PetEditModal から開く | 親ペット Dialog と二重になっても結果 scrollport は飼主検索側 | 親は `overflow-y-auto`、子は C3 と同じ契約 | いいえ |

OwnersListTable を誤って開いた場合の分離:

| ID | 条件 | 期待 | 現行 |
| --- | --- | --- | --- |
| L0 多数ページ | 21 行以上 | Pagination でページ移動。ページ本文は PageLayout がスクロール | 20 行/ページ。フィルタが画面上端から消えるのはページ owner の帰結 |
| L1 0件 | フィルタ 0 | 空行メッセージ | DataTable empty |
| L2 キーボード | 行リンク / 操作メニュー | ページスクロールでフォーカスが見える | モーダル listbox ではない |

## CHART-FIT と共用する実機残り（UNKNOWN）

本 unit はカルテ 9 タブを直さない。採取プロトコルだけ共用する。

| 残り | 状態 |
| --- | --- |
| 実機 `innerWidth` × `innerHeight` | UNKNOWN。1366×625 申告と同一視しない |
| 100% と既報 125% の実測 CSS 領域 | UNKNOWN。割り算で作らない |
| OS スケール | UNKNOWN |
| Chrome 版 | UNKNOWN。Win8 は 109 まで、という公式事実は互換ゲート。実機未確認 |
| 最大化の有無 | 観察。修正手段にしない |
| 対象画面（カルテ sticky 飼主 / ペット編集 / 誤って一覧） | 再現時に記録。未入手なら実機セル BLOCKED |

互換性ゲートは CHART-FIT 票「Windows 8 / Chrome の互換性ゲート」に従う。寸法修正の PASS を旧 Chrome の PASS にしない。

## 完了 / PO・停止

- 検索・選択・閉じるへ到達できる **前後証拠**（C0–C5、該当なら C6）。
- 1366×625 **CSS px 基準**と **実機**を別 run。片方の成功でもう片方を閉じない。
- 物理寸法 ≠ CSS viewport。
- **単に再度最大化を案内して閉じない。**
- 正確な画面 / 旧 Chrome が未確認なら該当実機受入を残す。
- マスタ一覧高さの既存修正を再実装しない。OwnersListTable のページスクロールを「直した」ことにしてモーダルを閉じない。

## 最小提案（本 unit は実装しない）

製品コードは変更しない。実装単位が開いたときの方向だけ記す。

1. **要件:** 最大化の有無に依存せず、飼主検索結果の末行「選択」と検索欄・閉じるへ到達する。存在すべき操作を隠す `overflow:hidden` クリップや文字縮小で「収まった」ことにしない。
2. **削除:** 最大化案内を完了条件にしない。OwnersListTable と TreatmentSearchDialog をこの症状の修正対象に足さない。
3. **簡素化:** OwnerSearchModal の結果領域に、既存マスタ検索と同じく **明示天井 + `overflow-y-auto` + `min-h-0`** を足す（`max-h-[calc(80vh-…)]` または親 `overflow-hidden` + 子 `flex-1 min-h-0`）。床 `min-h-[200px]` を実機 innerHeight で再評価。API 失敗と 0 件を同じ EmptyState にしている点は本票の高さ問題とは別ギャップ。
4. サイクル短縮・自動化は手動で C3/C5 が通ってから。viewport フィクスチャ無しの見た目テストを PASS にしない。

## 再現記録（2026-09-23・ローカル基準 1366×625 CSS px）

実機採取とは別 run。Vite dev を孤立 Docker コンテナ（`ekarte-frontend:latest` + `frontend_node_modules` volume、host :3400）で起動し、スクラッチ harness（`frontend/repro-owner-height.html` + `src/repro-owner-height.tsx`。`GET /v1/owners` を stub し行数・total・入れ子を query param で制御。製品コードからは import されず、run 後に削除）で実コンポーネント `OwnerSearchModal` を実ブラウザ（Playwright Chromium、`page.setViewportSize`）に描画。`window.innerWidth=1366` / `innerHeight=625` は実測で確認済み（申告値の転用ではない）。

### 結果（現行コード = 修正済み `flex-1 min-h-0 max-h-[calc(80vh-12rem)] overflow-y-auto`）

| ID | 結果 | 実測 |
| --- | --- | --- |
| C0 | PASS | 未検索「検索してください」表示。検索欄・Close(X) が viewport 内。`body { overflow: hidden }` で Radix scroll-lock 実測 |
| C1 | PASS | 0件→「該当する飼主が見つかりません」到達。検索欄・閉じる残置 |
| C2 | PASS | 3行。結果領域 scrollHeight=clientHeight=248 でスクロール不要。末行「選択」bottom 487 ≤ 625 |
| C3 | PASS | 100行。結果 `div` が scrollport: clientHeight 308 / scrollHeight 6941。scrollTop 6633 で末行「選択」が scrollport 内（top 473–517）・elementFromPoint=選択 button。クリック→確認 Dialog「飼主変更の確認」→「変更する」→ `selected:100` で modal close |
| C4 | PASS | `total=150 > 100` で「先頭100件」バナー可視（role=status） |
| C5 | PASS | Enter で検索発火。可視外行の「選択」button を `focus()` で scrollport が自動スクロール（top 7106→473）。ESC で閉じる |
| C6 | PASS | 親 Dialog（`sm:max-w-[1000px] max-h-[90vh] overflow-y-auto` = PetEditModal と同クラス）内の入れ子 Root でも結果 scrollport は独立稼働（scrollTop 6633、末行 hit-test 可） |

証跡 PNG（worktree ルート）: `slack-owner-height-c0-unsearched-1366x625.png` / `-c1-empty-1366x625.png` / `-c2-3rows-1366x625.png` / `-c3-100rows-scrolled-last-row-1366x625.png` / `-c3-confirm-dialog-1366x625.png` / `-c6-nested-last-row-1366x625.png` / `-legacy-broken-innerH380.png`

### 修正前クラス（`flex-1 overflow-auto min-h-[200px]`）との対比 — 機序の訂正

DOM の class 差し替えで legacy セットを同一 fixture に再現して計測した。**L55–59 の仮説は 1366×625 では再現しない。** `overflow:auto` の要素は scroll container であり block 方向の automatic minimum size は 0 になるため、definite な親高さ（`max-h-[80vh]` が効いた 500px）に対して flex 子は正しく縮む。実測（legacy @625）: clientHeight 322 / scrollHeight 6977 / `scrollTop=5000` 可・末行 hit-test 可 —— 申告 viewport では legacy でもスクロールは成立する。

実際に legacy が破綻するのは **床 `min-h-[200px]` が dialog 内容収容高を超えた時**:

| innerHeight (CSS px) | legacy 結果領域 | 状態 |
| --- | --- | --- |
| 625 | 322px、scrollport 成立 | PASS（報告値そのままでは再現しない） |
| 450 | 200px（床）が dialog 360px に辛うじて内蔵（下端差 -7px） | PASS 境界 |
| 380 | 200px が dialog 下端を **+49px**、viewport を **+11px** 越え。最大 scroll で末行 button bottom 378 vs dialog 下端 342 | 末行が枠外・下端切れ（PNG 証跡あり） |
| 340 | 最大 scroll でも末行「選択」 bottom 374 > viewport 340 → **off-viewport・物理的に到達不能**。fixed 版は同条件で results 80px・末行 top 217–261 で到達可 | FAIL=症状一致 |

発火条件はおよそ `0.8 × innerHeight < ヘッダ+検索+padding(実測約153px) + 200px` → **innerHeight ≲ 440 CSS px**。申告「1366×625」が画面解像度で実機 innerHeight がブラウザ UI・OS 表示スケール・ズームで ≲440 まで下がっていたなら「最大化してもスクロールできない」と整合する（最大化は CSS px の innerHeight を変えない）。実機 innerWidth×innerHeight・ズーム・OS スケール・Chrome 版は引き続き **UNKNOWN**（実機受入は別 run で残す）。

### 結論の確認

- 報告症状のスクロール owner は **OwnerSearchModal の結果 `div`**（`[data-testid=owner-search-results]`）で確定。呼出元は MedicalRecordFormModals / PetEditModal の 2 箇所のみ（grep 確認済み）、OwnersListTable は import せず別 surface。
- 現行コードは 1366×625 および 340・380 の低下 innerHeight でも全ケース到達可。legacy の破綻は床 `min-h-[200px]` 起因であり、現行の `min-h-0` + 明示天井が除去している。
- 実機（Win8/Chrome ≤109、実 innerHeight、ズーム/スケール）の受入は本 run では閉じない。

## 参照

- 症状正本: [todo-issue.md#slack-owner-height](../../../todo-issue.md#slack-owner-height)
- 共用 viewport: [UAT-R2-CHART-FIT.md](../todo-campaign-20260918/UAT-R2-CHART-FIT.md)
- マスタ高さ（別 ID）: [todo-verification.md#uat-r2-master-list-height](../../../todo-verification.md#uat-r2-master-list-height)
- 飼主付け替え仕様: [common-dialogs.md §2.2](../../../docs/spec/screens/common-dialogs.md)
