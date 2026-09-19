# UAT-R2-CHART-FIT — カルテ表示の机上再現票

状態: **机上整理済み／実機再現・改善判定は UNKNOWN**。この票はレイアウト変更や医院受入の証拠ではない。

## 確認済みの事実と未確定の入力

| 項目 | 現時点の記録 |
| --- | --- |
| 端末・ブラウザズーム | ノート PC、ブラウザ 125%。機種・画面解像度は UNKNOWN。[^todo] |
| CSS viewport 幅・高さ | UNKNOWN。`window.innerWidth` / `window.innerHeight` の実測値はない。 |
| OS 表示スケール | UNKNOWN。ブラウザズームとは別に記録する。 |
| 対象カルテタブ | UNKNOWN。現行のタブは 9 件だが、問題の発生タブは未特定。[^tabs] |
| サイドバーの希望状態 | UNKNOWN。展開時に収まらないとの申告はあるが、医院が常時展開を必要とするか未確認。[^feedback] |
| 隠れる情報・操作 | UNKNOWN。必須情報、保存操作、フォーカスのどれがどの状態で届かないか未計測。 |
| 要件責任者（個人名） | UNKNOWN。医院・PO から責任者と業務上の目的を確認する。[^philosophy] |

現行コードではサイドバーの初期状態は `window.innerWidth < 1280` で折りたたみとなり、幅の変化も media query で反映する。展開幅は `w-[220px]`。これは CSS viewport に依存する条件であり、ノート PC・125% という記録だけでは展開状態も overflow も確定しない。[^sidebar][^width]

カルテフォームは `LAYOUT.fullHeight` を使い、問診タブのときだけタブルートに `flex-1 min-h-0` を付ける。表示領域はタブ依存で、問診の右カラムにある履歴抜粋の見え方も対象タブごとに確認する。現行構造や過去の申告から実機で「1画面に収まらない」寸法を推定しない。[^form][^interview][^feedback]

## 次回の同一端末での採取表

医院の対象端末・同じカルテ条件で、変更前と変更後を対にして記録する。まず対象タブ、サイドバーの希望状態、隠れる具体的な操作を医院に確認し、条件が決まるまでは製品レイアウトの実装を止める。[^todo]

| 条件・観察 | 変更前 | 変更後 | 判定に必要な記録 |
| --- | --- | --- | --- |
| 端末条件 | 未採取 | 未採取 | 端末、画面解像度、OS 表示スケール、ブラウザ名・バージョン、ズーム、CSS viewport 幅×高さ |
| 対象画面 | 未採取 | 未採取 | カルテタブ名、サイドバー実際の状態と医院の希望、ウィンドウ最大化状態 |
| 可視範囲 | 未採取 | 未採取 | 画面全体とスクロール位置、必要領域の上下端、縦横 overflow の有無、見えない箇所 |
| 必須情報 | 未採取 | 未採取 | 診療上必要な情報をタブごとに列挙し、閲覧・入力できるか |
| 保存操作 | 未採取 | 未採取 | 保存コントロールの可視性、スクロールを含む到達手順、操作できるか |
| キーボード操作 | 未採取 | 未採取 | Tab 移動順、フォーカス表示、隠れた領域や保存への到達・復帰 |
| 操作数 | 未採取 | 未採取 | 指定した同じ診療タスクのクリック・キー操作・スクロール回数 |

受入観察は、指定タブの必須情報を欠かさず見て入力でき、保存操作とキーボードフォーカスへ到達できることを同一端末の前後記録で示す。全画面内に収まったか、操作数が改善したかはその比較が揃うまで **UNKNOWN**。医院が展開サイドバーを必要とする場合、カルテページだけ畳む案は PO の判断に戻す。情報・タブの削除や小さすぎる文字での収まりは受入根拠にしない。[^todo][^feedback]

[^todo]: 元 main の `todo-issue.md`「UAT-R2-CHART-FIT」90–94 行。候補 worktree の[同じ節](../../../todo-issue.md#uat-r2-chart-fit)は 43–47 行。両者の端末条件・完了条件を照合した。
[^feedback]: [医院 UAT 記録の UAT-R2-CHART-FIT](../stg-uat-clinic-feedback-q1-q4.md#uat-r2-chart-fit)、402–408 行。
[^sidebar]: [Sidebar.tsx](../../../frontend/src/components/shared/Layout/Sidebar.tsx)、28–29・63–72 行。
[^width]: [design-tokens.ts](../../../frontend/src/lib/design-tokens.ts)、611–614 行。
[^form]: [MedicalRecordFormReadyPanels.tsx](../../../frontend/src/features/medical-records/routes/MedicalRecordFormReadyPanels.tsx)、182・198–212・244–245 行。
[^interview]: [MedicalRecordInterview.tsx](../../../frontend/src/features/medical-records/components/MedicalRecordInterview.tsx)、81–107 行。
[^tabs]: [medical-record-form-model.ts](../../../frontend/src/features/medical-records/routes/medical-record-form-model.ts)、4–19 行。
[^philosophy]: [product-philosophy.md](../../product-philosophy.md) の「① 要件を疑う」。
