# FE-010: 医師選択ボタン - 実装未完成

## 問題

カルテ編集ページ上部の医師選択ボタン「田中 太郎」をクリックしても、医師選択UI（モーダル/ドロップダウン）が表示されない。

---

## 根本原因

MedicalRecordForm.tsx のコールバック実装が不完全：

```typescript
// 行158
onStaffClick={() => setStaffName(staffName)}
```

このコールバックは医師を **変更せず**、現在の値を再設定しているだけ。医師選択UI がない。

---

## コンポーネント分析

### MedicalRecordForm.tsx

**現在の実装:**
```typescript
// 行64: staff 状態定義
const [staffName, setStaffName] = useState(() => user?.displayName ?? "");

// 行158: 医師選択ボタン
onStaffClick={() => setStaffName(staffName)}  // ❌ 何もしない
```

**問題:**
- 医師選択モーダル/ドロップダウンが実装されていない
- コールバックで医師リストの取得がない
- UI として医師を選択する手段がない

### PatientInfoCard.tsx

コンポーネントは正しく実装されている（ボタンクリック時に `onStaffClick` を呼び出す）。親から正しいコールバックが渡されていないのが問題。

---

## 修正対応

### オプション 1: モーダル実装（推奨）

1. `features/medical-records/components/StaffSelectionModal.tsx` を作成
2. スタッフリスト取得 API: `useGetStaffList()` hook
3. MedicalRecordForm で modal 状態管理
4. ボタンクリック時にモーダルを開く

**実装例:**
```typescript
// MedicalRecordForm.tsx
const [isStaffModalOpen, setIsStaffModalOpen] = useState(false);

<PatientInfoCard
  staffName={staffName}
  onStaffClick={() => setIsStaffModalOpen(true)}
/>

{isStaffModalOpen && (
  <StaffSelectionModal
    selectedStaff={staffName}
    onSelect={(newStaffName) => {
      setStaffName(newStaffName);
      setIsStaffModalOpen(false);
    }}
    onClose={() => setIsStaffModalOpen(false)}
  />
)}
```

### オプション 2: ドロップダウン実装

Popover + Select で医師リストを表示。

### オプション 3: ボタン削除

医師選択機能が不要な場合は、一時的にボタンを削除。

---

## テスト環境
- 記録 ID: 17
- テスト日時: 2026-03-16 13:15 JST

---

## 優先度
**🟡 MEDIUM** - ユーザー操作ブロック・医師情報が更新できない

---

## ブロッカー
- スタッフリスト取得 API の確認が必要
- 医師選択 UI の要件（モーダル vs ドロップダウン）確認が必要

