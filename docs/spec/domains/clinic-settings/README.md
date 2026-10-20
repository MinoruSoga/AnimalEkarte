# D17 医院・設定入口

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | domain.clinic-settings.index／ドメイン入口 |
| 決定状態 | 2026-10-20追加本人回答は末尾の該当節を参照。混在。既存記載を継承（決定根拠未確認の部分を含む）。追加合意の決定者・日付・根拠は各節のFRと関連正本を参照 |
| 実装状態 | 既存記載を継承（今回コード再監査なし）。目標仕様との既知差分を保持 |
| 実機検証状態 | 未実施。文書静的確認を受入PASSとしない |
| 文書確認日 | 2026-10-20（状態軸・後続合意への参照） |

> 2026-10-15の整理を起点とし、各節に記載した後続合意を適用範囲内の旧記述より優先する。既存継承・合意済み・技術提案を区別し、文書整理を実装／実機検証の完了としない。

## 目的・範囲

P1：医院の操作境界と設定変更の影響を明確にする。個別マスタの全業務内部は各ドメインの責務。

## 正本と優先順

- [docs/spec/screens/19-clinic-settings.md](../../screens/19-clinic-settings.md)
- [docs/spec/screens/20-master-settings.md](../../screens/20-master-settings.md)
- [docs/architecture/auth.md](../../../architecture/auth.md)

最新の明示合意が適用範囲内の旧記述に優先する。コードは実装事実の根拠で、業務承認の代替ではない。既存FR／ACの本文は元正本に維持し、本書の接続要件で改番・複製しない。

## 業務要件

<a id="requirement-fr-clinic-settings-clinic-directory"></a>
### FR-clinic-settings-clinic-directory — 医院と法人の設定単位

