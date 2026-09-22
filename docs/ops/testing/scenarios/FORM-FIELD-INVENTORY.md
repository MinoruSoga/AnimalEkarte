# フォーム×項目 棚卸し (Form Field Inventory)

> **目的**: 受入 V シリーズの永続化フォームについて、検証済み exact field key と未収録 gap を管理する。
> **使い方**: 左の項目を 1 行ずつ [FIELD-LEVEL-PROTOCOL.md](FIELD-LEVEL-PROTOCOL.md) で実施。手順の補足は V01〜V05。
> **更新規則**: 画面に入力項目を追加したら、本表と該当 V を同 PR で更新する。
> **ステータス**: V02〜V05 の exact field key は保存 request builder と突合して収録済み（BE 受理だが FE 非送出の key は計上せず注記）。V01 の定義依存行（健診/検査の fixture field）は envelope を確定済みとしたうえで、承認済み定義の id 列挙は実行時作業として残る。一意フォーム総数や「全フォーム/全項目の受入完了」は依然主張しない — inventory 収録は実行完了を意味しない。route inventory は 86 product pages だが page 数と form 数は別。

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

保存は2段: `POST /api/v1/medical-records/:mrId/checkups`（下表の静的 key）→ 値あり時のみ `PUT /checkups/:cid/field-results`（`{results:[…]}` 全置換。省略=送信しない、`results:[]`=意図的全削除）。**編集 PATCH は動的項目を含まない**（動的項目は create 専用）。`results[]` 項目キー: `checkup_type_field_id` + `value_number|value_text|value_bool|value_list` のいずれか1種（`field_type` で決定: number→value_number、boolean→value_bool、single_select/text→value_text、multi_select/checklist→value_list）。列挙元: `GET /v1/masters/checkup-types/:id/fields` の `id`/`field_type`。

| fieldKey                                             | R/O | 型        | F 重点                                                                                                                                      |
| :--------------------------------------------------- | :-- | :-------- | :------------------------------------------------------------------------------------------------------------------------------------------ |
| date                                                 | R   | date      | F1 F4                                                                                                                                       |
| checkup_type_id                                      | R   | select FK | F1 F4 C3-1                                                                                                                                  |
| doctor_id                                            | O   | select FK | F4                                                                                                                                          |
| result                                               | O   | text      | F4 F5                                                                                                                                       |
| next_date                                            | O   | date      | F4 F5                                                                                                                                       |
| results[].checkup_type_field_id / value_*（上記4種） | C   | 定義依存  | envelope 確定済み。実行前に承認済み健診定義の field.id と field_type を run report inventory に列挙する。定義 id が未完の間は V01 完了不可 |

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

`POST /api/v1/examinations` / `PATCH /:id`（`items` ネスト）または `PUT /:id/items`（全置換・`items`省略=全削除）。`status` は PATCH のみ。`items[]` 項目キー: `exam_type_field_id`（テンプレ紐付け・手動行は null）/ `name`（R・空 name 行は送信前に除外）/ `inspection_value` / `normal_value` / `unit` / `reference_value` / `sort_order`。`status`/`is_abnormal`/基準値評価はサーバー導出で送信しない。列挙元: `GET /v1/masters/examination-types/:id` の `exam_type_fields`（`id`/`name`/`unit`/`normal_value` は行の pre-fill 元）。

| fieldKey                                                                     | R/O | 型          | F 重点                                                                                                                              |
| :--------------------------------------------------------------------------- | :-- | :---------- | :---------------------------------------------------------------------------------------------------------------------------------- |
| exam_type_id                                                                 | R   | select      | F1 F4 C3-1                                                                                                                          |
| medical_record_id                                                            | O   | id          | create 紐付け F4                                                                                                                    |
| pet_id                                                                       | C   | id          | 承認済み患者変更時のみ PATCH F4                                                                                                     |
| doctor_id                                                                    | R   | select      | FE 必須（「担当医を選択してください」）F1 F4                                                                                        |
| date                                                                         | O/R | date        | FE JST 当日補完・BE required F1 F4                                                                                                  |
| machine                                                                      | O   | text        | F4 F5                                                                                                                               |
| result_summary                                                               | O   | text        | F4 F5                                                                                                                               |
| status                                                                       | O   | enum        | PATCH のみ `pending|in_progress|result_entered|completed|confirmed` F4                                                              |
| items[].exam_type_field_id / name(R) / inspection_value / normal_value / unit / reference_value / sort_order | C | rows | shape 確定済み（上記）。実行前に対象 `exam_type_id` の `exam_type_fields` を run report inventory に列挙する。未完の間は V01 完了不可 |

### vaccination-form（独立）— [V01 §8](V01-clinical-forms.md)

独立フォームも `pet_id`、`medical_record_id`（該当時）、`vaccine_id`、`date`、`lot1..lot4`、`next_date`、`supplemental`、`next_schedule_type`、`remarks` を個別に記録する。未来日接種拒否・次回予定境界（接種日と同日拒否）を含め、タブ表の適用可能な全Fを実施する。

### checkup-form（独立クイック）— [V01 §9](V01-clinical-forms.md)

