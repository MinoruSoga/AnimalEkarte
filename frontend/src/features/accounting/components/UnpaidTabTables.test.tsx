import type { ReactNode } from "react";
import { render, screen, within } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it } from "vitest";

import type { Accounting } from "../api/transforms";
import type { PeriodUnpaidResponse } from "../api/get-unpaid-billings";
import { UnpaidBillingTable, UnpaidPeriodTable } from "./UnpaidTabTables";

function renderTables(ui: ReactNode) {
  return render(<>{ui}</>, {
    wrapper: ({ children }) => <MemoryRouter>{children}</MemoryRouter>,
  });
}

const BILLING_ROW: Accounting = {
  id: "77",
  clinicId: "1",
  medicalRecordId: undefined,
  ownerId: "31",
  ownerName: "山田花子",
  petId: "41",
  petName: "チョコ",
  petSpecies: undefined,
  status: "waiting",
  scheduledDate: "2026-06-10",
  completedAt: undefined,
  items: [],
  payment: undefined,
  paymentSplits: undefined,
  totalAmount: 5000,
  totalRefundedAmount: 0,
  outstandingAmount: 5000,
  memo: undefined,
};

const PERIOD_ROWS: PeriodUnpaidResponse["data"] = [
  {
    owner_id: 1,
    owner_name: "山田花子",
    pet_id: 10,
    pet_name: "チョコ",
    prev_period_carryover: 3000,
    current_period_unpaid: 5000,
    period_end_carryover: 8000,
    latest_scheduled: "2026-06-20",
  },
  {
    owner_id: 2,
    owner_name: "鈴木一郎",
    pet_name: "",
    prev_period_carryover: 0,
    current_period_unpaid: 2000,
    period_end_carryover: 2000,
    latest_scheduled: "2026-06-05",
  },
];

describe("UnpaidBillingTable", () => {
  it("最新未納日列にその会計の scheduled_date を YYYY-MM-DD で表示する", () => {
    renderTables(<UnpaidBillingTable billings={[BILLING_ROW]} endDate="2026-06-30" />);

    const headers = screen.getAllByRole("columnheader").map((h) => h.textContent);
    expect(headers).toContain("最新未納日");

    const row = screen.getByRole("row", { name: /山田花子/ });
    expect(within(row).getByText("2026-06-10")).toBeInTheDocument();
  });
});

describe("UnpaidPeriodTable", () => {
  it("最新未納日列に各行の latest_scheduled を YYYY-MM-DD で表示する", () => {
    renderTables(<UnpaidPeriodTable rows={PERIOD_ROWS} />);

    const headers = screen.getAllByRole("columnheader").map((h) => h.textContent);
    expect(headers).toContain("最新未納日");

    const yamadaRow = screen.getByRole("row", { name: /山田花子/ });
    expect(within(yamadaRow).getByText("2026-06-20")).toBeInTheDocument();
    const suzukiRow = screen.getByRole("row", { name: /鈴木一郎/ });
    expect(within(suzukiRow).getByText("2026-06-05")).toBeInTheDocument();
  });
});
