# 見積の業務契約

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | domain.estimates.business-rules／業務仕様 |
| 決定状態 | 2026-10-20追加本人回答は末尾の該当節を参照。2026-10-17追加決定は末尾と進捗入口を参照。従来：既存契約を継承。E1は依頼者2026-10-15承認（Sentinel_5b14560fa7308191a92a2882c517a23d）。期限内の提示単価・提示後の改訂は追加承認（Sentinel_1fda6cf121048191943efab56cc16a23）。 2026-10-23：D2通常削除から親未確定条件だけを廃止（下記決定）。移行原本証拠の保全/通常台帳との分離は既決方針。EMR-547は共通構造・保存tableの役割・関係制約（chunk/field/payload上限は未選定）・readonly APIを2026-10-23選定。6入れ子型・verification_ref検証記録/供給binding・配送local→stable UUID耐久対応と実データprofile/manifest/集約入力は未閉鎖、全wire完成/実装済みではない。 |
| 実装状態 | 現行差を分離。今回実装変更・網羅監査なし。EMR-547の保存/read/UI接続は未実装。 |
| 実機検証状態 | 未実施。受入条件は検査定義。今回tests/build/codegen/DB/import/browserはNOT_RUN。 |
| 文書確認日 | 2026-10-23（移行証拠保全の設計境界/語彙追補。従来確認2026-10-20は既存範囲） |

## 目的・対象範囲

