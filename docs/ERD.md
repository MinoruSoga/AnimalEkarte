# 動物病院管理システム ER図 (Entity Relationship Diagram)

本ドキュメントは、システム内の全エンティティとその関連を定義します。

> **Note**: 本ドキュメントは**目標設計（Target Design）**です。
> 現在実装されているテーブルは `backend/migrations/001_init.sql` を参照してください。
> GORM AutoMigrate で自動実行されます。詳細は [MIGRATION.md](./MIGRATION.md) を参照。

---

## Mermaid ER図

```mermaid
erDiagram
    %% ===== Core Entities =====

    Owner {
        string ownerId PK
        string ownerName
        string ownerNameKana
        string company
        string postalCode
        string address1
        string address2
        string homePostalCode
        string homeAddress1
        string homeAddress2
        string birthDate
        string phone
        string companyPhone
        string email
        string remarks
        boolean isDangerous
        number discountRate
        MembershipType membershipType
    }

    Pet {
        string id PK
        string ownerId FK
        string ownerName
        string petNumber
        string name
        string petNameKana
        PetSpecies species
        PetGender gender
        PetStatus status
        string birthDate
        string breed
        string color
        string weight
        string neuteredDate
        AcquisitionType acquisitionType
        DangerLevel dangerLevel
        string food
        string environment
        string phone
        string lastVisit
        string insuranceName
        string insuranceDetails
    }

    MedicalRecord {
        string id PK
        string recordNo
        string date
        string ownerId FK
        string ownerName
        string petId FK
        string petName
        PetSpecies species
        string chiefComplaint
        string treatmentPolicy
        string physicalExam
        string diagnosisDetails
        string diagnosis1Category FK
        string diagnosis1Name FK
        string diagnosis2Category FK
        string diagnosis2Name FK
        string doctor
        MedicalRecordStatus status
    }

    Hospitalization {
        string id PK
        string hospitalizationNo
        string ownerId FK
        string ownerName
        string petId FK
        string petName
        PetSpecies species
        HospitalizationType hospitalizationType
        string startDate
        string endDate
        HospitalizationStatus status
        string cageId FK
        string doctorName
        string memo
        string ownerRequest
        string staffNotes
    }

    CarePlanItem {
        string id PK
        string hospitalizationId FK
        CarePlanType type
        string name
        string description
        PlanTiming[] timing
        CarePlanStatus status
        string notes
        string masterId FK
        number unitPrice
        string category
    }

    DailyRecord {
        string id PK
        string hospitalizationId FK
        string date
    }

    VitalRecord {
        string id PK
        string time
        number temperature
        number heartRate
        number respirationRate
        number weight
        string notes
        string staff
    }

    CareLogRecord {
        string id PK
        string time
        CareLogType type
        CareLogStatus status
        string value
        string staff
        string notes
    }

    StaffNoteRecord {
        string id PK
        string time
        string content
        string staff
    }

    Accounting {
        string id PK
        string medicalRecordId FK
        string ownerId FK
        string ownerName
        string petId FK
        string petName
        PetSpecies petSpecies
        AccountingStatus status
        string scheduledDate
        string completedAt
        string memo
    }

    AccountingItem {
        string id PK
        string code
        ItemCategory category
        string name
        number unitPrice
        number quantity
        TaxRate taxRate
        boolean isInsuranceApplicable
        ItemSource source
    }

    PaymentInfo {
        number subtotal
        number taxTotal
        number totalAmount
        string insuranceName
        number insuranceRatio
        number insuranceAmount
        number discountAmount
        number billingAmount
        number receivedAmount
        number changeAmount
        PaymentMethod method
    }

    ReservationAppointment {
        string id PK
        Date start
        Date end
        string ownerName
        string petName
        string petId FK
        VisitType visitType
        string type
        string doctor
        boolean isDesignated
        ReservationStatus status
        string notes
    }

    TrimmingRecord {
        string id PK
        string date
        string petId FK
        string petNumber
        string petName
        string ownerName
        PetSpecies species
        string weight
        string styleRequest
        string staff
        TrimmingStatus status
        string courseId FK
        string[] optionIds FK
        string bw
        BodyWeightUnit bwUnit
        string bt
        string usedShampoo
        string usedRibbon
        string treatment
        string medicine
        string charge
        string remarks
        string finalCheck
        string styleImage
        string completedImage
    }

    ExaminationRecord {
        string id PK
        string medicalRecordId FK
        string petId FK
        string date
        string ownerName
        string petName
        string testType
        string doctor
        ExaminationStatus status
        string resultSummary
        string machine
    }

    ExaminationRecordItem {
        string name
        string inspectionValue
        string normalValue
        string result
    }

    VaccinationRecord {
        string id PK
        string medicalRecordId FK
        string petId FK
        string ownerName
        string petName
        string vaccineName
        string date
        string nextDate
        NextScheduleType nextScheduleType
        string doctor
        string supplemental
        string lot1
        string lot2
        string lot3
        string lot4
        string remarks
    }

    MasterItem {
        string id PK
        string code
        string name
        string category
        number price
        MasterItemStatus status
        string description
        string inventoryId FK
        number defaultQuantity
        string species
        string interval
        string parentId FK
    }

    MasterItemInspection {
        string name
        string inspectionValue
        string normalValue
    }

    InventoryItem {
        string id PK
        string name
        InventoryCategory category
        number quantity
        string unit
        number minStockLevel
        string location
        string expiryDate
        string supplier
        string lastRestocked
        InventoryStatus status
    }

    ClinicInfo {
        string name
        string branchName
        string postalCode
        string address
        string phoneNumber
        string faxNumber
        string registrationNumber
        string directorName
        string email
        string website
        string logoUrl
    }

    Appointment {
        string id PK
        string time
        string ownerName
        PetSpecies petType
        string petName
        AppointmentVisitType visitType
        string serviceType
        NextAppointmentStatus nextAppointment
        boolean isDesignated
        string doctor
        string petId FK
    }

    TreatmentPlan {
        string id PK
        string treatmentContent
        string memo
        boolean insurance
        number unitPrice
        number quantity
        number discount
        number discountAmount
        number subtotal
    }

    TreatmentItem {
        string id PK
        boolean selected
        TreatmentStatus status
        string content
        string memo
        boolean insurance
        number unitPrice
        number quantity
        number discountRate
        number discountAmount
        string inventoryId FK
    }

    VitalEntry {
        string id PK
        string recordedAt
        string staff
        number temperature
        number heartRate
        number respirationRate
        number weight
        string notes
    }

    %% ===== Relationships =====

    Owner ||--o{ Pet : "飼育"
    Pet ||--o{ MedicalRecord : "診療記録"
    Pet ||--o{ Hospitalization : "入院記録"
    Pet ||--o{ TrimmingRecord : "トリミング記録"
    Pet ||--o{ ReservationAppointment : "予約"
    Pet ||--o{ Appointment : "来院"

    MedicalRecord ||--o{ ExaminationRecord : "検査"
    MedicalRecord ||--o{ VaccinationRecord : "予防接種"
    MedicalRecord ||--o{ TreatmentItem : "治療プラン/実施"
    MedicalRecord ||--o{ VitalEntry : "バイタル記録"
    MedicalRecord |o--o| Accounting : "会計紐付け"

    Hospitalization ||--o{ CarePlanItem : "ケアプラン"
    Hospitalization ||--o{ DailyRecord : "日次記録"
    Hospitalization ||--o{ TreatmentPlan : "治療プラン"

    DailyRecord ||--o{ VitalRecord : "バイタル"
    DailyRecord ||--o{ CareLogRecord : "ケアログ"
    DailyRecord ||--o{ StaffNoteRecord : "スタッフメモ"

    Accounting ||--o{ AccountingItem : "明細"
    Accounting |o--|| PaymentInfo : "支払情報"

    ExaminationRecord ||--o{ ExaminationRecordItem : "検査項目"

    MasterItem ||--o{ MasterItemInspection : "検査項目定義"
    MasterItem |o--o| InventoryItem : "在庫連携"
    MasterItem |o--o| MasterItem : "親子関係(diagnosis)"
    MasterItem |o--o{ CarePlanItem : "ケアプラン連携"
    MasterItem |o--o{ TreatmentItem : "治療連携"
    MasterItem |o--o{ TrimmingRecord : "コース/オプション連携"

    MedicalRecord }o--o{ MasterItem : "診断名参照"
```

