import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { Tabs } from "@/components/ui/tabs";
import { STYLE, STATE } from "@/lib/design-tokens";

import { MedicalRecordMountedTab } from "./MedicalRecordTabsShared";

// EMR-412 X7/R7: 閲覧専用 fieldset は入力値を ink のまま保ち（opacity 1）、
// 境界を hairline へ落とす。一方で disabled アクションボタンの既定演出は変えない。
describe("MedicalRecordMountedTab readonly fieldset (EMR-412 X7)", () => {
  it("locked fieldset は STATE.readonlyInputs を合成し、ボタン.opacity は上書きしない", () => {
    render(
      <Tabs defaultValue="問診">
        <MedicalRecordMountedTab
          tab="問診"
          activeTab="問診"
          mountedTabs={new Set(["問診"])}
          isLocked
        >
          <input disabled defaultValue="所見" />
          <button type="button" disabled>
            操作
          </button>
        </MedicalRecordMountedTab>
      </Tabs>,
    );

    const fieldset = screen.getByRole("group");
    // 入力は opacity 1 + hairline 境界（STATE の完全リテラルを STYLE 経由で合成）。
    expect(fieldset.className).toContain(STATE.readonlyInputs);
    expect(fieldset.className).toContain(STYLE.readonlyInputs);
    expect(fieldset.className).toContain("[&_input:disabled]:opacity-100");
    expect(fieldset.className).toContain("[&_input:disabled]:border-[#E6E6E6]");
    // アクションボタンの disabled 演出（opacity 50）は共有 side のまま（劣化させない）。
    expect(fieldset.className).not.toContain("[&_button:disabled]:opacity-100");
    expect(fieldset.className).not.toContain("[&_button:disabled]:border-");
  });
});
