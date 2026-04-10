# BUG-308: HospitalizationForm — ハードコード URL パス1箇所

## 概要

`frontend/src/features/hospitalization/routes/HospitalizationForm.tsx:160` で `navigate(\`/hospitalization/${hospitalizationId}\`)` にリテラル文字列が使われていた。

## 影響ファイル

| ファイル | 違反箇所 |
|---------|---------|
| `frontend/src/features/hospitalization/routes/HospitalizationForm.tsx` | line 160 |

## 違反箇所と修正

```tsx
// Before
onClick={() => navigate(`/hospitalization/${hospitalizationId}`)}

// After
onClick={() => navigate(paths.hospitalization.detail.getHref(String(hospitalizationId)))}
```

## 適用ルール

- `config/paths.ts` でURL管理: ハードコードされた URL パス文字列は禁止

## ステータス

✅ 修正済み