[営業時間・休診日の共通基準とLINE受付範囲](#clinic-hours-common-basis-20261020)は2026-10-20本人回答へ従う。医院ごとの基本情報・連絡先・有効状態・帳票設定と、法人共通のインボイス登録番号を分ける。法人共通設定の影響が全医院へ及ぶことを編集時に示す。scope=allは非system_adminへ割当外医院を開示する根拠にしない。最新auth §4.10を優先する。

決定状態・根拠：既存仕様継承。全医院影響の表示は既存の法人共通設定という範囲を明確にするUI設計。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

<a id="requirement-fr-clinic-settings-portal-authorization"></a>
### FR-clinic-settings-portal-authorization — 設定入口と変更権限

認証済み利用者は設定入口へアクセスできる。各カード・遷移先・操作は対象権限と医院scopeで制約する。閲覧できることを編集可と解釈しない。価格・割引・保険・支払方法は会計第六群の管理者と権限付きスタッフという既決に従う。

決定状態・根拠：既存仕様・会計合意継承。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

<a id="requirement-fr-clinic-settings-clinic-retirement"></a>
### FR-clinic-settings-clinic-retirement — 医院無効化と履歴保持

医院無効化は現在の利用対象から除くが、過去の診療・会計・職員履歴の一括削除を意味しない。既存関連データがある医院の物理削除は拒否する。無効化した医院に対する現在のAPI権限判定はC04・auth正本に従う。未終了業務を全件解消するまで無効化を一律禁止する新ルールは今回追加しない。発行済み帳票と設定変更の関係はD10の確定snapshot正本へ委ねる。

決定状態・根拠：既存医院設定・auth・帳票ownerの契約継承。追加の無効化前一括解消ルールは対象外。上記の該当既存正本に基づく整理で、新たな承認を捏造しない。

## 受入条件

全条件は文書上の期待結果。実機未実施。技術方式の提案は実装設計として区別する。

| 条件ID | Given（前提） | When（操作） | Then（期待結果） | 独立確認 | 成功条件 | 確認手順／証跡 |
|:---|:---|:---|:---|:---|:---|:---|
| AC-clinic-settings-01 | 一般職員にA院のみ所属 | scope=allで一覧を見る | 割当外B院を取得できない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-clinic-settings-clinic-directory](#requirement-fr-clinic-settings-clinic-directory)。実機未実施 |
| AC-clinic-settings-02 | 設定トップ閲覧可・対象editなし | 対象設定を直接更新 | 更新拒否、画面表示だけで認可を代替しない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-clinic-settings-portal-authorization](#requirement-fr-clinic-settings-portal-authorization)。実機未実施 |
| AC-clinic-settings-03 | 参照診療データが存在する | 医院を物理削除 | 削除を拒否して既存記録を保持 | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-clinic-settings-clinic-retirement](#requirement-fr-clinic-settings-clinic-retirement)。実機未実施 |
| AC-clinic-settings-04 | 医院が無効化され現在の有効集合から外れた | 旧医院を選択した状態で操作 | 現在の医院authorityを再評価して拒否・復旧し、過去履歴を一括削除しない | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-clinic-settings-clinic-retirement](#requirement-fr-clinic-settings-clinic-retirement)。実機未実施 |
| AC-clinic-settings-05 | 法人登録番号を変更する | 保存前に影響確認 | 全医院共通設定であることが分かる（表示方法はUI設計） | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-clinic-settings-clinic-directory](#requirement-fr-clinic-settings-clinic-directory)。実機未実施 |
| AC-clinic-settings-06 | A院の設定を編集中に選択医院が変わった | 旧入力を保存しようとする | 対象医院を再確認しB院へA院の入力を自動転用しない（C04接続） | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-clinic-settings-portal-authorization](#requirement-fr-clinic-settings-portal-authorization)。実機未実施 |
| AC-clinic-settings-07 | 医院固有番号と法人共通番号が異なる | 一方を編集 | 別エンティティを意図せず上書きせず、帳票の採用値はD10正本に従う | 単独ケースで確認 | Thenの結果と失敗時の不変 | [FR-clinic-settings-clinic-directory](#requirement-fr-clinic-settings-clinic-directory)。実機未実施 |

## 未決・責務境界

新しい個別質問を増やさず既存の現在医院認可・履歴保持へ整合する。医院全体無効化前に全未終了業務の解消を強制する新しい規則は今回追加しない。法人情報の発行済帳票への扱いはD10帳票snapshot正本へ委ねる。scope=allの旧記述は既存auth正本との整合修正。

## レビュー状態

基準版の本文と関連する静的根拠を照合。独立レビュー・統合リンク検査は統合担当で実施する。仕様記述を実装完了、受入PASS、実環境変更承認へ変換しない。

## 操作・状態・例外の全域点検

- 医院一覧／詳細／作成／部分更新／削除、法人番号の取得／更新を区別する。医院と法人の登録番号を同一値として強制しない。
- 有効／無効は現在のアクセス集合へ影響する。無効化前の将来予約等の全件解消を一律必須にする新規規則は今回追加せず、既決の職員指名停止ガードを医院全体の新たな強制条件へ自動拡張しない。履歴保持と利用停止の既存認可は維持する。
- 全院閲覧、選択医院編集、法人共通編集はscopeを分ける。画面を開いた時の古い権限を保存許可にしない。
- 帳票表示セクション・順序・フッター・ロゴ表示は設定範囲。ロゴ画像アップロードは本画面の既存機能に含めない。発行済み帳票の再現はD10正本へ接続する。
- 設定ポータルはカードの権限表示・セクション非表示・遷移先のガードを持つ。直接URL/APIでも認可を省略しない。
- 保存中の医院切替・権限失効はC04へ接続し、対象を再確認する。未保存入力の確認UIはC01、保存競合はC05へ委ねる。移行や医院設定の実変更は本仕様整理の完了条件外。

AC05の根拠は[法人情報セクション](../../screens/19-clinic-settings.md#13-法人情報セクション-companyinvoicesection)の全医院へ反映する既存設定単位。独立した承認段階や追加ロールは作らず、影響の表示方法はUI設計であり未決業務の追加ではない。

<a id="decision-alignment-20261017"></a>
## 2026-10-17確定仕様との接続

[決定根拠・適用範囲](../SPECIFICATION_PROGRESS.md#confirmed-business-decisions-20261017)：ユーザーの2026-10-16〜17 JSTの直接決定、文書更新承認は10月17日03:20 JST。以下は目標仕様・未実行の受入条件であり、実装・実機合格を示さない。

[業務正本・受入条件](../owners-pets/lifecycle.md#common-owner-pet-20261017)。氏名または動物種＋ペット名の候補をスタッフ確認し、同一飼主は警告して新規停止、別人は可、新ペットは既存飼主へ追加する。他院既存患者の自院利用と共通最新基本情報を採用し、院別診療・金銭記録と認可は維持する。旧他院行／ID一律拒否を自院利用へ機械適用しない。共有ID・既存重複移行・link/unlink方式は未選定で、重複作成を自院利用条件にしない。

<a id="clinic-hours-common-basis-20261020"></a>
## 営業時間・休診日とLINE受付範囲（2026-10-20本人回答）

決定者はユーザー。質問 Sentinel_07eb2535e184819190e1145f884b69cc に対する本人 Sentinel_6498c7a23a8c8191b672ba0dea52c1c3 の2026-10-20回答「4 ok」により採用。営業時間・休診日は医院設定を共通の基準とし、LINEの受付可能時間をその範囲内で別設定する。院内では従来どおり警告確認で時間外登録でき、LINEへその例外を設けない。レジの営業日区分は別扱いとし、医院営業時間／LINE受付時間へ統一しない。

[LINEの分類別・担当者別開始時刻](../../line/reservation-spec.md#requirement-fr-line-reservations-business-start-times)と受付期間・勤務／休憩・能力・ホテル入退室の既決条件を維持する。共通基準の採用を、LINE個別設定や既存保存値の自動上書きにしない。設定の物理保存先・競合検出・既存値移行は技術設計へ分離し、未照合設定を現在有効な受付範囲として補完しない。

既存FR-clinic-settings-clinic-directoryとLINE受付条件の追加確認として、医院時間内のLINE別設定／時間外LINE拒否／院内警告了承／取消、レジの別日界を対照確認する。実装・実機受入は未実施。
