# 一覧・会計明細の優先改善計画（EMR-379 / EMR-384）

作成日: 2026-10-13
状態: 実装前の計画。実行状態・指摘・検証証跡の正本はPlane。
依頼: visual-qa / improve-ui の結果を踏まえ、改善計画を追加する。
責任者: ユーザー依頼を受けたCodex Lead。Devin CLIは読み取り専用で根拠を確認する。

## 対象と順序

最初の実装対象は **[EMR-379](https://app.plane.so/baritechllc/browse/EMR-379/)（カルテ一覧）と [EMR-384](https://app.plane.so/baritechllc/browse/EMR-384/)（会計明細）**。どちらも、識別情報や金額項目が1〜2文字ずつ折り返され、一覧を読む際の視線移動と誤読の可能性を増やしている。新しい機能や操作を追加せず、既存情報を読める表示に整える。

ソース確認の基準: `965b9473b0b1c2144dbd3c20c93369bb3fa33752`。監査画像の申告基準は `ed04c0237f81ddc2890d39b528d3608a58180c8b`。画像撮影中にmainが更新されたため、画像全件が当該SHAと一致する保証はない。実装着手時は対象ファイルの更新とclaimを再確認し、固定した候補で再現画像を取得する。

## 共通方針

- 表示する列・値・業務上の意味を保持し、列の幅・折り返し・スクロールを修正する。
- 色・余白・文字サイズは既存の `DESIGN.md`、`docs/spec/design-system.md`、`docs/spec/design-states.md` と実装トークンに従う。
- 共有 `Table` / `DataTable` 全体への幅・nowrapの強制は行わない。画面ごとの情報量が異なり、既存の狭幅対応（BUG-458）へ波及するため。
- 既存部品の再利用を優先する。共有部品に表示制御を足す場合は、既定値を保持した明示的なopt-inとし、対象画面だけで使う。
- CSSの幅指定だけのテストでは見た目の改善を保証できない。候補を固定した実ブラウザの確認を受入条件に含める。

## EMR-379 — カルテ一覧の識別情報を読める幅にする

### 根拠と変更対象

| 対象ソース | 現状と役割 |
|---|---|
| `frontend/src/features/medical-records/hooks/use-medical-records-columns.tsx:52` | 診療日等に幅指定がある一方、飼主名・ペット名の列には幅指定がない |
| `frontend/src/features/medical-records/routes/MedicalRecordsListPanels.tsx:96` | 本文の飼主名・ペット名セル。ペット名リンクと死亡バッジを併置する |
| `frontend/src/components/shared/DataTable/DataTable.tsx:79` | `min-w-0` とスクロールコンテナ。固定 `min-w-[640px]` を避ける既存判断がある |
| `frontend/src/components/shared/DataTable/list-table-col.ts` | ステータスだけは最小幅とnowrapを既に共有化済み |

画像: `shot-20261012T184308602071.png`（氏名の縦積みをLead確認）、`cdp-m-medrec-768.png` / `cdp-m-medrec-390.png`（狭幅の追加確認用）。保存先は監査記録を参照。画像はPlane未添付。

### 実装方針

1. カルテ一覧専用の列定義と行セルに、識別情報用の最小幅・折り返し規則を同じ定義から適用する。ヘッダだけを直して本文を取り残さない。
2. ヘッダはnowrap、本文の名前は最小幅を確保した上で折り返し可能とする。通常長の飼主名・ペット名を1文字ずつ折り返さず、長い名前は全文を読める幅で折り返す。死亡バッジは縮めず、ペット名リンクの操作対象を維持する。全文を確認できない省略へ置き換えない。
3. まず列ごとの最小幅で既存スクロールを成立させる。実ブラウザで安定しない場合に限り、`DataTable` に `<Table>` へ追加する省略可能な `tableClassName` を足す。既存の `className` は外側コンテナ用のまま保持し、テーブルの既定クラス `w-full min-w-0` を変更しない。カルテ一覧だけで、ブレークポイントごとの**表示列**の最小幅を反映する。非表示列まで幅へ算入しない。
4. 現行の `md` / `lg` による補助列の表示条件を維持する。日付・識別情報・ステータス・操作の列を、スクロールしても到達できない位置へクリップしない。
5. 検索placeholderの途中切断は、単独で不具合と断定しない。`PropertyFilter.tsx:179` は全文を `aria-label` に設定済み。今回は検索ラベルを追加せず、実画面で視覚的説明が足りないと確認された場合にだけ別途判断する。

最小幅の最終値は初回の再現計測で確定する。文字の表示幅・セルpadding・死亡バッジを含む必要幅を測り、画面固有の定義へ置く。単なる「列幅配分が原因」の断定は避け、コンテナの利用可能幅とtableの実測を記録する。

### 受入条件

- 1568×873 / 1280×800 / 768×1024 / 390×844 CSS pxで、通常長の日本語氏名・ペット名が1文字ずつ縦積みにならない。
- 長い氏名・ペット名、死亡ペット、他医院の読み取り専用行、複数医院表示でも識別情報・死亡表示を読み取れる。
- ステータス・許可された行操作へマウス・タッチ・キーボードで到達できる。リンク・ソート・ページ送り・絞り込みの挙動を保持する。
- テーブルが横幅を超える場合、テーブル領域でスクロールでき、`documentElement.scrollWidth` がviewportを超えるページ全体の横スクロールを新たに起こさない。
- 権限のない編集・削除や他医院の操作が表示上の変更によって有効化されない。
- 共有部品にopt-inを足す場合、省略した既存利用側とBUG-458の回帰テストを維持する。

### 実装時の限定検証

候補worktreeをマウントしたDocker環境で、既存の `medical-records-columns.test.tsx`、`MedicalRecords.test.tsx`、`MedicalRecords.species-status.test.tsx` を中心に確認する。共有部品を変更する場合は `DataTable.test.tsx` も含める。

```bash
docker compose exec frontend npx vitest run src/features/medical-records/routes/medical-records-columns.test.tsx src/features/medical-records/routes/MedicalRecords.test.tsx src/features/medical-records/routes/MedicalRecords.species-status.test.tsx src/components/shared/DataTable/DataTable.test.tsx
```

DOMテストは識別情報・バッジ・権限・ソート等の回帰を扱い、折り返し・実幅・スクロールはブラウザで確認する。ブラウザ操作は既存ルーティング規則を用いる。既存E2Eの参照先は `frontend/e2e/medical-records-pagination-sort.spec.ts` と `medical-records-patient-search.spec.ts`。必要なUI操作は当該テストとの役割を照合してから選ぶ。

## EMR-384 — 会計明細ヘッダと編集値を読み取れる状態にする

### 根拠と変更対象

| 対象ソース | 現状と役割 |
|---|---|
| `frontend/src/features/accounting/components/ItemListCard.tsx:461` | 明細テーブルの11列。既存の指定幅の合計は980px（100+180+90+60+90+100+70+80+60+100+50）。多くは幅指定のみで、ヘッダnowrapは税額列のみ |
| `frontend/src/features/accounting/components/AccountingItemRow.tsx:156` | 品目名セルは現在nowrap。品目名入力w-56、数量w-20等の編集UIがヘッダ指定幅より大きくなる |
| `frontend/src/features/accounting/components/AccountingDetailPanels.tsx:277` | desktopで左明細と右400pxの決済情報を並べる。左領域の幅・overflow条件を確認する |
| `frontend/src/components/ui/table.tsx:12` | 横スクロール用wrapperを既に持つ。ここへ全テーブル共通の最小幅を設定しない |

画像: `shot-20261012T190856215454.png`（ヘッダ縦積みをLead確認）、`shot-20261012T190911419481.png`（編集状態の追加確認用）。画像はPlane未添付。

### 実装方針

1. 明細の全ラベル付きヘッダをnowrapにする。ヘッダだけを広げず、編集入力やSelectを持つ本文列にも必要な最小幅を適用する。
2. 対象テーブルに画面固有の最小幅を設定する。**980pxは現行ヘッダ指定幅の合計から得た下限の初期値**。各列の最小幅は「ヘッダの必要幅」と「最も幅を使う編集状態の入力/Select/ボタン/バッジ + セルpadding」の大きい方を採り、表全体の最小幅はその合計で決める。品目名入力224px、数量入力80px、金額入力96px等はpadding分が別に必要。候補上の計測値・採用値をチケットに記録する。
3. 既存 `Table` wrapperを横スクロールの主体とする。親のflex childとCardへ必要な `min-w-0` 制約を与え、ページ全体のoverflowや祖先の `overflow-hidden` による右端クリップを防ぐ。
4. 多重スクロール領域とstickyヘッダの組み合わせを実測する。横スクロールは表内へ収め、縦方向のヘッダ追従と最終行・合計表示の可視性を保持する。
5. 品目名セル全体のnowrapを外し、幅を制限した品目名テキストだけを `whitespace-normal` で折り返す新しい表示へ変更する。編集入力・金額・バッジのnowrapと最小幅は維持し、長い品目名も全文を読める。横スクロール中も数量・課税区分・税率・割引・削除操作を使えるようにし、Select/Dialog等のoverlayを切らない。セルのpaddingや文字サイズは変えない。

計算・税区分・保険・金額丸め・会計確定・監査ログ・印刷仕様は本票の変更範囲に含めない。

### 受入条件

- 登録・編集の両方で、同じ4種類のviewportにおけるヘッダが1〜2文字ずつ縦積みにならず、一行で読める。
- 表を横スクロールすれば全11列と右端操作へ到達できる。ページ全体の横overflowを起こさない。
- 空明細、長い品目名、複数明細、当日/過去の未請求区分、保険あり/なし、編集権限なしを確認する。
- 数量・割引・課税区分・税率の編集と各overlayを操作でき、表示値・再計算結果は変更前と一致する。
- ヘッダの縦スクロール追従、最終行、税抜小計・消費税・合計、決済パネルと確定ボタンを確認できる。
- 別画面の `Table` や印刷帳票へ幅・折り返し変更が波及しない。

### 実装時の限定検証

```bash
docker compose exec frontend npx vitest run src/features/accounting/components/ItemListCard.test.tsx src/features/accounting/components/AccountingItemRow.test.tsx src/features/accounting/routes/AccountingDetail.test.tsx
```

既存の品目追加・編集・計算・権限テストを保持する。JSDOM上のクラス確認でレイアウトPASSとしない。ブラウザでは表のscrollWidth/clientWidth、右端列への到達、入力/Selectの操作、stickyヘッダを確認する。既存フローの参照先は `frontend/e2e/accounting-smoke.spec.ts` と `accounting-flow.spec.ts`。

## 着手・検証・受入の手順

1. 実装着手時にPlane、claim、対象HEADと関連ファイルの既存変更を再確認する。優先2票は別worktreeで扱い、共有部品へ同時に書き込まない。
2. 隔離環境で監査時の状態を再現し、before画像と計測を保存する。再現しなければ追加の変更を先に入れず、候補SHA・CSS・データ・viewportの違いを調べる。
3. 必要な回帰テストを先に用意し、列・コンテナの限定修正を実装する。既存のデザイントークンとnearestルールに従う。
4. 上記のscoped Dockerチェックと、変更ファイルのlint/format、適用されるデザイン監査を実施する。mainの常駐コンテナを候補の検証に流用しない。
5. 指定viewport・状態でafter画像と同じ計測を取得し、コードレビューと実画面確認を各票に記録する。画像の共有は承認された非公開手段を使い、公開URLへのアップロードで代用しない。
6. 実装レビュー・main統合・STG/UAT・人間受入は別々に記録する。この計画の追加だけでチケットをReady/Running/Doneへ進めない。

## 後続の改善順序

| 順序 | 対象 | 次に確定すること |
|---|---|---|
| 1 | EMR-379 / EMR-384 | 本書の優先2計画 |
| 2 | EMR-378 / EMR-382 | 名前・ステータス・日付の折り返し。優先2票の修正結果を踏まえ、共通原因か個別原因かを判断 |
| 3 | EMR-380 / EMR-387 | カルテ右パネル・担当医の見切れと、FABの操作要素への重なり。実際の操作阻害と影響面を確定 |
| 4 | EMR-377 / EMR-381 / EMR-383 | 受付列・通知・取得失敗状態。既存配置仕様と再現操作を照合してから修正案を一つに絞る |
| 5 | EMR-385 / EMR-386 / EMR-388 | フィルタ文言・新規パネル名・チップ省略。意図的な表示との区別と対象範囲を確認 |
| 別枠 | EMR-372のIU所見 | IU-002/004の誤認・重複とIU-007の描画経路を修正。投与量の安全注意（IU-005）等は正本との対応を確定後に計画化 |

全対象監査の残り（権限で未到達の画面、データ0件の編集面、LINE予約フロー、LIFF token状態、行メニュー等）はEMR-373で追跡する。今回の優先2計画と、全対象監査の完了を同じ判定にしない。

## この計画で実施した確認

Codex Leadによる対象ソース・規約・既存テストの読み取りと、前回確認した監査画像の所見を根拠とする。製品コードの変更、Dockerテスト、ブラウザ再撮影、計算の実測、UATはこの計画作成では実施していない。

Devin CLI（要求UID `swe-2-high`、exportで実モデル名 `SWE-2 High` を確認）の読み取り専用確認で、幅合計の訂正、品目名のnowrap解除対象、ヘッダ/本文の両方への適用、`tableClassName` の既定値、検索の既存aria-labelを再確認し、計画へ反映した。同じセッション `slender-count` で修正後の計画文書を再レビューし、仕様適合・品質ともPASS。`superpowers:verification-before-completion` のinvokeと実Sourceをexportで確認した。これは計画文書のレビュー結果であり、製品修正やブラウザ検証のPASSではない。CLI終了・実モデル確認・skill確認・計画の受入はローカル台帳に別々に記録する。
