# FE-002: 医療記録作成API - リクエスト契約更新確認

**Status**: Open
**Priority**: High
**Affects**: Medical Record creation feature
**Date Created**: 2026-03-16
**Related**: BE-015

## 問題

BE-015 で Medical Record 作成API のリクエスト契約が変更されました。
FE側の `CreateMedicalRecordRequest` がこの新しいコントラクトに対応しているか確認が必要です。

## BE-015 での変更内容

### 新規フィールド追加
- `visit_date` (string, "YYYY-MM-DD"形式) — FE送信フィールド
- `visit_type` (string) — FE送信フィールド（BEで無視）
- `record_no` (optional) — 自動生成に対応（送信不要）
- `chief_complaint` (optional) — 問診情報
- `chief_complaint_category_id` (optional)
- `plan` (optional) — 臨床計画
- `assessment` (optional)
- `diagnosis_1_category_id`, `diagnosis_1_name_id`
- `diagnosis_2_category_id`, `diagnosis_2_name_id`

### フィールド型変更
- `date` → `visit_date` に変更（またはdateを廃止）
- ID フィールド: string 型に統一（BE側で uint64 に変換）

## 確認項目

### ファイル: `frontend/src/features/medical-records/api/types.ts`

1. `CreateMedicalRecordRequest` インターフェースが新しいフィールドに対応しているか
2. 以下フィールドの型が正しいか：
   - `visit_date: string` (YYYY-MM-DD)
   - `visit_type: string`
   - `chief_complaint?: string`
   - `plan?: string`
   - `assessment?: string`
   - Diagnosis フィールド（複数対応）

### ファイル: `frontend/src/features/medical-records/hooks/useMedicalRecordForm.ts`

1. フォーム送信時に新フィールドを正しく構築しているか
2. `visit_date` の形式変換（JS Date → "YYYY-MM-DD"）が実装されているか
3. `record_no` の自動生成を期待していないか（BE側で自動生成されるため）

## 修正対応

### 方法1: 型定義を更新
```typescript
// types.ts
export type CreateMedicalRecordRequest = Omit<
  BackendMedicalRecord,
  'id' | 'clinic_id' | 'created_at' | 'updated_at' | 'record_no'
> & {
  visit_date?: string;  // "YYYY-MM-DD"
  visit_type?: string;
  chief_complaint?: string;
  plan?: string;
  assessment?: string;
  // ... diagnosis fields
};
```

### 方法2: フォームコンポーネント対応
- `visit_date` フィールドを追加（DatePicker）
- 送信前に形式を "YYYY-MM-DD" に変換
- `record_no` フィールドを削除（自動生成のため）

## 優先度

**🔴 HIGH** - 医療記録作成機能が完全に動作しないため

## ブロッカー

BE-015 が実装完了したため、FE側の対応が必須です。
