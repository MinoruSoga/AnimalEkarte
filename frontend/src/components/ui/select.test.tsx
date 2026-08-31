import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogTitle,
} from "./dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "./select";

describe("Select inside Dialog", () => {
  it("opens and selects an option without a focus loop", async () => {
    const user = userEvent.setup();
    const onValueChange = vi.fn();

    render(
      <Dialog open>
        <DialogContent>
          <DialogTitle>予約区分</DialogTitle>
          <DialogDescription>予約区分を選択します。</DialogDescription>
          <Select onValueChange={onValueChange}>
            <SelectTrigger aria-label="予約区分">
              <SelectValue placeholder="選択してください" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="consultation">診察</SelectItem>
              <SelectItem value="vaccination">予防接種</SelectItem>
            </SelectContent>
          </Select>
        </DialogContent>
      </Dialog>,
    );

    await user.click(screen.getByRole("combobox", { name: "予約区分" }));
    await user.click(screen.getByRole("option", { name: "診察" }));

    expect(onValueChange).toHaveBeenCalledWith("consultation");
    expect(screen.getByRole("combobox", { name: "予約区分" })).toHaveTextContent("診察");
  });
});
