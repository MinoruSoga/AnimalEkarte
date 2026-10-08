import { render, screen, within } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { PatientSelectionResults } from "./PatientSelectionResults";
import type { Pet } from "@/types";

describe("EMR-318 patient result scrolling", () => {
  it("keeps all returned patients in one keyboard accessible shrinking scroll region", () => {
    const pets = Array.from(
      { length: 20 },
      (_, i) =>
        ({
          id: String(i),
          name: `患者${i}`,
          ownerId: "1",
          ownerName: "飼主",
          species: "犬",
          status: "生存",
        }) as Pet,
    );
    render(
      <PatientSelectionResults
        hasSearchConditions
        isSearchPending={false}
        error={null}
        isBusy={false}
        pets={pets}
        selectedPetIds={new Set()}
        onSelect={vi.fn()}
      />,
    );
    const region = screen.getByRole("region", { name: "患者検索結果" });
    expect(region).toHaveAttribute("tabindex", "0");
    expect(region).toHaveClass("min-h-0", "overflow-auto");
    expect(within(region).getAllByRole("button")).toHaveLength(20);
    expect(within(region).getByRole("button", { name: "選択: 患者19 (ID 19)" })).toBeEnabled();
  });
});
