import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CategoryChipsFilter } from "./CategoryChipsFilter";

const CATEGORIES = ["診察", "検査", "処置"];

describe("CategoryChipsFilter", () => {
  it("categories が空配列のときチップが1つも描画されずクラッシュしない", () => {
    render(
      <CategoryChipsFilter categories={[]} activeCategory={null} onSelectCategory={() => {}} />,
    );
    expect(screen.queryAllByRole("button")).toHaveLength(0);
  });

  it("activeCategory が null のとき解除チップは表示されない", () => {
    render(
      <CategoryChipsFilter
        categories={CATEGORIES}
        activeCategory={null}
        onSelectCategory={() => {}}
      />,
    );
    expect(screen.queryByRole("button", { name: /解除/ })).not.toBeInTheDocument();
  });

  it("全カテゴリチップが表示される", () => {
    render(
      <CategoryChipsFilter
        categories={CATEGORIES}
        activeCategory={null}
        onSelectCategory={() => {}}
      />,
    );
    for (const category of CATEGORIES) {
      expect(screen.getByRole("button", { name: category })).toBeInTheDocument();
    }
  });

  it("非選択チップをクリックするとそのカテゴリで onSelectCategory が呼ばれる", async () => {
    const user = userEvent.setup();
    const onSelectCategory = vi.fn();
    render(
      <CategoryChipsFilter
        categories={CATEGORIES}
        activeCategory={null}
        onSelectCategory={onSelectCategory}
      />,
    );

    await user.click(screen.getByRole("button", { name: "検査" }));

    expect(onSelectCategory).toHaveBeenCalledWith("検査");
  });

  it("選択中チップを再クリックすると null で onSelectCategory が呼ばれる (トグル解除)", async () => {
    const user = userEvent.setup();
    const onSelectCategory = vi.fn();
    render(
      <CategoryChipsFilter
        categories={CATEGORIES}
        activeCategory="検査"
        onSelectCategory={onSelectCategory}
      />,
    );

    await user.click(screen.getByRole("button", { name: "検査" }));

    expect(onSelectCategory).toHaveBeenCalledWith(null);
  });

  it("選択中は解除チップが表示され、クリックで null が呼ばれる", async () => {
    const user = userEvent.setup();
    const onSelectCategory = vi.fn();
    render(
      <CategoryChipsFilter
        categories={CATEGORIES}
        activeCategory="診察"
        onSelectCategory={onSelectCategory}
      />,
    );

    const clearButton = screen.getByRole("button", { name: /解除/ });
    expect(clearButton).toBeInTheDocument();
    await user.click(clearButton);

    expect(onSelectCategory).toHaveBeenCalledWith(null);
  });

  it("選択中チップに aria-pressed=true が設定される", () => {
    render(
      <CategoryChipsFilter
        categories={CATEGORIES}
        activeCategory="処置"
        onSelectCategory={() => {}}
      />,
    );
    expect(screen.getByRole("button", { name: "処置" })).toHaveAttribute("aria-pressed", "true");
    expect(screen.getByRole("button", { name: "診察" })).toHaveAttribute("aria-pressed", "false");
  });
});
