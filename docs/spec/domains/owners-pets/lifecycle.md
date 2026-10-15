# 飼主・ペット台帳の業務仕様

| 項目 | 内容 |
|:---|:---|
| 文書ID・種別 | D04 / 業務ライフサイクル・受入条件 |
| 決定状態 | 既存記載を継承（決定根拠は個別参照。新しい一括業務承認を意味しない） |
| 実装状態 | 既存記載を継承。2026-10-15に下記限定ソースを静的照合、実装追従の全件再監査なし |
| 実機検証状態 | 未実施 |
| 文書確認日 | 2026-10-15 |

## 目的・対象範囲

医院の飼主と患者を正確に識別し、基本台帳の登録・変更・副飼主・検索・参照保護を扱う。画面は飼主一覧／新規／詳細、マスタは動物種類の計4製品ルート。具体的入力項目・UI・API一覧の正本は[一覧](../../screens/03-owners-list.md)・[フォーム](../../screens/04-owners-form.md)・[動物種類](../../screens/settings/master-animal-species.md)に保持する。本書は画面をまたぐ業務契約とACの正本であり、画面本文の複製ではない。

- 台帳を登録し診療等で同じ患者を選ぶ。主飼主・副飼主・既存scope内の同一個体連携を混同しない。
- 飼主変更時も過去の診療・請求・預り金を別の飼主へ自動移管しない。現在台帳の変更と個々の履歴の帰属を分ける。
- スタッフ安全情報を業務画面で確認し、飼主向け応答へ漏らさない。
- 優先度の新たな採番は行わず、臨床安全・参照保護は[共通原則](../../specification.md)に従う。

語彙・関係は[台帳オントロジー](../../ontology/owners-pets.md)、保存・再送・競合は[技術契約](../../../design/domains/owners-pets/technical-spec.md)。

## 担当者・権限・前提

飼主の閲覧は owners:view、登録は create、変更は edit、削除は delete。ペット詳細は閲覧だけで開けるが保存できない。副飼主の読取／変更も owners:view／edit。動物種は全医院共有であり、変更／削除／並替はシステム管理者のみ。clinic側master権限から全体変更を許可しない。

複数所属の医院横断一覧は所属検証済医院に限定。他医院行は閲覧データのみでリンク・編集・レポート・削除を提供しない。医院所属・IDの照合はサーバで行い、氏名一致や検索結果の近似からIDを決めない。

## 契機・入口・出口

一覧検索→同じowner/pet IDの詳細参照→新規／変更保存が基本経路。初回飼主と複数ペットの同時登録は全体成功または全体失敗。既存飼主フォームの保存、副飼主全置換、個別ペット更新、LINE操作は別の操作単位であり、画面全体を一括transactionとみなさない。

<a id="owners-pets-rules"></a>
## 状態遷移・引継ぎデータ

<a id="requirement-fr-owners-pets-register-scope"></a>
### FR-owners-pets-register-scope

飼主は医院スコープで登録し、同時登録する全ペットは同じ医院・作成された飼主へ結び付ける。途中でペット・マスタ検証・再取得に失敗した場合、飼主だけ／一部ペットだけを登録済みにしない。根拠は[既存同時登録契約](../../../architecture/cross-domain-orchestration-catalog.md) PATH-OWNER-PET-REGISTER。これは業務atomicityの継承であり全操作の監査必須化を新設しない。

<a id="requirement-fr-owners-pets-identity-and-search"></a>
### FR-owners-pets-identity-and-search

飼主Noはowners.idの表示別名、ペット番号はpets.pet_number。検索は名前・カナ・電話・番号を扱い、複数語AND／各語内の既存項目OR、空白のみ0件、種別は専用filterを使う。20件ページのサーバ絞込みとURL状態を正本とし、ページ内ソート／全件取得を再導入しない。健診履歴の判定はペット自身の受診履歴とJST暦日境界に従う。詳細は[一覧](../../screens/03-owners-list.md)。

<a id="requirement-fr-owners-pets-primary-secondary"></a>
### FR-owners-pets-primary-secondary

ペットの主飼主は一人、副飼主は別リンク。副飼主追加で請求宛先・割引・LINE認証の主飼主を置き換えない。続柄trim後1〜50文字、同じ飼主の重複・主副の重複・他医院飼主は拒否する。副飼主置換は取得versionと完全な希望一覧を送る。競合時は再取得して再判断し、空一覧への自動置換をしない。

