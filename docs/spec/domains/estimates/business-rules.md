# 見積の業務契約

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | domain.estimates.business-rules／業務仕様 |
| 決定状態 | 既存契約を継承。E1は依頼者2026-10-15承認（Sentinel_5b14560fa7308191a92a2882c517a23d）。 |
| 実装状態 | 現行差を分離。今回実装変更・網羅監査なし |
| 実機検証状態 | 未実施。受入条件は検査定義 |
| 文書確認日 | 2026-10-15 |

## 目的・対象範囲

D11は費用の説明、見積の保存・提示・承認／却下・後継訂正を扱う。見積承認は診療実施・カルテ確定・請求確定・入金・在庫出庫を意味しない。D10の明細必須という請求条件を、既存の見積下書きまで遡及適用しない。実装上の独立フォームとカルテタブの違いは[技術整理](../../../design/domains/finance-operations/technical-spec.md#estimate-current-gap)へ。

## 担当者・権限・前提

同一医院の見積を扱い、閲覧・作成・編集・削除はそれぞれ estimates の操作権限に従う。割引編集は追加の discount 権限を確認する。新しい職種限定や全スタッフ開放はしない。飼主・ペット・カルテ参照を持つ場合は同医院・相互整合を検査し、他医院のIDへ付け替えない。後継作成者の医院所属も検査する。

## 契機・入口・出口

[一覧](../../screens/22-estimate-list.md)から[作成／編集](../../screens/23-estimate-form.md)・[詳細](../../screens/26-estimate-detail.md)へ進む。カルテ見積タブもD11の別入口。保存結果は見積ID・番号・状態・明細・関連参照を返し、カルテの画面遷移だけで保存成功としない。将来の会計受渡しは下記E1の承認範囲内で別に扱う。

## 状態遷移・引継ぎデータ

<a id="requirement-fr-estimates-state-protection"></a>
### FR-estimates-state-protection：状態と確定保護

既存契約を継承。新規作成は下書き／送付済み。下書き・送付済みは権限付きで編集でき、承認済み／却下へ移行すると原見積を編集・削除・下書きへ戻せない。送付済みという状態変更だけで外部へ自動送信したことにはならない。見積の状態とカルテの確定状態は別で、親カルテ確定後の通常子更新禁止も保持する。

<a id="requirement-fr-estimates-successor"></a>
### FR-estimates-successor：原本を残す後継訂正

既存契約を継承。承認済み／却下見積の訂正は理由付きの新しい後継下書きとして作る。原番号・状態・明細を保持し、新ID／番号から原見積を参照する。元明細・ヘッダ値・宛先・有効期限等のコピーと、今回の変更を区別する。承認済み目標では、親カルテが確定済みでも親や原見積を再開せず、専用の後継訂正経路で必要な変更を保存し、確認した内容を未確定会計へ受け渡すところまで可能にする。現行は後継createのみ親draft保護の例外で、その後の通常updateが親確定を理由に拒否する実装差がある。専用経路でこの差を解消し、一般unlockは追加しない。後継が作れたことだけで編集・会計化の完了としない。

<a id="requirement-fr-estimates-money-source"></a>
### FR-estimates-money-source：金額の出所を明示

現行継承：独立フォームはヘッダ金額を直接入力し、カルテタブは明細を持つ。保存済み明細がある見積ではサーバーが明細から金額を再計算し、ヘッダだけの更新で明細との不整合を作らない。明細省略と明細空配列（全行削除）は異なる。現在の一律10%表示・ヘッダ割引・保険額の扱いをD10の新計算契約に適合済みとはしない。

承認済みE1：新規見積は入口を問わず明細方式へ統一し、D10の税・割引・端数・計算根拠保持を再利用する。保険は概算であることを明示し、窓口精算時に再確認する。既存ヘッダのみ見積は当時内容を保全し、明細があると偽装しない。新方式切替と旧ヘッダ見積の保全・移行停止は[選定技術契約](../../../design/domains/finance-operations/technical-spec.md#estimate-selected-contract)に従う。

<a id="requirement-fr-estimates-billing-handoff"></a>
### FR-estimates-billing-handoff：実請求との分離

D10／D03の既決境界を継承：見積保存／承認／後継作成のみで請求・実入金・在庫実績を生成しない。請求確定はD10、臨床実施はD03等、在庫実績はD13が所有する。

承認済みE1：スタッフが飼主・対象ペット・実際に請求する明細・金額を確認して未確定会計へ受け渡す。元見積／版／明細の参照を保持し、現在の価格／税／割引との相違を確認する。見積承認や転記を実施記録の作成に代用せず、既に請求した同一実績を再請求しない。会計化の成功判定は受渡し先の保存結果で行う。

<a id="requirement-fr-estimates-validity"></a>
### FR-estimates-validity：有効期限と説明

既存の有効期限は見積提示条件を確認する日付であり、期限到来で自動入金・自動取消を行う契約はない。承認済みE1：期限切れを警告し、条件を再確認して後継見積を作る。無回答を期限延長や旧価格保証の承認としない。

## 異常系・保存・取消

- 通常削除は権限に加え編集可能状態・明細依存なし・親カルテ未確定を検査。明細が残る見積を削除して消さない。
- 保存中の二重クリック抑止と、サーバーの同時更新保証は別。承認と明細置換が競合した場合、確定後の明細更新を通さない。
- 後継作成の必須監査失敗では新下書きを残さない。通常CRUDの監査は現行best-effortであり、全操作fail-closed実装済みとは書かない。
- 応答消失時は現在の見積・後継関係を読み直し、同じ意図の後継を盲目的に再作成しない。操作ID・内容一致・結果照会・版競合時の再確認は[選定技術契約](../../../design/domains/finance-operations/technical-spec.md#finance-selected-operation-contract)に従う。
- 一覧の検索は取得済み先頭20件内という現行制約を明示し、結果なしを医院内に見積がないことと混同しない。

## 受入条件

既存継承／D10境界の条件は検査定義。E1依存条件は承認された目標の検査定義。実装／実機は未検証。

| 条件ID | 前提／操作（Given / When） | 期待結果（Then） | 確認手順／証跡 |
|:---|:---|:---|:---|
| AC-estimates-create-state | 新規作成で承認済み／却下を指定 | 新規確定状態を拒否 | [FR-estimates-state-protection](#requirement-fr-estimates-state-protection)。状態と応答を照合、未実施 |
| AC-estimates-lock-race | 承認と明細変更を同時要求 | 承認後の変更が原見積に混入しない | [FR-estimates-state-protection](#requirement-fr-estimates-state-protection)。原本／変更結果を照合、未実施 |
| AC-estimates-no-unlock | 承認済み見積を下書きへ変更／削除 | 原本を保持して拒否 | [FR-estimates-state-protection](#requirement-fr-estimates-state-protection)。保存値を照合、未実施 |
| AC-estimates-successor | 確定見積に理由付き後継作成 | 別番号の下書きと原見積参照、原本不変 | [FR-estimates-successor](#requirement-fr-estimates-successor)。新旧ID／項目を照合、未実施 |
| AC-estimates-successor-audit | 後継作成の監査保存失敗 | 後継を成功として残さない | [FR-estimates-successor](#requirement-fr-estimates-successor)。rollback境界を照合、未実施 |
| AC-estimates-parent-finalized | 確定カルテ配下の原見積へ通常変更 | 通常更新を拒否、専用後継との違いを保持 | [FR-estimates-successor](#requirement-fr-estimates-successor)。現行編集差を検出、未実施 |
| AC-estimates-successor-edit-handoff | 確定親の承認済み見積から後継を作り必要箇所を変更 | 親・原見積は不変、専用経路で変更保存後に確認して未確定会計へ渡せる | [FR-estimates-successor](#requirement-fr-estimates-successor)。承認E1目標と現行update拒否差を検査、未実施 |
| AC-estimates-items-omitted | 明細あり見積へヘッダのみ変更 | 既存明細を消さず明細由来金額の整合を保つ | [FR-estimates-money-source](#requirement-fr-estimates-money-source)。nilと空配列を別ケース、未実施 |
| AC-estimates-items-empty | 明細空配列を明示して保存 | 省略扱いせず現行全行置換契約を適用 | [FR-estimates-money-source](#requirement-fr-estimates-money-source)。旧合計残留を確認、未実施 |
| AC-estimates-delete-dependency | 未確定見積に明細が残る状態で削除 | 依存を保持して拒否 | [FR-estimates-state-protection](#requirement-fr-estimates-state-protection)。原本／明細を照合、未実施 |
| AC-estimates-scope | 別医院または不整合飼主／ペットを指定 | 参照を受理せず情報を混在させない | 前提。医院／関連参照を照合、未実施 |
| AC-estimates-no-effects | 見積を保存／承認／後継作成 | 請求確定・入金・在庫増減を発生させない | [FR-estimates-billing-handoff](#requirement-fr-estimates-billing-handoff)。各owner結果を照合、未実施 |
| AC-estimates-list-scope | 先頭ページにない見積名を検索 | 全件検索済みと表示しない | 異常系。ページスコープ表示を照合、未実施 |
| AC-estimates-unified-calculation | E1の目標で、同明細を二つの入口で保存 | 同じD10計算結果と根拠を保持 | 承認E1。両入口のsnapshot比較、未実施 |
| AC-estimates-handoff-confirm | E1の目標で、見積から会計へ受渡し | 対象確認と元明細参照付き未確定会計、二重請求なし | 承認E1。受渡し元／先結果を照合、未実施 |
| AC-estimates-expired | E1の目標で、有効期限を過ぎて再利用 | 警告し再確認・後継へ、無断価格保証なし | 承認E1。境界日と後継を照合、未実施 |

## 未決事項・関連仕様・根拠

E1（明細統一・会計受渡し・期限切れ）は2026-10-15に依頼者承認済み。質問 Sentinel_75a3a1bf81b481919ba4c34aa633feb4 の「2. 見積から会計」に対する回答 Sentinel_5b14560fa7308191a92a2882c517a23d「2 ok」が根拠。明細方式、D10共通計算、概算保険、スタッフ確認後の未確定会計転記、確定原本保持と変更／期限切れ後継を対象とする。既存実装の適合済みを意味しない。後継の重複排除、親カルテ確定後の後継編集、帳票と金額表示、API／schemaは[設計差](../../../design/domains/finance-operations/technical-spec.md#estimate-current-gap)。旧ヘッダだけの見積を自動請求する移行は承認しない。
