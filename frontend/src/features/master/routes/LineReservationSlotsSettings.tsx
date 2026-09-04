import { useCallback, useEffect, useMemo } from "react";
import { useSearchParams } from "react-router";
import { CalendarDays, Info } from "lucide-react";
import { FormHeader } from "@/components/shared/Form/FormHeader";
import { PermissionBadges } from "@/components/shared/PermissionBadges/PermissionBadges";
import { C, ICON, LAYOUT } from "@/lib/design-tokens";
import { ResourceMasterReservationType } from "@/types/generated/models";
import { useGetReservationTypes } from "../api/reservation-types";
import { ReservationTypeAvailableSlotsCalendar } from "../components/ReservationTypeAvailableSlotsCalendar";
import { ReservationTypeTree } from "../components/ReservationTypeTree";

export function LineReservationSlotsSettings() {
  const { data: types = [], isLoading } = useGetReservationTypes();

  const allLeaves = useMemo(() => types.filter((t) => t.isLeaf), [types]);
  const activeLeaves = useMemo(() => allLeaves.filter((t) => t.isActive), [allLeaves]);

  const [searchParams, setSearchParams] = useSearchParams();

  // typeId 正規化ロジック
  useEffect(() => {
    if (!types.length) return; // まだロード中

    const typeIdParam = searchParams.get("typeId");

    if (!typeIdParam) {
      // 未指定: activeLeaf[0] ?? allLeaf[0]
      const fallback = activeLeaves[0] ?? allLeaves[0];
      if (fallback) {
        setSearchParams(
          (p) => {
            const next = new URLSearchParams(p);
            next.set("typeId", fallback.id);
            return next;
          },
          { replace: true },
        );
      }
      return;
    }

    const found = types.find((t) => t.id === typeIdParam);
    if (!found) {
      // 存在しない ID: fallback
      const fallback = activeLeaves[0] ?? allLeaves[0];
      if (fallback) {
        setSearchParams(
          (p) => {
            const next = new URLSearchParams(p);
            next.set("typeId", fallback.id);
            return next;
          },
          { replace: true },
        );
      }
      return;
    }

    if (!found.isLeaf) {
      // parent: 配下の最初の leaf に正規化
      const firstLeaf = allLeaves.find((t) => t.parentId === found.id);
      if (firstLeaf) {
        setSearchParams(
          (p) => {
            const next = new URLSearchParams(p);
            next.set("typeId", firstLeaf.id);
            return next;
          },
          { replace: true },
        );
      }
      return;
    }
    // leaf: そのまま
  }, [types, searchParams, setSearchParams, allLeaves, activeLeaves]);

  // 選択中の leaf
  const selectedType = allLeaves.find((t) => t.id === searchParams.get("typeId"));

  const handleTypeChange = useCallback(
    (value: string) => {
      setSearchParams(
        (prev) => {
          const params = new URLSearchParams(prev);
          params.set("typeId", value);
          return params;
        },
        { replace: true },
      );
    },
    [setSearchParams],
  );

  const breadcrumb = selectedType
    ? selectedType.depth === 1
      ? `${selectedType.parentName} / ${selectedType.name}`
      : selectedType.name
    : null;

  return (
    <div className={`flex-1 flex flex-col h-full ${C.bgPage} min-w-0 w-full`}>
      <FormHeader
        title="LINE予約枠"
        description="LINE予約で選択できる開始時刻を日別に設定します"
        icon={<CalendarDays className={`${ICON.page} ${C.text}`} />}
        action={<PermissionBadges resource={ResourceMasterReservationType} />}
      />

      <div className={`flex items-center gap-1.5 px-4 pt-3 pb-0 text-sm ${C.text50}`}>
        <Info className={`${ICON.smXs} shrink-0`} />
        ここで登録した開始時刻は、営業時間から自動生成される予約可能枠に追加されます（重複する時刻は追加されません）。枠を登録しても他の時刻が予約不可になることはなく、枠が未登録の日も営業時間から自動生成されます。
      </div>

      <div className="flex-1 min-h-0 flex flex-col md:flex-row gap-0 overflow-hidden mt-3">
        {/* 左: ツリーパネル */}
        <div
          className={`${LAYOUT.treeNavPanel.responsiveWidth} shrink-0 max-h-64 md:max-h-none border-b md:border-b-0 md:border-r ${C.borderMedium} overflow-y-auto ${C.bgWhite}`}
        >
          {isLoading ? (
            <p className={`text-sm ${C.text40} p-4`}>読み込み中...</p>
          ) : (
            <ReservationTypeTree
              types={types}
              selectedId={selectedType?.id ?? null}
              onSelect={handleTypeChange}
            />
          )}
        </div>

        {/* 右: カレンダーパネル */}
        <div className="flex-1 min-h-0 flex flex-col p-4 gap-3 overflow-hidden">
          {breadcrumb ? <p className={`text-sm font-medium ${C.text60}`}>{breadcrumb}</p> : null}

          {isLoading ? (
            <p className={`text-sm ${C.text40} py-4`}>読み込み中...</p>
          ) : selectedType ? (
            <ReservationTypeAvailableSlotsCalendar
              key={selectedType.id}
              clinicId={selectedType.clinicId}
              reservationTypeId={selectedType.id}
            />
          ) : (
            <p className={`text-sm ${C.text40} py-4`}>予約区分がありません</p>
          )}
        </div>
      </div>
    </div>
  );
}
