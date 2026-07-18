# src/hooks — Shared Global Hooks

## 責務

このディレクトリは**クロスカッティングな共有フック**のみ配置する。

- ✅ ここに置くもの: 複数 feature にまたがる共有フック（認証、権限、ページタイトル、モーダル状態など）
- ❌ ここに置かないもの: 特定 feature 専用のフック → `features/xxx/hooks/` に配置する

### 例外: フック本体ではない支持ファイル

`auth-context.ts`（`createContext` の定義のみ、フックではない）は例外として本ディレクトリに同居する。唯一の消費者 `use-auth.ts` と同じ場所に置くのが最も凝集的で、`features/auth` に置くと層逆転（features → hooks/lib への依存は許可・逆方向は禁止）になるため。命名規則(`use-kebab-case.ts`)の対象外（FE7-4(a)・2026-07-18）。

## 命名規則

```
use-kebab-case.ts   例: use-modal-state.ts, use-pagination.ts
```

テストファイル: `use-xxx.test.ts` を同ディレクトリに並置する。

## React フックルール (MANDATORY)

```typescript
// ✅ トップレベルでのみ呼び出す
function useMyHook() {
  const [state, setState] = useState(false);  // 常にトップレベル
}

// ❌ 条件分岐・ループ内での呼び出し禁止
if (condition) {
  const [state] = useState(false);
}
```

## 安定参照 (useCallback / useMemo)

コールバックを外部コンポーネントへ渡す場合は `useCallback` でラップする。
依存配列を省略しない。

```typescript
// ✅
const handleClose = useCallback(() => setOpen(false), []);

// ❌
const handleClose = () => setOpen(false);  // 毎レンダーで新しい参照
```

## Query Cache 共有パターン

複数 feature が同じエンティティを参照する場合、このディレクトリに shared hook を置いて query key を統一する。

```typescript
// src/hooks/use-pet.ts — 18 feature から参照
export function useGetPet(petId: string) {
  return useQuery({
    queryKey: ["pet", petId],  // features/pets と同じキーでキャッシュ共有
    ...
  });
}
```

## フック一覧

### ユーティリティ系（汎用）

| ファイル | 用途 |
|---------|------|
| `use-auth.ts` | 認証状態・ログインユーザー情報 |
| `use-permission.ts` | 権限チェック (`can("edit")` 等) |
| `use-modal-state.ts` | モーダル開閉状態の汎用管理 |
| `use-pagination.ts` | ページネーション状態管理 |
| `use-side-peek-dirty.ts` | サイドピーク未保存変更フラグ |
| `use-unsaved-changes.ts` | 未保存変更の離脱ガード |
| `use-title.ts` | ページタイトル設定 |
| `use-sortable-data.ts` | ドラッグ&ドロップ並べ替えデータ管理 |
| `use-sortable-list.ts` | ソータブルリスト UI ロジック |
| `use-reduced-motion.ts` | `prefers-reduced-motion` メディアクエリ |
| `use-staff-validation.ts` | スタッフバリデーション共有ロジック |
| `use-reservation-type-color-map.ts` | 予約タイプカラーマッピング |
| `use-clinic-scope.ts` | #86 拠点横断表示: URL ?clinics= driven の医院スコープ管理（selectedClinicIds / handleToggleClinic / clinicNameById） |

### Cross-feature データ系（Query Cache 共有）

| ファイル | 参照元 feature 数 | 用途 |
|---------|-----------------|------|
| `use-pet.ts` | 19 | ペット情報取得（accounting / trimming / medical-records / owner-report 他） |
| `use-animal-species.ts` | 2 | 動物種マスタ取得（owners / medical-records） |
| `use-pet-selection.ts` | 4 | ペット選択ロジック（examinations / trimming / hospitalization 他） |
| `use-pet-selection-page.ts` | 4 | ペット選択ページパターン（accounting / trimming / medical-records 他） |
| `use-owner.ts` | 4 | オーナー情報取得（medical-records / owners / owner-report） |
| `use-master-items.ts` | 6 | マスターアイテム汎用取得（trimming / examinations 他） |
| `use-treatment-master.ts` | 3 | 診療マスターデータ（medical-records / shared components） |
| `use-reservation-types.ts` | 1+ | 予約タイプ＋グループ取得（shared/ReservationFormModal） |
| `use-vaccinations.ts` | 2 | ワクチン接種 **作成** mutation（medical-records / vaccinations。query は `use-pet-vaccinations.ts` に分離） |
| `use-staffs.ts` | 2 | スタッフ一覧取得（reception / medical-records） |
| `use-clinic-holidays.ts` | 複数 | クリニック休診日取得 |
| `use-owner-line-tags.ts` | 3+ | LINE/LSTEP タグ取得（owners / medical-records / reservations / shared） |
| `use-trimming-course-types.ts` | 2 | トリミングコース種別マスタ（master / trimming） |
| `use-pet-vaccinations.ts` | 2 | ペット別予防接種履歴取得（medical-records / owner-report） |
| `use-pet-checkup-results.ts` | 2 | ペット別健診結果取得（checkups / owner-report） |
| `use-examinations.ts` | 2 | 検査記録一覧取得（examinations / medical-records） |
| `use-update-examination.ts` | 2 | 検査記録更新 mutation（examinations / medical-records） |
| `use-cash-register-closes.ts` | 2 | レジ締め一覧取得（cash-register / accounting の締め後編集判定） |
| `use-update-reservation.ts` | 2 | 予約更新 mutation（reservations / reception のモーダル編集保存） |
| `use-get-reservations.ts` | 2 | 予約一覧取得（reservations / medical-records の当日予約再利用判定） |
| `use-create-reservation.ts` | 2 | 予約作成 mutation（reservations / medical-records のカルテ自動作成） |
| `use-record-pet-death.ts` | 2 | ペット死亡記録 mutation（owners / pets の PetDeceasedRecordButton→PetDeceasedDialog） |
| `use-revoke-pet-death.ts` | 2 | ペット死亡記録解除 mutation（owners / pets の PetDeceasedRecordButton→PetDeceasedBanner） |
| `use-clinic-tax-rates.ts` | 2 | 病院マスタ消費税率取得（accounting-reports / accounting の AccountingDocument） |
| `use-get-reservation.ts` | 2+ | 予約詳細取得（reservations / components/shared/ReservationFormModal） |
| `use-update-reservation-route.ts` | 2+ | 予約経路更新 mutation（reservations / components/shared/ReservationFormModal 想定） |
| `use-reservation-type-unavailable-times.ts` | 2+ | 予約タイプ別予約不可時間取得（master / components/shared/ReservationFormModal。作成・削除 mutation は `features/master/api/reservation-type-unavailable-times.ts` に残置） |
