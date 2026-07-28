import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { ClinicHolidayModal } from "./ClinicHolidayModal";

vi.mock("../../api/clinic-holidays", () => ({
  useCreateClinicHoliday: () => ({ mutateAsync: vi.fn() }),
  useDeleteClinicHoliday: () => ({ mutateAsync: vi.fn() }),
}));

describe("ClinicHolidayModal state sync", () => {
  it("同じ日付・理由のexistingが新しいobjectになっても入力途中の理由をresetしない", async () => {
    const user = userEvent.setup();
    const existing = {
      id: 1,
      clinic_id: 1,
      date: "2026-07-22",
      reason: "定例休診",
      created_at: "2026-07-01T00:00:00Z",
      updated_at: "2026-07-01T00:00:00Z",
    };
    const { rerender } = render(
      <ClinicHolidayModal
        open
        onClose={vi.fn()}
        date="2026-07-22"
        existing={existing}
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
        existing={{ ...existing }}
        canEdit
      />,
    );

    expect(reason).toHaveValue("入力途中の理由");
  });
});
