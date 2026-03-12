# 画面仕様書 バリデーション・データ型補完ドキュメント

**更新日**: 2026-03-11
**バージョン**: 2.0（完全版）

本ドキュメントは、SCREENS.mdとSCREENS_MASTER.mdで記載されている画面仕様に対して、バリデーションルール、エラーメッセージ、データ型定義を補完します。

> **関連ドキュメント**:
> - [SCREENS.md](./SCREENS.md): メイン機能の画面仕様
> - [SCREENS_MASTER.md](./SCREENS_MASTER.md): マスタ管理の画面仕様

---

## 目次

- [1. ダッシュボード](#1-ダッシュボード)
- [2. 予約管理](#2-予約管理)
- [3. 飼主・ペット管理](#3-飼主ペット管理)
- [4. 電子カルテ](#4-電子カルテ)
- [5. 入院管理](#5-入院管理)
- [6. トリミング](#6-トリミング)
- [7. 検査管理](#7-検査管理)
- [8. 会計](#8-会計)
- [9. 予防接種](#9-予防接種)
- [10. 定期健診](#10-定期健診)
- [12. 在庫管理](#12-在庫管理)
- [13. シフト管理](#13-シフト管理)
- [14. 認証](#14-認証)
- [共通ユーティリティ関数](#共通ユーティリティ関数)

---

## 1. ダッシュボード

### 1.1 ダッシュボード（カンバンボード） - データ型定義

**データ型:**

```typescript
// ダッシュボード予約データ型
interface Appointment {
  id: string;
  petId: string; // ペットID
  ownerId: string; // 飼主ID
  ownerName: string; // 飼主名
  petName: string; // ペット名
  petType: PetSpecies; // ペット種
  time: string; // 予約時刻 (HH:MM)
  serviceType: string; // 診療区分（マスタID）
  serviceTypeName: string; // 診療区分名
  doctor: string; // 担当医（マスタID）
  doctorName: string; // 担当医名
  visitType: 'first' | 'revisit'; // 初診/再診
  status: AppointmentStatus; // ステータス
  isNominated: boolean; // 指名フラグ
  nextAppointment?: 'scheduled' | 'unconfirmed'; // 次回予約状態
  notes?: string; // メモ
  createdAt: string; // 作成日時 (ISO 8601)
  updatedAt: string; // 更新日時 (ISO 8601)
}

type AppointmentStatus =
  | 'scheduled' // 受付予約
  | 'checked_in' // 受付済
  | 'in_consultation' // 診療中
  | 'waiting_payment' // 会計待ち
  | 'completed'; // 会計済

// カンバンカラムデータ型
interface ColumnData {
  id: AppointmentStatus;
  title: DashboardColumnTitle;
  items: Appointment[];
}

type DashboardColumnTitle = '受付予約' | '受付済' | '診療中' | '会計待ち' | '会計済';

// 週次統計データ型
interface WeeklyChartPoint {
  date: string; // YYYY-MM-DD
  count: number; // 件数
}

interface WeeklyStatsResult {
  scheduled: WeeklyChartPoint[]; // 受付予約（7日分）
  checked_in: WeeklyChartPoint[]; // 受付済（7日分）
  in_consultation: WeeklyChartPoint[]; // 診療中（7日分）
  waiting_payment: WeeklyChartPoint[]; // 会計待ち（7日分）
  completed: WeeklyChartPoint[]; // 会計済（7日分）
}
```

**ステータス遷移ルール:**
```typescript
const ALLOWED_STATUS_TRANSITIONS: Record<AppointmentStatus, AppointmentStatus[]> = {
  scheduled: ['checked_in', 'completed'], // 受付予約 → 受付済 or 会計済（キャンセル扱い）
  checked_in: ['in_consultation', 'completed'], // 受付済 → 診療中 or 会計済
  in_consultation: ['waiting_payment', 'completed'], // 診療中 → 会計待ち or 会計済
  waiting_payment: ['completed'], // 会計待ち → 会計済
  completed: [] // 会計済（終了状態、遷移なし）
};

// ステータス遷移可能か検証
const canTransitionTo = (
  currentStatus: AppointmentStatus,
  targetStatus: AppointmentStatus
): boolean => {
  return ALLOWED_STATUS_TRANSITIONS[currentStatus].includes(targetStatus);
};
```

**DnD操作エラーメッセージ:**
- 不正な遷移: 「このステータスへの移動はできません」
- 精算未確認で会計へ移動: 「精算が確認されていません。先にカルテまたは会計登録を行ってください」

---

## 2. 予約管理

### 2.1 予約カレンダー - バリデーション詳細

#### ReservationFormFields フォーム項目

| フィールド | 入力部品 | バリデーション | 必須 | エラーメッセージ |
|-----------|---------|---------------|------|-----------------|
| 日付 | Popover + Calendar | 有効な日付 | ✅ | 「日付は必須です」 |
| 時間帯（開始） | Select（30分刻み） | HH:MM形式 | ✅ | 「開始時間は必須です」 |
| 時間帯（終了） | Select（30分刻み） | 開始時刻より後 | ✅ | 「終了時間は必須です」<br>「終了時間は開始時間より後にしてください」 |
| 予約区分 | Select（マスタ連動） | マスタから選択 | ✅ | 「予約区分を選択してください」 |
| 初診/再診 | RadioGroup | first or revisit | ✅ | 「初診/再診を選択してください」 |
| 担当者 | Select（マスタ連動） | マスタから選択（active） | ✅ | 「担当者を選択してください」 |
| メモ | Textarea | 500文字以内 | - | 「メモは500文字以内で入力してください」 |

#### ペット検索・選択（Step 1）

| フィールド | 入力部品 | バリデーション | 必須 | エラーメッセージ |
|-----------|---------|---------------|------|-----------------|
| 検索テキスト | Input（PatientSearch） | 50文字以内 | - | - |
| ペット選択 | PatientSelectionTable | 選択必須 | ✅ | 「ペットを選択してください」 |

#### カスタムバリデーション

**予約重複チェック:**
```typescript
const checkReservationOverlap = (
  petId: string,
  date: string,
  startTime: string,
  endTime: string,
  excludeId?: string
): boolean => {
  // 同じペット、同じ日付で時間帯が重複する予約があるかチェック
  const existingReservations = getReservationsByPetAndDate(petId, date);
  return existingReservations.some(res => {
    if (excludeId && res.id === excludeId) return false;
    return !(endTime <= res.startTime || startTime >= res.endTime);
  });
};
```

**重複エラーメッセージ:**
- 「この時間帯に既に予約が入っています」（toast.warning）

**時間帯妥当性チェック:**
```typescript
const validateTimeRange = (startTime: string, endTime: string): boolean => {
  const start = parseTime(startTime);
  const end = parseTime(endTime);
  return end > start;
};
```

**データ型定義:**

```typescript
interface ReservationAppointment {
  id: string;
  petId: string; // ペットID
  ownerId: string; // 飼主ID
  ownerName: string; // 飼主名
  petName: string; // ペット名
  petType: PetSpecies; // ペット種
  date: string; // 予約日 (YYYY-MM-DD)
  startTime: string; // 開始時刻 (HH:MM)
  endTime: string; // 終了時刻 (HH:MM)
  serviceTypeId: string; // 予約区分ID
  serviceTypeName: string; // 予約区分名
  visitType: 'first' | 'revisit'; // 初診/再診
  doctorId: string; // 担当医ID
  doctorName: string; // 担当医名
  isNominated: boolean; // 指名フラグ
  status: ReservationStatus; // ステータス
  notes?: string; // メモ
  createdAt: string; // 作成日時 (ISO 8601)
  updatedAt: string; // 更新日時 (ISO 8601)
}

type ReservationStatus =
  | 'confirmed' // 予約確定
  | 'checked_in' // 受付済
  | 'in_consultation' // 診療中
  | 'waiting_payment' // 会計待ち
  | 'completed' // 完了
  | 'cancelled'; // キャンセル

interface ReservationFormData {
  petId: string;
  date: string;
  startTime: string;
  endTime: string;
  serviceTypeId: string;
  visitType: 'first' | 'revisit';
  doctorId: string;
  notes?: string;
}

type CalendarView = 'month' | 'week';
```

---

## 3. 飼主・ペット管理

### 3.2 飼主登録/編集 - バリデーション詳細

#### 飼主情報フォーム

| フィールド | 入力部品 | バリデーション | 必須 | エラーメッセージ |
|-----------|---------|---------------|------|-----------------|
| 飼主名 | Input | 50文字以内 | ✅ | 「飼主名は必須です」<br>「飼主名は50文字以内で入力してください」 |
| 飼主名（カナ） | Input | 50文字以内、全角カタカナ | ✅ | 「飼主名（カナ）は必須です」<br>「飼主名（カナ）は全角カタカナで入力してください」 |
| 会社名 | Input | 100文字以内 | - | 「会社名は100文字以内で入力してください」 |
| 郵便番号 | Input | `^\d{3}-\d{4}$` | - | 「郵便番号は000-0000の形式で入力してください」 |
| 住所1 | Input | 100文字以内 | - | 「住所1は100文字以内で入力してください」 |
| 住所2 | Input | 100文字以内 | - | 「住所2は100文字以内で入力してください」 |
| 自宅住所1 | Input | 100文字以内 | - | 「自宅住所1は100文字以内で入力してください」 |
| 自宅住所2 | Input | 100文字以内 | - | 「自宅住所2は100文字以内で入力してください」 |
| 飼主生年月日 | NotionDatePicker | 過去日のみ | - | 「生年月日は過去の日付を選択してください」 |
| 電話番号 | Input | 日本形式 | ✅ | 「電話番号は必須です」<br>「有効な電話番号を入力してください（例: 03-1234-5678）」 |
| 電話番号（自宅） | Input | 日本形式 | - | 「有効な電話番号を入力してください」 |
| 会社電話番号 | Input | 日本形式 | - | 「有効な電話番号を入力してください」 |
| メールアドレス | Input (type=email) | RFC 5322準拠 | - | 「有効なメールアドレスを入力してください」 |
| 備考・特記事項 | Textarea | 1000文字以内 | - | 「備考は1000文字以内で入力してください」 |
| 危険人物 | Switch | boolean | - | - |
| 値引率 (%) | Input (type=number) | 0〜100 | - | 「値引率は0〜100の範囲で入力してください」 |
| 会員区分 | Button群 | 選択必須 | ✅ | 「会員区分を選択してください」 |

**全角カタカナ検証:**
```typescript
const isFullWidthKatakana = (value: string): boolean => {
  return /^[ァ-ヶー\s]+$/.test(value);
};
```

**電話番号自動フォーマット:**
```typescript
const formatPhoneOnChange = (value: string) => {
  const digits = value.replace(/\D/g, '');
  if (digits.length === 10) {
    // 固定電話: 03-1234-5678
    return `${digits.slice(0, 2)}-${digits.slice(2, 6)}-${digits.slice(6)}`;
  }
  if (digits.length === 11) {
    // 携帯電話: 090-1234-5678
    return `${digits.slice(0, 3)}-${digits.slice(3, 7)}-${digits.slice(7)}`;
  }
  return value;
};
```

#### ペット編集モーダル - バリデーション詳細

| フィールド | 入力部品 | バリデーション | 必須 | エラーメッセージ |
|-----------|---------|---------------|------|-----------------|
| ペット名 | Input | 50文字以内 | ✅ | 「ペット名は必須です」<br>「ペット名は50文字以内で入力してください」 |
| ペット名カナ | Input | 50文字以内、全角カタカナ | - | 「ペット名カナは全角カタカナで入力してください」 |
| 種 | Select | - | ✅ | 「種を選択してください」 |
| 性別 | Select | - | ✅ | 「性別を選択してください」 |
| 生年月日 | NotionDatePicker | 過去日のみ | ✅ | 「生年月日は必須です」<br>「生年月日は過去の日付を選択してください」 |
| 品種 | Input | 50文字以内 | - | 「品種は50文字以内で入力してください」 |
| 毛色 | Input | 50文字以内 | - | 「毛色は50文字以内で入力してください」 |
| 避妊去勢日 | NotionDatePicker | 過去日、生年月日以降 | - | 「避妊去勢日は生年月日以降の日付を選択してください」 |
| 入手種別 | Select | - | - | - |
| 危険度 | Select | - | - | - |
| フード | Input | 100文字以内 | - | 「フードは100文字以内で入力してください」 |
| 保険名 | Select | - | - | - |
| 保険詳細（負担割合） | Select | - | - | - |
| 備考・特記事項 | Textarea | 1000文字以内 | - | 「備考は1000文字以内で入力してください」 |

**データ型定義:**

```typescript
// 飼主データ型
interface Owner {
  id: string;
  name: string; // 飼主名
  nameKana: string; // 飼主名カナ
  companyName?: string; // 会社名
  postalCode?: string; // 郵便番号
  address1?: string; // 住所1
  address2?: string; // 住所2
  homeAddress1?: string; // 自宅住所1
  homeAddress2?: string; // 自宅住所2
  birthDate?: string; // 生年月日 (YYYY-MM-DD)
  phone: string; // 電話番号（必須）
  homePhone?: string; // 電話番号（自宅）
  companyPhone?: string; // 会社電話番号
  email?: string; // メールアドレス
  notes?: string; // 備考・特記事項
  isDangerous: boolean; // 危険人物フラグ
  discountRate?: number; // 値引率 (0-100)
  membershipType: MembershipType; // 会員区分
  createdAt: string; // 作成日時 (ISO 8601)
  updatedAt: string; // 更新日時 (ISO 8601)
}

type MembershipType = 'non_member' | 'member' | 'deceased' | 'other_clinic' | 'provisional';

// ペットデータ型
interface Pet {
  id: string;
  ownerId: string; // 飼主ID
  name: string; // ペット名（必須）
  nameKana?: string; // ペット名カナ
  species: PetSpecies; // 種（必須）
  gender: PetGender; // 性別（必須）
  birthDate: string; // 生年月日（必須、YYYY-MM-DD）
  breed?: string; // 品種
  color?: string; // 毛色
  neuteredDate?: string; // 避妊去勢日 (YYYY-MM-DD)
  acquisitionType?: AcquisitionType; // 入手種別
  dangerLevel?: DangerLevel; // 危険度
  food?: string; // フード
  insuranceCompany?: InsuranceCompany; // 保険名
  insuranceRatio?: PetInsuranceRatio; // 保険詳細（負担割合）
  notes?: string; // 備考・特記事項
  status: PetStatus; // ステータス
  createdAt: string; // 作成日時 (ISO 8601)
  updatedAt: string; // 更新日時 (ISO 8601)
}

type PetSpecies = 'dog' | 'cat' | 'other';
type PetGender = 'male' | 'female' | 'unknown';
type AcquisitionType = 'purchase' | 'adoption' | 'rescue' | 'other';
type DangerLevel = 'low' | 'medium' | 'high';
type InsuranceCompany =
  | 'anicom'
  | 'ipet'
  | 'pet_and_family'
  | 'rakuten'
  | 'axa'
  | 'sbi'
  | 'fpc'
  | 'other';
type PetInsuranceRatio = '50' | '70' | '90' | '100' | 'other';
type PetStatus = 'active' | 'deceased';
```

---

## 4. 電子カルテ

### 4.3 カルテ入力/編集 - 全タブバリデーション詳細

#### タブ1: 基本情報

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 診療日 | 有効な日付 | ✅ | 「診療日は必須です」 |
| 診療区分 | マスタから選択 | ✅ | 「診療区分を選択してください」 |
| 担当医 | マスタから選択 | ✅ | 「担当医を選択してください」 |
| 初診/再診 | 選択必須 | ✅ | - |
| 体重 (kg) | 0以上、小数点2桁 | - | 「体重は0以上の数値を入力してください」 |
| 体温 (℃) | 0以上、小数点1桁 | - | 「体温は0以上の数値を入力してください」 |
| 心拍数 (/min) | 0以上、整数 | - | 「心拍数は0以上の整数を入力してください」 |
| 呼吸数 (/min) | 0以上、整数 | - | 「呼吸数は0以上の整数を入力してください」 |
| 主訴 | 5000文字以内 | - | 「主訴は5000文字以内で入力してください」 |

#### タブ2: 診察

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 診察所見 | 5000文字以内（Markdown） | - | 「診察所見は5000文字以内で入力してください」 |
| 診察項目 | マスタから選択（複数可） | - | - |
| 診察項目単価 | 0以上 | - | 「単価は0以上の数値を入力してください」 |

#### タブ3: 診断

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 診断カテゴリ | マスタから選択 | - | - |
| 診断名 | マスタから選択（複数可） | - | - |
| 診断詳細 | 5000文字以内（Markdown） | - | 「診断詳細は5000文字以内で入力してください」 |

#### タブ4: 検査

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 検査種別 | マスタから選択 | ✅ | 「検査種別を選択してください」 |
| 検査日 | 有効な日付 | ✅ | 「検査日は必須です」 |
| 検査項目 | マスタから選択 | - | - |
| 検査値 | 各項目の形式に応じて | - | 「有効な値を入力してください」 |
| 正常範囲 | 読み取り専用 | - | - |
| 結果概要 | 5000文字以内（Markdown） | - | 「結果概要は5000文字以内で入力してください」 |

#### タブ5: 処置

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 処置項目 | マスタから選択（複数可） | - | - |
| 処置項目単価 | 0以上 | - | 「単価は0以上の数値を入力してください」 |
| 処置詳細 | 5000文字以内（Markdown） | - | 「処置詳細は5000文字以内で入力してください」 |

#### タブ6: 処方

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 薬剤 | マスタから選択 | ✅ | 「薬剤を選択してください」 |
| 用量 | 0以上、小数点2桁 | ✅ | 「用量は0以上の数値を入力してください」 |
| 単位 | マスタから選択 | ✅ | 「単位を選択してください」 |
| 回数 | 1以上、整数 | ✅ | 「回数は1以上の整数を入力してください」 |
| 日数 | 1以上、整数 | ✅ | 「日数は1以上の整数を入力してください」 |
| 総量 | 自動計算（読み取り専用） | - | - |
| 用法 | 200文字以内 | - | 「用法は200文字以内で入力してください」 |

**処方薬総量自動計算:**
```typescript
const calculateTotalAmount = (dose: number, times: number, days: number): number => {
  return dose * times * days;
};
```

#### タブ7: 予防接種

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 予防接種名 | マスタから選択 | ✅ | 「予防接種名を選択してください」 |
| 接種日 | 有効な日付 | ✅ | 「接種日は必須です」 |
| 次回接種予定日 | 接種日以降の日付 | - | 「次回接種予定日は接種日以降の日付を選択してください」 |
| ロット番号 | 50文字以内 | - | 「ロット番号は50文字以内で入力してください」 |
| 製造元 | 100文字以内 | - | 「製造元は100文字以内で入力してください」 |
| 接種部位 | 50文字以内 | - | 「接種部位は50文字以内で入力してください」 |
| 副反応 | 500文字以内 | - | 「副反応は500文字以内で入力してください」 |

#### タブ8: 定期健診

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 健診種別 | マスタから選択 | ✅ | 「健診種別を選択してください」 |
| 健診日 | 有効な日付 | ✅ | 「健診日は必須です」 |
| 次回健診予定日 | 健診日以降の日付 | - | 「次回健診予定日は健診日以降の日付を選択してください」 |
| 健診項目 | マスタから選択（複数可） | - | - |
| 結果 | 5000文字以内（Markdown） | - | 「結果は5000文字以内で入力してください」 |

#### タブ9: 会計連携

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 明細項目 | 自動集計（読み取り専用） | - | - |
| 小計 | 自動計算（読み取り専用） | - | - |
| 消費税 | 自動計算（読み取り専用） | - | - |
| 合計 | 自動計算（読み取り専用） | - | - |

**データ型定義:**

```typescript
// 電子カルテデータ型
interface MedicalRecord {
  id: string;
  petId: string; // ペットID
  ownerId: string; // 飼主ID
  date: string; // 診療日 (YYYY-MM-DD)
  serviceTypeId: string; // 診療区分ID
  doctorId: string; // 担当医ID
  visitType: 'first' | 'revisit'; // 初診/再診

  // バイタルサイン
  weight?: number; // 体重 (kg)
  temperature?: number; // 体温 (℃)
  heartRate?: number; // 心拍数 (/min)
  respiratoryRate?: number; // 呼吸数 (/min)

  // 主訴・診察
  chiefComplaint?: string; // 主訴 (Markdown)
  examinationNotes?: string; // 診察所見 (Markdown)
  consultationItems: ConsultationItem[]; // 診察項目

  // 診断
  diagnosisCategoryId?: string; // 診断カテゴリID
  diagnosisNames: string[]; // 診断名ID配列
  diagnosisDetails?: string; // 診断詳細 (Markdown)

  // 検査
  examinations: Examination[]; // 検査記録

  // 処置
  procedures: Procedure[]; // 処置記録

  // 処方
  prescriptions: Prescription[]; // 処方薬

  // 予防接種
  vaccinations: Vaccination[]; // 予防接種記録

  // 定期健診
  checkups: Checkup[]; // 定期健診記録

  // 会計連携
  accountingItems: AccountingItem[]; // 会計明細

  status: 'draft' | 'completed'; // ステータス
  createdAt: string; // 作成日時 (ISO 8601)
  updatedAt: string; // 更新日時 (ISO 8601)
  createdBy: string; // 作成者ID
  updatedBy: string; // 更新者ID
}

interface ConsultationItem {
  id: string;
  itemId: string; // 診察項目マスタID
  itemName: string; // 診察項目名
  price: number; // 単価 (税込)
}

interface Examination {
  id: string;
  examinationTypeId: string; // 検査種別マスタID
  examinationTypeName: string; // 検査種別名
  date: string; // 検査日 (YYYY-MM-DD)
  items: ExaminationItem[]; // 検査項目
  summary?: string; // 結果概要 (Markdown)
  status: 'pending' | 'in_progress' | 'completed'; // ステータス
}

interface ExaminationItem {
  id: string;
  itemName: string; // 検査項目名
  value: string; // 検査値
  unit: string; // 単位
  normalRange: string; // 正常範囲
  isAbnormal: boolean; // 異常値フラグ
}

interface Procedure {
  id: string;
  procedureId: string; // 処置項目マスタID
  procedureName: string; // 処置項目名
  price: number; // 単価 (税込)
  details?: string; // 処置詳細 (Markdown)
}

interface Prescription {
  id: string;
  medicineId: string; // 薬剤マスタID
  medicineName: string; // 薬剤名
  dose: number; // 用量
  unit: string; // 単位
  frequency: number; // 回数
  days: number; // 日数
  totalAmount: number; // 総量（自動計算）
  instructions?: string; // 用法
  price: number; // 単価 (税込)
}

interface Vaccination {
  id: string;
  vaccineId: string; // 予防接種マスタID
  vaccineName: string; // 予防接種名
  date: string; // 接種日 (YYYY-MM-DD)
  nextDate?: string; // 次回接種予定日 (YYYY-MM-DD)
  lotNumber?: string; // ロット番号
  manufacturer?: string; // 製造元
  injectionSite?: string; // 接種部位
  adverseReaction?: string; // 副反応
  price: number; // 単価 (税込)
}

interface Checkup {
  id: string;
  checkupTypeId: string; // 健診種別マスタID
  checkupTypeName: string; // 健診種別名
  date: string; // 健診日 (YYYY-MM-DD)
  nextDate?: string; // 次回健診予定日 (YYYY-MM-DD)
  items: string[]; // 健診項目ID配列
  result?: string; // 結果 (Markdown)
  price: number; // 単価 (税込)
}

interface AccountingItem {
  id: string;
  category: string; // カテゴリ (診察/検査/処置/処方/予防接種/定期健診)
  name: string; // 項目名
  quantity: number; // 数量
  unitPrice: number; // 単価 (税込)
  totalPrice: number; // 金額 (税込)
  source: 'medical_record' | 'manual'; // ソース
}
```

---

## 5. 入院管理

### 5.3 入院登録/編集 - バリデーション詳細

#### 基本情報

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 入院タイプ | 選択必須（入院/ホテル） | ✅ | 「入院タイプを選択してください」 |
| 期間（開始日） | 有効な日付 | ✅ | 「開始日は必須です」 |
| 期間（終了日） | 開始日以降の日付 | ✅ | 「終了日は必須です」<br>「終了日は開始日以降の日付を選択してください」 |
| ケージ・個室 | マスタから選択 | ✅ | 「ケージを選択してください」 |
| 担当医 | マスタから選択 | ✅ | 「担当医を選択してください」 |
| 飼主からのリクエスト | 1000文字以内 | - | 「リクエストは1000文字以内で入力してください」 |
| スタッフへの連絡事項 | 1000文字以内 | - | 「連絡事項は1000文字以内で入力してください」 |

#### ケアプラン

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 種類 | 選択必須 | ✅ | 「種類を選択してください」 |
| 名称 | 200文字以内 | ✅ | 「名称は必須です」<br>「名称は200文字以内で入力してください」 |
| 詳細・指示量 | 200文字以内 | - | 「詳細・指示量は200文字以内で入力してください」 |
| タイミング | 最低1つ選択 | ✅ | 「タイミングを選択してください」 |
| メモ・特記事項 | 500文字以内 | - | 「メモは500文字以内で入力してください」 |
| ステータス | 選択必須 | ✅ | - |

---

## 6. トリミング

### 6.3 トリミング登録/編集 - バリデーション詳細

#### 基本情報

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| トリミング日 | 有効な日付 | ✅ | 「トリミング日は必須です」 |
| 担当者 | マスタから選択 | ✅ | 「担当者を選択してください」 |
| ペット | マスタから選択 | ✅ | 「ペットを選択してください」 |
| トリミング種類 | マスタから選択 | ✅ | 「トリミング種類を選択してください」 |
| トリミング部位 | 50文字以内 | - | 「トリミング部位は50文字以内で入力してください」 |
| トリミング詳細 | 500文字以内 | - | 「トリミング詳細は500文字以内で入力してください」 |
| トリミング料金 | 0以上 | - | 「トリミング料金は0以上の数値を入力してください」 |

**データ型定義:**

```typescript
interface Trimming {
  id: string;
  petId: string; // ペットID
  ownerId: string; // 飼主ID
  date: string; // トリミング日 (YYYY-MM-DD)
  staffId: string; // 担当者ID
  staffName: string; // 担当者名
  trimmingTypeId: string; // トリミング種類ID
  trimmingTypeName: string; // トリミング種類名
  trimmingParts?: string; // トリミング部位
  trimmingDetails?: string; // トリミング詳細
  price: number; // トリミング料金
  status: 'completed' | 'cancelled'; // ステータス
  createdAt: string; // 作成日時 (ISO 8601)
  updatedAt: string; // 更新日時 (ISO 8601)
}
```

---

## 7. 検査管理

### 7.3 検査登録/編集 - バリデーション詳細

#### 基本情報

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 検査日 | 有効な日付 | ✅ | 「検査日は必須です」 |
| 担当者 | マスタから選択 | ✅ | 「担当者を選択してください」 |
| ペット | マスタから選択 | ✅ | 「ペットを選択してください」 |
| 検査種別 | マスタから選択 | ✅ | 「検査種別を選択してください」 |
| 検査項目 | マスタから選択（複数可） | - | - |
| 検査値 | 各項目の形式に応じて | - | 「有効な値を入力してください」 |
| 正常範囲 | 読み取り専用 | - | - |
| 結果概要 | 5000文字以内（Markdown） | - | 「結果概要は5000文字以内で入力してください」 |

**データ型定義:**

```typescript
interface Examination {
  id: string;
  petId: string; // ペットID
  ownerId: string; // 飼主ID
  date: string; // 検査日 (YYYY-MM-DD)
  staffId: string; // 担当者ID
  staffName: string; // 担当者名
  examinationTypeId: string; // 検査種別ID
  examinationTypeName: string; // 検査種別名
  items: ExaminationItem[]; // 検査項目
  summary?: string; // 結果概要 (Markdown)
  status: 'pending' | 'in_progress' | 'completed'; // ステータス
  createdAt: string; // 作成日時 (ISO 8601)
  updatedAt: string; // 更新日時 (ISO 8601)
}

interface ExaminationItem {
  id: string;
  itemName: string; // 検査項目名
  value: string; // 検査値
  unit: string; // 単位
  normalRange: string; // 正常範囲
  isAbnormal: boolean; // 異常値フラグ
}
```

---

## 8. 会計

### 8.3 会計精算 - バリデーション詳細

#### 手動明細追加

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 区分 | 選択必須 | ✅ | 「区分を選択してください」 |
| 品目名 | 100文字以内 | ✅ | 「品目名は必須です」<br>「品目名は100文字以内で入力してください」 |
| 単価（税込） | 0以上、整数 | ✅ | 「単価は必須です」<br>「単価は0以上の数値を入力してください」 |

#### ペット保険

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 保険ON/OFF | boolean | - | - |
| 負担割合 | 選択必須（保険ON時） | ✅ (保険ON時) | 「負担割合を選択してください」 |
| 保険負担額 | 自動計算（読み取り専用） | - | - |

#### 決済情報

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 請求金額 | 自動計算（読み取り専用） | - | - |
| 支払方法 | 選択必須 | ✅ | 「支払方法を選択してください」 |
| お預かり金額 | 請求金額以上 | ✅ | 「お預かり金額は必須です」<br>「お預かり金額が不足しています（不足額: ¥{amount}）」 |
| お釣り | 自動計算（読み取り専用） | - | - |

**保険負担額自動計算:**
```typescript
const calculateInsuranceCoverage = (
  subtotal: number,
  insuranceRatio: number
): number => {
  return Math.floor(subtotal * (insuranceRatio / 100));
};
```

**お釣り自動計算:**
```typescript
const calculateChange = (
  receivedAmount: number,
  totalAmount: number
): number => {
  return receivedAmount - totalAmount;
};
```

**データ型定義:**

```typescript
interface Accounting {
  id: string;
  petId: string; // ペットID
  ownerId: string; // 飼主ID
  medicalRecordId?: string; // カルテID（カルテ連携時）
  hospitalizationId?: string; // 入院ID（入院連携時）
  scheduledDate: string; // 会計日 (YYYY-MM-DD HH:mm)
  items: AccountingItem[]; // 明細項目
  payment?: PaymentInfo; // 支払情報
  status: AccountingStatus; // ステータス
  source?: 'medical_record' | 'hospitalization' | 'manual'; // ソース
  createdAt: string; // 作成日時 (ISO 8601)
  updatedAt: string; // 更新日時 (ISO 8601)
}

interface AccountingItem {
  id: string;
  category: ItemCategory; // カテゴリ
  name: string; // 項目名
  unitPrice: number; // 単価 (税込)
  quantity: number; // 数量
  source: ItemSource; // ソース
  isInsuranceCovered: boolean; // 保険適用フラグ
}

interface PaymentInfo {
  method: PaymentMethod; // 支払方法
  receivedAmount: number; // お預かり金額
  changeAmount: number; // お釣り
  insuranceName?: string; // 保険名
  insuranceRatio?: InsuranceRatio; // 負担割合
  insuranceCoverage?: number; // 保険負担額
  patientPayment?: number; // 自己負担額
}

type AccountingStatus = 'waiting' | 'completed' | 'cancelled' | 'pending';
type ItemCategory =
  | 'consultation'
  | 'examination'
  | 'procedure'
  | 'medicine'
  | 'vaccination'
  | 'checkup'
  | 'hospitalization'
  | 'trimming'
  | 'food'
  | 'product'
  | 'other';
type ItemSource = 'medical_record' | 'hospitalization' | 'manual';
type PaymentMethod = 'cash' | 'card' | 'emoney';
type InsuranceRatio = 50 | 70 | 90 | 100;
```

---

## 9. 予防接種

### 9.3 予防接種登録/編集 - バリデーション詳細

#### 基本情報

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 予防接種名 | マスタから選択 | ✅ | 「予防接種名を選択してください」 |
| 接種日 | 有効な日付 | ✅ | 「接種日は必須です」 |
| 次回接種予定日 | 接種日以降の日付 | - | 「次回接種予定日は接種日以降の日付を選択してください」 |
| ロット番号 | 50文字以内 | - | 「ロット番号は50文字以内で入力してください」 |
| 製造元 | 100文字以内 | - | 「製造元は100文字以内で入力してください」 |
| 接種部位 | 50文字以内 | - | 「接種部位は50文字以内で入力してください」 |
| 副反応 | 500文字以内 | - | 「副反応は500文字以内で入力してください」 |

**データ型定義:**

```typescript
interface Vaccination {
  id: string;
  petId: string; // ペットID
  ownerId: string; // 飼主ID
  vaccineId: string; // 予防接種マスタID
  vaccineName: string; // 予防接種名
  date: string; // 接種日 (YYYY-MM-DD)
  nextDate?: string; // 次回接種予定日 (YYYY-MM-DD)
  lotNumber?: string; // ロット番号
  manufacturer?: string; // 製造元
  injectionSite?: string; // 接種部位
  adverseReaction?: string; // 副反応
  price: number; // 単価 (税込)
  status: 'completed' | 'cancelled'; // ステータス
  createdAt: string; // 作成日時 (ISO 8601)
  updatedAt: string; // 更新日時 (ISO 8601)
}
```

---

## 10. 定期健診

### 10.3 定期健診登録/編集 - バリデーション詳細

#### 基本情報

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 健診種別 | マスタから選択 | ✅ | 「健診種別を選択してください」 |
| 健診日 | 有効な日付 | ✅ | 「健診日は必須です」 |
| 次回健診予定日 | 健診日以降の日付 | - | 「次回健診予定日は健診日以降の日付を選択してください」 |
| 健診項目 | マスタから選択（複数可） | - | - |
| 結果 | 5000文字以内（Markdown） | - | 「結果は5000文字以内で入力してください」 |

**データ型定義:**

```typescript
interface Checkup {
  id: string;
  petId: string; // ペットID
  ownerId: string; // 飼主ID
  checkupTypeId: string; // 健診種別マスタID
  checkupTypeName: string; // 健診種別名
  date: string; // 健診日 (YYYY-MM-DD)
  nextDate?: string; // 次回健診予定日 (YYYY-MM-DD)
  items: string[]; // 健診項目ID配列
  result?: string; // 結果 (Markdown)
  price: number; // 単価 (税込)
  status: 'completed' | 'cancelled'; // ステータス
  createdAt: string; // 作成日時 (ISO 8601)
  updatedAt: string; // 更新日時 (ISO 8601)
}
```

---

## 12. 在庫管理

### 12.2 在庫登録/編集 - バリデーション詳細

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| 品目名 | 100文字以内 | ✅ | 「品目名は必須です」<br>「品目名は100文字以内で入力してください」 |
| カテゴリ | マスタから選択 | ✅ | 「カテゴリを選択してください」 |
| SKU/コード | 50文字以内、半角英数字 | - | 「SKU/コードは半角英数字で入力してください」 |
| 現在庫数 | 0以上、整数 | ✅ | 「現在庫数は必須です」<br>「現在庫数は0以上の整数を入力してください」 |
| 単位 | 50文字以内 | ✅ | 「単位は必須です」 |
| 発注点 | 0以上、整数 | - | 「発注点は0以上の整数を入力してください」 |
| 単価（税込） | 0以上 | - | 「単価は0以上の数値を入力してください」 |
| 仕入先 | 100文字以内 | - | 「仕入先は100文字以内で入力してください」 |
| 保管場所 | 100文字以内 | - | 「保管場所は100文字以内で入力してください」 |
| 備考 | 500文字以内 | - | 「備考は500文字以内で入力してください」 |

**在庫不足アラート:**
```typescript
const isLowStock = (currentStock: number, reorderPoint: number): boolean => {
  return currentStock <= reorderPoint;
};
```

**データ型定義:**

```typescript
interface InventoryItem {
  id: string;
  name: string; // 品目名
  category: string; // カテゴリ
  sku?: string; // SKU/コード
  currentStock: number; // 現在庫数
  unit: string; // 単位
  reorderPoint?: number; // 発注点
  unitPrice?: number; // 単価 (税込)
  supplier?: string; // 仕入先
  location?: string; // 保管場所
  notes?: string; // 備考
  status: 'active' | 'inactive'; // ステータス
  createdAt: string; // 作成日時 (ISO 8601)
  updatedAt: string; // 更新日時 (ISO 8601)
}

interface StockTransaction {
  id: string;
  inventoryItemId: string; // 在庫品目ID
  type: 'in' | 'out' | 'adjustment'; // 種別
  quantity: number; // 数量
  beforeStock: number; // 変更前在庫数
  afterStock: number; // 変更後在庫数
  reason?: string; // 理由
  notes?: string; // 備考
  performedBy: string; // 実施者ID
  performedAt: string; // 実施日時 (ISO 8601)
}
```

---

## 13. シフト管理

### 13.1 シフト管理カレンダー - バリデーション詳細

#### シフト登録/編集

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| スタッフ | マスタから選択 | ✅ | 「スタッフを選択してください」 |
| 日付 | 有効な日付 | ✅ | 「日付は必須です」 |
| 開始時刻 | HH:MM形式 | ✅ | 「開始時刻は必須です」<br>「有効な時刻を入力してください」 |
| 終了時刻 | 開始時刻より後 | ✅ | 「終了時刻は必須です」<br>「終了時刻は開始時刻より後にしてください」 |
| シフト種別 | 選択必須 | ✅ | 「シフト種別を選択してください」 |
| 休憩時間（分） | 0以上、整数 | - | 「休憩時間は0以上の整数を入力してください」 |
| メモ | 200文字以内 | - | 「メモは200文字以内で入力してください」 |

**労働時間自動計算:**
```typescript
const calculateWorkingHours = (
  startTime: string,
  endTime: string,
  breakMinutes: number
): number => {
  const start = new Date(`2000-01-01T${startTime}`);
  const end = new Date(`2000-01-01T${endTime}`);
  const diffMinutes = (end.getTime() - start.getTime()) / (1000 * 60);
  return (diffMinutes - breakMinutes) / 60; // 時間単位で返す
};
```

**シフト重複チェック:**
```typescript
const hasOverlap = (
  staffId: string,
  date: string,
  startTime: string,
  endTime: string,
  existingShifts: Shift[]
): boolean => {
  return existingShifts.some(shift =>
    shift.staffId === staffId &&
    shift.date === date &&
    !(endTime <= shift.startTime || startTime >= shift.endTime)
  );
};
```

**データ型定義:**

```typescript
interface Shift {
  id: string;
  staffId: string; // スタッフID
  date: string; // 日付 (YYYY-MM-DD)
  startTime: string; // 開始時刻 (HH:MM)
  endTime: string; // 終了時刻 (HH:MM)
  shiftType: ShiftType; // シフト種別
  breakMinutes: number; // 休憩時間（分）
  workingHours: number; // 労働時間（自動計算、時間単位）
  notes?: string; // メモ
  status: 'scheduled' | 'confirmed' | 'completed' | 'cancelled'; // ステータス
  createdAt: string; // 作成日時 (ISO 8601)
  updatedAt: string; // 更新日時 (ISO 8601)
}

type ShiftType = 'regular' | 'early' | 'late' | 'night' | 'holiday';
```

---

## 14. 認証

### ログイン画面 - バリデーション詳細

| フィールド | バリデーション | 必須 | エラーメッセージ |
|-----------|---------------|------|-----------------|
| メールアドレス | RFC 5322準拠 | ✅ | 「メールアドレスは必須です」<br>「有効なメールアドレスを入力してください」 |
| パスワード | 8文字以上 | ✅ | 「パスワードは必須です」<br>「パスワードは8文字以上で入力してください」 |

**認証エラー:**
- ステータスコード401: 「メールアドレスまたはパスワードが正しくありません」
- ステータスコード403: 「アカウントが無効化されています。管理者にお問い合わせください」
- ステータスコード500: 「ログインに失敗しました。しばらくしてから再度お試しください」

**データ型定義:**

```typescript
interface LoginCredentials {
  email: string;
  password: string;
}

interface AuthUser {
  id: string;
  name: string;
  email: string;
  userType: UserType;
  clinicIds: string[]; // 所属医院ID配列
  permissions: Permission[];
  lastLoginAt?: string; // 最終ログイン日時 (ISO 8601)
}

type UserType = 'system_admin' | 'clinic_admin' | 'staff';

interface Permission {
  resource: string; // リソース名
  actions: ('view' | 'create' | 'edit' | 'delete')[]; // 許可されたアクション
}
```

---

## 共通ユーティリティ関数

### 日付フォーマット

```typescript
// YYYY-MM-DD → YYYY年MM月DD日
export const formatDateJP = (dateStr: string): string => {
  const date = new Date(dateStr);
  return `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日`;
};

// ISO 8601 → YYYY/MM/DD HH:mm
export const formatDateTime = (isoStr: string): string => {
  const date = new Date(isoStr);
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, '0');
  const d = String(date.getDate()).padStart(2, '0');
  const h = String(date.getHours()).padStart(2, '0');
  const min = String(date.getMinutes()).padStart(2, '0');
  return `${y}/${m}/${d} ${h}:${min}`;
};
```

### 通貨フォーマット

```typescript
export const formatCurrency = (amount: number): string => {
  return `¥${amount.toLocaleString('ja-JP')}`;
};
```

### 消費税計算

```typescript
const TAX_RATE = 0.10; // 10%

export const calculateTax = (subtotal: number): number => {
  return Math.floor(subtotal * TAX_RATE);
};

export const calculateTotal = (subtotal: number): number => {
  return subtotal + calculateTax(subtotal);
};
```

---

**このドキュメントは継続的に更新されます。**
