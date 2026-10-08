import { render, screen, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { Pet } from "@/types";
import { PatientSelectionResults } from "./PatientSelectionResults";

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

const BASE_PET = {
  id: "pet-1",
  ownerId: "owner-1",
  ownerNumber: 1,
  ownerName: "山田 太郎",
  phone: "",
  name: "ポチ",
  species: "犬",
  gender: "雄",
  status: "生存",
} satisfies Pet;

function renderResults(pet: Pet) {
  render(
    <PatientSelectionResults
      hasSearchConditions={true}
      isSearchPending={false}
      error={null}
      isBusy={false}
      pets={[pet]}
      selectedPetIds={new Set()}
      onSelect={vi.fn()}
    />,
  );
}

describe("PatientSelectionResults 誕生月バッジ (EMR-301)", () => {
  beforeEach(() => {
    // 現在JST月を 2026-10 に固定して決定的にする
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-10-10T00:00:00+09:00"));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("今月(JST)誕生月のペット行はペット名セルに Cake バッジを表示する", () => {
    renderResults({ ...BASE_PET, birthDate: "2018-10-05" });

    expect(screen.getByRole("button", { name: "今月誕生日" })).toBeInTheDocument();
  });

  it.each([
    ["誕生月でない", "2018-09-05"],
    ["birthDate 未登録", undefined],
  ])("%s のペットには Cake バッジを表示しない", (_label, birthDate) => {
    renderResults({ ...BASE_PET, birthDate });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });

  it("死亡ペット（status=死亡）には誕生月でも Cake バッジを表示しない", () => {
    renderResults({ ...BASE_PET, status: "死亡", birthDate: "2018-10-05" });

    expect(screen.queryByRole("button", { name: "今月誕生日" })).not.toBeInTheDocument();
  });
});
