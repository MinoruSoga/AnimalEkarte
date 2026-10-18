# 入院・ホテルの用語と関係

文書確認日2026-10-15。既存語彙と個別合意、2026-10-15追加承認の訂正・料金方針を整理する。人数や新しい指名UIを推定しない。[業務正本](../domains/hospitalization/lifecycle.md)・[技術契約](../../design/domains/hospitalization/technical-spec.md)。

| 概念 | 意味 | 関係・境界 |
|:---|:---|:---|
| 臨床入院 | 診療のための入院管理 | 予定入院予約／診察後決定の両入口。同じ入院へ合流 |
| ホテル宿泊 | 預かり・宿泊管理 | 通常診療の指名／シフト／同日制約と別。実滞在の頭数管理 |
| 入院／宿泊予定 | 将来の受入予定 | 実際の入院／入室開始ではない |
| 滞在実績 | 実際の受入開始から退院／退室まで | 予定終了だけでは終了しない。予定予約と同じ宿泊は二重計上しない |
| 入院開始 | 来院後スタッフが担当医決定済みで明示確定 | 入院記録作成と整合する。通常カルテ作成とは別 |
| 指名医師 | 予約入口での指名 | 最新回答で指名は非必須。担当医師・記録actorと同義ではない |
| 担当 | 最新ユーザー訂正で担当医師を意味する | 非医師の責任担当1名必須案は不採用。宿泊管理の担当と実ケア入力actorを分ける |
| ケア記録actor | 実際に情報を入力した認証スタッフ | 担当医師／実施者と自動同一視しない。現行日次ログに実施者項目なし |
| 治療プラン | 診療内容・金額明細の計画。入院中は追加・変更・中止できる | 元内容・変更履歴を保持し現在内容／中止と区別。継続ケア・master・実施記録とは別。完了後は専用訂正 |
| ケアプラン項目 | 投薬・処置・給餌等の継続ケア項目 | 変更可能な独立子。itemだけ手入力その他を許可 |
| 日次記録 | 日付別のバイタル・ケアログ・スタッフメモ | ケア項目の自動実施記録ではない。日次入力を通常カルテ確定としない |
| 入院・宿泊プランmaster | 体格・per_day/per_night・参考単価等の定義 | 現行退院会計の単価／日数積算では参照しない |
| ケージ／連結グループ | 物理的な収容場所と代表・構成員 | 一階層。頭数上限とは別。未割当の滞在も存在 |
| 退院／退室 | 実滞在終了の明示状態遷移 | soft-delete、会計作成、会計確定、返金とは別 |
| 延泊希望／調整待ち | 変更案と超過を解消する調整状態 | 実滞在を残し、上限内を再確認して確定。状態のDB表現は設計 |
| 退院会計 | 入院参照付き会計とケア由来明細 | 会計ID・確定履歴はD10所有。治療プラン概算と同一ではない |

予約ID・hospitalization ID・通常medical_record ID・日次record ID・billing IDを区別。多重度の新たな推定はしない。同じペットが同じ日に別の滞在を持つ可能性を日付一致だけで排除しない。旧HospStatus reserved/admitted/dischargedと、提案／通信不明など操作状態を混同しない。

予約領域の同時宿泊・区間・実未退室の正本は[LINE仕様](../line/reservation-spec.md#hotel-counting-boundaries-1826)、入院開始は[予約区分](../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-inpatient-start-confirm)。担当の最新修正は業務正本の2026-10-15訂正を読む。

## 追加承認の語彙

- 死亡後の生前分処理: 死亡状態を保持したまま、生前診療の後日入力・誤記訂正、既存滞在の退院処理・D10精算を理由と履歴付きで行う。死亡後の新たな入院実施や自動返金・在庫戻しではない。

- 理由付き専用訂正: 完了後の原記録を保持し、理由と変更履歴を関連づける操作。自動返金・在庫巻戻しを含まない。
- 必要前金: スタッフが個別確認入力する必要額。D10で受領した残高とは別。
- 取消料: スタッフが個別確認入力する料金。予約取消だけで自動請求・返金しない。

<a id="stay-identity-and-time-concept"></a>
## 滞在の同一性・時間・日次参照

一つの実滞在は対象ペットと医院に対応する。予定入院と診察後決定は同じ入院へ合流する入口で、予約ID/通常カルテID/入院ID/日次記録ID/会計IDを同一にしない。正確な予約↔滞在↔請求の物理多重度が未記載なら未選定のまま保持し、全入口を1:1外部キーと固定しない。

| 概念／関係 | 意味・時間 | 正本・受入への追跡 |
|:---|:---|:---|
| 予定受入/退院と実滞在 | 予定区間と実際の開始/終了は別。未退室は予定超過後も滞在中 | [FR-hospitalization-hotel-stay-capacity](../domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-hotel-stay-capacity)、ホテルは頭数/全期間/二重計数防止を継承 |
| 入院開始→入院記録 | 来院後担当医決定済みで明示開始、記録作成と整合。失敗は開始成功としない | [FR-hospitalization-entry-and-start](../domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-entry-and-start)、受入は同書[hospitalization-acceptance](../domains/hospitalization/lifecycle.md#hospitalization-acceptance) |
| 入院治療プラン/ケアプラン/日次記録 | 診療明細計画/継続ケア項目/日付別実績入力という独立概念 | [FR-hospitalization-distinct-plans-and-records](../domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-distinct-plans-and-records)。変更を実施・金銭・在庫へ自動変換しない |
| 実退院日時 | 実際の臨床入院終了日時。予定退院・操作入力日時・支払日時とは別 | [FR-hospitalization-discharge-and-billing](../domains/hospitalization/lifecycle.md#requirement-fr-hospitalization-discharge-and-billing)。退院と会計作成/確定/入金は分離 |
| 退院情報の日次projection | 実退院のJST暦日[Dの00:00,翌日00:00)に属する退院日時・飼主名・ペット名・入院日 | [実日正本](../domains/cash-register/business-rules.md#closing-discharge-actual-jst-day-20261017)、AC-closing-discharge-actual-jst-day-interval／AC-closing-discharge-calendar-day-not-closing-period／AC-closing-discharge-four-fields。予定日や金銭営業日で代用しない |

退院情報の表示は金銭計上・請求・件数を変更せず、ホテル退室やD05独立帳票へ対象を拡張しない。退院情報の行粒度・固定版の保存方式は設計で、未知の日時を当日で埋めない。
