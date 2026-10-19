# D17 医院・設定入口

> 2026-10-15の仕様整理候補。PR #670の基準版から既存合意・既存仕様・未承認案を分離した。業務合意と実装／実機検証を別管理する。D22/D23は2026-10-15の一括承認を反映。技術設計案は業務承認と分ける。

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

医院ごとの基本情報・連絡先・有効状態・帳票設定と、法人共通のインボイス登録番号を分ける。法人共通設定の影響が全医院へ及ぶことを編集時に示す。scope=allは非system_adminへ割当外医院を開示する根拠にしない。最新auth §4.10を優先する。

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
