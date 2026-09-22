# UAT-R2-CHART-FIT — 1366×625のカルテUI修正・検証票

状態: **寸法再現・UI修正設計 READY／実装・実機受入未実施**。2026-09-19依頼者回答を反映。従来の「解像度・対象タブの回答待ち」を解除し、下記の基準で着手する。状態の正本は [TODO](../../../todo-issue.md#uat-r2-chart-fit)。

## 確認済みの事実と未確定の入力

| 項目 | 現時点の記録 |
| --- | --- |
| 端末・ブラウザズーム | 依頼者指定はWindows 8、Chrome、15.6インチ、1366×625。既報のブラウザ125%も検証対象に残す。 |
| CSS viewport 幅・高さ | ローカル基準は1366×625 CSS pxとする。申告寸法が物理画面/ブラウザ領域のどちらかと実機CSS値は未採取。担当QAが `window.innerWidth` / `window.innerHeight` を採取し、基準値と別に記録する。 |
| OS 表示スケール | UNKNOWN。ブラウザズームとは別に記録する。 |
| 対象カルテタブ | 全9タブ。問診、診察/治療プラン、治療、予防接種、定期健診、検査、画像、見積書、会計(医師確認)。[^tabs] |
| サイドバー | 展開/折畳みの両方を検証。展開時にも操作可能とし、自動折畳みだけで合格にしない。常時展開を強制する新仕様にはしない。[^feedback] |
| 隠れる情報・操作 | 開発/QAが各タブの表示、保存、エラー、選択ダイアログ、フォーカスを測る作業。依頼者へ追加の特定を求めない。 |
| 要件責任者（個人名） | 要件出所は9月19日の依頼者回答。個人名と受入担当の参照は製品仕様変更/受入の実行票へ記録する。[^philosophy] |

現行コードではサイドバーの初期状態は `window.innerWidth < 1280` で折りたたみとなり、幅の変化も media query で反映する。展開幅は `w-[220px]`。これは CSS viewport に依存する条件であり、ノート PC・125% という記録だけでは展開状態も overflow も確定しない。[^sidebar][^width]

カルテフォームは `LAYOUT.fullHeight` を使い、問診タブのときだけタブルートに `flex-1 min-h-0` を付ける。表示領域はタブ依存で、問診の右カラムにある履歴抜粋の見え方も対象タブごとに確認する。現行構造や過去の申告から実機で「1画面に収まらない」寸法を推定しない。[^form][^interview][^feedback]

## 9タブ overflow 原因と最小UI案（基準 1366×625）

対象タブ名の正本は `MEDICAL_RECORD_TABS`（問診 / 診察/治療プラン / 治療 / 予防接種 / 定期健診 / 検査 / 画像 / 見積書 / 会計(医師確認)）。[^tabs]

**共通シェル（全タブ）**

- フォーム根は `LAYOUT.fullHeight`（`flex-1 min-h-0 flex flex-col h-full`）。[^form][^fullheight]
- `PageLayout` は `FormHeader` をスクロール外に固定し、本文だけ `overflow-y-auto`。本文に `px-3 py-6`。[^pagelayout]
- `UnifiedTabsRoot` へ `flex-1 min-h-0` を付けるのは **問診のときだけ**。他タブは高さチェーンが切れ、本文が `min-h-[500px]` や固定ヘッダと積み上がると PageLayout のページ縦スクロールに逃げる。[^form]
- 患者識別・来院種別・診察日・タブ9個は sticky。タブ列は `overflow-x-auto` でラベルを畳まない。同居ペットも横スクロール。[^sticky]
- サイドバー初期折畳みは `window.innerWidth < 1280`、以降は `(max-width: 1279px)`。展開幅 `w-[220px]`、折畳み `w-[56px]`。ローカル基準 1366 CSS px では初期展開。実機の `innerWidth` は未採取のため **UNKNOWN**（計算で捏造しない）。合格条件に常時折畳みを使わない。[^sidebar][^width]
- 保存/印刷/確定は `fixed bottom-6 right-6`。会計タブではこの塊を出さず、タブ内の「チェック完了」が同じ固定位置。[^actions]
- 処置検索ダイアログは `max-h-[80vh]`、一覧 `overflow-y-auto max-h-[calc(80vh-12rem)]`。625 高では 80vh でもヘッダ+検索を残し一覧だけ局所スクロールする。[^searchdlg]

**禁止（全タブ共通の最小案）:** 臨床情報や操作を `overflow:hidden` で切って隠す、文字サイズの一律縮小、サイドバー強制折畳みだけで合格、タブ削除。Chrome 109 / Tailwind v4 / `esnext` は下節の **別ゲート** のまま。寸法修正の PASS を旧Chromeの PASS にしない。

実機 CSS viewport（`innerWidth`×`innerHeight`）、OS 表示スケール、Chrome 版は未採取 → **UNKNOWN**。

| タブ | 現行レイアウトに基づく overflow 原因 | 最小UI案（1366×625、情報非隠蔽） |
| --- | --- | --- |
| 問診 | タブ本体だけ `flex-1 min-h-0`。中身は `lg:grid-cols-12` の3列（3/4/5）`h-full`。Tailwind `lg` は 1024px。1366−220 でも3列のまま。右列「問診抜粋」はカード `overflow-hidden` + 内部 `ScrollArea`。履歴ヘッダに `w-48` 検索。高さ 625 から FormHeader・sticky・`py-6` を引くと3列の min-h-0 チェーンが崩れると抜粋が hidden で切れる。[^interview][^history] | 3列は維持してよい（幅は足りる）。各列を `min-h-0` + **明示 overflow-y-auto / ScrollArea** にし、カードの `overflow-hidden` はクリップ用途に使わない。検索は見出し下へ折返し。sticky の患者/タブと固定保存は残す。実測 innerWidth は UNKNOWN。 |
| 診察/治療プラン | `DiagnosisHeader` が `grid-cols-12` **固定高 `h-[300px]` shrink-0**。その下に ClinicalPlan、治療プラン（`overflow-hidden` の flex カード）、合計、さらにタブ外の次回予定・推奨理由。問診以外はタブルートに min-h-0 が無く、300px ヘッダだけで本文残りを食う。12列は 1366−220 でも横に詰まる。[^diagnosis][^plan][^clinical] | ヘッダの `h-[300px]` をやめ、列を `minmax(0,1fr)` + 列内スクロール。狭いときは 12列を 1〜2段へ wrap（主訴 / 身体所見 / 診断）。治療プラン表は `overflow-y-auto` の局所領域（hidden で行を切らない）。次回予定は同じ本文スクロール末尾。文字縮小・列削除なし。 |
| 治療 | `TreatmentsTab` は `flex flex-col gap-3 pb-24` + 表 `overflow-x-auto`。列は種別・内容・保険・単価・数量・値引き・小計・メモ・操作で、展開サイドバー時は横 overflow。高さは PageLayout スクロール任せで、末行・追加コントロールが固定保存に隠れるため `pb-24`。[^treatments][^treatheads] | 表は横スクロールを局所に残し、内容/メモを wrap。ページ全体ではなく表ボディを縦スクロール。追加行と合計を表の外の到達可能な帯に置く。保存は既存 fixed を維持。hidden でセル切捨て禁止。 |
| 予防接種 | フォームが `h-[calc(100vh-220px)] min-h-[500px] overflow-y-auto pb-20 lg:grid-cols-5`。625 では `100vh-220=405` だが **`min-h-[500px]` が勝ち**、ヘッダ類と足すとルートが 625 を超える。`lg` 5列は 1146px 幅で1列あたり狭い。外側保存は出さず内側「接種記録追加」。[^vaccine][^actions] | `min-h-[500px]` を外し、残り高さへ `flex-1 min-h-0`。5列は 2列 wrap（履歴 / 入力）。局所 `overflow-y-auto`。内側追加操作と患者/タブを画面内に残す。強制サイドバー折畳みは使わない。 |
| 定期健診 | `flex flex-col gap-3 pb-24`。表は `overflow-x-auto`、日付/種別/次回/担当/結果/操作に `min-w-[10rem]` 等。追加・編集行と動的結果フィールドで縦に伸び、PageLayout スクロール + 固定保存の下に操作が潜る。[^checkups][^checkuptable] | フィルタ/追加フォームを表の上で wrap。表は横スクロール局所、ボディは縦スクロール。結果フィールドは行展開ではなく下の局所パネル。`pb-24` は固定保存との重なり回避のみ。情報削除なし。 |
| 検査 | 予防接種と同じ `h-[calc(100vh-220px)] min-h-[500px] overflow-y-auto pb-20`。フィルタ（検索・期間・取込）+ グループ一覧。`min-h-[500px]` が 625 シェルと競合。[^exam] | `min-h-[500px]` 削除。フィルタは wrap して sticky 直下に残す。一覧だけ局所縦スクロール。取込ダイアログは既存 max-h パターン。root の overflow:hidden 禁止。 |
| 画像 | 検査と同じ高さクラス + フィルタ（検索・期間・ソート・アップロード）+ グループ。サムネグリッドが短い高さでページ全体を押し下げ、`min-h-[500px]` が 625 を超える。[^image] | 同上。ギャラリーを `min-h-0 overflow-y-auto` の局所領域。アップロード操作はフィルタ帯に残す。画像を hidden でトリムして「収まった」ことにしない。 |
| 見積書 | 同じ `min-h-[500px]` 高さ。科目 + 明細表 + 合計 + **コメント/備考 `grid-cols-2`** + `pb-24`。625 では明細とコメントが固定保存と重なり、2列コメントが横に狭い。[^estimate] | `min-h-[500px]` 削除。明細ボディ局所スクロール。コメント/備考は 1列 wrap。保存は見積タブではバイタルを出さない既存挙動を維持し、保存ボタンは fixed のまま到達可能。 |
| 会計(医師確認) | `h-full` + 表カード `overflow-hidden` + 検査/予防接種の ExtraLines + 合計。タブルートに min-h-0 が無いので hidden が明細を切り得る。確認ボタンは `fixed bottom-6 right-6`（このタブだけ FloatingActions が null）。[^bill][^actions] | カードを `overflow-y-auto` に変更（hidden クリップ禁止）。ExtraLines と合計は表の下の到達可能な帯、または同一局所スクロール末尾。チェック完了は画面内に残す。サイドバー展開でも操作可能。 |

実装時の検証は既存「着手手順」の 1366×625 × sidebar 展開/折畳み。この表は設計入力であり、実機採取セルは未実施のまま UNKNOWN。Windows 8 / Chrome 109 の判定は次節。

## 着手手順と合格条件

1. [カルテ配置](../../../frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx)・各タブ・[検索ダイアログ](../../../frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx) の固定幅/高さ、flex/gridの最小サイズ、スクロール責務を特定。再現する表示/操作テストを先に作り、小さいUI差分で直す。
2. 基準1366×625 CSS pxで全9タブ×sidebar展開/折畳みを確認する。短文/長文、空/長一覧、入力エラー、選択ダイアログ/保存確認を含める。比較用の大きいviewportでも回帰確認する。15.6インチからCSS値を計算しない。
3. rootの不要な横スクロール/切取りをなくす。患者識別、タブ切替、主要保存操作を画面内に保ち、本文・長い一覧は明示した領域の縦スクロールで末尾まで到達可能にする。「見切れない」は情報/操作が到達不能でないこととし、全履歴をスクロールなし1画面に押し込まない。列の折返し/上下配置・余白を調整し、情報削除、文字の一律縮小、`overflow:hidden` による隠蔽を使わない。
4. キーボードで全入力・保存・エラーへ到達し、フォーカスが固定ヘッダ/フッタに隠れないことを確認。ダイアログの検索/選択/閉じると元の位置への復帰を確認する。マスタ一覧高さの既存修正は再実装せず [受入タスク](../../../todo-verification.md#uat-r2-master-list-height) と同じ証拠を使う。
5. 同一実機で100%/既報125%を確認し、その時のCSS viewport、OS倍率、Chrome版をQAが採取する。`1366/1.25` の計算や `deviceScaleFactor` だけを125%の実機再現としない。縮小された実測領域でも同じ到達条件を満たす。環境未確保は実機ケースだけBLOCKEDとし、基準寸法の調査・設計を止めない。

## Windows 8 / Chrome の互換性ゲート

2026-09-19の公式資料照合では、Windows 8/8.1対応Chromeは109が最終で更新サポート終了。一方このrepoの [package.json](../../../frontend/package.json) はTailwind v4、[Vite設定](../../../frontend/vite.config.ts) は `build.target: esnext`。Tailwind v4の基準はChrome111以降であり、**現行構成は指定端末を互換性保証できる状態ではない**。[Google公式要件](https://support.google.com/chrome/a/answer/7100626?hl=ja)・[Tailwind公式互換性](https://tailwindcss.com/docs/compatibility) を根拠とする。

- 次作業: 合成データだけの隔離検証先で実機Chrome版を確認し、JS起動・CSS表示・入力/保存・ダイアログの不具合を寸法問題と分けて記録。最新Chromiumのviewport試験は旧Chromeの代替証拠ではない。
- 旧Chrome固有の不足があれば、限定互換対応の実現性/影響とサポート対象端末への更新案を依頼者へ提示する。Windows 8要件を黙って削除せず、基盤downgrade・依存追加・OS変更は別の採用判断にする。ブラウザの保護機能を無効化して通さない。
- 旧環境受入はこの判断と対象版の実機検証まで未完了。寸法のUI修正を終えても、本ID全体を完了にはしない。

## 変更前後の採取表

基準環境と実機を別runにし、同じ合成カルテ・タブ・sidebar状態の変更前後を対で保存する。実患者情報は使わない。

| 条件・観察 | 変更前 | 変更後 | 判定に必要な記録 |
| --- | --- | --- | --- |
| 端末条件 | 未採取 | 未採取 | 端末、画面解像度、OS 表示スケール、ブラウザ名・バージョン、ズーム、CSS viewport 幅×高さ |
| 対象画面 | 未採取 | 未採取 | カルテタブ名、サイドバー実際の状態と医院の希望、ウィンドウ最大化状態 |
| 可視範囲 | 未採取 | 未採取 | 画面全体とスクロール位置、必要領域の上下端、縦横 overflow の有無、見えない箇所 |
| 必須情報 | 未採取 | 未採取 | 診療上必要な情報をタブごとに列挙し、閲覧・入力できるか |
| 保存操作 | 未採取 | 未採取 | 保存コントロールの可視性、スクロールを含む到達手順、操作できるか |
| キーボード操作 | 未採取 | 未採取 | Tab 移動順、フォーカス表示、隠れた領域や保存への到達・復帰 |
| 操作数 | 未採取 | 未採取 | 指定した同じ診療タスクのクリック・キー操作・スクロール回数 |

完了条件は全9タブの基準寸法/実機runで上記の到達条件を満たし、必要な互換性判断と実機証拠が揃うこと。操作数の前後比較も残す。未実行・互換性未確認をPASSにせず、寸法修正と旧環境受入を別欄にする。

[^feedback]: [医院 UAT 記録の UAT-R2-CHART-FIT](../stg-uat-clinic-feedback-q1-q4.md#uat-r2-chart-fit)、402–408 行。
[^sidebar]: [Sidebar.tsx](../../../frontend/src/components/shared/Layout/Sidebar.tsx)、28–29・63–72 行。
[^width]: [design-tokens.ts](../../../frontend/src/lib/design-tokens.ts)、611–614 行。
[^form]: [MedicalRecordFormReadyPanels.tsx](../../../frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx)、182・198–212・244–245 行。
[^interview]: [MedicalRecordInterview.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordInterview.tsx)、81–107 行。
[^tabs]: [medical-record-form-model.ts](../../../frontend/src/features/medical-records/routes/medical-record-form-model.ts)、4–19 行。
[^philosophy]: [product-philosophy.md](../../product-philosophy.md) の「① 要件を疑う」。
[^fullheight]: [design-tokens.ts](../../../frontend/src/lib/design-tokens.ts)、654–660 行。
[^pagelayout]: [PageLayout.tsx](../../../frontend/src/components/shared/PageLayout/PageLayout.tsx)、45–56 行。
[^sticky]: [MedicalRecordStickyHeader.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordStickyHeader.tsx)、44–53・193–216 行。
[^actions]: [MedicalRecordFormActions.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordFormActions.tsx)、53–56・116–119 行。会計タブの固定操作は [MedicalRecordBillCheck.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordBillCheck.tsx)、335–360 行。
[^searchdlg]: [TreatmentSearchDialog.tsx](../../../frontend/src/components/shared/TreatmentSearchDialog/TreatmentSearchDialog.tsx)、175・202 行。
[^history]: [InterviewHistory.tsx](../../../frontend/src/features/medical-records/components/InterviewHistory.tsx)、43–74 行。
[^diagnosis]: [DiagnosisHeader.tsx](../../../frontend/src/features/medical-records/components/DiagnosisHeader.tsx)、55 行。
[^plan]: [MedicalRecordDiagnosisPlan.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordDiagnosisPlan.tsx)、186・229–233 行。
[^clinical]: [MedicalRecordClinicalTabs.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordClinicalTabs.tsx)、93–116 行。
[^treatments]: [TreatmentsTab.tsx](../../../frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTab.tsx)、48–51 行。
[^treatheads]: [TreatmentsTabParts.tsx](../../../frontend/src/features/medical-records/components/TreatmentsTab/TreatmentsTabParts.tsx)、14–23 行。
[^vaccine]: [MedicalRecordVaccination.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordVaccination.tsx)、96–98 行。
[^checkups]: [CheckupsTab.tsx](../../../frontend/src/features/medical-records/components/CheckupsTab/CheckupsTab.tsx)、157–163 行。
[^checkuptable]: [CheckupsTabTable.tsx](../../../frontend/src/features/medical-records/components/CheckupsTab/CheckupsTabTable.tsx)、16–25・85 行。
[^exam]: [MedicalRecordExamination.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordExamination.tsx)、68 行。
[^image]: [MedicalRecordImage.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordImage.tsx)、82 行。
[^estimate]: [MedicalRecordEstimate.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordEstimate.tsx)、222・250 行。
[^bill]: [MedicalRecordBillCheck.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordBillCheck.tsx)、286–336 行。
