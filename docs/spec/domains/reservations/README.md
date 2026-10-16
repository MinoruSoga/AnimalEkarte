# 予約

17:23 JSTの[LINE予約6点合意](../../line/reservation-spec.md#line-agreement-1723)へ従う。公開された診療・トリミング・ホテルは予約対象、入院は医院相談。ホテルには医師指名・医師シフト・通常予約の同日制約を適用しない。受付制限・入力保持・受付前取消・必要設定不足の扱いと受入条件はLINE正本を参照する。ホテル受入基準は17:35合意の医院別同時宿泊頭数・全宿泊期間判定で、実装・実機受入は未確認。

院内予約の登録・担当者別カレンダー・変更・取消・受付連携と、予約区分／担当者の対応区分／シフト管理への依存を確認する入口です。2026-10-14 18:26の一括合意を反映。17:35時点の完了判断は後続MECEレビューで訂正し、確認範囲と残る事項を下記に記録。実装・移行・実機受入は未完了で、現在の結論は下記最終レビューを参照します。


用語・概念関係は[予約オントロジーの語彙正本](../../ontology/reservations.md#reservation-ontology)へ統一する。予約区分、予約区分グループ（表示用）、予約区分の業務分類を区別し、担当者の総称と画面の指名医師／担当トリマー、職種と個人の対応可能予約区分を同義にしない。画面ラベル・引用・コード識別子は継承し、用語整理から未決業務を確定しない。


<a id="approved-reservation-scope"></a>
## 今回の確定仕様索引

基準82aa52f3時点の合意済み固有FRは69件。今回の分類別・担当者別開始時刻1FRを加え、基準5a4c292eでは70件。勤務複製／CSV取込／プレビュー確認3FRまでの索引は73件。2026-10-15の初診／再診・次回予定4FRを加え、現在の索引は77件、継承参照は別表の2件。過去に70行を70件の確定要件とした集計は、重複3件と継承2件を区別せず訂正した履歴であり、今回の追加による70件とは別。定義アンカーの検査数は承認件数ではない。

| FR | 承認日・根拠／適用範囲 | 正本の所在 |
|:---|:---|:---|
| FR-reservations-pet-history-visit-type | 2026-10-15、Sentinel_5bade7ee85988191b8e0ab59a6e81fd7第1項と、対象種別の追加回答Sentinel_ff5f2f7a9b8481918f976fda892e3478「yes」。実装・実機未確認 | [ペットの実受診歴による初診／再診候補](pages/reservation-management.md#requirement-fr-reservations-pet-history-visit-type) |
| FR-reservations-date-only-next-intent | 2026-10-15、Sentinel_5bade7ee85988191b8e0ab59a6e81fd7「1 yes 2 yes 3 yes 4 yes」。実装・実機未確認 | [日付だけの次回予定](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-date-only-next-intent) |
| FR-reservations-next-intent-explicit-consumption | 2026-10-15、Sentinel_5bade7ee85988191b8e0ab59a6e81fd7「1 yes 2 yes 3 yes 4 yes」。実装・実機未確認 | [今回来院への明示消化](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-next-intent-explicit-consumption) |
| FR-reservations-next-intent-cancel-no-revival | 2026-10-15、Sentinel_5bade7ee85988191b8e0ab59a6e81fd7「1 yes 2 yes 3 yes 4 yes」。実装・実機未確認 | [予約取消後の手動再設定](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-next-intent-cancel-no-revival) |
| FR-reservations-preconsultation-date-record-correction | 合意済み。2026-10-14 17:17 JST「ok」。具体的保存・表示は設計。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [診療前の日付訂正とカルテ](pages/reservation-management.md#requirement-fr-reservations-preconsultation-date-record-correction) |
| FR-reservations-staff-role-label | 合意済み。2026-10-13、追加対話。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [区分別の指名表示名](pages/reservation-management.md#requirement-fr-reservations-staff-role-label) |
| FR-reservations-duty-staff-selection | 合意済み。2026-10-13、追加対話。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [全時間の勤務・休憩適格性](pages/reservation-management.md#requirement-fr-reservations-duty-staff-selection) |
| FR-reservations-unassigned-without-duty | 合意済み。2026-10-13、追加対話。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [未登録・勤務なしの指名なし予約](pages/reservation-management.md#requirement-fr-reservations-unassigned-without-duty) |
| FR-reservations-partial-shift-evaluation | 合意済み。2026-10-13 22:21 JST、ユーザー「ok」。検出方式・対象範囲は未決。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [一部未登録時の候補・判定](pages/reservation-management.md#requirement-fr-reservations-partial-shift-evaluation) |
| FR-reservations-clinic-duty-scope | 合意済み。2026-10-13 22:31 JST、ユーザー「ok」。医院間重複シフトの保存不可は[シフト正本](../../screens/24-shift-calendar.md#requirement-fr-shifts-cross-clinic-overlap)、移動時間は[登録者が手動確保](../../screens/24-shift-calendar.md#requirement-fr-shifts-manual-travel-gap)。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [予約先医院の勤務先判定](pages/reservation-management.md#requirement-fr-reservations-clinic-duty-scope) |
| FR-reservations-new-patient-minimum | 合意済み。2026-10-13、ユーザー対話。[決定根拠](pages/reservation-management.md#agreement-2026-10-13)。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [最小入力](pages/reservation-management.md#requirement-fr-reservations-new-patient-minimum) |
| FR-reservations-internal-warning-override | 合意済み。2026-10-13、ユーザー対話。[決定根拠](pages/reservation-management.md#agreement-2026-10-13)。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [院内警告確認](pages/reservation-management.md#requirement-fr-reservations-internal-warning-override) |
| FR-reservations-staff-calendar | 合意済み。2026-10-13、ユーザー対話。[決定根拠](pages/reservation-management.md#agreement-2026-10-13)。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [担当者別カレンダー](pages/reservation-management.md#requirement-fr-reservations-staff-calendar) |
| FR-reservations-end-time-preservation | 合意済み。2026-10-13、ユーザー対話。[決定根拠](pages/reservation-management.md#agreement-2026-10-13)。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [終了時刻](pages/reservation-management.md#requirement-fr-reservations-end-time-preservation) |
| FR-reservations-same-day-schedule | 合意済み。2026-10-13 22:09 JST、ユーザー「ok」。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [通常時間指定予約の同日制約](pages/reservation-management.md#requirement-fr-reservations-same-day-schedule) |
| FR-reservations-staff-overlap-warning | 合意済み。2026-10-13、ユーザー対話。[決定根拠](pages/reservation-management.md#agreement-2026-10-13)。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [担当者重複](pages/reservation-management.md#requirement-fr-reservations-staff-overlap-warning) |
| FR-reservations-capacity-warning | 合意済み。2026-10-13、ユーザー対話。[決定根拠](pages/reservation-management.md#agreement-2026-10-13)。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [対応能力](pages/reservation-management.md#requirement-fr-reservations-capacity-warning) |
| FR-reservations-drag-confirmation | 合意済み。2026-10-13、ユーザー対話。[決定根拠](pages/reservation-management.md#agreement-2026-10-13)。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [ドラッグ確認](pages/reservation-management.md#requirement-fr-reservations-drag-confirmation) |
| FR-reservations-cancel-history | 合意済み。2026-10-13、ユーザー対話。[決定根拠](pages/reservation-management.md#agreement-2026-10-13)。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [取消履歴](pages/reservation-management.md#requirement-fr-reservations-cancel-history) |
| FR-reservations-restore-preconsultation-cancellation | 合意済み。2026-10-14 17:10 JST「ok」。復元先・手動修正は17:17合意、保存方式は設計。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [診療前の誤取消からの復元](pages/reservation-management.md#requirement-fr-reservations-restore-preconsultation-cancellation) |
| FR-reservations-edit-entry | 合意済み。2026-10-13、ユーザー対話。[決定根拠](pages/reservation-management.md#agreement-2026-10-13)。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [予約変更入口](pages/reservation-management.md#requirement-fr-reservations-edit-entry) |
| FR-reservations-post-consultation-edit-limits | 合意済み。2026-10-13 23:21 JST、ユーザー「ok」。別医師変更は[23:43合意](pages/reservation-management.md#requirement-fr-reservations-doctor-handoff)。特別訂正・段階判定等は未決。[12:22合意の判定時間・延長警告](pages/reservation-management.md#handoff-time-and-extension)へ従う。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [診療開始後・会計完了後の通常操作](pages/reservation-management.md#requirement-fr-reservations-post-consultation-edit-limits) |
| FR-reservations-doctor-handoff | 合意済み。2026-10-13 23:43 JST、ユーザー「ok」。11:15に引継ぎ後の担当時間と指名なし変更を合意。[12:22合意の判定時間・延長警告](pages/reservation-management.md#handoff-time-and-extension)へ従う。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [診療終了までの医師変更・引継ぎ](pages/reservation-management.md#requirement-fr-reservations-doctor-handoff) |
| FR-reservations-concurrent-edit | 合意済み。2026-10-13、ユーザー対話。[決定根拠](pages/reservation-management.md#agreement-2026-10-13)。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [同時編集](pages/reservation-management.md#requirement-fr-reservations-concurrent-edit) |
| FR-reservations-checkin-idempotency | 合意済み。2026-10-13、ユーザー対話。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [受付導線・重複防止](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-checkin-idempotency) |
| FR-reservations-manual-no-show | 追加境界は2026-10-14 18:26 JST一括合意。合意済み。2026-10-13、ユーザー対話。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [手動未来院・再受付](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-manual-no-show) |
| FR-reservations-checkin-record-atomicity | 合意済み。2026-10-13、ユーザー対話。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [受付と通常カルテの同時成功](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-checkin-record-atomicity) |
| FR-reservations-start-record-failure-retry | 合意済み。2026-10-14 10:59 JST、直前の明示確認へユーザー「ok」。結果不明時の技術方法は実装判断。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [未作成時の作成と作成／読み込み失敗時の開始未確定・再試行](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-start-record-failure-retry) |
| FR-reservations-checkin-record-reuse | 合意済み。2026-10-14 10:17 JST、ユーザー「ok」。未作成／失敗時は10:59の作成・読み込み失敗契約へ従う。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [受付時作成カルテの再利用・重複なし](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-checkin-record-reuse) |
| FR-reservations-confirmed-drag-consultation-end | 合意済み。2026-10-14 10:43 JST、直前の明示確認へユーザー「ok」。誤終了取消は11:56の専用操作。その他の逆方向・他遷移等は未決。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [確認付きドラッグの診療終了](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-confirmed-drag-consultation-end) |
| FR-reservations-undo-consultation-end | 合意済み。2026-10-14 11:56 JST、当該質問への明示返信「ok」。実行条件は12:07合意。会計境界は業務5点合意で確定。権限識別子・会計訂正内部は設計。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [誤終了の専用取消操作・理由入力と確認・データ保持](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-undo-consultation-end) |
| FR-reservations-manual-consultation-state | 合意済み。2026-10-14 00:18 JST、ユーザーの明示同意。開始操作は10:07・終了操作は10:12合意。確認付きドラッグ終了は10:43合意。誤終了取消は11:56合意。その他の逆方向・権限・期限等は未決。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [実業務に基づく開始・終了の明示確定](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-manual-consultation-state) |
| FR-reservations-future-assignment-protection | 合意済み。2026-10-13の回答に加え22:00 JSTに予定終了超過でも診療中なら対象と確認。未終了状態の保護は17:17合意、具体的判定方式は設計。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [未来・現在診療中予約の保護](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-future-assignment-protection) |
| FR-reservation-types-business-routing | 合意済み。2026-10-14 13:14 JST、ユーザー「ok」。ホテル・入院接続は最新D05正本へ従い、実装・実機適合は未確認。既存予約がある区分の分類変更は14:14合意で不可。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [業務分類と担当者・記録接続](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-business-routing) |
| FR-reservation-types-business-classification-lock | 合意済み。2026-10-14 14:14 JST、ユーザー「ok」。未来・過去を含む。未使用変更・未設定補完は15:58合意。管理権限・確認方法等は未決。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [既存予約がある区分の分類変更禁止](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-business-classification-lock) |
| FR-reservation-types-business-classification-initialization | 合意済み。2026-10-14 15:58 JST、ユーザー「ok」。管理権限・補完の具体的方法等は未決。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [業務分類の変更・未設定補完](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-business-classification-initialization) |
| FR-reservation-types-hotel-inpatient-entry | 合意済み。2026-10-14 16:07 JST、ユーザー回答。開始操作は16:36、担当医決定・記録時期は16:44合意。失敗時の開始未確定・理由表示・再試行は17:17合意。UI・保存方式は設計。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [ホテル担当と入院の両入口](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-hotel-inpatient-entry) |
| FR-reservation-types-inpatient-start-confirm | 合意済み。2026-10-14 16:36 JST、ユーザー「ok」。担当医・記録時期は16:44合意。失敗時の開始未確定・理由表示・再試行は17:17合意。画面配置・保存方式は設計。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [来院後の入院開始確定と同じ入院の二重登録防止](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-inpatient-start-confirm) |
| FR-reservation-types-shared-duration | 合意済み。2026-10-13、ユーザー対話。根拠は下記。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [共通所要時間と既存予約](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-shared-duration) |
| FR-reservation-types-occupation-gate-retirement | 合意済み。2026-10-13 20:43 JST、共通化提案にユーザー「ok」。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [紐付け職種による予約可否の廃止](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-occupation-gate-retirement) |
| FR-reservation-types-duration-required | 合意済み。2026-10-13 20:19「必須だね」。上下限は21:28合意を参照。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [作成時の所要時間必須](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-duration-required) |
| FR-reservation-types-duration-step | 合意済み。2026-10-13 20:20:48「5分刻みでいいよ」。上下限は21:28合意を参照。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [所要時間の5分刻み](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-duration-step) |
| FR-reservation-types-duration-range | 合意済み。2026-10-13 21:28 JST、ユーザー「ok」。両端を含む。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [所要時間の5〜480分](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-duration-range) |
| FR-reservation-types-unset-duration-review | 合意済み。2026-10-13 21:35 JST、ユーザー「ok」。実データの有無・件数は未確認。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [既存所要時間未設定への対応](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-unset-duration-review) |
| FR-reservation-types-unset-duration-new-use | 合意済み。2026-10-13 21:41 JST、ユーザー「ok」。実データ有無は未確認。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [未設定区分の新規利用停止](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-unset-duration-new-use) |
| FR-reservation-types-delete-restore-history | 合意済み。2026-10-14 12:51 JST、ユーザー「ok」。所属保持・配下新規利用停止・復元有効化後の個別有効だけ再開は13:00合意。無効中個別編集は17:17合意、保存方式は設計。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [削除・復元と履歴保持](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-delete-restore-history) |
| FR-reservation-types-disable-history | 合意済み。2026-10-13、ユーザー対話。根拠は下記。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [無効化と履歴保持](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-disable-history) |
| FR-reservation-types-name-history | 合意済み。2026-10-13 21:14 JSTの原則と21:19 JSTの開始日時境界。取得／補完方式は未決。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [名称変更の未来反映・過去保持](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-name-history) |
| FR-reservation-types-inactive-existing-edit | 合意済み。2026-10-13 20:58 JST、ユーザー「ok」。指名条件は維持。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [既存予約の無効区分保持と変更](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-inactive-existing-edit) |
| FR-reservation-types-group-disable-new-use | 合意済み。2026-10-13、ユーザーのグループ無効化回答。実装方式は未決。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [グループ無効化と新規利用停止](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-group-disable-new-use) |
| FR-reservation-types-group-reactivation-intent | 合意済み。2026-10-13 20:55 JST、ユーザー「ok」。実装方式は未決。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [再有効化と個別有効意図の保持](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-group-reactivation-intent) |
| FR-reservation-types-group-inheritance | 合意済み。2026-10-13の色継承と2026-10-14 12:56 JSTの所属0または1・グループ階層なし。業務分類は別FRの既合意を参照。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [任意所属とグループ継承](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-group-inheritance) |
| FR-reservation-types-group-color-picker | 合意済み。2026-10-13、ユーザー対話。根拠は下記。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [10presetと詳細RGB](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-group-color-picker) |
| FR-staff-positive-reservation-capabilities | 合意済み。2026-10-13、ユーザー対話。根拠は下記。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [対応可能だけの選択](../../screens/settings/master-staff.md#requirement-fr-staff-positive-reservation-capabilities) |
| FR-staff-shared-reservation-eligibility | 合意済み。2026-10-13 20:43 JST、ユーザー「ok」。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [院内／LINE共通の担当可能者基準](../../screens/settings/master-staff.md#requirement-fr-staff-shared-reservation-eligibility) |
| FR-staff-capability-booking-guard | 合意済み。2026-10-13、ユーザー対話。根拠は下記。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [対応区分解除時の予約保護](../../screens/settings/master-staff.md#requirement-fr-staff-capability-booking-guard) |
| FR-staff-deactivation-booking-guard | 合意済み。2026-10-13 21:06 JST、ユーザー「ok」。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [無効化・退職時の指名予約保護](../../screens/settings/master-staff.md#requirement-fr-staff-deactivation-booking-guard) |
| FR-shifts-protect-assigned-bookings | 合意済み。2026-10-13、ユーザー対話。根拠は下記。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [シフト変更と指名予約保護](../../screens/24-shift-calendar.md#requirement-fr-shifts-protect-assigned-bookings) |
| FR-shifts-overnight-duty | 合意済み。2026-10-14 17:17 JST「ok」。具体的保存・表示は設計。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [日をまたぐ勤務](../../screens/24-shift-calendar.md#requirement-fr-shifts-overnight-duty) |
| FR-shifts-holiday-warning-confirm | 合意済み。2026-10-14 12:43 JST、ユーザー「ok」。院内休診日登録は17:17合意、LINE条件は17:23・17:35・18:26合意のLINE正本へ従う。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [休診日への変更と既存予約警告](../../screens/24-shift-calendar.md#requirement-fr-shifts-holiday-warning-confirm) |
| FR-shifts-cross-clinic-overlap | 合意済み。2026-10-13 22:44 JST、ユーザー「はい」。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [同一スタッフの医院間勤務重複を拒否](../../screens/24-shift-calendar.md#requirement-fr-shifts-cross-clinic-overlap) |
| FR-shifts-manual-travel-gap | 合意済み。2026-10-13 22:57 JST、ユーザー「ok」。23:04の「ok」は進行への回答。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [登録者が医院間移動時間を確保](../../screens/24-shift-calendar.md#requirement-fr-shifts-manual-travel-gap) |
| FR-shifts-copy-selected-range | 合意済み。2026-10-15 Sentinel_e99d06bd6f0c8191b7bd18fd34d562f9「4 ok」。期間・スタッフを選ぶ複製。相対日付保持・全件確認、明示置換、確認範囲原子保存は合意を具体化する選定設計 | [シフト複製](../../screens/24-shift-calendar.md#requirement-fr-shifts-copy-selected-range) |
| FR-shifts-excel-csv-import | 同第4項でExcel作成CSV取込を合意。形式・列・上限の具体値は未確定 | [CSV取込](../../screens/24-shift-calendar.md#requirement-fr-shifts-excel-csv-import) |
| FR-shifts-bulk-preview-confirm | 同第4項で差分・エラー・予約影響の反映前確認と無断上書き禁止を合意。既存予約保護・医院scopeを維持 | [確認・反映](../../screens/24-shift-calendar.md#requirement-fr-shifts-bulk-preview-confirm) |
| FR-line-reservations-public-business-routing | 合意済み。17:23回答。実装・実機未確認。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [公開対象と業務分類による接続](../../line/reservation-spec.md#requirement-fr-line-reservations-public-business-routing) |
| FR-line-reservations-acceptance-bounds | 追加境界は2026-10-14 18:26 JST一括合意。合意済み。17:23回答。実装・実機未確認。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない 最新訂正によりLINEは予定枠基準、実診療状態を空き条件にしない。 | [受付期間・休診日・時間外・対応可能人数](../../line/reservation-spec.md#requirement-fr-line-reservations-acceptance-bounds) |
| FR-line-reservations-business-start-times | 2026-10-15追加合意1。通常診療は営業時間から生成、トリミングは登録開始時刻限定。時刻変更・担当者別設定を承認。実装・実機未確認 | [分類別・担当者別開始時刻](../../line/reservation-spec.md#requirement-fr-line-reservations-business-start-times) |
| FR-line-reservations-no-nomination | 合意済み。17:23回答に加え、2026-10-15追加合意2で担当未定受付・院内後決定を承認。実装・実機未確認。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [指名なしとホテルの担当分離](../../line/reservation-spec.md#requirement-fr-line-reservations-no-nomination)・[追加合意2](../../line/reservation-spec.md#line-no-nomination-unassigned) |
| FR-line-reservations-reselect-after-change | 合意済み。17:23回答。実装・実機未確認。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [状況変化時の入力保持と日時再選択](../../line/reservation-spec.md#requirement-fr-line-reservations-reselect-after-change) |
| FR-line-reservations-owner-cancel | 合意済み。17:23回答。実装・実機未確認。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [来院受付前の飼主取消・変更案内](../../line/reservation-spec.md#requirement-fr-line-reservations-owner-cancel) |
| FR-line-reservations-required-setup | 合意済み。17:23回答。実装・実機未確認。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [必要設定不足の新規選択不可と宿泊期間](../../line/reservation-spec.md#requirement-fr-line-reservations-required-setup) |
| FR-line-reservations-hotel-concurrent-stay-limit | 追加境界は2026-10-14 18:26 JST一括合意。17:35合意済み。実装・実機未確認。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [医院別同時宿泊上限](../../line/reservation-spec.md#requirement-fr-line-reservations-hotel-concurrent-stay-limit) |
| FR-reservations-cross-business-type-change | 合意済み。2026-10-14 18:26 JST、一括「ok」。予約業務と記録接続の境界のみ | [受付前後の分類跨ぎ変更](pages/reservation-management.md#requirement-fr-reservations-cross-business-type-change) |
| FR-reservations-record-reviewed-correction | 合意済み。業務5点一括承認 Sentinel_c713c4fdc74c819182979c4b8d2fef07。内部設計・実機未確認 | [記録確認訂正](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-record-reviewed-correction) |
| FR-reservations-controlled-reverse-transition | 合意済み。同一5点承認。内部設計・実機未確認 | [その他逆遷移](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-controlled-reverse-transition) |
| FR-reservation-types-legacy-review | 合意済み。同一5点承認。資料・実データ・設計未確認 | [旧マスタ確認補完](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-legacy-review) |

<a id="inherited-reservation-references"></a>
### 継承仕様の参照（今回の新承認には含めない）

| FR | 決定状態・適用範囲 | 正本の所在 |
|:---|:---|:---|
| FR-reservations-deceased-pet-exclusion | 既存記載を継承。ID付与で合意済みにしない。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [既存の要件本文](pages/reservation-management.md#requirement-fr-reservations-deceased-pet-exclusion) |
| FR-reservations-active-general-record-unique | 既存記載を継承。ID付与で合意済みにしない。今回の予約業務と接続境界のみ、対象外の既存本文を一括確定しない | [既存の要件本文](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-active-general-record-unique) |

受入条件は各正本の表へ、語彙・関係・多重度・制約は[ontology](../../ontology/reservations.md)へ参照する。後日の追加合意（17:17院内8点・17:23 LINE6点・17:35ホテル上限）は各FRの補足節・根拠と下記レビューで確認する。

## 読む順・業務ルールの所在

まず [全体要件と臨床安全原則](../../specification.md) §2.1 を確認し、次に以下の既存本文を読みます。

[予約管理](pages/reservation-management.md#agreement-2026-10-13)で途中合意と根拠→[入力・カレンダー](pages/reservation-management.md#display-inputs)→[操作](pages/reservation-management.md#operations)→[保存・同時編集](pages/reservation-management.md#persistence)→[受付・カルテの横断契約](../../workflows/reservation-to-record-flow.md#checkin-contract)を読む。その後[関連マスタと未確認境界](pages/reservation-management.md#master-dependencies)、[残差一覧](pages/reservation-management.md#implementation-gaps)へ進む。入口にルール本文を複製しません。

## 受入条件の所在

今回合意したFRに対応するGWTは[予約の受入条件](pages/reservation-management.md#acceptance)と[横断受入条件](../../workflows/reservation-to-record-flow.md#acceptance)。既存[S22](../../../ops/testing/scenarios/S22-reservation-conflict-and-staff.md)・[S04 LINE予約](../../../ops/testing/scenarios/S04-liff-reservation-journey.md)は過去の条件・手順で、新合意へ更新済みとは限りません。実機は未確認、証跡の入口は[docs/work/](../../../work/README.md)です。

<a id="reservation-document-priority"></a>
## 確定仕様・旧資料・未承認案の優先関係

2026-10-14 17:55 JST、再編対象を明示した質問（Sentinel_c39e373c663481918072b572207ac945）への「ok」（Sentinel_0a14b95b04f081919c1f255301e3be8d）でこの分離を承認。今回の予約確定仕様の識別・参照整理のみで、新しい業務ルールの承認ではない。

1. [今回の確定FR索引](#approved-reservation-scope)から参照する最新の業務正本・受入条件と、[オントロジー](../../ontology/reservations.md)を優先する。最新合意の適用範囲内で旧記述に優先し、範囲外は既存の有効仕様を維持する。
2. [旧資料](../../../work/archive/reservations/2026-10-14-decision-history.md)は当時の経緯・証跡で、現在の業務許可の根拠にしない。
3. [未承認／設計事項](../../../work/design/reservations/pending-and-design.md)は確定業務ではない。設計・調査・移行・実装・実機受入の残りを業務承認状態と分ける。

docs/spec/ は置き場を表し、配下全体の承認を表さない。仕様の承認と実装済み・実機PASSを同義にしない。混在ファイルの対象外本文・臨床安全・通常カルテ保存§2・会計／未納／レジ締めは保持する。

## 未決事項と実装状態

共通所要時間・無効化・グループ色・肯定形適格性・勤務／休憩・予約保護に加え、[院内／LINEの担当可能者共通基準](../../screens/settings/master-staff.md#requirement-fr-staff-shared-reservation-eligibility)と[区分側職種可否の廃止](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-occupation-gate-retirement)を2026-10-13 20:43 JSTに合意。予約区分自体の親子分類・色値・名称履歴の取得・補完方式（未来への日時変更時の現名称使用・履歴保持は17:17合意）・医院別設定方式と未登録警告対象の調査・予約保護境界・情報を失わない旧職種データ移行・LINE6点は17:23合意、ホテル受入基準は17:35合意、計算・保存方式は設計、具体的UI・保存等は設計。現在の記述完了判断は17:35最終レビューを参照。実装・実機受入は未完了。起票・コード着手はしない。

グループ無効化の[新規利用停止・既存予約保持](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-group-disable-new-use)も合意済み。院内のみへ限定せず、未来日の既存予約も保持する。再有効化は[個別有効意図保持](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-group-reactivation-intent)を合意。区分フラグ等の実装方式は未決。既存予約の無効区分保持と日時・指名・メモ変更は[追加合意](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-inactive-existing-edit)を参照。

スタッフ無効化・退職時も[指名予約保護](../../screens/settings/master-staff.md#requirement-fr-staff-deactivation-booking-guard)を適用する21:06 JST合意を反映。未来・現在診療中の予約は解消まで無効化不可、過去完了履歴は保持し、それだけで永久禁止しない。関連する未決と実装残差を残し、現在の記述完了判断は17:35最終レビューを参照。実装・実機受入は未完了。

予約区分名変更の[未来反映・過去保持](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-name-history)は21:14、名称変更確定時点と予約開始日時の比較は21:19 JSTに合意。開始時刻ちょうどを含む開始済みは旧名称、開始前は新名称。閲覧時点による巻戻しはしない。履歴取得・既存補完等は未決で、現在の記述完了判断は17:35最終レビューを参照。実装・実機受入は未完了。

標準所要時間は21:28 JSTの[必須・5〜480分（両端含む）・5分刻み](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-duration-range)まで合意済み。実予約の手動終了は別契約を維持し、既存所要時間未設定への自動補完なし・一覧提示は21:35合意済み。新規選択不可は21:41合意済みで、補完担当は[旧マスタ確認・補完](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-legacy-review)で確定。判定の実データ調査・設定時期・導線は設計／検証。

21:35 JSTの[既存所要時間未設定への対応](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-unset-duration-review)は条件付き合意。自動補完をせず一覧で設定を求め、登録済み予約時刻を保持する。未設定の有無・件数は未確認で、新規選択不可は21:41合意済みで、補完担当は[旧マスタ確認・補完](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-legacy-review)で確定。判定の実データ調査・設定時期・導線は設計／検証。

21:41 JSTの[未設定区分の新規選択不可](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-unset-duration-new-use)と[院内の説明・設定ボタン](pages/reservation-management.md#unset-duration-guidance)を合意。設定完了だけで個別／グループ無効を解除せず、登録済み予約を保持する。実データの有無は未確認で、LINEの設定不足案内は17:23合意の選択不可・必要に応じ医院連絡へ従う。具体的UIは設計。

21:52 JSTの[休憩端点の合意例と受入条件](pages/reservation-management.md#break-boundary)を追加。休憩開始ちょうどに終了／休憩終了ちょうどに開始は指名可能、正の時間幅で重なる場合は不可。勤務時間全体と他の指名条件を満たす前提で、予約重複の警告許容とは別に扱う。

22:00 JSTの[予定終了超過でも現在診療中なら保護対象](../../workflows/reservation-to-record-flow.md#consultation-overtime)を追加。3操作に共通だが、影響しない変更の一律拒否・診療終了自動判定・予約時間自動延長の承認ではない。他の状態境界と計算／終了の保存失敗・訂正手順は未決で、現在の記述完了判断は17:35最終レビューを参照。実装・実機受入は未完了。

22:09 JSTの[通常予約は終了が開始より後・同日内](pages/reservation-management.md#requirement-fr-reservations-same-day-schedule)を合意。翌日に出る自動終了は修正を求め保存不可。宿泊ホテル・入院は別期間管理の境界だけを参照し、深夜の実診療延長を強制終了しない。日跨ぎ勤務の具体的保存・表示方式や期間連携の詳細は未決。

22:21 JSTの[一部シフト未登録時の候補・判定](pages/reservation-management.md#requirement-fr-reservations-partial-shift-evaluation)を合意。候補と人数は登録済みの適格情報に基づき、不足警告へ情報不足の説明を併記する。全未登録の判定不可と区別し、未登録者の勤務を推測しない。検出方式と対象範囲は未決。

22:31 JSTの[予約先医院の勤務先範囲](pages/reservation-management.md#requirement-fr-reservations-clinic-duty-scope)を合意。指名候補・人数不足判定は予約先医院勤務だけ、他院同時間勤務者は除外。所属から勤務を推測せず、未登録の区別を維持。医院間の勤務重複保存不可は22:44 JSTの[シフト正本](../../screens/24-shift-calendar.md#requirement-fr-shifts-cross-clinic-overlap)で合意済み。[登録・更新の受入条件](../../screens/24-shift-calendar.md#追加合意の受入条件)へ進む。移動時間は22:57 JSTの[登録者による手動確保](../../screens/24-shift-calendar.md#requirement-fr-shifts-manual-travel-gap)を合意。勤務時間の間を空けて確保し、自動計算や固定1時間の必須値は採用しない。勤務情報の取得方法等は未決。

<a id="internal-reservation-batch-1717"></a>
## 院内予約8点の一括合意（17:17 JST）

決定者はユーザー。院内予約8点を明示した一括確認「この8点を仕様に反映してよいですか？変更したい番号だけでも大丈夫です。その後、LINE予約をまとめて確認します。」（Sentinel_99085f0f5aa4819187183153215846b9）への2026-10-14 17:17 JST「ok」（Sentinel_5010e24570b88191a4d6bd204ab00163）が根拠。個別承認編集であり、定時報告は読み取りのみ。この17:17時点でLINE案は未承認だった。現在は17:23合意に従う。実装・実機受入は未確認、予約ドメイン全体は未完了。

<a id="reservation-final-review-1735"></a>
## 予約ドメイン仕様整理の最終整合性レビュー（17:35合意反映後）

**17:35時点の判断（後続レビューで訂正）：合意事項の収録・移設までを完了としたが、意味上の整合性完了とはしない。** ホテルの最後の業務判断も17:35に合意済みで、当時は以下の範囲で未決業務を検出しなかったが、後続レビューで業務境界4点・文書不整合3点を検出し、18:26に下記修正を承認。既存の臨床承認待ち、支払い確定後訂正、専用記録内部、予約区分自体の既存親子機能の改訂は今回追加決定せず、既存仕様／各領域へ残す。

| レビュー範囲 | 照合した合意・境界 | 正本・受入条件の入口 |
|:---|:---|:---|
| 区分マスタ・所要時間・グループ・履歴 | 必須所要時間5〜480分／5分刻み、既存時刻保持、所属0または1・階層なし、無効／削除／復元・個別意図、名称履歴、業務分類変更／補完 | [正本](../../screens/settings/master-reservation-type.md#agreement-2026-10-13) |
| 担当者・医院別適格性・勤務終了 | 肯定形の対応可能区分、職種分類維持、医院別判定・他院流用なし、一医院離脱と全体退職、手動解消・他院継続・履歴保持 | [正本](../../screens/settings/master-staff.md#agreement-2026-10-13) |
| シフト・勤務・休憩・保護 | 全時間勤務／休憩外、端点合意、全／部分未登録、医院間重複禁止・移動間隔手動、日跨ぎ勤務可、変更削除の未終了予約保護 | [正本](../../screens/24-shift-calendar.md#agreement-2026-10-13) |
| 院内入力・能力警告・編集 | 最小患者入力、照合と任意補完、未定列、15分クリック、所要時間自動入力と手動終了保持、警告確認・休診日、同時編集保護 | [正本](pages/reservation-management.md#display-inputs) |
| 状態・取消・復元・日付訂正 | 診療開始後の通常変更制限、担当交代と未定、誤取消の同一予約復元／適格性再確認、未来院手動、未記入と記入済みの日付訂正分離 | [正本](pages/reservation-management.md#operations) |
| 受付・診療開始終了・誤終了訂正 | 受付とカルテ同時成功、作成済み再利用、未作成／失敗・再試行、明示開始終了・確認ドラッグ、支払い前誤終了取消・勤務外時の継続前手動修正 | [正本](../../workflows/reservation-to-record-flow.md#checkin-contract) |
| 通常カルテ保存・臨床保護 | 保存ルーティング§2と最新D03の診療終了後通常編集・履歴保持へ従い、旧確定後ロックの現行差、日付固定処理と合意済み限定訂正を区別。臨床承認待ちは対象外のまま保持 | [正本](../medical-records/lifecycle.md#hybrid-sync) |
| 入院・ホテル管理 | 予定入院と診察後入院の両入口、来院後明示開始、開始までの担当医、開始時記録・失敗未確定再試行、同じ入院重複防止、ホテル担当医師任意（臨床入院の開始前担当医決定は保持） | [正本](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-hotel-inpatient-entry) |
| LINE・ホテル全期間上限 | 公開診療／トリミング／ホテル・入院相談、受付制限、指名なし・ホテル医師条件なし、入力保持、受付前取消・直接変更復元なし、設定不足、医院別全期間宿泊上限 | [正本](../../line/reservation-spec.md#line-agreement-1723) |
| 会計・対象外境界 | 支払い前の終了取消と支払い後の会計訂正を分け、会計・未納・レジ締め本体を改訂しない | [正本](../accounting/README.md#reservation-correction-handoff) |

現在の決定状態は上表と各正本の最新合意を優先し、下記の時刻付き記録は当時の判断範囲を保持する。削除／復元・医院別対応可否・LINE・ホテル受入基準の古い未決記載、会計入口の医師再確認未合意、カルテ日付固定との見かけの矛盾を追従した。予約区分→必須の固定業務分類→記録／管理接続と、生成時点・院内／LINE・医師条件／宿泊上限・業務決定／現行モデルを区別するオントロジーを、FR・受入条件・画面参照と照合した。未承認案を合意済みに格上げせず、コードの現状やIssue完了を承認の根拠にしない。

残る設計・調査・検証：医院別設定・上限・宿泊期間の保存と計算、未登録警告対象の既存所属／勤務情報の照合、能力計算・同時確定・入力保持・履歴／監査、具体的な画面・権限識別子、既存分類・所要時間未設定や不整合データの調査と情報を失わない移行、各受入条件の実機確認。これらを新しい業務質問へ細分化しない。調査で既合意を変える必要が出た場合だけ、その差を別途確認する。

レビュー担当は本タスクの文書編集エージェント。確認範囲は上表の予約・マスタ・シフト・受付・カルテ・入院ホテル・LINE・会計境界と関連画面／索引。全88ルート対応と通常カルテ保存§2、既存の対象外仕様を保持し、旧パス案内・ローカルリンク・アンカー・FR／受入条件・差分を限定検査する。実機テスト・フルCI・実装変更は行わない。

<a id="mece-corrections-1826"></a>
## MECEレビューの7指摘と18:26一括是正

決定者はユーザー。4案と文書不整合3点をまとめた質問（Sentinel_71abdb50fcec8191988196aeb8e25f04）への2026-10-14 18:26 JST「ok」（Sentinel_4e15997c948081918e456378427a7091）が根拠。定時報告とは別の個別承認編集。実装・移行・実機受入は未確認。

| 指摘／具体的反例 | 是正した正本・受入条件 | 現在の確認範囲 |
|:---|:---|:---|
| 同じ犬の予約と宿泊管理を二重計数／取消・延泊・未退室で受付可否が変わる | [ホテル計数決定表](../../line/reservation-spec.md#hotel-counting-boundaries-1826) | 予約済み・滞在中、除外状態、同一宿泊1回、端点・延泊・未退室の合意 |
| 火曜入室・木曜退室、水曜休診 | [LINE入退室日時条件](../../line/reservation-spec.md#hotel-entry-exit-hours-1826) | 入室・退室の条件と滞在途中の休診日を分離 |
| 診療受付後ホテルへ変更／ホテル受付後診療へ変更 | [分類跨ぎ変更決定表](pages/reservation-management.md#requirement-fr-reservations-cross-business-type-change) | 受付前は再確認で可、受付後は通常編集不可・記録確認訂正 |
| 未来院後にシフト／能力／医院別指名を外し、遅刻患者を再受付 | [担当再確認・カルテ再利用](../../workflows/reservation-to-record-flow.md#no-show-recheck-1826) | 不適格なら手動修正してから同じ予約を受付 |
| 確定索引70行の重複・継承混在 | [確定索引](#approved-reservation-scope)・[継承参照](#inherited-reservation-references) | 一意な合意FRと継承参照を分離。定義数と合意数を別集計 |
| LINEの承認済み条件に旧い未承認／未決表示が残る | [LINE正本](../../line/reservation-spec.md)と関連参照 | 当時の引用・archiveは保持し、現行案内だけ最新合意へ追従 |
| オントロジーが入院記録内部内容まで承認と読める | [関係正本](../../ontology/reservations.md#reservation-concept-relations) | 開始・失敗時挙動の合意と、今回未承認の記録内部を分離 |

18:26合意反映後の再確認：独立した状態境界・ホテル条件・正本／文書構造の3視点で前回7指摘の反例を照合し、指摘箇所は記述上解消した。追加で検出した旧いLINE案内・分類未決表示・管理表の完了判断、受入条件表の配置を追従した。新しい業務決定を追加せず、無制限のMECE・perfect・実装済みとは報告しない。限定検査は仕様文書と今回の設計参照125文書、1,671リンク・1,020節参照でリンク切れなし。定義アンカー76件・受入条件ID177件に重複なし。合意索引66固有FRと継承参照2件は別集計。旧アンカー、移設済みarchive／design本文、88ルート対応とカルテ保存契約を保持し、docs-symbol-driftとdiff空白検査も成功。docs/spec/の18フォルダすべてに説明READMEを確認。これらは静的な文書確認で、実機受入ではない。関連領域の記録確認訂正、延泊超過時の解消、保存／同時処理と実機検証の未確認は残す。新たな業務判断は本承認から推測して追加しない。


<a id="remaining-business-decisions"></a>
<a id="残る業務判断の確認順"></a>
## 追加合意の確定箇所一覧

2026-10-14 17:17 JSTまでの個別合意を反映した、対話用の確認一覧。優先順の整理であり、以下の未決事項の承認ではない。院内8点は下表の正本へ反映済み。LINE6点は17:23に一括合意。ホテルの受入基準は17:35合意済み、設計・検証を残す。各単位は必要な業務結果をまとめて決め、実装の細部ごとに質問を増やさない。既存仕様の継承部分や未実施の受入を未決要件と取り違えない。

| 順 | 業務判断の単位 | 現在の決定状態 | 正本 |
|:---|:---|:---|:---|
| 1 | 誤取消の復元先と不適格な指名の手動修正 | 17:17一括合意済み。実装設計・検証は別 | [fr-reservations-restore-preconsultation-cancellation ](pages/reservation-management.md#internal-batch-restore) |
| 2 | 勤務時間外の担当医がいる誤終了取消 | 17:17一括合意済み。実装設計・検証は別 | [fr-reservations-undo-consultation-end ](../../workflows/reservation-to-record-flow.md#internal-batch-end-undo) |
| 3 | 予定時刻を過ぎた未終了予約の保護 | 17:17一括合意済み。実装設計・検証は別 | [fr-reservations-future-assignment-protection ](../../workflows/reservation-to-record-flow.md#internal-batch-protection) |
| 4 | 診療前の予約日訂正とカルテ日付 | 17:17一括合意済み。実装設計・検証は別 | [fr-reservations-preconsultation-date-record-correction ](pages/reservation-management.md#requirement-fr-reservations-preconsultation-date-record-correction) |
| 5 | グループ無効中の個別有効意図の編集 | 17:17一括合意済み。実装設計・検証は別 | [fr-reservation-types-group-reactivation-intent ](../../screens/settings/master-reservation-type.md#internal-batch-group-intent) |
| 6 | 予約日を未来へ変更した場合の区分名 | 17:17一括合意済み。実装設計・検証は別 | [fr-reservation-types-name-history ](../../screens/settings/master-reservation-type.md#internal-batch-rescheduled-name) |
| 7 | 休診日の院内予約 | 17:17一括合意済み。実装設計・検証は別 | [fr-reservations-internal-warning-override ](pages/reservation-management.md#internal-batch-holiday) |
| 8 | 日をまたぐ勤務の登録 | 17:17一括合意済み。実装設計・検証は別 | [fr-shifts-overnight-duty ](../../screens/24-shift-calendar.md#requirement-fr-shifts-overnight-duty) |
| 9 | 来院時の患者照合と任意補完 | 17:17一括合意済み。実装設計・検証は別 | [fr-reservations-new-patient-minimum ](pages/reservation-management.md#internal-batch-patient-identification) |
| 10 | ホテル担当医師の任意設定 | 旧17:17の必須条件は2026-10-15最新承認で任意へ更新済み。実装・実機検証は別 | [fr-reservation-types-hotel-inpatient-entry ](../../screens/settings/master-reservation-type.md#internal-batch-hotel-care) |
| 11 | 入院記録作成失敗時の開始未確定と再試行 | 17:17一括合意済み。実装設計・検証は別 | [fr-reservation-types-inpatient-start-confirm ](../../screens/settings/master-reservation-type.md#internal-batch-inpatient-failure) |
| 最後 | LINE固有の受付・公開・変更／取消 | 17:23の6点合意済み。ホテル受入基準は17:35合意、計算・保存方式は設計、実装・実機未確認 | [LINE正本](../../line/reservation-spec.md#line-agreement-1723) |

業務判断の質問にしない事項は、以下のように扱う。これも実装・実機受入の完了を意味しない。

- **合意済みの具体化**: 所要時間5〜480分／5分刻み、手動終了保持、10色の順とRGB自由指定、未定列、予約重複警告、勤務／休憩適格性、取消履歴、受付の同時成功、医院間重複拒否、移動時間手動確保、診療開始後／会計完了後の通常操作制限とメモ編集可、診療終了までの適格医師への引継ぎと終了後の通常変更不可、実業務に基づくスタッフの開始・終了明示確定と予定時刻の自動遷移なし、当日カードモーダル「カルテ作成」が診療開始を兼ねること、同モーダル「診察を終了する」での診療終了確定、受付時作成済みカルテを開いて診療開始する再利用・重複なし、確認付きドラッグ終了とキャンセル時の診療中維持、未作成時の作成と作成／読み込み失敗時の開始未確定・理由表示・再試行、交代先の引継ぎ後担当時間での判定と終了までの指名なし変更可・終了後の通常変更不可、理由入力・確認による専用の誤終了取消とカルテ／会計保持、支払い確定前の受付操作可能スタッフによる日付非制限の取消と理由記録、交代時から予定終了までの判定・予定終了一致／超過時の現在時点判定と、勤務終了／休憩開始時の延長警告・自動終了／交代なし。表示文言・タッチ操作・確認画面・未定への戻し方は合意した結果を満たす設計にまとめ、方針を変える場合だけ確認する。
- **実装者の設計・検証**: URLと予約IDの引継ぎ、版管理／ロック／再試行、通信断で結果不明となった場合の整合方法、グループ状態と個別状態の保存方式、名称履歴の保存・取得、他院勤務の取得・同一スタッフ照合、区分混在・部分重複の能力判定アルゴリズム、入力案保持、具体的な勤務時刻の比較。業務上の保護範囲は上表で決めるが、それを実現する方式は別に扱う。
- **資料・データ・運用の確認**: 添付参照色の正確なRGB値、既存未設定durationや不整合の有無・件数、旧職種データを失わない移行案、既存未設定区分を権限のある担当者が設定する段取り、既存権限／入力形式・UATの確認。情報・既存個人設定を勝手に変更せず、調査や移行で業務結果を変える必要が出た場合だけ該当単位へ戻す。宿泊期間と専用記録の内部仕様は関連領域の確認として保持する。

<a id="next-business-review"></a>
<a id="次にまとめて確認する業務判断整理案"></a>
### 過去の対話整理への参照

当時の確認順・未決記載は[日付付き旧記録](../../../work/archive/reservations/2026-10-14-decision-history.md#chronology)へ移設した。現在の確定索引・FR・受入条件を優先する。

<a id="five-business-decisions-current"></a>
### 業務5点反映後の現在状態

決定者はユーザー。親の5案一括確認 Sentinel_4c765f9ae6b88191902a7abfed6ac6c1「この5点でよいでしょうか？」への2026-10-14の返信 Sentinel_c713c4fdc74c819182979c4b8d2fef07「ok」が根拠。承認は業務5点に限定し、技術方式の詳細・他ドメイン内部・実装／実機受入を確定しない。

訂正担当と整合確認・会計確定境界・その他逆遷移は[横断正本](../../workflows/reservation-to-record-flow.md#five-business-decisions)、延泊超過の実績保持と調整後確定・例外なしは[ホテル運用](../../line/reservation-spec.md#hotel-extension-over-capacity)、未設定／不整合の確認補完は[マスタ正本](../../screens/settings/master-reservation-type.md#requirement-fr-reservation-types-legacy-review)へ反映した。これらを未決として再質問しない。

予約側に残るのは技術方式・API／保存形式・権限識別子・資料／実データ照合と設計／受入の確認。記録訂正・請求修正／返金・締め後訂正・専用記録は各ownerの最新正本へ接続する。ホテル担当医師は任意で、別責任担当・非医師担当・一名必須を追加しない。臨床入院開始と記録時点を再び未決に戻さない。

予約側の業務・連携境界の記述完了条件は、決定した主体・操作・結果・例外が正本／受入条件／語彙へ対応し、技術上の保留と関連ドメインへの引継ぎが明示され、独立レビュー・リンク／索引検査で不整合がないこと。予約ドメイン全体の設計完了・実装済みを意味しない。他ドメインの文書整備後に連携を最終照合し、全ドメインの文書完了までは実装しない。

## 共通事項

追加合意は[予約区分のFR／AC](../../screens/settings/master-reservation-type.md#agreement-2026-10-13)→[スタッフ適格性](../../screens/settings/master-staff.md#agreement-2026-10-13)→[シフト保護](../../screens/24-shift-calendar.md#agreement-2026-10-13)→[未来予約の横断境界](../../workflows/reservation-to-record-flow.md#requirement-fr-reservations-future-assignment-protection)の順に読む。予約側は候補・警告の依存だけを持ち、マスタ本文を重複しない。[共通用語・権限](../../shared/README.md)・[仕様全体の入口](../../README.md)。

過去のオントロジーレビュー記録は[旧記録](../../../work/archive/reservations/2026-10-14-decision-history.md#ontology-review-history)、現在の正本は[オントロジー](../../ontology/reservations.md)。今回の定義移設は本文・関係図と参照先を照合し、業務ルールを追加せずに実施した。

## このフォルダの役割と使い方

予約領域の業務ルール・受入条件・未決事項を辿る入口。

pagesに画面詳細を置き、保存契約など領域固有本文は既存の配置を維持する。横断フロー・共通マスタ・語彙をここで再定義しない。

子フォルダ：[pages/](pages/README.md)

主要文書：主要文書は本書の既存索引・上記の読む順を参照します。

読む順：本書の「読む順・業務ルールの所在」 → [画面詳細](pages/README.md) → 関連する横断契約・マスタ。

正本・状態：[決定状態と検証状態の規約](../../authoring-guide.md#status)に従い、各節・FRの承認根拠と実装／実機状態を別に読む。READMEへの掲載・フォルダ配置は承認や実装済みを表さず、対象外の既存仕様を新たに確定しない。既存の別名入口・索引の責務は維持し、本書は案内を補う。

技術的な実現方式と現行差分は[予約技術仕様](../../../design/domains/reservations/technical-spec.md)へ参照する。業務FR・受入条件は本索引の参照先が正本。具体的タスク・依存関係・進捗はPlane、実施記録はdocs/workで管理する。

### 技術4方針の文書具体化

2026-10-14 20:43 JST、Sentinel_570e060eda888191b8aeea7b5653eb07「ok」で、固定指名考慮の仮割当判定、操作別API／更新版／同一操作再送キー、受付と記録の同一transaction・確定後通知、切替済み更新経路の旧write停止を承認。[技術仕様§4.2〜§4.4・§6・§7](../../../design/domains/reservations/technical-spec.md#capacity-assignment-contract)へ具体化した。仮担当は保存しない。技術方針を新しい業務遷移の許可や実装済みへ読み替えない。権限ID・色・実データ・他ドメイン内部は未確認のまま。全ドメイン文書完了前は実装しない。

### LINE予定枠基準の訂正反映

ユーザー Sentinel_08ce4d91de608191a05c30220331a832 の訂正と Sentinel_2334916604f88191830b7a194c74bf85／Sentinel_61b62c3381b88191ad0d5af32bdcffb5 のdocs限定編集許可に従い、[LINE受付条件の正本](../../line/reservation-spec.md#line-scheduled-slot-basis)・受入条件と[技術算法](../../../design/domains/reservations/technical-spec.md#capacity-assignment-contract)を整合した。LINEの実診療状態・予定超過チェック案は不採用。予定枠重複・シフト・対応区分を保持し、ホテル実滞在計数や院内の未終了指名予約保護へ広げない。これを残るLINE業務質問から除外し、予約全体の設計完了・実装可能とは扱わない。

### API・保存・切替対応表と他ドメイン参加条件

Sentinel_f19ab03e8c2c81919c6f589bcf563811のdocs限定許可で、[保存詳細](../../../design/domains/reservations/technical-spec.md#reservation-storage-details) → [closed schema／既存grant](../../../design/domains/reservations/technical-spec.md#reservation-api-closed-contract) → [受け手契約](../../../design/domains/reservations/technical-spec.md#reservation-receiving-contracts) → [全writer取得規約](../../../design/domains/reservations/technical-spec.md#reservation-global-lock-contract)の順に技術契約を選定した。FR／ACの正本は既存参照先を維持する。[完了判定](../../../design/domains/reservations/technical-spec.md#reservation-document-completion-gates)では、予約接続の文書具体化済みと、C05保持方針への実適合・関連domain最新本文との照合・実データ／性能検証を分離する。LINE予定枠基準を維持し、全ドメイン文書完了前は実装しない。


<a id="reservation-document-completion-current"></a>
### 予約ドメインの文書整理完了境界

2026-10-14、当面の保存・廃棄の扱いはユーザー Sentinel_9d85a7955b608191985bb02c2c9356a5 の「ok」で承認。[C05最新方針](../../shared/retention-audit.md)で業務原本・履歴・必要監査を保持し、通常操作は取消／非表示／論理削除とする。年数自動削除・一括物理削除は通常製品範囲外で必要時は別手続とし、原本所在不明を自動／手動の実体削除へ含めない。具体的年数・法令適合・法的無期限保存は決定していない。既存の論理削除・非表示・復元・履歴参照の合意を保持する。

当時の完了範囲は、予約区分／表示group・指名担当者／個人対応・シフト／休憩・院内／LINE・ホテル／予定入院からの予約接続を含む69固有FRと継承参照2件の文書整理、および[技術接続契約](../../../design/domains/reservations/technical-spec.md#reservation-document-completion-gates)。予約側固有の業務判断を残す段階は終え、受付へ進む。全仕様79FR／194ACは全体集計であり、予約69件と混同しない。

関連領域の訂正・精算／返金／締め後処理・専用記録は[D05](../hospitalization/lifecycle.md)・[D04](../owners-pets/lifecycle.md)・[D06](../trimming/lifecycle.md)、[D10](../accounting/README.md)、[D11〜13/16](../../../design/domains/finance-operations/technical-spec.md)の最新正本を参照する。ホテル担当医師は任意、保持は[C05](../../shared/retention-audit.md)に従い、旧未決のまま残さない。実装前の旧データ・preset値・DB制約・API生成／互換／認可・全writer・並行処理・性能検証、全ドメイン確定後の連携再照合は未実施。文書整理完了を実装開始・利用開始・実機受入済みへ読み替えない。

本節と技術仕様§9が現在状態。過去の対話記録・[設計／未承認事項](../../../work/design/reservations/pending-and-design.md)には移設時点の記述を保持しており、既決事項を再び未決へ戻す根拠にしない。未承認の画面方式案や対象外内部仕様も一括承認したと扱わない。

## 主飼主変更との接続更新

2026-10-15の[D04追加合意4・5](../owners-pets/lifecycle.md#agreement-owner-contacts-transfer-20261015)を参照する。登録時の連絡先重複は警告・候補を示して別飼主登録を許容し、主飼主変更は今後の予約・LINE通知先への影響一覧を確認後に切り替える。予約への引継ぎは[ページ補足](pages/reservation-management.md#owner-transfer-impact-handoff)、保存は[予約技術契約](../../../design/domains/reservations/technical-spec.md#owner-transfer-impact-confirmation)へ。D04所有のFR/ACを本書の69固有FRへ二重計上せず、LINE未連携の自動新規紐付け・過去請求／預り金移管を追加しない。

## 2026-10-15会議4点の索引更新

最新の通常カルテ編集・復元と終了履歴は[D03](../medical-records/lifecycle.md#medical-record-agreement-autosave-20261015)、専用施術の開始／終了・料金は[D06](../trimming/lifecycle.md)、勤務の複製／CSVは[D14](../../screens/24-shift-calendar.md#shift-copy-csv-agreement-20261015)へ。当時の73固有FRの索引と技術traceの一致を確認した。旧70／69の検査記録は当時の範囲。新しい予約枠数・時間既定や無断上書き許可は追加しない。

### 初診／再診・次回予定の4点追加合意（2026-10-15）

2026-10-15の4点確認（Sentinel_0755eb8b662c8191939153f29229607e）への回答 Sentinel_5bade7ee85988191b8e0ab59a6e81fd7「1 yes 2 yes 3 yes 4 yes」を上記77件の索引・FR／受入条件・技術traceへ接続した。初診／再診の[受診歴対象種別](pages/reservation-management.md#visit-history-included-types)も2026-10-15、対象範囲を確認した質問 Sentinel_0cfe19c6936881918fdeaf851ab09042 へのユーザー回答 Sentinel_ff5f2f7a9b8481918f976fda892e3478「yes」で確定。4点と対象種別の業務決定、API適合・実装前検証を分ける。今回の文書検査／独立レビュー結果は別途報告し、過去のレビュー記録を今回の検証済み証拠にしない。

<a id="decision-alignment-20261017"></a>
## 2026-10-17確定仕様との接続

[決定根拠・適用範囲](../SPECIFICATION_PROGRESS.md#confirmed-business-decisions-20261017)：ユーザーの2026-10-16〜17 JSTの直接決定、文書更新承認は10月17日03:20 JST。以下は目標仕様・未実行の受入条件であり、実装・実機合格を示さない。

[業務正本・受入条件](pages/reservation-management.md#line-revisit-intake-20261017)。予約完了・予約詳細から任意の再診問診へ進み、受付前に訂正する。予約選択ペットごと7項目を自動取込し、カルテ未作成でも別保存、飼主由来・最新／旧回答と日時を保持する。未回答で予約・受付を止めない。日時変更は継承、新予約・ペット変更は未回答、取消履歴保持。既存のプロフィール転記・入力保持だけを問診本文取込済みの証拠にしない。本人LINE認証・対象患者認可を維持し、共通患者の自院利用を他院記録への認可免除としない。
