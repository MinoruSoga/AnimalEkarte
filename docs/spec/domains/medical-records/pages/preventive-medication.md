# 予防薬の専用管理・処方箋・年間投薬表

| 管理項目 | 記載 |
|:---|:---|
| 文書ID・種別 | page.medical-records.preventive-medication／D03 業務・画面仕様 |
| 決定状態 | 2026-10-15、Sentinel_0adf405e47508191abf775b6c4af22a3 の回答2「yes 現状、カルテページ右上のレポートボタンを押したら開かれるところで表示するようにしています」。専用管理・印刷の維持と既存レポート入口を採用 |
| 実装状態 | docs限定。固定headで既存レポート入口と一般薬歴表示を確認。専用2帳票の実装適合は未確認 |
| 実機検証状態 | 未実施。ユーザー申告、静的観察、目標を分ける |
| 文書確認日 | 2026-10-15 |

<a id="preventive-medication-scope"></a>
## 範囲・責務・既存入口

予防薬専用の処方箋と、薬剤・年を識別して1〜12月の投薬状況を確認する年間表を維持する。一般の「薬・処方」履歴だけへの統合で代替しない。D03が予防薬の処方内容・投薬確認記録を所有し、D21が既存レポート入口の参照と専用帳票出力の入口を受ける。D08のワクチン接種実績／証明書とは別機能であり、D08へ処方・投薬writeを移管しない。

