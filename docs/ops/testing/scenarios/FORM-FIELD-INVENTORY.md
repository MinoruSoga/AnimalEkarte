# フォーム×項目 棚卸し (Form Field Inventory)

> **目的**: 受入 V シリーズの永続化フォームについて、検証済み exact field key と未収録 gap を管理する。
> **使い方**: 左の項目を 1 行ずつ [FIELD-LEVEL-PROTOCOL.md](FIELD-LEVEL-PROTOCOL.md) で実施。手順の補足は V01〜V05。
> **更新規則**: 画面に入力項目を追加したら、本表と該当 V を同 PR で更新する。
> **ステータス**: inventory は再構築中。下表は検証済み exact field key の部分一覧であり、一意フォーム総数や「全フォーム/全項目完了」はまだ主張しない。route inventory は 86 product pages だが page 数と form 数は別。

凡例: **R**=必須 / **O**=任意 / **C**=条件付き必須 / **S**=システム（入力不可→F は N/A）。fieldKey は保存 request の wire key を使う。response 名や UI state 名が異なる場合は表直前の対応表を参照する。未検証の UI-only helper/context は永続 field と数えない。

---

## V01 臨床（算定保留）

### medical-record-clinical-plan PATCH — `/api/v1/medical-records/:id/clinical-plan` — [V01 §1](V01-clinical-forms.md)

Owner: clinical_plan PATCH child resource. The parent medical-record and inquiry fields in the following subsection are not clinical-plan PATCH fields.

| UI/system state             | wire key            | R/O | 型        | 制約・特記                           | F 重点 |
| :-------------------------- | :------------------ | :-- | :-------- | :----------------------------------- | :----- |
| physicalExam                | physical_exam       | O   | text      | 身体検査所見                         | F0 F4  |
| diagnosis1CategoryId        | diagnosis_type_id   | O   | select FK | 診断1区分                            | F0 F4  |
| diagnosis1NameId            | diagnosis_name_id   | O   | select FK | 診断1名称。区分連動                  | F0 F4  |
| diagnosis2CategoryId        | diagnosis_2_type_id | O   | select FK | 診断2区分                            | F0 F4  |
| diagnosis2NameId            | diagnosis_2_name_id | O   | select FK | 診断2名称。区分連動                  | F0 F4  |
| assessment                  | diagnosis_details   | O   | text      | 診断詳細                             | F0 F4  |
| plan                        | treatment_policy    | O   | text      | 治療方針                             | F0 F4  |
| existingClinicalPlanVersion | version             | S   | number    | CAS system version; not user-entered（返却 version を次の PATCH へ送る。省略 legacy path は CAS 対象外） | F0 F4  |

### medical-record-form — parent record / inquiry (not clinical-plan PATCH) — `/medical-records/new|/:id` — [V01 §1](V01-clinical-forms.md)

| fieldKey             | ラベル概要 | R/O | 型          | 制約・特記                                               | F 重点     |
| :------------------- | :--------- | :-- | :---------- | :------------------------------------------------------- | :--------- |
| pet_id               | 対象ペット | R   | id          | 親カルテcontext。new は query。無しは select-pet         | F0         |
| attending_vet        | 担当医     | O   | select      | 親medical recordの即時 PATCH                             | F0 F4      |
| visit_type           | 来院種別   | O   | select/enum | 親medical recordの即時 PATCH + 詳細キャッシュ invalidate | F0 F4      |
| chief_complaint_type | 主訴区分   | O   | select FK   | inquiry PATCH                                            | F0 F4 C3-1 |
| chief_complaint      | 主訴       | O   | text        | inquiry PATCH                                            | F0 F4      |

### medical-record-treatments-tab — 治療タブ — [V01 §2](V01-clinical-forms.md)

