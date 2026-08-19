# フォーム×項目 棚卸し (Form Field Inventory)

> **目的**: 受入 V シリーズがカバーする全永続化フォームと、**項目単位 F プロトコル**の対象一覧を定義する。  
> **使い方**: 左の項目を 1 行ずつ [FIELD-LEVEL-PROTOCOL.md](FIELD-LEVEL-PROTOCOL.md) で実施。手順の補足は V01〜V05。  
> **更新規則**: 画面に入力項目を追加したら、本表と該当 V を同 PR で更新する。  
> **最新更新**: 2026-08-14  
> **合計**: **85 フォーム**（旧 84 + inventory-form）

凡例: **R**=必須 / **O**=任意 / **C**=条件付き必須 / **S**=システム（入力不可→F は N/A）

---

## V01 臨床（18）

### medical-record-form — `/medical-records/new|/:id` — [V01 §1](V01-clinical-forms.md)

| fieldKey | ラベル概要 | R/O | 型 | 制約・特記 | F 重点 |
|:--|:--|:--|:--|:--|:--|
| pet_id | 対象ペット | R | id | new は query。無しは select-pet | F0 |
| attending_vet | 担当医 | O | select | 即時保存の要実測あり | F0 F4 |
| visit_type | 来院種別 | O | select/enum | 同上 | F0 F4 |
| soap_s | SOAP-S | O | text | 空保存可の要実測 | F0 F4 F5 |
| soap_o | SOAP-O | O | text | | F0 F4 F5 |
| soap_a | SOAP-A | O | text | | F0 F4 F5 |
| soap_p | SOAP-P | O | text | | F0 F4 F5 |
| chief_complaint_type | 主訴区分 | O | select FK | | F0 F4 C3-1 |
| chief_complaint | 主訴 | O | text | | F0 F4 |
| treatment_policy | 治療方針 | O | text | | F0 F4 |
| diagnosis1_type | 診断1区分 | O | select | | F0 F4 |
| diagnosis1_name | 診断1名称 | O | select FK | 区分連動 | F0 F4 |
| diagnosis2_* / diagnosis3_* | 診断2・3 | O | select | 同上セット | F0 F4 |

### medical-record-treatments-tab — 治療タブ — [V01 §2](V01-clinical-forms.md)

| fieldKey | ラベル概要 | R/O | 型 | 制約・特記 | F 重点 |
|:--|:--|:--|:--|:--|:--|
| item_id | 項目（処置/薬剤） | R | select FK | | F0 F1 F4 |
| quantity | 数量 | R | number | >0。薬量 hard gate | F1 F3 F4 |
| unit_price | 単価 | O | money | ≥0 | F3 F4 |
| dose_override_reason | 上限超過理由 | C | text | 絶対上限時ブロック | F1 F6 |

### medical-record-vitals — VitalsModal — [V01 §3](V01-clinical-forms.md)

| fieldKey | ラベル概要 | R/O | 型 | 制約・特記 | F 重点 |
|:--|:--|:--|:--|:--|:--|
| recorded_at | 記録日時 | R | datetime | 未来日は要実測 | F1 F2 F4 |
| temperature | 体温 | O | number | 範囲ガード要実測 | F3 F4 F5 |
| weight_kg | 体重 | O | number | kg/g 切替 | F3 F4 |
| heart_rate / respiratory_rate / etc. | その他バイタル | O | number | 実装にある計測値すべて | F0 F4 F5 |

> 実装上の計測フィールドはモーダル表示ラベルを inventory 実行時に全列挙し、本表へ追記してから F4 する（【要実測】昇格）。

### medical-record-checkups-tab — [V01 §4](V01-clinical-forms.md)

| fieldKey | R/O | 型 | F 重点 |
|:--|:--|:--|:--|
| performed_on | R | date | F1 F4 |
| checkup_type_id | R | select FK | F1 F4 C3-1 |
| dynamic_* | C | 定義依存 | 定義ごと F0 F1 F4 |
| result_note | O | text | F4 F5 |
| next_due_on | O | date | F4 F5 |

### medical-record-vaccination-tab — [V01 §5](V01-clinical-forms.md)

| fieldKey | R/O | 型 | F 重点 |
|:--|:--|:--|:--|
| vaccine_id | R | select | F1 F4 |
| vaccinated_on | R | date | F1 F4 |
| lot1..lot4 | O | text | F4 F5 |
| note | O | text | F4 F5 |
| next_schedule_mode | O | radio | F4 |
| next_due_on | O | date | F4（手入力永続） |