入口はカルテ右上の「レポート」から開く既存の[Owner Report](../../../screens/39-owner-report.md#owner-report-preventive-output)。別のサイドメニューや新しい独立ナビゲーションは要求しない。レポートの選択患者に対応する予防薬記録・年間表を参照し、専用の処方箋／年間表を選んで印刷する。管理・訂正はD03の認可された処方／投薬管理へ同じ患者と対象記録を渡す。D21のページ自体に臨床writeを持たせず、view権限を編集権限へ昇格させない。

### 根拠と観察の区別

| 種別 | 確認内容・限界 |
|:---|:---|
| 今回のユーザー決定 | 専用機能・印刷を残す。「現在レポートで表示」との申告は入口指定の根拠であり、今回の実機確認ではない |
| 旧原資料 | [IMG_1827](https://drive.google.com/file/d/1atDQBSJxBdtkRgLZe_mtXZkVEFyKwvT1/view)は処方箋名、処方日、予防薬名、投与条件、期間、既往歴コピー、備考、担当者、印刷。[IMG_1826](https://drive.google.com/file/d/18b_LTjnrbvVFe9Htqb1CgLY7Uc6-fqQZ/view)は名称・年度・1〜12月、連続薬数と参照。○の厳密な意味や連続薬数の計算法は写真だけでは確定しない。個人名・患者別数値は転載しない |
| 旧手順 | [マニュアル#2](https://drive.google.com/file/d/1KzKz30mXwkQHG9s8m9G6a3lm5jiNlf4L/view)「カルテ入力３」B288:B307・image015/017は専用処方箋の登録／印刷。[マニュアル#1](https://drive.google.com/file/d/1p0wyop7CvgKzuvw5P0_ZEo1elcMMz9Av/view)「カルテ入力」B206:B208は処方箋登録とメイン画面での投与実績設定を別操作としている。旧ボタン・分数ナビ等をそのまま新UI要件にしない |
| 固定headの静的観察 | 908b1def96c91885c01b038bd71af44f500207bf の MedicalRecordStickyHeader → openOwnerReport → OwnerReport は `/owners/:id/report?petId=`。OwnerClinicalHistoryPanel／clinical-briefing は一般治療履歴の medicine を「薬・処方」として日付別表示。この読取範囲で専用処方箋・1〜12月表・専用印刷は確認できず、採用仕様への適合確認が必要 |

<a id="requirement-fr-medical-records-preventive-prescription"></a>
### FR-medical-records-preventive-prescription：専用処方記録を維持する

| 記録内容 | 意味・保護 |
|:---|:---|
| 医院・患者・処方記録／版、関連カルテ | 同じ患者の対象処方を辿る。氏名・薬剤名の一致だけで同一記録としない |
| 予防処方箋名、処方日、予防薬名 | 表示名と薬剤参照／当時表示を保持。複数処方を記録・再参照でき、master改名で過去帳票の当時値を再生成しない |
| 投与条件、処方期間 | 記録した条件を保存。年／年月など原情報の精度を保持し、具体頻度・用量・臨床閾値を本仕様で選定しない |
| 既往歴、備考、担当者 | 既往歴コピーは利用者が対象を選び元記録／版とコピー時点を保持。現在の既往歴と当時コピー、担当者と実編集者を区別 |

処方記録の保存と投薬確認は別の状態・操作。処方日・期間・数量・交付済み・会計済みだけで「投薬済み」にしない。一般処方と参照を共有できる場合も専用処方箋の内容・同一性と年間表との関係を保ち、同一処方を再createして二重請求・在庫減算しない。

既存D03認可・自動保存／子資源の保存境界、終了後通常編集、元版／編集履歴、競合／結果不明照合を継承する。必要な保存と監査が成功した対象だけ成功と表示し、未対応wireで項目を捨てない。既往歴コピーも既存の選択・上書き確認・複製取消契約へ接続し原記録は不変。タブクリア／取消を投薬済み取消や別ownerの在庫・金額取消へ拡張しない。

<a id="requirement-fr-medical-records-annual-preventive-medication"></a>
### FR-medical-records-annual-preventive-medication：1〜12月の投薬状況を専用管理する

対象患者・対象年・予防薬／処方記録を識別し、1〜12月を固定列として見られる。複数薬剤・複数年・同月複数記録を混同せず、月セルから元記録と根拠を確認できる。年度ラベルを使用しても1〜12月の暦年を別の開始月へ暗黙に読み替えない。

- 処方／交付、投薬予定、実際の投薬確認を識別する。確認済み実績があるときだけ投薬済みを表示し、確認元（院内実施記録・飼主申告等）と確認時点、分かる範囲の実投薬時期と精度、記録者／版を辿れるようにする。
- 未入力、「わからない」という回答、未投薬という確認、取得失敗を分ける。空セル・未取得・旧○だけから投薬済み／未投薬を断定しない。医学的頻度・自動予定を新設しない。
- 同月の複数記録を一つの○で上書き・消去しない。予定と実績や矛盾する申告は識別可能な明細を保つ。新しい投薬回数上限は作らない。
- 訂正は元値・目的別訂正に必要な理由・版／履歴を保持し、削除や復元で元事実・会計・在庫を黙って取り消さない。過去分・死亡後の生前分記録にはD03既存専用経路を適用する。
- 旧「連続薬数」の計算式は今回決めない。対応未確認なら原値と出所を保持して未確認とし、12か月の○数や連続月数へ置き換えない。

<a id="requirement-fr-medical-records-preventive-report-output"></a>
### FR-medical-records-preventive-report-output：既存レポート入口から専用2帳票を参照・印刷する

選択患者の予防処方箋と年間投薬表を表示・選択し、保存済み対象処方／年／薬剤範囲・保存版をプレビューして専用帳票として印刷する。予防処方箋は処方日・薬名・条件・期間・必要な既往歴／備考・担当を、年間表は対象年・薬剤・1〜12月・状態の意味／凡例を対応付ける。未保存入力を自動保存して印刷しない。画像化された旧紙だけの出力は構造化管理を代替しない。

D21一般レポート全体の印刷・共有は追加しない。スタッフ専用ページ／section認可・一頭contextを維持し、今回明示された予防薬専用2帳票だけを例外とする。接種・検査・健診等の他sectionを無条件に同梱せず、飼主向け公開、リンク共有、LINE送信は追加しない。

医院・患者・処方／投薬記録と版、現在閲覧権限を出力直前にも照合する。年間表の対象年／選択薬剤の必要ページを取得し切り捨てを隠さない。欠損・競合・取得失敗・版不一致・権限不足なら出力停止。未入力と確認できたセルは意味を明記して出力できる。プレビュー後に対象や保存版が変われば再取得・再確認する。

プレビューを閉じる、印刷取消、ブラウザ印刷完了が確認できない場合に投薬済み・発行完了・請求済みへ状態変更しない。同じ保存版の再印刷も新処方・実績・在庫／会計・配信を発火しない。

<a id="preventive-medication-ac"></a>
## 受入条件（3FR／15AC）

| 条件ID | Given | When | Then | 対応FR |
|:---|:---|:---|:---|:---|
| AC-medical-records-prevention-fields | 専用処方が複数ある | 入力・保存・再参照 | 処方箋名／日、薬名、条件、期間、既往歴・備考・担当と各ID／版を保持 | preventive-prescription |
| AC-medical-records-prevention-history-copy | 既往歴を処方へコピー | 選択・置換確認後に取消 | 元記録／版を辿れ、元回答不変。複製直前へ対象だけ復元 | preventive-prescription |
| AC-medical-records-prevention-no-auto-dose | 旧紙に薬量・頻度例がある | 処方を新規入力 | 例を初期値・臨床閾値・自動予定へ採用しない | preventive-prescription |
| AC-medical-records-prevention-write-boundary | 処方保存／通常編集・クリア | 会計・投薬・在庫状態を確認 | 処方と実績を区別しowner保護と元版・ログを保持。新請求／投薬済み／在庫調整の自動生成なし | preventive-prescription |
| AC-medical-records-prevention-write-failure | 失敗・競合・結果不明・履歴失敗 | 再送検討 | 入力保持、元操作／保存版を照合。別ID重複作成や成功誤表示なし | preventive-prescription |
| AC-medical-records-prevention-annual-grid | 複数薬剤・複数年 | 年を選び表示 | 対象患者・年・薬剤と1〜12月を保持。薬名だけで別処方を統合しない | annual-preventive-medication |
| AC-medical-records-prevention-state-meaning | 処方、予定、投薬確認、未投薬確認、不明、未入力 | 月セルを表示 | 状態と確認元を区別。処方／交付／会計を投薬実績へ変換しない | annual-preventive-medication |
| AC-medical-records-prevention-month-multiple | 同じ薬・月に複数記録／相違する申告 | 年間表から明細を開く | 元ID・時期精度・出所・版を保持し単一○で消さない | annual-preventive-medication |
| AC-medical-records-prevention-legacy-unknown | 旧○や連続薬数の意味が未確認 | 移行値を参照 | 原値／出所を保持し未確認表示。投薬済み・未投薬や新集計へ推測変換なし | annual-preventive-medication |
| AC-medical-records-prevention-correction | 訂正又は過去・死亡後生前分記録 | 保存・再参照 | 既存専用保護と元履歴を保持。未来の新実施や自動返金／在庫戻しなし | annual-preventive-medication |
| AC-medical-records-prevention-report-entry | カルテに選択患者あり | 右上レポートを開く | 既存D21で同患者の専用処方箋／年間表を参照。一般薬歴だけで代替しない | preventive-report-output |
| AC-medical-records-prevention-two-print | 保存済み処方／対象年・薬剤を選択 | プレビューして印刷 | 範囲と版の一致する専用2帳票だけ出力。D21全体や他section印刷／外部共有なし | preventive-report-output |
| AC-medical-records-prevention-print-stale | 医院／患者切替、版更新、権限失効、不完全取得 | 出力を要求 | 古い患者・版や欠落年表を出力せず再照合・再確認 | preventive-report-output |
| AC-medical-records-prevention-print-cancel | プレビュー中又は印刷結果不明 | 閉じる／取消／再印刷 | 臨床write・投薬済み・請求／在庫・配信を発火せず、実印刷完了を断定しない | preventive-report-output |
| AC-medical-records-prevention-report-view | レポート閲覧のみの権限 | 参照・管理入口・出力 | D03編集権限へ昇格せず、同患者と現在権限で認可された保存済み範囲だけ扱う | preventive-report-output |

## 未決と利用開始前確認

採用範囲の画面／保存・版・認可・専用印刷設計は上記で進められ、通常の安全設計について追加質問は不要。「旧○や連続薬数をどの事実として移すか」は旧データ移行時に元仕様／データで照合する業務意味で、未確認のまま自動移行しない。今回採用を妨げる追加の必須業務質問はない。薬量・製品別頻度・適応と臨床閾値承認は既存の別工程。API／DB・印刷レイアウト、全件取得・実機受入は[技術契約](../../../../design/domains/medical-records/technical-spec.md#medical-record-intake-prevention-contract)の実装確認事項とする。
