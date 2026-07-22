import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import type { Pet } from "@/types";
import { PetSelectionResultsTable } from "./PetSelectionResultsTable";

const PET = {
  id: "pet-1",
  ownerId: "owner-1",
  ownerNumber: 1,
  ownerName: "山田 太郎",
  phone: "",
  name: "ポチ",
  species: "犬",
  gender: "雄",
} satisfies Pet;

describe("PetSelectionResultsTable row actions", () => {
  it("行は選択せず、固有名の44px native buttonだけが選択する", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(<PetSelectionResultsTable pets={[PET]} onSelect={onSelect} />);

    await user.click(screen.getByText("山田 太郎"));
    expect(onSelect).not.toHaveBeenCalled();

    const selectButton = screen.getByRole("button", { name: "選択: ポチ (ID pet-1)" });
    expect(selectButton.tagName).toBe("BUTTON");
    expect(selectButton).toHaveClass("min-h-11", "min-w-11");
    await user.click(selectButton);
    expect(onSelect).toHaveBeenCalledWith(PET);
  });

  it("死亡個体は非色依存の名前で選択不可になる", () => {
    render(
      <PetSelectionResultsTable
        pets={[{ ...PET, status: "死亡" }]}
        onSelect={vi.fn()}
      />,
    );

    expect(
      screen.getByRole("button", { name: "死亡・選択不可: ポチ (ID pet-1)" }),
    ).toBeDisabled();
  });
});