---

## エンティティ一覧

### コアエンティティ

| # | エンティティ | 定義場所 | 説明 |
|---|-------------|---------|------|
| 1 | `Owner` | `features/owners/types` | 飼主（顧客） |
| 2 | `Pet` | `types/index.ts` + `features/owners/types` | ペット（患者）。一覧用の軽量IF(`Pet`)と編集用の完全IF(`PetFormData`/`PetInfo`)に分離 |
| 3 | `MedicalRecord` | `types/index.ts` | 電子カルテ |
| 4 | `Hospitalization` | `types/index.ts` | 入院/ホテル記録 |
| 5 | `Accounting` | `features/accounting/types` | 会計レコード |
| 6 | `ReservationAppointment` | `types/index.ts` | 予約 |
| 7 | `TrimmingRecord` | `types/index.ts` + `features/trimming/types` | トリミング記録。一覧用(`TrimmingRecord`)とフォーム用(`TrimmingFormData`)に分離 |
| 8 | `ExaminationRecord` | `types/index.ts` | 検査記録 |
| 9 | `VaccinationRecord` | `types/index.ts` + `features/medical-records/types` | 予防接種記録。フォーム用の追加フィールドは`VaccinationFormData`に分離 |
| 10 | `MasterItem` | `types/index.ts` | マスタデータ |
| 11 | `InventoryItem` | `types/index.ts` | 在庫品目 |
| 12 | `ClinicInfo` | `features/clinic/types` | 病院情報（シングルトン） |

