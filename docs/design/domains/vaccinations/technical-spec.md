# D08 予防接種の技術契約

| 項目 | 内容 |
|---|---|
| 文書ID | design-D08-vaccinations-technical |
| 種別 | 技術契約・実装差分 |
| 決定状態 | 2026-10-17追加決定は末尾と進捗入口を参照。従来：既存記載を継承（個別の決定根拠は本文参照。根拠未確認部分は承認済みとしない）＋D03合意目標への接続。次回日は種類別設定または手入力へ統一する既合意、および過去接種の当時の次回日を保持する最新第4項合意を反映。対象外の未決は保持 |
| 実装状態 | 本文の現コード観察と目標契約を分離。コードに存在するだけでは業務承認済み・目標実装済みとしない |
| 実機検証状態 | 今回未実施。既存証跡の引用は今回の実機受入を意味しない |
| 文書確認日 | 2026-10-17追加差分（従来確認2026-10-15） |

既存仕様の継承とD03の合意済み目標を整理した文書候補。ソフトウェア仕様の整理、実装完了、実機受入、医師による具体的な臨床数値の承認は別の判定とする。

## 範囲
[業務正本](../../../spec/domains/vaccinations/lifecycle.md)の5FR／28AC。基準commit `41fcf3eb0a8befb7c0b5f613d1df749a60e30be8`の静的照合で、今回runtime未実施。D03[保存§2](../../../spec/domains/medical-records/lifecycle.md#hybrid-sync)・[型／writer契約](../medical-records/technical-spec.md#medical-record-reconciled-contract)を保持。

## 入口・資源・保存結果
独立とカルテ内は同じGET/POST/PATCH/DELETE /api/v1/vaccinations。カルテ内GETはpet_id＋page/limit、POSTは保存済みmedical_record_idを送る。独立はrecord NULLで保存し得る。接種自身にfinalized状態を新設しない。同日複数は単件POSTの順次実行、バッチAPIを推定しない。[EMR-105](../../../work/emr/emr-105-vaccine-multi-20261002.md)の過去検証は保存済みカルテ内3件実DB／2件モックフォーム等の限定証拠で、今回の再実行や全UI受入とは別。

[service](../../../../backend/internal/medicalrecord/vaccination_service.go)はcreate/updateで同医院の関連と親状態をtx内に照合し、updateは旧／新親を安定した順序でlockし元relation tupleを再照合する。deleteは親→子lock・同一tx削除監査で、失敗時同期を発火しない。create/updateの全編集者履歴は現slogだけでD03の必須監査を充足したとしない。

## 型・日付・clear・既定
| 項目 | 現契約 | 適合／停止条件 |
|:---|:---|:---|
| ID | [request](../../../../backend/internal/medicalrecord/vaccination_request.go)：pet／record／doctor optional数値pointer、vaccine必須 | 医院＋IDと認証actorを分離。不正IDを0へ補完しない |
| 接種日 | date必須、serviceはJST暦日で未来拒否 | UTC instant差と暦日を混同しない。新規予定を接種実績にしない |
| 次回日 | optional、serviceは指定時に接種日より後を検証 | 通常独立新規の本日以降条件と、専用後日入力での当時の期限経過日保存を分離。APIに同じ本日条件／専用例外が実装済みと断定しない |
| 間隔 | 旧UIの標準候補3weeks／4weeks／1year／other、独立既定1year・カルテ内4weeks | 目標は両入口ともワクチン種類別設定または手入力。入口固有の固定既定を解消し手入力値を黙って上書きしない。具体間隔・設定なしの初期値・新たな優先順位はこの合意で決めない。実装適合未確認 |
| LOT等 | lot1〜4、supplemental、remarks | 空文字と未送信を区別、他行に値を持ち越さない |
| PATCH nullable | optional pointerのnullは未送信と区別できない | doctor／next_date／record解除済みと誤表示しない。対応する専用clearが未定義なら解除を送らない |
| 期待版 | version任意、nilは互換上照合skip | 新adapterは期待版を保持。旧clientが無版で更新できることを競合防止完成としない |
| snapshot | vaccine参照と価格preload等 | 当時名称／価格／間隔の完全snapshotありと仮定しない。D03/D10の当時値確認へ |

現vaccines masterは[master routes](../../../../backend/internal/medicalrecord/routes_masters.go)のmaster-medical操作別grantによるCRUD/reorder。接種実績のvaccinations grantとは別で、checkup masterのcheckups grantとも区別する。

次回日の統一は[最新の業務合意](../../../spec/domains/vaccinations/lifecycle.md#fr-vaccinations-schedule)へ従う。種類設定の参照と手入力日を両入口で同じ意味に対応させる。既存schedule値を新しい既定・臨床的標準として固定しない。設定の保存／読取・入力状態の保持・両フォームadapter・再表示を照合し、現行の入口別初期化を対応済みと扱わない。設定変更から過去の保存済み接種／次回日を自動再計算しない。技術対応はAPI／DBの適合調査で具体化し、本docs-only変更でwire schemaや実装を追加しない。

## 過去接種の当時の次回日を保持するadapter

[最新第4項合意](../../../spec/domains/vaccinations/lifecycle.md#vaccination-retrospective-next-date)の対象は過去接種の専用後日入力。実接種日と当時のnext_dateを履歴値として保存し、入力時刻・actorと区別する。next_dateが入力時点より前でも実接種日より後なら期限経過だけで拒否せず、今日以降への補正や現masterによる再計算をしない。通常の独立新規入力には既存の本日以降条件を維持し、未来接種日の拒否と実接種日との前後関係を両経路で保つ。

対象医院／患者／元記録・親状態・grant・必要理由／履歴を既存専用経路で照合し、通常POST/PATCH全体の検証解除で代用しない。D03の過去入力・確定後追記、および死亡後生前分記録へ接続する。選択した履歴値は再取得・projectionでも同じ日付で返す。現行APIの接種日より後という検証だけでは、専用UI／入力目的の識別・履歴／認可対応まで完成していない。必要adapterのAPI／DB適合は後続の実装確認であり、本変更で新wire fieldや許可経路を捏造しない。

## 死亡登録後の生前接種の専用経路

[死亡後の生前分記録](../../../spec/domains/vaccinations/lifecycle.md#fr-vaccinations-correction)を最新合意の目標として扱う。専用経路は死亡状態を維持したまま、生前分であること・対象医院／患者／元記録・必要権限・理由・現在版を保存時に再照合する。実施時点不明は推測せず確認まで停止。原記録と前後差分・理由・actor／確認者・入力時刻・元記録関連を保持し、必要履歴失敗は部分成功にしない。結果不明は元操作を照合し別IDで重複作成しない。現行の通常create/PATCH・一律死亡拒否を専用経路の実装済み証拠にせず、API／版／receipt／権限対応は適合確認に残す。死亡状態の一時解除、通常writeへの無制限な死亡gate解除で代用しない。

実接種日と現在の登録時刻を分け、D03確定親への原記録保持・専用追記を維持する。次回日の統一合意から死亡後の新しい接種／予約を許可せず、通知の死亡除外も解除しない。

## writer・失敗と受け手
- 全入口CRUD、関連変更、カルテ内フォーム、専用訂正／転記／過去入力／復元adapter、タグ同期を列挙する。旧親finalizedによる一律拒否は最新D03の通常編集目標へ合わせる適合差であり、現在の患者／医院・権限・版・履歴を検査する。患者取り違え・死亡後専用処理・削除の保護を普通APIで回避しない。
- フォームのgrant・死亡判定だけでサーバ全経路保護済みとしない。現在clinic/pet/record/doctor/vaccine関係とactorを検査し、scope変更後の古い応答は破棄／再取得する。
- 送信失敗は入力保持。通信断はunknownとして保存先ID／操作結果を照会。現単件POSTに汎用operation receiptがあると仮定せず、結果不明の再createを止める。既存ID／実操作の照合方式は実装で具体化する。
- D03から継承する必須編集ログは変更前後・対象・理由・actor／時刻・結果を同じ業務単位で保持する。現delete監査だけを全writeの充足としない。
- [現service](../../../../backend/internal/medicalrecord/vaccination_service.go)のタグ同期はcommit後、失敗はログであり接種保存をrollbackしない。同期／配信失敗通知・再試行・停止・auditはD15へ渡す。現在ログのみで人への失敗通知まで満たすとしない。
- D20/D21へclinic/pet/接種ID・date・vaccine表示情報・next_date・取得状態・更新版／時刻を渡す。保存済み原事実の受渡しで、D08のview／createは飼主公開認可ではない。D20の飼主公開条件とD21のスタッフ専用Owner Report閲覧認可は各所有者の既存／合意正本を適用する。D21に飼主公開や新しい承認操作を追加せず、D20の承認済み公開条件に従う。本人に紐付く生存ペットの保存済み接種実績を単位とし、親カルテdraft／紐付きなしでも新たな親finalized条件や接種承認操作を要求しない。未保存・取消／削除は非公開、訂正保存後は更新する。根拠は[健康手帳正本](../../../spec/screens/38-liff-pet-health.md#liff-health-document-contract)の2026-10-15承認 Sentinel_7a67fd30e47c819192cd36d416f0a2bf。実装適合は別途検証する。
- 会計金額snapshotはD10、在庫副作用はD13。臨床訂正成功と会計／在庫修正receiptを別にする。

## AC追跡
| 技術部分 | 業務AC suffix |
|:---|:---|
| 同じ資源・単件保存 | standalone, in-chart, same-day, reset, lots |
| 日付・マスタ候補 | future, master-interval, manual, next-invalid, retrospective-expired-next, normal-new-next-past, retrospective-protection, default-boundary, no-setting-default |
| write保護・結果 | finalized, scope, deceased, deceased-boundary, delete-one, conflict, unknown, edit-log, clear-copy-guard |
| 一覧・読取 | filter-scope, navigation, projection |
| master／他owner | master-history, notify-failure, no-side-effect |

## 完了判定外
種不一致の医学的妥当性、製品別の具体的間隔承認、医院実grant、旧snapshot、create/update監査の充足、全client期待版・receipt、D15配信運用・D20公開条件の実装、実ブラウザ／実データ受入は未確認。機能追加の未承認事項と既存挙動の記載を分離する。


D21 Owner Reportの接種履歴sectionはスタッフ閲覧専用で、接種情報の印刷・共有を同画面から追加しない。最新採用の[予防薬専用処方箋・年間投薬表](../../../spec/screens/39-owner-report.md#owner-report-preventive-output)だけがD21の限定印刷例外であり、D03の予防薬管理とD08の接種実績／証明書を区別する。D21一般レポート全体や本sectionへの出力許可に拡張しない。


### 最新カルテ契約へのadapter

[通常編集・自動保存／復元](../../../spec/domains/vaccinations/lifecycle.md#vaccination-normal-edit-autosave)に従い、旧親finalizedだけを拒否理由にする判定を目標から外す。ただし患者／親recordの同一性、現在grant、対象版、日付制約、死亡後の専用理由／履歴と必要監査は維持する。復元操作は対象の接種ID・基準版・現在版・前後値・操作identityを照合し、必要な変更履歴と保存結果を一体にする。新規接種・別タブ・通知／在庫／会計の操作を復元に連鎖させない。既存PATCHの監査・期待版／receipt不足を対応済みとはせず、結果不明は同一操作照会後に続ける。AC-vaccinations-finalized／AC-vaccinations-clear-copy-guardがこのadapterの受入条件。

<a id="decision-alignment-20261017"></a>
## 2026-10-17確定仕様との接続

[決定根拠・適用範囲](../../../spec/domains/SPECIFICATION_PROGRESS.md#confirmed-business-decisions-20261017)：ユーザーの2026-10-16〜17 JSTの直接決定、文書更新承認は10月17日03:20 JST。以下は目標仕様・未実行の受入条件であり、実装・実機合格を示さない。

[業務正本・受入条件](../../../spec/domains/vaccinations/lifecycle.md#certificate-editions-20261017)。通常縦横・狂犬病の3様式は旧見本に従う。訂正後新発行・旧版履歴保持、再印刷は最新版、旧訂正版は有効印刷不可。共通基本情報更新で発行版は自動変更しない。旧印刷時自動登録等は復活させない。専用証明書実装の所在・適合と仕様採用を分ける。
