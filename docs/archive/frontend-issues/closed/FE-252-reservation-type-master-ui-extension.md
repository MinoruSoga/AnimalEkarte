# FE-252: 予約区分マスタ編集画面に「予約不可時間」タブ & 「対応職種」セクション追加

**Status**: Open
**Priority**: High
**Affects**: features/master — 予約区分設定画面
**Date Created**: 2026-04-16
**Related**: TASK-001, BE-116（先に完了必要）

## Summary

カルテ側の予約区分マスタ編集画面（`features/master/`）に、
BE-116 で実装した API を使って「予約不可時間」と「対応職種」の設定 UI を追加する。

## 現状のコード

```typescript
// frontend/src/features/master/routes/ReservationTypeSidePanel.tsx
// CategorySidePanel (memo) が MasterSidePanel で予約区分フォームを描画している。
// 現状は名称・グループ・LINE予約設定の各フィールドのみ。
// 不可時間・職種紐付けのセクションは存在しない。

// frontend/src/features/master/routes/ReservationTypeSettings.tsx:486
// categoryEditTarget !== null の場合に <CategorySidePanel /> をレンダリングしている。
// CategorySidePanel に渡される item: ReservationType には
//   item.id      → reservationTypeId
//   item.clinicId → clinicId
// が含まれる（api/reservation-types.ts の transformReservationType で導出済み）。

// frontend/src/features/master/api/occupations.ts:68
// useGetAllOccupations() が /v1/masters/occupations を取得し Occupation[] を返す。
// 職種追加ドロップダウンはこの hook を再利用する（新規実装不要）。
```

## 必要な変更

### 1. 型定義

`frontend/src/features/master/api/types.ts`（または新規作成）に追加:

```typescript
// models.ts から導出（make codegen 後に利用可能）
import type {
  ReservationTypeUnavailableTime,
  ReservationTypeOccupation,
} from "@/types/generated/models";

// API リクエスト型（models.ts から Omit で導出）
export type CreateUnavailableTimeRequest = Omit<
  ReservationTypeUnavailableTime,
  "id" | "clinic_id" | "reservation_type_id" | "created_at" | "updated_at"
>;

export type LinkOccupationRequest = {
  occupation_id: number;
};
```

### 2. API hooks 追加

`frontend/src/features/master/api/reservation-type-unavailable-times.ts`:

```typescript
// 予約不可時間
export function useGetUnavailableTimes(clinicId: string, reservationTypeId: string) {
  return useQuery({
    queryKey: ["reservation-types", reservationTypeId, "unavailable-times"],
    queryFn: () => axios.get<{ data: ReservationTypeUnavailableTime[] }>(
      `/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/unavailable-times`
    ).then(r => r.data.data),
  });
}

export function useCreateUnavailableTime(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateUnavailableTimeRequest) =>
      axios.post(`/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/unavailable-times`, data),
    onSuccess: () => queryClient.invalidateQueries({
      queryKey: ["reservation-types", reservationTypeId, "unavailable-times"]
    }),
  });
}

export function useDeleteUnavailableTime(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) =>
      axios.delete(`/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/unavailable-times/${id}`),
    onSuccess: () => queryClient.invalidateQueries({
      queryKey: ["reservation-types", reservationTypeId, "unavailable-times"]
    }),
  });
}
```

`frontend/src/features/master/api/reservation-type-occupations.ts`:

```typescript
export function useGetReservationTypeOccupations(clinicId: string, reservationTypeId: string) {
  return useQuery({
    queryKey: ["reservation-types", reservationTypeId, "occupations"],
    queryFn: () => axios.get<{ data: ReservationTypeOccupation[] }>(
      `/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/occupations`
    ).then(r => r.data.data),
  });
}

export function useLinkOccupation(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (occupationId: number) =>
      axios.post(
        `/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/occupations`,
        { occupation_id: occupationId }
      ),
    onSuccess: () => queryClient.invalidateQueries({
      queryKey: ["reservation-types", reservationTypeId, "occupations"],
    }),
  });
}

export function useUnlinkOccupation(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (occupationId: number) =>
      axios.delete(
        `/v1/clinics/${clinicId}/reservation-types/${reservationTypeId}/occupations/${occupationId}`
      ),
    onSuccess: () => queryClient.invalidateQueries({
      queryKey: ["reservation-types", reservationTypeId, "occupations"],
    }),
  });
}
```

### 3. 予約不可時間コンポーネント

`frontend/src/features/master/components/ReservationTypeUnavailableTimesSection.tsx`:

