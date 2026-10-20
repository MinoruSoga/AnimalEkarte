# EMR-499 → EMR-461 有効状態 subset の実装と未確定case

本表はPR734の実装境界・質問の証跡。新しい業務仕様やDDLの採用ではない。PR740のschema/DTO/routerは別ownerであり、この文書で変更しない。

| 入力履歴 | 今回の判定 | 根拠・残差 |
|---|---|---|
| 実admit（actual_atあり）から終了eventなし | admitted | [開始の明示確定](../spec/domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-entry-and-start)。予定終了超過は実退院ではない |
| admit → 有効discharge、実退院時刻が入院以上 | released | [明示退院](../spec/domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-discharge-and-billing)。通常の再check-inは追加しない |
| admit前のcancel/no_show | released | [計数正本](../spec/domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-hotel-stay-capacity)の取消・未来院除外 |
| no movement（予定entry/exitだけあり） | unknown | **未確定Q1**: stay行の作成だけで予約成立を表すか、予約受入/成立の別参照が必要か。予定日時の存在からreservedを捏造しない。499の明示契約が必要 |
| correction・無効化・訂正参照あり | 当該stayのみunknown | **未確定Q2**: 置換後effectの指定、元effect継承か、無効化、actual_at NULLの意味。corrected_event_id/reasonのDDLだけから推測しない |
| discharge単独、admit後cancel/no_show、退院後再admit、実時刻逆転、scope/ID/revision不正 | 当該stayのみunknown | **未確定Q3**: 専用訂正を通した有効履歴の解決規則。最大revisionのevent_kindでは決めない |

`StayCageEffectiveStateReader` は医院付きStayCaseIDについて状態と`EvidenceVersion`を返す。今回の根拠版は検査したmovement履歴の最新revisionで、CAS/fenceやstay行versionの代用ではない。全履歴をrevision順に検査し、同revision・同ID・scope不整合をunknownにする。予定期間は占有解放に使わない。

## readerとlockの成立範囲

readerはambient transaction内のstay行をID昇順`FOR SHARE`で保護し、同じtransactionでmovementを読む。候補読取にambient transactionがなければ短いtransactionを作り、stay読取とevent読取の間のINSERT競合を避ける。

採用済み001の [`stay_movement_guard`](https://github.com/MinoruSoga/AnimalEkarte/blob/e60ba31e86f88e0e0a6a870989ac491fd8872654/backend/migrations/001_init.sql#L10388) はmovementのBEFORE INSERTで同stay行を`FOR UPDATE`する。このINSERTとの直列化は静的な対応関係であり、実DBのlock/2接続試験PASSではない。

採用済み001には [`stay_movement_events_immutable`](https://github.com/MinoruSoga/AnimalEkarte/blob/e60ba31e86f88e0e0a6a870989ac491fd8872654/backend/migrations/001_init.sql#L10529) のBEFORE UPDATE OR DELETE→`contract_history_immutable`も存在する。DDL上の原event不変条件は確認済みで、実DB適用/受入は未検証。

**未確定Q4**: 499全writerの取得順と、461のhospitalization/cage-group→stay行の順との整合。現reader注入だけで訂正writer全体の安全性を認定しない。

## 現在の未完範囲

通常caseのreaderをcompositionへ注入し、nil依存だけで全stayを常時503にする構成は除去する。一方、Q1–Q3の対象stayはunknownであり、対象ケージの空き確認を止める。候補一覧にunknownが混じる場合も部分集合を空きと偽らない。

hospitalization create/update/swap等の旧status/type/start/end利用の499 writer接続、必要認可・監査・receipt、PR740のknown/unknown・current_assignment等へのruntime追随、実DB/API/UI受入は別途必要。DDL変更の必要性はQ1/Q2の意味・表現が確定してから446/466へ具体差分を渡す。今回はDDLを追加しない。