| fieldKey              | ラベル概要        | R/O | 型        | 制約・特記                                            | F 重点   |
| :-------------------- | :---------------- | :-- | :-------- | :---------------------------------------------------- | :------- |
| item_id               | 項目（処置/薬剤） | R   | select FK |                                                       | F0 F1 F4 |
| quantity              | 数量              | R   | number    | >0。薬量 hard gate                                    | F1 F3 F4 |
| unit_price            | 単価              | O   | money     | ≥0                                                    | F3 F4    |
| dose_deviation_reason | 用量逸脱理由      | C   | text      | 下限/推奨乖離時必須。絶対上限は理由不可・保存ブロック | F1 F6    |

### medical-record-vitals — VitalsModal — [V01 §3](V01-clinical-forms.md)

| fieldKey         | ラベル概要 | R/O | 型       | 制約・特記                               | F 重点   |
| :--------------- | :--------- | :-- | :------- | :--------------------------------------- | :------- |
| recorded_at      | 記録日時   | R   | datetime | 未来日時は FE 拒否                       | F1 F2 F4 |
| temperature      | 体温       | O   | number   | FE 30〜45℃（45.0 受理・45.1 拒否）       | F3 F4 F5 |
| heart_rate       | 心拍数     | O   | number   |                                          | F0 F4 F5 |
| respiration_rate | 呼吸数     | O   | number   | exact key（`respiratory_rate` ではない） | F0 F4 F5 |
| weight           | 体重       | O   | number   | exact key（`weight_kg` ではない）        | F3 F4 F5 |
| weight_unit      | 体重単位   | O   | enum     | weight と組で保存                        | F0 F4    |
| note             | 備考       | O   | text     |                                          | F0 F4 F5 |

### medical-record-checkups-tab — [V01 §4](V01-clinical-forms.md)

| fieldKey               | R/O | 型        | F 重点                                                                                                                   |
| :--------------------- | :-- | :-------- | :----------------------------------------------------------------------------------------------------------------------- |
| performed_on           | R   | date      | F1 F4                                                                                                                    |
| checkup_type_id        | R   | select FK | F1 F4 C3-1                                                                                                               |
| （fixture field keys） | C   | 定義依存  | 実行前に承認済み健診定義の exact key を run report inventory に列挙する。この source inventory が未完の間は V01 完了不可 |
| result_note            | O   | text      | F4 F5                                                                                                                    |
| next_due_on            | O   | date      | F4 F5                                                                                                                    |

### medical-record-vaccination-tab — [V01 §5](V01-clinical-forms.md)

| fieldKey           | R/O | 型         | F 重点                     |
| :----------------- | :-- | :--------- | :------------------------- |
| pet_id             | S   | id         | F0/N/A（親カルテcontext）  |
| medical_record_id  | S   | id         | F0/N/A（保存済み親カルテ） |
| vaccine_id         | R   | select     | F1 F4                      |
| date               | R   | date       | F1 F4                      |
| lot1               | O   | text       | F4 F5                      |
| lot2               | O   | text       | F4 F5                      |
| lot3               | O   | text       | F4 F5                      |
| lot4               | O   | text       | F4 F5                      |
| next_date          | O   | date       | F4 F5（接種日以下を拒否）  |
| supplemental       | O   | text       | F4 F5（補助説明）           |
| next_schedule_type | O   | enum/radio | F4                         |
| remarks            | O   | text       | F4 F5                      |

### medical-record-image-upload / addendum / examination-import / estimate-tab — [V01 §6](V01-clinical-forms.md)

| formId                            | fieldKey        | R/O | 型    | F 重点                  |
| :-------------------------------- | :-------------- | :-- | :---- | :---------------------- |
| medical-record-image-upload       | file            | R   | file  | F0 F2(MIME) F3(10MB) F4 |
| medical-record-addendum           | content         | R   | text  | F1 F4                   |
| medical-record-addendum           | reason          | R   | text  | F1 F3(500) F4           |
| medical-record-examination-import | examination_ids | R   | multi | F1 F4                   |
| medical-record-estimate-tab       | title           | O   | text  | F0 F4                   |
| medical-record-estimate-tab       | lines           | O   | grid  | F0 F4                   |

