# 動物病院管理システム 画面仕様書（マスタ管理編）

**更新日**: 2026-03-11  
**バージョン**: 3.0

本ドキュメントは、設定・マスタ管理機能の詳細仕様を定義します。

> **関連ドキュメント**:
> - [SCREENS.md](./SCREENS.md): メイン機能の画面仕様
> - [FORMS_SPECIFICATION.md](./docs/FORMS_SPECIFICATION.md): フォーム項目の詳細仕様

---

## 凡例

| 記号 | 意味 |
|------|------|
| `[R]` | ルートコンポーネント (`routes/` 配下) |
| `[C]` | 機能固有コンポーネント (`components/` 配下) |
| `[S]` | 共有コンポーネント (`/components/shared/`) |
| `[H]` | フック (`hooks/` 配下) |
| `[M]` | モーダル / ダイアログ |

---

## 11. 設定・マスタ管理

### 11.1 マスタ設定トップ

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings` |
| **コンポーネント** | `[R] MasterSettingsIndex` |
| **目的** | マスタカテゴリ一覧をカード形式で表示 |

**画面構成:**
- ヘッダー: タイトル「マスタ設定」（Settings アイコン）
- セクション分類（5セクション）
- カードグリッド（`grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3`）
- 最終セクションに `pb-16` 付き

**セクション構成:**
| # | セクション名 | カテゴリキー |
|---|---|---|
| 1 | 基本設定 | `clinic` |
| 2 | 診療関連マスタ | `serviceType`, `consultation`, `examination`, `procedure`, `vaccine`, `medicine`, `diagnosis_category`, `diagnosis_name` |
| 3 | 入院・ケージ管理 | `hospitalization`, `cage` |
| 4 | トリミング関連 | `trimming_course`, `trimming_option` |
| 5 | スタッフ・保険 | `staff`, `insurance` |

**カテゴリカード（`CategoryCard`）表示項目:**
| 項目 | 表示内容 |
|---|---|
| アイコン | `CATEGORY_CONFIG[key].IconComponent`（p-2 rounded-lg bg-[#F7F6F3]、ホバー時 bg-[#37352F] text-white） |
| カテゴリ名 | `cfg.label`（truncate） |
| 説明 | `cfg.description`（line-clamp-2） |
| 件数 | `{count}件登録済`（マスタカテゴリのみ、clinic は undefined） |
| 矢印 | ChevronRight（ホバー時 opacity 変化） |

**特記:**
- `CATEGORY_CONFIG` から全カードを自動導出
- コンパイル時に全カテゴリの網羅性を型チェック（`ExhaustiveKeyMap`）
- `useMasterItems()` で全マスタアイテムを取得し、カテゴリ別にカウント

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `MasterSettingsIndex` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `CategoryCard` | ローカル | カテゴリカード |
| `useMasterItems` | `[H]` | マスタデータ取得 |

**データ型:** `MasterCategory`, `MasterCardKey`, `MasterCategoryCard`, `MasterSection`, `CategoryConfig`

---

### 11.2 病院情報設定

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/clinic` |
| **コンポーネント** | `[R] ClinicSettings` |
| **目的** | 病院の基本情報を管理する |

**画面構成:**
- ヘッダー: タイトル「病院情報設定」（Building2 アイコン）、戻るボタン → `/settings`
- フォーム（`react-hook-form@7.55.0`）
- ローディング時: スケルトン表示（6行の `animate-pulse`）
- カードコンテナ: `bg-white p-6 rounded-lg shadow-sm border`

**フォーム項目:**
| フィールド | 入力部品 | グリッド | バリデーション | 必須 | エラーメッセージ |
|---|---|---|---|---|---|
| 病院名 | `Input`（register: `name`） | 2cols-左 | 50文字以内 | ✅ | 「病院名は必須です」 |
| 支店名 | `Input`（register: `branchName`） | 2cols-右 | 50文字以内 | - | 「支店名は50文字以内で入力してください」 |
| 郵便番号 | `Input`（register: `postalCode`） | 3cols-左 | `^\d{3}-\d{4}$` | - | 「郵便番号は000-0000の形式で入力してください」 |
| 住所 | `Input`（register: `address`） | 3cols-中右（col-span-2） | 200文字以内 | - | 「住所は200文字以内で入力してください」 |
| 電話番号 | `Input`（register: `phoneNumber`） | 2cols-左 | 日本形式 | ✅ | 「有効な電話番号を入力してください」 |
| FAX番号 | `Input`（register: `faxNumber`） | 2cols-右 | 日本形式 | - | 「有効なFAX番号を入力してください」 |
| 登録番号 | `Input`（register: `registrationNumber`） | full | 100文字以内 | - | - |
| 院長名 | `Input`（register: `directorName`） | full | 50文字以内 | - | 「院長名は50文字以内で入力してください」 |
| メールアドレス | `Input`（type=email, register: `email`） | 2cols-左 | RFC 5322準拠 | - | 「有効なメールアドレスを入力してください」 |
| WebサイトURL | `Input`（register: `website`） | 2cols-右 | URL形式 | - | 「有効なURLを入力してください」 |

