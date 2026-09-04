import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import {
  ReservationTypePickerDialog,
  type ReservationTypePickerGroup,
} from "./ReservationTypePickerDialog";

const GROUPS: ReservationTypePickerGroup[] = [
  {
    label: "診察",
    items: [
      { id: "1", name: "一般診療", color: "#111111", durationMinutes: 20 },
      { id: "2", name: "再診", color: "#222222", durationMinutes: 10 },
    ],
  },
  {
    label: "トリミング",
    items: [{ id: "3", name: "シャンプーコース", color: "#333333", durationMinutes: 60 }],
  },
];

function renderDialog(overrides?: {
  onSelect?: (id: string) => void;
  onOpenChange?: (o: boolean) => void;
}) {
  return render(
    <ReservationTypePickerDialog
      open
      onOpenChange={overrides?.onOpenChange ?? (() => {})}
      groups={GROUPS}
      selectedId=""
      onSelect={overrides?.onSelect ?? (() => {})}
    />,
  );
}

describe("ReservationTypePickerDialog", () => {
  it("開くとカテゴリチップと全グループのカードが表示される", async () => {
    renderDialog();

    expect(await screen.findByRole("button", { name: "診察" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "トリミング" })).toBeInTheDocument();

    expect(screen.getByTestId("res-type-card-1")).toBeInTheDocument();
    expect(screen.getByTestId("res-type-card-3")).toBeInTheDocument();
    // 所要時間が表示される
    expect(screen.getByText("約20分")).toBeInTheDocument();
  });

  it("カテゴリチップで当該グループのみに絞り込まれる", async () => {
    renderDialog();
    fireEvent.click(await screen.findByRole("button", { name: "トリミング" }));

    await waitFor(() => {
      expect(screen.queryByTestId("res-type-card-1")).not.toBeInTheDocument();
    });
    expect(screen.getByTestId("res-type-card-3")).toBeInTheDocument();
    // 解除ボタンが現れる
    expect(screen.getByRole("button", { name: "解除" })).toBeInTheDocument();
  });

  it("検索語で名前一致のカードのみ表示される", async () => {
    renderDialog();
    fireEvent.change(await screen.findByPlaceholderText("予約区分を検索..."), {
      target: { value: "再診" },
    });

    await waitFor(() => {
      expect(screen.queryByTestId("res-type-card-1")).not.toBeInTheDocument();
    });
    expect(screen.getByTestId("res-type-card-2")).toBeInTheDocument();
  });

  it("カードをクリックすると onSelect が呼ばれダイアログを閉じる", async () => {
    const onSelect = vi.fn();
    const onOpenChange = vi.fn();
    renderDialog({ onSelect, onOpenChange });

    fireEvent.click(await screen.findByTestId("res-type-card-3"));

    expect(onSelect).toHaveBeenCalledWith("3");
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it("検索後にクリアボタンで全件表示に戻る", async () => {
    const user = userEvent.setup();
    renderDialog();

    const input = await screen.findByPlaceholderText("予約区分を検索...");
    await user.type(input, "再診");
    expect(input).toHaveValue("再診");
    expect(screen.queryByTestId("res-type-card-1")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "検索をクリア" }));

    expect(input).toHaveValue("");
    expect(screen.getByTestId("res-type-card-1")).toBeInTheDocument();
    expect(screen.getByTestId("res-type-card-2")).toBeInTheDocument();
    expect(screen.getByTestId("res-type-card-3")).toBeInTheDocument();
  });
});