### examination-form — [V01 §7](V01-clinical-forms.md)

| fieldKey                      | R/O | 型          | F 重点                                                                                                                             |
| :---------------------------- | :-- | :---------- | :--------------------------------------------------------------------------------------------------------------------------------- |
| exam_type_id                  | R   | select      | F1 F4 C3-1                                                                                                                         |
| doctor_id                     | R   | select      | F1 F4                                                                                                                              |
| date                          | O/R | date        | F1/F4                                                                                                                              |
| （fixture result field keys） | C   | number/text | 実行前に `exam_type_field_id` ごとの exact key を run report inventory に列挙する。この source inventory が未完の間は V01 完了不可 |

### vaccination-form（独立）— [V01 §8](V01-clinical-forms.md)

独立フォームも `pet_id`、`medical_record_id`（該当時）、`vaccine_id`、`date`、`lot1..lot4`、`next_date`、`supplemental`、`next_schedule_type`、`remarks` を個別に記録する。未来日接種拒否・次回予定境界（接種日と同日拒否）を含め、タブ表の適用可能な全Fを実施する。

### checkup-form（独立クイック）— [V01 §9](V01-clinical-forms.md)

| fieldKey                       | R/O | F 重点                                                                                        |
| :----------------------------- | :-- | :-------------------------------------------------------------------------------------------- |
| checkup_type_id                | R   | F1 F4                                                                                         |
| performed_on                   | R   | F1 F4                                                                                         |
| （fixture checkup field keys） | C   | 実行前に exact key を run report inventory に列挙。この source inventory が未完の間は完了不可 |

### hospitalization-form — [V01 §10](V01-clinical-forms.md)

| fieldKey               | R/O | 型        | F 重点                        |
| :--------------------- | :-- | :-------- | :---------------------------- |
| pet_id                 | R   | id        | F0 F1                         |
| hospitalization_type   | R   | enum      | 既定「入院」 F0 F4            |
| start_date             | R   | date      | 既定当日 F1 F4                |
| end_date               | O   | date      | F4 F5                         |
| cage_id                | R   | select FK | BUG-037 F1 F4 C3-1            |
| owner_request          | O   | text      | 一覧主訴列 F4 F5              |
| doctor_id              | O   | select    | F4                            |
| memo                   | O   | text      | F4 F5                         |
| staff_notes            | O   | text      | F4 F5                         |
| is_insurance           | O   | boolean   | F4                            |
| insurance_company_name | C   | text      | ON 時 F1 F4                   |
| insurance_number       | C   | text      | ON 時 F1 F4                   |
| treatment_plans        | O   | rows      | 新規のみ。登録後は読取専用 F4 |

### hospitalization-care-plan / daily-vitals / daily-care-logs / daily-staff-notes — [V01 §11](V01-clinical-forms.md)

| formId                            | 必須 fieldKey | その他                                           | F 重点   |
| :-------------------------------- | :------------ | :----------------------------------------------- | :------- |
| hospitalization-care-plan         | name          | type, timing                                     | 各 F1/F4 |
| hospitalization-daily-vitals      | time          | 計測値は任意（カルテバイタルの 30〜45 制約なし） | F1 F4    |
| hospitalization-daily-care-logs   | time, type    | value/notes 任意                                 | F1 F4    |
| hospitalization-daily-staff-notes | time, content |                                                  | F1 F4    |

### trimming-form — [V01 §12](V01-clinical-forms.md)

保存 request は `frontend/src/features/trimming/hooks/trimming-form-utils.ts`。`record_shortcut` の時刻は現在時刻の秒・ミリ秒から生成される。画像の UI preview は存在するが、下記 create/update builder に画像キーは含まれないため、画像の永続化経路は coverage gap として別に追跡する。