3段チェーン: `POST /v1/medical-records {pet_id, owner_id, visit_date}` → `POST /medical-records/:mrId/checkups` → 値あり時 `PUT /checkups/:cid/field-results`（checkups-tab と同一 envelope・全置換）。

| fieldKey                                             | R/O | F 重点                                                                                          |
| :--------------------------------------------------- | :-- | :---------------------------------------------------------------------------------------------- |
| pet_id / owner_id / visit_date                       | R/S | 親カルテ自動作成の context。F0                                                                   |
| checkup_type_id                                      | R   | F1 F4                                                                                           |
| date                                                 | R   | F1 F4                                                                                           |
| next_date / doctor_id / result                       | O   | F4 F5                                                                                           |
| results[].checkup_type_field_id / value_*（上記4種） | C   | envelope 確定済み。定義 id/field_type のみ実行時列挙 — 未完の間は完了不可                        |

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
| （画像保存経路） | — | wire key は `style_image` / `completed_image`（string≤2048・同一 POST/PATCH body）で確定済み。ただし FE は File を preview のみに使い upload 手段が存在せず送信されないため永続化 PASS には数えない（既知 gap、[17-trimming-form.md](../../../spec/screens/17-trimming-form.md)） |

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

### clinic-master-side-panel — `/settings/clinic` — [V03 §7](V03-owner-pet-staff-forms.md)

`PATCH /api/v1/clinics/:clinic_id`（builder: `frontend/src/features/clinic-settings/lib/clinic-master-settings-model.ts` `buildUpdateClinicRequest`）。profile 文字列系は `""`→key 省略（既存値保持・クリア不可）、`accounting_document_*` と税率・`is_active` は常時送信。`POST /api/v1/clinics`（create・system_admin 限定）は profile 9 key のみで税率・`is_active`・`accounting_document_*` は非送信（BE 既定適用）。非 admin の他院 PATCH は 403。`logo_url` は BE PATCH が受理するが FE 型・UI とも非送信（ロゴアップロード機能なし）— 計上しない。

| fieldKey                                       | R/O | 型      | F 重点                                                              |
| :--------------------------------------------- | :-- | :------ | :------------------------------------------------------------------ |
| name                                           | R   | string  | FE「院名は必須です」が唯一のガード（BE omitempty は `""` 受理）F1 F4 |
| is_active                                      | O   | boolean | 常時送信。ステータス pill F4                                         |
| postal_code                                    | O   | string  | `""`→省略。BE `jp_postal` →400 F2 F4                                 |
| address                                        | O   | string  | `""`→省略。BE max=500 F4                                             |
| phone_number                                   | O   | string  | `""`→省略。BE `jp_phone` →400 F2 F4                                  |
| fax_number                                     | O   | string  | `""`→省略。BE `jp_phone` 同一 F2 F4                                  |
| registration_number                            | O   | string  | `""`→省略。BE max=100。医院の登録番号（領収書印字用）F4                |
| director_name                                  | O   | string  | `""`→省略。BE max=255 F4                                             |
| email                                          | O   | string  | `""`→省略。FE type=email + BE `jp_email` →400 F2 F4                  |
| website                                        | O   | string  | `""`→省略。BE max=500 F4                                             |
| standard_tax_rate                              | O   | number  | 常時送信。FE 0–100%→÷100・BE service 0–1 range F3 F4                 |
| reduced_tax_rate                               | O   | number  | 同上 F3 F4                                                          |
| accounting_document_show_logo                  | O   | boolean | 常時送信。ロゴ表示 F4                                                |
| accounting_document_show_registration_warning  | O   | boolean | 登録番号警告（既定 true）F4                                          |
| accounting_document_show_item_category         | O   | boolean | 項目カテゴリ F4                                                      |
| accounting_document_show_clinic_header         | O   | boolean | 病院情報ヘッダー F4                                                  |
| accounting_document_show_owner_pet_info        | O   | boolean | 飼主・ペット情報 F4                                                  |
| accounting_document_show_items_table           | O   | boolean | 明細テーブル F4                                                      |
| accounting_document_show_payment_summary       | O   | boolean | お会計サマリー F4                                                    |
| accounting_document_section_order              | O   | string[] | 全キー順列送信（欠落キーは末尾補完）。BE enum+重複拒否→400 F4         |
| accounting_document_footer_note                | O   | string  | 常時送信・`""` クリア可。FE maxLength=500 < BE max=1000（非対称境界）F3 F4 F5 |

---

## V04 設定マスタ（算定保留）

共通 save パイプライン: `useMasterSave`（`frontend/src/features/master/hooks/use-master-save.ts`）— `validate` → `toCreateRequest`/`toUpdateRequest` → `POST /api/v1/masters/<resource>` / `PATCH /{id}`。下表は SidePanel builder が**実際に送出する wire key** のみを列挙する。BE が受理するが FE 非送出の key（`sort_order`・`display_order`・`duration`・`time_condition` 等）は計上せず注記のみ。並び替えは専用ワイヤ `PATCH /api/v1/masters/<resource>/reorder {ids:number[]}` が owner。

