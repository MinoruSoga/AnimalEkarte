# D22 院内マニュアル

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | domain.internal-manual.index／ドメイン入口 |
| 決定状態 | 混在。既存記載を継承（決定根拠未確認の部分を含む）。追加合意の決定者・日付・根拠は各節のFRと関連正本を参照 |
| 実装状態 | 既存記載を継承（今回コード再監査なし）。目標仕様との既知差分を保持 |
| 実機検証状態 | 未実施。文書静的確認を受入PASSとしない |
| 文書確認日 | 2026-10-20（状態軸・後続合意への参照） |

> 2026-10-15の整理を起点とし、各節に記載した後続合意を適用範囲内の旧記述より優先する。既存継承・合意済み・技術提案を区別し、文書整理を実装／実機検証の完了としない。

## 目的・範囲

P2：操作案内を全スタッフが同じ公開版で参照できるようにする（公開方式は今回合意）。製品内マニュアルと開発仕様の管理資料は別責務。

## 業務正本と優先順

- [docs/spec/screens/35-internal-manual.md](../../screens/35-internal-manual.md)

最新の明示合意と本書のFR／ACを業務の根拠とする。

## 現行実装の参照

- [frontend/src/features/manual/routes/ManualPage.tsx](../../../../frontend/src/features/manual/routes/ManualPage.tsx)
- [backend/internal/manualarticle/service.go](../../../../backend/internal/manualarticle/service.go)

最新の明示合意が適用範囲内の旧記述に優先する。コードは実装事実の根拠で、業務承認の代替ではない。既存FR／ACの本文は元正本に維持し、本書の接続要件で改番・複製しない。

## 業務要件

<a id="requirement-fr-internal-manual-reading"></a>
### FR-internal-manual-reading — 記事閲覧・検索・印刷

