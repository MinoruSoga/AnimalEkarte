# マスタ設定 仕様書

**更新日:** 2026-03-12
**バージョン:** 3.0（SCREENS_MASTER.md準拠）

## 概要
- **画面の目的**: 診療・トリミング・入院・スタッフなど、システム全体で使用するマスタデータを一元管理する
- **URLパターン**: `/settings`（トップ）、`/settings/*`（各カテゴリ）
- **アクセス権限**: ユーザー種別による制限あり
  - システム管理者: 全マスタ編集可能
  - 医院管理者: 所属医院のマスタ編集可能
  - スタッフ: 閲覧のみ

---

## 11.1 マスタ設定トップ

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings` |
| **コンポーネント** | `[R] MasterSettingsIndex` |
| **目的** | マスタカテゴリ一覧をカード形式で表示 |

### 画面構成
- ヘッダー: タイトル「マスタ設定」（Settings アイコン）
- セクション分類（5セクション）
- カードグリッド（`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3`）

### セクション構成

| # | セクション名 | カテゴリキー |
|---|---|---|
| 1 | 基本設定 | `clinic` |
| 2 | 診療関連マスタ | `serviceType`, `consultation`, `examination`, `procedure`, `vaccine`, `medicine`, `diagnosis_category`, `diagnosis_name` |
| 3 | 入院・ケージ管理 | `hospitalization`, `cage` |
| 4 | トリミング関連 | `trimming_course`, `trimming_option` |
| 5 | スタッフ・保険 | `staff`, `insurance` |

### カテゴリカード（`CategoryCard`）表示項目

| 項目 | 表示内容 |
|---|---|
| アイコン | `CATEGORY_CONFIG[key].IconComponent`（ホバー時 bg-[#37352F] text-white） |
| カテゴリ名 | `cfg.label` |
| 説明 | `cfg.description`（2行制限） |
| 件数 | `{count}件登録済`（clinicは非表示） |
| 矢印 | ChevronRight |

---

## 11.2 病院情報設定

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/clinic` |
| **コンポーネント** | `[R] ClinicSettings` |
| **目的** | 病院の基本情報を管理する |

### フォーム項目

| フィールド | 入力部品 | バリデーション | 必須 |
|---|---|---|---|
| 病院名 | `Input` | 50文字以内 | ✅ |
| 支店名 | `Input` | 50文字以内 | - |
| 郵便番号 | `Input` | `^\d{3}-\d{4}$` | - |
| 住所 | `Input` | 200文字以内 | - |
| 電話番号 | `Input` | 日本形式 | ✅ |
| FAX番号 | `Input` | 日本形式 | - |
| 登録番号 | `Input` | 100文字以内 | - |
| 院長名 | `Input` | 50文字以内 | - |
| メールアドレス | `Input`（type=email） | RFC 5322準拠 | - |
| WebサイトURL | `Input` | URL形式 | - |

### データ型

```typescript
interface ClinicInfo {
  name: string;
  branchName?: string;
  postalCode?: string;
  address?: string;
  phoneNumber: string;
  faxNumber?: string;
  registrationNumber?: string;
  directorName?: string;
  email?: string;
  website?: string;
  logoUrl?: string;
}
```

---

## 11.3 診療項目マスタ（統合ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/treatment-items` |
| **コンポーネント** | `[R] TreatmentItemsSettings` |
| **目的** | 診察・検査・処置・予防接種・定期健診の5カテゴリを1ページで管理 |

### タブ構成

| タブラベル | カテゴリキー | showPrice | showCode | showParentItem |
|---|---|---|---|---|
| 診察 | `consultation` | true | false | true |
| 検査 | `examination` | true | false | true |
| 処置 | `procedure` | true | false | true |
| 予防接種 | `vaccine` | true | false | true |
| 定期健診 | `checkup` | true | false | true |

---

## 11.4 診断マスタ（統合ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/diagnosis` |
| **コンポーネント** | `[R] DiagnosisSettings` |
| **目的** | 診断カテゴリと診断名の2カテゴリを1ページで管理 |

### タブ構成

