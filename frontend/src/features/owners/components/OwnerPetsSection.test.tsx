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

interface RenderSectionOptions {
  pet?: PetFormData;
  canEdit?: boolean;
  canCreate?: boolean;
  canDelete?: boolean;
  onEditPet?: (pet: PetFormData) => void;
  onDeleteRequest?: (id: string, name: string) => void;
}

function renderSection({
  pet = PET,
  canEdit = true,
  canCreate = false,
  canDelete = false,
  onEditPet = vi.fn(),
  onDeleteRequest = vi.fn(),
}: RenderSectionOptions = {}) {
  render(
    <OwnerPetsSection
      pets={[pet]}
      ownerId="owner-1"
      canEdit={canEdit}
      canCreate={canCreate}
      canDelete={canDelete}
      onAddPet={vi.fn()}
      onEditPet={onEditPet}
      onDeleteRequest={onDeleteRequest}
    />,
    { wrapper: createTestWrapper({ router: true }) },
  );
  return { onEditPet, onDeleteRequest };
}

describe("OwnerPetsSection row actions", () => {
  it("行は編集せず、権限時だけ固有名の44px native buttonで編集する", async () => {
    const user = userEvent.setup();
    const { onEditPet } = renderSection();

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
    renderSection({ canEdit: false });
    expect(
      screen.queryByRole("button", { name: "詳細・編集: ペット ポチ (ID pet-1)" }),
    ).not.toBeInTheDocument();
  });

  it("作成権限があれば編集権限なしでもペット追加buttonを表示する", () => {
    renderSection({ canEdit: false, canCreate: true });

    expect(screen.getByRole("button", { name: "ペット追加" })).toBeInTheDocument();
  });

  it("生存ペットは既存の5つの新規業務遷移と削除操作を表示する", async () => {
    const user = userEvent.setup();
    renderSection({ canCreate: true, canDelete: true });

    await user.click(
      screen.getByRole("button", {
        name: "操作メニュー: ペット ポチ (ID pet-1)",
      }),
    );

    for (const name of ["予約作成", "カルテ作成", "トリミング", "入院登録", "会計登録", "削除"]) {
      expect(screen.getByRole("menuitem", { name })).toBeInTheDocument();
    }
  });

  it("死亡ペットは編集を維持しつつ5つの新規業務遷移と破壊的削除を表示しない", async () => {
    const user = userEvent.setup();
    renderSection({
      pet: { ...PET, status: "死亡", deceasedAt: "2026-07-11T00:00:00+09:00" },
      canCreate: true,
      canDelete: true,
    });

    await user.click(
      screen.getByRole("button", {
        name: "操作メニュー: ペット ポチ (ID pet-1)",
      }),
    );

    expect(screen.getByRole("menuitem", { name: "詳細・編集" })).toBeInTheDocument();
    for (const name of ["予約作成", "カルテ作成", "トリミング", "入院登録", "会計登録", "削除"]) {
      expect(screen.queryByRole("menuitem", { name })).not.toBeInTheDocument();
    }
  });
});
