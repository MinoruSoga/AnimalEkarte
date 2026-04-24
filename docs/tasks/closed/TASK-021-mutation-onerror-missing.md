# TASK-021: useMutation の onError で handleApiError 未呼び出し — 新規 reservation-type 系 API

## 概要

直近コミット `c8625318`（`feat(reservation-type): FE-252`）で追加された 2ファイルの `useMutation` に `onError` コールバックがなく、ミューテーションエラーが `handleApiError` を通らない。コンポーネント側の `onError` 指定に依存しているため、呼び出し側が漏らすとサイレントエラーになる。

## 優先度

HIGH

## 影響ファイル

| ファイル | 行 | 対象 mutation |
|---------|-----|--------------|
| `frontend/src/features/master/api/reservation-type-unavailable-times.ts` | L72 | `useCreateUnavailableTime` |
| `frontend/src/features/master/api/reservation-type-unavailable-times.ts` | L80 | `useDeleteUnavailableTime` |
| `frontend/src/features/master/api/reservation-type-occupations.ts` | L62 | `useLinkOccupation` |
| `frontend/src/features/master/api/reservation-type-occupations.ts` | L70 | `useUnlinkOccupation` |

## 規約違反

`.claude/CLAUDE.md`:
> catch ブロックでは必ず `handleApiError` を呼び出す

## 修正方針

```typescript
// reservation-type-unavailable-times.ts
import { handleApiError } from "@/lib/api-error";

export function useCreateUnavailableTime(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (req: CreateUnavailableTimeRequest) =>
      createUnavailableTime(clinicId, reservationTypeId, req),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: unavailableTimesKey(reservationTypeId) }),
    onError: (error) => handleApiError(error, "予約不可時間の追加"),  // 追加
  });
}

export function useDeleteUnavailableTime(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteUnavailableTime(clinicId, reservationTypeId, id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: unavailableTimesKey(reservationTypeId) }),
    onError: (error) => handleApiError(error, "予約不可時間の削除"),  // 追加
  });
}
```

```typescript
// reservation-type-occupations.ts
export function useLinkOccupation(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (occupationId: string) =>
      linkOccupation(clinicId, reservationTypeId, occupationId),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: occupationsKey(reservationTypeId) }),
    onError: (error) => handleApiError(error, "職種の紐付け"),  // 追加
  });
}

export function useUnlinkOccupation(clinicId: string, reservationTypeId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (occupationId: string) =>
      unlinkOccupation(clinicId, reservationTypeId, occupationId),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: occupationsKey(reservationTypeId) }),
    onError: (error) => handleApiError(error, "職種の紐付け解除"),  // 追加
  });
}
```

## あわせて対応（MEDIUM）

同ファイルの `useGetUnavailableTimes` / `useGetReservationTypeOccupations` に `staleTime`/`gcTime` を追加する:

```typescript
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/query-config";

export function useGetUnavailableTimes(clinicId: string, reservationTypeId: string) {
  return useQuery({
    queryKey: unavailableTimesKey(reservationTypeId),
    queryFn: () => getUnavailableTimes(clinicId, reservationTypeId),
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}
```
