# FE-150: MedicalRecordForm — バイタル記録ボタンと VitalsTab 保存ボタンに canEdit ガードなし

## 概要

`MedicalRecordForm.tsx` の「バイタル記録」ボタンに `canEdit` チェックがなく、閲覧のみユーザーでもクリックしてバイタル入力モーダルを開ける。さらにモーダル内の `VitalsTab` に「保存」ボタンがあり、閲覧のみユーザーでも実際に新しいバイタル記録を作成できてしまう。

## 影響範囲

- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx:425-435`
- `frontend/src/features/medical-records/components/VitalsModal.tsx`
- `frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx:215`
- 権限: `can_edit = false` のユーザー

## 現状の挙動（バグ）

```tsx
// MedicalRecordForm.tsx 425-435 — canEdit チェックなし
{activeTab !== "見積書" ? (
  <Button variant="outline" onClick={() => setIsVitalsOpen(true)} disabled={isNewRecord}>
    <HeartPulse className={ICON.action} /> バイタル記録
  </Button>
) : null}
```

1. 閲覧のみユーザーが既存カルテを開く
2. 「バイタル記録」ボタンが表示され、クリックするとモーダルが開く
3. VitalsTab 内の「保存」ボタンからバイタルデータを送信できる（API 403 が返るのみでUIガードなし）

## 期待する挙動

`canEdit` が false の場合、「バイタル記録」ボタン自体を非表示にする。

## 修正方針

```tsx
// MedicalRecordForm.tsx
const { canEdit } = usePermission("medical-records");

// バイタル記録ボタン:
{activeTab !== "見積書" && canEdit ? (
  <Button variant="outline" onClick={() => setIsVitalsOpen(true)} disabled={isNewRecord}>
    <HeartPulse className={ICON.action} /> バイタル記録
  </Button>
) : null}
```

VitalsTab の保存ボタンは canEdit を props で受け取るか、VitalsTab 内で `usePermission("medical-records")` を呼ぶことで対応。

## 優先度

HIGH — バックエンドが 403 を返すため実際のデータ変更は防げるが、ユーザーがエラーに遭遇する体験が悪い

## 関連

- `frontend/src/features/medical-records/routes/MedicalRecordForm.tsx`
- `frontend/src/features/medical-records/components/VitalsTab/VitalsTab.tsx`
- BUG-RBAC テスト 2026-04-07
