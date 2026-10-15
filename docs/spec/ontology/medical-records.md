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

項目ごとの状態と記録履歴を持ち、checkbox解除を実施取消へ変換しない。理由付き実施取消は追加合意の専用手順へ。実施記録と治療明細／課金の具体的連動は別ownerの接続判断。入院snapshotへ通常カルテの状態を自動適用しない。

## 訂正・例外4項目の概念

[合意正本](../domains/medical-records/lifecycle.md#medical-record-agreement-correction-four)。項目実施取消は理由と前後状態・操作主体／時刻を持ち、診療終了Undoと異なる。患者訂正は元患者の原記録を保持したまま必要内容を正しい患者へ転記し、元／先を対応付ける。患者IDの付替えや全件自動複写ではない。過去入力は実診療日と現在入力日時が異なり、当日カード／自動課金を生まない。非表示／復元は空誤作成の元ID・履歴を保持する操作で、個人表示設定・診療再開と異なる。以前の保存版・既存追記の保持と別ownerの請求訂正を維持する。

## 追加3項目の概念

[業務正本](../domains/medical-records/pages/medical-record-form.md#medical-record-agreement-integration-three)。画像の選択／D&Dは同一の入力検証へ渡す2入口。印刷対象範囲は認可された保存済み情報の選択で、関係する訂正・追記の識別を保持し公開範囲を拡大しない。同一検査結果は検査ownerの結果ID・結果版・変更履歴を持ち、複数画面から参照する。旧確定時を含む保存時参照版と後日結果版は別で、以前の保存版を不変とし現在の記録更新と区別する。訂正影響は関連ownerへの確認・操作の必要性で、金額／在庫の増減確定や実行済み状態と同義ではない。影響未確認・取得失敗と、照合済み影響なしを分離する。

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
