# 電子カルテ タブ詳細仕様補足

本ドキュメントは、`/SCREENS.md`の「4.3 カルテ入力/編集」の詳細仕様を補足します。

---

## Tab 2: 診察/治療プラン（`MedicalRecordDiagnosisPlan`）

### レイアウト構成

- **上部**: バイタル入力セクション（`VitalInputDialog`トリガーボタン＋バイタル履歴グラフ）
- **中部**: 診断ヘッダー（`DiagnosisHeader`、3カラム）
- **下部**: 治療プランテーブル＋集計

### 診断ヘッダー（`DiagnosisHeader`）

| カラム | コンポーネント | フィールド | 入力部品 | 備考 |
|---|---|---|---|---|
| 左 | `DiagnosisHeaderChiefComplaint` | 主訴 | 読み取り専用テキスト | 問診タブで入力した内容を表示 |
| 中央 | `DiagnosisHeaderPhysicalExam` | 身体所見 | Markdown Textarea | |
| 右 | `DiagnosisHeaderDiagnosis` | 診断詳細 | Markdown Textarea | |
| 右 | | 診断1 カテゴリ | Combobox | diagnosis_categoryマスタから選択 |
| 右 | | 診断1 診断名 | Combobox | diagnosis_nameマスタから選択、カテゴリでフィルタ |

### 治療プランテーブル（`TreatmentTable`）

#### データソース
- `formData.items`（TreatmentItem[]）

#### 表示カラム
種別、治療内容、メモ、保険、単価(税込)、数量、割引(%)、値引(￥)、小計、操作

#### 種別カラム詳細
- **表示**: 読み取り専用テキスト（`w-20`, `text-xs`, `text-slate-600`）
- **データ**: `item.type`（"診察" | "処置" | "薬剤" | "検査" | "予防接種" | "定期健診" | undefined）
- **設定方法**: マスタ検索ダイアログで項目選択時に自動設定（`getCategoryTreatmentType`関数使用）
- **未設定時**: 空欄表示（手動追加した空行の場合）

#### マスタ検索連携
- **トリガー**: テーブル下部の「+ 治療項目を追加」ボタン
- **ダイアログ**: `TreatmentSearchDialog`（`Command`コンポーネントベース）
- **検索対象マスタ**: 以下の条件を満たす項目
  - `status === "active"`（有効項目）
  - `price != null && price > 0`（価格設定済み）
  - `category in ["consultation", "examination", "procedure", "vaccine", "checkup", "hospitalization", "medicine"]`
- **選択時の処理フロー**:
  1. マスタアイテムの`category`フィールドを取得（英語キー）
  2. `getCategoryTreatmentType(category)`で種別を判定
     - `consultation` → "診察"
     - `procedure` → "処置"
     - `medicine` → "薬剤"
     - `examination` → "検査"
     - `vaccine` → "予防接種"
     - `checkup` → "定期健診"
     - `hospitalization` → undefined（種別なし）
  3. 新規`TreatmentItem`を作成:
     ```typescript
     {
       id: String(Date.now()),
       type: getCategoryTreatmentType(item.category),
       content: item.name,
       unitPrice: item.unitPrice,
       memo: "",
       insurance: true,
       quantity: 1,
       discountRate: 0,
       discountAmount: 0
     }
     ```
  4. `formData.items`配列に追加

#### 操作
- **各行の編集**: インライン編集（Input/NotionCheckbox）
- **削除**: `Trash2`アイコンボタン（確認なし即削除、ホバー時表示）
- **手動追加**: 「+ 行を追加」ボタン（空行挿入、種別なし）

### 集計（`TreatmentDetailedSummary`）

| 項目 | 表示 | 編集 | 備考 |
|---|---|---|---|
| 小計 | 読み取り専用 | - | 各行小計の合計（`Σ calcLineTotal`） |
| 割引率(%) | 入力可能 | Input（`useNumericInput`, 0-100範囲） | 全体割引率 |
| 値引額(￥) | 入力可能 | Input（`useNumericInput`） | 全体値引額 |
| 税込合計 | 読み取り専用 | - | `calcGrandTotal(小計, 割引率, 値引額)` |
| 消費税(内税10%) | 読み取り専用 | - | `extractInnerTax(合計, 10)` |

---

## Tab 3: 治療（`MedicalRecordTreatment`）

### レイアウト構成

- **上部**: 治療プランテーブル（未完了項目）＋集計
- **下部**: 治療済みテーブル（完了項目）＋集計

### 治療プランテーブル

#### データソース
- `planItems`（TreatmentItem[]）
- フィルタ: `status !== "完了"`

#### 表示カラム
[済]、種別、治療内容、メモ、保険、単価(税込)、数量、割引(%)、値引(￥)、小計、操作

#### 済チェックボックスカラム
- **幅**: `w-12`（48px）
- **部品**: `NotionCheckbox`（`#2383E2`アクセントブルー）
- **動作**: チェックON → `TreatmentMoveConfirmDialog`表示 → 確認後に`completedItems`へ移動
- **ラベル**: `<span className="sr-only">完了にする</span>`（スクリーンリーダー用、視覚的には非表示）
- **アクセシビリティ**: `role="checkbox"`, `aria-checked`属性

