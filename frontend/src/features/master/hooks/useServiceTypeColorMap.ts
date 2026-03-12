import { useMemo, useCallback } from "react";
import { useMasterItems } from "@/hooks/use-master-items";
import { C, PALETTE } from "@/lib/design-tokens";

interface ServiceTypeColor {
  bg: string;
  text: string;
  border: string;
  dot: string;
}

interface ServiceTypeEntry {
  name: string;
  color: ServiceTypeColor;
}

// Palette rotation for service types
const COLOR_PALETTES: ServiceTypeColor[] = [
  {
    bg: C.bgAccentLight,
    text: C.textAccentDark,
    border: "border-[#B8D4E3]",
    dot: `bg-[${PALETTE.dotBlue}]`,
  },
  {
    bg: C.bgStatusGreen,
    text: C.textStatusGreen,
    border: "border-[#B2D8D0]",
    dot: `bg-[${PALETTE.dotGreen}]`,
  },
  {
    bg: C.bgDiscountLight,
    text: C.textDiscount,
    border: "border-[#F0C9A8]",
    dot: `bg-[${PALETTE.dotOrange}]`,
  },
  {
    bg: C.bgStatusPurple,
    text: C.textStatusPurple,
    border: "border-[#D6C6E1]",
    dot: `bg-[${PALETTE.dotPurple}]`,
  },
  {
    bg: C.bgRedLight,
    text: C.textNotionRed,
    border: "border-[#F5CBC4]",
    dot: `bg-[${PALETTE.dotRed}]`,
  },
  {
    bg: "bg-[#EEE0DA]",
    text: "text-[#64473A]",
    border: "border-[#DFD0C8]",
    dot: `bg-[${PALETTE.dotBrown}]`,
  },
  {
    bg: "bg-[#F5E0E9]",
    text: "text-[#AD1A72]",
    border: "border-[#ECCBDA]",
    dot: `bg-[${PALETTE.dotPink}]`,
  },
  {
    bg: C.bgNotice,
    text: C.textNotice,
    border: "border-[#F2DBA7]",
    dot: `bg-[${PALETTE.dotYellow}]`,
  },
];

// Known service type names → palette index
const KNOWN_TYPE_INDEX: Record<string, number> = {
  "診療": 0,
  "検診": 1,
  "検査": 1,
  "トリミング": 2,
  "ワクチン": 3,
  "手術": 4,
  "入院": 5,
  "ホテル": 6,
};

// Pre-computed set of palette indices used by known types
const KNOWN_PALETTE_INDICES = new Set(Object.values(KNOWN_TYPE_INDEX));

const DEFAULT_COLOR: ServiceTypeColor = {
  bg: "bg-[#F1F0EE]",
  text: "text-[#9B9A97]",
  border: "border-[#E3E2E0]",
  dot: `bg-[${PALETTE.dotGray}]`,
};

/**
 * Returns a color map derived from the active serviceType master items.
 * Colors are assigned deterministically: known types use canonical colors,
 * unknown types rotate through COLOR_PALETTES.
 */
export function useServiceTypeColorMap() {
  const { data: serviceTypes } = useMasterItems("serviceType");

  const activeEntries = useMemo<ServiceTypeEntry[]>(() => {
    const active = serviceTypes.filter((s) => s.status === "active");
    let paletteIndex = 0;

    return active.map((item) => {
      const knownIdx = KNOWN_TYPE_INDEX[item.name];
      let color: ServiceTypeColor;

      if (knownIdx !== undefined) {
        color = COLOR_PALETTES[knownIdx];
      } else {
        // Rotate through palettes for unknown types (skip known indices)
        while (KNOWN_PALETTE_INDICES.has(paletteIndex % COLOR_PALETTES.length)) {
          paletteIndex++;
        }
        color = COLOR_PALETTES[paletteIndex % COLOR_PALETTES.length] ?? DEFAULT_COLOR;
        paletteIndex++;
      }

      return { name: item.name, color };
    });
  }, [serviceTypes]);

  const colorMap = useMemo(
    () => new Map(activeEntries.map((e) => [e.name, e.color])),
    [activeEntries]
  );

  const getColor = useCallback(
    (typeName: string): ServiceTypeColor => colorMap.get(typeName) ?? DEFAULT_COLOR,
    [colorMap]
  );

  return { activeEntries, getColor };
}