| fieldKey | R/O | F 重点 |
|:--|:--|:--|
| pet_id | R | F0 F1（create context） |
| appointment_id | S | 既存予約との紐付け（create context） |
| reservation_type_id | S | 予約区分（create context） |
| staff_id | R | F1 F4 |
| course_id | R | F1 F4 C3-1・無効マスタ #228 |
| option_ids | O | multi F4 |
| start_time | C | record_shortcut 時は既定生成、指定時 F4 |
| end_time | C | start_time と組。既定は90分後 F4 |
| status | C | 新規・既存予約なし時の initialStatus（pending/in_consultation） F4 |
| reservation_route | S | record_shortcut。新規経路の識別 |
| style_request | O | スタイル要望 F4 F5 |
| bw | O | 体重 F3 F4 F5 |
| bw_unit | O | 体重単位 F4 |
| bt | O | 体温 F3 F4 F5 |
| used_shampoo | O | 使用シャンプー F4 F5 |
| used_ribbon | O | 使用リボン F4 F5 |
| remarks | O | 備考 F4 F5 |
| （画像保存経路） | — | exact wire key 未確定。永続化 PASS には数えない |

---

## V02 会計・予約・在庫（算定保留）

### accounting-settlement-form — [V02 §1](V02-accounting-reservation-forms.md)

| fieldKey                | R/O | 型     | F 重点                   |
| :---------------------- | :-- | :----- | :----------------------- |
| payment_splits[].method | R   | select | F1 F4（method 重複禁止） |
| payment_splits[].amount | R   | money  | F1 F3(≥1)                |
| cash_tendered           | C   | money  | 現金時 預り≥金額         |
| change_override         | O   | money  | ≥0                       |
| post_close_reason       | C   | text   | 締め後必須 F1            |

### accounting-item-add-dialog — [V02 §2](V02-accounting-reservation-forms.md)

| fieldKey            | R/O             | F 重点               |
| :------------------ | :-------------- | :------------------- |
| name（手動）        | R               | F1 F4                |
| unit_price          | R               | F3(≥0) F4            |
| quantity            | R               | F3(>0) F4            |
| category            | R（手動）       | F1 F4                |
| other_reason        | C               | category=other 時 F1 |
| merchandise_item_id | R（マスタタブ） | F1 F4 C3-1           |
| tax_rate            | O               | F4                   |

### credit-correction-dialog — [V02 §3](V02-accounting-reservation-forms.md)

| fieldKey         | R/O | F 重点       |
| :--------------- | :-- | :----------- |
| amount           | R   | F1 F3(≥1) F4 |
| method           | R   | card/electronic_money F0 F4 |
| memo             | O   | F4 F5 |
| reason           | R   | F1 F4        |

### refund-dialog — [V02 §4](V02-accounting-reservation-forms.md)

| fieldKey | R/O | F 重点                 |
| :------- | :-- | :--------------------- |
| amount   | R   | F1 F3(1..残額) F4      |
| method   | R   | 使用済み手段のみ F0 F4 |
| reason   | O   | F4 F5                  |

### cash-register-close-form — [V02 §5](V02-accounting-reservation-forms.md)

| fieldKey    | R/O | F 重点          |
| :---------- | :-- | :-------------- |
| close_date  | R   | F0 F4           |
| period      | R   | am/pm/emg F0 F4 |
| actual_cash | R   | F1 F3(≥0) F4    |

### estimate-form — [V02 §6](V02-accounting-reservation-forms.md)

保存 request は `frontend/src/features/estimates/api/types.ts`。UI の camelCase と区別する。`owner_id` / `pet_id` / `medical_record_id` は create の紐付けで、通常 update DTO は受け取らない。期限クリアは update の `clear_valid_until` 契約を確認する。

| fieldKey        | R/O | F 重点                 |
| :-------------- | :-- | :--------------------- |
| title           | R   | F1 F4                  |
| status          | R   | 作成時 draft/sent のみ |
| owner_id         | O   | F4                     |
| pet_id           | O   | F4                     |
| medical_record_id | O   | F4                     |
| subtotal        | O   | ≥0 F3 F4               |
| tax_total        | O   | ≥0 F3 F4               |
| total_amount     | O   | ≥0 F3 F4               |
| insurance_amount | O   | ≥0 F3 F4               |
| discount_amount  | O   | ≥0 F3 F4・権限で F6    |
| valid_until      | O   | F4 F5                  |
| comment         | O   | F4 F5                  |
| notes           | O   | F4 F5                  |