### 入院サブエンティティ

| # | エンティティ | 親 | 説明 |
|---|-------------|-----|------|
| 13 | `CarePlanItem` | `Hospitalization` | ケアプラン項目 |
| 14 | `DailyRecord` | `Hospitalization` | 日次記録コンテナ |
| 15 | `VitalRecord` | `DailyRecord` | バイタルサイン |
| 16 | `CareLogRecord` | `DailyRecord` | ケアログ（食事/排泄/投薬等） |
| 17 | `StaffNoteRecord` | `DailyRecord` | スタッフメモ |
| 18 | `TreatmentPlan` | `Hospitalization` | 入院治療プラン |

### 電子カルテサブエンティティ

| # | エンティティ | 親 | 説明 |
|---|-------------|-----|------|
| 19 | `TreatmentItem` | `MedicalRecord` | 治療/処置項目 |
| 20 | `VitalEntry` | `MedicalRecord` | カルテ内バイタル |
| 21 | `ExaminationRecordItem` | `ExaminationRecord` | 検査結果項目 |

### 会計サブエンティティ

| # | エンティティ | 親 | 説明 |
|---|-------------|-----|------|
| 22 | `AccountingItem` | `Accounting` | 会計明細行 |
| 23 | `PaymentInfo` | `Accounting` | 支払情報（1:1） |

### マスタサブエンティティ

| # | エンティティ | 親 | 説明 |
|---|-------------|-----|------|
| 24 | `MasterItemInspection` | `MasterItem` | 検査マスタの検査項目定義 |

### ダッシュボードエンティティ（ビューモデル）

| # | エンティティ | 説明 |
|---|-------------|------|
| 25 | `Appointment` | カンバンカード（ダッシュボード用ビューモデル） |
| 26 | `ColumnData` | カンバンカラム（ビューモデル、Mermaid図からは省略） |

---

## リレーション詳細

### 1. Owner ↔ Pet（1:N）