**アクション（`pt-4 border-t`）:**
| ボタン | 動作 | 備考 |
|---|---|---|
| キャンセル | `/settings` へ遷移 | outline |
| 設定を保存 | `handleSubmit(onSubmit)` → `updateClinicInfo` + `reset` | Save アイコン、`isDirty` false 時 disabled |

**保存処理:**
1. バリデーション実行
2. エラーがある場合、該当フィールドにエラー表示
3. API呼び出し（updateClinicInfo）
4. 成功時:
   - トースト: 「病院情報を更新しました」（success）
   - フォームをリセット（isDirty = false）
5. 失敗時:
   - トースト: 「更新に失敗しました」（error）

**使用コンポーネント:**
| コンポーネント | 種別 | 説明 |
|---|---|---|
| `ClinicSettings` | `[R]` | メインページ |
| `PageLayout` | `[S]` | ページレイアウト |
| `NotionPropertyRow` | `[S]` | Notion風プロパティ行（label/required/align） |
| `NotionSectionLabel` | `[S]` | Notion風セクションラベル |
| `NotionSectionDivider` | `[S]` | Notion風薄罫線ディバイダー |
| `useClinicInfo` | `[H]` | 病院情報CRUD |

**データ型:**

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

### 11.3 診療項目マスタ（統合ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/treatment-items` |
| **コンポーネント** | `[R] TreatmentItemsSettings` |
| **目的** | 診察・検査・処置・予防接種・定期健診の5カテゴリを1ページで管理 |

**タブ構成:**
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
- 編集モード時は `PageLayout` 付きの独立フォームに遷移

---

### 11.4 診断マスタ（統合ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/diagnosis` |
| **コンポーネント** | `[R] DiagnosisSettings` |
| **目的** | 診断カテゴリと診断名の2カテゴリを1ページで管理 |

**タブ構成:**
| タブラベル | カテゴリキー | showPrice | showCode | showCategory |
|---|---|---|---|---|
| カテゴリ | `diagnosis_category` | false | false | true |
| 診断名 | `diagnosis_name` | false | false | true |

- ヘッダー: 「診断マスタ」（FolderTree アイコン）+ 戻るボタン → `/settings`
- タブ切替: `Tabs` / `TabsList` / `TabsTrigger` / `TabsContent`（Notion風スタイル）
- 各タブ内は `[C] SettingsContent`（`embedded` モード）で一覧/編集を描画

---

### 11.5 トリミングマスタ（統合ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/trimming` |
| **コンポーネント** | `[R] TrimmingSettings` |
| **目的** | トリミングコースとオプションの2カテゴリを1ページで管理 |

**タブ構成:**
| タブラベル | カテゴリキー | showPrice | showCode | showCategory | showParentItem |
|---|---|---|---|---|---|
| コース | `trimming_course` | true | false | false | true |
| オプション | `trimming_option` | true | false | false | true |

- ヘッダー: 「トリミングマスタ」（Scissors アイコン）+ 戻るボタン → `/settings`
- タブ切替: `Tabs` / `TabsList` / `TabsTrigger` / `TabsContent`（Notion風スタイル）
- 各タブ内は `[C] SettingsContent`（`embedded` モード）で一覧/編集を描画（ツリー表示対応）

---

### 11.6 マスタカテゴリ設定（個別ページ）

| 項目 | 内容 |
|------|------|
| **ルート** | `/settings/{category-slug}`（6パターン） |
| **コンポーネント** | `[R] Settings` → `[C] SettingsContent` |
| **目的** | 各マスタカテゴリのアイテムCRUD |