#### 種別カラム
診察/治療プランタブと同じ仕様（読み取り専用表示）

#### 操作
- **削除**: `Trash2`アイコン（確認なし即削除）
- **完了への移動フロー**:
  1. 済チェックボックスをON
  2. `TreatmentMoveConfirmDialog`で確認
     - タイトル: "治療済みに移動"
     - 説明: "[治療内容] を治療済みに移動しますか？"
     - ボタン: "キャンセル" / "移動する"
  3. 「移動する」クリック後:
     - `planItems`から該当項目を削除
     - `completedItems`に追加（`status: "完了"`を設定）
     - スクリーンリーダー通知（`useAnnounce`）: "治療済みに移動しました"
     - 治療済みテーブルの見出し（`h3`）にフォーカス移動（`completedHeadingRef.current?.focus()`）

### 治療済みテーブル

#### データソース
- `completedItems`（TreatmentItem[]）
- フィルタ: `status === "完了"`

#### 表示カラム
[戻す]、種別、治療内容、メモ、保険、単価(税込)、数量、割引(%)、値引(￥)、小計、操作

#### 戻すボタンカラム
- **幅**: `w-20`（80px）
- **部品**: `Button`（`variant="ghost"`, `size="sm"`、`text-teal-700`）
- **ラベル**: "プランに戻す"
- **アイコン**: `Undo2`（size-4）
- **動作**: クリック → `TreatmentMoveConfirmDialog`表示 → 確認後に`planItems`へ移動

#### 種別カラム
診察/治療プランタブと同じ仕様（読み取り専用表示）

#### 操作
- **削除**: なし（治療済み項目は削除不可、会計連携のため保持）
- **プランへの戻しフロー**:
  1. 「プランに戻す」ボタンクリック
  2. `TreatmentMoveConfirmDialog`で確認
     - タイトル: "治療プランに戻す"
     - 説明: "[治療内容] を治療プランに戻しますか？"
     - ボタン: "キャンセル" / "戻す"
  3. 「戻す」クリック後:
     - `completedItems`から該当項目を削除
     - `planItems`に追加（`status: "未完了"`を設定）
     - スクリーンリーダー通知: "治療プランに戻しました"
     - 治療プランテーブルの見出し（`h3`）にフォーカス移動（`planHeadingRef.current?.focus()`）

### 移動確認ダイアログ（`TreatmentMoveConfirmDialog`）

#### 表示条件
`pendingMove`ステートが`null`でない

#### Props

| Prop | 型 | 説明 |
|---|---|---|
| `open` | `boolean` | `pendingMove !== null` |
| `title` | `string` | "治療済みに移動" または "治療プランに戻す" |
| `description` | `string` | "[治療内容] を[移動先]に移動しますか？" |
| `onConfirm` | `() => void` | 移動処理実行＋フォーカス移動＋スクリーンリーダー通知 |
| `onCancel` | `() => void` | `pendingMove`をnullにリセット |

#### 内部構造
- **ベース**: `AlertDialog`コンポーネント
- **ボタン**: キャンセル（`variant="outline"`） / 確認（`variant="default"`）

### 集計（両テーブル）

- **治療プラン用集計**: `planItems`の小計・割引・合計を表示・編集可能
- **治療済み用集計**: `completedItems`の小計・割引・合計を表示（読み取り専用）

---

## Tab 7: 見積書（`MedicalRecordEstimate`）

### レイアウト構成

- **上部**: 件名入力フォーム
- **中部**: 明細テーブル（`TreatmentTable`）＋集計
- **下部**: コメント・備考入力（2カラム）＋PDF出力ボタン

### 件名入力（`EstimateForm`）
- **フィールド**: `subject`
- **部品**: `Input`
- **プレースホルダー**: "見積書の件名を入力"

### 明細テーブル（`TreatmentTable`）

#### データソース
- `formData.items`（診察/治療プランタブと共有）

#### 表示カラム
種別、治療内容、メモ、保険、単価(税込)、数量、割引(%)、値引(￥)、小計、操作

#### 特記事項
- **済チェックボックス**: なし
- **選択チェックボックス**: 表示（将来拡張用、現在は機能なし）
- **戻すボタン**: なし
- **削除ボタン**: あり
- **種別カラム**: 診察/治療プランタブと同じ（読み取り専用、マスタ選択時に自動設定）

### 集計
診察/治療プランタブと同じ（`TreatmentDetailedSummary`）

### コメント・備考
- **レイアウト**: 2カラム（`grid grid-cols-1 lg:grid-cols-2 gap-4`）
- **左**: コメント（`Textarea`、`rows={4}`）
- **右**: 備考（`Textarea`、`rows={4}`）

### アクション
- **PDF出力ボタン**: `PrintPreviewDialog`トリガー（確定済みカルテのみ有効）

---

## Tab 8: 会計(医師確認)（`MedicalRecordBillCheck`）

