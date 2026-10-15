# 予約ドメイン技術仕様

| 管理項目 | 内容 |
|:---|:---|
| 文書ID／役割 | design.reservations.technical／確定業務を実現する技術設計 |
| 作成根拠 | 2026-10-14、ユーザー Sentinel_4090fd2153c881918e1da545da9ab505：design配置と各フォルダの説明文書を承認 |
| 補完根拠 | ユーザー Sentinel_971a1312ff3481919fddad08aca44be1：レビュー6件とtransaction参加条件の文書補完・再レビューを承認。業務ルール追加や実装の承認ではない |
| 業務5点の根拠 | Sentinel_c713c4fdc74c819182979c4b8d2fef07：訂正・会計境界・その他逆遷移・延泊超過・旧マスタ補完を承認。技術方式の詳細全部は承認対象外 |
| 技術4方針の根拠 | 2026-10-14 20:43 JST、Sentinel_570e060eda888191b8aeea7b5653eb07「ok」：固定指名を先に考慮する仮割当判定、操作別API／版／再送キー、受付・記録の同一transactionと確定後通知、切替済み経路の旧write停止を承認 |
| LINE判定の訂正根拠 | Sentinel_08ce4d91de608191a05c30220331a832：LINEは実診療状態をチェックしない。個別編集許可 Sentinel_2334916604f88191830b7a194c74bf85／docs限定 Sentinel_61b62c3381b88191ad0d5af32bdcffb5。§4.2の予定枠基準へ反映 |
| 対応表の文書化根拠 | Sentinel_087e713d04548191b371bfdc4e19dbea「ok」：直前の読み取り調査に基づくAPI・保存項目・切替対象と他ドメイン参加条件のdocs限定具体化。現行API・DB・設定の更新や利用開始承認ではない |
| 全境界具体化の根拠 | Sentinel_f19ab03e8c2c81919c6f589bcf563811：関連domainを含めdocs内の技術契約を一括具体化。新業務判断は追加せず、実装・設定・公開は行わない |
| 当面の保存・廃棄の根拠 | Sentinel_9d85a7955b608191985bb02c2c9356a5：当面自動削除せず、具体的保存期間／削除手続きは全ドメイン共通で別途確定。法的年数・適合確認の承認ではない |
| 基準コード | ローカルmain cea3f94ddeba10632ea372afcbe5d89efedfee36。仕様文書の既存未コミット差分を含めて照合 |
| 状態 | 予約接続の技術契約は§3.2・§4.7・§5.1・§6.4へ選定。当面の共通保存方針は承認済み。具体的期間／削除手続き・関連owner内部文書・実装前検証は§9で分離。実装・実DB・実機検証は未実施 |
| 管理先 | 技術設計は本書、作業・依存関係・担当・進捗はPlane、実施記録はdocs/work。タスクファイルは作らない |

## 1. 目的・正本・制約