**ルートマッピング（6カテゴリ）:**
| スラグ | カテゴリキー | ラベル | アイコン | showPrice | showCode | showCategory | showParentItem |
|---|---|---|---|---|---|---|---|
| `service-type` | `serviceType` | 予約区分マスタ | Activity | false | false | false | false |
| `medicine` | `medicine` | 薬剤マスタ | Pill | true | false | false | true |
| `hospitalization` | `hospitalization` | 入院マスタ | Bed | true | false | false | true |
| `cage` | `cage` | ケージマスタ | Building2 | false | true | true | false |
| `staff` | `staff` | スタッフマスタ | Users | false | true | true | false |
| `insurance` | `insurance` | 保険マスタ | ShieldCheck | false | false | true | false |

#### 画面構成（リストモード）

- ヘッダー: カテゴリ名（`config.IconComponent` 付き）+ 戻るボタン → `/settings` + 新規登録ボタン
- 検索バー（`[S] SearchFilterBar`）: `showCode` 時は `{config.labels.code}、{config.labels.name}で検索...`、それ以外は `{config.labels.name}で検索...`
- データテーブル（`[S] DataTable`）

#### ツリー表示（`showParentItem: true` のカテゴリ）

**特徴:**
- ドラッグ中にテーブル上部に「トップレベルに移動」ドロップゾーンが出現
- 各行にドラッグハンドル（`GripVertical` アイコン）付き
- トップレベル項目＝「カテゴリ」として機能、Chevronで展開/折りたたみ
- 子項目数をカウントバッジで表示
- 最下層のみ金額表示、親は金額欄を空欄表示
- 操作列: 「+」ボタン（子項目インライン追加）+ 編集ボタン

**D&D操作:**
- ドラッグで項目を別の親にドロップして所属カテゴリ変更（自己参照・子孫へのドロップは防止）
- 並び順変更: 行の上端25%にドロップで「前に挿入」、下端25%で「後に挿入」、中央50%で「子項目化」
- `sortOrder`フィールドで永続化、`bulkUpdate`で兄弟全体のsortOrderを一括更新
- カスタムドラッグプレビュー: `setDragImage`でNotionライクなピル型ゴースト（GripVerticalアイコン＋項目名）を表示
- ホバー自動展開: 折りたたまれた親ノードの中央ゾーンに600msホバーで自動展開

**キーボードアクセシビリティ:**
- ドラッグハンドルに `role="button"` / `tabIndex={0}` / `aria-roledescription` / `aria-label` を付与
- `aria-grabbed` / `aria-dropeffect="move"` で状態通知
- `Alt+ArrowUp/Down` でキーボードのみの並び替え
- `Alt+ArrowLeft` で親の階層に昇格
- `Alt+ArrowRight` で直前の兄弟の子に降格（循環参照防止付き）
- 移動後はハンドルにフォーカス自動復帰
- `useAnnounce` で全D&D操作結果を `aria-live="polite"` リージョンにリアルタイム通知

**インライン追加行:**
- `Enter` で追加（連打で複数登録可）
- `Esc` で閉じる

#### リストモード テーブル列

| 列 | className / align | 表示内容 | 条件 |
|---|---|---|---|
| コード | `w-[120px]` | `item.code`（等幅フォント） | `showCode` 時のみ |
| 名称 | - | `item.name`（太字）、ドラッグハンドル付き。ツリー時はインデント+Chevron追加。フラット時も`GripVertical`+D&D並び替え対応 | 常時 |
| 所属カテゴリ | `w-[130px]` | `diagnosisCategories` からの名前解決 | `diagnosis_name` のみ |
| 分類 | `w-[100px]` | `item.category` | `showCategory` 時のみ（`showParentItem: true` のカテゴリでは非表示） |
| 単価(税込) | `w-[100px]`, align:right | `¥{price}` or 「-」（等幅フォント） | `showPrice` 時のみ |
| ステータス | `w-[100px]`, align:center | `StatusBadge`（`getMasterStatusColor` / `getMasterStatusLabel`） | 常時 |
| 操作 | `w-[80px]`, align:right | 編集ボタン（行クリックでも編集モードへ遷移） | 常時 |

#### 画面構成（編集モード）

- ヘッダー: 「{カテゴリ名} 編集/新規登録」、戻るボタン → リストモードへ
- カードコンテナ: `bg-white p-6 rounded-lg border shadow-sm space-y-4`

**共通フォーム項目（全カテゴリ共通）:**

