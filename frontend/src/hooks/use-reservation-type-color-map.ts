import { useMemo, useCallback } from "react";
import type React from "react";
import { useQuery } from "@tanstack/react-query";
import { axios } from "@/lib/axios";
import { QUERY_STALE_TIMES, QUERY_GC_TIMES } from "@/lib/react-query";
import { queryKeys } from "@/lib/query-keys";
import { PALETTE } from "@/lib/design-tokens";
import type { ReservationType, ReservationTypeGroup } from "@/types/generated/models";

// ─────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────

interface ReservationTypeForColorMap {
  id: string;
  name: string;
  color?: string;
  isActive: boolean;
  groupId?: string;
}

interface ReservationTypeGroupForColorMap {
  id: string;
  name: string;
  color?: string;
  isActive: boolean;
}

/**
 * ReservationTypeColor: RGB hex色をinline style で表現
 * - style: 予約枠背景色（カレンダーに使用）
 * - dotStyle: 凡例の色ドット
 * - hex: 元のhex値
 */
export interface ReservationTypeColor {
  style: React.CSSProperties;
  dotStyle: React.CSSProperties;
  hex: string;
  isInactive?: boolean;
}

/** 凡例エントリ (グループ名 + 色) */
export interface LegendEntry {
  name: string;
  color: ReservationTypeColor;
}

// ─────────────────────────────────────────────────
// Transforms
// ─────────────────────────────────────────────────

function transformReservationType(data: ReservationType): ReservationTypeForColorMap {
  return {
    id: String(data.id ?? 0),
    name: data.name,
    color: data.color,
    isActive: data.is_active,
    groupId: data.group_id ? String(data.group_id) : undefined,
  };
}

function transformReservationTypeGroup(
  data: ReservationTypeGroup,
): ReservationTypeGroupForColorMap {
  return {
    id: String(data.id ?? 0),
    name: data.name,
    color: data.color,
    isActive: data.is_active,
  };
}

// ─────────────────────────────────────────────────
// Internal hooks (query keys match features/master for cache sharing)
// ─────────────────────────────────────────────────

function useGetReservationTypes() {
  return useQuery({
    queryKey: queryKeys.masters.category("reservation-types"),
    queryFn: async (): Promise<ReservationTypeForColorMap[]> => {
      const { data } = await axios.get<ReservationType[]>("/v1/masters/reservation-types");
      return data.map(transformReservationType);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

function useGetReservationTypeGroups() {
  return useQuery({
    queryKey: queryKeys.masters.category("reservation-type-groups"),
    queryFn: async (): Promise<ReservationTypeGroupForColorMap[]> => {
      const { data } = await axios.get<ReservationTypeGroup[]>(
        "/v1/masters/reservation-type-groups",
      );
      return data.map(transformReservationTypeGroup);
    },
    staleTime: QUERY_STALE_TIMES.STATIC,
    gcTime: QUERY_GC_TIMES.LONG,
  });
}

// ─────────────────────────────────────────────────
// Color utilities
// ─────────────────────────────────────────────────

const DEFAULT_COLOR: ReservationTypeColor = {
  style: {
    backgroundColor: PALETTE.mutedBg,
    color: PALETTE.grayMedium,
    borderColor: PALETTE.grayLight,
  },
  dotStyle: {
    backgroundColor: PALETTE.grayMedium,
  },
  hex: PALETTE.grayMedium,
};

/** hex → RGBA style 変換 (bg: 10% opacity, border: 30% opacity) */
function hexToStyle(hex: string): ReservationTypeColor {
  const hex6 = hex.startsWith("#") ? hex : `#${hex}`;
  const red = parseInt(hex6.slice(1, 3), 16);
  const green = parseInt(hex6.slice(3, 5), 16);
  const blue = parseInt(hex6.slice(5, 7), 16);

  return {
    style: {
      backgroundColor: `rgba(${red}, ${green}, ${blue}, 0.1)`,
      color: hex6,
      borderColor: `rgba(${red}, ${green}, ${blue}, 0.3)`,
    },
    dotStyle: {
      backgroundColor: hex6,
    },
    hex: hex6,
  };
}

// ─────────────────────────────────────────────────
// Public hook
// ─────────────────────────────────────────────────

/**
 * Returns:
 * - `activeGroupEntries`: アクティブなグループ一覧（凡例表示用）
 * - `colorMap`: カテゴリ名 → 色（グループ色を優先、なければカテゴリ固有色）
 * - `getColor`: カテゴリ名から色取得
 *
 * 凡例はグループのみ表示。カテゴリはグループの色を継承する。
 * グループ未設定のカテゴリはカテゴリ固有色を使用。
 */
export function useReservationTypeColorMap() {
  const { data: categories = [] } = useGetReservationTypes();
  const { data: groups = [] } = useGetReservationTypeGroups();

  const hasGroups = groups.length > 0;

  const activeGroupEntries = useMemo<LegendEntry[]>(() => {
    if (hasGroups) {
      return groups
        .filter((g) => g.isActive)
        .map((g) => ({
          name: g.name,
          color: g.color ? hexToStyle(g.color) : DEFAULT_COLOR,
        }));
    }
    return categories
      .filter((c) => c.isActive)
      .map((c) => ({
        name: c.name,
        color: c.color ? hexToStyle(c.color) : DEFAULT_COLOR,
      }));
  }, [hasGroups, groups, categories]);

  const groupColorById = useMemo(
    () => new Map(groups.map((g) => [g.id, g.color ? hexToStyle(g.color) : DEFAULT_COLOR])),
    [groups],
  );

  const colorMap = useMemo(() => {
    const map = new Map<string, ReservationTypeColor>();
    for (const cat of categories) {
      const groupColor = cat.groupId ? groupColorById.get(cat.groupId) : undefined;
      const baseColor = groupColor ?? (cat.color ? hexToStyle(cat.color) : DEFAULT_COLOR);
      map.set(cat.name, cat.isActive ? baseColor : { ...baseColor, isInactive: true });
    }
    return map;
  }, [categories, groupColorById]);

  const getColor = useCallback(
    (typeName: string): ReservationTypeColor => colorMap.get(typeName) ?? DEFAULT_COLOR,
    [colorMap],
  );

  return { activeGroupEntries, activeEntries: activeGroupEntries, colorMap, getColor };
}