| タブラベル | カテゴリキー | showPrice | showCode | showCategory |
|---|---|---|---|---|
| カテゴリ | `diagnosis_category` | false | false | true |
| 診断名 | `diagnosis_name` | false | false | true |

---

## 11.5 トリミングマスタ（統合ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/trimming` |
| **コンポーネント** | `[R] TrimmingSettings` |
| **目的** | トリミングコースとオプションの2カテゴリを1ページで管理 |

### タブ構成

| タブラベル | カテゴリキー | showPrice | showParentItem |
|---|---|---|---|
| コース | `trimming_course` | true | true |
| オプション | `trimming_option` | true | true |

---

## 11.6 マスタカテゴリ設定（個別ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/{category-slug}`（6パターン） |
| **コンポーネント** | `[R] Settings` → `[C] SettingsContent` |

### ルートマッピング（6カテゴリ）

| スラグ | カテゴリキー | ラベル | アイコン | showPrice | showCode | showCategory |
|---|---|---|---|---|---|---|
| `service-type` | `serviceType` | 予約区分マスタ | Activity | false | false | false |
| `medicine` | `medicine` | 薬剤マスタ | Pill | true | false | false |
| `hospitalization` | `hospitalization` | 入院マスタ | Bed | true | false | false |
| `cage` | `cage` | ケージマスタ | Building2 | false | true | true |
| `staff` | `staff` | スタッフマスタ | Users | false | true | true |
| `insurance` | `insurance` | 保険マスタ | ShieldCheck | false | false | true |

### 画面構成（リストモード）

- ヘッダー: カテゴリ名 + 戻るボタン（`/settings`） + 新規登録ボタン
- 検索バー（`SearchFilterBar`）
- データテーブル（`DataTable`）

#### テーブル列

| 列 | 表示内容 | 条件 |
|---|---|---|
| コード | `item.code`（等幅フォント） | `showCode` 時 |
| 名称 | `item.name`（太字）、ドラッグハンドル付き | 常時 |
| 分類 | `item.category` | `showCategory` 時 |
| 単価(税込) | `¥{price}` | `showPrice` 時 |
| ステータス | `StatusBadge` | 常時 |
| 操作 | 編集ボタン | 常時 |

#### ツリー表示（`showParentItem: true` のカテゴリ）

- トップレベル項目＝カテゴリとして機能（Chevronで展開/折りたたみ）
- 子項目数をカウントバッジで表示
- 最下層のみ金額表示
- D&Dで階層変更・並び替え（`useDragAndDrop` + `useKeyboardReorder`）
- ホバー自動展開: 600msホバーで折りたたまれた親ノードを自動展開
- カスタムドラッグプレビュー（Notionライクなピル型ゴースト）

**D&D操作:**
- 並び順変更: 行の上端25%「前に挿入」、下端25%「後に挿入」、中央50%「子項目化」
- ドラッグ中テーブル上部に「トップレベルに移動」ドロップゾーン出現

**キーボードアクセシビリティ:**
- `Alt+ArrowUp/Down`: 並び替え
- `Alt+ArrowLeft`: 親の階層に昇格
- `Alt+ArrowRight`: 直前の兄弟の子に降格（循環参照防止付き）
- `useAnnounce` で全D&D操作結果をスクリーンリーダーに通知

---

## カテゴリ固有フォームセクション

### examination（検査マスタ）

| フィールド | 入力部品 | 備考 |
|---|---|---|
| 検査項目リスト | 動的リスト（追加/削除） | 3カラムグリッド per 行 |
| └ 項目名 | `Input` | placeholder: RBC |
| └ 単位 | `Input` | placeholder: mg/dL |
| └ 正常値 | `Input` | placeholder: 550-850 |
| └ 削除 | Trash2 ボタン（赤色 ghost） | |
| 項目追加 | `Button`（Plus アイコン） | ヘッダー右上に配置 |

### vaccine（予防接種マスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 対象種別 | `Select` | ✅ | 犬/猫/共通 |
| 標準接種間隔 | `Input` | - | placeholder: 1年 |

### medicine（薬剤マスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 剤形 | `Select` | ✅ | 錠剤/液剤/注射/外用薬/粉末 |
| 単位 | `Select` | ✅ | 1錠あたり/1mLあたり/1回分/1gあたり |

