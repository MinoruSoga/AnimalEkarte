# BUG-380: サイドパネル編集中に他行操作 / タブ切替で未保存変更がサイレント破棄される

**作成日**: 2026-04-15
**Status**: CLOSED — 全対象画面に展開完了
**Priority**: MEDIUM (UX・データ消失リスク)
**Affects**: 全マスタ設定画面 (共通 `SidePeekPanel` / `MasterSidePanel` + マスタのタブ UI)

---

## 概要

マスタ設定画面のサイドパネルで入力中（未保存）に以下を行うと**警告なしに入力内容が破棄される**。

1. 一覧の別行の「操作」アイコンクリック → その行の編集モードに切替（入力中データ消失）
2. マスタ内のタブ切替 → サイドパネル自動クローズ（入力中データ消失）
3. 「+ 新規登録」クリック → 現在の編集が破棄される
4. ブラウザタブクローズ / リロード → 警告なしで離脱

## 実装

### 1. 共通フック — `frontend/src/hooks/use-side-peek-dirty.ts`

```ts
useSidePeekDirty(): {
  isDirty: boolean;
  markDirty(): void;
  markClean(): void;
  confirmDiscard(): boolean;  // dirty なら window.confirm、OK で true
}
```

- dirty 時に `beforeunload` イベントを自動登録（タブクローズ警告）
- `confirmDiscard()` が true を返したら呼び出し元は遷移続行、false なら中止

### 2. `useMasterCRUD` 拡張

```ts
useMasterCRUD<T>({ ..., dirtyGuard?: { confirmDiscard: () => boolean } })
```

`dirtyGuard` を指定すると `handleEdit` / `handleNew` / `handleClose` で破棄確認を挟む。

### 3. 適用画面 (全 14 画面)

| マスタ画面 | タブ有無 | 備考 |
|-----------|--------|------|
| `AnimalSpeciesSettings` | - | 先行実装 |
| `OccupationSettings` | - | |
| `InsuranceSettings` | - | |
| `InterviewTemplateSettings` | - | |
| `PermissionGroupSettings` | - | |
| `StaffSettings` | - | |
| `HospitalizationSettings` | - | |
| `CageSettings` | - | |
| `MerchandiseItemSettings` | - | |
| `ChiefComplaintSettings` | - | |
| `DiagnosisSettings` | あり (カテゴリ/病名) | 両 CRUD が dirtyGuard を共有 + タブ切替ガード |
| `TrimmingSettings` | あり (コース/オプション) | 両 CRUD が dirtyGuard を共有 + タブ切替ガード |
| `ReservationTypeSettings` | - (GroupPanel / CategoryPanel 切替) | カスタム構造。handleGroupEdit 等を全て confirmDiscard でラップ |
| `MedicineSettings` | - | カスタム formData 構造。`updateForm` で markDirty、save 成功で markClean |

### 4. サイドパネル側の変更

各 SidePanel に `onDirtyChange?: (dirty: boolean) => void` prop を追加し、
`useEffect(() => onDirtyChange?.(isDirty), [isDirty, onDirtyChange])` で親に伝搬。

### 5. タブ切替ガード

`DiagnosisSettings` と `TrimmingSettings` の `handleTabChange` 冒頭で
`if (!dirty.confirmDiscard()) return;` を挟む。

## 動作

### 別行クリック時
1. ユーザーが未保存で別行の編集アイコンをクリック
2. `handleEdit` が `confirmDiscard()` を呼び出し
3. 未保存あり → `window.confirm`「未保存の変更があります。破棄してよろしいですか?」
4. OK → 編集対象切替、Cancel → 現在の編集を保持

### タブ切替時 (Diagnosis / Trimming)
1. ユーザーが未保存でタブクリック
2. `handleTabChange` が `confirmDiscard()` を呼び出し
3. 未保存あり → 同じダイアログ
4. OK → タブ切替 + 両 CRUD の editTarget クリア、Cancel → タブそのまま

### タブクローズ / リロード時
1. `useSidePeekDirty` 内部の `beforeunload` リスナーがブラウザ標準警告を起動

## 完了条件

- [x] `useSidePeekDirty` フック作成
- [x] `useMasterCRUD` に `dirtyGuard` オプション追加
- [x] 全 14 マスタ画面に適用
- [x] タブ切替ガード (DiagnosisSettings / TrimmingSettings)
- [x] カスタム構造のマスタ (ReservationTypeSettings / MedicineSettings) へ適用
- [x] TypeScript 型チェック通過
- [x] ESLint エラー 0

## 参照コミット

- 初版 + AnimalSpeciesSettings: `7840b12a`
- 残 13 画面展開 + タブガード: (本コミット)
