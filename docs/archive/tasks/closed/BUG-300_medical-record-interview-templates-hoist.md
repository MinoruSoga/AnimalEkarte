# BUG-300: MedicalRecordInterview — templates 配列がコンポーネント内で定義されている

## 概要

`rendering-hoist-jsx` 違反。`MedicalRecordInterview.tsx` 内の `templates` 配列がコンポーネント関数本体内で定義されており、レンダーのたびに新しい配列参照が生成される。`memo()` を適用しているにもかかわらず、このオブジェクトが子コンポーネント（`InterviewChiefComplaint`）の prop として渡されるため、memo の効果が無効化される。

## 影響ファイル

- `frontend/src/features/medical-records/components/MedicalRecordInterview.tsx`

## 違反箇所（修正前）

```tsx
// lines 57–62（コンポーネント関数内）
export const MedicalRecordInterview = memo(function MedicalRecordInterview({...}) {
  const templates = [                          // ← VIOLATION
    { label: "定期検診", text: "# 定期検診\n特に異常なし。食欲・元気あり。" },
    { label: "ワクチン",  text: "# 混合ワクチン接種\n体調良好。" },
    { label: "下痢・嘔吐", text: "# 消化器症状\n・嘔吐：あり（回数：　）\n・下痢：あり（性状：　）\n・食欲：なし" },
    { label: "皮膚",     text: "# 皮膚症状\n・痒み：あり\n・発赤：あり\n・部位：" },
  ];
  // ...
  <InterviewChiefComplaint templates={templates} ... />
```

## 修正内容

```tsx
// モジュールレベルに巻き上げ（DEFAULT_HISTORY_ITEMS の直上）
const INTERVIEW_TEMPLATES: { label: string; text: string }[] = [
  { label: "定期検診", text: "# 定期検診\n特に異常なし。食欲・元気あり。" },
  { label: "ワクチン",  text: "# 混合ワクチン接種\n体調良好。" },
  { label: "下痢・嘔吐", text: "# 消化器症状\n・嘔吐：あり（回数：　）\n・下痢：あり（性状：　）\n・食欲：なし" },
  { label: "皮膚",     text: "# 皮膚症状\n・痒み：あり\n・発赤：あり\n・部位：" },
];

// コンポーネント内の const templates = [...] を削除し、参照を差し替え
<InterviewChiefComplaint templates={INTERVIEW_TEMPLATES} ... />
```

## 適用ルール

- `rendering-hoist-jsx`: 静的な JSX・オブジェクト・配列はコンポーネント外のモジュール定数に巻き上げる
- `rerender-memo`: memo() は props の参照が安定している場合にのみ有効。配列をインラインで渡すと毎レンダー新参照が生成され memo が機能しない

## ステータス

✅ 修正済み（commit: 対応コミットにて）
