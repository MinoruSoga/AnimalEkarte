# BUG-209: catch / onError で handleApiError 未呼び出し（26箇所）

| 項目 | 内容 |
|------|------|
| 優先度 | **High** |
| カテゴリ | エラーハンドリング |
| 影響 | サーバーの具体的なエラー（409 Conflict 等）がユーザーに伝わらない |

## 概要

プロジェクト規約では catch ブロックで `handleApiError` を呼ぶことが必須だが、
`toast.error("固定メッセージ")` のみで済ませている箇所が存在する。
catch ブロック 9箇所 + useMutation onError 17箇所 = 計26箇所。

## catch ブロック違反（9箇所 — grep で確認済み）

| ファイル | 行 | 現状 |
|---------|-----|------|
| `features/owners/routes/OwnersList.tsx` | 313 | `catch { toast.error("削除に失敗しました") }` |
| `features/vaccinations/hooks/use-vaccination-form.ts` | 198 | `catch { toast.error("保存に失敗しました") }` |
| `features/examinations/hooks/use-examination-form.ts` | 128 | `catch { toast.error("保存に失敗しました") }` |
| `features/trimming/hooks/use-trimming-form.ts` | 174 | `catch { toast.error("保存に失敗しました") }` |
| `features/master/hooks/use-master-save.ts` | 68 | `catch { toast.error("保存に失敗しました") }` |
| `features/master/hooks/use-master-save.ts` | 85 | `catch { toast.error("保存に失敗しました") }` |
| `features/hospital-settings/routes/ClinicMasterSettings.tsx` | 207 | `catch { toast.error("保存に失敗しました") }` |
| `features/shifts/components/ShiftFormDialog/ShiftFormDialog.tsx` | 127 | `catch { /* コメントのみ */ }` |
| `features/medical-records/components/MedicalRecordEstimate.tsx` | 127 | `catch { toast.error("保存に失敗しました") }` |

## useMutation onError 違反（17箇所 — grep で確認済み）

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
// catch ブロック
// Before
} catch {
  toast.error("保存に失敗しました");
}
// After
} catch (error) {
  handleApiError(error, "保存");
}

// onError
// Before
onError: () => toast.error("更新に失敗しました"),
// After
onError: (error) => handleApiError(error, "更新"),
```

## 参照実装

`features/owners/hooks/use-owner-form.ts:227-230` — 正しく `handleApiError(error, "保存")` を呼んでいる。

## 準拠すべきプロジェクト規約

### `.claude/CLAUDE.md`
> Error Handling: catch ブロックでは必ず handleApiError を呼び出す