| フィールド | 入力部品 | グリッド | バリデーション | 必須 | エラーメッセージ |
|---|---|---|---|---|---|
| カテゴリ | `Select`（カテゴリ候補リスト） | full | 循環参照防止 | - | - |
| コード | `Input` | 2cols-左 | 50文字以内 | ✅ (showCode時) | 「コードは必須です」 |
| 名称 | `Input` | 2cols-右 | 50文字以内 | ✅ | 「名称は必須です」 |
| 分類 | `Input` | 2cols-左 | 50文字以内 | - | - |
| 単価(税込) | `Input`（type=number） | 2cols-右 | 0以上 | - | 「単価は0以上の数値を入力してください」 |
| [カテゴリ固有セクション] | `MasterItemFormSections` | - | - | - | - |
| 備考/詳細 | `Input` | full | 200文字以内 | - | - |
| ステータス | `RadioGroup`（有効/無効） | full | - | ✅ | - |

**カテゴリ固有セクション詳細は以下を参照**

---

## カテゴリ固有フォームセクション

### examination（検査マスタ）

| フィールド | 入力部品 | バリデーション | 必須 | 備考 |
|---|---|---|---|---|
| 検査項目リスト | 動的リスト（追加/削除） | - | - | 3カラムグリッド per 行 |
| └ 項目名 | `Input` | 50文字以内 | - | placeholder: RBC |
| └ 単位 | `Input` | 20文字以内 | - | placeholder: 例: mg/dL |
| └ 正常値 | `Input` | 50文字以内 | - | placeholder: 550-850 |
| └ 削除 | Trash2 ボタン | - | - | 赤色 ghost |
| 項目追加 | `Button`（Plus アイコン） | - | - | ヘッダー右上に配置 |

### vaccine（予防接種マスタ）

| フィールド | 入力部品 | バリデーション | 必須 | デフォルト | 選択肢 |
|---|---|---|---|---|---|
| 対象種別 | `Select` | - | ✅ | dog | 犬/猫/共通 |
| 標準接種間隔 | `Input` | 50文字以内 | - | - | placeholder: 例: 1年 |

### medicine（薬剤マスタ）

| フィールド | 入力部品 | バリデーション | 必須 | デフォルト | 選択肢 |
|---|---|---|---|---|---|
| 剤形 | `Select` | - | ✅ | tablet | 錠剤/液剤/注射/外用薬/粉末 |
| 単位 | `Select` | - | ✅ | per_tablet | 1錠あたり/1mLあたり/1回分/1gあたり |

### staff（スタッフマスタ）

**特記事項:**
- 社員番号フィールドは存在せず、コード列も非表示
- 職種は job_title マスタから動的に取得（単一選択、Combobox形式）
- 所属医院は複数選択可能（チェックボックス形式のドロップダウンメニュー）
- ユーザー種別により権限管理を実施（システム管理者/医院管理者/スタッフ）
- セクション構成: スタッフ詳細（職種・資格番号） / 所属情報（所属医院） / アカウント発行・権限設定（メールアドレス・パスワード・ユーザー種別）

| フィールド | 入力部品 | バリデーション | 必須 | エラーメッセージ |
|---|---|---|---|---|
| 職種 | `Combobox`（job_title マスタ連動） | - | ✅ | 「職種を選択してください」 |
| 資格番号 | `Input` | 50文字以内 | - | - |
| 所属医院 | `DropdownMenu`（CheckboxItem） | - | ✅ | 「所属医院を選択してください」 |
| メールアドレス | `Input`（type=email） | RFC 5322準拠 | ✅ | 「有効なメールアドレスを入力してください」 |
| パスワード | `Input`（type=password） | 8文字以上 | ✅ (新規時) | 「パスワードは8文字以上で入力してください」 |
| ユーザー種別 | `Combobox` | - | ✅ | 「ユーザー種別を選択してください」 |

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

| フィールド | 入力部品 | バリデーション | 必須 | デフォルト | 選択肢 |
|---|---|---|---|---|---|
| ケージタイプ | `Select` | - | ✅ | general | ICU（酸素室）/犬舎/猫舎/共用 |
| サイズ | `Select` | - | ✅ | medium | 小型/中型/大型 |

### insurance（保険マスタ）

| フィールド | 入力部品 | バリデーション | 必須 | デフォルト | 選択肢 |
|---|---|---|---|---|---|
| 補償割合 (%) | `Select` | - | ✅ | 70 | 50%/70%/80%/100% |
| 請求先電話番号 | `Input` | 日本形式 | - | - | placeholder: 例: 0120-XXX-XXX |

### trimming_course（トリミングコースマスタ）

