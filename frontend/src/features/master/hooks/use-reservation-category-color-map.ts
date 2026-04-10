import { useMemo, useCallback } from "react";
import type React from "react";
import { useGetReservationCategories } from "@/features/master/api/reservation-categories";
import { useGetReservationCategoryGroups } from "@/features/master/api/reservation-category-groups";
import { PALETTE } from "@/lib/design-tokens";

/**
 * ReservationCategoryColor: RGB hex色をinline style で表現
 * - style: 予約枠背景色（事業に使用）
 * - dotStyle: 凡例の色ドット
 * - hex: 元のhex値（デバッグ/参照用）
 */
interface ReservationCategoryColor {
  style: React.CSSProperties;
  dotStyle: React.CSSProperties;
  hex: string;
}

/** 凡例エントリ (グループ名 + 色) */
interface LegendEntry {
  name: string;
  color: ReservationCategoryColor;
}

const DEFAULT_COLOR: ReservationCategoryColor = {
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

/**
 * hex → RGBA style 変換
 * bg: 10% opacity, border: 30% opacity
 */
function hexToStyle(hex: string): ReservationCategoryColor {
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

/**
 * Returns:
 * - `activeGroupEntries`: アクティブなグループ一覧（凡例表示用）
 * - `colorMap`: カテゴリ名 → 色（グループ色を優先、なければカテゴリ固有色）
 * - `getColor`: カテゴリ名から色取得
 *
 * 凡例はグループのみ表示。カテゴリはグループの色を継承する。
 * グループ未設定のカテゴリはカテゴリ固有色を使用。
 */
export function useReservationCategoryColorMap() {
  const { data: categories = [] } = useGetReservationCategories();
  const { data: groups = [] } = useGetReservationCategoryGroups();

  /** グループが登録されているか */
  const hasGroups = groups.length > 0;

  /** 凡例エントリ: グループがあればグループ一覧、なければアクティブカテゴリ（後方互換） */
  const activeGroupEntries = useMemo<LegendEntry[]>(() => {
    if (hasGroups) {
      return groups
        .filter((g) => g.isActive)
        .map((g) => ({
          name: g.name,
          color: g.color ? hexToStyle(g.color) : DEFAULT_COLOR,
        }));
    }
    // グループ未登録時はカテゴリをそのまま凡例に使う（後方互換）
    return categories
      .filter((c) => c.isActive)
      .map((c) => ({
        name: c.name,
        color: c.color ? hexToStyle(c.color) : DEFAULT_COLOR,
      }));
  }, [hasGroups, groups, categories]);

  /** グループID → 色 マップ */
  const groupColorById = useMemo(
    () => new Map(groups.map((g) => [g.id, g.color ? hexToStyle(g.color) : DEFAULT_COLOR])),
    [groups],
  );

  /** カテゴリ名 → 色マップ（グループ色優先） */
  const colorMap = useMemo(() => {
    const map = new Map<string, ReservationCategoryColor>();
    for (const cat of categories) {
      if (!cat.isActive) continue;
      // グループが設定されていればグループ色を優先
      const groupColor = cat.groupId ? groupColorById.get(cat.groupId) : undefined;
      map.set(cat.name, groupColor ?? (cat.color ? hexToStyle(cat.color) : DEFAULT_COLOR));
    }
    return map;
  }, [categories, groupColorById]);

  const getColor = useCallback(
    (typeName: string): ReservationCategoryColor => colorMap.get(typeName) ?? DEFAULT_COLOR,
    [colorMap],
  );

  // 後方互換: activeEntries は activeGroupEntries のエイリアス
  return { activeGroupEntries, activeEntries: activeGroupEntries, colorMap, getColor };
}
