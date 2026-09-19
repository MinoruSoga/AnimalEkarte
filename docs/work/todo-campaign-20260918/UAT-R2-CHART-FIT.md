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
