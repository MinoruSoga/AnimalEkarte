# BUG-160: ワクチン管理・検査管理の取込ボタンが view-only ユーザーに表示される

## 概要
`/vaccinations` の「データ取込」ボタンと `/examinations` の「検査データ取込」ボタンが、
create 権限のない（view-only）ユーザーに表示される。
取込操作は create と同義であり、`canCreate` ガードの対象とすべきだった。

## 再現手順
1. view-only 権限グループのユーザー（can_view=true, can_create=false）でログイン
2. `/vaccinations` にアクセス
3. **結果**: ページ右上に「データ取込」ボタンが表示される
4. `/examinations` にアクセス
5. **結果**: ページ右上に「検査データ取込」ボタンが表示される

## 期待する動作
- `can_create = false` のユーザーには「データ取込」「検査データ取込」ボタンを非表示にする

## 現状コード（修正前）

### `frontend/src/features/vaccinations/routes/VaccinationList.tsx:228-237`
```tsx
// ❌ canCreate チェックなし — 常に表示される
<Button variant="outline" className="h-10 text-base gap-2 bg-white" onClick={() => {}}>
  <FileSpreadsheet className={ICON.action} />
  データ取込
</Button>
{canCreate ? (
  <PrimaryButton onClick={handleCreate}>
    <Plus className={`mr-1.5 ${ICON.action}`} />
    新規登録
  </PrimaryButton>
) : null}
```

### `frontend/src/features/examinations/routes/ExaminationsList.tsx:249-258`
```tsx
// ❌ canCreate チェックなし — 常に表示される
<Button variant="outline" className="h-10 text-sm gap-2 bg-white" onClick={() => {}}>
  <FileSpreadsheet className={ICON.action} />
  検査データ取込
</Button>
{canCreate ? (
  <PrimaryButton onClick={handleCreate}>
    <Plus className={`mr-1.5 ${ICON.action}`} />
    新規検査登録
  </PrimaryButton>
) : null}
```

### 比較: 正しい実装（プロジェクト内参照実装）
```tsx
// ✅ BUG-156/157/158 で修正済みの各ページ — canCreate でガード
{canCreate ? (
  <Button ...>新規登録</Button>
) : null}
```

## 影響範囲

| 対象 | 詳細 | 状態 |
|------|------|------|
| `frontend/src/features/vaccinations/routes/VaccinationList.tsx:228` | データ取込ボタン | **修正済み** |
| `frontend/src/features/examinations/routes/ExaminationsList.tsx:249` | 検査データ取込ボタン | **修正済み** |

## 修正方針

### 1. `VaccinationList.tsx` — canCreate ガードで囲む
```tsx
{canCreate ? (
  <Button variant="outline" className="h-10 text-base gap-2 bg-white" onClick={() => {}}>
    <FileSpreadsheet className={ICON.action} />
    データ取込
  </Button>
) : null}
{canCreate ? (
  <PrimaryButton onClick={handleCreate}>
    <Plus className={`mr-1.5 ${ICON.action}`} />
    新規登録
  </PrimaryButton>
) : null}
```

### 2. `ExaminationsList.tsx` — canCreate ガードで囲む
```tsx
{canCreate ? (
  <Button variant="outline" className="h-10 text-sm gap-2 bg-white" onClick={() => {}}>
    <FileSpreadsheet className={ICON.action} />
    検査データ取込
  </Button>
) : null}
{canCreate ? (
  <PrimaryButton onClick={handleCreate}>
    <Plus className={`mr-1.5 ${ICON.action}`} />
    新規検査登録
  </PrimaryButton>
) : null}
```

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md` — フロントエンドベストプラクティス
> **Conditional Render**: 必ず `? (...) : null`（`&&` 禁止）

`canCreate` を条件として三項演算子で表示制御する。

### プロジェクト内参照実装
- `frontend/src/features/vaccinations/routes/VaccinationList.tsx:232` — `{canCreate ? (<PrimaryButton...>) : null}` が正しいパターン（修正前から存在）
- `frontend/src/features/medical-records/` — 全ボタンが権限チェック済み

## 優先度
**Medium** — セキュリティ実害は軽微（ボタン onClick は未実装・no-op）だが、RBAC の一貫性違反

## 関連チケット
- BUG-156: 薬剤・診断・ケージマスタの作成ボタンをcreate権限で制御
- BUG-157: 薬剤マスタのインライン追加ボタンをcreate権限で制御
- BUG-158: サイドパネルの保存・削除ボタンをview-onlyで非表示化

## 関連ファイル
- `frontend/src/features/vaccinations/routes/VaccinationList.tsx:228` — データ取込ボタン
- `frontend/src/features/examinations/routes/ExaminationsList.tsx:249` — 検査データ取込ボタン
- `frontend/src/features/auth/index.ts` — usePermission hook