認証済み全スタッフが画面別／業務別の記事を閲覧・検索し、表示中の記事を印刷できる。操作を確認しやすいよう説明対象の画面スクリーンショットを記事に添える。根拠は[10月14日メモ](https://docs.google.com/document/d/1UnyGdGyIU8tSZnwRrAXZAxXBk61YrShCH2Eqaz7GoaU)の「取説ページ／急ぎではないが、画面スクショ付きで」（2026-10-14 13:57:02 JSTに曽我稔がSlackで共有。メモの原発言者は未記載）。新しい納期・全記事全画面の収録義務・自動撮影機能を推定しない。記事がない場合は未発見と示し、別記事を目的の記事として表示しない。マニュアル閲覧から診療情報の変更権限を得ない。

決定状態・根拠：既存仕様継承。印刷は現行画面実装の参照（実機未確認）。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

<a id="requirement-fr-internal-manual-publication"></a>
### FR-internal-manual-publication — 編集権限と公開版

医院共通の記事を対象編集権限者が編集し、公開を確定した版を全認証スタッフが閲覧する。一般職員へ編集権限を与えることと、公開済み記事を読ませることを分離する。未保存・公開未確定の編集案は公開版として表示せず、保存失敗を公開済みと扱わない。検索・印刷も同じ公開版を使う。

決定状態・根拠：合意済み：2026-10-15 ユーザー Sentinel_7a67fd30e47c819192cd36d416f0a2bf の「ok」。確認文 Sentinel_75a3a1bf81b481919ba4c34aa633feb4／Sentinel_9c0eb79acfd481919738187e9769f8cd と対象確認 Sentinel_8d0032ca45b881919b3d2ceb5339f792。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

<a id="requirement-fr-internal-manual-version-retention"></a>
### FR-internal-manual-version-retention — 変更履歴と削除

編集・公開・過去版復元の履歴を残す。旧版復元で現在版を更新しても、過去の変更記録を消さない。版の識別・編集者・時点の保存方式と競合検出は技術設計へ分ける。記事削除と静的版への復帰は現行の操作意味を維持し、今回の承認を履歴の永久削除許可へ拡張しない。

決定状態・根拠：合意済み：2026-10-15 ユーザー Sentinel_7a67fd30e47c819192cd36d416f0a2bf の「ok」。確認文 Sentinel_75a3a1bf81b481919ba4c34aa633feb4／Sentinel_9c0eb79acfd481919738187e9769f8cd と対象確認 Sentinel_8d0032ca45b881919b3d2ceb5339f792。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

## 受入条件

全条件は文書上の期待結果。実機未実施。技術方式の提案は実装設計として区別する。

| 条件ID | Given（前提） | When（操作） | Then（期待結果） | 独立確認 | 成功条件 | 確認手順／証跡 |
|:---|:---|:---|:---|:---|:---|:---|
| AC-internal-manual-01 | 記事が存在する認証済み利用者 | 画面別・業務別検索から記事を開き印刷 | 選択した記事本文と説明対象の画面スクリーンショットを閲覧・印刷できる | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-internal-manual-reading](#requirement-fr-internal-manual-reading)。実機未実施 |
| AC-internal-manual-02 | 存在しないcategory/slug | 直接記事URLを開く | 未発見と表示する | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-internal-manual-reading](#requirement-fr-internal-manual-reading)。実機未実施 |
| AC-internal-manual-03 | 編集権限のないスタッフ | 公開済み改訂記事を読む／編集を試す | 最新版閲覧は可、編集は拒否 | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-internal-manual-publication](#requirement-fr-internal-manual-publication)。実機未実施 |
| AC-internal-manual-04 | 編集内容入力済みで保存失敗 | 保存する | 公開成功を表示せず入力を保持 | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-internal-manual-publication](#requirement-fr-internal-manual-publication)。実機未実施 |
| AC-internal-manual-05 | 変更履歴のある記事 | 過去版を明示復元 | 新しい版として保存し旧履歴を保持 | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-internal-manual-version-retention](#requirement-fr-internal-manual-version-retention)。実機未実施 |
| AC-internal-manual-06 | 未保存または公開未確定の編集内容がある | 一般スタッフが記事を検索・印刷 | 未確定の編集案は公開済み本文として表示しない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-internal-manual-publication](#requirement-fr-internal-manual-publication)。実機未実施 |
| AC-internal-manual-07 | 同じ記事を別端末で編集している | 古い表示版から保存 | 競合を検出して確認を要求し新しい版を無警告に上書きしない（技術提案） | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-internal-manual-version-retention](#requirement-fr-internal-manual-version-retention)。実機未実施 |
| AC-internal-manual-08 | 静的記事に編集overrideがある | 削除する | 静的版への復帰か追加記事の非表示かという現行の結果を区別し、公開・編集・復元履歴を無断で抹消しない。具体UIは設計 | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-internal-manual-version-retention](#requirement-fr-internal-manual-version-retention)。実機未実施 |

## 未決・責務境界

公開版全職員閲覧・編集／公開／過去版復元の履歴は今回合意済み。独立した下書き承認workflow、医院別記事、既読必須を今回新設しない。旧実装の権限付きoverride閲覧との差分、公開版と検索／印刷の一貫性、原子的保存・競合・receiptは技術正本の選定契約、履歴UI実装は検証の残件であり業務回答待ちではない。

## レビュー状態

基準版の本文と関連する静的根拠を照合。独立レビュー・統合リンク検査は統合担当で実施する。仕様記述を実装完了、受入PASS、実環境変更承認へ変換しない。

## 操作・状態・例外の全域点検

- 画面別／業務別一覧、検索、記事直接URL、記事未発見、モバイルナビ、印刷を対象にする。検索と印刷が本文とは別の旧版を使うことを許容しない契約は今回の公開版合意に従う。
- 現行の保存はDB overrideであり、下書き／公開の分離モデルが実装済みとは言わない。今回の合意は編集権限者が公開を確定する操作を定義するもので、未保存編集や非公開案を全員へ露出する承認ではない。独立した下書き承認ワークフローを自動追加しない。
- 競合保存は元版識別と最新取得で検出し、入力案と最新公開版を比較して再確認する技術提案。後勝ち無警告上書き・失敗時公開成功表示を避ける。実装方式／保存履歴schemaはdesignで決める。
- 編集権限が失効した場合の保存拒否・非公開編集案の閲覧制限をC04へ接続する。一般公開READは認証済みスタッフ向けであり、未認証Internet公開ではない。
- 公開／編集／復元の操作者・時点と版の関連を履歴に保持する。理由の必須入力欄は本合意から新設しない。復元と破壊的な履歴削除を分ける。記事削除と静的版への復帰のUI意味を操作前に確認する。
- マニュアル本文は医院共通で、個人患者記録の新たな共有場所を設ける意味ではない。医療記録の原本は各業務ownerが管理する。


## 保存・例外の追加受入条件

| 条件ID | Given（前提） | When（操作） | Then（期待結果） | 独立確認 | 成功条件 | 確認手順／証跡 |
|:---|:---|:---|:---|:---|:---|:---|
| AC-internal-manual-09 | 編集公開の必須履歴／監査保存が失敗 | 公開する | 内容版・公開pointer・履歴・receiptを全てrollbackし旧公開版を維持 | 単独確認 | Thenの結果と失敗時の不変 | [選定保存契約](../../../design/domains/internal-manual/technical-spec.md#manual-selected-save-contract)。実機未実施 |
| AC-internal-manual-10 | 同じ記事を別操作が更新済み | 古い期待版で保存 | 競合拒否。公開原本不変、入力案を保持して比較へ | 単独確認 | Thenの結果と失敗時の不変 | [選定保存契約](../../../design/domains/internal-manual/technical-spec.md#manual-selected-save-contract)。実機未実施 |
| AC-internal-manual-11 | 保存応答が不明 | 同じ操作IDで照合 | 既存結果を返し新たな版／公開／復元を重複作成しない。不明なら停止 | 単独確認 | Thenの結果と失敗時の不変 | [選定保存契約](../../../design/domains/internal-manual/technical-spec.md#manual-selected-save-contract)。実機未実施 |
| AC-internal-manual-12 | 旧overrideの公開状態または復元元が不明 | 全員公開または復元する | 自動公開・架空履歴で補完せず対象確認まで停止 | 単独確認 | Thenの結果と失敗時の不変 | [選定保存契約](../../../design/domains/internal-manual/technical-spec.md#manual-selected-save-contract)。実機未実施 |

公開／編集／復元の操作ID・期待版・原子的な内容／履歴／監査／receipt保存、旧版競合・結果不明時停止は[選定技術契約](../../../design/domains/internal-manual/technical-spec.md#manual-selected-save-contract)へ従う。技術方式を後決めの業務未決に戻さない。

## 複数ページ記事の印刷の補足受入条件

保存済みSlack SLK-043〜046（6/10、投稿者識別子は本リポジトリに記録しない）の「1ページしか印刷されない」という不具合と修正報告を、既存の表示中記事印刷契約へ接続する。修正予定・テスト環境反映の報告を実機合格へ転用しない。SLK-047の「左の部分」は添付PDFの対象箇所を未確認であり、記事中の説明用スクリーンショットを切除する許可にはしない。

| 条件ID | 前提／操作 | 期待結果 | 根拠・検証状態 |
|:---|:---|:---|:---|
| AC-internal-manual-multipage-print | 複数ページにわたる公開記事を表示し、ブラウザ印刷／PDF保存する | 選択記事の本文・説明対象スクリーンショットを最後のページまで出力でき、先頭1ページだけに欠落しない。別記事や未公開編集案を混ぜない | SLK-043〜046、既存記事印刷契約。未実施 |
