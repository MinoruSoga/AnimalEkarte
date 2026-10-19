# カルテの概念・関係

[2026-10-15最新合意](../domains/medical-records/lifecycle.md#medical-record-agreement-autosave-20261015)で通常カルテの医師確定ボタン／終了後通常編集禁止を置換し、自動保存は維持する。旧finalizedは保存された履歴属性と現実装の状態であり、現在の編集可否・スタッフ履歴掲載を単独で決めない。

2026-10-14初回5項目と入力4項目の部分合意。業務本文・承認根拠・受入条件は[保存正本](../domains/medical-records/lifecycle.md#medical-record-agreement-first-five)。予約語彙は[予約正本](reservations.md)を継承し、以下から専用記録や他領域の内部を確定しない。

| 用語 | 定義・区別 | 制約の所在 |
|:---|:---|:---|
| 通常カルテ | 診療の記録。予約、施術・入院・ホテル専用記録とは別 | [受付接続](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-reception-and-retrospective-entry) |
| 編集者 | 実際に入力・編集した権限あるスタッフ。担当医本人とは限らない | [編集ログ](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-authorized-editor-audit) |
| 確認医師 | 医院が認め、目的別の臨床内容・訂正・評価不能手入力を確認する医師。担当医以外も代行可能 | 同上。新しい臨床責任者役割を設けない |
| カルテ担当医／予約指名医師 | 記録の担当と予約の指名。編集者・確認医師と自動的に同一にしない | [予約概念](reservations.md) |
| 作成中／確定済み | draft／finalizedは旧実装の状態。最新目標は終了後も通常編集でき、以前の保存版・確定履歴・追記を保持する | [通常編集](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-normal-edit-after-end) |
| 治療プラン未完了警告 | 未実施・後日予定が残ることを診療終了時に知らせ、確認後は終了可能 | 自動保存時には出さず、終了を実施済みへ変換しない。4状態と現在値の取得／版を照合 |
| 空の誤作成 | 臨床入力がなく、削除を妨げる請求等関連がない誤作成。draftだけでは空とは言わない | [削除・訂正](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-empty-delete-reviewed-correction) |
| 記録確認訂正／訂正追記 | 患者・予約日等の目的別訂正確認と、元記載に対応する既存追記履歴。通常本文更新を追記だけに限定しない | 同上。必要内容転記・元／先対応と以前の保存版を保持 |
| 過去未登録診療の追加入力 | 過去の診療で未登録だった記録を理由・医師確認付きで専用入力する業務。確定記録の追記とは別 | [過去入力](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-reception-and-retrospective-entry) |
| 薬量評価不能／理由付き手入力 | 必要情報不足で計算・安全評価ができない状態と、医師確認・理由を記録して入力する業務 | [薬量境界](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-unevaluable-dose-manual)。上限超過解除とは別 |

関係：一つの通常カルテは複数回編集され、それぞれの編集者と履歴を持つ。担当医と別の医師による確認も記録する。既存の訂正追記と以前の保存版を保持したまま、現在の通常カルテは終了後も編集できる。過去版の不変性と現在版の更新可能性を区別する。通常受付の同一予約・active記録の一意性は既存予約契約、過去入力の同一性・重複判定は今後の専用設計へ分ける。治療プラン・子入力・請求の依存は単なる本文空欄判定から除外しない。

## 入力4項目の概念

業務正本は[追加合意](../domains/medical-records/pages/medical-record-form.md#medical-record-agreement-input-four)。

| 用語 | 定義・関係 | 制約 |
|:---|:---|:---|
| 選択複写 | 過去問診／治療明細の選択対象を今回記録へ適用する。原記録は不変 | 問診上書きは確認、新明細は今回価格、無効／不一致は通知再選択。逐次保存の部分成功を維持 |
| Template挿入 | 登録済み定型文を手動選択して問診の未保存入力に入れる | 基本追記、置換のみ確認。自動提案／元template更新なし |
| 当時内容／今回価格 | 既存記録が保持する名称・単位・価格と、新しい複写明細が使う今回の価格 | master改定／無効化で既存値を変更しない。根拠のない旧値補完をしない |
| 治療プラン項目状態 | 未実施・実施済み・後日予定・中止の4状態 | 未実施／後日予定を終了時警告、中止は理由。警告確認後終了可、自動保存時警告なし |
| 実施記録者／記録時刻 | 実施済み等を記録した操作主体とその記録時刻 | 実際の実施者／実績時刻と区別し、未指定項目を必須化しない |

項目ごとの状態と記録履歴を持ち、checkbox解除を実施取消へ変換しない。理由付き実施取消は追加合意の専用手順へ。実施・準備完了の請求対象時点とプラン同期の業務境界は[D10最新合意](../domains/accounting/lifecycle.md#doctor-review-sync-billing-20261017)で確定し、各writerの具体対応は設計。入院snapshotへ通常カルテの状態を自動適用しない。

## 訂正・例外4項目の概念

[合意正本](../domains/medical-records/lifecycle.md#medical-record-agreement-correction-four)。項目実施取消は理由と前後状態・操作主体／時刻を持ち、診療終了Undoと異なる。患者訂正は元患者の原記録を保持したまま必要内容を正しい患者へ転記し、元／先を対応付ける。患者IDの付替えや全件自動複写ではない。過去入力は実診療日と現在入力日時が異なり、当日カード／自動課金を生まない。非表示／復元は空誤作成の元ID・履歴を保持する操作で、個人表示設定・診療再開と異なる。以前の保存版・既存追記の保持と別ownerの請求訂正を維持する。

## 追加3項目の概念

[業務正本](../domains/medical-records/pages/medical-record-form.md#medical-record-agreement-integration-three)。画像の撮影／選択／D&Dは同一の入力検証へ渡す入口（最新画面正本）であり、端末撮影からの保存も原本保護へ従う。印刷対象範囲は認可された保存済み情報の選択で、関係する訂正・追記の識別を保持し公開範囲を拡大しない。同一検査結果は検査ownerの結果ID・結果版・変更履歴を持ち、複数画面から参照する。旧確定時を含む保存時参照版と後日結果版は別で、以前の保存版を不変とし現在の記録更新と区別する。訂正影響は関連ownerへの確認・操作の必要性で、金額／在庫の増減確定や実行済み状態と同義ではない。影響未確認・取得失敗と、照合済み影響なしを分離する。

## 最終照合の境界用語

[現契約／目標／互換](../../design/domains/medical-records/technical-spec.md#medical-record-reconciled-contract)を区別する。未送信の意味はAPI／操作ごとの保持・置換契約に従い、PATCHの保持とdose param PUTの未送信optionalのNULL置換を区別する。nullはAPIが明示解除を表現できる場合だけ解除で、0・空文字・未取得と同義にしない。版内項目IDと跨版の結果同一性、confirmation IDとaccounting IDを区別。GetOrCreate初期化はGETでもwriter候補、印刷projectionは保存済み版の副作用なしread。編集者ログはslog／best-effortと業務必須履歴を区別し、旧snapshot不明は現在masterで補完しない。内容判定unknownと、確認済みempty／enteredを別の結果として保持する。

## 自動保存を維持した入力復元と終了・履歴

| 用語 | 意味・区別 | 正本 |
|:---|:---|:---|
| タブ開始基準 | そのタブを開いた時点の値・存在状態・版。後続の自動保存値と別 | [タブクリア](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-clear-tab-open-state) |
| 複製直前基準 | 対象の履歴複製を行う直前の値・版・対象。タブ開始基準とは異なる | [複製取消](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-history-copy-undo) |
| 復元操作 | 自動保存済みの変更も基準へ戻し、新しい変更履歴を記録する操作。物理削除や他者／別owner効果の逆実行ではない | 同上 |
| 診療終了 | 同じ予約を診療中から会計待ちへ進める明示操作。通常カルテ・請求・入金の確定と別 | [画面から終了](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-screen-consultation-end) |
| スタッフ向け診療履歴 | 診療終了したカルテを表示する履歴。診療中の途中保存やfinalizedだけでは対象にしない | [終了済み履歴](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-ended-record-history) |


<a id="medical-record-diagnosis-vitals-ontology"></a>
## 第1・第2の複数診断とバイタル測定履歴

2026-10-15、Sentinel_4ec55f21a17c819199d16ad1add270a4 の回答1・2に基づく概念更新。業務と受入の正本は[診断・バイタル2項目](../domains/medical-records/pages/medical-record-form.md#medical-record-agreement-diagnosis-vitals-20261015)。初診判定の受診歴対象は[予約／受付の別合意](../domains/reservations/pages/reservation-management.md#visit-history-included-types)へ従い、この診断・バイタル回答から定義しない。

| 用語 | 定義・関係／多重度 | 制約・正本 |
|:---|:---|:---|
| 第1診断・第2診断 | 通常カルテの診断を分ける既存の2区分。各区分に複数の診断を登録でき、単一の主診断指定へ統合しない | [複数診断](../domains/medical-records/pages/medical-record-form.md#requirement-fr-medical-records-diagnoses-multiple-per-category)。既存の未設定を保持し、新たな必須件数・上限を追加しない |
| 診断病名の疾患カテゴリ | 個々の診断病名がマスタ上で属する分類。第1／第2という記録上の区分とは別 | 同上。名称と所属カテゴリ整合・医院scope・当時名称を維持 |
| バイタル測定 | 対象患者・カルテに対応する測定値／単位・測定日時を持つ記録。一つのカルテに同日を含め複数測定があり得る | [最新値・履歴](../domains/medical-records/pages/medical-record-form.md#requirement-fr-medical-records-vital-measured-time-history)。未入力と0、実測日時と登録／編集日時を分離 |
| 最新バイタル値・測定時刻 | 表示する最新の値と、その値の測定元に属する時刻。測定日も識別できる表示にする | 同上。異なる測定の値を同じ時刻へ混同せず、日時不明・不完全履歴を補完しない |
| バイタル測定履歴 | 同日を含む以前の測定を参照する履歴。測定ごとの変更履歴とも、終了済みカルテを並べるスタッフ診療履歴とも別 | 同上。診療中も当該カルテの測定履歴を参照でき、閲覧だけのwriteなし |

保存・終了後通常編集・基準への復元・元版／必要ログ・別owner保護は既存概念を継承する。診断の多重度追加を新しい疾患分類や自動診断へ拡張せず、最新測定表示から薬量評価の医学的期限・閾値を導出しない。

<a id="medical-record-intake-prevention-ontology"></a>
## 初診問診と予防薬の概念

| 用語 | 定義・関係 | 正本／保護 |
|:---|:---|:---|
| 初診問診回答 | 当該患者・カルテの聞取時点の回答。病歴・接種／投薬歴・異常歴・感染症検査歴等を区別して複数保持 | [初診問診](../domains/medical-records/pages/initial-intake.md)。現在D04属性やD07／D08の確認済み実績と別 |
| 回答状態・取得状態 | あり／なし／不明という回答／未入力と、未取得／失敗のシステム状態 | 同上。未入力・不明を陰性／異常なし・投薬なしに変換しない |
| 出来事の時期／聞取・記録時点 | 過去の接種・投薬・検査等の時期と、回答を聞き取り入力した日時 | 年／年月／概略表現を保持し架空の日を補完しない |
| 自己申告検査歴 | FeLV／FIV等の申告結果と最終時期 | D07の確認済み結果・その版とは別。問診保存から結果を生成・確定しない |
| 予防処方記録 | 対象患者への予防薬処方箋名・日・薬・条件・期間等の保存単位 | [予防薬正本](../domains/medical-records/pages/preventive-medication.md)。D03所有。一般薬歴と参照可能でも接種実績と同一にしない |
| 投薬確認記録 | 実際の投薬についての確認内容・情報源・時期・記録者／保存版 | 処方・交付・予定・請求とは別。複数確認／同月複数を保持 |
| 年間投薬表 | 患者・年・薬剤／処方を識別した1〜12月の投薬状況projection | 元記録に戻れる表示。空欄・不明・旧○を投薬済み又は未投薬へ推測変換しない |
| 専用2帳票 | 予防薬処方箋と年間投薬表の保存済み範囲・版に対応する出力 | D21既存レポート入口から利用する限定例外。一般レポート全体や接種証明書の出力と別 |

一患者に複数の問診回答時点・予防処方・投薬確認があり、年／月セルは複数の元記録を参照し得る。薬名だけを同一性キーにしない。問診から予防処方へ既往歴をコピーしても元回答は変わらず、コピー元／保存版を保持する。帳票表示・印刷は処方保存・投薬確認・在庫増減・会計・配信のトリガーとしない。

## 薬量の自動範囲・複数パターン

[薬量追加採用の正本・追加4AC](../domains/medical-records/lifecycle.md#dose-source-adoption-20261015)へ従う。用量パターンは薬剤／動物種に加え目的・経路で区別する承認対象で、今回の明示選択と当時版を持つ。薬剤masterの1件と一対一に限定しない。単純体重計算から1回製剤量を得る範囲と、医師確認・理由付き手入力を区別する。目的／経路別選択は複雑protocol自動実行ではなく、具体値の臨床承認#261も代替しない。現行単一用量パラメータの記述は実装参照で、複数パターン対応済みの根拠ではない。

<a id="outpatient-treatment-plan-concept"></a>
## 通常診療の治療プラン項目集合

通常カルテに属する、薬剤・処置・検査・物販など今回の治療項目の集合。各項目の状態、今回の単価・数量等、元マスタ／複写元と当時の値を区別する。項目と状態の業務契約は[既存入力契約](../domains/medical-records/pages/medical-record-form.md#medical-record-agreement-input-four)、現在単価等は[薬量表示・明細単価](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-dose-display-unit-price)へ従う。

[統合治療プラン画面](#medical-record-current-concepts-20261018)はこの項目集合を表示・編集し、スタッフ問診・予約由来の飼主原文・過去履歴も扱う画面責務である。画面内の情報すべてをこの項目集合や一つの保存単位に含めない。[入院治療プラン](hospitalization.md#inpatient-treatment-plan-concept)はD05の別概念・保存責務で、その完了後訂正条件を通常診療へ自動適用しない。物理的な集約・多重度・一括transactionはこの区別から選定しない。

<a id="medical-record-current-concepts-20261018"></a>
## 統合プラン・原文・要約・薬量の最新定義

| 推奨語／別名 | 意味・識別・関係 | 制約・追跡先 |
|:---|:---|:---|
| 統合治療プラン画面（画面上の「治療プラン」） | 問診入力と既存右側の過去カルテ履歴を移し、[通常診療の治療項目集合](#outpatient-treatment-plan-concept)も表示・編集する画面責務。治療明細・保存単位と同義ではない | [FR-medical-records-integrated-plan](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-integrated-plan)、AC-medical-records-integrated-history。入力と履歴同時参照・選択複写を維持。統合保存単位は設計 |
| 飼主由来再診事前問診 | 予約×対象ペットの原文。スタッフ聞取の初診問診・診療本文とは別 | [予約語彙](reservations.md#revisit-intake-concept)。AC-medical-records-owner-intake-protected：スタッフ原文編集・clear・過去複写で上書き/消去/他予約へ流用しない |
| 会話要約専用保存項目 | 来院カルテ単位に保存する要約。診療本文、録音、文字起こし全文、スタッフ貼付フォームと別 | [FR-medical-records-conversation-summary](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-conversation-summary)、AC-medical-records-summary-hidden。今回は非表示、音声入力・外部連携はPhase2、自動診療本文上書きなし |
| 1回量／用法／処方総量 | 一回の製剤量／投与方法・回数等／処方する総量という異なる値 | [FR-medical-records-dose-display-unit-price](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-dose-display-unit-price)、AC-medical-records-dose-three-values。薬剤行で常時区別し根拠は詳細、新用法templateや日数から総量計算は採用しない |
| 承認済用量パターン | 薬剤・種・目的・経路等に対応する計算根拠の候補とその版 | [薬量採用正本](../domains/medical-records/lifecycle.md#dose-source-adoption-20261015)、AC-dose-source-simple-scope／AC-dose-source-outside-scope／AC-dose-source-pattern-choice／AC-dose-source-approval-limit。1薬剤=1パターンと固定しない。犬猫単純体重計算、未承認値/評価不能/上限超過の安全境界を維持 |
| 個別明細単価 | 薬剤・処置・検査・物販の今回行に適用する単価。マスタ現在価格と別 | 同FR、AC-medical-records-priced-line-edit。権限内で実施前後に変更でき、通常理由任意・変更者日時前後値保持。その他手入力行・確定済会計訂正の理由必須を解除しない |

<a id="record-technical-version-relations"></a>
## 保存単位・編集世代・発行版の技術案への対応

以下は既決概念を実現する[詳細設計案](../../design/domains/medical-records/technical-spec.md#integrated-save-edit-right-proposal)の語彙対応。物理PK・API採用・業務追加承認ではない。

| 識別する概念 | 関係・制約 | 実現方式の正本 |
|:---|:---|:---|
| 統合画面と保存単位 | 一つのタブが複数ownerの保存結果を参照する。画面統合は全owner一括txと同義でない | [D03保存単位](../../design/domains/medical-records/technical-spec.md#integrated-save-edit-right-proposal)。成功subsetと未保存／不明を区別 |
| 飼主回答target世代 | 予約×petの有効対象に一つの現世代、各世代に0..n回答版。petを戻しても旧版を自動再採用しない提案 | [D01回答設計](../../design/domains/reservations/technical-spec.md#revisit-answer-storage-proposal)。スタッフ原文編集・copy・clearの対象外 |
| 編集権世代／編集session | 一recordに有効保持0..1、移転／解放後の旧世代は新write不可。同actorの複数sessionも別 | [D03 fence案](../../design/domains/medical-records/technical-spec.md#integrated-save-edit-right-proposal)。閲覧権・実編集者・保存版・医師確認とは別 |
| ローカル下書き／保存結果 | 未保存値・元版・未解決操作を持つ。臨床保存版・編集権・監査の代用ではない | [logout保持の本人決定](../domains/medical-records/lifecycle.md#draft-logout-recovery-20261018)で、logout後も本人専用・同じ本人の再ログイン後だけ回収。他staff／別患者への表示適用なし。回収と編集権／保存成功は別、最新比較・明示選択を維持 |
| 医師確認の臨床世代 | [D10の確認対象](accounting.md#accounting-review-version-relations)を参照。全保存版や金額版と別 | 金額のみ変更で医師再確認にしない。D03/D10の対象集合・世代を確認receiptへ対応 |
| 証明系譜・発行版・artifact | 一系譜に複数の不変発行版、有効currentは高々一つという技術案。描画候補は有効発行ではない | [D08固定版案](../../design/domains/vaccinations/technical-spec.md#certificate-edition-proposal)。Q-V1は本人決定で通常縦横の同じ最新版を共有し印刷時に様式選択、固有項目保持。Q-V2は既決適用として訂正保存成功で旧内容の有効出力停止、失敗時は保持 |

<a id="record-edit-right-concept"></a>
## 編集権と未保存入力

| 概念 | 定義・多重度／区別 | 制約・受入への追跡 |
|:---|:---|:---|
| カルテ編集権 | 一つの通常カルテを現在編集できる権利の保持。ある時点の有効保持は0または1 | [FR-medical-records-single-editor](../domains/medical-records/lifecycle.md#medical-record-edit-exclusivity-20261018)。閲覧認可、編集操作の権限、実編集者履歴と別。同じユーザーの別端末も迂回しない |
| 閲覧／編集開始 | 開くのみは閲覧、「編集を開始」は取得操作。取得成立後に編集する | [最新3点の正本](../domains/medical-records/lifecycle.md#record-edit-right-lifecycle-20261018)、AC-medical-records-editor-view-no-acquire／AC-medical-records-editor-explicit-start。全項目保存ボタン方式にはしない |
| 保持／通常解放 | 同カルテのタブ移動・単なる無操作では維持。別画面離脱時は保存結果確認後に解放 | 同正本、AC-medical-records-editor-tab-and-idle／AC-medical-records-editor-leave-result。失敗・結果不明を成功表示しない |
| 強制取得／移転 | 既存医院scopeと対象編集権限を持つスタッフの確認付き取得。旧権利を移す操作 | 同FR、AC-medical-records-single-editor-force-acquire／AC-medical-records-force-acquire-editor-permission／AC-medical-records-force-acquire-confirm-cancel。確認取消は非取得・元状態不変。旧端末の編集保存停止・移転表示、遅延保存受理拒否、移転前成立済保存保持 |
| 切断／再接続 | 通信状態の変化。業務の診療終了や編集権移転の証明ではない | AC-medical-records-editor-disconnect-stop／AC-medical-records-editor-reconnect-check。入力保持・編集送信停止、復帰時の権と直前保存結果照合。単なる再接続で自動再送しない |
| 未保存入力の回復 | 保存済み版と別に保持する入力を再読み込み後も復元可能にすること | AC-medical-records-editor-reload-recover／AC-medical-records-editor-reacquire-select。再取得後に最新内容と比較して必要部分を明示選択適用。自動上書き・破棄・再送なし。タブ開始時クリア/複製直前取消とは別基準 |

編集権のAPI・セッション識別・原子取得/解放・保存受理の実現、切断失効回収方式・秒数、入力の保存場所・保持期間・比較選択UIは未選定。業務上の維持・停止・再取得後比較を未決へ戻さず、端末間下書き共有や理由必須化も追加しない。既存hybrid保存と他ownerの保護を維持する。

<a id="certificate-edition-concept"></a>
## 接種事実と証明書版

接種実績はD08の記録、証明書は認可された保存済み内容を発行時に固定した版である。通常予防接種縦/横、狂犬病予防注射済証という3様式を、別の接種事実3件とは扱わない。訂正後は新発行し旧内容を履歴に保持する。再印刷は最新版、旧訂正版は有効証明として印刷しない。共通基本情報の更新で発行版を自動変更せず、同じ発行版の再印刷は当時内容の再現である。発行番号書式・記録との物理多重度・版保存方式は本整理で決めない。

正本：[証明書の版と3様式](../domains/vaccinations/lifecycle.md#certificate-editions-20261017)、AC-vaccinations-certificate-three-layouts／AC-vaccinations-certificate-correction／AC-vaccinations-certificate-basic-update／AC-vaccinations-certificate-no-side-effect。12:53 UTC本人第3項Aによる[縦横の最新版共有](../domains/vaccinations/lifecycle.md#certificate-shared-layout-edition-20261018)は同じ通常証明書の同じ版から印刷様式を選ぶ業務関係で、layout変更と新発行を区別する。未保存印刷・印刷時自動登録・preview省略を復活しない。専用実装の所在/適合と採用仕様を区別する。
