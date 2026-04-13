# BUG-321: ペット登録フォーム内で削除済み動物種の表示状態が不確定

## 優先度
MEDIUM

## 関連セクション
- 14.16 動物種・品種マスタ管理テスト - 「ペット登録中の動物種/品種と既存マスタ選択の整合」

## 問題の説明

ペット登録フォーム（既存ペット編集時）で、削除済みの動物種が以下のいずれかになっている可能性がある：
1. フォーム上に表示される（非表示であるべき）
2. 表示されない（既存選択値として表示されるべき）
3. グレーアウト表示される

既存のペットが削除済み動物種を参照している場合、その種類をフォーム上で適切に表示・編集できるかが不明確。

## 期待動作

### シナリオ 1: 削除済み種類への既存参照
1. ペット「ポチ」が動物種「ハムスター」を参照
2. ハムスター種別を（is_active = false に）削除
3. ペット編集フォームで「ポチ」を開く
4. 動物種セレクトボックスに以下が表示される：
   - 有効な種類：「犬」「猫」「鳥」等（通常表示）
   - 「ハムスター」（グレーアウト、無効化不可の理由を表示、または「現在利用不可」と表示）
5. 種類を変更して保存可能

### シナリオ 2: 新規ペット登録
1. 新規ペット登録フォームで動物種セレクト
2. 有効な種類のみ表示（削除済みは表示しない）

## 技術詳細

### Frontend: 実装状況
- **ファイル**: `frontend/src/features/owners/routes/OwnerForm.tsx` (ペット入力セクション)
- **コンポーネント**: `PetFormSection` または `PetInput`
- **问题**: animals/models.ts に `GetAnimalSpecies()` hook があるが、削除済み種類の扱いが不明確
  - フィルタリング方法：`is_active === true` で有効のみ？
  - フォール back：既存ペット参照時に削除済み種類も追加で取得？

### 修正方法

**フロントエンド: 2段階フェッチ**
```typescript
// hooks/use-animal-species.ts
export function useAnimalSpecies(opts?: { includeInactive?: boolean }) {
  const { data: activeSpecies } = useQuery({
    queryKey: ['animal-species', 'active'],
    queryFn: () => getAnimalSpecies({ isActive: true })
  });

  // 編集モード: 既存ペットの種類を含める
  const { data: speciesForForm } = useQuery(
    enabled: editingPetId != null,
    queryKey: ['animal-species', 'all', editingPetId],
    queryFn: async () => {
      const all = await getAnimalSpecies();
      return all.map(s => ({
        ...s,
        isInactive: !s.isActive,
        label: s.isActive ? s.name : `${s.name} (利用不可)`,
      }));
    }
  );

  return editingPetId ? speciesForForm : activeSpecies;
}
```

**バックエンド: 検証**
- `owner_service.go` の `UpdatePet()` メソッドで、`animal_species_id` の有効性チェック
  - 既存ペット参照の場合、削除済みでも許可（validation skip）
  - 新規選択時、有効なもののみ許可

## テスト確認項目
- [ ] 削除済み種類を参照するペット編集時、フォームで種類を表示できる
- [ ] 削除済み種類はグレーアウト表示、選択不可に見せる
- [ ] 新規ペット登録時、削除済み種類は選択肢に表示しない
- [ ] 既存選択を変更→保存が可能
- [ ] 削除済み種類のまま保存しようとしたら適切なエラーメッセージ表示

## 修正優先度
- MEDIUM（既存参照がある場合のみ影響）
- 現在のテスト実施で詳細が判明するため、修正前に NG 項目を確認する
