# 飼主・ペット台帳の業務仕様

| 項目 | 内容 |
|:---|:---|
| 文書ID・種別 | D04 / 業務ライフサイクル・受入条件 |
| 決定状態 | 既存契約を継承し、追加合意4・5、死亡状態を保つ後日記録等、および[台帳・会員区分・LINE希望の共同確認](#agreement-owner-membership-dm-20261015)を区別して反映。対象外を一括承認しない |
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

<a id="agreement-owner-contacts-transfer-20261015"></a>
## 2026-10-15追加合意の範囲

決定者はユーザー。6点の確認（Sentinel_4230d91e7dd88191b67f30714ed073ba）への「ok」（Sentinel_ae272463dc248191b51fbef540e04109）により、4「電話・メール重複の警告と既存飼主候補を示し、別飼主登録も許可」と5「今後の予約・LINE通知先への影響一覧を確認して主飼主を切替、過去請求・預り金は元飼主に残す」を承認した。業務上の目標であり、旧409拒否や影響一覧の現行実装が変更済みという意味ではない。

重複判定の閾値・新たな正規化方式・確認理由やチェック欄等の追加必須条件は選定していない。飼主の自動統合、過去履歴移管、LINE未連携先の自動新規紐付け、送信済み／結果不明通知の自動再送を承認したものではない。

<a id="agreement-owner-membership-dm-20261015"></a>
## 飼主・ペット台帳の共同確認（2026-10-15）

質問Sentinel_4c24df2a3a0081918a0aabf8fdb9d7d3と回答Sentinel_56f0796762c0819185c612e7d17a3a89「ok」、補足Sentinel_1291b2ee79588191966adb6a7cc91dbbにより、主飼主1人／副飼主複数、副飼主追加で請求先・LINE通知先を自動変更しないこと、飼主とペットの注意事項の分離・ペット高中低／高は理由必須／スタッフ限定、全医院共通の動物種と追加・名称変更・無効化のシステム管理者限定を確認した。

会員区分は既存ラベルを保持。「退亡者」は連絡なく来院しなくなった人であり死亡を意味しない。「他診/準」は意味を未確定のままラベルを使い、意味や連動動作を新設しない。会員区分変更で料金・予約受付・LINE配信可否を自動変更しない。直接回答Sentinel_14eb5ae0bbd88191b9c5cecc3e73403b「そうだね、自動変更しないね」もこの非連動を確認している。DMはLINEの希望であり、初回質問の郵送解釈は後の明示回答で置き換わった。

質問Sentinel_8889f8843dbc819191f388e76388a06fへの回答Sentinel_6725944026748191841188b59f9c0db1「ok」で、DM不要は案内・販促を止め、予約確認・変更・予約リマインドは既存安全条件の下で送信対象とする。さらに質問Sentinel_b4119035d6848191818e63ec13356558への回答Sentinel_0b29068a42688191be8516f3dd7e704c「yes」で、DM未設定も案内・販促を送らず、配信希望を確認して必要へ変更する。予約通知の扱いは同じ。送信目的と停止・本人／宛先条件の判定は[D15正本](../../line/lstep-integration.md#crm-dm-purpose-contract)へ追跡する。

以上は今回確認した業務範囲であり、全既存台帳条件・全過去会話／Issueの承認確認完了を意味しない。画面配置・数値制限・削除／並替等の既存条件は継承として残す。物理API・実DB・外部配信の適合は未検証。

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

<a id="requirement-fr-owners-pets-contact-duplicate-warning"></a>
### FR-owners-pets-contact-duplicate-warning

電話・メールが同医院の既存飼主と重複する登録では、重複を警告して既存飼主の候補を表示し、家族等で連絡先を共用していても別の飼主として登録できる。候補は照合の参考であり、同一人物／同一ownerと自動判定しない。候補を示しただけで入力先ownerを置換したり、既存owner・pet・identity group・LINE紐付けを統合したりしない。候補参照は既存の医院scope・閲覧権限内に限定する。

連絡先の重複だけを登録拒否の理由にしない。他の必須項目・形式・医院整合・認可の検証と、新規飼主＋複数petの同時成功は維持する。別登録は独立owner IDを持ち、連絡先一致を同一操作の保存receiptに使わない。検出方式の変更や追加必須確認を本合意から推定しない。旧仕様の重複409拒否は[現行差](../../../design/domains/owners-pets/technical-spec.md#owner-contact-duplicate-target)へ分離する。

<a id="requirement-fr-owners-pets-identity-and-search"></a>
### FR-owners-pets-identity-and-search

飼主Noはowners.idの表示別名、ペット番号はpets.pet_number。検索は名前・カナ・電話・番号を扱い、複数語AND／各語内の既存項目OR、空白のみ0件、種別は専用filterを使う。20件ページのサーバ絞込みとURL状態を正本とし、ページ内ソート／全件取得を再導入しない。健診履歴の判定はペット自身の受診履歴とJST暦日境界に従う。詳細は[一覧](../../screens/03-owners-list.md)。

<a id="requirement-fr-owners-pets-primary-secondary"></a>
### FR-owners-pets-primary-secondary

ペットの主飼主は一人、副飼主は複数の別リンクとして登録できる。副飼主追加で請求先・LINE通知先を自動変更せず、割引判定やLINE認証の主飼主も置き換えない。続柄trim後1〜50文字、同じ飼主の重複・主副の重複・他医院飼主は拒否する。副飼主置換は取得versionと完全な希望一覧を送る。競合時は再取得して再判断し、空一覧への自動置換をしない。

<a id="requirement-fr-owners-pets-primary-transfer"></a>
### FR-owners-pets-primary-transfer

主飼主変更は同医院の明示選択で行い、変更確定前に、今後の予約とLINE通知先への影響を一覧で確認してから新飼主へ切り替える。予約の対象・変更前後の飼主・LINE通知先がどう変わるか、付替えできず連携解除となる場合を判別できるようにする。予約がない場合は0件を示し、影響取得失敗を0件扱いしない。会員区分／値引率が同じ場合にも予約・LINEの影響確認を省略しない。会員区分の差だけで価格を変更せず、独立した値引率等による実際の会計金額への影響警告は維持する。現副飼主を主飼主へ変更する場合、先に副飼主リンクを解除する既存ガードを維持する。

確認の取消では主飼主・予約・LINE紐付けを切り替えない。確認後に対象予約・帰属・LINE対応・必要な版／権限が変わった場合は、古い一覧への確認だけで変更せず再取得・差分確認する。対象表示のための新しい公開範囲を設けず、既存の認可・医院scopeを維持する。確認と保存の安全な接続は[技術契約](../../../design/domains/owners-pets/technical-spec.md#owner-transfer-target)へ。

予約への既存移管は[技術契約](../../../design/domains/owners-pets/technical-spec.md#owner-transfer-current)の同医院・同pet・JST当日以降pending/confirmedに限定する。過去予約、来院受付後、tentativeやカルテ付き予約等を「今後の予約」の語だけで対象へ広げない。影響一覧にはLINEの付替え／解除／変更なしを区別して示し、新飼主に確認済みLINE IDと同医院の一意なcustomer対応がない場合は通知可能と表示せず、自動新規紐付けをしない。元からLINE customerがないmanual予約を、新飼主のLINEへ自動接続しない。

過去の請求と預り金は、成功した主飼主変更後も元のownerに残し、金額・残高・履歴を自動移管しない。過去カルテや別医院データの付替えも許可しない。LINEの配信除外・本人確認・停止gate・通知結果不明の保留は各正本を維持し、主飼主変更から解除／再開／再送を起動しない。

<a id="requirement-fr-owners-pets-safety-and-private-notes"></a>
### FR-owners-pets-safety-and-private-notes

死亡状態と安全特記は別概念。台帳の死亡日・理由を記録しても既存記録を消さず、各機能の死亡ゲートへ接続する。生死の通常属性PATCHによる変更は禁止された現行writer境界を維持し、死亡／復活の専用処理を使う（[技術契約](../../../design/domains/owners-pets/technical-spec.md)）。2026-10-15の追加確認（質問Sentinel_2b63d8f5aad88191aa2b4dff79fb5ac2、回答Sentinel_45be2cab0fe4819188c7c0f9a211358c「ok」）により、生前の診療の後日入力・誤記訂正・退院処理・精算は、死亡登録を取り消さず理由と履歴を残す専用経路で可能にする。具体的な認可・対象記録・保存と監査は[共通臨床安全](../../shared/clinical-safety.md)と各臨床／会計正本へ従う。これは新規未来予約・新たな施術／医療行為を許可するものではなく、generic status PATCHや復活処理による迂回を行わない。

飼主への対応上の注意is_dangerousと、ペットの咬みつき等の安全注意danger_level/danger_reasonは別物で、それぞれスタッフ内部情報として扱う。高のみ理由必須、中は任意、低／未設定のバッジ非表示。スタッフ向け中立表示とpopoverを使い、スタッフ専用Owner Reportの限定DTO、およびLIFF・line-reserveの飼主向け応答にペット特記フィールドを含めない。詳細・アクセシビリティ条件はフォーム正本へ。

<a id="requirement-fr-owners-pets-delete-dependencies"></a>
### FR-owners-pets-delete-dependencies

飼主／ペットの論理削除は参照保護を満たす場合だけ行う。飼主に主飼主ペットまたは副飼主リンクが残る場合は拒否。ペットにカルテが紐付く場合は拒否する既存ガードを維持する。エラーメッセージの「先に削除」を、保存すべき臨床記録の削除指示とみなさない。カルテの入力済み・確定済み保護はD03が優先し、削除のための依存切断・ID付替え・履歴消去を新たに許可しない。

自動削除をしない当面方針と、将来の保存期間／正式削除手続きの未決は[共通方針](../../shared/README.md#interim-retention-policy)に分離する。

<a id="requirement-fr-owners-pets-species-master"></a>
### FR-owners-pets-species-master

動物種は全医院共通の有効／無効・表示順付きマスタ。追加・名称変更・無効化はシステム管理者だけが行う。使用中の動物種削除は409で拒否し、新規の選択では無効を除外する。犬／猫名称の品種サジェストは現行仕様であり、動物種マスタ変更がワクチン対象種・投薬量enumを自動変更するものではない。master変更の必要監査はsame-txで失敗時rollback。

<a id="requirement-fr-owners-pets-membership-dm"></a>
### FR-owners-pets-membership-dm

会員区分は非会員・会員・退亡者・他診/準のラベルを保持し、退亡者をペットや飼主の死亡記録へ対応付けない。他診/準の意味は未確定で、新しい価格・受付・配信の動作を与えない。区分変更だけで料金・予約受付・LINE配信可否を自動変更しない。個別値引率・予約条件・本人／宛先や独立した配信停止は別の契約として保持する。

DMはLINEの配信希望で、未設定・必要・不要を区別する。未設定／不要では案内・販促を送らず、未設定から必要へは希望を確認して明示変更する。既存未設定値を必要へ自動補完しない。予約確認・変更・予約リマインドは未設定／不要だけでは除外せず、[D15目的別条件](../../line/lstep-integration.md#crm-dm-purpose-contract)と既存のアカウント・ブロック・本人／宛先等の安全条件へ渡す。DM値変更自体は送信・過去通知の再送・配信停止の解除・LINE紐付け変更を起動しない。

## 異常系・保存・取消

| 状況 | 検知／表示と次の操作 | 境界 |
|:---|:---|:---|
| 同医院のメール／電話重複 | 警告と既存飼主候補を示し、別飼主として登録可能 | 連絡先一致だけでは拒否・統合しない。旧BE409は現行実装差 |
| 副飼主version競合 | 再読込して現在一覧と照合 | 他者の追加を古いPUTで消さない |
| 主飼主移管の予約競合／既存カルテ | 全変更失敗、理由を確認 | 一部だけ移管しない。既存履歴を削除して再実行しない |
| 主飼主の影響一覧取得失敗／確認後変化 | 未取得・0件・変更ありを分離し、再取得・差分確認 | 不明な影響や古い確認で切替成功にしない。期待版だけ更新して無警告再送しない |
| 新規保存の通信断・結果不明 | 入力を保持して既存登録結果を照合 | timeoutだけで未登録と判定して別IDを作らない。結果照会・操作IDの実装差は技術契約 |
| 詳細／マスタ取得失敗 | 未取得と0件を分離 | 空一覧や未設定値で上書きしない |
| 医院／患者切替後の遅延応答 | 元contextへ隔離して破棄 | 新医院フォームへ混ぜない |
| 未保存編集で離脱 | 確認後に破棄、取消なら編集維持 | 永続済み副飼主／LINE操作をブラウザ取消で巻戻さない |

<a id="owners-pets-acceptance"></a>
## 受入条件

以下は継承仕様と技術的安全条件、追加合意4・5、および上記の台帳・会員区分・LINE希望の共同確認を具体化したGWT。死亡後の後日記録等の受入条件は共通臨床安全と各臨床／会計正本へ置く。実行証跡は未取得。追加合意の対象外を新たに承認したものではない。

| 条件ID | 前提／操作 | 期待結果 | 確認手順／証跡 |
|:---|:---|:---|:---|
| AC-owners-pets-register-atomic | 新規飼主と複数ペットを送信し、2頭目の検証／書込／再取得を失敗させる | 飼主も全ペットも未確定。一部成功を表示しない | 同時登録の失敗点別確認。未実施 |
| AC-owners-pets-scope | 他医院のowner／保険／pet IDを登録・更新へ指定する | 拒否して現医院の台帳は不変 | 所属違反を個別確認。未実施 |
| AC-owners-pets-search | 空白だけ、複数語、カナ、飼主No、種名を入力する | 空白のみ0件、複数語AND、カナ正規化、No照合。種名は専用filter | 一覧の検索条件と20件境界。未実施 |
| AC-owners-pets-checkup-boundary | JST当日からN年前ちょうどの健診と別ペットの健診がある | 境界日は含め、別ペット履歴を混入しない | 1／2／3年・なしの各filter。未実施 |
| AC-owners-pets-read-only | owners:viewのみでpet深いリンクを開く | 詳細閲覧可、入力・保存不可。query除去後に同リンク再利用可 | 所属医院の一覧と直接URL。未実施 |
| AC-owners-pets-cross-clinic-list | 複数所属医院を選択する | 他医院行の操作リンクを抑止、サーバ所属検証 | URL改変／権限なし医院。未実施 |
| AC-owners-pets-secondary-replace | 副飼主を有効versionで追加・削除する | 複数副飼主の全置換が成功、主飼主・請求先・LINE通知先と割引判定は不変 | 空配列による明示解除も確認。未実施 |
| AC-owners-pets-secondary-conflict | 古いversion、重複owner、主飼主、他医院を送る | 競合／入力拒否、既存リンク保持 | 競合後の再取得を確認。未実施 |
| AC-owners-pets-owner-transfer-confirm | 変更先の会員区分／値引率が同じ／異なる場合と予約0件の場合に影響一覧を確認し、取消す | 予約・LINE影響確認を省略せず、独立した値引率等の実際の会計影響警告も維持。会員区分差だけで価格を変えない。取消では主飼主・予約・LINE紐付け不変 | [FR](#requirement-fr-owners-pets-primary-transfer)。確認・取消・0件と取得失敗を分離。未実施 |
| AC-owners-pets-owner-transfer-atomic | 同日以降pending/confirmed予約移管で監査または対象照合が失敗 | petと予約移管をともにrollback、過去請求・預り金不変 | 現行移管の範囲・失敗点別。未実施 |
| AC-owners-pets-danger-privacy | 高／中／低／未設定を表示し飼主向け応答を読む | 飼主注意とペット注意を分離しスタッフ限定。ペット高のみ理由必須、所定の中立表示。飼主向け応答に特記フィールドなし | aria/title／JSONまで確認。未実施 |
| AC-owners-pets-delete-guard | 主飼主ペット／副飼主リンクを持つ飼主、カルテ付きペットを削除する | 削除拒否、既存記録保持。カルテ削除への自動連鎖なし | 依存種別ごとに確認。未実施 |
| AC-owners-pets-species-authority | 一般clinic管理者が共通動物種を変更する | 拒否。システム管理者操作だけ許可 | create/update/delete/reorder別確認。未実施 |
| AC-owners-pets-species-reference-audit | 使用中種を削除、またはmaster監査を失敗させる | 参照中削除拒否／監査失敗時変更なし | tx内監査の失敗注入。未実施 |
| AC-owners-pets-optional-clear | 名前由来／出逢いストーリーで省略・null・空白を保存する | 省略は保持、null／空白はNULL、長さ制約維持 | 500／2000文字の境界。未実施 |
| AC-owners-pets-unknown-result | 保存応答を失い再読込する | 結果不明を失敗と断定せず照合、重複createなし | 遅延成功／失敗の両ケース。未実施 |
| AC-owners-pets-stale-update | Given 表示後に他操作が同じ記録を更新、When 元表示版で保存、Then | 現在の原記録は不変、入力保持、競合を通知し再取得・差分確認する。新版への無警告再送不可 | 結果不明は同一操作照合、別create不可。未実施 |
| AC-owners-pets-contact-duplicate-warning | 電話のみ／メールのみ／両方が同医院の既存飼主と重複する登録入力をする | 重複警告と権限内の既存飼主候補を提示。既存ownerへ自動置換・統合しない | [FR](#requirement-fr-owners-pets-contact-duplicate-warning)。他医院・閲覧権限外の候補を混入しない。未実施 |
| AC-owners-pets-contact-duplicate-separate | 共用連絡先の警告・候補を表示後、別飼主として登録する | 他の検証を満たせば独立owner IDで保存でき、既存owner・pet・LINE紐付けは不変。同時pet登録の原子性を保持 | [FR](#requirement-fr-owners-pets-contact-duplicate-warning)。電話／メールをreceiptにしない。未実施 |
| AC-owners-pets-owner-transfer-impact | 移管対象予約とLINE通知先への影響があるpetで主飼主変更を進める | 切替前に対象予約と変更前後の帰属・LINE付替え／解除／変更なしを一覧表示し、確認後だけ保存する | [FR](#requirement-fr-owners-pets-primary-transfer)。表示対象と保存対象を照合。未実施 |
| AC-owners-pets-owner-transfer-impact-stale | 影響一覧を確認後、対象予約・owner帰属・LINE対応・必要な版／権限が変化する | 古い確認だけで切替せず、現在状態を再取得し差分確認。部分移管なし | [FR](#requirement-fr-owners-pets-primary-transfer)。保存時再検証と認可失効を確認。未実施 |
| AC-owners-pets-owner-transfer-line-boundary | 元LINE予約／manual予約と、新飼主LINEの確認済一意対応あり／なしを組み合わせる | 元LINE予約だけ既存条件で付替え又は解除。未連携への自動新規紐付けなし。manualはNULLと既存customer_fields保持。通知停止・除外・結果不明保留を解除しない | [FR](#requirement-fr-owners-pets-primary-transfer)。通知到達を台帳保存成功と同一視しない。未実施 |
| AC-owners-pets-owner-transfer-history | 過去請求・預り金残高・過去予約があるpetの主飼主変更を成功させる | 対象未来予約以外の履歴と、元ownerの過去請求・預り金帰属／金額／残高を保持 | [FR](#requirement-fr-owners-pets-primary-transfer)。成功前後のID・金額・履歴比較。未実施 |
| AC-owners-pets-membership-labels | 退亡者／他診/準を保存・表示する | 既存ラベル保持。退亡者を死亡扱いせず、他診/準へ未確認の意味・動作を加えない | FR-owners-pets-membership-dm。台帳状態との分離。未実施 |
| AC-owners-pets-membership-independent | 会員区分のみ変更する | 料金・予約受付・LINE配信可否を自動変更しない。独立した値引率／停止／本人条件は保持 | 同FR。変更前後の各ownerを比較。未実施 |
| AC-owners-pets-dm-line-preference | DMを未設定／必要／不要で保存する | LINE希望として区別。郵送設定へ置換せず、値変更だけで送信・再送・停止解除・宛先変更しない | 同FR・D15目的別条件。未実施 |
| AC-owners-pets-dm-unset-confirmation | DM未設定の飼主の配信希望を確認し必要へ明示変更する | 確認前は案内・販促対象外。自動補完なし。変更後も送信時の既存安全条件と目的を検証 | 同FR・D15。予約通知例外は維持。未実施 |

## 未決事項・関連仕様・根拠

- 今回の台帳・会員区分・LINE希望の確認範囲は確定。他診/準の意味は未確定だがラベル維持・自動連動なしとして使うため、今回の保存／配信条件を未定に戻さない。全業務の共同確認は継続中。
- 追加合意4・5は本文の範囲で確定。重複判定の閾値／新たな正規化／追加必須確認、影響確認の物理API／payloadは未選定。一般更新の全項目履歴・競合version・再送receipt・参照制約の全writer適合は技術検証の残件であり実装済みとしない。
- 主飼主変更とLINE予約紐付け／過去履歴・預り金の整合はD01/D10/D18/D20とのX04/X06照合対象。identity groupのlink/unlinkはD19へ、飼主レポートの各section公開はD21へ、配信停止／タグ同期はD15へ残す。
- 死亡／復活の外部配信連携はD15の専用経路。汎用ペット更新からstatusを書けない既存制約を維持する。
- 原文詳細は冒頭の3正本。ソース根拠・現在の限界は技術契約。確認対象refはPR670 / 41fcf3eb0a8befb7c0b5f613d1df749a60e30be8。