### staff（スタッフマスタ）

**特記事項:**
- 社員番号フィールドは存在せず、コード列も非表示
- 職種は `job_title` マスタから動的取得（単一選択、Combobox形式）
- 所属医院は複数選択可能（チェックボックスDropdownMenu形式）
- ユーザー種別により権限管理（システム管理者/医院管理者/スタッフ）

| フィールド | 入力部品 | 必須 | エラーメッセージ |
|---|---|---|---|
| 職種 | `Combobox`（job_title マスタ連動） | ✅ | 「職種を選択してください」 |
| 資格番号 | `Input` | - | - |
| 所属医院 | `DropdownMenu`（CheckboxItem） | ✅ | 「所属医院を選択してください」 |
| メールアドレス | `Input`（type=email） | ✅ | 「有効なメールアドレスを入力してください」 |
| パスワード | `Input`（type=password） | ✅（新規時） | 「パスワードは8文字以上で入力してください」 |
| ユーザー種別 | `Combobox` | ✅ | 「ユーザー種別を選択してください」 |

**スタッフマスタ 一覧表示列:**

| 列 | 表示内容 |
|---|---|
| 名称 | `item.name`（ドラッグハンドル付き） |
| 職種 | job_title マスタから解決した名称 |
| 所属医院 | clinics配列から解決した医院名（カンマ区切り） |
| メールアドレス | `item.email` |
| 最終ログイン | `item.lastLoginAt`（yyyy/MM/dd HH:mm） |
| ステータス | StatusBadge（有効/無効） |
| 操作 | RowActionDropdown（編集） |

### cage（ケージマスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| ケージタイプ | `Select` | ✅ | ICU（酸素室）/犬舎/猫舎/共用 |
| サイズ | `Select` | ✅ | 小型/中型/大型 |

### insurance（保険マスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 補償割合 (%) | `Select` | ✅ | 50%/70%/80%/100% |
| 請求先電話番号 | `Input` | - | placeholder: 0120-XXX-XXX |

### trimming_course（トリミングコースマスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 対象サイズ | `Select` | ✅ | 小型犬/中型犬/大型犬/猫 |
| 所要時間 (分) | `Input`（type=number） | - | Clock アイコン付き |

### trimming_option（トリミングオプションマスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 追加所要時間 (分) | `Input`（type=number） | - | Clock アイコン付き |
| 併用可否 | `Select` | ✅ | 併用可/単独のみ |

### hospitalization（入院マスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 対象体格 | `Select` | ✅ | 小型/中型/大型 |
| 料金単位 | `Select` | ✅ | 1日あたり/1泊あたり |

### consultation（診察マスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 適用区分 | `Select` | ✅ | 常時/初診/再診/時間外/緊急 |
| 標準診察時間 | `Input` | - | placeholder: 15分 |

### procedure（処置マスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 所要時間(目安) | `Input` | - | placeholder: 30分 |
| 麻酔要否 | `Select` | ✅ | 不要/局所麻酔/鎮静/全身麻酔 |

### checkup（定期健診マスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 推奨受診間隔 | `Input` | - | placeholder: 1年 |
| 対象年齢 | `Select` | ✅ | 全年齢/幼齢(〜1歳)/成年(1〜7歳)/シニア(7歳〜) |

### serviceType（予約区分マスタ）

| フィールド | 入力部品 | 必須 | 備考 |
|---|---|---|---|
| 表示カラー | カラーピッカー（丸ボタン群） | ✅ | `SERVICE_TYPE_COLOR_VALUES`、選択時チェック+ring |
| プレビュー | 読み取り専用バッジ | - | 選択色+名称のプレビュー表示 |

### diagnosis_name（診断名マスタ）

| フィールド | 入力部品 | 必須 | エラーメッセージ |
|---|---|---|---|
| 診断カテゴリ | `Select`（diagnosis_category 連動） | ✅ | 「診断カテゴリを選択してください」 |

### diagnosis_category（診断カテゴリマスタ）

固有セクションなし。共通フィールド（コード、名称）のみ。

---

## 共通フォーム項目（全カテゴリ共通）