| フィールド | 入力部品 | バリデーション | 必須 | デフォルト | 選択肢 |
|---|---|---|---|---|---|
| 対象サイズ | `Select` | - | ✅ | small | 小型犬/中型犬/大型犬/猫 |
| 所要時間 (分) | `Input`（type=number） | 1以上 | - | - | Clock アイコン付き、placeholder: 例: 60 |

### trimming_option（トリミングオプションマスタ）

| フィールド | 入力部品 | バリデーション | 必須 | デフォルト | 選択肢 |
|---|---|---|---|---|---|
| 追加所要時間 (分) | `Input`（type=number） | 1以上 | - | - | Clock アイコン付き、placeholder: 例: 15 |
| 併用可否 | `Select` | - | ✅ | yes | 併用可/単独のみ |

### hospitalization（入院マスタ）

| フィールド | 入力部品 | バリデーション | 必須 | デフォルト | 選択肢 |
|---|---|---|---|---|---|
| 対象体格 | `Select` | - | ✅ | small | 小型/中型/大型 |
| 料金単位 | `Select` | - | ✅ | per_day | 1日あたり/1泊あたり |

### consultation（診察マスタ）

| フィールド | 入力部品 | バリデーション | 必須 | デフォルト | 選択肢 |
|---|---|---|---|---|---|
| 適用区分 | `Select` | - | ✅ | anytime | 常時/初診/再診/時間外/緊急 |
| 標準診察時間 | `Input` | 50文字以内 | - | - | placeholder: 例: 15分 |

### procedure（処置マスタ）

| フィールド | 入力部品 | バリデーション | 必須 | デフォルト | 選択肢 |
|---|---|---|---|---|---|
| 所要時間(目安) | `Input` | 50文字以内 | - | - | placeholder: 例: 30分 |
| 麻酔要否 | `Select` | - | ✅ | none | 不要/局所麻酔/全身麻酔 |

### checkup（定期健診マスタ）

| フィールド | 入力部品 | バリデーション | 必須 | デフォルト | 選択肢 |
|---|---|---|---|---|---|
| 推奨受診間隔 | `Input` | 50文字以内 | - | - | placeholder: 例: 1年 |
| 対象年齢 | `Select` | - | ✅ | all | 全年齢/幼齢(〜1歳)/成年(1〜7歳)/シニア(7歳〜) |

### serviceType（予約区分マスタ）

| フィールド | 入力部品 | バリデーション | 必須 | 備考 |
|---|---|---|---|---|
| 表示カラー | カラーピッカー（丸ボタン群） | - | ✅ | `SERVICE_TYPE_COLOR_VALUES`、選択時チェック+ring |
| プレビュー | 読み取り専用バッジ | - | - | 選択色+名称のプレビュー表示 |

### diagnosis_name（診断名マスタ）

| フィールド | 入力部品 | バリデーション | 必須 | エラーメッセージ |
|---|---|---|---|---|
| 診断カテゴリ | `Select`（diagnosis_category マスタ連動） | - | ✅ | 「診断カテゴリを選択してください」 |

**特記:**
- `MasterLink` 付き
- 未登録時は警告メッセージ表示

### diagnosis_category（診断カテゴリマスタ）

**固有セクションなし**
共通フィールド（コード、名称、カテゴリ、[単価(税込)]）のみ

### inquiry_template（問診定型文マスタ）

| フィールド | 入力部品 | バリデーション | 必須 | デフォルト | 選択肢 |
|---|---|---|---|---|---|
| 区分 | `Select` | - | ✅ | chief_complaint | 主訴/既往歴/現在の投薬/アレルギー情報/備考 |
| タイトル | `Input` | 100文字以内 | ✅ | - | - |
| 内容 | `Textarea` | 1000文字以内 | ✅ | - | - |
| 並び順 | `Input` (number) | 0以上の整数 | - | 0 | - |
| ステータス | `Select` | - | ✅ | active | 有効/無効 |

**バリデーション詳細:**
- タイトル: 必須、1〜100文字
- 内容: 必須、1〜1000文字
- 区分（category）: `chief_complaint` / `history` / `current_medications` / `allergy_info` / `notes`

### chief_complaint（主訴区分マスタ）

| フィールド | 入力部品 | バリデーション | 必須 | デフォルト | 選択肢 |
|---|---|---|---|---|---|
| 区分名 | `Input` | 100文字以内 | ✅ | - | - |
| コード | `Input` | 半角英数字・ハイフン・アンダーバー、50文字以内 | - | - | - |
| 並び順 | `Input` (number) | 0以上の整数 | - | 0 | - |
| ステータス | `Select` | - | ✅ | active | 有効/無効 |