### medical-record-image-upload / addendum / examination-import / estimate-tab — [V01 §6](V01-clinical-forms.md)

| formId | fieldKey | R/O | 型 | F 重点 |
|:--|:--|:--|:--|:--|
| medical-record-image-upload | file | R | file | F0 F2(MIME) F3(10MB) F4 |
| medical-record-addendum | content | R | text | F1 F4 |
| medical-record-addendum | reason | R | text | F1 F3(500) F4 |
| medical-record-examination-import | examination_ids | R | multi | F1 F4 |
| medical-record-estimate-tab | title / lines | O | text/grid | F0 F4 |

### examination-form — [V01 §7](V01-clinical-forms.md)

| fieldKey | R/O | 型 | F 重点 |
|:--|:--|:--|:--|
| exam_type_id | R | select | F1 F4 C3-1 |
| staff_id | R | select | F1 F4 |
| performed_on | O/R | date | 仕様に従い F1/F4 |
| dynamic_result_* | C | number/text | 定義ごと F0 F4 |

### vaccination-form（独立）— [V01 §8](V01-clinical-forms.md)

タブと同型 + 未来日接種拒否・次回予定境界（接種日と同日拒否）。項目は vaccination-tab に準じ **全項目 F**。

### checkup-form（独立クイック）— [V01 §9](V01-clinical-forms.md)

| fieldKey | R/O | F 重点 |
|:--|:--|:--|
| checkup_type_id | R | F1 F4 |
| performed_on | R | F1 F4 |
| dynamic_* | C | F0 F4 |

### hospitalization-form — [V01 §10](V01-clinical-forms.md)

| fieldKey | R/O | 型 | F 重点 |
|:--|:--|:--|:--|
| pet_id | R | id | F0 F1 |
| admitted_at | R | datetime | F1 F4 |
| planned_discharge_at | O | datetime | F4 F5 |
| cage_id | O | select FK | F4 C3-1 |
| plan_id | O | select FK | F4 C3-1 |
| note | O | text | F4 F5 |

### hospitalization-care-plan / daily-vitals / daily-care-logs / daily-staff-notes — [V01 §11](V01-clinical-forms.md)

| formId | 必須 fieldKey | その他 | F 重点 |
|:--|:--|:--|:--|
| hospitalization-care-plan | name | type, timing | 各 F1/F4 |
| hospitalization-daily-vitals | date + 計測1+ | | F1 F4 |
| hospitalization-daily-care-logs | date, content | | F1 F4 |
| hospitalization-daily-staff-notes | date, content | | F1 F4 |

### trimming-form — [V01 §12](V01-clinical-forms.md)

| fieldKey | R/O | F 重点 |
|:--|:--|:--|
| pet_id | R | F1 |
| course_id | R/O | F0 F4 C3-1・無効マスタ #228 |
| option_ids | O | multi F4 |
| staff_id | O | F4 |
| scheduled_at | R | F1 F4 |
| note | O | F4 F5 |

---

## V02 会計・予約・在庫（12）

### accounting-settlement-form — [V02 §1](V02-accounting-reservation-forms.md)

| fieldKey | R/O | 型 | F 重点 |
|:--|:--|:--|:--|
| payment_splits[].method | R | select | F1 F4（method 重複禁止） |
| payment_splits[].amount | R | money | F1 F3(≥1) |
| cash_tendered | C | money | 現金時 預り≥金額 |
| change_override | O | money | ≥0 |
| post_close_reason | C | text | 締め後必須 F1 |

### accounting-item-add-dialog — [V02 §2](V02-accounting-reservation-forms.md)

| fieldKey | R/O | F 重点 |
|:--|:--|:--|
| name（手動） | R | F1 F4 |
| unit_price | R | F3(≥0) F4 |
| quantity | R | F3(>0) F4 |
| category | R（手動） | F1 F4 |
| other_reason | C | category=other 時 F1 |
| merchandise_item_id | R（マスタタブ） | F1 F4 C3-1 |
| tax_rate | O | F4 |

### credit-correction-dialog — [V02 §3](V02-accounting-reservation-forms.md)

| fieldKey | R/O | F 重点 |
|:--|:--|:--|
| corrected_amount | R | F1 F3(≥1) F4 |
| reason | R | F1 F4 |

### refund-dialog — [V02 §4](V02-accounting-reservation-forms.md)

| fieldKey | R/O | F 重点 |
|:--|:--|:--|
| amount | R | F1 F3(1..残額) F4 |
| method | R | 使用済み手段のみ F0 F4 |
| reason | O | F4 F5 |

### cash-register-close-form — [V02 §5](V02-accounting-reservation-forms.md)

