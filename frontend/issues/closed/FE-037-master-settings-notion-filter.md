# FE-037: マスタ設定13ページの NotionFilter 移行

**親タスク**: [TASK-009](../../docs/tasks/open/TASK-009-notion-filter-migration-and-size-fix.md)
**ステータス**: Open
**作成日**: 2026-03-18

---

## 概要

マスタ設定ページ13ページがNotionFilter未適用。現状は件数表示 + 検索アイコンのみで、「+ フィルタを追加」「並べ替え」ボタンがない。

## 現状（全マスタ設定ページ共通）

```
スタッフマスタ
├── 12件（件数のみ）
├── 検索アイコン（右上）Q
└── テーブル
```

## 変更後

```
スタッフマスタ
├── NotionFilter ツールバー
│   ├── 12件
│   ├── + フィルタを追加
│   ├── 並べ替え
│   └── 検索トグル
└── テーブル
```

## 対象ページと追加フィルタプロパティ

| ページ | フィルタプロパティ |
|--------|------------------|
| StaffSettings | ステータス(select: 有効/無効), 職種(select) |
| AnimalSpeciesSettings | ステータス(select: 有効/無効) |
| CageSettings | ステータス(select: 有効/無効) |
| TrimmingSettings | ステータス(select: 有効/無効) |
| MedicineSettings | ステータス(select: 有効/無効), 薬効分類(select) |
| DiagnosisSettings | （検索のみ） |
| HospitalizationSettings | ステータス(select: 有効/無効) |
| ServiceTypeSettings | ステータス(select: 有効/無効) |
| InterviewTemplateSettings | ステータス(select: 有効/無効) |
| ChiefComplaintSettings | ステータス(select: 有効/無効) |
| TreatmentPlanMaster | ステータス(select: 有効/無効) |
| InsuranceSettings | ステータス(select: 有効/無効) |
| JobTitleSettings | ステータス(select: 有効/無効) |

## 対象ファイル

- `frontend/src/features/master/routes/StaffSettings.tsx`
- `frontend/src/features/master/routes/AnimalSpeciesSettings.tsx`
- `frontend/src/features/master/routes/CageSettings.tsx`
- `frontend/src/features/master/routes/TrimmingSettings.tsx`
- `frontend/src/features/master/routes/MedicineSettings.tsx`
- `frontend/src/features/master/routes/DiagnosisSettings.tsx`
- `frontend/src/features/master/routes/HospitalizationSettings.tsx`
- `frontend/src/features/master/routes/ServiceTypeSettings.tsx`
- `frontend/src/features/master/routes/InterviewTemplateSettings.tsx`
- `frontend/src/features/master/routes/ChiefComplaintSettings.tsx`
- `frontend/src/features/master/routes/TreatmentPlanMaster.tsx`
- `frontend/src/features/master/routes/InsuranceSettings.tsx`
- `frontend/src/features/master/routes/JobTitleSettings.tsx`

## 実装方針

- `MasterListPage` コンポーネント（TASK-007/008で作成済み）にNotionFilterを統合するのが理想的
- 各ページで個別に `filterProperties` を定義し、`MasterListPage` に渡すパターン
- TASK-008（MasterCRUDPage高レベルラッパー）との統合を考慮

## 受入条件

- [ ] 全13ページにNotionFilterツールバー表示
- [ ] 各ページでステータスフィルタが動作
- [ ] 並べ替え機能が動作
- [ ] 検索トグルで検索入力が展開/折りたたみ
- [ ] 既存のCRUD操作（作成/編集/削除/並べ替え）に影響なし
- [ ] `docker compose exec frontend npm run lint` エラーなし
- [ ] `docker compose exec frontend npm run build` 成功