<a id="requirement-fr-owners-pets-primary-transfer"></a>
### FR-owners-pets-primary-transfer

主飼主変更は同医院の明示選択で行い、会員区分／値引率が異なる場合は会計への影響確認を挟む。現副飼主を主飼主へ変更する場合、先に副飼主リンクを解除する既存ガードを維持する。

予約への既存移管は[技術契約](../../../design/domains/owners-pets/technical-spec.md#owner-transfer-current)に限定される。台帳変更から過去カルテ・確定済会計・預り金・別医院データの移管許可を推定しない。特にD10の預り金は元飼主残高を保ち、主飼主変更に伴う自動譲渡をしない。

<a id="requirement-fr-owners-pets-safety-and-private-notes"></a>
### FR-owners-pets-safety-and-private-notes

死亡状態と安全特記は別概念。台帳の死亡日・理由を記録しても既存記録を消さず、各機能の死亡ゲートへ接続する。生死の通常属性PATCHによる変更は禁止された現行writer境界を維持し、死亡／復活の専用処理を使う（[技術契約](../../../design/domains/owners-pets/technical-spec.md)）。他ドメインの臨床操作可否を台帳側で新設しない。

飼主is_dangerousとペットdanger_level/danger_reasonは別物。高のみ理由必須、中は任意、低／未設定のバッジ非表示。スタッフ向け中立表示とpopoverを使い、スタッフ専用Owner Reportの限定DTO、およびLIFF・line-reserveの飼主向け応答にペット特記フィールドを含めない。詳細・アクセシビリティ条件はフォーム正本へ。

<a id="requirement-fr-owners-pets-delete-dependencies"></a>
### FR-owners-pets-delete-dependencies

飼主／ペットの論理削除は参照保護を満たす場合だけ行う。飼主に主飼主ペットまたは副飼主リンクが残る場合は拒否。ペットにカルテが紐付く場合は拒否する既存ガードを維持する。エラーメッセージの「先に削除」を、保存すべき臨床記録の削除指示とみなさない。カルテの入力済み・確定済み保護はD03が優先し、削除のための依存切断・ID付替え・履歴消去を新たに許可しない。

自動削除をしない当面方針と、将来の保存期間／正式削除手続きの未決は[共通方針](../../shared/README.md#interim-retention-policy)に分離する。

<a id="requirement-fr-owners-pets-species-master"></a>
### FR-owners-pets-species-master

動物種は全医院共通の有効／無効・表示順付きマスタ。使用中の動物種削除は409で拒否し、新規の選択では無効を除外する。犬／猫名称の品種サジェストは現行仕様であり、動物種マスタ変更がワクチン対象種・投薬量enumを自動変更するものではない。master変更の必要監査はsame-txで失敗時rollback。

## 異常系・保存・取消

| 状況 | 検知／表示と次の操作 | 境界 |
|:---|:---|:---|
| 同医院のメール／電話重複 | BE409を表示し内容を修正 | 勝手に既存飼主へ統合しない |
| 副飼主version競合 | 再読込して現在一覧と照合 | 他者の追加を古いPUTで消さない |
| 主飼主移管の予約競合／既存カルテ | 全変更失敗、理由を確認 | 一部だけ移管しない。既存履歴を削除して再実行しない |
| 新規保存の通信断・結果不明 | 入力を保持して既存登録結果を照合 | timeoutだけで未登録と判定して別IDを作らない。結果照会・操作IDの実装差は技術契約 |
| 詳細／マスタ取得失敗 | 未取得と0件を分離 | 空一覧や未設定値で上書きしない |
| 医院／患者切替後の遅延応答 | 元contextへ隔離して破棄 | 新医院フォームへ混ぜない |
| 未保存編集で離脱 | 確認後に破棄、取消なら編集維持 | 永続済み副飼主／LINE操作をブラウザ取消で巻戻さない |

<a id="owners-pets-acceptance"></a>
## 受入条件

以下は継承仕様と技術的安全条件を具体化したGWT。実行証跡は未取得。新しい業務承認ではない。

| 条件ID | 前提／操作 | 期待結果 | 確認手順／証跡 |
|:---|:---|:---|:---|
| AC-owners-pets-register-atomic | 新規飼主と複数ペットを送信し、2頭目の検証／書込／再取得を失敗させる | 飼主も全ペットも未確定。一部成功を表示しない | 同時登録の失敗点別確認。未実施 |
| AC-owners-pets-scope | 他医院のowner／保険／pet IDを登録・更新へ指定する | 拒否して現医院の台帳は不変 | 所属違反を個別確認。未実施 |
| AC-owners-pets-search | 空白だけ、複数語、カナ、飼主No、種名を入力する | 空白のみ0件、複数語AND、カナ正規化、No照合。種名は専用filter | 一覧の検索条件と20件境界。未実施 |
| AC-owners-pets-checkup-boundary | JST当日からN年前ちょうどの健診と別ペットの健診がある | 境界日は含め、別ペット履歴を混入しない | 1／2／3年・なしの各filter。未実施 |
| AC-owners-pets-read-only | owners:viewのみでpet深いリンクを開く | 詳細閲覧可、入力・保存不可。query除去後に同リンク再利用可 | 所属医院の一覧と直接URL。未実施 |
| AC-owners-pets-cross-clinic-list | 複数所属医院を選択する | 他医院行の操作リンクを抑止、サーバ所属検証 | URL改変／権限なし医院。未実施 |
| AC-owners-pets-secondary-replace | 副飼主を有効versionで追加・削除する | 全置換が成功、主飼主／割引／LINEの意味は不変 | 空配列による明示解除も確認。未実施 |
| AC-owners-pets-secondary-conflict | 古いversion、重複owner、主飼主、他医院を送る | 競合／入力拒否、既存リンク保持 | 競合後の再取得を確認。未実施 |
| AC-owners-pets-owner-transfer-confirm | 会員区分／値引率が違う飼主へ変更し、確認を取消す | 保存せず旧主飼主を保持 | 確認と取消を別々に確認。未実施 |
| AC-owners-pets-owner-transfer-atomic | 同日以降pending/confirmed予約移管で監査または対象照合が失敗 | petと予約移管をともにrollback、過去請求・預り金不変 | 現行移管の範囲・失敗点別。未実施 |
| AC-owners-pets-danger-privacy | 高／中／低／未設定を表示し飼主向け応答を読む | 高のみ理由必須、所定の中立表示。飼主向け応答に特記フィールドなし | aria/title／JSONまで確認。未実施 |
| AC-owners-pets-delete-guard | 主飼主ペット／副飼主リンクを持つ飼主、カルテ付きペットを削除する | 削除拒否、既存記録保持。カルテ削除への自動連鎖なし | 依存種別ごとに確認。未実施 |
| AC-owners-pets-species-authority | 一般clinic管理者が共通動物種を変更する | 拒否。システム管理者操作だけ許可 | create/update/delete/reorder別確認。未実施 |
| AC-owners-pets-species-reference-audit | 使用中種を削除、またはmaster監査を失敗させる | 参照中削除拒否／監査失敗時変更なし | tx内監査の失敗注入。未実施 |
| AC-owners-pets-optional-clear | 名前由来／出逢いストーリーで省略・null・空白を保存する | 省略は保持、null／空白はNULL、長さ制約維持 | 500／2000文字の境界。未実施 |
| AC-owners-pets-unknown-result | 保存応答を失い再読込する | 結果不明を失敗と断定せず照合、重複createなし | 遅延成功／失敗の両ケース。未実施 |
| AC-owners-pets-stale-update | Given 表示後に他操作が同じ記録を更新、When 元表示版で保存、Then | 現在の原記録は不変、入力保持、競合を通知し再取得・差分確認する。新版への無警告再送不可 | 結果不明は同一操作照合、別create不可。未実施 |

## 未決事項・関連仕様・根拠

- 既存仕様から新規の重大な台帳業務判断は追加抽出なし。一般更新の全項目履歴・競合version・再送receipt・参照制約の全writer適合は技術検証の残件であり実装済みとしない。
- 主飼主変更とLINE予約紐付け／過去履歴・預り金の整合はD01/D10/D18/D20とのX04/X06照合対象。identity groupのlink/unlinkはD19へ、飼主レポートの各section公開はD21へ、配信停止／タグ同期はD15へ残す。
- 死亡／復活の外部配信連携はD15の専用経路。汎用ペット更新からstatusを書けない既存制約を維持する。
- 原文詳細は冒頭の3正本。ソース根拠・現在の限界は技術契約。確認対象refはPR670 / 41fcf3eb0a8befb7c0b5f613d1df749a60e30be8。

