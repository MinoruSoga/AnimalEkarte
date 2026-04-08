# BUG-189: Frontend catch/onError で handleApiError 未呼び出し（26箇所）

| 項目 | 内容 |
|------|------|
| 優先度 | **High** |
| カテゴリ | エラーハンドリング |
| 影響箇所 | catch 9箇所 + onError 17箇所 = 26箇所 |

## 概要

プロジェクト規約で全 catch ブロックに `handleApiError` の呼び出しが必須だが、
`toast.error()` のみで済ませている箇所が 9 箇所ある。
また `useMutation` の `onError` コールバックで `handleApiError` を使わず
`toast.error()` のみの箇所が 17 箇所ある。

サーバーが返す具体的なエラー情報（409 Conflict、バリデーション詳細等）がユーザーに伝わらない。

## catch ブロック（9箇所）

| ファイル | 行 | 現状 |
|---------|-----|------|
| `features/owners/routes/OwnersList.tsx` | 313 | `toast.error("削除に失敗しました")` |
| `features/vaccinations/hooks/use-vaccination-form.ts` | 198 | `toast.error("保存に失敗しました")` |
| `features/examinations/hooks/use-examination-form.ts` | 128 | `toast.error("保存に失敗しました")` |
| `features/trimming/hooks/use-trimming-form.ts` | 174 | `toast.error("保存に失敗しました")` |
| `features/master/hooks/use-master-save.ts` | 68 | `toast.error("保存に失敗しました")` |
| `features/master/hooks/use-master-save.ts` | 85 | `toast.error("保存に失敗しました")` |
| `features/hospital-settings/routes/ClinicMasterSettings.tsx` | 207 | `toast.error("保存に失敗しました")` |
| `features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx` | 127 | 空 catch（コメントのみ） |
| `features/medical-records/components/MedicalRecordEstimate.tsx` | 127 | `toast.error("保存に失敗しました")` |

## onError コールバック（17箇所）

| ファイル | 行 |
|---------|-----|
| `features/master/routes/DiagnosisSettings.tsx` | 540, 551, 581, 593 |
| `features/master/routes/TrimmingSettings.tsx` | 573, 587, 616, 630 |
| `features/master/routes/MedicineSettings.tsx` | 668, 688 |
| `features/master/routes/TreatmentPlanMaster.tsx` | 723, 728 |
| `features/master/hooks/use-master-save.ts` | 72, 89 |
| `features/master/routes/ServiceTypeSettings.tsx` | 123 |
| `features/hospitalization/routes/HospitalizationForm.tsx` | 108 |
| `features/owners/routes/OwnerForm.tsx` | 529 |

## 修正パターン

```typescript
// catch — Before
} catch {
  toast.error("保存に失敗しました");
}
// catch — After
} catch (error) {
  handleApiError(error, "保存");
}

// onError — Before
onError: () => toast.error("更新に失敗しました"),
// onError — After
onError: (error) => handleApiError(error, "更新"),
```

### 参照実装: `features/owners/hooks/use-owner-form.ts:227-229`
```typescript
} catch (error) {
  handleApiError(error, "保存");
  return { ...prevState, success: false, timestamp: Date.now() };
}
```

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md`
> catch ブロックでは必ず handleApiError を呼び出す

## 優先度
**High** — master settings 系に集中。サーバーの 409 Conflict メッセージが握り潰される。
