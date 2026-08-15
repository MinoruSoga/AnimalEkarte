import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { CashRegisterClosePage } from "./CashRegisterClosePage";
import type { ClosePreviewResult } from "../api/get-cash-register-preview";

const { mockPreviewState } = vi.hoisted(() => ({
  mockPreviewState: {
    data: undefined as ClosePreviewResult | undefined,
    isLoading: false,
  },
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: {
      clinics: [{ clinicId: "1", clinicName: "テスト動物病院", isMain: true }],
      clinic: { name: "テスト動物病院" },
    },
    currentClinicId: "1",
    hasPermission: () => true,
  }),
}));

vi.mock("@/hooks/use-current-clinic-name", () => ({
  useCurrentClinicName: () => "テスト動物病院",
}));

vi.mock("../hooks/use-cash-register-close-form", () => ({
  useCashRegisterCloseForm: () => ({
    date: "2026-07-21",
    period: "am",
    previewEnabled: false,
    handleDateChange: vi.fn(),
    handlePeriodChange: vi.fn(),
    enablePreview: vi.fn(),
  }),
}));

vi.mock("../api/get-cash-register-preview", () => ({
  useGetCashRegisterPreview: () => mockPreviewState,
}));

vi.mock("../api/create-cash-register-close", () => ({
  useCreateCashRegisterClose: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

describe("CashRegisterClosePage", () => {
  beforeEach(() => {
    mockPreviewState.data = undefined;
  });

  it("プレビュー Primary CTA は DESIGN.md の pill 形状を使う", () => {
    render(<CashRegisterClosePage />);

    const previewButton = screen.getByRole("button", { name: "プレビュー" });
    expect(previewButton).toHaveClass("rounded-full");
    expect(previewButton).not.toHaveClass("rounded-xs");
  });

  it("プレビュー後の印刷操作を44px以上に保つ", () => {
    mockPreviewState.data = {
      date: "2026-07-21",
      period: "am",
      periodStart: "2026-07-21T00:00:00+09:00",
      periodEnd: "2026-07-21T11:59:59+09:00",
      isHoliday: false,
      isAlreadyClosed: false,
      aggregate: {
        categories: {},
        paymentMethods: [],
        theoreticalCash: 0,
        taxBreakdown: {
          standard: { taxableAmount: 0, taxAmount: 0 },
          reduced: { taxableAmount: 0, taxAmount: 0 },
        },
      },
      billingDetails: [],
    };

    render(<CashRegisterClosePage />);

    expect(screen.getByTestId("close-print-button")).toHaveClass("min-h-11");
  });
});