| fieldKey | R/O | F 重点 |
|:--|:--|:--|
| close_date | R | F0 F4 |
| period | R | am/pm/emg F0 F4 |
| actual_cash | R | F1 F3(≥0) F4 |

### estimate-form — [V02 §6](V02-accounting-reservation-forms.md)

| fieldKey | R/O | F 重点 |
|:--|:--|:--|
| title | R | F1 F4 |
| status | R | 作成時 draft/sent のみ |
| amounts.* | O | ≥0 F3 F4 |
| owner/pet/record links | O | F4 |
| valid_until | O | F4 F5 |
| comment | O | F4 F5 |
| discount | O | 権限で F6 |

### reservation-form-modal / reception-walkin / reception-status — [V02 §7–9](V02-accounting-reservation-forms.md)

| fieldKey | R/O | F 重点 |
|:--|:--|:--|
| pet_id / owner | R | F1 F4 |
| reservation_type_id | R | F1 F4 C3-1 |
| start_at / end_at | R | F1 F4・枠衝突 |
| staff_id | O | F4 |
| memo | O | F4 F5 |
| status（受付） | R | 遷移のみ F0 F4 |

### shift-form-dialog — [V02 §10](V02-accounting-reservation-forms.md)

| fieldKey | R/O | F 重点 |
|:--|:--|:--|
| staff_id | R | F1 F4 |
| date | R | F1 F4 |
| start/end | R | F1 F4 |
| template_id | O | F4 C3-1 |

### clinic-holiday-modal — [V02 §11](V02-accounting-reservation-forms.md)

| fieldKey | R/O | F 重点 |
|:--|:--|:--|
| date | R | F1 F4 |
| reason | O | F4 F5 |

### inventory-form — `/inventory/new|/:id` — **[V02 §12 新設](V02-accounting-reservation-forms.md)**

| fieldKey | ラベル | R/O | 型 | 制約 | F 重点 |
|:--|:--|:--|:--|:--|:--|
| name | 品名 | R | text | 非空 | F0 F1 F4 |
| category | カテゴリ | R | enum | medicine/consumable/food/other | F0 F1 F4 |
| unit | 単位 | R | text | 非空（HTML required + BE required） | F0 F1 F4 |
| quantity | 現在庫数 | R | int | ≥0 | F1 F3 F4 |
| minStockLevel | 最低在庫数 | R | int | ≥0 | F1 F3 F4 |
| location | 保管場所 | O | text | | F4 F5 |
| expiryDate | 使用期限 | O | date | | F4 F5 |
| supplier | 仕入先 | O | text | | F4 F5 |
| lastRestocked | 最終入庫日 | O | date | | F4 F5 |

---

## V03 飼主・組織（7）

### owner-create-edit — [V03 §1](V03-owner-pet-staff-forms.md)

| fieldKey | R/O | 型 | F 重点 |
|:--|:--|:--|:--|
| owner_name | R | text | F1 F4 |
| owner_name_kana | R(new)/O(edit) | text | F1/F4 |
| phone | R | phone | F1 F2 F4 C3-2 |
| email | O | email | F2 F4 F5 C3-2 |
| postal_code / address* | O | postal/text | F2 F4 |
| home_postal_code / home_address* | O | | F2 F4 |
| company / company_phone | O | | F4 F5 |
| membership_type | R | enum4 | F0 F4 |
| discount_rate | O | 0–100 | F3 F4 F5 |
| is_dangerous | O | bool | F4 |
| birth_date | O | date | F4 F5（null PATCH） |
| note | O | text | F4 F5 |
| clinic_id（登録先） | O | select | F4 C3-1 |

### pet-edit-modal / pet-add-pending — [V03 §2–3](V03-owner-pet-staff-forms.md)

| fieldKey | R/O | F 重点 |
|:--|:--|:--|
| name | R | F1 F4 |
| animal_species_id | R | F1 F4 C3-1 |
| sex | R(FE) | F1 F4 |
| breed | O | F4 F5 |
| birth_date | O | F4 F5 |
| weight | O | 0–200 FE F3 F4 |
| color | O | F4 F5 |
| microchip | O | F4 F5 |
| blood_type | O | F4 F5 |
| neutered_on | O | F4 F5 |
| food / environment | O | F4 F5 |
| insurance_id | O | F4 C3-1 |
| danger_level | O | enum F4 |
| danger_reason | C | high 時必須 F1 F4 |

### pet-deceased-dialog — [V03 §4](V03-owner-pet-staff-forms.md)