**F5 系統的注意（要実測）**: builder の `x || undefined`（key 省略）と `→ null` は Go update map が非 nil のみを拾うため **PATCH では既存値が残りクリア不可**。クリア可は常時送信系（`""`・`[]`・`0` 保存）のみ。対象外: interview-template `content`（`""` 保存）、campaign `target_categories`/`target_item_ids`（`[]`）、insurance `coverage_rate`（空→`0`）、数値系 `price`/`unit_price`（空→`0`）。

**create の `is_active`**: payment-method・trimming-course-type の create POST は BE struct が `is_active` を持たず dead key。animal-species・diagnosis-type/name・trimming-course・trimming-option は create で `true` 固定送信。`is_active` の F4（OFF 永続）は PATCH 経路でのみ意味を持つ。

L-step/LINE settings は V05 が唯一の owner。V04 では実行・集計しない。

### V04 §1 標準マスタ SidePanel 群（16 フォーム）

`name` 系は共通で FE `maxLength=100`（trimming-course-type のみ 50）+ BE `required`。BE 側上限の有無は不統一（max=255 あり: animal-species/merchandise/insurance/occupation/chief-complaint/campaign、上限なし: 他）— F3 の実測はフォーム別。

#### master-animal-species — `/settings/animal-species`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | BE `required,max=255`。グローバル一意（`WHERE is_active=true`） | F1 F3 F4 C3-2 |
| is_active | O | boolean | create `true` 固定 / update フォーム値 | F4 |
| （sort_order） | S | int | create のみ `0` 固定送信・UI 非編集 | F6 |

#### master-diagnosis-type — `/settings/diagnosis?tab=diagnosis_type`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | BE `required`（上限タグなし）。(clinic_id,name) 一意 | F1 F4 C3-2 |
| description | O | string | `|| undefined` → PATCH クリア不可（要実測） | F4 |
| is_active | O | boolean | create `true` 固定 / update フォーム値 | F4 |

#### master-diagnosis-name — `/settings/diagnosis?tab=diagnosis_name`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | BE `required` | F1 F4 |
| diagnosis_type_id | R | uint64 | BE `required`。FE「カテゴリを選択してください」 | F1 F4 C3-1 |
| description | O | string | `|| undefined` → クリア不可 | F4 |
| is_active | O | boolean | create `true` 固定 / update フォーム値 | F4 |

#### master-chief-complaint — `/settings/interview/chief-complaint`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | BE `required,min=1,max=255`。(clinic_id,name) 一意 | F1 F3 F4 C3-2 |
| description | O | string | `|| undefined` → クリア不可 | F4 |
| is_active | O | boolean | フォーム値 | F4 |

#### master-interview-template — `/settings/inquiry-templates`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| category | R | string | BE `required,min=1,max=255`。自由テキスト入力 | F1 F3 F4 |
| title | R | string | BE `required,min=1,max=255`。**wire key は `name` ではなく `title`** | F1 F3 F4 |
| content | O | string | 常時送信・`""` で保存（クリア可） | F4 F5 |
| is_active | O | boolean | フォーム値 | F4 |

#### master-reservation-type-group — `/settings/reservation-type`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | BE `required`（上限なし） | F1 F4 |
| color | O | string | `|| undefined` → クリア不可。BE 制約なし（形式検証なし・要実測） | F2 F4 |
| is_active | O | boolean | フォーム値 | F4 |

#### master-hospitalization-plan — `/settings/hospitalization`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | BE `required`。(clinic_id,name) 一意 | F1 F4 C3-2 |
| price | O | int64 | create `|| undefined`（0→省略）/ update 常時送信（空→0）。BE 上下限なし・負値可（要実測） | F3 F4 |
| body_size | O | enum `small|medium|large` | BE `omitempty,oneof`。update `null`→patch skip→**クリア不可**（要実測） | F2 F4 |
| billing_unit | O | enum `per_day|per_night` | body_size と同様（クリア不可） | F2 F4 |
| tax_type | O | enum `included|excluded|exempt` | BE `omitempty,oneof`。FE 既定 `excluded` | F2 F4 |
| tax_rate | O | float64 | 0.1/0.08 選択。BE `*float64` 範囲タグなし | F4 |
| description | O | string | `|| undefined` → クリア不可 | F4 |
| is_active | O | boolean | フォーム値 | F4 |

#### master-cage — `/settings/cage`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | BE `required`。(clinic_id,name) 一意 | F1 F4 C3-2 |
| cage_type | R | enum `icu|dog|cat|general` | BE `required,oneof`。FE 既定 `general` | F1 F2 F4 |
| cage_size | R | enum `small|medium|large` | BE `required,oneof`。FE 既定 `medium` | F1 F2 F4 |
| price | O | int64 | 常時送信・空→0。BE 上下限なし（負値要実測） | F3 F4 |
| description | O | string | `|| undefined` → クリア不可 | F4 |
| is_active | O | boolean | フォーム値 | F4 |