| フィールド | 入力部品 | バリデーション | 必須 |
|---|---|---|---|
| コード | `Input` | 50文字以内 | ✅（showCode時） |
| 名称 | `Input` | 50文字以内 | ✅ |
| 分類 | `Input` | 50文字以内 | - |
| 単価(税込) | `Input`（type=number） | 0以上 | - |
| [カテゴリ固有セクション] | `MasterItemFormSections` | - | - |
| 備考/詳細 | `Input` | 200文字以内 | - |
| ステータス | `RadioGroup`（有効/無効） | - | ✅ |

## 編集モード アクション

| ボタン | 動作 | 備考 |
|---|---|---|
| 削除 | `handleDelete` → staff時は `StaffImpactDialog` / 他は `ConfirmDialog` | 赤文字、編集時のみ表示 |
| キャンセル | リストモードへ | outline |
| 保存 | `handleSave` | Save アイコン、`PrimaryButton` |

**スタッフ固有ダイアログ（`StaffImpactDialog`）:**
- ステータス変更・名称変更・削除時に影響範囲を確認
- `staffName`, `action`（rename/deactivate/delete）, `usage`（使用箇所数）を表示

---

## 使用コンポーネント一覧

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `Settings` | `[R]` | メインページ（リスト/編集切替） |
| `MasterSettingsIndex` | `[R]` | マスタカテゴリ一覧 |
| `SettingsContent` | `[C]` | リスト/編集表示本体（embeddedモード対応） |
| `MasterItemEditForm` | `[C]` | マスタ項目編集フォーム |
| `MasterItemFormSections` | `[C]` | カテゴリ固有フォームセクション（ディスパッチャー） |
| `MasterFlatDataTable` | `[C]` | フラットテーブル表示 |
| `MasterTreeDataTable` | `[C]` | ツリーテーブル表示 |
| `MasterStatusDot` | `[C]` | ステータスドット表示 |
| `MasterInlineAdd` | `[C]` | インライン追加フォーム |
| `MasterSidePeek` | `[C]` | サイドパネル（アニメーション付き） |
| `ExaminationSection` | `[C]` | 検査マスタ固有フォーム |
| `VaccineSection` | `[C]` | 予防接種マスタ固有フォーム |
| `MedicineSection` | `[C]` | 薬剤マスタ固有フォーム |
| `StaffSection` | `[C]` | スタッフマスタ固有フォーム |
| `CageSection` | `[C]` | ケージマスタ固有フォーム |
| `InsuranceSection` | `[C]` | 保険マスタ固有フォーム |
| `TrimmingCourseSection` | `[C]` | コースマスタ固有フォーム |
| `TrimmingOptionSection` | `[C]` | オプションマスタ固有フォーム |
| `HospitalizationSection` | `[C]` | 入院マスタ固有フォーム |
| `DiagnosisNameSection` | `[C]` | 診断名マスタ固有フォーム |
| `ConsultationSection` | `[C]` | 診察マスタ固有フォーム |
| `ProcedureSection` | `[C]` | 処置マスタ固有フォーム |
| `CheckupSection` | `[C]` | 定期健診マスタ固有フォーム |
| `ServiceTypeSection` | `[C]` | 予約区分マスタ固有フォーム |
| `SectionWrapper` | `[C]` | セクション共通ラッパー |
| `NotionPropertyRow` | `[S]` | Notion風プロパティ行 |
| `PageLayout` | `[S]` | ページレイアウト |
| `SearchFilterBar` | `[S]` | 検索フィルタバー |
| `DataTable` / `DataTableRow` | `[S]` | データテーブル |
| `PrimaryButton` | `[S]` | 新規登録・保存ボタン |
| `StatusBadge` | `[S]` | ステータスバッジ |
| `StaffImpactDialog` | `[S][M]` | スタッフ変更影響確認ダイアログ |
| `ConfirmDialog` | `[S][M]` | 削除確認ダイアログ |
| `MasterLink` | `[S]` | マスタ設定リンク |
| `useMasterItemEditor` | `[H]` | CRUD操作フック |
| `useMasterItems` | `[H]` | マスタデータ取得フック |
| `useDragAndDrop` | `[H]` | マウスD&Dフック |
| `useKeyboardReorder` | `[H]` | キーボード並び替えフック（Alt+Arrow全4方向） |

