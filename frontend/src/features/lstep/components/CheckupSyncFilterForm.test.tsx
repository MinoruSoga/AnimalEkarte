import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useAnimalSpecies } from "@/hooks/use-animal-species";
import { describe, expect, it, vi } from "vitest";

import { CheckupSyncFilterForm } from "./CheckupSyncFilterForm";

vi.mock("@/hooks/use-animal-species", () => ({
  useAnimalSpecies: vi.fn(() => ({
    activeSpecies: [{ id: 99, name: "犬", is_active: true }],
    isLoading: false,
    isError: false,
  })),
}));

describe("species name search boundary", () => {
  it("selects the authoritative name, and omits blank species", async () => {
    const user = userEvent.setup();
    const onSearch = vi.fn();
    render(<CheckupSyncFilterForm onSearch={onSearch} isLoading={false} />);
    await user.selectOptions(screen.getByLabelText(/検診種別/), "annual");
    await user.selectOptions(screen.getByLabelText(/動物種/), "犬");
    await user.click(screen.getByRole("button", { name: "対象者を検索" }));
    await waitFor(() =>
      expect(onSearch).toHaveBeenLastCalledWith({ checkup_type: "annual", species: "犬" }),
    );
    await user.selectOptions(screen.getByLabelText(/検診種別/), "annual");
    await user.selectOptions(screen.getByLabelText(/動物種/), "");
    await user.click(screen.getByRole("button", { name: "対象者を検索" }));
    await waitFor(() => expect(onSearch).toHaveBeenLastCalledWith({ checkup_type: "annual" }));
  });
  it("failed lookup keeps other filters usable and explains the text fallback", () => {
    vi.mocked(useAnimalSpecies).mockReturnValueOnce({
      activeSpecies: [],
      isLoading: false,
      isError: true,
    } as ReturnType<typeof useAnimalSpecies>);
    render(<CheckupSyncFilterForm onSearch={vi.fn()} isLoading={false} />);
    expect(screen.getByLabelText(/動物種/)).not.toBeDisabled();
    expect(screen.getByLabelText(/検診種別/)).not.toBeDisabled();
    expect(
      screen.getByText("動物種マスタを取得できないため、直接入力できます"),
    ).toBeInTheDocument();
  });

  it("PR665: denied species lookup falls back to a usable text filter and submits", async () => {
    const user = userEvent.setup();
    const onSearch = vi.fn();
    vi.mocked(useAnimalSpecies).mockReturnValueOnce({
      activeSpecies: [],
      isLoading: false,
      isError: true,
    } as ReturnType<typeof useAnimalSpecies>);
    render(<CheckupSyncFilterForm onSearch={onSearch} isLoading={false} />);

    const speciesInput = screen.getByLabelText(/動物種/);
    expect(speciesInput.tagName).toBe("INPUT");
    expect(speciesInput).not.toBeDisabled();
    await user.type(speciesInput, "犬");
    await user.selectOptions(screen.getByLabelText(/検診種別/), "annual");
    await user.click(screen.getByRole("button", { name: "対象者を検索" }));
    await waitFor(() =>
      expect(onSearch).toHaveBeenLastCalledWith({ checkup_type: "annual", species: "犬" }),
    );
  });
});

// EMR-385: 「最終来院日（以降・任意）」の括弧内で「任/意」の1文字折返しが
// 発生していた。ラベルの日本語まとまり（主語部 / 括弧部）をそれぞれ
// 分割不能な単位にして、まとまり同士の境界でのみ折返せるようにする。
// jsdom は実レイアウトを評価しないため、ここでは「分割不能単位として
// レンダリングされる構造」と label-for/input-id 接続の維持を固定する。
const PAREN_LABELS: [labelText: string, inputId: string][] = [
  ["最終来院日（以降・任意）", "last_visit_after"],
  ["最終来院日（以前・任意）", "last_visit_before"],
  ["最終健診実施日（以降・任意）", "last_checkup_after"],
  ["最終健診実施日（以前・任意）", "last_checkup_before"],
];

describe("CheckupSyncFilterForm — EMR-385 ラベル折返し防止", () => {
  it.each(PAREN_LABELS)("%s のラベルはまとまり単位でしか折返せない", (labelText, inputId) => {
    render(<CheckupSyncFilterForm onSearch={vi.fn()} isLoading={false} />);

    // label htmlFor ↔ input id の接続が維持されている
    const input = screen.getByLabelText(labelText);
    expect(input).toHaveAttribute("id", inputId);

    const label = document.querySelector(`label[for="${inputId}"]`);
    expect(label).not.toBeNull();
    const nowrapParts = label!.querySelectorAll("span.whitespace-nowrap");
    expect(nowrapParts.length).toBeGreaterThanOrEqual(2);
    // 不分割単位の結合が元のラベル全文と一致する（括弧内での途中切れなし）
    expect([...nowrapParts].map((node) => node.textContent).join("")).toBe(labelText);
  });
});