### reservation-form-modal / reception-walkin / reception-status — [V02 §7–9](V02-accounting-reservation-forms.md)

wire key は `frontend/src/features/reservations/api/transforms.ts`。UI の start/end/doctor とは別名。受付 status は同じ予約 API の更新値。

| fieldKey            | R/O | F 重点         |
| :------------------ | :-- | :------------- |
| pet_id              | R   | F1 F4          |
| owner_id            | R   | F1 F4          |
| reservation_type_id | R   | F1 F4 C3-1     |
| start_time            | R   | F1 F4・枠衝突  |
| end_time              | R   | F1 F4・枠衝突  |
| doctor_id            | O   | F4             |
| notes                | O   | F4 F5          |
| status（受付）      | R   | 遷移のみ F0 F4 |

### shift-form-dialog — [V02 §10](V02-accounting-reservation-forms.md)

| fieldKey             | R/O | F 重点                                 |
| :------------------- | :-- | :------------------------------------- |
| staff_id             | S   | F0/N/A（launch context）               |
| date                 | S   | F0/N/A（launch context）               |
| start_time           | C   | F1 F4（勤務時。UI stateは`startTime`） |
| end_time             | C   | F1 F4（勤務時。UI stateは`endTime`）   |
| shift_type           | R   | F1 F4（`off`/`paid_leave`は時刻なし）  |
| notes                | O   | F4 F5                                  |
| breaks[].break_start | C   | F1 F3 F4（休憩行）                     |
| breaks[].break_end   | C   | F1 F3 F4（開始後）                     |

`template_id`は入力補助であり、shift保存payloadの永続fieldではない。

### clinic-holiday-modal — [V02 §11](V02-accounting-reservation-forms.md)

| fieldKey | R/O | F 重点                                                                 |
| :------- | :-- | :--------------------------------------------------------------------- |
| date     | S   | F0/N/A（launch context 由来の read-only。編集/必須空 F1 は適用しない） |
| reason   | O   | F4 F5                                                                  |

### inventory-form — `/inventory/new|/:id` — **[V02 §12 新設](V02-accounting-reservation-forms.md)**

UI form 名は `minStockLevel` / `expiryDate` / `lastRestocked`。保存時は `use-inventory-form-model.ts` が以下の snake_case へ変換する。

| fieldKey      | ラベル     | R/O | 型   | 制約                                | F 重点   |
| :------------ | :--------- | :-- | :--- | :---------------------------------- | :------- |
| name          | 品名       | R   | text | 非空                                | F0 F1 F4 |
| category      | カテゴリ   | R   | enum | medicine/consumable/food/other      | F0 F1 F4 |
| unit          | 単位       | R   | text | 非空（HTML required + BE required） | F0 F1 F4 |
| quantity      | 現在庫数   | R   | int  | ≥0                                  | F1 F3 F4 |
| min_stock_level | 最低在庫数 | R   | int  | ≥0                                  | F1 F3 F4 |
| location      | 保管場所   | O   | text |                                     | F4 F5    |
| expiry_date    | 使用期限   | O   | date |                                     | F4 F5    |
| supplier      | 仕入先     | O   | text |                                     | F4 F5    |
| last_restocked | 最終入庫日 | O   | date |                                     | F4 F5    |

---

## V03 飼主・組織（算定保留）

### owner-create-edit — [V03 §1](V03-owner-pet-staff-forms.md)