UI 仕様:
- 既存設定一覧をテーブルで表示（種別・曜日 or 日付・開始〜終了・削除ボタン）
- 「追加」ボタン押下でインラインフォームを表示
  - 種別選択: `weekly` / `specific` のセレクト
  - `weekly` 選択時: 曜日ドロップダウン（日〜土）
  - `specific` 選択時: 日付ピッカー（`NotionDatePicker` を使用）
  - 開始時刻・終了時刻: 30分単位のセレクト（`"00:00"` 〜 `"23:30"`）
  - 保存は `useActionState` + `SubmitButton` パターン
- バリデーションエラーメッセージ仕様（`FormFieldError` コンポーネントで表示）:
  - 種別未選択: `"種別を選択してください"`
  - `weekly` で曜日未選択: `"曜日を選択してください"`
  - `specific` で日付未入力: `"日付を選択してください"`
  - 開始時刻・終了時刻いずれか未選択: `"時刻を選択してください"`
  - 終了時刻 ≤ 開始時刻: `"終了時刻は開始時刻より後にしてください"`
  - API 409 (重複): `handleApiError` がトーストで `"この時間帯はすでに登録されています"` を表示

```typescript
export function ReservationTypeUnavailableTimesSection({
  clinicId,
  reservationTypeId,
}: Props) {
  const { data: times } = useGetUnavailableTimes(clinicId, reservationTypeId);
  const deleteTime = useDeleteUnavailableTime(clinicId, reservationTypeId);

  const [state, formAction] = useActionState(async (_: unknown, formData: FormData) => {
    try {
      // formData から値を取り出して createUnavailableTime を呼ぶ
    } catch (error) {
      handleApiError(error, "予約不可時間の追加");
      return { success: false };
    }
    return { success: true };
  }, null);

  return (
    // ... UI 実装
  );
}
```

### 4. 対応職種コンポーネント

`frontend/src/features/master/components/ReservationTypeOccupationsSection.tsx`:

UI 仕様:
- 紐付け済み職種をバッジで表示（削除ボタン付き）
- 「職種を追加」ドロップダウン（クリニックの occupations 一覧から未紐付けのみ表示）
- 追加は即時 API 呼び出し（フォーム不要）
- 職種一覧は既存の `useGetAllOccupations`（`api/occupations.ts:68`）を使用（新規実装不要）
  - 全職種から `linkedOccupations` に含まれる id を除外してドロップダウンに表示する

### 5. 予約区分編集画面への組み込み

`frontend/src/features/master/routes/ReservationTypeSidePanel.tsx` の `CategorySidePanel` に追加する。
`MasterSidePanel` の children 末尾（LINE予約設定 `</div>` の後）に追加:

```typescript
// ReservationTypeSidePanel.tsx の CategorySidePanel props に以下を追加:
// item: ReservationType | null → item が null（新規作成モード）のときは両セクションを非表示
// item.id → reservationTypeId, item.clinicId → clinicId として渡す

// MasterSidePanel children 末尾に追加:
{item !== null ? (
  <>
    <ReservationTypeUnavailableTimesSection
      clinicId={item.clinicId}
      reservationTypeId={item.id}
    />
    <ReservationTypeOccupationsSection
      clinicId={item.clinicId}
      reservationTypeId={item.id}
    />
  </>
) : null}
```

- `item === null`（新規作成モード）では両セクションを表示しない（IDが未確定のため）
- 両コンポーネントは `@/features/master/components/` に配置し、
  `features/master/index.ts` には追加不要（SidePanel 内でのみ使用）

## UI 操作フロー

1. 管理者が「設定」→「予約区分」→ 対象区分の編集を開く
2. 画面下部に「予約不可時間」セクションが表示される
3. 「追加」ボタンを押し、種別・曜日or日付・時刻を入力して「保存」
4. 一覧に追加され、削除ボタンで個別削除可能
5. 「対応職種」セクションでドロップダウンから職種を選択して紐付け
6. バッジの×ボタンで職種の紐付けを解除

## プロジェクトルール遵守チェック

- [ ] `any` 型なし
- [ ] `FC` / `forwardRef` なし
- [ ] 条件レンダー `? ... : null`（`&&` 禁止）
- [ ] `useActionState` + `SubmitButton` でフォーム送信
- [ ] Design Tokens（`C`, `STYLE`）使用、Hex カラー直書き禁止
- [ ] 型は `models.ts` から導出（手書き interface 禁止）
- [ ] Feature Indexing: 外部からのインポートは `index.ts` 経由

## 依存関係

- BE-116 の全エンドポイントが実装済みであること
- `make codegen` により `models.ts` に新型が追加されていること

## 完了条件

- [ ] 予約不可時間を追加・削除できる（weekly / specific 両対応）
- [ ] 対応職種を追加・削除できる
- [ ] `pnpm lint` 0 errors
- [ ] `pnpm build` が通る（型エラーなし）
- [ ] 既存の予約区分 CRUD 機能に影響なし
