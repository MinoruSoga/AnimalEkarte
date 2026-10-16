import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { PastRecordHistoryPanel } from "./PastRecordHistoryPanel";
import type { PastRecordHistoryItem } from "./PastRecordHistoryPanel";

const item: PastRecordHistoryItem = {
  id: "record-42",
  date: "2026-10-13",
  title: "定期診察",
  subtitle: "経過良好",
};

describe("PastRecordHistoryPanel 履歴の操作 (EMR-412)", () => {
  it("コールバックがある履歴は native button で元の項目を返す", async () => {
    const user = userEvent.setup();
    const onItemClick = vi.fn<(record: PastRecordHistoryItem) => void>();
    render(
      <PastRecordHistoryPanel
        title="診療履歴"
        searchPlaceholder="履歴を検索"
        items={[item]}
        onItemClick={onItemClick}
      />,
    );

    const button = screen.getByRole("button", { name: /定期診察/ });
    expect(button.tagName).toBe("BUTTON");
    expect(button).toHaveAttribute("type", "button");
    await user.click(button);

    expect(onItemClick).toHaveBeenCalledTimes(1);
    expect(onItemClick.mock.calls[0]?.[0]).toBe(item);
  });

  it("コールバック未指定の履歴は表示のみで操作ボタンにしない", () => {
    render(
      <PastRecordHistoryPanel title="診療履歴" searchPlaceholder="履歴を検索" items={[item]} />,
    );

    expect(screen.getByText("定期診察")).toBeInTheDocument();
    expect(screen.getByText("経過良好")).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: /定期診察/ })).not.toBeInTheDocument();
  });
});