| fieldKey            | R/O            | 型        | F 重点              |
| :------------------ | :------------- | :-------- | :------------------ |
| owner_name          | R              | text      | F1 F4               |
| owner_name_kana     | R(new)/O(edit) | text      | F1/F4               |
| phone               | R              | phone     | F1 F2 F4 C3-2       |
| email               | O              | email     | F2 F4 F5 C3-2       |
| postal_code         | O              | postal    | F2 F4               |
| address1            | O              | text      | F4 F5               |
| address2            | O              | text      | F4 F5               |
| home_postal_code    | O              | postal    | F2 F4               |
| home_address1       | O              | text      | F4 F5               |
| home_address2       | O              | text      | F4 F5               |
| company             | O              | text      | F4 F5               |
| company_phone       | O              | phone     | F2 F4 F5            |
| membership_type     | R              | enum4     | F0 F4               |
| discount_rate       | O              | 0–100     | F3 F4 F5            |
| is_dangerous        | O              | bool      | F4                  |
| birth_date          | O              | date      | F4 F5（null PATCH） |
| remarks             | O              | text      | F4 F5               |
| dm_preference       | O              | tri-state | F0 F4 F5            |
| clinic_id（登録先） | O              | select    | F4 C3-1             |

### pet-edit-modal / pet-add-pending — [V03 §2–3](V03-owner-pet-staff-forms.md)

保存 request は `frontend/src/lib/transforms/pet.ts`（pending nested create は `frontend/src/types/owner.ts`）。`name_kana` の response 側名称は `pet_name_kana`。性別は `gender`、マイクロチップは `microchip_number`、去勢避妊日は `neutered_date`。

| fieldKey          | R/O   | F 重点            |
| :---------------- | :---- | :---------------- |
| name              | R     | F1 F4             |
| name_kana     | O     | F4 F5             |
| animal_species_id | R     | F1 F4 C3-1        |
| gender               | R(FE) | F1 F4             |
| breed             | O     | F4 F5             |
| birth_date        | O     | F4 F5             |
| weight            | O     | 0–200 FE F3 F4    |
| color             | O     | F4 F5             |
| microchip_number         | O     | F4 F5             |
| blood_type        | O     | F4 F5             |
| neutered_date       | O     | F4 F5             |
| food              | O     | F4 F5             |
| environment       | O     | F4 F5             |
| insurance_id      | O     | F4 C3-1           |
| danger_level      | O     | enum F4           |
| danger_reason     | C     | high 時必須 F1 F4 |
| acquisition_type  | O     | F0 F4 F5          |
| remarks           | O     | F4 F5             |

### pet-deceased-dialog — [V03 §4](V03-owner-pet-staff-forms.md)

| fieldKey    | R/O | F 重点             |
| :---------- | :-- | :----------------- |
| deceased_at | R   | F1 F2(未来拒否) F4 |
| reason      | O   | F4 F5              |

### staff-side-panel — [V03 §5](V03-owner-pet-staff-forms.md)

| fieldKey                 | R/O | F 重点                            |
| :----------------------- | :-- | :-------------------------------- |
| name                     | R   | F1 F4                             |
| email                    | O   | F2 F4 C3-2                        |
| password                 | C   | email 時必須・8+英数混在 F1 F2 F3 |
| occupation_id            | O   | F4 C3-1                           |
| permission_group_ids     | O   | 2段階保存 F4                      |
| clinic_ids               | O   | F4                                |
| reservation_capabilities | O   | F4                                |
| line_display_name        | O   | F4 F5                             |

### permission-group-side-panel — [V03 §6](V03-owner-pet-staff-forms.md)

| fieldKey                      | R/O | F 重点                                                         |
| :---------------------------- | :-- | :------------------------------------------------------------- |
| name                          | R   | F1 F4 C3-2                                                     |
| description                   | O   | F4 F5                                                          |
| color                         | O   | F4                                                             |
| permissions[resource][action] | O   | **全 resource × view/create/edit/delete** を F4（ON/OFF 代表） |

### clinic-master-side-panel — [V03 §7](V03-owner-pet-staff-forms.md)