#### master-merchandise-item — `/settings/merchandise-items`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | BE `required,max=255`。(clinic_id,name) `WHERE is_active=true` 一意 | F1 F3 F4 C3-2 |
| category | R | enum `food|goods|other` | BE `required,oneof`。FE 既定 `goods` | F1 F2 F4 |
| unit_price | O | int64 | 常時送信・空→0。BE `min=0` → 負値 400 | F3 F4 |
| tax_type | R | enum `included|excluded|exempt` | BE `required,oneof`。FE 既定 `excluded` | F1 F2 F4 |
| tax_rate | O | float64 | 0.1/0.08。BE `omitempty,min=0,max=1` | F3 F4 |
| is_active | O | boolean | フォーム値 | F4 |

#### master-insurance — `/settings/insurance`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | BE `required,max=255`。(clinic_id,name) 一意 | F1 F3 F4 C3-2 |
| coverage_rate | O | int | BE `omitempty,min=0,max=100`（0 受理・101/-1 拒否）。FE 同一境界。**空入力→`0` 送信**（省略でなく 0 保存） | F2 F3 F4 |
| description | O | string | `|| undefined` → クリア不可 | F4 |
| contact_phone | O | string | `|| undefined` → クリア不可 | F4 |
| is_active | O | boolean | フォーム値 | F4 |

#### master-occupation — `/settings/occupations`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | BE `required,min=1,max=255`。(clinic_id,name) 一意 | F1 F3 F4 C3-2 |
| description | O | string | `|| undefined` → クリア不可。BE `max=2000` | F3 F4 |
| is_active | O | boolean | フォーム値 | F4 |

#### master-trimming-course — `/settings/trimming?tab=course`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | BE `required`（上限なし）。(clinic_id,name) 一意 | F1 F4 C3-2 |
| price | O | int64 | `toNullableNumber` 空→`null`→PATCH クリア不可（要実測）。BE 上下限なし | F3 F4 |
| target_size | O | enum `small|medium|large|cat` | BE `omitempty,oneof`。`null`→クリア不可 | F2 F4 |
| course_type_id | O | uint64 FK | 空→`undefined`→省略・クリア不可 | F4 C3-1 |
| duration | O | int | 空→`null`→クリア不可 | F4 |
| description | O | string | `|| undefined` → クリア不可 | F4 |
| is_active | O | boolean | create `true` 固定 / update フォーム値 | F4 |

#### master-trimming-option — `/settings/trimming?tab=option`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | BE `required`。(clinic_id,name) 一意 | F1 F4 C3-2 |
| price | O | int64 | 空→`null`→クリア不可 | F3 F4 |
| duration | O | int | 同上 | F4 |
| is_combinable | O | boolean | 常時送信。既定 ON | F4 |
| description | O | string | `|| undefined` → クリア不可 | F4 |
| is_active | O | boolean | create `true` 固定 / update フォーム値 | F4 |

#### master-trimming-course-type — `/settings/trimming-course-type`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | BE `required`（上限なし）。FE `maxLength=50`。(clinic_id,name) 一意 | F1 F3 F4 C3-2 |
| is_active | O | boolean | **create POST は BE が `is_active` を持たず dead key**・PATCH のみ有効 | F4 F6 |

#### master-campaign — `/settings/campaigns`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | BE `required,max=255` | F1 F3 F4 |
| start_date | R | date | BE `required` + DateOnly parse → 非形式 400 | F1 F2 F4 |
| end_date | R | date | 同上。FE `end<start` 拒否 + BE `validateCampaignPeriod` も拒否 | F1 F2 F4 |
| discount_type | R | enum `rate|amount` | BE `required,oneof`。FE 既定 `rate` | F1 F2 F4 |
| discount_value | O | float64 | BE `min=0`。**rate 時の max=100 は HTML 属性のみ → rate>100 も BE 受理**（要実測） | F3 F4 |
| target_categories | O | string[] | BE `omitempty,dive,oneof`。常時送信・`[]`→クリア可 | F4 F5 |
| target_item_ids | O | uint64[] | 常時送信・`[]`→クリア可 | F4 F5 |
| is_active | O | boolean | フォーム値 | F4 |

#### master-payment-method — `/settings/payment-methods`（`/v1/masters/` 外: `/api/v1/payment-methods`）

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | BE `required,max=255`。FE 一覧内重複拒否。(clinic_id,name) `WHERE deleted_at IS NULL` 一意 | F1 F4 C3-2 |
| is_active | O | boolean | **create は dead key（BE struct 非保持）**・PATCH 有効。`system_key` 保持行の OFF は 409 | F4 F6 |
| （system_key） | S | — | **wire 非送信**（FE/BE request ともに存在しない）。DB immutable 識別子・GET 応答のみ | F6 |

### V04 §2 診療項目マスタ 5 タブ — `/settings/treatment-items?tab=…`

5 タブ共通: `name`(R・FE 必須)・`price`(O・`<0` 拒否)・`description`(O・`||undefined`→クリア不可)・`is_active`(O)・`parent_id`(O・truthy 時のみ送信・子持ち時非表示)・`clear_parent_id`(O・PATCH のみ・parentId=""→`true`)。**タブ別 wire 差異**: `tax_type`/`tax_rate` は consultation・procedure のみ送信、`is_non_insurance` は examination のみ、`anesthesia`（`none|local|sedation|general`・BE create required）は procedure のみ。vaccine・checkup は上記 4+parent のみ。**パネルに表示されるが非送信の UI-only 項目を永続 field に数えない**。検査タブは API resource が `checkup-types` でなく `examination-types`。

