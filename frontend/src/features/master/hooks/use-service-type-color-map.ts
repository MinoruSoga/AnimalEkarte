import { useMemo, useCallback } from "react";
import type React from "react";
import { useGetServiceTypes } from "@/features/master/api/service-types";

/**
 * ServiceTypeColor: RGB hex色をinline style で表現
 * - style: 予約枠背景色（事業に使用）
 * - dotStyle: 凡例の色ドット
 * - hex: 元のhex値（デバッグ/参照用）
 */
interface ServiceTypeColor {
  style: React.CSSProperties;
  dotStyle: React.CSSProperties;
  hex: string;
}

interface ServiceTypeEntry {
  name: string;
  color: ServiceTypeColor;
}

const DEFAULT_COLOR: ServiceTypeColor = {
  style: {
    backgroundColor: "#F1F0EE",
    color: "#9B9A97",
    borderColor: "#E3E2E0",
  },
  dotStyle: {
    backgroundColor: "#9B9A97",
  },
  hex: "#9B9A97",
};

/**
 * hex → RGBA style 変換
 * bg: 10% opacity, border: 30% opacity
 */
function hexToStyle(hex: string): ServiceTypeColor {
  const hex6 = hex.startsWith("#") ? hex : `#${hex}`;

  // hex → RGB 変換（簡易版：16進数直接パース）
  const r = parseInt(hex6.slice(1, 3), 16);
  const g = parseInt(hex6.slice(3, 5), 16);
  const b = parseInt(hex6.slice(5, 7), 16);

  return {
    style: {
      backgroundColor: `rgba(${r}, ${g}, ${b}, 0.1)`,     // 10%
      color: hex6,                                          // 100%
      borderColor: `rgba(${r}, ${g}, ${b}, 0.3)`,         // 30%
    },
    dotStyle: {
      backgroundColor: hex6,  // 100%
    },
    hex: hex6,
  };
}

/**
 * Returns a color map derived from active serviceType records from DB.
 * Each service type gets its color from the `color` field（hex値）.
 */
export function useServiceTypeColorMap() {
  const { data: serviceTypes = [] } = useGetServiceTypes();

  const activeEntries = useMemo<ServiceTypeEntry[]>(() => {
    if (!serviceTypes || serviceTypes.length === 0) {
      return [];
    }

    return serviceTypes
      .filter((st) => st.isActive)
      .map((st) => ({
        name: st.name,
        color: st.color ? hexToStyle(st.color) : DEFAULT_COLOR,
      }));
  }, [serviceTypes]);

  const colorMap = useMemo(
    () => new Map(activeEntries.map((e) => [e.name, e.color])),
    [activeEntries]
  );

  const getColor = useCallback(
    (typeName: string): ServiceTypeColor => colorMap.get(typeName) ?? DEFAULT_COLOR,
    [colorMap]
  );

  return { activeEntries, colorMap, getColor };
}
