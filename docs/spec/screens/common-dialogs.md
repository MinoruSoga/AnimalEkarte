# 共通ダイアログ・共有コンポーネント 仕様書 (Shared UI Components)

## 概要
本システムは、全画面での UX の一貫性と、医療ミスを防ぐための「安全機能」を担保するため、高度に共通化されたコンポーネントライブラリを使用しています。

---

## 1. 臨床安全・ナビゲーション

### 1.1 患者情報カード (`PatientInfoCard`)
トリミング・定期健診・検査・入院・予防接種の各フォーム画面上部に常駐する「単一の真実」（`features/trimming`, `features/checkups`, `features/examinations`, `features/hospitalization`, `features/vaccinations`）。
- **臨床アラート**: 死亡 (`deceased`) ステータスを【死亡】バッジで強調。
- **属性表示**: 名前、年齢、性別、最新体重、担当医、保険、次回来院予定。`petDetails` は `formatPatientPetDetails` が組み立てる。`species` があるときは先頭に実データを出す。欠損の年齢・性別・去勢避妊だけ「不明」。独立登録（予防接種・定期健診・入院新規）も同じ関数を使う。
- **クイックアクション**: `onOwnerClick` / `onStaffClick` 等のクリックコールバックを任意で受け取る。現状 `onOwnerClick` を渡す利用画面はない。担当者変更はトリミングフォームの担当スタッフ選択と、入院登録・編集フォームの担当医選択で使用する（入院は `canSubmit` 時のみ）。飼主付け替えの `OwnerSearchModal` 起動はカルテ画面側（下記例外）の導線であり、本カード経由の実装は存在しない。
- **例外**: カルテ画面（06-medical-records-form.md）は `PatientInfoCard` を使わず、専用の `MedicalRecordStickyHeader` が共有 `PatientContextHeader` を組み込んで同等の飼主（`OwnerSearchModal` 起動）/担当医クリック導線を実装している。

### 1.2 離脱防止ガード (`NavigationBlocker`)
React Router 7 の `useBlocker` を活用し、入力データの損失を物理的に防ぎます。
- **トリガー**: フォーム変更（`isDirty`）がある状態でのページ離脱。
- **挙動**: SPA内遷移はカスタム`ConfirmDialog`（`NavigationBlocker`自身）で確認。タブを閉じる/リロード等のブラウザレベル離脱はネイティブの `beforeunload` ダイアログで確認（こちらは併用する `useUnsavedChanges` フックが担当し、`NavigationBlocker` 単体の機能ではない）。

---

## 2. 検索・選択モーダル (Modals)

### 2.1 診療項目検索 (`TreatmentSearchDialog`)
数千項目のマスタから目的の処置を瞬時に特定します。
- **フィルタ**: 診察、検査、処置、予防、入院、薬剤のカテゴリチップ切り替え。
- **検索**: インクリメンタルなキーワード検索。

### 2.2 飼主検索・付け替え (`OwnerSearchModal`)
既存カルテの飼主を誤って登録した場合や、譲渡時の変更に使用します。
- **安全確認**: `OwnerSearchModal` 自身は選択のたびに汎用の「飼主変更の確認」ダイアログ（「飼主を「A」→「B」に変更します。よろしいですか？」）を無条件で表示する（値引率・会員区分の差異は判定しない）。会員区分・値引率が異なる場合の金額変動警告（BUG-373）は呼び出し元（`OwnerForm.tsx` の `handlePetChangeOwner`）側の別ロジックであり、`OwnerSearchModal` の機能ではない。

### 2.3 担当医選択 (`StaffSelectionModal`)
稼働中（在職）スタッフを職種別にグルーピングし、名前検索で臨床担当者を割り当てます。
- **注**: 実装は `features/medical-records/components/StaffSelectionModal.tsx` にあり、現状カルテ画面（`MedicalRecordForm`）専用。他画面から共有利用されているコンポーネントではない。

---

## 3. 入力・フィルタリング補助

### 3.1 属性入力 (`PropertyInput`)
Notion の操作感を踏襲したボーダーレス入力。
- **特徴**: 通常時はテキスト、フォーカスで入力枠が出現。画面を煩雑にせず、かつ高速な編集を可能にします。

### 3.2 高機能フィルタ (`PropertyFilter`)
- **動的条件**: ユーザーがその場で「かつ/または」条件を組み合わせてデータを抽出。
- **遅延描画**: `useDeferredValue` は `PropertyFilter` 自身ではなく利用側の一覧画面が検索キーワードへ適用し、大規模データのフィルタ中もタイピングの遅延が発生しません（遅延中の視覚フィードバックは共有 `FilteringIndicator` が担当）。

---

## 4. 技術仕様 (Common Standards)

- **アクセシビリティ**: 全てのモーダルは WAI-ARIA 準拠。`ESC` キーでの終了、フォーカストラップを完備。
- **パフォーマンス**: 頻繁に再描画される親画面の影響を避けるため、主要なモーダル・カード系部品（`PatientInfoCard`, `OwnerSearchModal`, `TreatmentSearchDialog`, `StaffSelectionModal`, `PropertyFilter` 等）は `memo()` 化されています。ただし本ページで扱う `PropertyInput` と `NavigationBlocker` は非 `memo()`。

---