| タブ | endpoint | 追加 wire key |
| :-- | :-- | :-- |
| consultation | `/v1/masters/consultations` | tax_type(`omitempty,oneof`)・tax_rate(`min=0,max=1`) |
| examination | `/v1/masters/examination-types` | is_non_insurance(bool)。**+ 検査項目サブリソース**: `POST/PATCH …/fields[/:fid] {name(R),inspection_value,normal_value,unit}`・`PUT …/fields/:fid/reference-ranges {ranges:[{animal_species_id(R),ref_min,ref_max,qualitative_min,qualitative_max}]}`（重複種別拒否・numeric XOR qualitative・min≤max）・reorder `PATCH …/fields/reorder {ids[]}` |
| procedure | `/v1/masters/procedures` | tax_type(create `required,oneof`)・tax_rate・anesthesia(create `required,oneof none|local|sedation|general`・既定 `none`) |
| vaccine | `/v1/masters/vaccines` | なし（共通のみ） |
| checkup | `/v1/masters/checkup-types` | なし（共通のみ） |

### V04 §3 master-medicine — `/settings/medicine`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | FE 必須。カテゴリなし・price≤0・剤形なしの組合せは FE 拒否（「親カテゴリ、単価、または剤形のいずれかを入力してください」） | F1 F4 |
| price | O | int64 | カテゴリノードは強制 0・UI disabled。非カテゴリ `min=0` | F3 F4 F6 |
| parent_id / clear_parent_id | O | uint64/bool | clear は PATCH のみ | F5 |
| description | O | string | そのまま送信（`""` 可） | F4 F5 |
| is_active | O | boolean | | F4 |
| dosage_form | O | enum `tablet|liquid|injection|topical|powder` | `""`→undefined | F4 F5 |
| medicine_unit | O | enum `per_tablet|per_ml|per_dose|per_gram` | `""`→undefined | F4 F5 |
| tax_type / tax_rate | O | enum/float | カテゴリは disabled。tax_rate 0–1 | F3 F4 F6 |
| is_non_insurance | O | boolean | | F4 |
| calculation_type | O | enum `none|per_weight` | | F4 F6 |
| strength | C | float64 | `calculation_type≠none` 時のみ送信。BE `gt=0`・per_weight 必須（400） | C F3 F6 |
| frequency_per_day | O | int | 同上 gating。`gt=0` | F3 F6 |
| default_duration_days | O | int | 同上 | F3 F6 |

投与量パラメータ（`calculation_type=per_weight` 時のみ描画）: `PUT /v1/masters/medicines/:id/dose-params/:species`（`:species` は `dog|cat` の URL パス・body 外）/`DELETE` 同パス。body: `dose_basis`(O・`per_administration|per_day`)、`dose_per_kg`(R・`gt=0`)、`min_mg_per_kg`(O・≤max かつ ≤dose_per_kg)、`max_mg_per_kg`(C・absolute_max_dose なし時必須)、`absolute_max_dose`(C・同上)、`rounding_step`+`rounding_mode`(C・セットで有/無・mode は `up|down|nearest`)、`notes`(O)。全項目 F0/F4、`gt=0` 境界 F3、either/or・ペア条件 F6。
**注**: BE PATCH は `clear_strength` を受理するが FE 経路なし — per_weight→none 切替で strength は残存（要実測）。

### V04 §4 master-reservation-type — `/settings/reservation-type`

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | FE 必須 | F1 F4 |
| description | O | string | `""`→undefined | F4 |
| is_active | O | boolean | **create は `true` 固定**（toggle 無視）/ update 実値 | F4 S |
| group_id | O | uint64 | 未選択=「未分類」。**clear_group_id 非送信 → グループ解除不可**（要実測） | F4 F5 |
| reservation_display_name | O | string | `""`→undefined（空は name フォールバック） | F4 F5 |
| duration_minutes | O | int | FE `Number(v)||15` フォールバック（非数値/空/0→15）。min=5 max=480 は属性のみ | F2 F3 F4 |
| short_name | O | string | `""`→undefined | F4 F5 |
| reservation_visible | O | boolean | 予約ページに表示 | F4 |
| reservation_comment | O | string | `""`→undefined | F4 F5 |
| reservation_image_url | O | string | `""`→undefined。形式検証なし | F4 F5 |
| show_short_name | O | boolean | 略称を使用 | F4 |
| reservation_day_option | O | enum `none|weekday|saturday|anyday` | BE `omitempty,oneof` | F4 |
| is_internal | O | boolean | 内部サービス | F4 |

紐付け職種（既存項目のみ）: `POST /v1/masters/reservation-types/:id/occupations {occupation_id(R)}` / `DELETE …/occupations/:linkId`。不可時間帯: `POST …/unavailable-times {unavailable_type(R,weekly|specific), day_of_week?|specific_date?, start_time(R), end_time(R)}` / `DELETE …/unavailable-times/:id`。

