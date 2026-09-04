import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ClinicHolidayModal } from "./ClinicHolidayModal";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

const holidayMutations = vi.hoisted(() => ({
  create: vi.fn().mockResolvedValue({ id: 1, clinic_id: 1, date: "2026-08-11", reason: "" }),
  remove: vi.fn().mockResolvedValue(undefined),
}));

const toastError = vi.hoisted(() => vi.fn());

vi.mock("../../api/clinic-holidays", () => ({
  useCreateClinicHoliday: () => ({ mutateAsync: holidayMutations.create }),
  useDeleteClinicHoliday: () => ({ mutateAsync: holidayMutations.remove }),
}));

vi.mock("sonner", () => ({
  toast: { error: toastError, success: vi.fn() },
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
  toastError.mockClear();
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
      <ClinicHolidayModal open onClose={onClose} date="2026-08-11" existing={undefined} canEdit />,
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
    expect(toastError).not.toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
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
    expect(toastError).not.toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });
});

describe("ClinicHolidayModal mutation permission re-check", () => {
  it("canEdit=false の定休日設定は mutate せず toast.error する", async () => {
    const onClose = vi.fn();
    const user = userEvent.setup();
    const { rerender } = render(
      <ClinicHolidayModal open onClose={onClose} date="2026-08-11" existing={undefined} canEdit />,
    );

    await user.type(screen.getByRole("textbox", { name: "理由・メモ（任意）" }), "設備点検");
    const form = screen.getByRole("textbox", { name: "理由・メモ（任意）" }).closest("form");
    expect(form).not.toBeNull();

    rerender(
      <ClinicHolidayModal
        open
        onClose={onClose}
        date="2026-08-11"
        existing={undefined}
        canEdit={false}
      />,
    );

    await act(async () => {
      form?.requestSubmit();
    });

    await waitFor(() => {
      expect(toastError).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(holidayMutations.create).not.toHaveBeenCalled();
    expect(holidayMutations.remove).not.toHaveBeenCalled();
    expect(onClose).not.toHaveBeenCalled();
  });

  it("canEdit=false の定休日解除は mutate せず toast.error する", async () => {
    const onClose = vi.fn();
    const { rerender } = render(
      <ClinicHolidayModal
        open
        onClose={onClose}
        date="2026-07-22"
        existing={existingHoliday}
        canEdit
      />,
    );

    const form = screen.getByRole("textbox", { name: "理由・メモ（任意）" }).closest("form");
    expect(form).not.toBeNull();

    rerender(
      <ClinicHolidayModal
        open
        onClose={onClose}
        date="2026-07-22"
        existing={existingHoliday}
        canEdit={false}
      />,
    );

    await act(async () => {
      const submitter = document.createElement("button");
      submitter.type = "submit";
      submitter.name = "intent";
      submitter.value = "remove";
      form?.appendChild(submitter);
      form?.requestSubmit(submitter);
      submitter.remove();
    });

    await waitFor(() => {
      expect(toastError).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(holidayMutations.remove).not.toHaveBeenCalled();
    expect(holidayMutations.create).not.toHaveBeenCalled();
    expect(onClose).not.toHaveBeenCalled();
  });
});