```
Owner.ownerId = Pet.ownerId
```
- 1人の飼主が複数のペットを飼育
- Pet側にownerName等の非正規化フィールドあり（表示用）

### 2. Pet ↔ MedicalRecord（1:N）

```
Pet.id = MedicalRecord.petId
```
- 1匹のペットに対して複数の診療記録
- MedicalRecord側にpetName, ownerName等の非正規化フィールドあり

### 3. Pet ↔ Hospitalization（1:N）

```
Pet.id = Hospitalization.petId
```
- 1匹のペットに対して複数の入院記録

### 4. MedicalRecord ↔ Accounting（1:0..1）

```
Accounting.medicalRecordId = MedicalRecord.id
```
- 診療記録に対して最大1件の会計レコード
- 会計レコードはカルテなしでも作成可能（物販等）

### 5. MedicalRecord ↔ ExaminationRecord（1:N）

```
ExaminationRecord.medicalRecordId = MedicalRecord.id
```
- 1つのカルテに複数の検査を紐付け

### 6. MedicalRecord ↔ VaccinationRecord（1:N）

```
VaccinationRecord.medicalRecordId = MedicalRecord.id
```
- 1つのカルテに複数のワクチン接種記録

### 7. Hospitalization ↔ CarePlanItem（1:N）

```
CarePlanItem.hospitalizationId = Hospitalization.id
```
- 入院レコードに対して複数のケアプランアイテム

### 8. Hospitalization ↔ DailyRecord（1:N）

```
DailyRecord.hospitalizationId = Hospitalization.id
```
- 入院レコードに対して日付ごとの記録

### 9. DailyRecord ↔ VitalRecord / CareLogRecord / StaffNoteRecord（1:N each）

```
DailyRecord.vitals[]      → VitalRecord
DailyRecord.careLogs[]    → CareLogRecord
DailyRecord.staffNotes[]  → StaffNoteRecord
```
- 各日次記録に複数のバイタル/ログ/メモを格納（ネスト構造）

### 10. Accounting ↔ AccountingItem（1:N）

```
AccountingItem は Accounting.items[] として保持
```

### 11. Accounting ↔ PaymentInfo（1:0..1）

```
PaymentInfo は Accounting.payment として保持
```
- 支払い完了後にセットされる

### 12. MasterItem ↔ InventoryItem（N:0..1）

```
MasterItem.inventoryId = InventoryItem.id
```
- 薬剤マスタ等が在庫品目と紐付く

### 13. MasterItem ↔ MasterItem（自己参照: 親子）

```
MasterItem.parentId = MasterItem.id
```
- 診断名マスタ → 診断カテゴリマスタの親子関係

### 14. MasterItem ↔ CarePlanItem（マスタ参照）

```
CarePlanItem.masterId = MasterItem.code
```
- ケアプランがマスタ項目を参照（単価・名称の自動入力）

### 15. Hospitalization ↔ MasterItem(cage)

```
Hospitalization.cageId = MasterItem.id (category=cage)
```
- ケージマスタから入院先ケージを参照

### 16. TrimmingRecord ↔ MasterItem(trimming_course)（N:0..1）

```
TrimmingRecord.courseId = MasterItem.id (category=trimming_course)
```
- トリミングコースマスタからコースを参照（料金・名称の自動入力）

### 17. TrimmingRecord ↔ MasterItem(trimming_option)（N:M）

```
TrimmingRecord.optionIds[] = MasterItem.id (category=trimming_option)
```
- トリミングオプションマスタから複数オプションを参照
- バックエンド実装時は中間テーブル（`TrimmingRecordOption`）が必要

### 18. MedicalRecord ↔ MasterItem(diagnosis)（診断名参照）

```
MedicalRecord.diagnosis1Category = MasterItem.id (category=diagnosis_category)
MedicalRecord.diagnosis1Name     = MasterItem.id (category=diagnosis_name)
MedicalRecord.diagnosis2Category = MasterItem.id (category=diagnosis_category)
MedicalRecord.diagnosis2Name     = MasterItem.id (category=diagnosis_name)
```
- カルテに最大2つの診断（カテゴリ+診断名）を紐付け
- 各診断はdiagnosis_categoryマスタとdiagnosis_nameマスタへのFK

