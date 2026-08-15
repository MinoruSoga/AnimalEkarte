import { describe, expect, it } from "vitest";

import { C, LAYOUT, PALETTE, STYLE } from "./design-tokens";

function classesOf(value: string): Set<string> {
  return new Set(value.split(/\s+/));
}

function expectClasses(value: string, expected: readonly string[]): void {
  const classes = classesOf(value);

  for (const expectedClass of expected) {
    expect(classes.has(expectedClass), `missing class: ${expectedClass}`).toBe(true);
  }
}

function expectNoClasses(value: string, forbidden: readonly string[]): void {
  const classes = classesOf(value);

  for (const forbiddenClass of forbidden) {
    expect(classes.has(forbiddenClass), `conflicting class: ${forbiddenClass}`).toBe(false);
  }
}

describe("brand and primary color tokens", () => {
  it("design-system.md の brand と semantic primary を同じ teal に統一する", () => {
    expect(PALETTE.brand).toBe("#038B94");
    expect(PALETTE.brandHover).toBe("#027078");
    expect(PALETTE.actionPrimary).toBe(PALETTE.brand);
    expect(PALETTE.actionPrimaryActive).toBe(PALETTE.brandHover);
    expect(PALETTE.brandLight).toBe("#E1F3F4");
    expect(PALETTE.brandDark).toBe("#025F66");
    expect(PALETTE.actionPrimaryLight).toBe(PALETTE.brandLight);
    expect(PALETTE.actionPrimaryDark).toBe(PALETTE.brandDark);
    expect(C.bgBrandIdentity).toBe("bg-[#038B94]");
    expect(C.textOnBrand).toBe("text-white");
    expect(C.textBrandIdentity).toBe("text-[#025F66] dark:text-[#079BA5]");
    expect(C.hoverBgBrandIdentity).toBe("hover:bg-[#027078]");
    expect(C.bgBrandIdentityLight).toBe("bg-[#E1F3F4]");
    expect(C.textOnBrandIdentity).toBe("text-white");
    expect(C.hoverTextOnBrandIdentity).toBe("hover:text-white");
    expect(C.bgActionPrimary).toBe("bg-[#038B94]");
    expect(C.hoverBgActionPrimary).toBe("hover:bg-[#027078]");
    expect(C.textOnActionPrimary).toBe("text-white");
    expect(C.hoverTextOnActionPrimary).toBe("hover:text-white");
    expect(C.activeTextOnActionPrimary).toBe("active:text-white");
    expect(C.focusRingActionPrimary).toBe("focus:ring-[#038B94]");
    expect(C.dataCheckedBgActionPrimary).toBe("data-[state=checked]:bg-[#038B94]");
    expect(C.dataActiveBorderB).toBe("data-[state=active]:border-b-[#038B94]");
    expect(PALETTE.focusRingActionPrimary).toBe("focus:shadow-focus-primary");
    expect(C.dataActiveText).toBe(
      "data-[state=active]:text-[#025F66] dark:data-[state=active]:text-[#079BA5]",
    );
    expect(C.focusRingBrandIdentity).toBe("focus:ring-[#038B94]");
    expect(C.focusRingBrand40).toBe("focus:ring-[#038B94]");
    expect(C.focusRingAccent40).toBe("focus-visible:ring-[#038B94]");
    expect(C.focusRingAccent30).toBe("focus:ring-[#038B94]");
  });
});

describe("LAYOUT typography tokens", () => {
  it("pageTitle が DESIGN.md の heading-2 数値に一致する", () => {
    expect(LAYOUT.pageTitle).toEqual({
      fontSize: "26px",
      fontWeight: 700,
      lineHeight: "1.23",
      letterSpacing: "-0.625px",
    });
  });
});

describe("STYLE table tokens", () => {
  it("header は canvas-soft / hairline / eyebrow と 12px 16px padding を使う", () => {
    expectClasses(STYLE.tableHeaderRow, ["border-b", C.borderLight, C.bgPage]);
    // text-2xs は globals.css で 12px / 1.33 / +0.125px に固定されている。
    expectClasses(STYLE.tableHeaderCell, ["text-2xs", "font-semibold", "px-4", "py-3"]);
    expectNoClasses(STYLE.tableHeaderCell, ["px-2", "p-2"]);
  });

  it("body cell は body-sm と 12px 16px padding を使う", () => {
    // text-sm は globals.css で 15px / 1.33 に固定されている。
    expectClasses(STYLE.tableCell, ["text-sm", "font-normal", "px-4", "py-3"]);
    expectNoClasses(STYLE.tableCell, ["text-base", "p-2", "py-2.5"]);
  });
});

describe("STYLE icon button touch targets", () => {
  it.each([STYLE.iconBtn20, STYLE.iconBtn28, STYLE.iconBtn32])(
    "compact glyphを維持しながら44x44px以上のhit areaを持つ",
    (classes) => {
      expectClasses(classes, ["min-h-11", "min-w-11"]);
    },
  );
});
