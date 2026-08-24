import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ClinicHolidayModal } from "./ClinicHolidayModal";

const holidayMutations = vi.hoisted(() => ({
  create: vi.fn().mockResolvedValue({ id: 1, clinic_id: 1, date: "2026-08-11", reason: "" }),
  remove: vi.fn().mockResolvedValue(undefined),
}));

vi.mock("../../api/clinic-holidays", () => ({
  useCreateClinicHoliday: () => ({ mutateAsync: holidayMutations.create }),
  useDeleteClinicHoliday: () => ({ mutateAsync: holidayMutations.remove }),
}));

const existingHoliday = {
  id: 1,
  clinic_id: 1,
  date: "2026-07-22",
  reason: "定例休診",
  created_at: "2026-07-01T00:00:00Z",
  updated_at: "2026-07-01T00:00:00Z",
};

beforeEach(() => {
  holidayMutations.create.mockClear();
  holidayMutations.remove.mockClear();
});

describe("ClinicHolidayModal state sync", () => {
  it("同じ日付・理由のexistingが新しいobjectになっても入力途中の理由をresetしない", async () => {
    const user = userEvent.setup();
    const { rerender } = render(
      <ClinicHolidayModal
        open
        onClose={vi.fn()}
        date="2026-07-22"
        existing={existingHoliday}
        canEdit
      />,
    );

    const reason = screen.getByRole("textbox", { name: "理由・メモ（任意）" });
    await user.clear(reason);
    await user.type(reason, "入力途中の理由");

    rerender(
      <ClinicHolidayModal
        open
        onClose={vi.fn()}
        date="2026-07-22"
        existing={{ ...existingHoliday }}
        canEdit
      />,
    );

    expect(reason).toHaveValue("入力途中の理由");
  });
});

describe("ClinicHolidayModal clinic-holidays write path", () => {
  it("定休日に設定すると useCreateClinicHoliday で clinic-holidays に保存する", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(
      <ClinicHolidayModal
        open
        onClose={onClose}
        date="2026-08-11"
        existing={undefined}
        canEdit
      />,
    );

    await user.type(screen.getByRole("textbox", { name: "理由・メモ（任意）" }), "設備点検");
    await user.click(screen.getByRole("button", { name: "定休日に設定" }));

    await waitFor(() => {
      expect(holidayMutations.create).toHaveBeenCalledWith({
        date: "2026-08-11",
        reason: "設備点検",
      });
    });
    expect(holidayMutations.remove).not.toHaveBeenCalled();
    await waitFor(() => {
      expect(onClose).toHaveBeenCalled();
    });
  });

  it("定休日を解除すると useDeleteClinicHoliday で clinic-holidays から削除する", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    render(
      <ClinicHolidayModal
        open
        onClose={onClose}
        date="2026-07-22"
        existing={existingHoliday}
        canEdit
      />,
    );

    await user.click(screen.getByRole("button", { name: "定休日を解除" }));

    await waitFor(() => {
      expect(holidayMutations.remove).toHaveBeenCalledWith("2026-07-22");
    });
    expect(holidayMutations.create).not.toHaveBeenCalled();
    await waitFor(() => {
      expect(onClose).toHaveBeenCalled();
    });
  });
});