---

## 列挙型（Enum）一覧

### グローバル列挙型（`/types/index.ts`）

| 列挙型 | 値 | 用途 |
|--------|-----|------|
| `PetSpecies` | 犬, 猫, 鳥, その他 | ペット種別 |
| `PetStatus` | 生存, 死亡 | ペット状態 |
| `MedicalRecordStatus` | 作成中, 確定済 | カルテステータス |
| `HospitalizationType` | 入院, ホテル | 入院区分 |
| `HospitalizationStatus` | 入院中, 退院済, 予約 | 入院ステータス |
| `CarePlanType` | food, medicine, treatment, instruction, item | ケアプラン種別 |
| `CarePlanStatus` | active, completed, discontinued | ケアプランステータス |
| `CareLogType` | food, excretion, medicine, treatment, other | ケアログ種別 |
| `CareLogStatus` | completed, partial, skipped | ケアログステータス |
| `PlanTiming` | morning, noon, night | 実施タイミング |
| `ReservationStatus` | confirmed, pending, cancelled, checked_in, in_consultation, accounting, completed | 予約ステータス |
| `ReservationType` | 診療, 検診, 検査, 手術, トリミング, ワクチン, 入院, ホテル | 予約種別 |
| `VisitType` | first, revisit | 来院種別 |
| `VisitReason` | checkup, sick, prevention | 来院理由 |
| `TrimmingStatus` | 完了, 予約, 進行中 | トリミングステータス |
| `ExaminationStatus` | 依頼中, 検査中, 完了 | 検査ステータス |
| `MasterItemStatus` | active, inactive | マスタ有効性 |
| `MasterCategory` | examination, vaccine, medicine, staff, insurance, cage, serviceType, consultation, procedure, hospitalization, trimming_course, trimming_option, diagnosis_category, diagnosis_name | マスタカテゴリ（14種） |
| `InventoryCategory` | medicine, consumable, food, other | 在庫カテゴリ |
| `InventoryStatus` | sufficient, low, out_of_stock | 在庫ステータス |
| `SortOrder` | desc, asc | ソート順 |
| `CalendarView` | month, week | カレンダー表示 |
| `TextAlign` | left, center, right | テーブル列配置 |

### Feature固有列挙型

| 列挙型 | Feature | 値 | 用途 |
|--------|---------|-----|------|
| `AccountingStatus` | accounting | waiting, completed, cancelled, pending | 会計ステータス |
| `PaymentMethod` | accounting | cash, credit_card, electronic_money | 支払方法 |
| `ItemCategory` | accounting | examination, test, procedure, surgery, medicine, food, goods, other | 会計品目カテゴリ |
| `TaxRate` | accounting | 0.1, 0.08 | 税率 |
| `ItemSource` | accounting | medical_record, manual | 品目ソース |
| `InsuranceRatio` | accounting | 0.5, 0.7, 0.9, 1.0 | 保険負担割合（会計側） |
| `DocumentType` | accounting | receipt, statement | 書類種別 |
| `MembershipType` | owners | 非会員, 会員, 退亡者, 他診/準 | 会員種別 |
| `PetGender` | owners | 雄, 雌, 不明 | ペット性別 |
| `AcquisitionType` | owners | 購入, 譲渡, 保護, その他 | 入手経路 |
| `DangerLevel` | owners | 低, 中, 高 | 危険度 |
| `InsuranceCompany` | owners | アニコム, アイペット, ... | 保険会社 |
| `PetInsuranceRatio` | owners | 50%, 70%, 90%, 100%, その他 | ペット保険負担割合（飼主側） |
| `TreatmentStatus` | medical-records | 未完了, 完了, - | 治療ステータス |
| `NextScheduleType` | medical-records | 3weeks, 4weeks, 1year, other | 次回接種間隔 |
| `ExaminationResultStatus` | medical-records | normal, high, low | 検査結果状態 |