### V04 §5 reservation-type-available-slots — `/settings/reservation-type`（§4 パネル内・`/line-reservation/slots` と同一コンポーネント）

`POST /v1/masters/reservation-types/:id/available-slots` / `DELETE …/:slotId`。

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| available_type | R | enum `weekly|specific` | BE `required,oneof` | F1 |
| start_time | R | string `HH:MM` | 15分刻み select。BE required+時刻 parse | F1 F2 |
| is_active | S | boolean | **`true` 固定送信**（UI 非編集） | S |
| day_of_week | C | int8 | weekly のみ送信・BE 必須+0–6 | C |
| specific_date | C | date | specific のみ。**FE 検証なし** → 空は BE 400 | C F2 |
| （重複） | — | — | type+day/date+start_time 重複は BE 409 | — |

### V04 §6 締め時間設定 — `/settings/closing-time`（3 フォーム）

closing-standard-time `PATCH /v1/closing-settings`: `closing_am_pm_boundary`(R・HH:MM・**両終了時刻より前必須**)、`closing_weekday_end`(R)、`closing_sunday_end`(R)、`closed_weekdays`(O・int64[]・0–6 重複なし)。全項目 F1/F2（time 形式）・F4。

closing-holiday `POST /v1/closing-settings/holidays` / `DELETE …/:date`: `date`(R・YYYY-MM-DD)、`reason`(O・`""`→undefined・BE max=500)。同日再 POST は 409（insert-only）。F1/F2/F3/F4。

closing-special-period `POST /v1/closing-settings/special-periods` / `DELETE …/:id`: `note`(O・max=1000)、`start_date`(R)、`end_date`(R・start>end 拒否)、`am_pm_boundary`(R・pm_end より前)、`pm_end`(R)。MasterSidePanel は `<form noValidate>` — required 属性はブラウザ非強制で **BE `binding:"required"` が唯一のゲート**。期間重複は 409。F1(BE)/F2/F3/F4。

### V04 §7 master-shift-template — `/settings/shift-templates`

`POST /v1/shift-templates` / `PATCH /:id`。

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| name | R | string | FE 保存ボタン disabled 制御。BE `required,max=255` | F1 F3 F4 |
| shift_type | R | enum `full|morning|afternoon|off|paid_leave` | BE `required,oneof` | F1 |
| start_time | C | string `HH:MM` | type∈{off,paid_leave} 以外で FE 必須・非表示時 create=undefined/update=null | C F6 F5 |
| end_time | C | string `HH:MM` | 同上 | C F6 F5 |
| breaks[] | O | array | 両端 set のペアのみ送信・非表示型は `[]`。BE `max=50`・各 break_start/break_end `required` | F4 F6 |
| notes | O | string | BE `max=2000` | F3 F4 F5 |
| is_active | O | boolean | StatusPill | F4 |

並び替え: `PATCH /v1/shift-templates/reorder {ids:number[]}`。

### V04 §10 company-invoice-section — `/settings/clinic` 上部

**別 entity・別 endpoint**: `PATCH /api/v1/company`（singleton・id なし）。clinic の `registration_number`（獣医師会番号等・V03 §7 側）とは別物 — 混同注意。

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| invoice_registration_number | O | string | 常時送信・`""` クリア可。**T+13桁の形式検証は FE/BE ともになし**（BE `omitempty,max=100` のみ）→ F2 形式は「検証なし」が正 | F3(>100) F4 F5 |

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
| liff-account-link            | link_token, line_id_token（自動実行・入力欄なし）                               | V05-5・F0 分岐のみ                                    |
| line-reserve-create          | course_id, staff_id, date, start_time, end_time, customer_fields{customer_name, phone, owner_name, pets[]}, request_text, trimming_course_id, trimming_option_ids — 下表参照 | V05-6 |
| line-reserve-cancel          | （body なし・`DELETE /api/liff/{clinicId}/my-reservations/{id}`）               | V05-7・F0 F4                                          |
| line-reservation-settings    | 全量 PUT 25 key（status ほか）+ secret 2 key 非送信 — 下表参照                  | V05-8                                                 |
| line-reservation-page-editor | header_text, request_example, reservation_notice, cancel_notice, privacy_policy（編集は5項目・wire は全量 PUT 25 key） | V05-9（唯一の owner） |
| line-reservation-slots       | available_type(`specific`固定), specific_date, start_time, is_active(`true`固定) | V05-10                                                |
| owner-line-customer-link     | owner_id（紐付け=id・解除=null）                                                | V05-11・F0 F1 F4 F6 C3-2                              |
| lstep-settings               | PATCH 28 key（secret3+text2+enum1+numeric23+bool1）— 下表参照                   | V05-12                                                |
| lstep-trigger-priority       | items[]{trigger_type, priority}（全トリガー一括置換 PATCH）                     | V05-13・F0 F1 F3 F4 F6                                |
| lstep-tag-code-mappings      | entries[]{code_type, codes[]}（tagName 単位全量置換 PUT・4 固定タグ）           | V05-14・F0 F1 F2 F4                                   |
| lstep-tag-config             | 3 追加フォーム: prefix+category / condition_code+tag_name / purpose+tag_prefix — 下表参照 | V05-15                          |
| lstep-csv-import             | file（multipart・パート名 `file`・.csv・BE 上限 ~51MB）                         | V05-16・F0 F1 F3 F4                                   |
| lstep-bulk-tag-remove        | （body なし・tagName/ownerId は URL パス・逐次 DELETE）                         | V05-17・F0 F6                                         |
| lstep-checkup-sync-create    | checkup_type, owner_ids[](1–100), tag_name + preview query 12 key — 下表参照    | V05-18                                                |