| fieldKey | R/O | F 重点 |
|:--|:--|:--|
| deceased_at | R | F1 F2(未来拒否) F4 |
| reason | O | F4 F5 |

### staff-side-panel — [V03 §5](V03-owner-pet-staff-forms.md)

| fieldKey | R/O | F 重点 |
|:--|:--|:--|
| name | R | F1 F4 |
| email | O | F2 F4 C3-2 |
| password | C | email 時必須・8+英数混在 F1 F2 F3 |
| occupation_id | O | F4 C3-1 |
| permission_group_ids | O | 2段階保存 F4 |
| clinic_ids | O | F4 |
| reservation_capabilities | O | F4 |
| line_display_name | O | F4 F5 |

### permission-group-side-panel — [V03 §6](V03-owner-pet-staff-forms.md)

| fieldKey | R/O | F 重点 |
|:--|:--|:--|
| name | R | F1 F4 C3-2 |
| description | O | F4 F5 |
| color | O | F4 |
| permissions[resource][action] | O | **全 resource × view/create/edit/delete** を F4（ON/OFF 代表） |

### clinic-master-side-panel — [V03 §7](V03-owner-pet-staff-forms.md)

| fieldKey | R/O | F 重点 |
|:--|:--|:--|
| name | R | F1 F4 |
| postal/address/phone/email | O | F2 F4 |
| standard_tax_rate / reduced_tax_rate | O | 0–100% F3 F4 |
| accounting_document_* toggles / footer / order | O | **表示されている帳票設定項目すべて** F4 |

---

## V04 設定マスタ（30）

標準 SidePanel 行は [V04 §1 差分表](V04-settings-master-forms.md) を正とし、各行について:

| 共通 fieldKey | R/O | F 重点 |
|:--|:--|:--|
| name（または文書上の必須名） | R | **必ず F1 + F4** |
| is_active | O | F4 |
| sort_order（D&D） | O | F4 |
| 差分表の追加カラムすべて | 各行 | 型に応じ F2 F3 F4 |

追加で項目単位が厚いフォーム:

| formId | 追加項目（すべて F0+F4、制約あるものは F1–F3） |
|:--|:--|
| master-medicine | 剤形・単位・価格・strength・回数/日・既定日数・calculation_type・dose params |
| master-treatment-item | 名称・価格・説明・親・tab 種別（5 タブ） |
| master-insurance | name・補償率 0–100 |
| master-campaign | name・期間・対象商品 |
| master-payment-method | name・system_key 行の F6 |
| master-reservation-type | 区分名・グループ・職種・公開・色 等 UI 全項目 |
| reservation-type-available-slots | 枠時刻・日付 |
| master-shift-template | テンプレ名・時間帯 |
| closing-settings 系 | 画面上の全入力 |
| lstep settings（V04 重複分） | V05 と役割分担し、片方で F 完了すれば他方は参照 |

実行時: SidePanel を開き **DOM/スナップショット上の入力コントロールを全列挙**し、本 inventory に欠けがあれば追記してから F を回す。

---

## V05 認証・LINE（18）

| formId | 主要 fieldKey（すべて F 適用） | 参照 |
|:--|:--|:--|
| auth-login | email, password | V05-1 |
| auth-change-password | current, new, confirm | V05-2 |
| auth-forgot-password | email | V05-3 |
| auth-reset-password | password, confirm | V05-4 |
| liff-account-link | （自動・入力なし） | F0 分岐のみ |
| line-reserve-create | customer_*, course, staff, date, time, request | **ステップ上の全入力** V05-6 |
| line-reserve-cancel | cancel action | F0 F4 |
| line-reservation-settings | 稼働・受付ルールの全入力 | V05-8 |
| line-reservation-page-editor | テキストエリア全件 | V05-9 |
| line-reservation-slots | 日付・開始時刻 | V05-10 |
| owner-line-customer-link | link/unlink 操作 | V05-11 |
| lstep-settings | secret×3, text, numeric 閾値 | V05-12 |
| lstep-tag-config 等 | 差分表の必須ペア | V05-13–17 |
| lstep-checkup-sync-create | 抽出条件 + tag_name | V05-18 |

---

## カバレッジ更新チェックリスト（開発者）

- [ ] 新規永続フォーム → 本ファイルに formId + fieldKey 追加 + 該当 V に § 追加  
- [ ] 既存フォームに項目追加 → fieldKey 行追加  
- [ ] 必須/境界変更 → R/O と F 重点を更新  
- [ ] route-inventory（84 product pages）と form 数がずれる場合、本ファイルの「合計」コメントで理由を残す（page ≠ form）  
