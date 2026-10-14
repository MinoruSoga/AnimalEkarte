import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MasterSelectTrigger } from "./MasterSelectTrigger";

const selectedCourse = { name: "プレミアムトリミングコース", price: 12000 };

describe("MasterSelectTrigger — 変更ラベルの折返し防止 (EMR-392/VQ-016)", () => {
  // VQ-016: 選択済みコース名が幅を取り、「変更」ラベルが2行に割れていた。
  // 変更ラベル+アイコンの compact グループを shrink-0/nowrap で1行に保ち、
  // 名称本文の折返し・truncate 方針は variant ごとの現状を維持する。
  it.each(["inline", "block"] as const)(
    "%s variant: 変更ラベル+アイコンのグループは shrink-0/whitespace-nowrap で1行に保つ",
    (variant) => {
      render(
        <MasterSelectTrigger
          selectedItem={selectedCourse}
          placeholder="コースを選択"
          icon={<span>icon</span>}
          onClick={() => {}}
          variant={variant}
        />,
      );

      const changeGroup = screen.getByText("変更").parentElement;
      expect(changeGroup).toHaveClass("shrink-0", "whitespace-nowrap");
      // 名称・価格は切り詰めたり削除したりしない
      expect(screen.getByText(selectedCourse.name)).toBeInTheDocument();
      expect(screen.getByText("¥12,000")).toBeInTheDocument();
    },
  );

  it("block variant の名称本文は折返し可能なまま（truncate させない）", () => {
    render(
      <MasterSelectTrigger
        selectedItem={selectedCourse}
        placeholder="コースを選択"
        icon={<span>icon</span>}
        onClick={() => {}}
        variant="block"
      />,
    );

    expect(screen.getByText(selectedCourse.name)).not.toHaveClass("truncate");
  });

  it("クリックで onClick が呼ばれる（hit target を維持）", async () => {
    const user = userEvent.setup();
    const onClick = vi.fn();
    render(
      <MasterSelectTrigger
        selectedItem={selectedCourse}
        placeholder="コースを選択"
        icon={<span>icon</span>}
        onClick={onClick}
        variant="block"
      />,
    );

    await user.click(screen.getByRole("button"));
    expect(onClick).toHaveBeenCalledTimes(1);
  });
});