D11は費用の説明、見積の保存・提示・承認／却下・後継訂正を扱う。見積承認は診療実施・カルテ確定・請求確定・入金・在庫出庫を意味しない。D10の明細必須という請求条件を、既存の見積下書きまで遡及適用しない。実装上の独立フォームとカルテタブの違いは[技術整理](../../../design/domains/finance-operations/technical-spec.md#estimate-current-gap)へ。

<a id="requirement-fr-estimates-legacy-evidence-preservation"></a>
## FR-estimates-legacy-evidence-preservation：旧見積の原値・由来の参照保全

Minoru Sogaの2026-10-23保全/閲覧整備承認のうち見積部分。旧見積の割引原値と丸め値/規則、保険等の生成定数・未取得・意味不明を区別して保持/表示し、通常見積の計算revisionへ未知値を詰めない。通常編集/提示/再計算/帳票/会計化やformal importの保留解除を許可しない。

| 条件ID | 前提/操作 | 期待結果 | 確認手順/証跡 |
|---|---|---|---|
| AC-estimates-legacy-evidence-discount-provenance | 旧割引のraw/丸め値/負値/欠損と生成定数を含む旧見積を参照 | 元値・丸め値・規則/由来を別に保持/表示。生成insurance=0/rate=0/tax/status等を元情報の確認済み事実と見せず、割後単価へ割引を二重適用しない。通常見積の原本を変更しない | 合成値でraw/derived/表示/通常見積不変を対照。NOT_RUN |

本FR/1ACに加え、[会計側6共通AC](../accounting/lifecycle.md#requirement-fr-accounting-legacy-evidence-preservation)の関係履歴/再送/現在認可/read-only/通常台帳分離/機密を継承する。見積証拠は現在の見積閲覧権限を用い、会計閲覧から推定許可しない。新しいFR/AC IDは既存5FR/24ACの確認範囲と分ける。実装・実機は未完、選定済み共通型/保存/readと未閉鎖依存は[移行証拠設計](../../../design/domains/accounting/legacy-finance-evidence.md#legacy-evidence-read-api-ui-candidates)へ。

## 担当者・権限・前提

同一医院の見積を扱い、閲覧・作成・編集・削除はそれぞれ estimates の操作権限に従う。割引編集は追加の discount 権限を確認する。新しい職種限定や全スタッフ開放はしない。飼主・ペットは確認済み共通患者の自院利用を認め、見積・カルテ参照の医院・相互整合と操作認可を検査する。他医院の診療・見積IDへ無権限に付け替えない。共通基本情報と院別ID対応は設計で接続する。後継作成者の医院所属も検査する。

## 契機・入口・出口

[一覧](../../screens/22-estimate-list.md)から[作成／編集](../../screens/23-estimate-form.md)・[詳細](../../screens/26-estimate-detail.md)へ進む。カルテ見積タブもD11の別入口。保存結果は見積ID・番号・状態・明細・関連参照を返し、カルテの画面遷移だけで保存成功としない。将来の会計受渡しは下記E1の承認範囲内で別に扱う。

## 状態遷移・引継ぎデータ

<a id="requirement-fr-estimates-state-protection"></a>
### FR-estimates-state-protection：状態と確定保護

現行実装では新規作成は下書き／送付済みで、下書き・送付済みは権限付きで編集でき、承認済み／却下へ移行すると原見積を編集・削除・下書きへ戻せない。2026-10-15追加承認の目標では、飼主へ提示済みなら返答待ちでも提示原本を保持し、金額・明細の直接上書きや削除で失わせない。変更は理由付き後継を作り再提示する。提示版を識別して承認／却下を記録する状態遷移は可能だが、内容の改変とは分離する。未提示下書きの通常編集は維持し、送付済みの現行編集可を新目標の例外にしない。送付済みという状態変更だけで外部へ自動送信したことにはならない。見積の状態とカルテの確定状態は別に扱う。親draftを要求する旧更新gateは現行実装との差として扱い、目標では未提示下書きの通常編集を親カルテの確定状態だけで禁止しない。見積自身の提示状態・期待版・医院と親子関係・権限を検査し、提示原本や確定見積の保護を維持する。2026-10-23の[通常削除条件の限定訂正](#estimates-delete-parent-state-20261023)により、通常削除も親カルテ未確定を条件にしない。

<a id="estimates-delete-parent-state-20261023"></a>
### 2026-10-23：通常削除の親カルテ条件だけを廃止

本人が「親カルテが未確定であること」だけを見積の通常削除条件から外すと回答した（Sentinel_ec171df8a1408191977cfb41a21cb12d）。見積自身の提示原本/承認・却下保護、編集可能状態、期待版、医院/親子関係、削除権限、明細/受渡し依存と履歴保全は維持する。診療終了・医師確認・会計確定を代替の削除禁止条件にしない。削除に伴って親や見積原本を再開せず、元履歴を消さない。これはEMR-495 D2の既決条件訂正であり、製品削除writerや試験の変更/実行は今回行わない。

<a id="requirement-fr-estimates-successor"></a>
### FR-estimates-successor：原本を残す後継訂正

承認済み／却下見積の訂正は理由付きの新しい後継下書きとして作る。追加承認により、提示済み・返答待ちの見積も金額・明細を変更する場合は同じ後継方式に含め、変更内容を再提示する。元見積への承認を内容の異なる後継の承認へ自動継承しない。原番号・状態・明細を保持し、新ID／番号から原見積を参照する。元明細・ヘッダ値・宛先・有効期限等のコピーと、今回の変更を区別する。承認済み目標では、親カルテが確定済みでも親や原見積を再開せず、専用の後継訂正経路で必要な変更を保存し、確認した内容を未確定会計へ受け渡すところまで可能にする。現行は後継createのみ親draft保護の例外で、その後の通常updateが親確定を理由に拒否する実装差がある。後継は元版・後継関係・理由・権限・未提示の編集可能状態を専用経路で検査し、この現行差を解消する。通常の未提示下書き編集も親確定だけを拒否理由にしないが、提示原本・確定見積を通常編集へ戻す一般unlockは追加しない。後継が作れたことだけで編集・会計化の完了としない。

<a id="requirement-fr-estimates-money-source"></a>
### FR-estimates-money-source：金額の出所を明示

現行継承：独立フォームも[明細行エディタを持つ](../../screens/23-estimate-form.md#12-金額セクション)。明細ありでは金額を明細から導出し、明細なし（既存のヘッダのみ見積、および新規作成で明細未入力）ではヘッダ金額の手入力が残る。カルテタブも明細を持つ。保存済み明細がある見積ではサーバーが明細から金額を再計算し、ヘッダだけの更新で明細との不整合を作らない。明細省略と明細空配列（全行削除）は異なる。現在の一律10%表示・ヘッダ割引・保険額の扱いをD10の新計算契約に適合済みとはしない。

承認済みE1：新規見積は入口を問わず明細方式へ統一し、D10の税・割引・端数・計算根拠保持を再利用する。保険は概算であることを明示し、窓口精算時に再確認する。既存ヘッダのみ見積は当時内容を保全し、明細があると偽装しない。新方式切替と旧ヘッダ見積の保全・移行停止は[選定技術契約](../../../design/domains/finance-operations/technical-spec.md#estimate-selected-contract)に従う。

<a id="requirement-fr-estimates-billing-handoff"></a>
### FR-estimates-billing-handoff：実請求との分離

D10／D03の既決境界を継承：見積保存／承認／後継作成のみで請求・実入金・在庫実績を生成しない。請求確定はD10、臨床実施はD03等、在庫実績はD13が所有する。

承認済みE1：スタッフが飼主・対象ペット・実際に請求する明細・金額を確認して未確定会計へ受け渡す。元見積／版／明細の参照を保持し、現在の価格／税／割引との相違を確認する。期限内の承認済み見積と内容・数量が同じ場合は、マスタ単価が変わっていても提示した単価を使用する。追加処置や内容・数量の変更は改めて説明し、提示済み見積を修正する場合は後継を再提示する。差分確認だけを理由に保証対象の単価を現在マスタへ自動置換しない。見積承認や転記を実施記録の作成に代用せず、既に請求した同一実績を再請求しない。会計化の成功判定は受渡し先の保存結果で行う。

<a id="requirement-fr-estimates-validity"></a>
### FR-estimates-validity：有効期限と説明

既存の有効期限は見積提示条件を確認する日付であり、期限到来で自動入金・自動取消を行う契約はない。承認済みE1：期限切れを警告し、条件を再確認して後継見積を作る。無回答を期限延長や旧価格保証の承認としない。

<a id="estimate-price-presentation-agreement-20261015"></a>
### 2026-10-15追加合意：期限内の提示単価と提示原本

承認済み・有効期限内・内容と数量が同じという条件を満たす場合に提示単価を使う。今回の価格保証は単価の取り扱いであり、税率・保険負担額・割引の適用条件や請求総額を無条件で固定する承認ではない。D10の適用計算・概算保険の窓口再確認・既存割引条件を維持し、単価とその他の差分を分けて確認する。税務条件の新規決定や自動請求確定を追加しない。返答待ちを承認済みとせず、期限切れは従来どおり条件再確認と後継再提示へ進む。提示原本の単価・数量・明細・当時の計算根拠は、後継や会計の再計算で書き換えない。

## 異常系・保存・取消

- 通常削除は権限に加え見積自身の編集可能状態・期待版・医院/関連整合・明細/受渡し依存なしを検査し、履歴を保全する。親カルテ未確定は条件にしない。明細が残る見積を削除して消さない。
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
| AC-estimates-parent-finalized | 確定カルテ配下で未提示下書きと提示済み原見積の金額・明細変更をそれぞれ要求 | 未提示下書きは見積自身の状態・版・関係・権限が適合すれば親確定だけで拒否しない。提示済み原見積は直接変更を拒否し専用後継へ。通常削除も親未確定を条件とせず、見積自身の削除条件・履歴保全を維持 | [FR-estimates-state-protection](#requirement-fr-estimates-state-protection)・[FR-estimates-successor](#requirement-fr-estimates-successor)。旧親draft更新gateとの現行差を検査、未実施 |
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
| AC-estimates-quoted-price | 承認済み・期限内・内容と数量が同じ見積を会計へ受け渡し、現在マスタ単価だけが変更済み | 提示単価を使い、元版・行参照と単価の出所を保持する | [FR-estimates-validity](#requirement-fr-estimates-validity)。追加承認目標、未実施 |
| AC-estimates-changed-content-explain | 追加処置または見積内容・数量を変更する | 改めて説明し、提示済み見積の変更は後継を作り再提示する。旧原本は不変 | [FR-estimates-successor](#requirement-fr-estimates-successor)。追加承認目標、未実施 |
| AC-estimates-presented-pending-original | 提示済み・返答待ちの見積で金額・明細の直接更新／原本削除を要求 | 原本を保持して直接変更を拒否し、後継改訂の経路を示す | [FR-estimates-state-protection](#requirement-fr-estimates-state-protection)。追加承認目標、未実施 |
| AC-estimates-presented-pending-successor | 提示済み・返答待ちの見積から理由付き改訂を作成 | 別ID／番号・原版参照を持つ後継を作り再提示。旧承認や返答を新内容へ自動継承しない | [FR-estimates-successor](#requirement-fr-estimates-successor)。追加承認目標、未実施 |
| AC-estimates-presentation-edit-race | 同じ未提示下書きに提示と明細変更を同時要求 | 提示した版を固定し、提示後の変更が原本へ混入しない。競合は再読取・再確認する | [FR-estimates-state-protection](#requirement-fr-estimates-state-protection)。追加承認目標、未実施 |
| AC-estimates-quoted-price-other-rules | 提示単価保証対象の見積で税・保険・割引の条件を確認 | 提示単価を保持しつつD10適用計算と保険再確認を維持。単価保証を総額や保険額の無条件固定としない | [FR-estimates-billing-handoff](#requirement-fr-estimates-billing-handoff)。追加承認目標、未実施 |
| AC-estimates-quoted-price-no-approval | 期限内でも返答待ちの見積を扱う | 承認済み限定の単価保証条件を満たしたと推定しない。原本保護は維持する | [FR-estimates-validity](#requirement-fr-estimates-validity)。追加承認目標、未実施 |
| AC-estimates-presented-successor-retry | 返答待ち原本からの後継作成で応答消失後に同じ操作を再送 | 結果照会／同一操作IDで同じ後継へ収束し、重複後継や外部自動送信を作らない | [FR-estimates-successor](#requirement-fr-estimates-successor)。追加承認目標、未実施 |

## 未決事項・関連仕様・根拠

E1（明細統一・会計受渡し・期限切れ）は2026-10-15に依頼者承認済み。質問 Sentinel_75a3a1bf81b481919ba4c34aa633feb4 の「2. 見積から会計」に対する回答 Sentinel_5b14560fa7308191a92a2882c517a23d「2 ok」が根拠。明細方式、D10共通計算、概算保険、スタッフ確認後の未確定会計転記、確定原本保持と変更／期限切れ後継を対象とする。既存実装の適合済みを意味しない。後継の重複排除、親カルテ確定後の後継編集、帳票と金額表示、API／schemaは[設計差](../../../design/domains/finance-operations/technical-spec.md#estimate-current-gap)。旧ヘッダだけの見積を自動請求する移行は承認しない。


### カルテ本文の通常編集／復元との区別

[D03の会議追加合意](../medical-records/lifecycle.md#medical-record-agreement-autosave-20261015)はカルテ本文の通常編集・自動保存・タブ復元を扱う。見積の承認／却下原本・専用後継・会計転記の既決条件を解除する承認ではない。カルテタブに見積を表示していても、タブクリア／複製Undoから原見積削除・後継取消・請求の逆実行を行わない。旧親finalizedの現行write gateと、見積自身の原本／後継保護を区別し、カルテ旧値の一括解除で財務系writerを解禁しない。D11の対象操作は現在の見積状態・版・医院／患者・操作権限と専用経路で照合する。

追加合意の根拠：質問 Sentinel_ff8ad168acec8191be01b7c559dfa858 の「期限内の価格」「提示後の変更」に対し、依頼者 Sentinel_1fda6cf121048191943efab56cc16a23 が「1 yes / 2 yes」と回答。今回2点を超える税率固定・保険額保証・実施／請求の自動確定は承認範囲に含めない。

<a id="decision-alignment-20261017"></a>
## 2026-10-17確定仕様との接続

[決定根拠・適用範囲](../SPECIFICATION_PROGRESS.md#confirmed-business-decisions-20261017)：ユーザーの2026-10-16〜17 JSTの直接決定、文書更新承認は10月17日03:20 JST。以下は目標仕様・未実行の受入条件であり、実装・実機合格を示さない。

[業務正本・受入条件](../medical-records/lifecycle.md#requirement-fr-medical-records-dose-display-unit-price)。通常プラン4種明細の単価編集・通常理由任意を、提示済み見積の原本保護や後継訂正理由必須へ拡張しない。共通患者の自院利用はD04に従い、同一医院の見積記録・参照整合を検証する。他院由来の共通患者であることだけを一律拒否理由にせず、未認可の他院見積ID付替えは拒否する。

## 主飼主変更時の未提示見積の帰属

[D04の2026-10-20本人回答](../owners-pets/lifecycle.md#owner-transfer-draft-retention-20261020)に従い、未提示見積も元飼主へ残し、主飼主変更から自動移管しない。必要なら別途変更するという選択を、提示済み原本の直接書換えや無条件の対象付替えへ拡張しない。見積自身の現在状態・医院／飼主／ペット・親参照・認可・版を検査し、[原本を残す後継訂正](#requirement-fr-estimates-successor)と既決の提示済み原本保護を維持する。主飼主変更成功だけで見積変更・再提示・会計化を成功扱いしない。受入はD04の既存AC-owners-pets-owner-transfer-historyと本書の原本保護条件を照合する（未実施）。
