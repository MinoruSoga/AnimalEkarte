# FE-006: 主訴区分（Chief Complaint Category）がハードコード・マスタデータ非連動

## 問題
医療記録（カルテ）の「問診」タブで「主訴区分」（Chief Complaint Category）を選択しても、データが保存されない。原因は、コンボボックスが：
1. **ハードコード状態** - DB のマスタデータを読み込まず、3 つの固定値のみ表示
2. **状態管理がない** - 選択値をコンポーネント内で保持していない
3. **API 送信時に含まれない** - フォーム送信時に主訴区分が送信されない

## 詳細分析

### コンポーネント構造（InterviewChiefComplaint.tsx）
```typescript
// ❌ 問題: hardcoded SelectItem
<Select>
  <SelectTrigger>
    <SelectValue placeholder="選択してください" />
  </SelectTrigger>
  <SelectContent>
    <SelectItem value="checkup">定期検診</SelectItem>
    <SelectItem value="sick">傷病</SelectItem>
    <SelectItem value="prevention">予防</SelectItem>
  </SelectContent>
</Select>
```

**問題点:**
- props に `chiefComplaintCategoryId`（選択値）がない
- props に `setChiefComplaintCategoryId`（setter）がない
- Select に `value` binding がない
- Select に `onValueChange` handler がない
- マスタデータ API 呼び出しがない

### 親コンポーネント（MedicalRecordInterview.tsx）
```typescript
// ❌ 問題: カテゴリ state がない
interface MedicalRecordInterviewProps {
  chiefComplaint: string;           // ✅ テキスト state あり
  setChiefComplaint: ...
  treatmentPolicy: string;          // ✅ テキスト state あり
  setTreatmentPolicy: ...
  // ❌ 以下が欠落:
  // chiefComplaintCategoryId?: number | null;
  // setChiefComplaintCategoryId?: (id: number | null) => void;
}
```

### ネットワーク検証
ブラウザ DevTools で確認:
- **API requests**: `GET /api/v1/me` × 2, `GET /api/v1/medical-records/17`, `GET /api/v1/pets/15` のみ
- **マスタデータ取得**: なし（FETCH/XHR で master 関連エンドポイント呼び出しなし）
- **DB 側**: chief_complaint_categories テーブルに 6 件のマスタデータが存在（確認済み）

## 根本原因
UI は「マスタデータを動的に読み込む機能」を実装していない：

1. **フロントエンド設計の欠損**
   - InterviewChiefComplaint に主訴区分の state がない
   - MedicalRecordInterview に chiefComplaintCategoryId の状態管理がない
   - API フック（useGetChiefComplaintCategories など）がない

2. **コンポーネント設計**
   - InterviewChiefComplaint は「主訴詳細テキスト」のみを受け取り
   - Select が hardcoded で、マスタデータ props がない
   - onValueChange handler がない

3. **API 連携**
   - backend には `GET /api/v1/medical-records/:id/chief-complaint-categories` などのエンドポイントがあるかも不明
   - frontend が master data を fetch していない

## テスト環境
- 記録 ID: 17
- ペット ID: 15
- テスト日時: 2026-03-16 11:40 JST

### 再現手順
1. カルテ編集画面を開く（ID 17）
2. 「問診」タブをクリック
3. 「主訴区分」コンボボックスを確認
   - **期待**: DB の chief_complaint_categories から 6 つのオプション（食欲不振、嘔吐・下痢、皮膚・被毛異常、耳症状、眼症状、その他）
   - **実際**: 3 つの固定オプション（定期検診、傷病、予防）
4. いずれかのオプションを選択して保存
   - **期待**: 選択値が DB に保存される
   - **実際**: 選択値は無視され、保存されない

## 修正対応
### フロントエンド（優先度: **HIGH**）
1. **state 追加** (MedicalRecordInterview.tsx)
   - `chiefComplaintCategoryId: number | null`
   - `setChiefComplaintCategoryId: (id: number | null) => void`
   - InterviewChiefComplaint に props として渡す

2. **コンポーネント修正** (InterviewChiefComplaint.tsx)
   - マスタデータ API hook (`useGetChiefComplaintCategories` など) を追加
   - Select に `value` と `onValueChange` binding
   - 動的オプション生成 (`.map()`)
   - API 로딩 중 Skeleton/Spinner 표시

3. **API 整備** (features/medical-records/api/get-chief-complaint-categories.ts)
   - `useGetChiefComplaintCategories()` hook 作成
   - API endpoint: `GET /api/v1/chief-complaint-categories` or `GET /api/v1/master/chief-complaint-categories`

4. **フォーム送信** (useMedicalRecordForm.ts)
   - `chiefComplaintCategoryId` を form payload に含める
   - backend API call 時に送信

### バックエンド
- **確認項目**:
  1. Master data API endpoint が存在するか確認
  2. exists: `/api/v1/chief-complaint-categories` or similar
  3. Response schema 確認
  4. Medical record PATCH/POST に `chief_complaint_category_id` フィールドがあるか確認

## ブロッカー
- BE-003, BE-004, BE-005 の根本原因は **もしかして** これと関連しているかもしれない（state 欠損 → form payload 送信不完全）
- 「保存成功」UI 表示 → 実際は 4xx/5xx エラー返却 → frontend がエラー握りつぶし？
- Network 調査必要

