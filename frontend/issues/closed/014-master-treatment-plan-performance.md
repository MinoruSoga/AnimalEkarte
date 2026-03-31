---
status: closed
closed_at: 2026-03-16
---

# [master] TreatmentPlanMaster: パフォーマンス問題（tabConfigs・全 hook 常時呼び出し）

## 優先度
低

## 種別
パフォーマンス

## 対象ファイル
`frontend/src/features/master/routes/TreatmentPlanMaster.tsx`

## 問題

### 1. `tabConfigs` が useMemo なし

`tabConfigs: Record<string, TreatmentTabConfig>` がコンポーネント関数本体で定義されており（L559〜L739）、
`useMemo` が使われていない。5 エンティティのミューテーション hook の参照を含む巨大なオブジェクトが、
毎レンダリングで再生成されている。

```tsx
// 現状（問題）
const tabConfigs: Record<string, TreatmentTabConfig> = {
  consultation: { ..., onCreate: createConsultation, ... },
  examination: { ..., onCreate: createExaminationType, ... },
  // ...5エンティティ分
};

// 修正後
const tabConfigs = useMemo<Record<string, TreatmentTabConfig>>(() => ({
  consultation: { ..., onCreate: createConsultation, ... },
  // ...
}), [createConsultation, createExaminationType, /* ...全依存 */]);
```

### 2. アクティブタブに関わらず 5 エンティティ分の mutation hook が常時呼ばれている

タブ切替前から全エンティティの `useMutation` が呼ばれており、不要な状態保持が発生している。
React Query の `useMutation` はステートフルなため、使われていないタブのミューテーション hook が
メモリを消費し続ける。

アクティブタブのみの hook を呼ぶ設計（`useMemo` + 動的 hook 選択）は hooks のルール上不可能なので、
`tabConfigs` を `useMemo` でメモ化することで再生成コストを削減する対応が現実的。

### 3. `TreatmentTabContent` 内で `useTransition` を使っていない

`handleSave` が `onUpdate`・`onCreate` を直接呼び出しており、他のマスタページが `startSaveTransition` で
非緊急マークしている設計と不統一。サブミット中のローディング状態管理も不足している。

## 修正方針

1. `tabConfigs` を `useMemo` でメモ化する
2. `TreatmentTabContent` の `handleSave` を `useTransition` でラップし、`isSavePending` を返す
3. サブミットボタンに `disabled={isSavePending}` を追加する