**バリデーション詳細:**
- 区分名: 必須、1〜100文字
- コード: 任意、半角英数字・ハイフン・アンダーバーのみ（正規表現: `/^[a-zA-Z0-9_-]*$/`）、50文字以内

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
| `MasterItemFormSections` | `[C]` | カテゴリ固有フォームセクション（ディスパッチャー） |
| `ExaminationSection` | `[C]` | 検査マスタ固有: 検査項目リスト |
| `VaccineSection` | `[C]` | 予防接種マスタ固有: 対象種別・接種間隔 |
| `MedicineSection` | `[C]` | 薬剤マスタ固有: 剤形・単位 |
| `StaffSection` | `[C]` | スタッフマスタ固有: 役職・資格番号 |
| `CageSection` | `[C]` | ケージマスタ固有: タイプ・サイズ |
| `InsuranceSection` | `[C]` | 保険マスタ固有: 補償割合・電話 |
| `TrimmingCourseSection` | `[C]` | コースマスタ固有: 対象サイズ・所要時間 |
| `TrimmingOptionSection` | `[C]` | オプションマスタ固有: 追加時間・併用可否 |
| `HospitalizationSection` | `[C]` | 入院マスタ固有: 体格・料金単位 |
| `DiagnosisNameSection` | `[C]` | 診断名マスタ固有: 親カテゴリ選択 |
| `ConsultationSection` | `[C]` | 診察マスタ固有: 適用区分・標準診察時間 |
| `ProcedureSection` | `[C]` | 処置マスタ固有: 所要時間・麻酔要否 |
| `CheckupSection` | `[C]` | 定期健診マスタ固有: 推奨受診間隔・対象年齢 |
| `ServiceTypeSection` | `[C]` | 予約区分マスタ固有: 表示カラーピッカー・プレビュー |
| `SectionWrapper` | `[C]` | セクション共通ラッパー |
| `NotionPropertyRow` | `[S]` | Notion風プロパティ行 |
| `PageLayout` | `[S]` | ページレイアウト |
| `SearchFilterBar` | `[S]` | 検索フィルタバー |
| `DataTable` / `DataTableRow` | `[S]` | データテーブル |
| `PrimaryButton` | `[S]` | 新規登録・保存ボタン |
| `StatusBadge` | `[S]` | ステータスバッジ |
| `StaffImpactDialog` | `[S][M]` | スタッフ変更影響確認ダイアログ |
| `ConfirmDialog` | `[S][M]` | 削除確認 |
| `MasterLink` | `[S]` | マスタ設定リンク |
| `useMasterItemEditor` | `[H]` | CRUD操作フック |
| `useMasterItems` | `[H]` | マスタデータ取得 |

---

## データ型定義

```typescript
// マスタアイテム基本型
interface MasterItem {
  id: string;
  name: string;
  code?: string;
  category?: string;
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
type Anesthesia = "none" | "local" | "general";
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

// マスタフォームデータ
interface MasterFormData extends MasterItem {
  // カテゴリ固有フィールド
  examinationItems?: ExaminationItem[];
  species?: VaccineSpecies;
  interval?: string;
  dosageForm?: DosageForm;
  unit?: MedicineUnit;
  role?: string;
  licenseNumber?: string;
  clinics?: string[];
  email?: string;
  password?: string;
  userType?: UserType;
  cageType?: CageType;
  size?: CageSize;
  coverageRate?: CoverageRate;
  billingPhone?: string;
  targetSize?: TargetSize;
  duration?: number;
  additionalDuration?: number;
  combinable?: Combinable;
  bodySize?: BodySize;
  billingUnit?: BillingUnit;
  timeType?: ConsultationTime;
  standardDuration?: string;
  estimatedDuration?: string;
  anesthesia?: Anesthesia;
  recommendedInterval?: string;
  targetAge?: CheckupTargetAge;
  color?: string;
  diagnosisCategoryId?: string;
}

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
   - キーボードショートカット（Alt+矢印キー）で操作

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
- **MasterSelectModal**: 内アイテムリスト `<button>` 要素でキーボード操作対応
- **キーボードナビゲーション**: Tab, Shift+Tab, Enter, Esc, Alt+矢印キー
- **フォーカス管理**: モーダル開閉時、D&D操作後にフォーカス復帰
- **スクリーンリーダー対応**: `aria-live`, `aria-label`, `aria-describedby` 適切に設定
