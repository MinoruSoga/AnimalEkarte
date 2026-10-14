# 予約の未承認事項・設計／移行／検証の残り

**確定業務仕様ではない。** 18:26に計数・入退室日時・分類跨ぎ変更・未来院再受付の境界は[一括是正](../../../spec/domains/reservations/README.md#mece-corrections-1826)で合意済み。以下は移設時点の記載を保持した参照で、解消済み境界を再び業務未決へ戻さない。残る具体的訂正・延泊解消操作・技術方式等は区別して扱う。 [確定索引](../../../spec/domains/reservations/README.md#approved-reservation-scope)・各FR・[オントロジー正本](../../../spec/ontology/reservations.md)を優先する。ここには既存の未決記載、UI・保存・計算等の詳細設計、コードの既知ギャップ、調査／移行・未実施受入を区別して保持する。既合意を変える案は個別承認まで未承認のまま。今回の分離だけで権限・専用記録・対象外領域の要件を確定しない。

業務5点は Sentinel_c713c4fdc74c819182979c4b8d2fef07 で合意済み。[最新状態](../../../spec/domains/reservations/README.md#five-business-decisions-current)を優先する。下記の訂正主体・会計境界・その他逆遷移・延泊超過運用・補完担当の旧い未決記載は、当時の参照であり現在の業務未決ではない。技術方式、資料／データ照合、関連内部は引き続き未確定。

| 種別 | 読み方 |
|:---|:---|
| 未承認・対象外の追加改訂 | 新しい業務結果を許可しない。既存契約を維持し、変更する場合だけ承認を確認 |
| 詳細設計・実装ギャップ | 確定FRを満たす方式の検討／実装確認。新しい業務質問へ細分化しない |
| 調査・移行・実機検証 | 未確認の実在・件数・実装状態を推測せず、既存情報を保持して検証 |

<a id="reservation-details"></a>
## 予約操作・引継ぎの未決／設計・調査

移設元：[正本](../../../spec/domains/reservations/pages/reservation-management.md)。以下は移設時点の未決／設計・ギャップ記載を保持したもの。

### 未決・未確認

以下には業務判断・実装設計・既存データの調査・未実施の受入が混在する。ユーザーとの確認は[予約ドメインの有限一覧](../../../spec/domains/reservations/README.md#remaining-business-decisions)で整理し、合意済みの再確認や実装の細分化を追加質問へ転用しない。元の未決記載は決定済みへ格上げしない。

- 予約区分の親子分類、正確な色値、名称履歴の取得・補完方式（未来への日時変更時の現名称使用・履歴保持は17:17合意）、グループ状態とactiveフラグの実装関係（無効中の個別編集は17:17合意で可）、LINE固有編集導線・他の状態条件、未設定duration、既存予約編集での再計算。グループ無効時の新規利用停止と未来日を含む既存保持は合意済み。
- 医院別対応可否は既合意、一医院勤務終了は担当者マスタの医院別指名停止契約で確定。具体的医院別設定・保存方式、未登録検出方式・未登録検出の対象母集団・他院勤務情報の取得方法、通常予約の勤務開始／終了の端点・日跨ぎ勤務の具体的保存・表示方式。休憩端点の接触／正幅重複は21:52合意済み。医院間勤務時間の重複拒否は[22:44合意](../../../spec/screens/24-shift-calendar.md#requirement-fr-shifts-cross-clinic-overlap)による。移動時間は[22:57の手動確保合意](../../../spec/screens/24-shift-calendar.md#requirement-fr-shifts-manual-travel-gap)による。固定1時間を必須にしない。担当可能者は個人対応区分＋シフトへ共通化し、職種は分類として残す。旧紐付け職種からの情報を失わない移行は未決。
- 区分混在・部分重複を含む対応能力判定アルゴリズム。合意例以外を推定しない。
- 誤登録訂正等の特別フロー、通信断で結果不明時の整合方法（実装判断）、他の逆方向・遷移・誤終了取消の具体的ロール・終了の保存失敗／訂正手順・会計完了の定義、診療前の誤取消復元の実装設計、その他の保持カルテの表示・操作。別医師への変更は23:43合意で診療終了まで可、終了後の通常操作では不可。通常操作での日時・予約区分変更と取消は23:21合意で不可、メモ編集は可。LINE固有条件は17:23合意のLINE正本へ従い、ホテル受入基準は17:35合意、計算・保存方式は設計。 [12:22合意の判定時間・延長警告](../../../spec/domains/reservations/pages/reservation-management.md#handoff-time-and-extension)へ従う。
- 患者の重複・名前未登録時の識別、来院後の任意情報補完、宿泊期間連携の詳細、休診日を時間外に含めるか、通常保存の結果不明・再送。
- 終了説明と空白選択のタッチ方式、担当者を未定へ戻す詳細、警告確認と変更確認の具体的画面。
- 権限変更、受付直後と診療開始の境界、既存カルテの引継ぎ・予約日時変更時のカルテ日付整合、来院区分の引継ぎ。
- シフト／担当区分の既指名予約保護における17:17合意の未終了予約保護を満たす時刻・状態の照合方式・終了の保存失敗／訂正手順・影響時間幅の計算詳細。現在診療中は開始済みでも検査対象、過去対応済み・取消済・未来院の履歴は検査外。データ不整合救済と未登録日の列構成、LINE6点は17:23合意、ホテル受入基準は17:35合意、計算・保存方式は設計、具体的UI・保存等は設計。

2026-10-14 00:18 JSTの[開始・終了の明示確定](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-manual-consultation-state)は合意済み。予定時刻による自動遷移は採用しない。開始は10:07 JST合意の[当日カードモーダル「カルテ作成」](../../../spec/workflows/reservation-to-record-flow.md#consultation-start-card-modal)が兼ねる。終了は[10:12合意の同じモーダル「診察を終了する」](../../../spec/workflows/reservation-to-record-flow.md#consultation-end-card-modal)で確定。確認付きドラッグ終了は[10:43合意](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-confirmed-drag-consultation-end)で決定。誤終了取消は[11:56契約](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-undo-consultation-end)で決定。その他の逆方向・遷移、取消の具体的ロールは引き続き未決／要確認。引継ぎ後の担当時間による判定と終了までの指名なし変更・終了後の通常変更不可は11:15合意済み。未作成時の作成・作成／読み込み失敗時の開始未確定・理由表示と再試行は[10:59合意](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-start-record-failure-retry)で決定し、作成済みカルテの再利用・重複なしは[10:17合意](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-checkin-record-reuse)で決定した。予約ドメイン全体は未完了。 [12:22合意の判定時間・延長警告](../../../spec/domains/reservations/pages/reservation-management.md#handoff-time-and-extension)へ従う。

- Windowsスクロール・実DB並行要求など、既存の別途受入は未確認のまま。

<a id="master-dependencies"></a>


<a id="workflow-gaps"></a>
## 横断フローの設計・実装ギャップ

移設元：[正本](../../../spec/workflows/reservation-to-record-flow.md)。以下は移設時点の未決／設計・ギャップ記載を保持したもの。

### 7. Deferred / known gaps

- available-staffs endpointは未導入。[予約時間全体の勤務・休憩適格性](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-duty-staff-selection)、[未来・現在診療中の指名予約の変更保護](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-future-assignment-protection)へ現実装は未追従／未確認。単日の勤務者絞込みを全時間判定の代替にしない。
- LINE日付判定の職種guardが現実装に残り、個人対応区分＋シフトの共通基準へ未追従。職種データの廃止・移行は未決で、情報と個人設定を保持する。職種分類そのものを廃止する合意ではない。

- 旧excluded名のAPI・部品は現実装に残るが、[対応可能のみの統一合意](../../../spec/screens/settings/master-staff.md#requirement-fr-staff-positive-reservation-capabilities)では対応不可方式を廃止し、旧実効範囲を維持する変換方針。実装・DBの移行手順は未決。
- 本書は route / E2E の実行結果を主張しない。docs-only refresh では runtime を再実行していない。

2026-10-13の静的差分: 予約時の通常カルテbest-effort作成、取消後のdraft自動削除、時間経過の未来院自動変更、受付からの直接予約編集は今回の合意と異なる。[ドメイン別残差](../../../spec/domains/reservations/pages/reservation-management.md#implementation-gaps)へ集約し、本書の更新を実装完了にしない。



<a id="master-details"></a>
## 区分マスタの設計・継承境界

移設元：[正本](../../../spec/screens/settings/master-reservation-type.md)。以下は移設時点の未決／設計・ギャップ記載を保持したもの。

個別削除区分の復元操作・既存編集導線等、予約区分自体の親子分類の必要性、正確なpreset／共通標準色の値、名称履歴の取得・補完方式（未来への日時変更時の現名称使用・履歴保持は17:17合意）、グループ状態と区分activeの実装関係（無効中の個別状態編集は17:17合意で可）、LINE固有編集導線・特別訂正など残る状態条件、既存所要時間未設定の判定・担当／設定時期・導線の詳細、情報を失わない旧職種データ移行、LINE6点は17:23合意、ホテル受入基準は17:35合意、計算・保存方式は設計、具体的UI・保存等は設計。グループ無効時の所属区分新規利用不可と既存保持は合意済み。 [12:22合意の判定時間・延長警告](../../../spec/domains/reservations/pages/reservation-management.md#handoff-time-and-extension)へ従う。

<a id="staff-details"></a>
## 担当者の設計・移行事項

移設元：[正本](../../../spec/screens/settings/master-staff.md)。以下は移設時点の未決／設計・ギャップ記載を保持したもの。

17:17合意の未終了予約保護を満たす時刻・状態の照合方式、診療終了の保存失敗／訂正手順、医院別設定・保存方式と未登録警告対象の調査、旧個人対応方式変換の検証、旧紐付け職種データの情報を失わない移行、無効区分や新グループへの追加表示は未決。個人設定を勝手に書き換えず、LINE6点は17:23合意、ホテル受入基準は17:35合意、計算・保存方式は設計として保持。

<a id="shift-details"></a>
## シフトの設計・調査事項

移設元：[正本](../../../spec/screens/24-shift-calendar.md)。以下は移設時点の未決／設計・ギャップ記載を保持したもの。

17:17合意の未終了予約保護を満たす時刻・状態の照合方式、診療終了の保存失敗／訂正手順、部分未登録の検出方式・未登録検出の対象母集団・他院勤務情報の取得・同一スタッフ照合方法、勤務開始／終了の端点・日跨ぎ勤務の具体的保存・表示方式、院内休診日警告の表示設計・LINE合意への実装追従、入力案保持の具体的方法、LINE6点は17:23合意、ホテル受入基準は17:35合意、計算・保存方式は設計、具体的UI・保存等は設計。


<a id="end-time-explanation-proposal"></a>
## 終了時刻説明の未承認表示案

移設元：[予約画面の終了時刻説明](../../../spec/domains/reservations/pages/reservation-management.md)。以下の表示方式案は未承認のまま保持する。

説明をホバーのみへ依存させずタッチでも読めるようにする案は親の提案であり、表示方法の承認は未確認。
