import { useState } from "react";
import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import { SearchableSelect, type SearchableSelectGroup } from "./searchable-select";

const GROUPS: SearchableSelectGroup[] = [
  {
    label: "診察",
    options: [
      { value: "1", label: "一般診療" },
      { value: "2", label: "再診" },
    ],
  },
  {
    label: "トリミング",
    options: [
      { value: "3", label: "シャンプーコース" },
      { value: "4", label: "フルコース" },
    ],
  },
];

/** 制御コンポーネントを駆動する最小ハーネス */
function Harness({ onChange }: { onChange?: (v: string) => void }) {
  const [value, setValue] = useState("");
  return (
    <SearchableSelect
      value={value}
      onValueChange={(v) => {
        setValue(v);
        onChange?.(v);
      }}
      groups={GROUPS}
      groupFilter
      triggerTestId="ss-trigger"
      searchPlaceholder="検索..."
    />
  );
}

describe("SearchableSelect — groupFilter(カテゴリチップ)", () => {
  it("開くとカテゴリチップと全グループのオプションが表示される", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.click(screen.getByTestId("ss-trigger"));

    // カテゴリチップ(グループラベルのボタン)
    await waitFor(() => {
      expect(screen.getByRole("button", { name: "診察" })).toBeInTheDocument();
    });
    expect(screen.getByRole("button", { name: "トリミング" })).toBeInTheDocument();

    // 全グループのオプションが見える
    expect(screen.getByRole("option", { name: "一般診療" })).toBeInTheDocument();
    expect(screen.getByRole("option", { name: "シャンプーコース" })).toBeInTheDocument();
  });

  it("チップをクリックすると当該グループのみに絞り込まれ、解除で全表示に戻る", async () => {
    const user = userEvent.setup();
    render(<Harness />);

    await user.click(screen.getByTestId("ss-trigger"));
    await waitFor(() => expect(screen.getByRole("button", { name: "トリミング" })).toBeInTheDocument());

    // 「トリミング」チップで絞り込み
    fireEvent.click(screen.getByRole("button", { name: "トリミング" }));

    await waitFor(() => {
      expect(screen.queryByRole("option", { name: "一般診療" })).not.toBeInTheDocument();
    });
    expect(screen.getByRole("option", { name: "シャンプーコース" })).toBeInTheDocument();

    // 「解除」で全表示に戻る
    fireEvent.click(screen.getByRole("button", { name: "解除" }));
    await waitFor(() => {
      expect(screen.getByRole("option", { name: "一般診療" })).toBeInTheDocument();
    });
  });

  it("オプション選択で onValueChange が呼ばれる", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(<Harness onChange={onChange} />);

    await user.click(screen.getByTestId("ss-trigger"));
    fireEvent.click(await screen.findByRole("option", { name: "フルコース" }));

    expect(onChange).toHaveBeenCalledWith("4");
  });
});