---

## データ型定義

```typescript
// マスタカテゴリ
const MASTER_CATEGORY_VALUES = [
  'examination', 'vaccine', 'medicine', 'staff', 'insurance', 'cage',
  'serviceType', 'consultation', 'procedure', 'hospitalization',
  'trimming_course', 'trimming_option', 'diagnosis_category', 'diagnosis_name', 'checkup'
] as const;
type MasterCategory = (typeof MASTER_CATEGORY_VALUES)[number];

// マスタアイテム基本型
interface MasterItem {
  id: string;
  name: string;
  code?: string;
  category: MasterCategory;
  price?: number;
  parentItemId?: string | null;
  sortOrder?: number;
  details?: string;
  status: "active" | "inactive";
  createdAt: string;
  updatedAt: string;
}

// カテゴリ別固有型
type ConsultationTime = "anytime" | "first_visit" | "revisit" | "after_hours" | "emergency";
type Anesthesia = "none" | "local" | "sedation" | "general";
type CheckupTargetAge = "all" | "puppy" | "adult" | "senior";
type VaccineSpecies = "dog" | "cat" | "both";
type DosageForm = "tablet" | "liquid" | "injection" | "external" | "powder";
type MedicineUnit = "per_tablet" | "per_ml" | "per_dose" | "per_gram";
type CageType = "icu" | "dog" | "cat" | "general";
type CageSize = "small" | "medium" | "large";
type CoverageRate = 50 | 70 | 80 | 100;
type TargetSize = "small" | "medium" | "large" | "cat";
type Combinable = "yes" | "no";
type BodySize = "small" | "medium" | "large";
type BillingUnit = "per_day" | "per_night";

interface ExaminationItem {
  id: string;
  itemName: string;
  unit: string;
  normalRange: string;
}
```

---

## ユーザー操作フロー

### リストモード
1. 検索バーでフィルタリング
2. 行クリックで編集モードへ遷移
3. 新規登録ボタンで編集モード（新規）へ遷移
4. ツリー表示の場合:
   - Chevronクリックで展開/折りたたみ
   - ドラッグ&ドロップで並び替え・階層変更
   - Alt+矢印キーで並び替え

### 編集モード
1. 共通フィールド入力
2. カテゴリ固有フィールド入力
3. ステータス切替（有効/無効ラジオ）
4. 保存 / キャンセル / 削除（編集時のみ）
5. スタッフカテゴリ: 名称変更・ステータス変更・削除時に影響確認ダイアログ

---

## アクセシビリティ

- **名称フィールド**: `aria-invalid` + `aria-describedby` → `FormFieldError`（`role="alert"`）接続
- **D&D操作**: `useAnnounce` でキーボード・マウス全操作結果をスクリーンリーダーに通知
- **MasterSelectModal**: `<button>` 要素でキーボード操作対応
- **キーボードナビゲーション**: Tab, Shift+Tab, Enter, Esc, Alt+矢印キー
- **フォーカス管理**: モーダル開閉時・D&D操作後にフォーカス復帰
- **スクリーンリーダー**: `aria-live`, `aria-label`, `aria-describedby` 適切に設定

---

## API連携

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/master-items` | マスタ項目一覧取得 | ✅ 実装済 |
| GET | `/api/v1/master-items/:id` | マスタ項目詳細取得 | ✅ 実装済 |
| POST | `/api/v1/master-items` | マスタ項目作成 | ✅ 実装済 |
| PUT | `/api/v1/master-items/:id` | マスタ項目更新 | ✅ 実装済 |
| DELETE | `/api/v1/master-items/:id` | マスタ項目削除 | ✅ 実装済 |
| GET | `/api/v1/clinics` | 病院情報取得 | ✅ 実装済 |
| PUT | `/api/v1/clinics/:id` | 病院情報更新 | ✅ 実装済 |

## 実装状況
- フロントエンド(ui-sample): ✅ 実装済（16カテゴリ、ツリー表示、D&D、キーボード操作）
- バックエンドAPI: ✅ 実装済（`/api/v1/master-items`）