### line-reserve-create — line-reserve アプリ（`frontend/line-reserve/`・独立 SPA）— V05-6

`POST /api/liff/{clinicId}/reservations`（Bearer LINE idToken）。409 は `{code:"SLOT_TAKEN", redirect_step}` で FE 画面遷移。BE は `trimming_style_request` も受理するが FE 非送信。

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| course_id | R | number | BE `required`。inactive 区分は BE 拒否 | F0 F1 F4 C3-1 |
| staff_id | R | number | `0`=指名なし。非 0 は BE clinic 所属チェック | F0 F1 F4 |
| date | R | string `YYYY-MM-DD` | BE `required`+DateOnly parse | F0 F1 F2 F4 |
| start_time / end_time | R | string `HHMM` | BE `required` | F0 F1 F2 F4 |
| customer_fields | R | object | BE: JSON ≤10KB・top ≤20 key・各 string ≤500 | F0 F3 F4 |
| customer_fields.customer_name | R | string | FE trim 非空 | F0 F1 F4 |
| customer_fields.phone | R | string | FE `/^[0-9+ ()-]+$/`+数字≥10桁・BE 同規約 | F0 F1 F2 F4 |
| customer_fields.owner_name | O | string | 空可 | F0 F4 F5 |
| customer_fields.pets[] | O | array | `{name, type, is_new}`。新規追加は name 非空 | F0 F1(行内) F4 C3-1 |
| request_text | O | string | BE ≤1000 字。**FE maxLength なし** → 1001 字で 400 | F0 F3 F4 F5 |
| trimming_course_id / trimming_option_ids | O | number/number[] | trimming 分岐時のみ送信 | F0 F4 F6 |

### line-reservation-settings — `/line-reservation/settings` — V05-8

`PUT /v1/clinics/{clinicId}/line-reservation-settings`（clinic 1 レコード・**全量 PUT** — UI 非編集項目も round-trip 送信）。F4 は編集可能項目に適用、round-trip-only 項目は F0+F4（値が消えないことの確認として V05-9 手順3 で代表実施）。

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| status | R | enum `running|stopped` | BE `required,oneof`。停止時 LIFF は MaintenancePage | F0 F1 F4 |
| national_holiday_closed | R | boolean | 祝日休診 | F0 F4 |
| closed_weekdays | R | string[] | 曜日番号 `"0"`–`"6"`（0=日） | F0 F4 |
| closed_dates | R | string[] | `YYYY-MM-DD`。UI 非編集・round-trip | F0 F4 |
| business_hours | R | object `{start,end} HHMM` | FE が `:` 除去して送信 | F0 F1 F2 F4 |
| business_hours_by_weekday | R | object | `{"0".."6":{start,end}}`・トグル OFF で `{}` | F0 F4 F6 |
| break_hours | R | array `[{start,end}]` | `HHMM`。BE 形式必須 | F0 F2 F4 |
| daily_limit / monthly_limit | R | number\|null | UI 非編集・round-trip。BE `min=0,max=100000` | F0 F4 |
| booking_window_min_days | R | number | FE `min=0`・BE `min=0,max=366` | F0 F3 F4 |
| booking_window_max_days | R | number | FE `min=1`・BE `min=0,max=366` | F0 F3 F4 |
| calendar_months | R | number | FE `min=1 max=6`・BE `min=0,max=12`（**境界非対称**） | F0 F3 F4 |
| phone_number | R | string | BE `max=32` | F0 F3 F4 |
| notification_email | R | string | FE type=email・BE `omitempty,email,max=254` | F0 F2 F4 |
| request_example | R | string | UI 非編集（編集は page-editor）。BE `max=2000` | F0 F4 |
| time_slot_mode | R | enum `minimize_gaps|allow_gaps` | BE `required,oneof` | F0 F1 F4 |
| time_slot_interval_minutes | R | number | FE `min=5 step=5`・BE `min=1,max=1440` | F0 F3 F4 |
| no_staff_mode | R | enum `first_available|top_priority` | BE `required,oneof` | F0 F1 F4 |
| show_no_staff_option | R | boolean | UI 非編集・round-trip | F0 F4 |
| additional_fields | R | json | UI 非編集・round-trip（顧客追加項目） | F0 F4 |
| header_text | R | string | UI 非編集。BE `max=2000` | F0 F4 |
| reservation_notice | R | string | UI 非編集。BE `max=10000` | F0 F4 |
| cancel_notice | R | string | UI 非編集。BE `max=10000` | F0 F4 |
| privacy_policy | R | string | UI 非編集。BE `max=100000` | F0 F4 |
| line_channel_id | R | string | BE `max=255` | F0 F4 |
| liff_id | R | string | BE `max=255` | F0 F4 |

