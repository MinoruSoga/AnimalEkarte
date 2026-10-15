# D08 予防接種の技術契約

| 項目 | 内容 |
|---|---|
| 文書ID | design-D08-vaccinations-technical |
| 種別 | 技術契約・実装差分 |
| 決定状態 | 既存記載を継承（個別の決定根拠は本文参照。根拠未確認部分は承認済みとしない）＋D03合意目標への接続。新規の未決事項は未決のまま |
| 実装状態 | 本文の現コード観察と目標契約を分離。コードに存在するだけでは業務承認済み・目標実装済みとしない |
| 実機検証状態 | 今回未実施。既存証跡の引用は今回の実機受入を意味しない |
| 文書確認日 | 2026-10-15 |

既存仕様の継承とD03の合意済み目標を整理した文書候補。ソフトウェア仕様の整理、実装完了、実機受入、医師による具体的な臨床数値の承認は別の判定とする。

## 範囲
[業務正本](../../../spec/domains/vaccinations/lifecycle.md)の5FR／23AC。基準commit `41fcf3eb0a8befb7c0b5f613d1df749a60e30be8`の静的照合で、今回runtime未実施。D03[保存§2](../../../spec/domains/medical-records/lifecycle.md#hybrid-sync)・[型／writer契約](../medical-records/technical-spec.md#medical-record-reconciled-contract)を保持。

## 入口・資源・保存結果
独立とカルテ内は同じGET/POST/PATCH/DELETE /api/v1/vaccinations。カルテ内GETはpet_id＋page/limit、POSTは保存済みmedical_record_idを送る。独立はrecord NULLで保存し得る。接種自身にfinalized状態を新設しない。同日複数は単件POSTの順次実行、バッチAPIを推定しない。[EMR-105](../../../work/emr/emr-105-vaccine-multi-20261002.md)の過去検証は保存済みカルテ内3件実DB／2件モックフォーム等の限定証拠で、今回の再実行や全UI受入とは別。

[service](../../../../backend/internal/medicalrecord/vaccination_service.go)はcreate/updateで同医院の関連と親状態をtx内に照合し、updateは旧／新親を安定した順序でlockし元relation tupleを再照合する。deleteは親→子lock・同一tx削除監査で、失敗時同期を発火しない。create/updateの全編集者履歴は現slogだけでD03の必須監査を充足したとしない。

## 型・日付・clear・既定
| 項目 | 現契約 | 適合／停止条件 |
|:---|:---|:---|
| ID | [request](../../../../backend/internal/medicalrecord/vaccination_request.go)：pet／record／doctor optional数値pointer、vaccine必須 | 医院＋IDと認証actorを分離。不正IDを0へ補完しない |
| 接種日 | date必須、serviceはJST暦日で未来拒否 | UTC instant差と暦日を混同しない。新規予定を接種実績にしない |
| 次回日 | optional、serviceは指定時に接種日より後を検証 | 独立フォームの新規本日以降条件はAPIに同じ条件があると断定しない |
| 間隔 | 3weeks／4weeks／1year／other、独立既定1year・カルテ内4weeks | マスタ指定優先・手動候補を保持、既定統一／種フィルタ未実装を臨床承認としない |
| LOT等 | lot1〜4、supplemental、remarks | 空文字と未送信を区別、他行に値を持ち越さない |
| PATCH nullable | optional pointerのnullは未送信と区別できない | doctor／next_date／record解除済みと誤表示しない。対応する専用clearが未定義なら解除を送らない |
| 期待版 | version任意、nilは互換上照合skip | 新adapterは期待版を保持。旧clientが無版で更新できることを競合防止完成としない |
| snapshot | vaccine参照と価格preload等 | 当時名称／価格／間隔の完全snapshotありと仮定しない。D03/D10の当時値確認へ |

現vaccines masterは[master routes](../../../../backend/internal/medicalrecord/routes_masters.go)のmaster-medical操作別grantによるCRUD/reorder。接種実績のvaccinations grantとは別で、checkup masterのcheckups grantとも区別する。

## writer・失敗と受け手
- 全入口CRUD、関連変更、カルテ内フォーム、専用訂正／転記／過去入力／復元adapter、タグ同期を列挙する。普通APIで親finalizedや患者取り違え保護を回避しない。
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
| 日付・マスタ候補 | future, master-interval, manual, next-invalid, default-boundary |
| write保護・結果 | finalized, scope, deceased, delete-one, conflict, unknown, edit-log |
| 一覧・読取 | filter-scope, navigation, projection |
| master／他owner | master-history, notify-failure, no-side-effect |

## 完了判定外
種不一致の医学的妥当性、製品別の具体的間隔承認、医院実grant、旧snapshot、create/update監査の充足、全client期待版・receipt、D15配信運用・D20公開条件の実装、実ブラウザ／実データ受入は未確認。機能追加の未承認事項と既存挙動の記載を分離する。


D21 Owner Reportはスタッフの閲覧専用であり、同画面からの印刷・共有を設けない。D03や検査自ドメインに既存の印刷経路があることを、D21経由の印刷許可へ拡張しない。
