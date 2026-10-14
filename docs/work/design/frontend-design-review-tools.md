# フロントエンドデザイン改善ツールの選定

**調査日:** 2026-10-13\
**対象:** Animal Ekarte の React 19 / TypeScript フロントエンド\
**目的:** デザイン逸脱の発見、実画面の視覚レビュー、アクセシビリティ修正に使う外部ツール候補を比較し、導入優先度を決める。

**採用状況:** 2026-10-13、ユーザーの採用指示により `improve-ui` と `visual-qa` をCodexのユーザースキルとして導入。インストール済みファイルと参照ファイルを確認済み。画面の監査・撮影・改善実装は未実施。

## 結論

最初に試す候補は **`ibelick/ui-skills` の `improve-ui`**。このプロジェクトには `DESIGN.md`、製品色の正本である `docs/spec/design-system.md`、状態・コントラストの `docs/spec/design-states.md`、実装トークン `frontend/src/lib/design-tokens.ts` がすでにある。improve-ui はそのプロジェクト固有の規約を根拠に、選んだ画面だけを監査し、修正計画を作る用途に合う。skill の説明ではソースを変更せず、`design-plans/` に計画だけを書き出す境界が明記されている。

実際の表示状態まで確認したい場合は、既存のブラウザ運用を使ってスクリーンショットを用意し、`dylanfeltus/skills` の `visual-qa` を補助レビューとして試す。これは視覚レビューの補助であり、現行の機械監査やブラウザ受け入れを置き換えない。

`impeccable` は広範なデザイン改善コマンドとフックを含むため、現時点では全体導入を推奨しない。採用する場合も、既存の製品規約を読み込ませた上で `critique` / `audit` のレビュー用途に限定して試行する。`polish` 等の実装系コマンドや edit hook は別途評価が必要。

## このプロジェクトで優先する基準

ツールの出力より、既存の正本と製品上の安全条件を優先する。

1. 色は [`docs/spec/design-system.md`](../../spec/design-system.md)、タイポグラフィ・形状・余白・寸法はルート [`DESIGN.md`](../../../DESIGN.md)、状態とコントラストは [`docs/spec/design-states.md`](../../spec/design-states.md) を参照する。
2. 実装上は `C` / `STYLE` 等の既存 design token と既存コンポーネントを使う。色の直書き、任意値追加、独自のデザイントークン提案をそのまま採用しない。
3. 危険・死亡・権限・診療状態の表示は臨床上の意味を持つ。見た目を整える目的で色、文言、状態、情報量を変えない。臨床安全上の挙動は [`ui-design-compliance.md`](../../spec/ui-design-compliance.md) の C6a も参照する。
4. ツールによるコード編集・依存追加・設定変更は、監査と計画の評価後に、対象と差分を確認してから別作業として行う。
5. スクリーンショットの所見はその撮影条件の範囲に限る。対象ルート、画面状態、viewport、ブラウザを記録し、未確認の状態を合格扱いしない。

## 候補一覧