**secret-bearing（値は扱わない・キー名のみ）**: `line_channel_secret` — BE request struct に存在しない（故意に非受付・canonical owner は別 API）。`line_access_token` — BE binding はあるが **FE 非送信**・response 非返却（write-only 相当・非含有は FE テストで回帰固定済み）。いずれも F0 のみ・値の永続確認対象外。

### lstep-settings — `/settings/integrations/lstep` — V05-12

`PATCH /v1/clinics/{clinicId}/lstep-settings`。**空欄=変更なし**（`setTrimmedString skipEmpty`）が基本で、正の整数のみ送信する数値系（0/負値は送信せず既存値維持）と混在。`liff_id` のみ空文字クリア可。secret 3 key は response が `*_masked` のみ返却する write-only。

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| lstep_api_key / line_channel_access_token / line_channel_secret | O | string | **secret（write-only・response 非エコー）**。空欄=変更なし | F0 F4 F5 |
| liff_id | O | string | **唯一 `""` でクリア可** | F0 F4 F5 |
| lstep_base_url | O | string | 空欄=変更なし。BE `max=512`+https/allowlisted host（`https://app.lstep.jp`→400） | F0 F2 F4 F5 |
| line_account_name | — | string | BE・FE 型に存在するが **UI 非送信** — 計上せず注記のみ | — |
| cpm_version | O | enum `v1|v2` | payload に入るのはこの2値のみ | F0 F4 |
| dormant_prevention_180/210/240/365_days・health_prevention_lookback_days・vaccine_deadline_days・cpm_v2_*_threshold・cpm_v1_*（23 数値キー） | O | int/int64 | `setPositiveInteger` — ≥1 のみ送信（0/負値は既存値維持） | F0 F3 F4 |
| is_sync_enabled | R | boolean | **常時送信**。無効化は ConfirmDialog 経由 | F0 F4 F6 |

ボタン類（永続しない操作）: `POST …/lstep-settings/test-connection`・`DELETE …/lstep-settings`（ConfirmDialog）。

### lstep-tag-config — `/settings/integrations/lstep` 内セクション — V05-15

**追加フォームは 3 種**（POST/DELETE は `requireSystemAdmin` — UAT 実行権限の前提）。

| formId | endpoint | fieldKey | R/O | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| auto-managed-prefixes | `POST /v1/lstep-tag-config/auto-managed-prefixes` | prefix | R | F0 F1 F4 C3-2（重複 409） |
| | | category | R enum `B|C1|C2|C3` | F0 F1 F2 F4 |
| condition-tag-mappings | `POST /v1/lstep-tag-config/condition-tag-mappings` | condition_code | R・BE `max=50`・同一コード 409 | F0 F1 F3 F4 C3-2 |
| | | tag_name | R | F0 F1 F4 |
| send-purpose-tag-prefixes | `POST /v1/lstep-tag-config/send-purpose-tag-prefixes` | purpose | R（V05 表に第3フォームとして追記済み） | F0 F1 F4 |
| | | tag_prefix | R | F0 F1 F4 |
| 行削除 | `DELETE …/{type}/{id}` → 204 | — | — | F0 F4 |

（`description` は BE が受理するが UI 非送信 — 計上しない）

### lstep-checkup-sync-create — `/lstep/checkup-sync` — V05-18

永続化: `POST /v1/clinics/{clinicId}/lstep/checkup-sync`。

| fieldKey | R/O | 型 | 制約・特記 | F 重点 |
| :-- | :-- | :-- | :-- | :-- |
| checkup_type | R | enum `annual|dental|blood|skin|cancer|other` | BE `required`+enum | F0 F1 F4 |
| owner_ids | R | string[] | BE `required,min=1,max=100`。FE 上限 100 で disabled | F0 F1 F3 F4 F6 |
| tag_name | R | string | FE trim 非空。BE `^[a-zA-Z0-9_\-]{1,100}$`+システム管理タグ拒否 | F0 F1 F2 F3 F4 |

前段プレビュー（非永続 GET `/checkup-sync/preview`・query keys — F0/F2/F5 のみ適用）: `checkup_type`(R)・`species`・`last_visit_after`/`last_visit_before`(date)・`min_age_years`/`max_age_years`(≥0・min>max エラー)・`has_chronic_condition`(`true|false`)・`cpm_stage`(`cpm_encounter|cpm_growing|cpm_core|cpm_spot|cpm_noah|cpm_dormant`)・`min_total_amount`・`min_annual_visit_count`・`last_checkup_after`/`last_checkup_before`(date)。

---

## カバレッジ更新チェックリスト（開発者）

- [ ] 新規永続フォーム → 本ファイルに formId + fieldKey 追加 + 該当 V に § 追加
- [ ] 既存フォームに項目追加 → fieldKey 行追加
- [ ] 必須/境界変更 → R/O と F 重点を更新
- [ ] route inventory は 86 product pages。page 数と unique persistent form 数を混同しない
- [ ] wildcard / UI 全項目 / 動的 placeholder が残る間は inventory incomplete とし、全フォーム完了や総数を主張しない
