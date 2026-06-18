import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import type { Pet } from "@/types";

import { PetSwitcher, OWNER_REPORT_TABPANEL_ID, ownerReportPetTabId } from "./PetSwitcher";

// PetSwitcher は id/name/species のみ参照するため、Pet の全フィールドは不要。
// 焦点を絞った単体テストとして最小フィクスチャを生成する。
const makePet = (id: string, name: string, species = "犬"): Pet =>
  ({ id, name, species }) as unknown as Pet;

const pets = [makePet("1", "ポチ"), makePet("2", "タマ", "猫"), makePet("3", "ハチ")];

function setup(selectedPetId: string | undefined = "1") {
  const onSelect = vi.fn();
  const utils = render(
    <PetSwitcher pets={pets} selectedPetId={selectedPetId} onSelect={onSelect} />,
  );
  const tab = (name: RegExp) => screen.getByRole("tab", { name });
  return { onSelect, tab, ...utils };
}

describe("PetSwitcher キーボードアクセシビリティ (WAI-ARIA Tabs)", () => {
  it("選択中タブのみ tabIndex=0、他は -1（roving tabindex）。選択移動で tabbable も追従する", () => {
    const { tab, rerender } = setup("1");
    expect(tab(/ポチ/)).toHaveAttribute("tabindex", "0");
    expect(tab(/タマ/)).toHaveAttribute("tabindex", "-1");
    expect(tab(/ハチ/)).toHaveAttribute("tabindex", "-1");

    rerender(<PetSwitcher pets={pets} selectedPetId="2" onSelect={vi.fn()} />);
    expect(tab(/ポチ/)).toHaveAttribute("tabindex", "-1");
    expect(tab(/タマ/)).toHaveAttribute("tabindex", "0");
    expect(tab(/ハチ/)).toHaveAttribute("tabindex", "-1");
  });

  it("選択が一致しない場合は先頭タブが tabbable（tablist が Tab 到達不能にならないフォールバック）", () => {
    const { tab } = setup("999");
    expect(tab(/ポチ/)).toHaveAttribute("tabindex", "0");
    expect(tab(/タマ/)).toHaveAttribute("tabindex", "-1");
    expect(tab(/ハチ/)).toHaveAttribute("tabindex", "-1");
  });

  it("ArrowRight/ArrowDown は次タブへフォーカスのみ移動する（手動アクティベーション）", async () => {
    const user = userEvent.setup();
    const { tab, onSelect } = setup("1");

    tab(/ポチ/).focus();
    await user.keyboard("{ArrowRight}");
    expect(tab(/タマ/)).toHaveFocus();

    await user.keyboard("{ArrowDown}");
    expect(tab(/ハチ/)).toHaveFocus();

    // フォーカス移動だけでは選択しない（ページ遷移なし・onSelect 未発火）。
    expect(onSelect).not.toHaveBeenCalled();
  });

  it("ArrowLeft/ArrowUp は前タブへフォーカス移動する", async () => {
    const user = userEvent.setup();
    const { tab } = setup("3");

    tab(/ハチ/).focus();
    await user.keyboard("{ArrowLeft}");
    expect(tab(/タマ/)).toHaveFocus();

    await user.keyboard("{ArrowUp}");
    expect(tab(/ポチ/)).toHaveFocus();
  });

  it("末尾→先頭 / 先頭→末尾 にフォーカスが巻き戻る（wrap）", async () => {
    const user = userEvent.setup();
    const { tab } = setup("1");

    tab(/ハチ/).focus(); // 末尾
    await user.keyboard("{ArrowRight}");
    expect(tab(/ポチ/)).toHaveFocus(); // 先頭へ wrap

    tab(/ポチ/).focus(); // 先頭
    await user.keyboard("{ArrowLeft}");
    expect(tab(/ハチ/)).toHaveFocus(); // 末尾へ wrap
  });

  it("Home/End で先頭/末尾タブへフォーカスする", async () => {
    const user = userEvent.setup();
    const { tab } = setup("1");

    tab(/タマ/).focus();
    await user.keyboard("{Home}");
    expect(tab(/ポチ/)).toHaveFocus();

    await user.keyboard("{End}");
    expect(tab(/ハチ/)).toHaveFocus();
  });

  it("フォーカス中タブを Enter/Space で activation し onSelect が発火する", async () => {
    const user = userEvent.setup();
    const { tab, onSelect } = setup("1");

    tab(/ポチ/).focus();
    await user.keyboard("{ArrowRight}"); // タマ（未選択）へフォーカス
    await user.keyboard("{Enter}");
    expect(onSelect).toHaveBeenCalledWith("2");

    onSelect.mockClear();
    tab(/ポチ/).focus();
    await user.keyboard("{ArrowRight}{ArrowRight}"); // ハチへフォーカス
    await user.keyboard(" "); // Space
    expect(onSelect).toHaveBeenCalledWith("3");
  });

  it("クリックでも従来通り選択できる", async () => {
    const user = userEvent.setup();
    const { tab, onSelect } = setup("1");

    await user.click(tab(/タマ/));
    expect(onSelect).toHaveBeenCalledWith("2");
  });

  it("tablist/tab/aria-selected/aria-controls/id の ARIA セマンティクスを保持する", () => {
    const { tab } = setup("1");

    expect(screen.getByRole("tablist", { name: "ペット切替" })).toBeInTheDocument();
    expect(tab(/ポチ/)).toHaveAttribute("aria-selected", "true");
    expect(tab(/タマ/)).toHaveAttribute("aria-selected", "false");
    expect(tab(/ポチ/)).toHaveAttribute("aria-controls", OWNER_REPORT_TABPANEL_ID);
    expect(tab(/ポチ/)).toHaveAttribute("id", ownerReportPetTabId("1"));
  });
});
