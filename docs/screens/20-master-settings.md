# マスタ設定 仕様書

**更新日:** 2026-03-12
**バージョン:** 4.0（SCREENS.md 11節準拠）

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
- 最終セクションに `pb-16` 付き

### セクション構成

| # | セクション名 | カテゴリキー |
|---|---|---|
| 1 | 基本設定 | `clinic` |
| 2 | カルテ | `treatmentItems` (診察/検査/処置/ワクチン/健診), `diagnosisGroup` (カテゴリ/診断名), `inquiry_template`, `medicine` |
| 3 | 診療関連マスタ | `serviceType` |
| 4 | 入院・ケージ管理 | `hospitalization`, `cage` |
| 5 | トリミング関連 | `trimmingGroup` (コース/オプション) |
| 6 | スタッフ・保険 | `staff`, `job_title`, `insurance` |

### カテゴリカード（`CategoryCard`）表示項目

| 項目 | 表示内容 |
|---|---|
| アイコン | `CATEGORY_CONFIG[key].IconComponent`（p-2 rounded-lg bg-[#F7F6F3]、ホバー時 bg-[#37352F] text-white） |
| カテゴリ名 | `cfg.label`（truncate） |
| 説明 | `cfg.description`（line-clamp-2） |
| 件数 | `{count}件登録済`（マスタカテゴリのみ、clinic は非表示） |
| 矢印 | ChevronRight（ホバー時 opacity 変化） |

### 特記

- `CATEGORY_CONFIG` から全カードを自動導出
- コンパイル時に全カテゴリの網羅性を型チェック（`ExhaustiveKeyMap`）
- `useMasterItems()` で全マスタアイテムを取得し、カテゴリ別にカウント

### 使用コンポーネント

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `MasterSettingsIndex` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `CategoryCard` | ローカル | カテゴリカード |
| `useMasterItems` | `[H]` | マスタデータ取得 |

---

## 11.2 病院情報設定

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/clinic` |
| **コンポーネント** | `[R] ClinicSettings` |
| **目的** | 病院の基本情報を管理する |

詳細は [19-clinic-settings.md](./19-clinic-settings.md) を参照。

---

## 11.3 診療項目マスタ（統合ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/treatment-items` |
| **コンポーネント** | `[R] TreatmentItemsSettings` |
| **目的** | 診察・検査・処置・予防接種・定期健診の5カテゴリを1ページで管理 |

### タブ構成

| タブラベル | カテゴリキー | showPrice | showCode | showCategory | showParentItem |
|---|---|---|---|---|---|
| 診察 | `consultation` | true | false | false | true |
| 検査 | `examination` | true | false | false | true |
| 処置 | `procedure` | true | false | false | true |
| 予防接種 | `vaccine` | true | false | false | true |
| 定期健診 | `checkup` | true | false | false | true |

- ヘッダー: 「診療項目マスタ」（Stethoscope アイコン）+ 戻るボタン → `/settings`
- タブ切替: `Tabs` / `TabsList` / `TabsTrigger` / `TabsContent`（Notion風スタイル）
- 各タブ内は `[C] SettingsContent`（`embedded` モード）で一覧/編集を描画（ツリー表示対応）

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

- ヘッダー: 「診断マスタ」（FolderTree アイコン）+ 戻るボタン → `/settings`
- タブ切替: `Tabs` / `TabsList` / `TabsTrigger` / `TabsContent`（Notion風スタイル）
- 各タブ内は `[C] SettingsContent`（`embedded` モード）で一覧/編集を描画

---

## 11.5 トリミングマスタ（統合ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/trimming` |
| **コンポーネント** | `[R] TrimmingSettings` |
| **目的** | トリミングコースとオプションの2カテゴリを1ページで管理 |

### タブ構成

| タブラベル | カテゴリキー | showPrice | showCode | showCategory | showParentItem |
|---|---|---|---|---|---|
| コース | `trimming_course` | true | false | false | true |
| オプション | `trimming_option` | true | false | false | true |

- ヘッダー: 「トリミングマスタ」（Scissors アイコン）+ 戻るボタン → `/settings`
- タブ切替: `Tabs` / `TabsList` / `TabsTrigger` / `TabsContent`（Notion風スタイル）
- 各タブ内は `[C] SettingsContent`（`embedded` モード）で一覧/編集を描画（ツリー表示対応）

---

## 11.6 マスタカテゴリ設定（個別ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/{category-slug}`（6パターン） |
| **コンポーネント** | `[R] Settings` → `[C] SettingsContent` |
| **目的** | 各マスタカテゴリのアイテムCRUD |

### ルートマッピング（6カテゴリ）

| スラグ | カテゴリキー | ラベル | アイコン | showPrice | showCode | showCategory | showParentItem |
|---|---|---|---|---|---|---|---|
| `service-type` | `serviceType` | 予約区分マスタ | Activity | false | false | false | false |
| `medicine` | `medicine` | 薬剤マスタ | Pill | true | false | false | true |
| `hospitalization` | `hospitalization` | 入院マスタ | Bed | true | false | false | true |
| `cage` | `cage` | ケージマスタ | Building2 | false | true | true | false |
| `staff` | `staff` | スタッフマスタ | Users | false | true | true | false |
| `insurance` | `insurance` | 保険マスタ | ShieldCheck | false | false | true | false |

### 画面構成（リストモード）

- ヘッダー: カテゴリ名（`config.IconComponent` 付き）+ 戻るボタン → `/settings` + 新規登録ボタン
- 検索バー（`[S] NotionFilter`）: `showCode` 時は `{config.labels.code}、{config.labels.name}で検索...`、それ以外は `{config.labels.name}で検索...`
- データテーブル（`[S] DataTable`）

#### テーブル列

| 列 | className / align | 表示内容 | 条件 |
|---|---|---|---|
| コード | `w-[120px]` | `item.code`（等幅フォント） | `showCode` 時のみ |
| 名称 | - | `item.name`（太字）、ドラッグハンドル付き。ツリー時はインデント+Chevron追加 | 常時 |
| 所属カテゴリ | `w-[130px]` | `diagnosisCategories` からの名前解決 | `diagnosis_name` のみ |
| 分類 | `w-[100px]` | `item.category` | `showCategory` 時のみ（`showParentItem: true` のカテゴリでは非表示） |
| 単価(税込) | `w-[100px]`, align:right | `¥{price}` or 「-」（等幅フォント） | `showPrice` 時のみ |
| ステータス | `w-[100px]`, align:center | `StatusBadge`（`getMasterStatusColor` / `getMasterStatusLabel`） | 常時 |
| 操作 | `w-[80px]`, align:right | 編集ボタン（行クリックでも編集モードへ） | 常時 |

### ツリー表示（`showParentItem: true` のカテゴリ）

- トップレベル項目＝「カテゴリ」として機能、Chevronで展開/折りたたみ
- 子項目数をカウントバッジで表示
- 最下層のみ金額表示、親は金額欄を空欄表示
- 操作列: 「+」ボタン（子項目インライン追加）+ 編集ボタン
- D&D で項目を別の親にドロップして所属カテゴリ変更（自己参照・子孫へのドロップは防止）

**D&D並び順変更ルール:**
- 行の上端25%: 「前に挿入」
- 行の下端25%: 「後に挿入」
- 中央50%: 「子項目化」
- ドラッグ中テーブル上部に「トップレベルに移動」ドロップゾーン出現
- カスタムドラッグプレビュー: Notionライクなピル型ゴースト（GripVerticalアイコン＋項目名）
- ホバー自動展開: 折りたたまれた親ノードの中央ゾーンに600msホバーで自動展開

**キーボードアクセシビリティ（D&D操作）:**
- `Alt+ArrowUp/Down`: 並び替え
- `Alt+ArrowLeft`: 親の階層に昇格
- `Alt+ArrowRight`: 直前の兄弟の子に降格（循環参照防止付き）
- 移動後はドラッグハンドルにフォーカス自動復帰
- `useAnnounce` で全操作結果を `aria-live="polite"` リージョンにリアルタイム通知

**インライン追加行:**
- `Enter` で追加（連打で複数登録可）
- `Esc` で閉じる

### 画面構成（編集モード）

- ヘッダー: 「{カテゴリ名} 編集/新規登録」、戻るボタン → リストモードへ
- カードコンテナ: `bg-white p-6 rounded-lg border shadow-sm space-y-4`

#### 共通フォーム項目（全カテゴリ共通）

| フィールド | 入力部品 | グリッド | 備考 |
|---|---|---|---|
| カテゴリ（親項目） | `Select`（カテゴリ候補リスト） | full | `showParentItem` 時のみ表示、循環参照防止済み。「なし（トップレベル）」選択可 |
| コード | `Input` | 2cols-左 | `showCode` 時のみ、必須（`*`マーク）、placeholder: `config.codePlaceholder` |
| 名称 | `Input` | 2cols-右 | 必須（`*`マーク）、placeholder: `config.namePlaceholder` |
| 分類 | `Input` | 2cols-左 | `showCategory` 時のみ、placeholder: `config.categoryPlaceholder` |
| 単価(税込) | `Input`（type=number）/ 「子項目で金額を設定」表示 | 2cols-右 | `showPrice` 時のみ。子を持つ親は入力不可 |
| [カテゴリ固有セクション] | `MasterItemFormSections` | - | 下記参照 |
| 備考 / 詳細 | `Input` | full | placeholder: 補足情報など |
| ステータス | `RadioGroup`（有効 / 無効） | full | radio ボタン2つ |

---

## カテゴリ固有フォームセクション

### examination（検査マスタ）

| フィールド | 入力部品 | 備考 |
|---|---|---|
| 検査項目リスト | 動的リスト（追加/削除） | 3カラムグリッド per 行 |
| └ 項目名 | `Input` | placeholder: RBC |
| └ 単位 | `Input` | placeholder: 例: mg/dL |
| └ 正常値 | `Input` | placeholder: 550-850 |
| └ 削除 | Trash2 ボタン | 赤色 ghost |
| 項目追加 | `Button`（Plus アイコン） | ヘッダー右上に配置 |

### vaccine（予防接種マスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 対象種別 | `Select` | - | 犬（dog）/ 猫（cat）/ 共通（both）、デフォルト: dog |
| 標準接種間隔 | `Input` | - | placeholder: 例: 1年 |

### medicine（薬剤マスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 剤形 | `Select` | - | 錠剤 / 液剤 / 注射 / 外用薬 / 粉末、デフォルト: tablet |
| 単位 | `Select` | - | 1錠あたり / 1mLあたり / 1回分 / 1gあたり、デフォルト: per_tablet |

### staff（スタッフマスタ）

**特記事項:**
- 社員番号フィールドは存在せず、コード列も非表示
- 職種は `job_titles` マスタから動的取得（単一選択、Combobox形式）
- 所属医院は複数選択可能（CheckboxItem型のDropdownMenu形式）
- ユーザー種別により権限管理（システム管理者/医院管理者/スタッフ）
- セクション構成: スタッフ詳細（職種・資格番号） / 所属情報（所属医院） / アカウント発行・権限設定（メールアドレス・パスワード・ユーザー種別）

| フィールド | 入力部品 | 必須 | エラーメッセージ |
|---|---|---|---|
| 職種 | `Combobox`（job_title マスタ連動、active のみ、`MasterLink` 付き） | - | - |
| 資格番号 | `Input` | - | placeholder: 例: 獣医第12345号 |
| 所属医院 | `DropdownMenu`（`DropdownMenuCheckboxItem`、複数選択） | 必須 | 「所属医院を選択してください」 |
| メールアドレス | `Input`（type=email） | 必須 | 「有効なメールアドレスを入力してください」 |
| パスワード | `Input`（type=password） | 必須（新規時のみ） | 「パスワードは8文字以上で入力してください」 |
| ユーザー種別 | `Combobox`（`USER_TYPE_VALUES`: システム管理者/医院管理者/スタッフ） | 必須 | 「ユーザー種別を選択してください」 |

**スタッフマスタ 一覧表示列:**

| 列 | 表示内容 |
|---|---|
| 名称 | `item.name`（ドラッグハンドル付き） |
| 職種 | job_title マスタから解決した名称 |
| 所属医院 | clinics配列から解決した医院名（カンマ区切り） |
| メールアドレス | `item.email` |
| 最終ログイン | `item.lastLoginAt`（yyyy/MM/dd HH:mm 形式） |
| ステータス | StatusBadge（有効/無効） |
| 操作 | RowActionDropdown（編集） |

### cage（ケージマスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| ケージタイプ | `Select` | - | ICU（酸素室）/ 犬舎 / 猫舎 / 共用、デフォルト: general |
| サイズ | `Select` | - | 小型 / 中型 / 大型、デフォルト: medium |

### insurance（保険マスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 補償割合 (%) | `Select` | - | 50% / 70% / 80% / 100%、デフォルト: 70 |
| 請求先電話番号 | `Input` | - | placeholder: 例: 0120-XXX-XXX |

### trimming_course（トリミングコースマスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 対象サイズ | `Select` | - | 小型犬 / 中型犬 / 大型犬 / 猫、デフォルト: small |
| 所要時間 (分) | `Input`（type=number、Clock アイコン付き） | - | placeholder: 例: 60 |

### trimming_option（トリミングオプションマスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 追加所要時間 (分) | `Input`（type=number、Clock アイコン付き） | - | placeholder: 例: 15 |
| 併用可否 | `Select` | - | 併用可（yes）/ 単独のみ（no）、デフォルト: yes |

### hospitalization（入院マスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 対象体格 | `Select` | - | 小型 / 中型 / 大型、デフォルト: small |
| 料金単位 | `Select` | - | 1日あたり / 1泊あたり、デフォルト: per_day |

### consultation（診察マスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 適用区分 | `Select` | - | 常時 / 初診 / 再診 / 時間外 / 緊急、デフォルト: anytime |
| 標準診察時間 | `Input` | - | placeholder: 例: 15分 |

### procedure（処置マスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 所要時間(目安) | `Input` | - | placeholder: 例: 30分 |
| 麻酔要否 | `Select` | - | 不要 / 局所麻酔 / 鎮静 / 全身麻酔、デフォルト: none |

### checkup（定期健診マスタ）

| フィールド | 入力部品 | 必須 | 選択肢 |
|---|---|---|---|
| 推奨受診間隔 | `Input` | - | placeholder: 例: 1年 |
| 対象年齢 | `Select` | - | 全年齢 / 幼齢(〜1歳) / 成年(1〜7歳) / シニア(7歳〜)、デフォルト: all |

### serviceType（予約区分マスタ）

| フィールド | 入力部品 | 必須 | 備考 |
|---|---|---|---|
| 表示カラー | カラーピッカー（`SERVICE_TYPE_COLOR_VALUES` 丸ボタン群） | - | 選択時チェック+ring、デフォルト: default |
| プレビュー | 読み取り専用バッジ | - | 選択色+名称のプレビュー表示 |

### diagnosis_name（診断名マスタ）

| フィールド | 入力部品 | 必須 | エラーメッセージ |
|---|---|---|---|
| 診断カテゴリ | `Select`（diagnosis_category マスタ連動、active のみ） | 必須 | 「診断カテゴリを選択してください」（未登録時は警告メッセージ表示） |

### diagnosis_category（固有セクションなし）

共通フィールド（コード、名称）のみ。

---

## 編集モード アクション

| ボタン | 位置 | 動作 | 備考 |
|---|---|---|---|
| 削除 | 左（編集時のみ） | `handleDelete` → staff 時は `StaffImpactDialog` / 他は `ConfirmDialog` | Trash2 アイコン、赤文字 |
| キャンセル | 右 | `handleCloseEdit` → リストモードへ | outline |
| 保存 | 右 | `handleSave` | Save アイコン、`[S] PrimaryButton` |

**スタッフ固有ダイアログ（`[S] StaffImpactDialog`）:**
- ステータス変更・名称変更・削除時に影響範囲を確認
- `staffName`, `action`（rename / deactivate / delete）, `usage`（使用箇所数）を表示

---

## 使用コンポーネント一覧

| コンポーネント | 種別 | 説明 |
|---|---|---|
| `Settings` | `[R]` | メインページ（リスト/編集切替） |
| `MasterSettingsIndex` | `[R]` | マスタカテゴリ一覧 |
| `TreatmentItemsSettings` | `[R]` | 診療項目マスタ統合ページ（5タブ） |
| `DiagnosisSettings` | `[R]` | 診断マスタ統合ページ（2タブ） |
| `TrimmingSettings` | `[R]` | トリミングマスタ統合ページ（2タブ） |
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
| `SectionWrapper` | `[C]` | セクション共通ラッパー（`SectionPropertyRow` = `NotionPropertyRow` re-export） |
| `NotionPropertyRow` | `[S]` | Notion風プロパティ行 |
| `PageLayout` | `[S]` | ページレイアウト |
| `NotionFilter` | `[S]` | 検索フィルタバー |
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
   - Alt+矢印キーでキーボード並び替え
   - 「+」ボタンでインライン子項目追加

### 編集モード
1. 共通フィールド入力
2. カテゴリ固有フィールド入力
3. ステータス切替（有効/無効ラジオ）
4. 保存 / キャンセル / 削除（編集時のみ）
5. スタッフカテゴリ: 名称変更・ステータス変更・削除時に影響確認ダイアログ
6. diagnosis_name カテゴリ: 親カテゴリ（diagnosis_category マスタ）からの選択

---

## アクセシビリティ

- **名称フィールド**: `aria-invalid` + `aria-describedby` → `FormFieldError`（`role="alert"`）接続
- **D&D操作**: `useAnnounce` でキーボード・マウス全操作結果をスクリーンリーダーに通知
- **ドラッグハンドル**: `role="button"` / `tabIndex={0}` / `aria-roledescription` / `aria-label` / `aria-grabbed` / `aria-dropeffect="move"`
- **MasterSelectModal**: `<button>` 要素でキーボード操作対応
- **キーボードナビゲーション**: Tab, Shift+Tab, Enter, Esc, Alt+矢印キー
- **フォーカス管理**: モーダル開閉時・D&D操作後にフォーカス復帰
- **スクリーンリーダー**: `aria-live`, `aria-label`, `aria-describedby` 適切に設定

---

## 機能詳細

### 1. 階層構造の管理（DnD）
- **循環参照の防止**: ドラッグ&ドロップによる親カテゴリ変更時、自分自身や自分の子孫を親に設定しようとする操作をロジックで無効化する。
- **一括ソート**: ドラッグ操作が完了するたびに、同じ階層内での並び順（`sort_order`）が更新され、API を通じて保存される。

### 2. サイドパネル編集（Side Peek）
- **コンテキスト維持**: 一覧画面を維持したまま、Notion 風のサイドパネルで詳細を編集可能。
- **自動保存とリセット**: 変更がある場合のみ保存ボタンが活性化し、キャンセル時には変更前の状態に戻るロジックを備える。

### 3. マスタ削除の影響範囲
- **整合性チェック**: 既にカルテや予約で使用されているマスタ項目を削除しようとした場合、物理削除ではなく「無効化（`is_active: false`）」を推奨する、または関連データの存在を警告する。

| メソッド | エンドポイント | 用途 | 状態 |
|---------|--------------|------|------|
| GET | `/api/v1/master-items` | マスタ項目一覧取得 | 実装済 |
| GET | `/api/v1/master-items/:id` | マスタ項目詳細取得 | 実装済 |
| POST | `/api/v1/master-items` | マスタ項目作成 | 実装済 |
| PATCH | `/api/v1/master-items/:id` | マスタ項目更新 | 実装済 |
| DELETE | `/api/v1/master-items/:id` | マスタ項目削除 | 実装済 |
| GET | `/api/v1/staffs` | スタッフ一覧取得 | 実装済 |
| PATCH | `/api/v1/staffs/:id` | スタッフ更新 | 実装済 |

## 実装状況

- フロントエンド: 実装済（`features/master/routes/MasterSettingsIndex.tsx` 等）
- バックエンドAPI: 実装済（`handler/master_handler.go`, `handler/staff_handler.go`）
