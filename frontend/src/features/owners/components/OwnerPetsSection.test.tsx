import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { createTestWrapper } from "@/testing/utils";
import type { PetFormData } from "../types";
import { OwnerPetsSection } from "./OwnerPetsSection";

const PET = {
  id: "pet-1",
  petNumber: "P001",
  petName: "ポチ",
  status: "生存",
  species: "犬",
  gender: "雄",
  birthDate: "2020-01-01",
  color: "茶",
  weight: "5",
  environment: "室内",
  remarks: "",
} satisfies PetFormData;

function renderSection(canEdit: boolean, onEditPet = vi.fn()) {
  render(
    <OwnerPetsSection
      pets={[PET]}
      ownerId="owner-1"
      canEdit={canEdit}
      canCreate={false}
      canDelete={false}
      onAddPet={vi.fn()}
      onEditPet={onEditPet}
      onDeleteRequest={vi.fn()}
    />,
    { wrapper: createTestWrapper({ router: true }) },
  );
  return onEditPet;
}

describe("OwnerPetsSection row actions", () => {
  it("行は編集せず、権限時だけ固有名の44px native buttonで編集する", async () => {
    const user = userEvent.setup();
    const onEditPet = renderSection(true);

    await user.click(screen.getByText("犬"));
    expect(onEditPet).not.toHaveBeenCalled();

    const editButton = screen.getByRole("button", {
      name: "詳細・編集: ペット ポチ (ID pet-1)",
    });
    expect(editButton.tagName).toBe("BUTTON");
    expect(editButton).toHaveClass("min-h-11", "min-w-11");
    await user.click(editButton);
    expect(onEditPet).toHaveBeenCalledWith(PET);
  });

  it("編集権限がなければ編集buttonを表示しない", () => {
    renderSection(false);
    expect(
      screen.queryByRole("button", { name: "詳細・編集: ペット ポチ (ID pet-1)" }),
    ).not.toBeInTheDocument();
  });
});
