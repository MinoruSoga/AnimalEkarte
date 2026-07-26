import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { C } from "@/lib/design-tokens";
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

const DANGER_REASON = "診察時に咬傷歴あり";

function renderHighDangerPet(dangerReason: string | undefined) {
  render(
    <PetSelectionResultsTable
      pets={[{ ...PET, dangerLevel: "高", dangerReason }]}
      onSelect={vi.fn()}
    />,
  );

  return screen.getByRole("button", {
    name: "ポチの危険理由を表示",
  });
}

describe("PetSelectionResultsTable row actions", () => {
  it("危険度が高の個体だけ非色警告を表示する", () => {
    render(
      <PetSelectionResultsTable
        pets={[
          { ...PET, id: "pet-high", name: "ハイ", dangerLevel: "高" },
          { ...PET, id: "pet-medium", name: "ミドル", dangerLevel: "中" },
          { ...PET, id: "pet-unknown", name: "未判定" },
        ]}
        onSelect={vi.fn()}
      />,
    );

    expect(screen.getAllByText("⚠ 危険")).toHaveLength(1);
    expect(screen.getByRole("button", { name: "ハイの危険理由を表示" })).toHaveClass(
      C.bgDanger10,
      C.danger,
      C.borderDanger20,
    );
    expect(screen.getByText("ハイ").parentElement).toHaveTextContent("⚠ 危険");
    expect(screen.getByText("ミドル").parentElement).not.toHaveTextContent(
      "⚠ 危険",
    );
    expect(screen.getByText("未判定").parentElement).not.toHaveTextContent(
      "⚠ 危険",
    );
  });

  it("保存済みの高危険度理由をclickで開き、同じtriggerの再clickで閉じる", async () => {
    const user = userEvent.setup();
    const trigger = renderHighDangerPet(DANGER_REASON);

    expect(trigger.tagName).toBe("BUTTON");
    expect(trigger).toHaveAttribute("type", "button");
    expect(trigger).toHaveAttribute("aria-expanded", "false");

    await user.click(trigger);
    expect(await screen.findByText(DANGER_REASON)).toBeInTheDocument();
    expect(trigger).toHaveAttribute("aria-expanded", "true");
    expect(trigger).toHaveAttribute("aria-controls");

    await user.click(trigger);
    await waitFor(() => {
      expect(screen.queryByText(DANGER_REASON)).not.toBeInTheDocument();
    });
    expect(trigger).toHaveAttribute("aria-expanded", "false");
  });

  it.each([
    ["undefined", undefined],
    ["空文字", ""],
    ["空白のみ", "   "],
  ])(
    "高危険度理由が%sならclickで理由未登録を表示する",
    async (_caseName, dangerReason) => {
      const user = userEvent.setup();
      const trigger = renderHighDangerPet(dangerReason);

      await user.click(trigger);

      expect(await screen.findByText("理由未登録")).toBeInTheDocument();
    },
  );

  it.each([
    ["Enter", "{Enter}"],
    ["Space", " "],
  ])(
    "%sで高危険度理由を開き、同じキーで閉じる",
    async (_keyName, key) => {
      const user = userEvent.setup();
      const trigger = renderHighDangerPet(DANGER_REASON);

      trigger.focus();
      await user.keyboard(key);
      expect(await screen.findByText(DANGER_REASON)).toBeInTheDocument();

      trigger.focus();
      await user.keyboard(key);
      await waitFor(() => {
        expect(screen.queryByText(DANGER_REASON)).not.toBeInTheDocument();
      });
    },
  );

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

  it("死亡個体は非色依存の名前で選択不可になる", async () => {
    const user = userEvent.setup();
    const onSelect = vi.fn();
    render(
      <PetSelectionResultsTable
        pets={[{ ...PET, status: "死亡" }]}
        onSelect={onSelect}
      />,
    );

    const selectButton = screen.getByRole("button", {
      name: "死亡・選択不可: ポチ (ID pet-1)",
    });
    expect(selectButton).toBeDisabled();
    expect(selectButton).toHaveTextContent("選択不可");
    expect(screen.getByText("死亡")).toBeInTheDocument();
    expect(selectButton.closest("tr")).toHaveClass(
      "opacity-60",
      "grayscale-[0.5]",
    );
    await user.click(selectButton);
    expect(onSelect).not.toHaveBeenCalled();
  });
});