| リポジトリ / 機能 | 得意なこと | Animal Ekarte との適合 | 留意点 | 判断 |
|---|---|---|---|---|
| [ibelick/ui-skills — improve-ui](https://github.com/ibelick/ui-skills/tree/main/skills/improve-ui) | 既存のデザインシステムと実装経路を根拠に、画面の逸脱を絞り込んで計画化 | **高い。** 既存の DESIGN.md・仕様・トークンを使う監査に合う。計画のみの運用境界も明示 | 計画の保存先として `design-plans/` を使う前提がある。初回は対象面を狭くし、計画をレビューしてから実装に移す | **第一候補** |
| [dylanfeltus/skills — visual-qa](https://github.com/dylanfeltus/skills/tree/main/visual-qa) | スクリーンショットを見てレイアウト、文字、色、階層、レスポンシブ等をレビュー | **中〜高。** 実レンダリングの見落としを補う用途に合う | 撮影・画面状態・viewportの準備が必要。指摘は画像上の証拠に基づくレビューであり、ソース修正やブラウザ操作の受け入れそのものではない | **補助候補** |
| [pbakaus/impeccable](https://github.com/pbakaus/impeccable) | `critique` / `audit` / `polish` などの設計レビュー・改善フロー、決定論的検出、ライブ確認 | **中。** Reactで使えるが、汎用的な見た目改善より本プロジェクトの規約準拠を優先する必要がある | `npx impeccable install` はスキルに加え編集時フック等を導入しうる。任意の初期値・anti-pattern規則を色や臨床UIの判断に流用しない。導入前に生成差分とhook内容の確認が必要 | **限定試行のみ** |
| [aladicf/better-react-web-ui](https://github.com/aladicf/better-react-web-ui) | React UIの実装・レビューに使うパターン集 | **中〜低。** Reactとの相性はあるが、このプロジェクト固有の規約を正確に監査する主目的とは異なる | Tailwind等への依存や一般的なスタイル提案が、既存トークン・コンポーネントを迂回しないか確認が必要 | **現時点では保留** |
| [cdmx-in/ui-review](https://github.com/cdmx-in/ui-review) | 複数viewportで溢れ、折り返し、横スクロール、重なり等を検出し、ソース位置と結び付ける | **中。** レスポンシブ不具合の診断に有用そう | README上はClaude Code skill と `agent-browser` CLI を前提としている。プロジェクトのブラウザ運用指定は `stagehand_*` MCP のため、互換性・実行経路は未確認 | **実行方式を確認してから** |
| [huanglvjing/spotpatch](https://github.com/huanglvjing/spotpatch) | ブラウザ上で要素を選び、Reactソース位置へ結び付けてレビュー・修正依頼を作る | **低〜中。** Vite向けのため構成上は候補になる | READMEではReact 19がViteアダプターのサポート範囲外で実験扱い。Vite設定やdev runtimeへの追加も必要になるため、現状は互換性リスクが高い | **採用しない（再評価待ち）** |
| [PugarHuda/jangkau](https://github.com/PugarHuda/jangkau) | axe-core等でアクセシビリティ違反を監査し、修正後の再監査・コンテンツ保持・画像差分確認を行う | **低〜中。** a11y修正の考え方は参考になる | 2026-10-12のハッカソンで作成された新しいエージェント。READMEの実行例はPython/Playwright/LLMキーを必要とし、HTML/CSS編集やPR作成まで可能。React/TSXの安全な修正適合は未確認 | **現時点では導入しない** |

「人気」やスター数は適合性の評価に使わない。公開READMEとスキル定義から読み取れる機能・前提で比較し、導入時には対応状況を再確認する。

## 採用後の監査範囲と実行手順

2026-10-13のユーザー指示により、対象は**すべての一覧・登録・編集画面**。ページだけでなく、モーダル・ダイアログ・サイドパネル・タブ内・インラインのフォームも含む。共用フォームでも登録/編集は別状態として確認する。

| 棚卸しする画面群 | 起点となるソース |
|---|---|
| 受付・飼主・予約 | `frontend/src/app/routes/clinical-general-routes.tsx` |
| カルテ・入院・トリミング・検査・予防接種・健診 | `frontend/src/app/routes/clinical-care-routes.tsx` |
| 在庫・見積・シフト | `frontend/src/app/routes/clinical-business-routes.tsx` |
| 会計・締め処理/履歴 | `frontend/src/app/routes/accounting-routes.tsx` |
| 各マスタ・設定・バグ報告/チャット履歴等の一覧 | `frontend/src/app/routes/settings-routes.tsx` |
| LINE予約設定・ページ編集・本人紐付け等の該当面 | `frontend/src/app/routes/operations-routes.tsx` |
| LIFF / LINE予約アプリ内の該当面 | `frontend/liff/src/App.tsx` / `frontend/line-reserve/src/App.tsx` |

この表は棚卸しの入口。監査実行時に参照先コンポーネントまで追い、全ページ/フォームの在庫を各Planeチケットへ作成する。アプリごとのデザイン規約を適用し、本体の規約をLIFF/LINE予約へ一律に適用しない。詳細専用・ダッシュボード・帳票・認証・説明ページは、一覧や登録編集機能を含む場合に当該部分を監査する。リダイレクトは重複計上しない。

1. 各チケットへ、アプリ・画面群・ルート/入口・画面種別・対象ソースを含む監査在庫を用意する。確認状態は「未着手・確認済み・理由付き未確認・対象外」で記録する。
2. 導入済みの `improve-ui` で、画面群ごとにルートからコンポーネント、token、CSSへ追跡する。1回の報告は画面群ごとに最大3件であり、全体の指摘数を3件に制限しない。計画はユーザーが選択した所見について作成する。
3. 全対象面について、既存のブラウザ手順で許可された環境のスクリーンショットを撮り、`visual-qa` で見た目の所見を確認する。通常/空/長文/入力エラー等の状態とviewportを記録し、取得できない状態は理由付き未確認にする。合成または匿名化済みデータを使う。
4. 所見を次の3分類でレビューする: **既存規約違反**（修正候補）、**正当な既存判断**（採用しない）、**仕様不足**（正本の追記が必要かを別途判断）。ツールの自動指摘だけで仕様やコードを変更しない。
5. 改善が採択された場合は通常のフロントエンド作業として差分を実装・レビューし、変更範囲に応じたDockerベースの検証とブラウザ確認を行う。静的監査、スクリーンショットレビュー、実ブラウザ受け入れの結果は区別して記録する。

### 導入記録と呼び出し方

Codexの `skill-installer` のPythonヘルパーで、確認したコミットを指定してインストールした。`npm` / `npx` は実行していない。スキル本体はユーザー設定側にあり、アプリの依存やMCP・hookの追加はない。

| スキル | インストール先 | 固定したコミット |
|---|---|---|
| `improve-ui` | `~/.codex/skills/improve-ui/` | [`f2264c30213a8a55bd0a6a7668d0438e4564e090`](https://github.com/ibelick/ui-skills/commit/f2264c30213a8a55bd0a6a7668d0438e4564e090) |
| `visual-qa` | `~/.codex/skills/visual-qa/` | [`b97a48f8f7bc0b5cef09f219cd79217190d6817e`](https://github.com/dylanfeltus/skills/commit/b97a48f8f7bc0b5cef09f219cd79217190d6817e) |

Codexでは次のターンから利用する。呼び出し例:

```text
$improve-ui すべての一覧・登録・編集画面を、モーダルやサイドパネルを含めて監査してください。
まず全対象の在庫を作り、画面群ごとに順に確認してください。
色は docs/spec/design-system.md、その他の外観は DESIGN.md、状態は
docs/spec/design-states.md に従ってください。
コードは変更せず、所見をPlane EMR-372へ記載してください。
```

```text
$visual-qa すべての一覧・登録・編集画面を、モーダルやサイドパネルを含めて
実画面でレビューしてください。全対象の在庫と撮影した画面状態・viewportを記録し、
画面群ごとに順に確認してください。
画像から確認できないキーボード操作・focus遷移・コントラスト比は未確認としてください。
コードは変更せず、所見をPlane EMR-373へ記載してください。
```

`visual-qa` 内の `browser` / `image` 等は汎用的なツール表記。実行時は利用可能なツールへ対応させ、ブラウザ操作はプロジェクトのルーティング規則に従う。画面の立ち上げやアクセス許可、実画像の用意が済んでいない場合は、画像レビューを実施済みとしない。スキル中の一般的な寸法・タイポグラフィ例は製品の正本を上書きしない。

`improve-ui` の計画作成先は `design-plans/`。通常の監査は読み取りのみで実施し、ユーザーが改善項目を選んだ後に計画を作る。この文書自体はツール選定資料であり、画面監査の所見や実装計画ではない。

### 指摘の記録先

監査後の優先改善案は [一覧・会計明細の優先改善計画](frontend-design-improvement-plan-20261013.md) を参照する。EMR-379・384の対象コード・変更方針・受入条件と、残る所見の対応順序をまとめている。

2026-10-13のユーザー指示により、スキルごとに専用の調査チケットを1件ずつ作成した。指摘と現在の状態の正本はPlaneとし、この文書へ指摘の一覧を複製しない。

| スキル | 専用チケット | 主な記録内容 |
|---|---|---|
| `improve-ui` | [EMR-372](https://app.plane.so/baritechllc/browse/EMR-372/) | 対象SHA・ルート、規約と実装経路の証拠、修正案、影響、確信度 |
| `visual-qa` | [EMR-373](https://app.plane.so/baritechllc/browse/EMR-373/) | 画面状態・viewport・証拠画像、対象要素、問題、優先度、未確認事項 |

上の例で全対象を指定して実行する。`improve-ui` はソースから監査できる。`visual-qa` は対象画像を添付するか、撮影する環境のURLを指定する。ローカル環境の起動・認証等が必要な場合は、既存のプロジェクト手順とアクセス範囲に従う。全対象の確認状態を追跡し、未確認の画面を監査済み扱いしない。

指摘には `IU-001` / `VQ-001` 等のIDを付け、本票の本文または日時付きコメントへ記録する。追試では既存記載を保持し、同じ原因の所見は相互参照する。指摘がない場合も対象と確認範囲を記載し、未監査と区別する。

`improve-ui` の既定の書き込み境界に対し、ユーザーは監査結果のPlaneへの記載を明示しているため、本票への記録を実施する。製品ソースの編集はこの監査には含めない。

## 参照したプロジェクト文書

- [DESIGN.md](../../../DESIGN.md)
- [UIデザインシステム](../../spec/design-system.md)
- [デザイントークン層](../../spec/design-token-layers.md)
- [デザイン状態・コントラスト](../../spec/design-states.md)
- [UIデザイン準拠監査](../../spec/ui-design-compliance.md)
- [フロントエンドルール](../../../frontend/CLAUDE.md)

## 外部参照

- [improve-ui skill](https://github.com/ibelick/ui-skills/tree/main/skills/improve-ui)
- [visual-qa skill](https://github.com/dylanfeltus/skills/tree/main/visual-qa)
- [Impeccable README](https://github.com/pbakaus/impeccable)
- [better-react-web-ui README](https://github.com/aladicf/better-react-web-ui)
- [ui-review README](https://github.com/cdmx-in/ui-review)
- [SpotPatch README](https://github.com/huanglvjing/spotpatch)
- [Jangkau README](https://github.com/PugarHuda/jangkau)
