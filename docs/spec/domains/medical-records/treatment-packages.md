# 診療項目パッケージ（EMR-518）

決定日: 2026-10-20。要件責任者・決定者: MinoruSoga。出典: [EMR-518](https://app.plane.so/baritechllc/browse/EMR-518/)。本人の「治療に限らず他のもパッケージ化」「治療プランで追加で選べる種類は最低限」と、証明書料金は発行前に手動修正・package構成は医院によるとの回答を採用した論理要件。実装・全DTO/FK固定・API公開の完了を意味しない。

目的は頻用項目の繰り返し選択を一度の明示追加へまとめること。構成品の再入力・二重実績・二重請求を減らす。比較する指標は同じ構成の選択操作数と所要時間で、基準値・改善実測は未計測。実施や請求の判断自体を自動化しない。

[既存D03の4状態](pages/medical-record-form.md#medical-record-agreement-input-four)と[3タブ連動](pages/medical-record-form.md#requirement-fr-medical-records-plan-domain-tabs)を継承する。用語は[カルテ概念](../../ontology/medical-records.md#treatment-package-concepts)、技術方向は[設計](../../../design/domains/medical-records/treatment-packages.md)、未完の契約固定は[作業記録](../../../work/emr/emr-518-treatment-packages-20261020.md)を参照する。

<a id="fr-package-composition"></a>
## FR-package-composition — 医院別の診療項目セット

医院は治療プランへ追加する診療項目を組み合わせたパッケージを管理できる。同一パッケージに診察・処置・薬剤・検査・予防接種・定期健診・物販を混在できる。正規の手入力その他も、既存の必須理由・手入力表示・元々master参照を持たない定義を保つ。未照合masterを手入力その他へ読み替えない。

現在のUIに未配線の種類も、確定仕様にある種類は初版から省略しない。検査は[D07 exam_types](../examinations/lifecycle.md)、予防接種は[D08 vaccines](../vaccinations/lifecycle.md)、定期健診は[D09 checkup_types](../checkups/lifecycle.md)として区別する。診察・処置・薬剤は[D03 master契約](pages/medical-record-form.md#medical-record-agreement-input-four)、物販は[D10会計](../accounting/lifecycle.md)と[在庫owner](../inventory/business-rules.md)へ従う。

既存の健診パッケージ・検査パネルは、それぞれのownerが定義した品目単位で含め、子項目・結果構造を保持する。新設する汎用パッケージ同士の入れ子は初版では作らないが、この制限を理由に既存健診パッケージを除外しない。

<a id="fr-package-apply"></a>
## FR-package-apply — 個別展開と副作用の分離

明示追加で構成品を個別の予定・選択明細として一括追加する。構成品ごとの種類・master参照・数量・当時名称/単位・価格・税・保険・順序と、元パッケージ/版/構成品との対応を保持する。親の合計を別請求行として加算しない。同じ品目を含む別の明示追加は別identityとし、名前・日付一致で統合しない。

選択・追加だけでは実施済み、検査結果、接種記録、健診完了、在庫減算、機器送信、証明書、外部通知、請求確定を起こさない。clinical品目は未実施で開始し、物販等は既存会計ownerの未処理の選択明細とする。医療行為の実施済みを偽造しない。

途中品目に不適合があれば全構成品の追加を成立させず、行別理由を表示する。結果不明は元操作を照会し、同じ操作の再送で重複追加しない。

<a id="fr-package-price-history"></a>
## FR-package-price-history — 当時値と現在masterの分離

master参照品は追加時に照合した現在の単品価格・税・保険等をsnapshot保存する。医院別構成に含まれる料金も各品目の既存計算契約へ従い、入力税込額保持と正確な数量を守る。金額は個別品目合算を初期設計とし、新しい固定総額料金・自動値引き・税配賦を作らない。

新しい全品目共通の数量scale上限、税区分、保険補完、医学的用量を導入しない。追加後のmaster/package変更で過去の適用snapshotを書き換えない。無効masterの新規追加は拒否し、既存表示・履歴は保持する。

<a id="fr-package-owner-boundary"></a>
## FR-package-owner-boundary — 個別業務の正本

実施、結果、訂正、請求適格、在庫は各既存ownerが担う。検査・接種・健診は単品追加と同じ予定identityで各タブの未実施表示・同一結果追跡へ接続する。プランと実績の双方から二重請求しない。数量を接種回数・検査回数・結果件数へ推測変換しない。package中止・削除・複写・取消だけで実績・確定会計・在庫を巻き戻さない。

<a id="rabies-package-boundary"></a>
### 狂犬病セットと証明書owner

狂犬病料金セットは医院別packageの一例で、一般機構に必須の特殊型ではない。登録料・済票料等の構成・採用は医院による。医院職員が既存の明示会計候補追加経路を使用する。全国・全医院共通の構成や請求発生イベント、新しい料金ルールエンジン、package選択だけでの請求確定を作らない。

証明書の4料金欄は[D08の発行前手動編集・発行版保持](../vaccinations/lifecycle.md#certificate-fees-manual-20261020)が正本。個別品目合算と証明書の手動確認額（合計欄を含む）は別であり、両者の一致を発行条件にしない。

<a id="package-acceptance"></a>
## 受入条件（12AC、全件NOT_RUN）

| 条件ID | 操作・期待結果 | 対応FR |
|:---|:---|:---|
| AC-package-mixed-items | 7master種＋正規manual_otherを混在し、typed参照・順序・正確数量を保って個別選択明細へ展開 | composition / apply |
| AC-package-no-selection-effects | clinicalは未実施、物販等はowner未処理明細。選択だけで実施・結果・出庫・証明書・請求確定なし | apply |
| AC-package-atomic-rejection | 無効master・scope不一致・失権・依存版変更・必要値欠落等が1件でもあれば全追加zero-write、行別理由表示 | apply |
| AC-package-retry-identity | 同key再送・応答断照会は同じ保存結果。新keyで盲目的再作成せず、新しい明示追加のみ別identity | apply |
| AC-package-snapshot-stable | 適用後のmaster/package変更でも当時値不変。preview後の改定を無確認適用しない | price-history |
| AC-package-exact-quantity | 浮動小数経由・整数丸めなし。owner別精度を守り、検査・接種回数へ推測変換しない | price-history / owner-boundary |
| AC-package-domain-trace | 検査・接種・健診の各タブから同じ予定/実績へ辿れる。健診をD07へ二重保存せず、パネル/子結果を保持 | owner-boundary |
| AC-package-execution-writer | 実施状態writerは一つ。結果待ち・結果取得失敗・確定と実施を区別し、結果受信だけでperformedにしない | owner-boundary |
| AC-package-cancel-history | 取消・削除・複写で過去実績・確定請求を自動巻戻しせず、元版・理由・actor/時刻・対応を保持 | owner-boundary |
| AC-package-no-double-billing | package親＋子、仮予定＋実績、同接種sourceの多重bindingによる二重請求なし | apply / owner-boundary |
| AC-package-clinic-composition | 医院A/Bで登録料・済票料等の構成が異なっても各設定を使用。全医院共通構成/発生条件へ固定しない | composition / owner-boundary |
| AC-package-certificate-fees | 発行前手動修正でも会計/master/package不変。発行版は確認値を保持し、preview/再印刷から料金を生成しない | owner-boundary / D08正本 |

条件は後続fixture・試験の入力であり実行証跡ではない。今回の追加は4FR/12ACで、既存の合意時点別件数を遡って改定しない。D08の料金境界は既存証明書FRの追補とし、同じAC本文を二重管理しない。