[予約の合意索引](../../../spec/domains/reservations/README.md#approved-reservation-scope)の69固有FRと[継承参照](../../../spec/domains/reservations/README.md#inherited-reservation-references)2件を実現する方式を検討する。件数は今回参照した索引の値であり、全仕様定義数や実装済み数ではない。[語彙](../../../spec/ontology/reservations.md)、[横断契約](../../../spec/workflows/reservation-to-record-flow.md)、[18:26是正](../../../spec/domains/reservations/README.md#mece-corrections-1826)を参照し、業務ルール本文を複製しない。

Spec Kitの[導入済み技術計画テンプレート](../../../../.specify/presets/animalekarte-canonical/templates/plan-template.md)のTechnical Context、Constitution Check、調査／判断待ち、Design and Contracts、Verification Planを参考にした。Task HandoffはPlaneへの参照へ置き換え、Spec Kitコマンド・全workflow・生成スクリプトは実行していない。

技術正本は[API契約](../../../../backend/docs/api.yaml)、[DB定義](../../../../backend/migrations/)、[全体設計](../../../architecture/README.md)、[ADR-006](../../../architecture/adr/006-backend-domain-package-boundaries.md)。appointmentsのwrite ownerはreservation、勤務・スタッフはstaff、既存入院処理はmedicalrecordにある。独立した二重write ownerを増やさず、consumer側の最小intentと同じtransactionで接続する。[臨床安全](../../../spec/specification.md)、医院・飼主・動物・スタッフの分離、既存認可・監査、[カルテ保存契約§2](../../../spec/domains/medical-records/lifecycle.md#hybrid-sync)を保持する。88ルート対応表は改訂しない。全ドメインのドキュメント完了が実装着手の前提であり、本書の補完・文書検査・静的レビューだけで実装を開始しない。

## 2. 現在コードとの対応・差分

「存在確認」はコード経路を読んだ意味で、適合・実機PASSではない。「差分確認」はその経路に合意と異なる動作がある。「未確認」は全経路で未実装と断定しない。

| 領域／正本 | 限定的な現在実装の根拠 | 差分・未確認 | 本書の設計箇所 |
|:---|:---|:---|:---|
| [動的区分と固定分類](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-business-routing) | [モデル](../../../../backend/internal/model/reservation_type.go)、[request](../../../../backend/internal/reservation/reservation_type_request.go)：general/trimmingの2値。親区分・グループは別項目 | 4業務分類への接続差分。名称からホテル・入院を除外する[handler](../../../../backend/internal/reservation/reservation_handler.go)も存在 | §3、§5 |
| [所要時間・未設定](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-duration-range) | [作成](../../../../backend/internal/reservation/reservation_type_service_core.go)：省略時15分、分類省略時general | 必須・5〜480分・5分刻み・未設定選択不可との差分。未設定データの実在・件数は未確認 | §3、§7 |
| [削除／復元・有効意図・名称履歴](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-delete-restore-history) | [区分削除](../../../../backend/internal/reservation/reservation_type_service_core.go)は子区分・予約使用ありを拒否。[グループ削除](../../../../backend/internal/reservation/reservation_type_group_service.go)は所属区分ありを拒否 | 復元可能非表示、所属保持、個別意図、名称履歴への適合未確認。既存の区分親子機能は勝手に廃止しない | §3、§7 |
| [個人・医院別対応と職種gate廃止](../../../spec/screens/settings/master-staff.md#requirement-fr-staff-shared-reservation-eligibility) | [能力モデル](../../../../backend/internal/model/staff_reservation_capability.go)はclinic/staff/typeを持つ。[validator](../../../../backend/internal/reservation/reservation_staff_capability_validator.go)は能力を確認。肯定形APIと旧excluded APIが併存 | 旧[画面](../../../../frontend/src/features/master/components/StaffSidePanelSections.tsx)に対応不可表現。院内経路の有効性・勤務・休憩等を共通化する必要。全経路監査は未実施 | §3、§4 |
| [医院別停止・既指名保護](../../../spec/screens/settings/master-staff.md#requirement-fr-staff-clinic-departure-booking-guard) | [所属service](../../../../backend/internal/staff/staff_clinic_assignment_service.go)と[関連テスト](../../../../backend/internal/staff/staff_clinic_assignment_reservation_race_test.go)がある | 所属解除の安全機構を再利用検討。未終了予約一覧、手動解消、B院継続、過去履歴保持の合意適合は未検証 | §3、§6 |
| [シフト・日跨ぎ・変更保護](../../../spec/screens/24-shift-calendar.md) | [service](../../../../backend/internal/staff/shift_entry_service.go)は更新時の行lock・休憩置換、削除は既存行確認後に削除。[共通検査](../../../../backend/internal/sharedkernel/validators.go)も参照 | 読んだ更新／削除経路に予約影響検査はない。医院間勤務重複、日跨ぎ日時、休診日警告、入力案保持は未検証 | §3、§4、§6 |
| [最小入力・終了保持・未定・院内警告](../../../spec/domains/reservations/pages/reservation-management.md#operations) | [フォーム](../../../../frontend/src/components/shared/ReservationFormModal/ReservationFormModal.tsx)は新規電話・ペット名・主訴を要求。[保存hook](../../../../frontend/src/features/reservations/hooks/use-reservation-save-actions.ts)はdoctor未指定を既存値へフォールバック | [pet作成request](../../../../backend/internal/pet/pet_request.go)はname必須、保存hookはowner→pet→予約を順次作成し途中IDを保持する。画面だけの任意化では不十分。明示未定への変更を欠落と区別する必要。カレンダー・終了保持・同時編集は再検証対象 | §4、§5 |
| [院内重複の警告確認](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-staff-overlap-warning) | [migration](../../../../backend/migrations/001_init.sql)にappointmentsのdoctor時間範囲EXCLUDE制約と同時開始一意制約 | 合意の院内重複許容とDB拒否が衝突。実DBで適用済みの制約状態は未確認 | §3、§7 |
| [受付とカルテ同時成功](../../../spec/workflows/reservation-to-record-flow.md#checkin-contract) | [handler](../../../../backend/internal/reservation/reservation_handler.go)はconfirmed/checked_in後にbest-effort作成、取消後draft削除。[lifecycle](../../../spec/domains/medical-records/lifecycle.md)の現行参照も照合 | 予約時作成・取消削除の差分、受付と記録の原子性が必要。現在のID・row lock・一意性は保全 | §5、§6 |
| [開始／終了・引継ぎ・誤終了取消](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-manual-consultation-state) | [更新service](../../../../backend/internal/reservation/reservation_service_update.go)、[受付アクション](../../../../frontend/src/features/reception/components/ReceptionDialogActionButtons.tsx)に状態操作がある | 合意モーダル・確認付き終了・理由記録・支払い境界・残り時間判定の全適合は未確認。汎用status変更だけで成立とは扱わない | §4、§5、§6 |
| [手動未来院・取消復元・分類跨ぎ](../../../spec/domains/reservations/README.md#mece-corrections-1826) | [intent repository](../../../../backend/internal/reservation/reservation_intent_repository.go)は4時間経過のMarkNoShow、[Lステップbatch](../../../../backend/internal/lstep/lstep_batch_noshow.go)から呼ぶ。汎用no_show更新を拒否 | 自動未来院を止め、手動intent・再受付再確認・取消復元・分類跨ぎ境界が必要。旧ロック／確定カルテ保護は保全 | §5、§6 |
| [入院・ホテル](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-hotel-inpatient-entry) | [モデル](../../../../backend/internal/model/hospitalization.go)はhotel/hospitalizationとdate期間・担当医。[service](../../../../backend/internal/medicalrecord/hospitalization_service.go)は未指定statusを開始日でadmitted/reservedに既定化 | 予定入院と明示開始・担当医・開始時記録・失敗時未確定との接続差分。任意担当医師、宿泊同一性・頭数上限・入退室時刻対応は未確認 | §3、§5、§6 |
| [LINE公開・厳格受付・ホテル](../../../spec/line/reservation-spec.md) | [作成service](../../../../backend/internal/reservation/liff_service_reservations.go)はValidateAndCreate。[型](../../../../frontend/line-reserve/src/types/models.ts)はcategory2値。[availability](../../../../backend/internal/reservation/liff_service_availability_staff.go)はactive/公開をfilter | 4分類、医院別宿泊上限、全期間再判定、勤務の保存時再確認、受付前取消・入力保持の合意適合は未検証 | §4、§5、§6 |
| [区分混在・部分重複の対応能力](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-capacity-warning) | [現在の計数](../../../../backend/internal/reservation/reservation_capacity.go)は区分・同開始時刻件数とMaxConcurrent中心 | 指名済み占有・未定予約・個人能力・勤務／休憩を区分間で照合する§4.2の仮割当算法への変更が必要。現行計数を新仕様の成立根拠にしない | §4.2、§6、§8、§9 |

## 3. データモデル案と変更理由

以下は提案設計。現行の列・型・FKの確定定義ではなく、承認された業務を満たす変更案である。具体的DDL・API wire更新は実装時に正本でレビューする。

| 業務事実 | 現在の格納・制約 | 提案する技術上の表現／保持条件 | 判断理由 |
|:---|:---|:---|:---|
| 予約区分と業務分類 | reservation_types.categoryは2値、name/group/parentは別 | 固定4分類の明示項目を追加する案。移行期は旧categoryと独立に扱い、確認済みmappingだけを保持。未設定を名前から推定せず新規選択を拒否 | 旧general/trimmingへの依存を一度に破壊せず、動的サービスと固定接続先を分離 |
| 所要時間 | int・既定15分 | 入力必須／範囲／刻みのserver検査とUI整合。移行前に既存default・未設定表現を調査し、設定済みと推測しない | 既定値で未設定を隠さず、既存予約終了値を保持 |
| 有効・削除・所属 | 個別is_active、group、DeletedAt、parent | 個別有効意図・所属group・削除／新規利用停止を区別して復元可能に保持。所属変更だけで停止を解除しない | 削除・復元・利用再開と履歴を分離。詳細flagは§3.2で選定 |
| 名称の履歴 | 現行master名と予約関連 | §3.2の名称revision／予約snapshotを選定。未来へ日時変更した際の名称と変更履歴も追跡 | 閲覧時点で過去名を巻き戻さず、補完不能データを捏造しない |
| 医院別能力・指名停止 | staff_reservation_capabilities、clinic assignment、スタッフ有効状態 | 既存医院別能力を基礎に、医院別指名停止と全体無効化を区別。担当履歴は現在の選択可否から分離 | B院勤務を継続し、過去記録の担当参照を保全 |
| 日跨ぎ勤務・休憩 | date＋start/end時刻と複数break | JST境界を明示できる開始／終了日時へ正規化する表現案。終了日・休憩跨日を明示し、同一人物の医院間正幅重複を検査 | 同日通常予約と日跨ぎ勤務を別概念にし、デフォルト勤務の推測を避ける |
| 予約・カルテ・履歴 | appointmentsとlinked medical record、既存一意性・lock | 予約IDと作成済みカルテIDを維持。取消・終了訂正は状態と操作履歴、理由は既存監査の目的別payloadへ | 再作成・自動削除を使わず、一つの予約と業務履歴を守る |
| 同時更新 | updated_at、row/advisory lock等が経路ごとに存在 | 変更対象のversionとintentの再送識別を明示。省略／明示null／値を区別するDTO | 未定への変更を「未変更」へ戻さず、stale入力を無条件に上書きしない |
| 宿泊同一性・担当・期間 | hotel/hospitalization、start/end date、doctor_id、daily_records | 同一宿泊を特定する関連を予約・宿泊管理で保持。予約指名／入院担当医／ホテル任意担当医師を別概念とし、予定・実入退室の日時対応を検討 | 1宿泊を二重計数せず、hotelを医師能力から推測しない。入院記録を既存daily recordへ無確認で同一視しない |
| 医院別ホテル上限 | 今回の読取で対応する設定項目を特定できていない | ホテル同時宿泊頭数の設定を医院単位で保持。予約区分のMaxConcurrentや空きケージ数を流用しない | 診療人数・ケージ・宿泊頭数の業務上の相違を保つ |
| 予約重複制約 | DBのEXCLUDE／同時開始一意制約 | 院内の警告確認許容とLINEの拒否を満たす制約再設計案。対象制約・consumer・既存データを調査し、最小変更をレビュー | DB制約だけを削除して整合性を弱めず、LINE競合防止をtransactionで維持 |

<a id="reservation-storage-matrix"></a>
### 3.1 保存項目・責務の対応表

以下は将来の保存設計。論理名を具体化したもので、実DBの列・tableを追加済みではない。migration／生成型は今回変更しない。業務5点と§4.4の決定を保持し、実データを確認したかのように初期値を埋めない。

| 保存対象／owner | 保存項目案 | 制約・更新条件／残る照合 |
|:---|:---|:---|
| appointments／reservation | version（BIGINT、正値、開始値1）、既存ID・予定開始／終了・指名ID・状態・受付時刻・取消前状態と履歴への参照 | 全更新入口が同じ版を増分更新。通常編集・状態操作・他ドメインからの予約補完／完了も参加。日時を自動延長せず、予定と実業務状態は別。取消済みの復元は医院認可付き専用readで同じID・版・取消履歴を照合し、任意の削除行を復活させない |
| reservation_operation_receipts／reservation | clinic_id、actor_kind＋actor_id、intent、operation_key（UUID）、business_fingerprint、result_ids、committed_version、completed_at、response_cache | scopeのFULL UNIQUEはclinic＋actor_kind＋actor_id＋intent＋key。staffとLIFF customerの同値IDを混同しない。claimはfingerprintまで、結果ID／確定版は業務write・監査と同時commit。raw患者入力・tokenをcacheや監査へ無目的に複製しない。成功／失敗キーは§3.2の非再利用・自動purgeなし。当面自動削除しない扱いは共通方針で承認済み。具体的期間・削除手続きはU11で別途確定 |
| staff_clinic_assignments／staff | 既存医院所属・capabilities_configured_at、医院別指名停止の意図と履歴 | [既存marker](../../../../backend/internal/model/account.go)は能力の明示空選択と未設定を区別する。勤務未登録のmarkerには使わない。医院別指名停止を全体staff無効と混同せず、具体的停止列・所属履歴との対応はstaff契約で照合 |
| shift_entries・shift_entry_breaks／staff | 勤務と休憩のstart_at／end_at（timestamptz）、医院・同一staff ID、JST基準日、勤務区分 | 現行[date＋time](../../../../backend/internal/model/staff.go)へ終了日を明示できる表現を追加する設計。開始／終了と休憩は同じ時間軸。日単位識別は保持し、終日非勤務の登録と勤務未登録を区別。旧終了時刻が開始より前でも翌日と推測せず確認対象にする |
| reservation_types・groups／reservation | 明示4業務分類、所要時間の設定状態、個別有効意図、group・parent、削除／復元状態、名称履歴の参照 | 旧categoryや15分defaultを確認済み値の根拠にしない。group停止で個別有効意図を上書きしない。既存parentを保持。名称履歴／予約表示値の保存構造と旧名の確認補完は別照合が残る |
| reservation_contract_states／reservation | clinic_id、contract_kind、revision、切替状態、適用分類・更新日時 | 能力・業務分類接続・lifecycle等の契約単位で管理する案。routeだけを順番に有効化して未対応入口へ迂回させない。切替更新も§6.3の規約へ参加。設定flagやclient申告だけで利用開始しない |
| audit_logs／既存audit participant | operation key／結果ID、対象ID、actor、理由・訂正確認・変更前後・切替revision | [既存metadata](../../../../backend/internal/model/audit_log.go)へ目的別に接続。業務成功に必要な監査は同一transaction。保管・権限と具体的payloadの境界はaudit ownerと照合。履歴を消して版・再送キーを再利用しない |

予約versionのwireは正の整数を正確に扱い、client側で精度を失う型へ変換しない。HTTP DTO／生成型は§4.7で十進文字列を選定、API正本反映は未実施。owner／petの空欄は§4.1のconsumer照合が必要で、この表だけで共有APIの必須解除を確定しない。

<a id="reservation-storage-details"></a>
### 3.2 名称・利用停止・保存方針の選定

Sentinel_f19ab03e8c2c81919c6f589bcf563811によるdocs限定の技術具体化。以下は設計上の選定で、現行DBの変更・業務権限拡張・法的保存年数の決定ではない。

| 保存対象 | 選定する契約 | 理由／停止条件 |
|:---|:---|:---|
| 区分の利用可否 | 個別有効意図、削除状態、group有効／削除、明示再開待ちを別事実として保持。無効／削除groupから所属変更する際は再開待ちを立て、所属変更だけでは解除しない。既に利用可能な区分の通常所属変更では新たに立てない | group再有効化はその所属の個別有効意図を尊重。区分復元は個別無効、group復元は子区分を一括復元／有効化しない。明示有効化で再開待ちを解除してもgroup等の他gateは残る |
| 名称revisionと予約snapshot | master ownerが名称revision、変更確定時刻、前後名を保持。reservation ownerが予約表示名snapshot・採用revision・出所を保持。同じgate内で改名確定時刻より予定開始が厳密に後の予約だけsnapshotを更新する | 等しい／過去開始は旧snapshot。閲覧時に現masterへ結合して過去名を変えず、後から旧revisionへ巻き戻さない。未来へ許可された日時変更時は現名称を採用して前後snapshotを監査 |
| 名称の旧データ | 復元できる既存値だけ移行し、確認不能は出所unknownとして保持 | 架空の過去名を補完しない。新規・過去日時への登録時もその登録で選択した現名称をsnapshot化し、その後の改名対象は上記開始境界で決める。大量未来予約の同期更新コストは実測対象 |
| 再送対応 | §3.1のreceipt storeにscope・キー・fingerprint・成功有無・結果ID・版の最小対応を保存。未成功も含め自動削除／キー再利用をしない。response cacheは保存必須ではなく、最小対応と現業務IDから結果を再構成する | cache TTLを設計の未決にしない。再構成不能ならresult_unavailableを返し新規作成へ落とさない。キーbindingを消すpurgeは本契約外 |
| 履歴・監査・個人情報 | 名称履歴、訂正、会計確定歴、宿泊関連、キーbindingを業務履歴として保護し、本文・個人情報の重複を最小化。fingerprintはserver秘密鍵付きの正規化HMAC、鍵versionを保持し旧鍵照合を維持する | 当面自動削除しない扱いは[共通方針](../../../spec/shared/README.md#interim-retention-policy)で承認済み。具体的保存期間・削除手続きは[U11](../../../delivery/DELIVERY_PACKAGE.md)で別途確定。年数を推測せず、法的無期限保存・適合確認済みとしない。廃棄要求が確定した際は非再利用tombstoneと結果照会の整合を再設計する |

## 4. API契約案と画面連携

現行wireの正本はbackend/docs/api.yaml。[予約route](../../../../backend/internal/reservation/routes.go)・[snapshot](../../../../backend/internal/reservation/routes_snapshot_test.go)、[スタッフroute検査](../../../../backend/internal/staff/routes_test.go)、[入院route](../../../../backend/internal/medicalrecord/routes_hospitalization.go)を参照する。将来の新URL・HTTP method・既存grant対応は§4.5・§4.7で選定し、API正本反映は未実施。

| 契約境界 | 現在の入口／利用画面 | 提案契約・処理 |
|:---|:---|:---|
| 区分・グループ管理 | masters/reservation-types、reservation-type-groups | 分類・所要時間・有効意図・復元・名称履歴の検査。未使用変更と未設定補完を別intentにし、保存時も既存予約を照合 |
| 医院別能力・停止 | capable-reservation-typesと旧excluded、staff clinic管理 | 肯定形の読書きを統一し、旧APIのcompatibility期間と解除条件を§4.3で設計。新規区分を旧負リストの反転で自動許可しない。保護予約の最終照合と影響一覧を返す |
| 勤務と候補照会 | shifts、on-duty-staffs、予約フォーム | 日単位勤務者一覧を全時間適格性の代用にしない。医院・区分・予約時間または引継ぎ時点で候補と判定可能／情報不足を返す |
| 新規患者と予約の同時登録 | owner／pet作成、予約フォーム | 最小入力のserver契約、空欄表現・既存consumer互換性、部分成功後のID再利用を§4.1で設計 |
| 対応能力照会 | カレンダー警告、LINE空き／確定 | 指名済み占有と未定予約を、区分間共有担当資源・部分重複へ照合。採用算法・未確認条件と停止境界は§4.2 |
| 通常予約保存 | reservations・予約管理／ドラッグ | 適格性と状態境界をserverで再確認。院内の許容警告と保存禁止を別codeとして返し、確認対象の入力・versionと対応付ける |
| 受付・再受付 | 受付カード、予約管理の既存入口 | 同じ予約のcheck-in intentと記録作成／再利用を同一transactionへ。未来院は再確認・必要手動修正後に受付 |
| 診療開始／終了・終了取消 | 受付カードモーダルと確認付き終了drag | start/end/undoを意味のあるintentへ分離。開始時読込失敗は開始未確定、undoは支払い境界・権限・理由をserverで確認 |
| 取消／復元・日付／分類変更 | 予約管理 | 通常編集、誤取消復元、受付後の記録確認訂正を区別。受付後の分類跨ぎを汎用PATCHで迂回しない |
| 入院開始・宿泊期間変更 | 入院／ホテル管理 | 予定入院・診察後決定の関連を照合し、臨床入院の開始と担当医・記録を同時確定（ホテル担当医師は任意）。延泊は変更後期間の再判定。超過時は実績保持・責任者調整後確定・初期例外なしの合意へ |
| LINE公開・空き・確定・取消 | LIFF既存入口／公開予約UI | 公開・設定・本人・医院を照合。診療／trimmingの適格性とhotel全期間計数を分類でdispatch。表示と保存時の双方で条件確認 |

提案DTOは対象ID、変更後入力、期待version、必要な確認意思・理由、再送識別を目的別に分ける。認証から医院・actorを決定し、bodyのclinicやownerを認可根拠にしない。成功は同じ予約・記録IDと保存後状態を返す。失敗は入力項目エラー、禁止理由、影響予約一覧、競合・結果不明を区別し、client入力を保持して再取得・比較・再試行へ導く。wire・grant・再送保持の選定は§4.7・§3.2へまとめ、API正本への実装反映は別。

画面では新規飼主名・動物種と任意補足を区別。開始＋所要時間の自動終了と手動終了保持は入力由来を管理して実現する。指名解除は明示null等で表し未送信と混同しない。警告確認tokenを再利用する場合は予約内容・version変更後に無条件流用させない。列・15分空白選択・ドラッグ確認・10色preset／詳細RGB・group色継承は[画面正本](../../../spec/domains/reservations/pages/reservation-management.md)と[マスタ正本](../../../spec/screens/settings/master-reservation-type.md)へ照合する。presetの正確な色値を推測しない。

### 4.1 新規患者の最小入力と部分成功

[最小入力の正本](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-new-patient-minimum)を予約専用登録intentで実現する方式を選定した。飼主名・動物種必須、電話・pet名・主訴任意。pet名は空文字＝未記録、共有APIの必須は保持し専用DTOとowner service intentで扱う。既存IDの再利用・補足変更・通信断後照会・grantは§4.7。画面のみの任意化や架空名・同名統合を採用しない。

段階ごとの成功owner／pet IDを同じregistration_attempt_idで保持する。補足変更で捨てず結果不明中は同じキーを照会／再送し、別患者への切替は先行結果確定後の別試行へ分ける。全登録の一括atomicityを新業務要件にせず、予約batch内部と受付／記録は既存atomicity契約を維持する。実consumer互換は実装前検証へ残す。

<a id="42-区分間の対応能力と部分重複"></a>
<a id="capacity-assignment-contract"></a>
### 4.2 区分間の対応能力と部分重複（技術方針承認済み）

[対応能力正本](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-capacity-warning)を実現するread/query責務をreservation側へまとめ、staff ownerから医院別能力・勤務・休憩の照合情報を取得する案。指名済み予約の占有と未定予約を分け、複数区分へ対応できる同一スタッフを区分ごとに独立した人数として重複計上しない。既存の同区分・同開始時刻件数や単純総人数だけで代用しない。

本算法は診療／トリミングの通常時間指定予約を対象とし、ホテルは宿泊上限、入院は担当医・開始条件の別判定へ接続する。20:43承認に従い、固定指名の占有を先に確定し、未定予約を適格スタッフへ仮に割当可能か判定する。判定用の仮割当は実担当・doctor ID・引継ぎとして保存しない。次を算法の設計契約にする。

1. 対象医院・対象時間に重なる予約を区分横断で取得し、取消等の計数対象・医院・有効性・医院別対応・勤務／休憩を正本どおり照合する。LINEでは下記の予定枠基準を用い、実診療の進行状態で占有時間を変えない。確定した勤務0人、全未登録、部分未登録、取得失敗を区別する。候補の未登録を勤務と推測しない。固定指名は現在の候補外でも既存占有として扱い、自動解除・他人への付替えで計算を成立させない。
2. 固定指名の重複は同じスタッフの占有時間の和集合として計数し、同じ人を複数の空き資源として数えない。固定指名同士の重複警告と、その後に未定予約を受け入れられるかの不足判定を別結果にする。院内の確認後許容をLINEの指名重複許可へ流用しない。
3. 未定予約ごとに予約時間全体で適格な候補集合を作る。半開区間の端点で占有を比較し、接触だけは重複扱いにしない。仮候補は固定指名の占有と重ならず、同じ候補に置く未定予約同士も正幅重複しない。同じ予約を各区間で異なる候補へ入替えるだけで成立としない。
4. 候補の少ない予約から決定的な順で探索し、行き止まりでは候補を戻して別組合せを調べる。貪欲に一度置いた失敗だけを不足確定にしない。時間・候補の共有が連鎖する予約も評価集合へ含め、直接重なる予約だけの切取りで外側の競合を見落とさない。相互に資源競合しない集合は分割可能。全未定予約を置ければ充足、全組合せを尽くして置けなければ不足とする。
5. 探索上限・時間制限・依存取得失敗で計算が完了しなければ判定不能とする。未完了を不足／充足の確定へ置換しない。出力は判定、登録情報の範囲、警告・禁止理由と根拠revisionであり、仮割当の保存・実担当自動決定は行わない。再計算時に異なる仮候補を使っても既存担当履歴は変わらない。

カレンダー警告・LINE空き・保存時再判定は同じ入力意味と算法を使い、保存時に最新データで再計算する。院内の警告確認・指名なし登録の既合意を維持し、LINEは拒否条件または判定不能を受入可へ変換しない。医院・区分をまたぐ共有担当資源の直列化を§6.2に接続する。診療中の引継ぎは残り時間／現在時点の正本で別評価し、この通常予約算法のために過去全体の勤務を再要求しない。

LINEの空き表示・確定時再判定は[予定枠基準の最新訂正](../../../spec/line/reservation-spec.md#line-scheduled-slot-basis)に従い、固定指名の占有・未定予約の需要を予約の予定開始／予定終了で算法へ渡す。実際の診療中／終了・予定終了超過・実診療の見込終了を取得／検査条件にしない。予定枠が空く場合は実診療の継続だけで判定不能へ落とさず、実終了済みでも予定枠の正幅重複を消さない。取消等の既存計数境界、勤務・休憩・対応区分、登録不足／取得失敗・探索中断の判定は維持する。予定時刻の自動変更や診療状態の自動遷移は行わない。ホテルの実滞在計数、院内の引継ぎと未終了指名予約の変更保護は別契約で、LINEの予定枠基準を流用しない。母集団と試験探索限度は§4.6で選定し、実取得・index・性能値は要検証で、文書上の算法決定を実機成立へ置換しない。

反例は、A指名Yの間にAだけが対応する未定X、開始時刻の異なる部分重複、途中で候補を切替えれば置けるが全時間を同じ人で置けない例、貪欲では失敗し再探索で成功する例、共有資源の連鎖、休憩端点、全／部分未登録、先行診療超過でも予定枠は空く例、実終了済みでも予定枠が重なる例、探索中断を含める。

### 4.3 旧excluded互換writeの停止・照合

[後から追加した区分は未チェック](../../../spec/screens/settings/master-staff.md#requirement-fr-staff-positive-reservation-capabilities)を互換期間にも守る。[旧変換](../../../../backend/internal/reservation/reservation_staff_repository.go)は保存時点のactive全区分からexcludedを引くため、旧画面の表示後に追加された区分を自動許可し得る。現在の変換をそのまま安全な互換writeとは扱わない。

20:43承認により、新方式へ切替済みの更新経路では旧excluded writeを停止する。旧payloadを最新全区分から反転するfallbackや、旧経路へ迂回しての更新を設けない。旧clientには方式切替の拒否理由を返し、入力を保持して対応画面の再取得へ案内する。read互換は業務事実を変更しないadapterだけを許可し、旧write再開を含めない。切替前の実効能力変換は確認した母集団を基準に照合し、切替後に追加された区分は未許可のまま。旧writeと新writeの同時要求も§6.2の切替状態照合へ参加させる。旧未対応consumerが残る場合は利用開始を止め、能力母集団revision案を旧write併存の自動許可としない。

### 4.4 操作別API・更新版・再送識別（技術方針承認済み）

以下は将来のAPI設計契約であり、現行backend/docs/api.yamlやrouteを変更済みとは扱わない。予約作成・通常編集、check-in／再受付、開始、終了、終了取消、取消／復元、未来院、引継ぎ、記録確認訂正を目的別intentとして扱い、汎用status更新から状態制限を迂回しない。推奨URL・HTTP methodは§4.5で具体化し、目的別grant・schemaを現行入口と他ドメインの契約へ照合してAPI正本へ定義する。新しい業務遷移を列挙だけで許可しない。

| 契約 | 具体化する設計 |
|:---|:---|
| 更新版 | 対象集約にserver管理の増分versionを持たせ、通常更新はexpected_versionとの比較とwrite／版増分を同じtransactionで行う。updated_atやclient時計だけを版の代用にしない。集約外の適格性・会計・区分設定もcommit前に再確認し、予約版だけで守れるとしない |
| 操作入力 | 対象ID・目的別changes・期待版・必要な理由／確認を分ける。医院・actorは認証から解決。省略は不変更、明示nullは許可された指名解除等、値は設定として検査する。新規作成は対象版を要求せず再送識別で保護する |
| 再送識別 | 同じclient操作にUUID形式のIdempotency-Keyを使い、医院＋actor＋intent＋キーをscopeとする。業務内容（対象・変更・理由等）のfingerprintと結果IDを対応付ける。同じキーで業務内容が違えば競合。同じキーの成功済み再送は再writeせず初回結果を返す |
| キーの確保・失敗後 | 認可後、scope＋キー＋業務fingerprintだけを短いtransactionで一意に確保する。既存キーの内容相違は拒否する。この最小対応は業務rollbackでも残し、成功済みとは扱わない。同内容の再試行は可能とする。業務transactionではこの対応行をlockし、同じキーの処理を直列化してから成功結果を照合する。通信断や時間切れだけで生存中の処理を追越して並行再実行しない |
| 成功の保存 | 初回結果ID／確定版へのキー対応・業務write・必要監査を同じtransactionでcommitする。失敗rollbackでは成功結果を残さず、確保済みキーは未成功のまま。待機した同内容再送は先行commitなら結果再利用、先行rollbackなら同じ内容で再実行する。既存active record一意性・業務IDlockも維持する |
| 版と再送の順 | 認可／scope確認後、成功済み同一操作の照合をexpected_version検査より先に行う。初回成功後に対象版が進んでも同じ操作を再適用しない。期待版・警告確認receiptは前提条件として業務fingerprintと区別し、失敗後の最新再照合を成功済み操作の新writeへ変えない |
| 応答・競合 | 成功は初回結果IDと確定版、必要なら別項目で現在状態を返す。初回結果と現在状態を混同して取消済み対象を復活させない。版競合・内容相違・禁止・入力不備・判定不能を別理由で返し、入力を保持して最新との差を確認し再適用する |
| 結果不明 | 通信断は失敗確定にせず、同じscope／キーの結果照会または同一内容再送で照合する。対象IDがない作成でも照会可能な契約を用意。成功済みなら同じIDで続け、内容を変える次操作は成功／失敗を確かめた後に別キーで行う |
| 保持 | response cacheのTTLと、業務IDへの最小キー対応を分ける。キャッシュ期限後も同じ操作を新規作成へ落とさず照合できるよう、業務履歴の存続中は最小対応を保持する設計。失敗／未成功キーの最小fingerprint対応もresponse cache期限で消さず、内容相違の新規操作へ転用しない。具体的保存は§3.1・§3.2、結果照会は§4.7。自動purgeなしの技術契約と業務保存年数を分離する。照合なしの削除・同じキーの新規扱いを許可せず、未決のTTLを安全な再作成許可にしない |

警告確認は変更入力・対象版・警告根拠へ結び付け、clientの確認booleanだけで最新条件の検査を免除しない。根拠が変われば再表示・再確認し、保存時のserver検査で禁止に変わった条件は拒否する。新規owner／petの段階登録は同じ登録試行へ各結果IDを対応付け、補足変更だけで作成済みIDを捨てない。キーの業務内容が変わる場合は既存IDへの訂正操作として分け、同じキーで別患者を作らない。空欄表現・既存grant・専用intentは§4.7で選定、実consumer互換は検証へ残す。

<a id="reservation-api-matrix"></a>
### 4.5 API・DTO・理由codeの対応表

Sentinel_087e713d04548191b371bfdc4e19dbeaで許可された文書具体化。下記は将来の推奨wireで、現行[API YAML](../../../../backend/docs/api.yaml)・[route](../../../../backend/internal/reservation/routes.go)・生成型へ未反映。新業務遷移やactor権限の追加ではない。通常編集・専用操作の入力schemaは閉じた項目集合とし、余分なstatus・禁止項目を黙って無視して成功扱いにしない。

| 目的 | 現行入口／推奨wire | DTO・結果／参加する契約 |
|:---|:---|:---|
| 作成／複数患者作成 | POST /api/v1/reservations、POST /api/v1/reservations/batchを保持 | Idempotency-Key＋予定枠・区分・患者・明示指名等。初期状態は業務intentで決め、任意status指定を除外。batchは各予約ID・版を結果へ対応させ、途中失敗で一部成功にしない |
| 通常編集 | PATCH /api/v1/reservations/{id}を保持 | expected_version、許可されたchanges、必要warning_receipt。statusは専用操作へ分離。指名解除はdoctor_id:null、未送信は不変更。予定終了が手動入力か自動計算かをclient draftで保持 |
| 受付／再受付 | POST /api/v1/reservations/{id}/check-in、/recheck-inを推奨 | 期待版・再送キー。同じ予約の状態とrecord ID／版を返す。受付とrecord participantは同一transaction、未来院等の既合意境界を§5で検査 |
| 診療開始／終了 | POST /api/v1/reservations/{id}/start-consultation、/end-consultationを推奨 | 開始は準備・読込成功済みrecord IDと期待版、終了は明示確認を照合。画面の同じカード操作を維持し、準備のための追加利用者ボタンは作らない |
| 誤終了取消 | POST /api/v1/reservations/{id}/undo-consultation-endを推奨 | 期待版・理由・確認・再送キー。受付操作権限と会計確定履歴を同一transactionで照合。現在勤務を取消そのものの追加条件にしない |
| 取消／復元／未来院 | POST /api/v1/reservations/{id}/cancel、/restore、/no-showを推奨 | 専用intent・期待版・必要確認／理由。復元は取消前状態、未来院は手動契機。既存DELETE／汎用status／自動batchで迂回しない |
| 診療中の引継ぎ | POST /api/v1/reservations/{id}/doctor-handoffを推奨 | 明示doctor_idまたは許可されたnullと期待版を入力する。交代時点はserverが確定前の現在時刻で決め、適格性を照合して結果／監査へ返す。client指定の過去時刻で判定せず、再送時に生成する現在時刻を業務fingerprintへ混ぜない。引継ぎ後時間の契約を使う。診療前の通常指名変更とは分ける |
| 候補・警告照会 | GET /api/v1/reservations/eligible-staff、POST /api/v1/reservations/validateを推奨 | 医院は認証scope、区分・予定枠・対象ID／版・通常予約／引継ぎの目的。候補・登録範囲・判定・warning_receiptを返すread-only契約。別医院の生勤務情報をclientへ公開しない |
| 結果照会 | GET /api/v1/reservation-operations/{key}?intent=...を推奨 | 認証医院・actor_kind／ID＋intentでscope。新規作成の対象IDなしでも照会可能。初回結果と現在状態を別に返す。未知キーを新規操作の成功／失敗確定と推測しない |
| LIFF空き・確定・取消 | 既存 /api/liff/:clinicId のavailable-dates／available-times／reservations／my-reservationsを保持しschemaを照合 | LIFF customerをactorとして再送・結果照会へ接続。予定枠基準、シフト・対応区分・本人医院scopeを維持。実診療状態／予定超過は入力・取得条件にしない。ホテルは全期間頭数の別契約 |
| 記録確認訂正／入院・ホテルの操作 | 予約の汎用PATCHでは扱わず関連ownerの専用intentへ | 訂正案・確認結果・同一対象ID・期待版を受け渡す。関連ownerの受け渡しschema・既存grantは§5.1・§4.7に定め、通常カルテと専用記録を変換しない |

更新DTOのdoctor_idには、[既存nullable IDの存在フラグ方式](../../../../backend/internal/owner/http_request.go)を用いる設計とする。標準encoding/jsonの二重pointerだけではnull／未送信が同じnilとなる。client型もoptional＋nullを明示し、未定を既存doctorへfallbackしない。期待版・warning_receiptは前提条件、業務内容は§4.4のfingerprintに正規化する。数値ID・明示null・配列の意味を揃え、JSON文字列の表示順だけで同一操作を別内容にしない。

| HTTP／理由code案 | 意味・画面の扱い |
|:---|:---|
| 409 reservation_version_conflict | 期待版不一致。入力案を保持して現在版・変更項目を比較し、利用者の確認後に再適用。自動上書きなし |
| 409 idempotency_content_conflict | 同じscope／キーで業務内容相違。内容を変えた再送として成功扱いにしない |
| 409 reservation_warning_changed | 警告根拠変更。入力を保持し再表示・再確認。保存時禁止へ変われば確認だけで通さない |
| 409 reservation_transition_not_allowed／reservation_record_review_required／reservation_accounting_finalized | 状態操作不可／通常編集ではなく記録確認訂正が必要／会計確定後Undo停止。別の逆遷移や返金で迂回しない |
| 409 reservation_assignment_ineligible／reservation_capacity_exceeded | 指名不適格／能力不足。院内許容警告とhard禁止を別結果にする。LINEへ例外確認を適用しない |
| 409 reservation_write_switched | 更新方式切替済みの旧write。入力を保持し対応画面へ。旧payloadの反転・旧経路復帰なし |
| 503 reservation_evaluation_unavailable | 必須依存取得失敗・探索未完了等。判定不能を不足／充足確定へ変えない。実診療超過はLINEでこの理由にならない |
| 400入力不備／401認証／403認可／404対象なし | 現行のerror区分を保持。医院外対象や他actorの操作結果を競合詳細で漏らさない |

上記codeはAPI設計上の候補名で、[既存apperrors](../../../../backend/internal/apperrors/errors.go)へ追加済みではない。成功はoperation_result（初回ID・確定版・結果）とcurrent_stateを区別する。警告はreason code・禁止／確認可・対象項目・根拠revisionとopaque receiptで返し、単一409をすべて「枠が埋まった」と表示しない。確認入力には業務内容・医院actor・対象版・根拠を結び付ける。

現在確認できるresourceはreservations、reception、medical-records、accounting、master-staff、master-reservation-type、shifts等。[権限名の実在](../../../../backend/internal/model/permission.go)と目的別操作へのgrant割当は別の確認。通常作成／編集は既存reservations create／editを基準にし、受付／Undo・訂正確認等は合意した主体と既存認可ownerへ対応表を渡す。新しい権限IDを推測しない。

<a id="reservation-roster-draft-contract"></a>
### 4.6 勤務母集団・入力保持・計算予算

勤務判定は当該医院の有効所属と個人対応区分等から候補母集団を取得し、勤務・休憩の登録結果を照合する。既指名の保護対象は候補から外れても照合から消さない。固定指名の占有と未定需要の意味は§4.2へ従う。母集団・勤務・休憩それぞれの取得成否を別に持ち、所属一覧のページ欠落や候補取得の上限切取りを「全員登録済み」としない。

[on-duty-staffsのJOIN](../../../../backend/internal/staff/shift_entry_repository.go)では勤務未登録者が消えるので母集団の代用にしない。登録済み非勤務／勤務0人、全未登録、一部未登録、取得失敗を区別する。capabilities_configured_atは能力の設定状態だけに使う。旧職種既定から新区分を自動許可せず、他院勤務はstaff ownerから同一staff IDの適格性として照合する。clientへ他院の勤務詳細を返す権限を追加しない。具体的な母集団の実在・欠損件数は未確認。

| clientが分けて保持する情報 | 更新・競合・再読込時の扱い |
|:---|:---|
| baseline（対象ID・医院・編集開始版）とdraft（入力案・変更項目・終了時刻由来） | 再取得したcurrentを別に保持し、draft全体をresetしない。診療状態・会計境界の変化を比較し禁止項目を再適用しない |
| operation key・intent・内容fingerprint・結果照会状態 | 通信断では結果不明を保持。結果照合前に別キーで作り直さず、成功後に業務内容を変える次操作は別キー |
| registration attempt IDとowner／pet／予約の成功ID | [現行refとJSON全入力key](../../../../frontend/src/features/reservations/hooks/use-reservation-save-actions.ts)だけへ依存せず、補足変更で成功IDを捨てない。必要な再読込後復旧は認証scopeのserver操作結果から行い、患者情報・tokenをURL／localStorageへ保存しない |
| warning_receipt・確認状態 | 対象版・変更入力・警告根拠が変われば古い確認を無効化して再表示。最新server検査を免除しない |
| 医院・actorのscope | 切替時は旧scopeのdraft／結果照会を新scopeへ送信しない。新scopeで認可を再確認。異なる医院／actorの成功IDを再利用しない |
| LINEの飼主入力と選択枠 | 状況変化時は飼主入力を保持して日時を再選択。409の内容相違・版競合・枠不足を区別し、結果不明はslot再選択の前に照会する |

計算は予定枠・能力・勤務・休憩を対象医院／日時範囲でbulk取得する。現行[単日／期間prefetch](../../../../backend/internal/reservation/liff_service_availability_slots.go)のN+1回避は再利用候補だが、勤務／休憩取得失敗でも続行するfallbackは新判定へ流用しない。期間の取得はstart<評価上限かつend>評価下限の重複で行い、開始日だけで日跨ぎ勤務を落とさない。§4.2の資源競合の連鎖を漏らさず、独立成分分割・探索状態のメモ化で再計算を抑える。取得上限で欠落した集合を充足と判定しない。

初期試験値候補は要求全体10万探索node／算法計算100ms（componentやslotごとの上限へ掛け算しない）。DB取得・lock待ち時間は別に測り、request deadlineも守る。未評価の日付／slotは明示的な判定不能であり、空き／満員として埋めない。保存時はlock後の最新入力で再評価し、プレビューcacheをcommit根拠にしない。node数・候補人数・予約数・SQL件数・取得時間・探索時間・lock待ち・中断理由を機密情報を含めず測定する。

値は測定前の候補であり、業務上限・性能保証ではない。[STG性能表](../../../ops/testing/STG-PERFORMANCE-CHECKLIST.md)のp95／p99も提案例で、採用済みSLOへ読み替えない。初期値・query／index・メモリ上限の受入はfixture・実測で確認する。LINEは[予定枠基準](../../../spec/line/reservation-spec.md#line-scheduled-slot-basis)を維持し、実診療の終了見込みを算法の依存へ追加しない。

<a id="reservation-api-closed-contract"></a>
### 4.7 閉じたschema・目的別grantの選定

本節は将来の予約boundaryの設計契約。API YAML／生成型／route・seedは未変更。既存resourceを用い、一般群がreception:viewのみで予約editを持つ現導線をreception:editへ一律置換しない。実医院のgrantはseedの存在から推測せずserverで再照合する。

共通型はID／version＝正規形十進文字列 `^[1-9][0-9]*$`、BIGINT上限以下。null・0・小数・指数表記・JSON数値を拒否しclientはNumberへ変換しない。共有APIの既存数値IDは変更せずadapterは精度が保証された整数だけ変換、丸め済み値を復元しない。フル日時はoffset付きRFC3339、serverは医院JST業務境界へ正規化する。日単位intended_date／LIFF dateとローカル時刻HHMMは下表の専用schemaだけ許可する。Idempotency-Keyは非zero UUID、intentはURLの登録済enumから解決。全write schemaはadditionalProperties:false相当で医院・actor・任意statusを拒否する。

| 操作／operationId | 閉じた入力（共通headerに再送キー） | 既存grant／actor |
|:---|:---|:---|
| createReservation／createReservationBatch（予約intent） | owner_id、pet_id、reservation_type_id、start_time、end_time、doctor_id（正ID/null）、visit_type（first/revisit）、is_designated（bool）、任意notes、必要warning_receipt。batchは共通予定枠・区分・指名・visit_type・is_designated・notesと、順序保持pets配列（各owner_id/pet_id）のclosed schemaで全体atomic。同一枠の既存multi-pet業務を維持し、各petごと異なる日時・区分・指名を許可する新用途へ広げない。sourceはstaff入口manual／LIFF入口lineをserver決定、初期pending／confirmedは既存予約intentから決め任意statusを受け付けない | reservations:create。受付を兼ね記録を新規作成する場合medical-records:createも検査 |
| createAndCheckInReservation／Batch（推奨専用check-in-create入口） | createと同じ患者・予定・区分入力、受付intentをURLで選択。checked_inと記録作成／再利用を同時commit。prevent_duplicate_receptionはclientで解除できずserverの重複受付保護を必須にする | reservations:create、診療の通常記録新規作成時のみmedical-records:create。他分類は専用owner grant、受付画面入口reception:view |
| updateReservationRoute（既存PATCH /api/v1/reservations/{id}/reservation-route） | expected_version、route（line/phone/reception/exam_room/record_shortcut）。独立した即時保存のintentと再送キーを持つ | reservations:edit。route表示変更をsource／LIFF本人認可・任意初期状態の代用にしない |
| reassignReservationPatient（推奨専用reassign-patient入口） | expected_version、owner_id、pet_id、必要warning_receipt。現在scope・owner/pet関連・死亡判定・状態／記録関連の既存保護を再照合 | reservations:edit＋患者検索owners:view。別pet付替えは継承正本どおりintended_date／source_medical_record_idを同じlock下で除去。記録の患者を黙って変更しない |
| updateReservation | expected_version、changes（start_time/end_time/reservation_type_id/doctor_id/visit_type/is_designated/notesだけ）、必要warning_receipt。患者差替えは専用reassign-patient、status・実時刻は専用状態intentへ | reservations:edit |
| checkInReservation／recheckInReservation | expected_version、必要warning_receipt | reservations:edit、診療の新規通常記録はmedical-records:create／本文読込はmedical-records:view。他分類は専用owner grant |
| startReservationConsultation | expected_version、prepared_record_ref、prepare_receipt | reservations:edit＋既存カルテ作成導線のmedical-records:create、本文読込medical-records:view |
| endReservationConsultation | expected_version、confirmed:true | reservations:edit。受付画面入口reception:view |
| undoReservationConsultationEnd | expected_version、reason（trim後非空）、confirmed:true | reservations:edit、受付画面入口reception:view。内部確定歴照合からaccounting編集権限を付与しない |
| cancelReservation／restoreReservation／markReservationNoShow | expected_version、confirmed:true。取消の任意reasonは監査、復元元状態はserver履歴から決定 | 取消reservations:delete、復元／手動未来院reservations:edit |
| handoffReservationDoctor | expected_version、doctor_id（正ID/null）、必要warning_receipt。実交代時刻はserver | reservations:edit |
| listReservationEligibleStaff／validateReservation | reservation_type_id、start_time、end_time、任意appointment_id／expected_version、purpose（booking/handoff）、任意doctor_id。validateは同じ通常変更案を評価 | reservations:view。候補情報master-staff:view、勤務詳細を返す場合shifts:view。内部適格性照合にmaster編集権限を付けない |
| getReservationOperation | path key、query intent。actor・医院は認証から | 元操作の必要grantを再検査。同一scopeのみ、本文は別の各view grantで取得 |
| createReservationPatient（推奨POST /api/v1/reservations/patient-registrations） | stage（owner/pet）、registration_attempt_id、owner段階はnameと任意phone、pet段階は成功owner_idとanimal_species_id・任意name。追加未知属性拒否 | reservations:create AND owners:create。owner／pet既存resourceはともにownersでpets grantを新設しない。患者再利用owners:view、既存補完owners:edit |
| 記録確認訂正 | 専用proposal（予約ID・期待版・理由・proposed_changes・record/billing refs）、confirm（correction_id・proposal_revision・dependency_revision） | 提案reception:view＋reservations:edit、記録確認medical-records:edit＋合意した記録担当確認。指名医師IDから確認actorを推測しない |
| 入院／hotel・trimming・billing内部参加 | §5.1のowner専用入力。既存owner URLからambient participantへ | 当該owner既存create/edit／取消／締め後等のgrantを保持。reservations grantだけで専用記録・会計を更新する入口を作らない |
| setNextVisitIntent／clearNextVisitIntent | 既存pet単位URLを保持。setはintended_date（YYYY-MM-DD）、任意source_medical_record_id、expected_target（noneまたはid＋version）。clearは対象id＋expected_version。UUID headerと結果照会へ参加 | 新規reservations:create、上書きcreate＋edit、削除edit。tentative時刻はserver sentinel、母集団へ数えない |
| confirmTentativeReservation | 既存confirm-tentative URL、expected_version、予定start/end、reservation_type_id、doctor_id（正ID/null）、visit_type、is_designated、任意notes／必要warning_receipt | reservations:edit。同じ予約IDを昇格、初期pending／confirmedはserverの予約intent、provenance保持 |
| createLiffReservation | 既存URLにcourse_id、staff_id（正ID/null）、date（YYYY-MM-DD）、start_time/end_time（HHMM）、任意customer_fields/request_text/trimming_course_id/trimming_option_ids/trimming_style_requestを明示。UUID header必須、IDは十進文字列 | LINE認証から医院／customer解決、既存owner/petリンクがある場合は本人患者所有を確認。未link customerを新たに一律拒否せず、customer actor本人予約scopeは全員に必須。区分・登録field・コース／optionの公開範囲をserver検査。予定枠基準、指名なしnull、hotelは専用期間schemaへ |
| cancelLiffReservation／getLiffReservationOperation | cancelはexpected_versionとUUID header、lookupはkey＋登録済intent。本人の同医院予約だけ対象 | staff RBACへ合流させず、既存受付前取消境界保持。結果照会は同じcustomer scope |
| LIFF hotel期間intent | course_id、予定入室／退室のoffset付き日時、本人予約に必要な登録fieldと任意request_text、UUID header。staff_idはnullのみ | 医師勤務なし、入退室の時間条件と全期間頭数をownerで再照合。飼主取消に版／キー適用、期間変更は医院側owner専用intent。LIFF日時変更／復元入口は設けず変更は医院連絡、超過overrideなし |

doctor_idの省略は不変更、nullは明示解除、空文字／0は禁止。notesは省略不変更、空文字消去、null禁止。必須ID・日時・versionはnull禁止。旧expected_updated_atを新versionへ推測変換しない。通常日時変更・分類変更は既存状態／記録入力の境界を検査し、閉じた項目に含まれることを無条件許可としない。開始prepare receiptは同じ予約／record ref・版・actorに対応するserverの準備結果で、read成功後に同じボタン内でstartへ渡す。record依存変更／read失敗なら再準備し、client booleanで準備を偽装しない。必要な準備／読込はDB lock保持外で行う。

次回予定のexpected_target:noneは同医院・petに既存tentativeがないことをgate／pet intent下で検査し、既存があれば競合。既存更新はid＋versionを必須にし、旧expected_reservation_id／expected_updated_atを推測変換しない。切替済みclearの無前提旧client writeも拒否する。LIFF ID数値／staff_id:0は旧wireであり、新clientは十進文字列／nullへ切替する。date＋HHMMはJST予定枠へ変換して検査する。customer_fieldsの許可キーと型は医院の公開登録field定義から生成し、任意の患者／actor／状態属性を挿入できない。LINE read consumerの互換は無副作用adapterに限り、参加不能旧writeは§7の拒否対象。新形式導入前に全該当clientを切替／停止し、旧payloadを最新版へ黙って変換してcommitしない。

登録専用intentはpet名の未送信／null／空白を空文字＝未記録へ正規化し、現pets.name NOT NULLに格納する技術選択。共有pet POST/PATCHの必須検査は維持し、専用DTOだけでなくpet ownerの専用service intentでblank許可する。表示は「未記録」をUIで示し架空名を保存しない。owner名／speciesは必須、同名統合なし。各段階は同じ試行の成功IDを保持し、任意補足変更は既存IDの補完操作とする。結果不明中は別キー新規登録を始めず同じscope／stageを照会する。結果照会のintentには登録済みpatient-owner/patient-petを含め、対応する同じgrantで照合する。

fingerprintは目的・対象・業務変更・理由・順序に意味があるbatch項目を正規化して含め、期待版・warning receipt・再送時の現在時刻を除外する。警告結果はcode、severity（confirmable/prohibited）、fields、dependency_revision、opaque receipt。receiptはscope・intent・業務fingerprint・期待版・根拠へ結び付け、保存時最新版の照合を免除しない。LINEにconfirmable overrideは返さない。

成功結果はoperation_result（operation_id、intent、result_refs、committed_versions、committed_at）とcurrent_stateを分離する。初回成功の再送は同じ結果を返しwrite／通知を増やさない。照会はsucceeded／claimed_uncommitted／not_found／result_unavailableを区別し、claimだけで失敗確定や実行中を推測しない。not_foundも通信失敗確定ではない。現在本文のview権限欠落は初回receiptと混同しない。snapshot不在は成功済IDを保持してresult_unavailableとし再作成を許可しない。§4.5の理由codeを採用し、単一409へ潰さない。

## 5. 状態と処理の接続

業務上の許可・禁止は参照先FRが正本。以下は処理を成立させる技術境界で、他分類へ診療状態を流用する許可ではない。

| 契機 | 技術境界案 | 保存結果／失敗・再送 |
|:---|:---|:---|
| 診療予約の作成 | 予約master・医院・担当・時間条件を確認して予約保存 | 予約だけを保存。通常カルテ作成は受付まで行わない |
| 診療check-in | appointmentを固定し、active recordを検索・再利用／作成 | 受付とカルテの両方成功でcommit。片方失敗はrollback、再送で増殖なし |
| start consultation | 作成済カルテを再利用して開く。未作成なら既存作成契約へ | 作成／読込失敗は開始未確定。通信断の結果照合を含め再送で別記録を作らない |
| end consultation | 明示intentで現status・権限を検査 | カルテ・会計を保持。終了後の通常編集制限を汎用status変更で解除しない |
| undo consultation end | [会計確定境界](../../../spec/workflows/reservation-to-record-flow.md#accounting-finalization-undo-boundary)と受付操作権限・理由を照合。未収・分割・0円確定後は停止、未保存入力は非確定、返金／取消後も自動再開なし。担当医の現在勤務適格性を取消の必須条件にしない | 勤務外でも取消可。指名・過去履歴を保持し、診療継続前に適格医師または未定へ手動修正。自動交代なし |
| cancel／restore | 同一予約を状態変更し取消前状態と履歴を保持 | 復元時の現在適格性を再確認。不適格なら手動修正。自動医師差替え・カルテ削除なし |
| manual no-show／recheck-in | 既存lifecycle保護を維持して手動遷移。再受付は現在適格性照合 | 自動4時間batchで変えない。再受付は同じ予約・既存カルテを使う |
| 日時／区分変更 | 状態・分類・記録入力有無に従う操作へdispatch | 未記入日付整合と記入済み訂正を分離。受付担当の理由／変更案→記録担当確認→予約・記録・請求の整合確認後確定へ。確定記録は追記、受付後分類跨ぎの通常保存は拒否 |
| inpatient start | 来院後のstaff intentで担当医決定を検査、同じ入院と記録作成を固定 | 失敗時は開始未確定・理由表示・再試行。予定だけでadmittedにしない |
| hotel accept | 宿泊管理ownerが同一宿泊・状態・受入条件を照合。担当医師は任意 | 担当医師未設定だけでは開始を拒否しない。指定時の参照／revisionと取得不明を§5.1に従い扱う。別責任担当・一名必須を追加しない |
| hotel period change | 予約済／滞在中と除外状態・実際の未退室を照合 | 予定接触端点は非重複。延泊後期間を再判定。超過は実滞在の記録を続け、責任者調整後に確定。例外承認・自動退室・他予約自動取消なし |
| LINE確定 | 選択後の入力とversionを保存時に再確認 | 不可なら入力保持・別日時へ。hotelは入退室条件＋全期間頭数を別判定 |

診療開始の同じカード操作は、カルテ準備・読み込みの成功確認後にstart intentを確定する段階方式を検討する。追加の利用者ボタンは設けない。準備した記録IDを保持し、読み込み失敗ならstart writeを呼ばず、再試行で再利用する。start応答が通信断で不明なら、業務結果を照合してから再送し、既知の読み込み失敗と同一視して自動Undoしない。受付の通常経路では、先行する受付・カルテ同時成功契約を維持する。

入院と通常カルテの別record、hotel管理、trimming専用記録は個別の関連を守る。診療後の入院決定と予定入院は同一入院へ重複なく接続する方式を設計し、通常カルテを入院記録へ変換する方式は採用済みと扱わない。

[勤務外でも可能な誤終了取消](../../../spec/workflows/reservation-to-record-flow.md#internal-batch-end-undo)は状態訂正と診療継続を分ける。取消後に診療中へ戻ったことだけで勤務外医師の診療継続を許可せず、継続前の手動担当修正の導線・server境界を設計する。新担当の判定は既存の引継ぎ後担当時間契約に従い、取消前の担当履歴は保持する。[ホテル担当の最新条件](../../../spec/screens/settings/master-reservation-type.md#internal-batch-hotel-care)は担当医師任意へ更新済み。旧受入前担当決定gateを残さず、臨床入院開始時の担当医決定とは分離する。

[記録確認訂正・その他逆遷移の正本](../../../spec/workflows/reservation-to-record-flow.md#five-business-decisions)へ従い、理由・変更案・記録担当確認と整合確認の結果を操作別intentへ渡す案。通常PATCHから迂回しない。具体的権限識別子・記録項目・会計訂正内部は関連設計へ委ねる。[旧マスタ確認補完](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-legacy-review)の担当・判断不能時停止も確定業務として扱い、保存形式や技術方式の採用とは分ける。

<a id="reservation-receiving-contracts"></a>
### 5.1 他ownerが返す境界契約

全participantは認証scope・actor・operation ID・ambient transactionを必須とし、同じ物理transactionで再照合・必要監査を行う。ID／revisionは§4.7の共通型。欠落／取得失敗／整合不能は成功へ変換しない。interfaceはconsumer側の最小契約として置き、owner service全体を相互importしない。

| owner／操作 | 入力と返却 | 同時確定／停止条件 |
|:---|:---|:---|
| medicalrecord／準備・入力判定 | appointment ID・版・対象record refs → record_ref、status、content_state(empty/entered/unknown)、dependency_revision、reason_codes | 問診・計画・処置・バイタル・検査・予防接種・健診・画像・追記等を含めownerが入力判定。自動ID・患者・初期日付だけは臨床入力ではない。子入力版を含むrevisionが必要で本体versionのみで代用しない。請求／見積依存の整合は別結果 |
| medicalrecord／訂正確認 | correction ID・理由・変更案・予約／record／billing refsと期待版 → correction_id、mode、consistency_result、relation_revision、必要addendum_id | 受付担当の案を記録担当が確認。確認後の子入力を再照合し予約・記録・請求結果と監査を同時commit。finalizedは追記のみで元本文／record.date／確定状態を解除しない。元記録日と訂正業務日を保持し、単純な日付一致を要求しない |
| billing／確定歴と予約完了 | exact appointment／record ID集合・会計版・operation ID → history_state(never/finalized/unknown)、finalization_ids、history_revision、normal_undo_allowed、reason_code | billing ownerが不変finalization event（医院・会計・確定時刻・operation key）と対象予約／記録リンクを会計更新・予約完了・監査と同時commit。未収／分割／0円を含む。rollback／未保存は履歴を作らない |
| billing／Undo境界 | 同じtransactionで対象予約の確定歴を照合 → neverだけ通常Undo許可 | 返金・取消・soft-delete・再請求で過去確定イベント／リンクを消さない。現在status・残額・CompletedAtだけで導出しない。後日現在の飼主／日付から確定対象を再推定せず確定時ID集合を保存。unknownは許可へ落とさない |
| hospitalization／hotel同一性 | 明示stay_key・既存hospitalization IDまたは新規operation、医院／患者・分類 → stay_ref、revision、planned_start/end、actual_start/end、result(created/reused/blocked) | stay_keyはowner管理の不変UUIDで医院内一意、既存hospitalization IDとの一対一関連と予約リンクを保持。両入口は明示同一targetを渡し再利用、患者／日付／名称だけで別滞在を自動統合しない。予定と実績を別保存 |
| hospitalization／実開始 | stay_ref・期待版・明示担当医 → admission_doctor_decided、record_ref、actual_start、result | 医師決定・開始記録・実開始を同時commit、失敗は全rollback。record_refはowner/kind/id/revisionの型付き参照でdaily_record ID固定・通常カルテへの変換をしない。同じoperationは同じ滞在／記録を再利用 |
| hotel／受入・延泊 | stay_ref・期待版・変更予定期間 → 任意担当医師の設定状態、capacity_assessment、capacity_revision、counted_stay_refs | [D05の最新担当条件](../../../spec/domains/hospitalization/lifecycle.md)へ従い、ホテル担当医師は任意。旧care_assignment_decidedを必須の受入gateとせず、別責任担当・非医師担当・一名必須も追加しない。指定時の参照／履歴はownerのref／revisionで保存し、未設定と取得不明を区別する。臨床入院開始の担当医決定とは別。通常カルテを追加必須にしない。同一滞在1回・実未退室を計数。超過時実績は継続、責任者調整後の再判定で延泊確定 |
| trimming／予約連動 | 予約ID・期待版・専用record_ref・operation → 同じ予約／専用記録refと結果 | Create／Update／Deleteを共通gate・予約版・切替検査へ参加させる。通常カルテへ変換せず、専用記録内部の新しい承認を本境界から作らない |

通常カルテの編集者／確認医師・必須ログ・確定警告・削除・過去入力・評価不能手入力は[カルテ初回合意](../../../spec/domains/medical-records/lifecycle.md#medical-record-agreement-first-five)と[部分技術設計](../medical-records/technical-spec.md)へ追従する。確認actorは医院が認めた医師で担当医以外も代行可。既存grantとscopeを保持し、新roleを作らない。

訂正案は外側でcorrection ID／proposal revisionを発行し、予約ID・変更項目・理由・確認者・確認時の依存revisionを保存する。新しい入力で確認snapshotが変われば再確認し、受付者のbooleanだけで記録担当確認を代用しない。会計確定後訂正・返金／締め後の精算方法や専用記録本文は各ownerの内部仕様であり、予約接続の返却型だけで確定したと扱わない。

## 6. Transaction・同時更新・冪等性案

appointment write ownerを窓口にしたambient transactionを共通にする。認可・医院scope、対象row・関連master、現在状態／versionとintentごとに必要な適格性を最終検査し、業務write・必要監査・結果取得をcommit前に成功させる。依存欠落・照合エラーはfail-closed。現在の一意性・親record lock・確定カルテ保護を先に読む。共通の勤務検査を全intentへ一律適用せず、誤終了取消は§5の例外に従う。認可・医院分離・支払い条件・監査依存の検査を免除する例外ではない。

- 予約保存とシフト／能力解除／無効化／医院別停止は、同じstaff・医院・予約影響の直列化規約へ参加させる案。勤務変更側は影響予約一覧と入力案を保持し、手動解消まで保存しない。時間経過の診療延長警告とシフト保存拒否は別判定。
- LINEの二重確定と院内同時予約は、医院・対象区分／担当・対象時間のbooking lock等で保存時再判定を成立させる。院内許容警告は意図確認後も最新条件を確認し、LINEの拒否条件を迂回させない。
- hotelは§6.4の共通gateと医院booking lockで同一宿泊の変更・医院別全期間計数・確定を直列化する方式を選定する。医院単位のlockは単純だが広範囲を待たせるため、実測して時間範囲方式とのtrade-offを判断する。同じ宿泊を予約と管理で2件として集計しない。
- check-in／記録作成／再受付／復元は同一appointmentのlockとactive記録一意性で処理。冪等keyだけを重複防止の根拠にせず、初回結果と現在状態・同じrecord IDを照合する。
- undoと支払い確定、日付訂正とカルテ入力／finalize、開始と入院記録、取消と再送は逆順実行を検証する。既存billing／medicalrecordのlock順を確認して循環待ちを避け、必要なreadをcommit前へ。
- 通信断で結果不明なら結果照会・再取得を経て同じintentを再送する設計。成功後の読取失敗を単純な保存失敗へ反転し、別予約を作り直させない。通知は既存best-effort contractを守り、業務の原子的成功と分ける。

### 6.1 記録作成participantと外側commit

[既存Transactor](../../../../backend/internal/persistence/transactor.go)はambient transactionへの参加を自動判定せず、新しいtransactionを開く。[通常カルテservice](../../../../backend/internal/medicalrecord/medical_record_service.go)と[AutoCreate](../../../../backend/internal/medicalrecord/medical_record_auto_create.go)の既存公開処理を外側transactionから単純に呼ぶ方式を採用しない。別transactionでの部分commitや外側予約lock待ちを避ける。

20:43承認に従い、consumer側の最小interfaceへambient transaction必須の記録作成／再利用participantを接続する。participantは同じtransactionの中でappointmentとの整合・一意性・記録write・必要監査を行い、record ID／処理結果／errorを呼出元へ返す。自らcommitせず、transaction欠落はfail-closed、作成失敗をbest-effortで握らない。外側の受付orchestratorが受付と記録の成功を確認して一度だけcommitし、取得結果もcommit前に準備する。具体的入出力は§5.1、lock参加は§6.4へ定める。consumer最小interfaceとowner participantの依存方向を維持し、write ownerの二重化やimport cycleを作らない。

通知等の既存post-commit副作用は外側commitの成功後だけ実行する。外側rollback時に通知・cache反映等で成功を示さず、副作用失敗で業務成功を保存失敗へ反転しない。再送で副作用を増殖させない結果識別・再実行条件を設計し、業務監査と通知を同じ成功条件へ混同しない。開始時のカルテ準備・読込→startの段階方式は§5を維持し、読込をDB transaction保持中に行う要求ではない。

### 6.2 直列化・副作用・切替状態の設計

初期方式は§6.4のDB共通gateを最上位にし、その内側の医院単位booking lockを予約作成／日時・区分・担当変更、LINE、能力再計算、ホテル計数／延泊確定へ適用する。占有・需要を変える受付／開始／終了・終了取消・取消／復元・未来院／再受付・引継ぎも対象とし、通常編集以外のintentを直列化から漏らさない。区分ごとのlockだけへ縮めて共有スタッフを二重消費しない。同一スタッフの医院間勤務重複は医院lockだけでは防げないため、スタッフ単位の医院横断直列化もstaff owner経由で行う。勤務・能力・無効化・医院別停止のwriteと予約保存が同じ規約へ参加し、影響予約を最新状態で検査する。

同一操作の対応を確保し、関係するstaff・clinicキーを一定順で取得した後、既存のlifecycle／appointment row／master／record／billingの保護順へ接続する。複数IDは安定順にする。取得順と参加表は§6.4で選定し、実writer／FK／独立transactionへの適用検証前にdeadlock-freeと宣言しない。参加できない旧経路は§7の停止条件へ扱う。必要な読み取り・監査・結果取得はcommit前に行い、client読込を待つためにDB lockを保持しない。

副作用はcommitした業務結果IDと種別で識別し、初回commit後に一度の送信試行として既存best-effort契約へ渡す。成功済み再送で新しい通知処理を増やさず、通知失敗は業務成功を反転しない。外部送信の結果不明と再配信は通知ownerの契約に委ね、exactly-once配信を主張しない。初期契約は既存best-effort一度試行を選定し、queue／outboxの新保証は採用せず、業務transaction内の必要監査と通知を混同しない。

切替状態は更新経路・医院・対象分類のどれで区切るかをconsumer一覧から定め、旧write拒否と新writeの最終検査で同じ状態を読む。切替writeも直列化へ参加し、要求開始時に旧方式だっただけで切替後の旧writeをcommitさせない。格納項目案は§3.1、入口対応は§7.2へ具体化。全体取得規約は§6.4、保存は§3.2へ選定。具体的保存期間・削除手続きの共通方針確定とquery／index・性能の並行検証を別に残す。ここはチケットや作業順の台帳ではない。

<a id="reservation-lock-participation"></a>
### 6.3 lock部分順序と他ドメインの参加条件

§6.4で全ownerの共通最上位gate・参加writer・下位保護順を選定した。receipt→staff→clinic→pet intent／lifecycleはgate後の予約系部分順序であり、全owner共通の唯一のrow順ではない。既存billing／締め、record／子入力、pet移管、staff／accountの保護はgate内で維持する。受け手のschema／結果・必要監査は§5.1、grantは§4.7へ対応する。

対象集合は事前read後、gate取得下で認可・版・関連を再確認する。変化時はwhole transactionを再実行し、取得済みrowの後へ上位gateを後付けしない。参加できない旧経路は切替／利用開始を停止する。予約側だけの無lock会計履歴readや、独立transactionのrecord呼出しで代用しない。予約版・切替検査・active記録一意性・pet移管NOWAIT・必要監査を保持する。

<a id="reservation-global-lock-contract"></a>
### 6.4 全writerの取得規約

初期方式としてDB共通のtransaction advisory gate `reservation-graph:write:v1` を選定する。既存のhashtextextended方式で固定namespace/keyを使い、commit／rollbackまで保持する。医院を跨ぐ直列化で並列性を下げるが、既存ownerの異なる下位順序を無理に反転せず接続できる。gateは他院データの読取権限を与えない。性能／実機deadlockの成立は未検証であり、実測前に小分けgateへ置換しない。

receiptを使う操作は先に短いmetadata-only claim transactionを完了する。全owner共通の必須prefixは最外の業務物理transaction開始 → 共通gate → 必要receipt行まで。cash close／auth／master等のgate参加自体は新しいclaimを必須にしない。以後は下表のowner固有部分順序を維持する。予約系上位資源を使う操作はstaff-global ID昇順 → clinic-booking ID昇順 → pet-intent／appointment-lifecycle → owner下位保護とし、policy／closeは各owner順へ置く。refund／itemのbilling→closeを共通prefixでclose→billingへ反転させない。gate前に業務row、receipt、policy、close、staffその他advisoryを保持しない。事前readで集めた集合はgate後再確認し、変化時はwhole transactionをやり直す。下表はgate取得後のowner部分順序であり、全ownerへ同じFK順を強制しない。

| write入口群 | 共通gate参加とgate後の下位順序 | 保持する条件 |
|:---|:---|:---|
| 予約通常／batch／状態／LIFF／intent／移行 | receipt → 必要staff-global → clinic-booking → pet-intent → lifecycle → appointment／関係FK | 全状態intent、旧write拒否、manual no-show、自動batch停止、予約版・切替状態へ参加。read-only候補はgate不要だが保存時再照合 |
| staff／所属／能力／シフト／schedule／テンプレート | 必要staff-global／clinic-booking → staff → assignment → capability／shift／break、各既存row順 | 対象医院横断staffと影響予約を照合。staff-account／occupation・account連動writeも最初からgate参加。行取得後にgateを後付けしない |
| owner／pet／主飼主移管 | gate → pet／ownerの既存順 → 対象appointment NOWAIT | 既存の移管対象日付・状態、active記録時拒否、LINE解除／付替え、NOWAIT保持。通常owner／pet writeもFK・対象整合へ影響する範囲は参加 |
| record本体／受付participant／補完／finalize／削除 | gate → owner/pet/staff等の既存FK → lifecycle／appointment → record等の既存owner順 | 保存契約§2・一意性・finalized保護保持。作成と受付はambient participant、別transactionのAutoCreate呼出し不可 |
| record子入力／追記／見積・関連clinical master | gate → parent record → child／master／actor／staff／FKの既存順 | 子入力を一律leaf扱いしない。独立repo transactionならその物理tx先頭でgate、外側gate保有時はambient専用participantに分離しbase DBから別txを開かない |
| billing完了／takeover／refund／取消／item訂正 | gate → close・FK・record／stay・billing等の各既存owner順 → exact appointment更新 | 完了のclose→billingとrefund／item更新のbilling→closeという既存逆辺をgateの上位参加で保護。履歴・必要監査同時commit |
| cash close／close影響write | gate → close境界等の既存順 | billingがclose取得前に参加する契約と揃える。締め本文の業務変更ではない |
| hospitalization／hotel／cage-swap／discharge-billing／trimming | gate → 既存ownerの関連FK・stay／record／billing／appointment順 | 予約ID・期間・実績・計数・担当・専用記録関連を変える全入口、Deleteを含む。予約接続のないcage変更も下位逆辺未確認なら初期は参加 |
| 区分／group／医院設定／公開枠・cap／切替状態 | gate → master等既存順 → 対象予約検査／snapshot更新 | rename、利用停止、分類補完、宿泊上限、営業枠など最終判定を変えるwriterを含む |
| auth／policy／account・監査 | policy変更やstaff/account連動はgate → 既存policy／account等の順。必要監査は同じambient tx | gate後に最終grant確認。監査participantは別tx・上位再取得なし。純粋password-reset等は他資源へblocking edgeがないことを確認したleafだけ除外 |

純readと、予約・staff・owner/pet・record・billing・close・policy・masterへblocking edgeを持たない単独leafだけ非参加とする。migration・CSV取込・backfill等の直接writerもgate参加または通常writer停止中に限定する。旧稼働writerを排出してからgate参加／拒否を同時切替し、復旧でも保護を維持する。未知のwriterは初期参加または切替停止とし、適用除外はwriter graphの証跡付きで判断する。全参加writerはgate再入を同じ物理transaction内だけで扱い、内部から新しいclaim・独立transaction・外部送信を開始しない。既存DB制約・row lock・NOWAITをgateで除去しない。writer漏れ、DDL/FK内部、旧client、逆順競合の実装検証は§8の利用開始条件に残す。

## 7. 移行方針案

追加表現と旧consumerの共存を可能にするexpand／照合／切替／整理方式を検討する。migrationの作成・適用は今回行わない。実DBや実データを読んでいないため件数・不整合の有無は未確認。

1. 調査では旧分類・所要時間既定・group有効／削除・parent、能力／excluded／職種紐付け、履歴名、シフト日時、予約status、宿泊関連・実入退室、DB制約を医院別に照合する。個人情報を文書へ出さず、件数と検証結果を実施記録へ残す。
2. 新表現は欠損を識別できる状態で追加する案。分類や色・勤務を名前・既定値で推測補完しない。未設定区分は確認後補完し、既存予約終了・担当・記録関連を保持する。旧実効能力を失わない変換案を検証してからconsumerを切替える。
3. API／UI互換性はreadとwriteの両方を確認。二重事実のwriteを恒久化せず、変換根拠と未処理を可視化する。既存親子区分や職種分類そのものは今回廃止しない。
4. 重複拒否DB制約の変更は、院内警告とLINE同時確定の代替保護が成立した上で対象を限定する。存在する制約、他consumer、既存重複・取消・未来院への影響を調べてからDDLをレビューする。制約を一括除去する計画にしない。
5. 切替前後の医院別件数、ID、履歴、現在能力、受付と記録、宿泊頭数、状態を照合する。戻し方は履歴や新状態を失わない互換方式を選び、DB巻戻しや自動再分類を復旧手段にしない。旧API等の撤去条件はconsumer調査と検証証跡で明示する。

### 7.1 新4分類の利用開始・旧consumer・復旧

新4分類は旧general/trimmingへ無損失変換できない。確認済みmappingを保持するだけで旧consumerへの公開・writeを許可しない。新分類の利用開始条件は、保存・受付・記録／管理接続・LINE・batch／shortcut等のconsumer調査と対応、未対応旧write拒否が成立していること。切替済みの更新経路は旧writeを停止し、名前や旧2値へのfallbackを禁止する。read経路も未対応の状態を誤った通常診療として表示・操作させない。医院・対象分類ごとの利用開始は§3.1のreservation_contract_statesと§6.4の共通gate下の切替方式を選定済み。実DDL・医院別revision・consumer対応確認は未実施のため利用開始保留とし、安全な接続が定まるまで利用開始を保留する。

反例は新hotel分類・旧category general・自由名称「お泊まり」。旧名称依存consumerでconfirmed／checked_inへ保存すると通常カルテを生成し得る。名称による除外を追加して救済せず、明示分類に対応したadapterか旧write拒否で守る。新形式の事実を作った後の復旧でも、そのadapter／拒否条件とID・履歴・宿泊関連を保持し、旧consumerだけへ戻す方式は採らない。API consumerの切替・旧write停止、§4.3の能力互換性を合わせて検証する。方式未確認・保護されていない未対応consumer残存は利用開始を止める条件である。未対応旧clientへの拒否表示・read互換、停止する旧write一覧と切替後の監視／復旧確認を定義し、停止した経路を無条件に再開するrollbackを採らない。

<a id="reservation-write-cutover-matrix"></a>
### 7.2 全write入口の切替・追従対応表

本表は調査で特定した必須入口群。新しい入口／consumerが見つかれば追加し、一覧の存在を全監査・切替済みの証明にしない。各行をactor／認可・版／再送・上位lock・分類dispatch・副作用・旧write拒否へ対応させ、切替状態は§3.1の医院×契約単位で検査する。

| 入口群／現在の根拠 | 切替時に揃える内容 | 未対応時の扱い |
|:---|:---|:---|
| 通常POST／PATCH／DELETE、batch、reservation-route：[route](../../../../backend/internal/reservation/routes.go)・[更新](../../../../backend/internal/reservation/reservation_service_update.go) | 作成／通常編集／状態intentを分離。メモ・経路変更も版を増分、制限項目を省略して迂回しない。削除は履歴保持の取消契約へ | 旧status・timestamp-only更新を新経路へ黙って変換せず拒否。任意の新分類を旧2値に保存しない |
| 医院別admin予約POST／DELETE：[admin repo](../../../../backend/internal/reservation/appointment_admin_repository.go) | 同じ予約owner・医院認可・分類・版・lock・取消保持へ接続 | 通常APIだけの追従でこの入口を残さない |
| LIFF空き／作成／本人取消：[LIFF service](../../../../backend/internal/reservation/liff_service_reservations.go) | customer scope・結果照会・版／再送と保存時算法。空きは予定枠、ホテルは別全期間計数 | 旧勤務／休憩取得失敗fallbackや2分類dispatchを使わない。既存自動委譲と判定用仮割当を別扱いにし、算法候補を実担当へ保存しない |
| tentative・次回予定確定／当日受付／吸収：[intent repo](../../../../backend/internal/reservation/reservation_intent_repository.go)・[handler](../../../../backend/internal/reservation/reservation_intent_handler.go) | 同じ予約・pet provenance・intent lock／CASを保全し、受付record participant・版／再送へ参加 | 新通常予約だけに適用して旧confirm／受付で誤記録を生成しない |
| trimmingの予約作成／更新／削除participant（CreateForTrimming／UpdateForTrimming／DeleteForTrimming）：[intent repo](../../../../backend/internal/reservation/reservation_intent_repository.go) | 専用記録・予約の型付き最小interfaceを保持。共通版／分類・必要booking lockへ参加 | 通常診療状態・通常カルテへfallbackしない |
| カルテからの予約補完／確定／削除：[crud](../../../../backend/internal/medicalrecord/medical_record_crud.go) | 既存ID・appointment関連・確定保護、予約補完writeの版・切替状態、§6.3の順序 | 新分類の未対応記録を作らず、通常予約handlerの名前判定だけで救済しない |
| 会計完了による予約状態更新：[complete](../../../../backend/internal/billing/accounting_complete_tx.go)・[予約participant](../../../../backend/internal/reservation/reservation_intent_repository.go) | ambientを維持し、共通lock・予約版・会計確定履歴・必要監査を同時照合 | 参加前に新Undo等を利用開始しない。会計側保護を削除して合わせない |
| 自動未来院batch：[batch](../../../../backend/internal/lstep/lstep_batch_noshow.go) | 4時間経過の自動遷移を新契約で停止。手動no-showは専用intentと同じlifecycle保護へ | batchを予約画面の切替対象外として残さない。実行中要求も切替後commit前に状態照合 |
| 能力・staff status／所属・医院別停止：[staff handler](../../../../backend/internal/staff/handler.go)・[staff update](../../../../backend/internal/staff/staff_service_update.go)・[reservation staff](../../../../backend/internal/reservation/reservation_staff_repository.go) | 肯定形PUT・明示空選択marker・既指名保護・医院横断staff lock。旧excluded PUTを停止。職種分類自体は保持 | 新区分の自動許可なし。スタッフ画面の[肯定形保存](../../../../frontend/src/features/master/api/staff-reservation-types.ts)は再利用し、[対応不可表示](../../../../frontend/src/features/master/components/StaffSidePanelSections.tsx)も追従 |
| shift通常更新／削除・LINE管理用schedule・template展開：[shift service](../../../../backend/internal/staff/shift_entry_service.go)・[repo](../../../../backend/internal/staff/shift_entry_repository.go) | 同じ勤務日時表現・未登録区別・既指名保護・共有lockへ接続。休憩置換も参加 | 別schedule入口を旧time／保護なしで保存しない。テンプレートからの書込consumerも照合 |
| 区分／groupの通常master・LINE管理用master、時間枠／公開設定：[type service](../../../../backend/internal/reservation/reservation_type_service_core.go)・[LIFF master](../../../../backend/internal/reservation/reservation_type_liff_service.go)・[route](../../../../backend/internal/reservation/routes.go) | 分類・有効意図・所要時間・名称履歴・対応母集団変化、影響予約の最新照合を同じ規約へ | masterだけ新しくして、旧read／writeから別分類・既定durationへ変換しない |
| ペット主飼主変更の未来予約移管：[pet service](../../../../backend/internal/pet/service.go)・[継承契約](../../../spec/workflows/reservation-to-record-flow.md) | 同じtransactionの未来pending／confirmed予約移管、予約版増分・切替状態・LINE連携解除／条件付き付替え・監査が参加。pet行保護後に共通advisoryを後付けせず入口から取得順を照合 | 対象日付／状態・カルテ存在時拒否・manualのcustomer_fields保持等の既存条件を維持。全予約や過去記録へ移管範囲を広げず、NOWAITの競合拒否を無条件待機へ変えない |
| 入院／ホテル管理のPOST／PATCH／DELETE／discharge-with-billing：[route](../../../../backend/internal/medicalrecord/routes_hospitalization.go)・[service](../../../../backend/internal/medicalrecord/hospitalization_service.go) | 同一宿泊の作成・期間／状態変更・削除・実退室・会計同時処理も計数／開始条件・共有lock・対象版へ参加。予定／実績と専用記録を保持。cage-swapや関連変更は同一対象／関連IDへ影響する場合に照合し、頭数上限とは分ける | ホテル実滞在の変更をLINE予約入口だけの切替対象外に残さない。ケージ数から上限を算出せず、任意担当医師の未設定を拒否理由にせず、記録・容量等の既存受入条件は維持する |
| 受付画面・予約フォーム／drag・カルテshortcut・LINE client：[保存hook](../../../../frontend/src/features/reservations/hooks/use-reservation-save-actions.ts) | 専用intent・期待版・明示null・入力保持・結果照会と認可へ接続 | APIの切替だけで旧payload・独立record生成・別予約再作成を許可しない |

切替は全経路の参加または旧write拒否を確認して契約単位で行う。旧read adapterは新形式を誤分類して操作させない。停止した入口は復旧時も拒否／正しいadapterを維持し、新形式の事実・ID・履歴を失うrollbackを採らない。具体的consumer一覧の完全性・医院別切替revisionの実在・既存実効能力・旧制約は実装前調査待ちで、今回設定やbatchを停止済みとは報告しない。

## 8. 検証方針と設計受入

今回は下記を実行していない。正本の各GWTへ対応する検証方法・対象と証跡を設計し、実施後はdocs/workへ記録、作業状態はPlaneへ反映する。Dockerで対象domain・経路を限定し、フルmake ci、実DB操作、migrationは別の実行許可に従う。

| 対象 | 検証方法案 | 必要な証跡／設計受入条件 |
|:---|:---|:---|
| 区分・group・名称・所要時間 | API境界とmaster service、UI入力・履歴、migration fixture | 全承認FRへ対応し、未設定／無効／削除／分類補完を混同しない。過去・未来・開始一致・未来への日時訂正を含む |
| 能力・勤務・医院停止 | unit＋実DB時間／scope検査＋UI対象一覧 | 全時間・休憩端点・日跨ぎ、A院とB院、全未登録／部分未登録／登録済0、入力保持・手動解消を確認 |
| 予約入力・警告・重複 | serverと画面の境界・重複DB制約・同時保存 | 院内の確認後許容とLINE拒否が共存。最小入力のowner／pet server境界・空欄・互換性、任意補足変更／通信断後の成功済みID再利用、終了保持、未定への変更、分類跨ぎ受付前後を確認 |
| 受付・診療・カルテ | handler/serviceの失敗注入、実DB逆順競合、card／modal UI | 受付だけ成功なし、既存カルテ再利用・読込失敗時未開始、開始／終了時計自動遷移なし、勤務外の誤終了取消可と継続前手動修正、未収／分割／0円会計確定後・未保存入力・返金／取消後のUndo境界、その他逆遷移の記録確認、取消と会計確定競合、participant欠落／記録失敗時の全rollback・外側commit後のみ副作用を確認 |
| 取消・未来院・再受付・訂正 | 同一予約を使った状態matrixと競合 | カルテ自動削除なし、不適格指名の手動修正、旧予約／記録の再利用、記入済日付・分類変更の通常操作拒否 |
| 入院／hotel | 予定と実開始、記録作成失敗、全期間計数と同時確定 | 両入院入口重複なし。hotel担当医師未設定でも他条件を満たせば開始可・設定時の参照保持、同一宿泊1回、取消等除外、同時刻接触、延泊超過中の実績保持／調整後確定と例外禁止、予定超過未退室、途中休診を各ACで確認 |
| LINE | 公開／本人医院分離、候補と保存条件、競合後入力保持 | hotel医師条件免除、入退室日時条件、未設定上限／設定不足、受付前取消、院内例外の不流用を確認 |
| 能力計算・互換切替 | 仮割当fixtureとAPI／UI・DB並行検証 | 跨区分・部分重複・全時間同一候補・貪欲失敗と再探索・連鎖外側競合・探索中断を照合し、仮担当が保存されないことを確認。LINEは予定枠基準とし、診療超過のみの拒否なし／実終了による予定重複の除外なしを確認。旧画面表示後の区分追加で自動許可なし。新hotel＋旧general＋名称「お泊まり」は旧write拒否または正しい接続、復旧でも保護を維持 |
| 回帰 | 既存write owner・認可・医院分離・record unique・保存契約・88ルート | 対象外の臨床承認・会計・trimming内部を変更せず、新しい他院参照・直接table writeを検出 |

同じキーの同時再送／成功後の古い期待版／現在状態が進んだ再送／成功前後の内容相違／業務rollback後の同内容再試行／キー確保後の通信断、警告根拠変更、版競合、外側rollbackと通知、切替と旧writeの両先行順序を設計検証へ含める。これらは検証計画で、今回は実行していない。

設計レビューでは正本FRとGWTへのtrace、全入口での最終検査、atomicity／冪等性、互換切替・未設定救済、判断待ちと停止境界が確認できることを条件とする。テストの存在や静的読取を実機PASSへ置換しない。

<a id="9-まとまりごとの判断待ち未確認"></a>
## 9. 文書の完了判定と残る判断

<a id="91-技術推奨案と残る判断業務5点の承認対象外"></a>
<a id="91-技術方針の決定状態と残る照合"></a>
### 9.1 技術方針の決定状態

Sentinel_f19ab03e8c2c81919c6f589bcf563811のdocs限定許可に基づき、予約との接続に必要な技術契約を§3.2・§4.7・§5.1・§6.4へ選定した。API schema／既存grant対応、最小患者登録、名称snapshot／停止flag、receiptの非再利用、全writer gate、記録訂正・不変会計確定歴・宿泊同一性と型付き記録参照は文書として具体化済み。現API／DDL／生成型は変更していない。

選定のtrade-offは、共通gateによる医院横断write待機、改名時の未来snapshot同期更新、子入力を含む依存revision、消さない最小キーbindingの保管コストである。これらは技術選択であり、新業務遷移・他院参照・grant拡張ではない。予約69固有FRと継承参照2件、全仕様79FR／194AC、LINE予定枠基準、カルテ保存契約§2、88ルート対応は維持する。

<a id="reservation-document-completion-gates"></a>
<a id="92-文書の残りと実装前検証を分ける"></a>
### 9.2 閉じた文書契約と実装前検証

| 対象 | 文書判定 | 残る確認の種類 |
|:---|:---|:---|
| API／nullable／version／警告／再送／結果照会／既存grant | §4.7で選定済み | API正本・生成型・route・実認可への反映と回帰試験 |
| 最小患者登録／成功ID再開 | §4.1・§4.7で専用intentと空文字を選定済み、共有APIの必須維持 | owner／pet表示・検索・import等の実consumer互換と通信断試験 |
| 名称履歴／個別意図／group停止／キー保存 | §3.2で選定済み、自動purgeなし・cache非必須 | 旧値の確認、rename大量更新、C05保持条件への全writer適合確認 |
| 全writer／lock／切替 | §6.4・§7.2で共通gate・owner部分順序・leaf除外・参加不可の停止を選定済み | 全実writer／FK／trigger／独立txの監査、切替排出、逆順競合、性能 |
| 記録確認訂正／会計確定歴／宿泊・専用記録接続 | §5.1で返却型・atomicity・停止条件を選定済み | owner実装への反映。記録内部・返金／締め後精算・ホテル任意担当医師はD05／D10／D12の最新正本へ従い、実装適合を別途確認する |
| FR／受入条件／独立レビュー | 既存正本を維持し技術責務へ対応 | 本改訂の独立レビュー・リンク／索引検査証跡を完了時に報告。検査通過は実装成立の証拠ではない |

**予約側の確定業務と技術接続契約の文書整理は完了。当面の自動削除停止も承認済み。ただし全ドメイン文書の完成・全writerの保持適合・実装可能という判定ではない。** 関連ドメイン内部の未決を予約APIの「schema未決」へ戻さず、各ownerの文書完了条件へ渡す。

### 9.3 業務判断と検証項目を分離する

当面自動削除しない扱いは Sentinel_9d85a7955b608191985bb02c2c9356a5 で承認済み。[C05共通方針](../../../spec/shared/retention-audit.md)で業務原本・履歴・必要監査を保持し、通常は取消／非表示／論理削除とする。年数による自動削除・一括物理削除は通常製品範囲外で必要時は別手続とし、新たな法定年数を置かない。これを予約固有の残る業務質問へ戻さず、法的年数を技術側で選ばない。ホテル担当医師は任意で、別責任担当・非医師担当・一名必須は追加しない。責任者の既存権限への割当・確定後訂正・請求修正／返金・締め後精算・専用記録は各ownerの最新正本へ従う。既決を未決の業務質問へ戻さず、実装適合を別途照合する。受付／記録担当という合意を、権限を持つ全員の役割へ自動拡張しない。

資料・実データの未知は、正確なpreset色、欠損／不整合／過去名称、旧実効能力、実DB制約、履歴復元可否。unknownは推測補完せず該当新規利用／通常Undoを止める。実装前検証は、DTO／nullableの生成、予約版全writer、receipt通信断と再送、勤務母集団完全性、10万node／100msの試験値とquery plan／メモリ／gate待機、record子入力と訂正、Undoと会計確定、旧client拒否・復旧保護である。いずれも今回は未実施。

既存Spec Kit旧出力先・生成skills・core名等の設定差分も未変更。全ドメインのドキュメント完了後に連携を最終照合し、その前に実装へ着手しない。

## 10. Spec Kitとの対応・配置

本書はテンプレートを参考に手動で作成した技術仕様であり、nativeのplan.md生成結果ではない。配置許可だけでSpec Kitのdefault出力先やツール設定を変更していない。coreはfeature内のspec.md／plan.md／tasks.mdを固定名で解決する。将来native planを使う場合、技術仕様への参照とファイル名対応をpresetで整合させ、tasks段階や全workflowを実行せずPlaneへ引き継ぐ運用が必要。本書と別の第二技術正本を生成しない。

現在の[Spec Kit運用](../../../ops/spec-kit.md)にはdocs/work/spec-kitの旧案内がある。[preset原本](../../../../.specify/presets/animalekarte-canonical/preset.yml)と生成skillsは今回未変更。仕様・技術設計・実施記録・Planeの新しい分担と、旧work資料の実態との差を保持し、既存資料を一括移動しない。

## 11. 合意FRへの対応索引

以下は業務本文のコピーではなく、今回の69固有FRから本書の検討責務を辿る索引。継承2件と対象外の業務正本は別扱い。索引の網羅は実装完了を示さない。

| 要件ID | 業務正本 | 技術設計の参照範囲 |
|:---|:---|:---|
| FR-reservations-preconsultation-date-record-correction | [診療前の日付訂正とカルテ](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-preconsultation-date-record-correction) | §4〜§6・§8 |
| FR-reservations-staff-role-label | [区分別の指名表示名](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-staff-role-label) | §4〜§6・§8 |
| FR-reservations-duty-staff-selection | [全時間の勤務・休憩適格性](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-duty-staff-selection) | §4〜§6・§8 |
| FR-reservations-unassigned-without-duty | [未登録・勤務なしの指名なし予約](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-unassigned-without-duty) | §4〜§6・§8 |
| FR-reservations-partial-shift-evaluation | [一部未登録時の候補・判定](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-partial-shift-evaluation) | §4〜§6・§8 |
| FR-reservations-clinic-duty-scope | [予約先医院の勤務先判定](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-clinic-duty-scope) | §4〜§6・§8 |
| FR-reservations-new-patient-minimum | [最小入力](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-new-patient-minimum) | §4〜§6・§8 |
| FR-reservations-internal-warning-override | [院内警告確認](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-internal-warning-override) | §4〜§6・§8 |
| FR-reservations-staff-calendar | [担当者別カレンダー](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-staff-calendar) | §4〜§6・§8 |
| FR-reservations-end-time-preservation | [終了時刻](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-end-time-preservation) | §4〜§6・§8 |
| FR-reservations-same-day-schedule | [通常時間指定予約の同日制約](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-same-day-schedule) | §4〜§6・§8 |
| FR-reservations-staff-overlap-warning | [担当者重複](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-staff-overlap-warning) | §4〜§6・§8 |
| FR-reservations-capacity-warning | [対応能力](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-capacity-warning) | §4〜§6・§8 |
| FR-reservations-drag-confirmation | [ドラッグ確認](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-drag-confirmation) | §4〜§6・§8 |
| FR-reservations-cancel-history | [取消履歴](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-cancel-history) | §4〜§6・§8 |
| FR-reservations-restore-preconsultation-cancellation | [診療前の誤取消からの復元](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-restore-preconsultation-cancellation) | §4〜§6・§8 |
| FR-reservations-edit-entry | [予約変更入口](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-edit-entry) | §4〜§6・§8 |
| FR-reservations-post-consultation-edit-limits | [診療開始後・会計完了後の通常操作](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-post-consultation-edit-limits) | §4〜§6・§8 |
| FR-reservations-doctor-handoff | [診療終了までの医師変更・引継ぎ](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-doctor-handoff) | §4〜§6・§8 |
| FR-reservations-concurrent-edit | [同時編集](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-concurrent-edit) | §4〜§6・§8 |
| FR-reservations-checkin-idempotency | [受付導線・重複防止](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-checkin-idempotency) | §4〜§6・§8 |
| FR-reservations-manual-no-show | [手動未来院・再受付](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-manual-no-show) | §4〜§6・§8 |
| FR-reservations-checkin-record-atomicity | [受付と通常カルテの同時成功](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-checkin-record-atomicity) | §4〜§6・§8 |
| FR-reservations-start-record-failure-retry | [未作成時の作成と作成／読み込み失敗時の開始未確定・再試行](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-start-record-failure-retry) | §4〜§6・§8 |
| FR-reservations-checkin-record-reuse | [受付時作成カルテの再利用・重複なし](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-checkin-record-reuse) | §4〜§6・§8 |
| FR-reservations-confirmed-drag-consultation-end | [確認付きドラッグの診療終了](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-confirmed-drag-consultation-end) | §4〜§6・§8 |
| FR-reservations-undo-consultation-end | [誤終了の専用取消操作・理由入力と確認・データ保持](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-undo-consultation-end) | §4〜§6・§8 |
| FR-reservations-manual-consultation-state | [実業務に基づく開始・終了の明示確定](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-manual-consultation-state) | §4〜§6・§8 |
| FR-reservations-future-assignment-protection | [未来・現在診療中予約の保護](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-future-assignment-protection) | §4〜§6・§8 |
| FR-reservation-types-business-routing | [業務分類と担当者・記録接続](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-business-routing) | §3・§4・§7・§8 |
| FR-reservation-types-business-classification-lock | [既存予約がある区分の分類変更禁止](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-business-classification-lock) | §3・§4・§7・§8 |
| FR-reservation-types-business-classification-initialization | [業務分類の変更・未設定補完](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-business-classification-initialization) | §3・§4・§7・§8 |
| FR-reservation-types-hotel-inpatient-entry | [ホテル担当と入院の両入口](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-hotel-inpatient-entry) | §3・§4・§7・§8 |
| FR-reservation-types-inpatient-start-confirm | [来院後の入院開始確定と同じ入院の二重登録防止](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-inpatient-start-confirm) | §3・§4・§7・§8 |
| FR-reservation-types-shared-duration | [共通所要時間と既存予約](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-shared-duration) | §3・§4・§7・§8 |
| FR-reservation-types-occupation-gate-retirement | [紐付け職種による予約可否の廃止](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-occupation-gate-retirement) | §3・§4・§7・§8 |
| FR-reservation-types-duration-required | [作成時の所要時間必須](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-duration-required) | §3・§4・§7・§8 |
| FR-reservation-types-duration-step | [所要時間の5分刻み](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-duration-step) | §3・§4・§7・§8 |
| FR-reservation-types-duration-range | [所要時間の5〜480分](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-duration-range) | §3・§4・§7・§8 |
| FR-reservation-types-unset-duration-review | [既存所要時間未設定への対応](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-unset-duration-review) | §3・§4・§7・§8 |
| FR-reservation-types-unset-duration-new-use | [未設定区分の新規利用停止](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-unset-duration-new-use) | §3・§4・§7・§8 |
| FR-reservation-types-delete-restore-history | [削除・復元と履歴保持](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-delete-restore-history) | §3・§4・§7・§8 |
| FR-reservation-types-disable-history | [無効化と履歴保持](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-disable-history) | §3・§4・§7・§8 |
| FR-reservation-types-name-history | [名称変更の未来反映・過去保持](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-name-history) | §3・§4・§7・§8 |
| FR-reservation-types-inactive-existing-edit | [既存予約の無効区分保持と変更](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-inactive-existing-edit) | §3・§4・§7・§8 |
| FR-reservation-types-group-disable-new-use | [グループ無効化と新規利用停止](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-group-disable-new-use) | §3・§4・§7・§8 |
| FR-reservation-types-group-reactivation-intent | [再有効化と個別有効意図の保持](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-group-reactivation-intent) | §3・§4・§7・§8 |
| FR-reservation-types-group-inheritance | [任意所属とグループ継承](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-group-inheritance) | §3・§4・§7・§8 |
| FR-reservation-types-group-color-picker | [10presetと詳細RGB](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-group-color-picker) | §3・§4・§7・§8 |
| FR-staff-positive-reservation-capabilities | [対応可能だけの選択](../../../spec/screens/settings/master-staff.md#requirement-fr-staff-positive-reservation-capabilities) | §3・§4・§7・§8 |
| FR-staff-shared-reservation-eligibility | [院内／LINE共通の担当可能者基準](../../../spec/screens/settings/master-staff.md#requirement-fr-staff-shared-reservation-eligibility) | §3・§4・§7・§8 |
| FR-staff-capability-booking-guard | [対応区分解除時の予約保護](../../../spec/screens/settings/master-staff.md#requirement-fr-staff-capability-booking-guard) | §3・§4・§7・§8 |
| FR-staff-deactivation-booking-guard | [無効化・退職時の指名予約保護](../../../spec/screens/settings/master-staff.md#requirement-fr-staff-deactivation-booking-guard) | §3・§4・§7・§8 |
| FR-shifts-protect-assigned-bookings | [シフト変更と指名予約保護](../../../spec/screens/24-shift-calendar.md#requirement-fr-shifts-protect-assigned-bookings) | §3・§4・§7・§8 |
| FR-shifts-overnight-duty | [日をまたぐ勤務](../../../spec/screens/24-shift-calendar.md#requirement-fr-shifts-overnight-duty) | §3・§4・§7・§8 |
| FR-shifts-holiday-warning-confirm | [休診日への変更と既存予約警告](../../../spec/screens/24-shift-calendar.md#requirement-fr-shifts-holiday-warning-confirm) | §3・§4・§7・§8 |
| FR-shifts-cross-clinic-overlap | [同一スタッフの医院間勤務重複を拒否](../../../spec/screens/24-shift-calendar.md#requirement-fr-shifts-cross-clinic-overlap) | §3・§4・§7・§8 |
| FR-shifts-manual-travel-gap | [登録者が医院間移動時間を確保](../../../spec/screens/24-shift-calendar.md#requirement-fr-shifts-manual-travel-gap) | §3・§4・§7・§8 |
| FR-line-reservations-public-business-routing | [公開対象と業務分類による接続](../../../spec/line/reservation-spec.md#requirement-fr-line-reservations-public-business-routing) | §3〜§6・§8 |
| FR-line-reservations-acceptance-bounds | [受付期間・休診日・時間外・対応可能人数](../../../spec/line/reservation-spec.md#requirement-fr-line-reservations-acceptance-bounds) | §3〜§6・§8 |
| FR-line-reservations-no-nomination | [指名なしとホテルの担当分離](../../../spec/line/reservation-spec.md#requirement-fr-line-reservations-no-nomination) | §3〜§6・§8 |
| FR-line-reservations-reselect-after-change | [状況変化時の入力保持と日時再選択](../../../spec/line/reservation-spec.md#requirement-fr-line-reservations-reselect-after-change) | §3〜§6・§8 |
| FR-line-reservations-owner-cancel | [来院受付前の飼主取消・変更案内](../../../spec/line/reservation-spec.md#requirement-fr-line-reservations-owner-cancel) | §3〜§6・§8 |
| FR-line-reservations-required-setup | [必要設定不足の新規選択不可と宿泊期間](../../../spec/line/reservation-spec.md#requirement-fr-line-reservations-required-setup) | §3〜§6・§8 |
| FR-line-reservations-hotel-concurrent-stay-limit | [医院別同時宿泊上限](../../../spec/line/reservation-spec.md#requirement-fr-line-reservations-hotel-concurrent-stay-limit) | §3〜§6・§8 |
| FR-reservations-cross-business-type-change | [受付前後の分類跨ぎ変更](../../../spec/domains/reservations/pages/reservation-management.md#requirement-fr-reservations-cross-business-type-change) | §4〜§6・§8 |
| FR-reservations-record-reviewed-correction | [記録確認訂正](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-record-reviewed-correction) | §4〜§6・§8・§9 |
| FR-reservations-controlled-reverse-transition | [その他逆遷移](../../../spec/workflows/reservation-to-record-flow.md#requirement-fr-reservations-controlled-reverse-transition) | §4〜§6・§8・§9 |
| FR-reservation-types-legacy-review | [旧マスタ確認補完](../../../spec/screens/settings/master-reservation-type.md#requirement-fr-reservation-types-legacy-review) | §3・§7〜§9 |

## カルテ訂正・例外の受け手補足

[カルテの追加部分設計](../medical-records/technical-spec.md#訂正例外4項目の接続設計)に接続する。項目実施取消は予約complete Undoではない。患者転記は元record／patient関係を保ち転記先を別IDで対応付け、原IDのpatient変更を同一ID保持の意味にしない。過去入力／空誤作成復元から当日受付・start／課金を発火しない。全writer／依存revision・結果不明照合とfinalized原本文／日付保護は継承し、会計訂正は別ownerへ。全域再照合・実装／runtimeは未完了。

## 後日結果・訂正影響の受け手補足

[カルテ追加技術契約](../medical-records/technical-spec.md#medical-record-integration-contracts)へ接続する。後日結果の追記writerは原記録・確定時参照版を保持し、通常子PATCH／unfinalizeで代用しない。既存の全writer gate・子依存revision・確認版照合に参加し、結果不明は同じ操作を照会する。画像／印刷は認可とread版を照合、印刷readからsave／start／finalizeなし。影響表示／ownerへの遷移は返金・在庫変更・予約Undoを発火しない。69件の既存trace・全owner処理の責務と確定歴を維持し、owner内部の訂正receiptは再取得で確認する。

## D03最終照合：participant・writer・互換の参照

[カルテ最終contract](../medical-records/technical-spec.md#medical-record-reconciled-contract)で、現request／responseと目標adapter・旧clientを分離。[全writer](../medical-records/technical-spec.md#medical-record-writer-matrix)にGET初期化・子入力・並替・関連変更・機器取込／revert・master・会計確認を含める。empty／entered／unknownと子依存版、医師確認と認証actor、元／先refと必要addendum、現在整合結果／receiptを受け取る。専用患者転記／過去／復元は通常CRUDや在庫副作用へfallbackしない。既存§5.1／共通gate・下位取得順・ambient・69件traceを保持し、全writer適合／切替排出・実DB／旧値／認可・並行性は利用開始前の未実施検証へ残す。
