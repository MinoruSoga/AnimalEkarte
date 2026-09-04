import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import type { Pet } from "@/types";

import { PetSwitcher } from "./PetSwitcher";

const makePet = (id: string, name: string, species = "犬"): Pet =>
  ({ id, name, species }) as unknown as Pet;

const pets = [makePet("1", "ポチ"), makePet("2", "タマ", "猫"), makePet("3", "ハチ")];

function setup(selectedPetId: string | undefined = "1") {
  const onSelect = vi.fn();
  render(<PetSwitcher pets={pets} selectedPetId={selectedPetId} onSelect={onSelect} />);
  return { onSelect, select: screen.getByRole("combobox", { name: "ペット切替" }) };
}

describe("PetSwitcher", () => {
  it("タブではなくコンパクトな選択欄として全ペットを表示する", () => {
    const { select } = setup("1");

    expect(select).toHaveValue("1");
    expect(select).toHaveAttribute("id", "owner-report-pet-switcher");
    expect(select).toHaveAttribute("name", "petId");
    expect(screen.getByText("ペット切替")).toHaveAttribute("for", "owner-report-pet-switcher");
    expect(screen.getByRole("option", { name: "ポチ（犬）" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "タマ（猫）" })).toBeInTheDocument();
    expect(screen.queryByRole("tablist")).not.toBeInTheDocument();
    expect(screen.queryByRole("tab")).not.toBeInTheDocument();
  });

  it("選択変更時にペットIDを通知する", async () => {
    const user = userEvent.setup();
    const { onSelect, select } = setup("1");

    await user.selectOptions(select, "2");

    expect(onSelect).toHaveBeenCalledWith("2");
  });

  it("選択IDが不正な場合は先頭ペットへフォールバックする", () => {
    const { select } = setup("999");

    expect(select).toHaveValue("1");
  });

  it("名称と種別が未登録でも空の選択肢にしない", () => {
    const onSelect = vi.fn();
    render(<PetSwitcher pets={[makePet("4", "", "")]} selectedPetId="4" onSelect={onSelect} />);

    expect(screen.getByRole("option", { name: "-" })).toBeInTheDocument();
  });

  it("ペットが0件なら未選択のselectとして表示する", () => {
    const onSelect = vi.fn();
    render(<PetSwitcher pets={[]} selectedPetId={undefined} onSelect={onSelect} />);

    const select = screen.getByRole("combobox", { name: "ペット切替" });
    expect((select as HTMLSelectElement).value).toBe("");
    expect(screen.queryByRole("option")).not.toBeInTheDocument();
  });
});
