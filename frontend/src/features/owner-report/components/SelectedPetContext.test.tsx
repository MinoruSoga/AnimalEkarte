import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import type { Pet } from "@/types";

import { SelectedPetContext } from "./SelectedPetContext";

const makePet = (overrides: Partial<Pet> = {}): Pet =>
  ({
    id: "1",
    ownerId: "42",
    name: "ポチ",
    species: "犬",
    ...overrides,
  }) as Pet;

describe("SelectedPetContext", () => {
  it("死亡ペットは文字付きの死亡表示と基本コンテキストを表示する", () => {
    render(
      <SelectedPetContext
        pet={makePet({
          status: "死亡",
          breed: "柴犬",
          birthDate: "2015-04-14",
          gender: "雄",
          weight: "8.2",
        })}
      />,
    );

    expect(screen.getByRole("heading", { name: "ポチ" })).toBeInTheDocument();
    expect(screen.getByText("死亡")).toBeInTheDocument();
    expect(screen.getByText("犬・柴犬")).toBeInTheDocument();
    expect(screen.getByText(/雄 ・ 8\.2 kg/)).toBeInTheDocument();
  });

  it("生存ペットでは死亡表示を出さず、未登録情報を捏造しない", () => {
    render(
      <SelectedPetContext
        pet={makePet({ name: "", species: "", status: "生存" })}
      />,
    );

    expect(screen.getByRole("heading", { name: "-" })).toBeInTheDocument();
    expect(screen.queryByText("死亡")).not.toBeInTheDocument();
  });
});