---

## データフロー概要

```
                           ┌────────────┐
                           │  MasterItem │
                           │  (14カテゴリ)│
                           └──────┬─────┘
                                  │ 参照
          ┌───────────────────────┼───────────────────────┐
          │                       │                       │
          ▼                       ▼                       ▼
    ┌───────────┐          ┌───────────┐          ┌───────────┐
    │TreatmentItem│         │CarePlanItem│         │InventoryItem│
    │(カルテ治療) │         │(入院ケア)  │         │(在庫管理)  │
    └──────┬────┘          └──────┬────┘          └───────────┘
           │                      │                    ▲
           │                      │                    │ consumeStock
           ▼                      ▼                    │
    ┌──────────────┐       ┌──────────────┐     ┌─────────────┐
    │ MedicalRecord │       │Hospitalization│     │  カルテ保存時  │
    │  (電子カルテ) │       │ (入院管理)    │     │  在庫消費連動  │
    └───────┬──────┘       └──────┬───────┘     └─────────────┘
            │                     │
            │  1:0..1             │ 1:N
            ▼                     ▼
    ┌──────────────┐       ┌──────────────┐
    │  Accounting   │       │  DailyRecord  │
    │   (会計)      │       │ (日次記録)    │
    └──────────────┘       └──────────────┘

    MasterItem(trimming_course/option) ←── TrimmingRecord（コース・オプション連携）
    MasterItem(diagnosis_category/name) ←── MedicalRecord（診断名参照）

    ┌──────────────┐       ┌──────────────┐
    │    Owner      │ 1:N  │     Pet       │
    │  (飼主)       │──────│   (ペット)    │
    └──────────────┘       └──────┬───────┘
                                  │ 1:N
                    ┌─────────────┼─────────────┐
                    │             │             │
                    ▼             ▼             ▼
             MedicalRecord  Hospitalization  TrimmingRecord
             ExaminationRecord  VaccinationRecord  ReservationAppointment
```

---

## 備考

1. **非正規化フィールド**: `ownerName`, `petName`, `species` 等が複数エンティティに冗長に保持されている。これはMock Data環境での表示利便性のため。バックエンド実装時はJOINに置き換えるか、意図的にデノーマライズを維持するか設計判断が必要。

2. **ClinicInfo**: シングルトンエンティティ（1レコードのみ）。`features/clinic/api/store.ts` でインメモリ管理。

3. **Appointment vs ReservationAppointment**: `Appointment` はダッシュボードのカンバンカード用のビューモデル。`ReservationAppointment` は予約管理画面の正式な予約データ。将来は同一エンティティに統合する可能性あり。

4. **TreatmentPlan vs TreatmentItem**: `TreatmentPlan` は入院管理の治療プラン。`TreatmentItem` はカルテ内の治療/処置項目。将来は共通化の余地あり。

5. **インメモリストア**: `master`, `clinic`, `hospitalization` は `api/store.ts` でミュータブルなインメモリストアを保持（CRUD操作のシミュレーション）。その他のfeatureは `api/mockData.ts` の静的データを参照。

6. **Pet のインターフェース分離**: 現在 `Pet`（`types/index.ts`）は一覧/検索用の軽量インターフェース。`PetFormData`/`PetInfo`（`features/owners/types`）がペット編集時の完全フィールドセットを持つ。ERD上は統合した完全なエンティティとして記載。バックエンド実装時は単一テーブルとして実装し、APIレスポンスのプロジェクション（SELECT列の選択）で軽量版を提供する想定。

7. **TrimmingRecord のインターフェース分離**: `TrimmingRecord`（`types/index.ts`）は一覧表示用。`TrimmingFormData`（`features/trimming/types`）がフォーム入力用の完全フィールドセットを持つ。ERD上は統合して記載。

8. **実装状況**: GORM Model は全22テーブル実装済。マイグレーションは `backend/migrations/001_init.sql` から自動実行。詳細は [MIGRATION.md](./MIGRATION.md) を参照。
