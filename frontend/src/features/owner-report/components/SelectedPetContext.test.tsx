import { render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

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
    render(<SelectedPetContext pet={makePet({ name: "", species: "", status: "生存" })} />);

    expect(screen.getByRole("heading", { name: "-" })).toBeInTheDocument();
    expect(screen.queryByText("死亡")).not.toBeInTheDocument();
  });
});

describe("SelectedPetContext 誕生月バッジ (EMR-301)", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-10T00:00:00+09:00"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("今月(JST)誕生月の生存ペットはペット名横に Cake バッジを表示する", () => {
    render(<SelectedPetContext pet={makePet({ birthDate: "2015-10-14", status: "生存" })} />);

    expect(screen.getByRole("button", { name: "今月誕生日" })).toBeInTheDocument();
  });

  it.each([
    ["誕生月でない", "2015-09-14" as const],
    ["birthDate 未登録", undefined],
  ])("%s のペットには Cake バッジを表示しない", (_label, birthDate) => {
    render(<SelectedPetContext pet={makePet({ birthDate, status: "生存" })} />);

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("死亡ペット（status=死亡）には誕生月でも Cake バッジを表示しない", () => {
    render(<SelectedPetContext pet={makePet({ birthDate: "2015-10-14", status: "死亡" })} />);

    expect(screen.getByText("死亡")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});
