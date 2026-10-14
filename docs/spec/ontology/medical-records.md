# カルテの概念・関係

2026-10-14初回5項目と入力4項目の部分合意。業務本文・承認根拠・受入条件は[保存正本](../domains/medical-records/lifecycle.md#medical-record-agreement-first-five)。予約語彙は[予約正本](reservations.md)を継承し、以下から専用記録や他領域の内部を確定しない。

| 用語 | 定義・区別 | 制約の所在 |
|:---|:---|:---|
| 通常カルテ | 診療の記録。予約、施術・入院・ホテル専用記録とは別 | [受付接続](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-reception-and-retrospective-entry) |
| 編集者 | 実際に入力・編集した権限あるスタッフ。担当医本人とは限らない | [編集ログ](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-authorized-editor-audit) |
| 確認医師 | 医院が認め、臨床内容・確定・訂正・評価不能手入力を確認する医師。担当医以外も代行可能 | 同上。新しい臨床責任者役割を設けない |
| カルテ担当医／予約指名医師 | 記録の担当と予約の指名。編集者・確認医師と自動的に同一にしない | [予約概念](reservations.md) |
| 作成中／確定済み | 通常入力可能な段階と、本文・子入力がロックされ訂正追記を使う段階 | [確定契約](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-explicit-finalize-plan-warning) |
| 治療プラン未完了警告 | 未実施の予定が残ることを確定前に知らせる警告。確認後に確定可能 | 同上。確定を実施済みへ置換しない。未実施／実施済み／後日予定／中止の状態は合意済み、現モデルとの保存対応は設計 |
| 空の誤作成 | 臨床入力がなく、削除を妨げる請求等関連がない誤作成。draftだけでは空とは言わない | [削除・訂正](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-empty-delete-reviewed-correction) |
| 記録確認訂正／訂正追記 | 記入済みの訂正を確認する業務と、確定済み原記録を保持して追記する方法 | 同上。必要内容転記・元／先対応は追加合意、具体的保存対応は設計残件 |
| 過去未登録診療の追加入力 | 過去の診療で未登録だった記録を理由・医師確認付きで専用入力する業務。確定記録の追記とは別 | [過去入力](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-reception-and-retrospective-entry) |
| 薬量評価不能／理由付き手入力 | 必要情報不足で計算・安全評価ができない状態と、医師確認・理由を記録して入力する業務 | [薬量境界](../domains/medical-records/lifecycle.md#requirement-fr-medical-records-unevaluable-dose-manual)。上限超過解除とは別 |

関係：一つの通常カルテは複数回編集され、それぞれの編集者と履歴を持つ。担当医と別の医師による確認も記録する。確定済み記録は複数の訂正追記を持てるが原記録を置換しない。通常受付の同一予約・active記録の一意性は既存予約契約、過去入力の同一性・重複判定は今後の専用設計へ分ける。治療プラン・子入力・請求の依存は単なる本文空欄判定から除外しない。

## 入力4項目の概念

業務正本は[追加合意](../domains/medical-records/pages/medical-record-form.md#medical-record-agreement-input-four)。

| 用語 | 定義・関係 | 制約 |
|:---|:---|:---|
| 選択複写 | 過去問診／治療明細の選択対象を今回記録へ適用する。原記録は不変 | 問診上書きは確認、新明細は今回価格、無効／不一致は通知再選択。逐次保存の部分成功を維持 |
| Template挿入 | 登録済み定型文を手動選択して問診の未保存入力に入れる | 基本追記、置換のみ確認。自動提案／元template更新なし |
| 当時内容／今回価格 | 既存記録が保持する名称・単位・価格と、新しい複写明細が使う今回の価格 | master改定／無効化で既存値を変更しない。根拠のない旧値補完をしない |
| 治療プラン項目状態 | 未実施・実施済み・後日予定・中止の4状態 | 未実施／後日予定を確定前警告、中止は理由。警告確認後確定可 |
| 実施記録者／記録時刻 | 実施済み等を記録した操作主体とその記録時刻 | 実際の実施者／実績時刻と区別し、未指定項目を必須化しない |

項目ごとの状態と記録履歴を持ち、checkbox解除を実施取消へ変換しない。理由付き実施取消は追加合意の専用手順へ。実施記録と治療明細／課金の具体的連動は別ownerの接続判断。入院snapshotへ通常カルテの状態を自動適用しない。

## 訂正・例外4項目の概念

[合意正本](../domains/medical-records/lifecycle.md#medical-record-agreement-correction-four)。項目実施取消は理由と前後状態・操作主体／時刻を持ち、診療終了Undoと異なる。患者訂正は元患者の原記録を保持したまま必要内容を正しい患者へ転記し、元／先を対応付ける。患者IDの付替えや全件自動複写ではない。過去入力は実診療日と現在入力日時が異なり、当日カード／自動課金を生まない。非表示／復元は空誤作成の元ID・履歴を保持する操作で、個人表示設定・診療再開と異なる。確定本文の保護と別ownerの請求訂正を維持する。

## 追加3項目の概念

[業務正本](../domains/medical-records/pages/medical-record-form.md#medical-record-agreement-integration-three)。画像の選択／D&Dは同一の入力検証へ渡す2入口。印刷対象範囲は認可された保存済み情報の選択で、関係する訂正・追記の識別を保持し公開範囲を拡大しない。同一検査結果は検査ownerの結果ID・結果版・変更履歴を持ち、複数画面から参照する。確定時参照版と後日結果版は別で、確定済み原記載は不変。訂正影響は関連ownerへの確認・操作の必要性で、金額／在庫の増減確定や実行済み状態と同義ではない。影響未確認・取得失敗と、照合済み影響なしを分離する。

## 最終照合の境界用語

[現契約／目標／互換](../../design/domains/medical-records/technical-spec.md#medical-record-reconciled-contract)を区別する。未送信の意味はAPI／操作ごとの保持・置換契約に従い、PATCHの保持とdose param PUTの未送信optionalのNULL置換を区別する。nullはAPIが明示解除を表現できる場合だけ解除で、0・空文字・未取得と同義にしない。版内項目IDと跨版の結果同一性、confirmation IDとaccounting IDを区別。GetOrCreate初期化はGETでもwriter候補、印刷projectionは保存済み版の副作用なしread。編集者ログはslog／best-effortと業務必須履歴を区別し、旧snapshot不明は現在masterで補完しない。内容判定unknownと、確認済みempty／enteredを別の結果として保持する。