### レイアウト構成

- **上部**: 明細テーブル（`TreatmentTable`、読み取り専用）＋集計
- **下部**: 固定フローティングアクション（チェック完了・会計へ進む）

### 明細テーブル（`TreatmentTable`）

#### データソース
- `completedItems`（治療タブの治療済みテーブルと自動同期）

#### 表示カラム
種別、治療内容、メモ、保険、単価(税込)、数量、割引(%)、値引(￥)、小計

#### 特記事項
- **済チェックボックス**: なし
- **選択チェックボックス**: なし
- **操作カラム**: なし（完全読み取り専用）
- **種別カラム**: 診察/治療プランタブと同じ（読み取り専用表示）
- **編集**: 全カラム編集不可

### 集計
- 治療済み項目の小計・割引・合計を表示（読み取り専用）

### 固定フローティングアクション

#### チェック完了ボタン
- **初期表示**: "チェック完了"（`CheckCircle2`アイコン、`variant="outline"`）
- **クリック後**: "未チェックに戻す"（`X`アイコン、`variant="outline"`）
- **動作**: トグル動作、クリックごとにステータス反転
- **トースト通知**: 
  - チェック完了時: "医師確認が完了しました"
  - 未チェックに戻す時: "医師確認を取り消しました"

#### 会計へ進むボタン / 会計を確認ボタン
- **新規カルテ（linkedAccountingId なし）**: 
  - ラベル: "会計へ進む"（`Receipt`アイコン）
  - `variant`: "default"（`bg-[#2383E2]`プライマリブルー）
  - disabled条件: `items.length === 0`
  - クリック時: 会計新規作成画面へ遷移（`/accounting/new?medicalRecordId=xxx`）
- **既存会計連携あり（linkedAccountingId あり）**:
  - ラベル: "会計を確認"（`ExternalLink`アイコン）
  - `variant`: "outline"
  - クリック時: 会計詳細画面へ遷移（`/accounting/:linkedAccountingId`）

---

## デモデータ修正内容

### 問題点
- デモデータの`TreatmentItem.category`フィールドが日本語（"診察"、"処置"等）になっていた
- マスタの`category`は英語キー（consultation, procedure等）のため、マッチせず項目が表示されなかった

### 修正内容
1. **デモデータの`category`を削除**: 
   - `TreatmentItem`型に`category`フィールドは存在しない（`type`フィールドが正しい）
2. **`type`フィールドに日本語ラベルを設定**:
   - 既存データ: `type: "診察"`, `type: "処置"`, `type: "薬剤"` 等
3. **マスタ連携時の自動設定ロジック追加**:
   - `MedicalRecordDiagnosisPlan.tsx`の`handleSelectTreatment`で`getCategoryTreatmentType`を呼び出し
   - `MedicalRecordTreatment.tsx`の`handleSelectTreatment`で同様に処理

### getCategoryTreatmentType関数

```typescript
function getCategoryTreatmentType(category: string): TreatmentType | undefined {
  const mapping: Record<string, TreatmentType> = {
    consultation: "診察",
    procedure: "処置",
    medicine: "薬剤",
    examination: "検査",
    vaccine: "予防接種",
    checkup: "定期健診",
  };
  return mapping[category];
}
```

### マスタカテゴリ一覧（英語キー）

| 英語キー | 日本語ラベル | 種別への変換 |
|---|---|---|
| `consultation` | 診察 | "診察" |
| `procedure` | 処置 | "処置" |
| `medicine` | 薬剤 | "薬剤" |
| `examination` | 検査 | "検査" |
| `vaccine` | 予防接種 | "予防接種" |
| `checkup` | 定期健診 | "定期健診" |
| `hospitalization` | 入院 | undefined（種別なし、入院は治療項目外） |

---

## まとめ

### 種別カラムの追加による変更点

1. **TreatmentTable**に「種別」カラムを追加（全タブ共通）
2. **種別の自動設定**: マスタ検索→項目選択時に`getCategoryTreatmentType`で自動判定
3. **デモデータ修正**: `category`フィールド削除、`type`フィールドに日本語ラベル設定
4. **マスタカテゴリ**: 英語キー（consultation, procedure等）で統一
5. **表示**: 読み取り専用テキスト（`w-20`, `text-xs`, `text-slate-600`）

### 影響範囲

| ファイル | 変更内容 |
|---|---|
| `MedicalRecordDiagnosisPlan.tsx` | `handleSelectTreatment`に`getCategoryTreatmentType`追加 |
| `MedicalRecordTreatment.tsx` | `handleSelectTreatment`に`getCategoryTreatmentType`追加 |
| `TreatmentTable.tsx` | 種別カラムの表示追加（`item.type`を表示） |
| `/features/medical-records/api/mockData.ts` | デモデータの`category`削除、`type`設定 |

### 今後の拡張可能性

- 見積書タブの選択チェックボックス機能実装（現在は表示のみ）
- 種別ごとの集計機能（診察費合計、薬剤費合計等）
- 種別フィルタ機能（特定種別のみ表示）
