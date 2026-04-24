# FE-009: マスタ編集ボタン - ナビゲーションパス未定義

## 問題

問診タブの「主訴区分」と「定型文挿入」のマスタ編集ボタン（⚙️）をクリックすると、JavaScript エラーが発生してナビゲートされない。

**エラー**: `Uncaught TypeError: Cannot read properties of undefined (reading 'chiefComplaint')`

---

## 根本原因

InterviewChiefComplaint コンポーネントが参照するパスが `config/paths.ts` に定義されていない：

```typescript
// InterviewChiefComplaint.tsx:51
navigate(paths.settings.interview.chiefComplaint.getHref())  // ❌ undefined

// InterviewChiefComplaint.tsx:75
navigate(paths.settings.interview.interviewTemplate.getHref())  // ❌ undefined
```

**存在しないパス**:
- `paths.settings.interview.chiefComplaint`
- `paths.settings.interview.interviewTemplate`

**存在する paths.settings**:
- clinic, staff, treatmentItems, diagnosis, trimming, medicine, serviceType, hospitalization, cage, insurance, jobTitle, inquiryTemplates

---

## コンポーネント分析

### InterviewChiefComplaint.tsx

```typescript
// 行51
onClick={() => navigate(paths.settings.interview.chiefComplaint.getHref())}

// 行75
onClick={() => navigate(paths.settings.interview.interviewTemplate.getHref())}
```

両ボタンとも `paths.settings.interview` 参照 → **エラーで navigate 不実行**

---

## 修正対応

### オプション 1: マスタ設定パス追加（推奨）

**config/paths.ts** に以下を追加：

```typescript
interview: {
  chiefComplaint: {
    path: "/settings/interview/chief-complaint",
    getHref: () => "/settings/interview/chief-complaint",
  },
  interviewTemplate: {
    path: "/settings/interview/templates",
    getHref: () => "/settings/interview/templates",
  },
},
```

### オプション 2: ボタン削除

マスタ設定ページが未実装の場合、一時的にボタンを削除し、後で実装。

---

## テスト環境
- 記録 ID: 17
- テスト日時: 2026-03-16 13:10 JST

---

## 優先度
**🟡 MEDIUM** - ナビゲーション失敗・ユーザー操作阻害

---

## ブロッカー
- マスタ設定ページ（/settings/interview/...）の実装状況確認が必要