| fieldKey                       | R/O | F 重点                                                                     |
| :----------------------------- | :-- | :------------------------------------------------------------------------- |
| name                           | R   | F1 F4                                                                      |
| postal_code                    | O   | F2 F4                                                                      |
| address                        | O   | F4                                                                         |
| phone                          | O   | F2 F4                                                                      |
| email                          | O   | F2 F4                                                                      |
| standard_tax_rate              | O   | 0–100% F3 F4                                                               |
| reduced_tax_rate               | O   | 0–100% F3 F4                                                               |
| （accounting document fields） | O   | exact keys 未収録。この行は coverage に数えず、収録完了まで V03 incomplete |

---

## V04 設定マスタ（算定保留）

V04 の exact field-key inventory は再構築中。共通して確認済みの keys は次だけであり、各 master 固有 key の収録が完了するまで V04 完了・総数を主張しない。

| form family                 | fieldKey   | R/O | F 重点 |
| :-------------------------- | :--------- | :-- | :----- |
| standard master side panels | name       | R   | F1 F4  |
| standard master side panels | is_active  | O   | F4     |
| standard master side panels | sort_order | O   | F4     |

L-step/LINE settings は V05 が唯一の owner。V04 では実行・集計しない。

---

### lab-device-item-master — `/settings/lab-device-item-masters` — [V04 §8](V04-settings-master-forms.md)

V04 が唯一の owner。V05 では数えない。

| fieldKey                | R/O | 型        | F 重点     |
| :---------------------- | :-- | :-------- | :--------- |
| name                    | R   | text      | F0 F1 F4   |
| sourceType              | R   | enum      | F0 F1 F4   |
| examTypeId              | O   | select FK | F0 F4 C3-1 |
| isActive                | O   | boolean   | F0 F4      |
| sortOrder               | O   | number    | F0 F3 F4   |
| items[].examTypeFieldId | O   | select FK | F0 F4 C3-1 |
| items[].isActive        | O   | boolean   | F0 F4      |

## V05 認証・LINE（算定保留）

| formId                       | 主要 fieldKey（すべて F 適用）                                                  | 参照                                                  |
| :--------------------------- | :------------------------------------------------------------------------------ | :---------------------------------------------------- |
| auth-login                   | email, password                                                                 | V05-1                                                 |
| auth-change-password         | current, new, confirm                                                           | V05-2                                                 |
| auth-forgot-password         | email                                                                           | V05-3                                                 |
| auth-reset-password          | password, confirm                                                               | V05-4                                                 |
| liff-account-link            | （自動・入力なし）                                                              | F0 分岐のみ                                           |
| line-reserve-create          | —                                                                               | exact keys 未収録。収録完了まで V05 incomplete        |
| line-reserve-cancel          | cancel action                                                                   | F0 F4                                                 |
| line-reservation-settings    | —                                                                               | exact keys 未収録。収録完了まで V05 incomplete        |
| line-reservation-page-editor | header_text, request_example, reservation_notice, cancel_notice, privacy_policy | V05-9（唯一の owner）                                 |
| line-reservation-slots       | 日付・開始時刻                                                                  | V05-10                                                |
| owner-line-customer-link     | link/unlink 操作                                                                | V05-11                                                |
| lstep-settings               | —                                                                               | exact keys 未収録。収録完了まで V05 incomplete        |
| lstep-tag-config             | —                                                                               | exact keys 未収録。収録完了まで V05 incomplete        |
| lstep-checkup-sync-create    | tag_name                                                                        | exact filter keys 未収録。収録完了まで V05 incomplete |

---

## カバレッジ更新チェックリスト（開発者）

- [ ] 新規永続フォーム → 本ファイルに formId + fieldKey 追加 + 該当 V に § 追加
- [ ] 既存フォームに項目追加 → fieldKey 行追加
- [ ] 必須/境界変更 → R/O と F 重点を更新
- [ ] route inventory は 86 product pages。page 数と unique persistent form 数を混同しない
- [ ] wildcard / UI 全項目 / 動的 placeholder が残る間は inventory incomplete とし、全フォーム完了や総数を主張しない
