import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { CashRegisterClosePage } from "./CashRegisterClosePage";
import type { ClosePreviewResult } from "../api/get-cash-register-preview";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

const { mockPreviewState, permissionState, mockMutateAsync, mockToast } = vi.hoisted(() => ({
  mockPreviewState: {
    data: undefined as ClosePreviewResult | undefined,
    isLoading: false,
  },
  permissionState: {
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  },
  mockMutateAsync: vi.fn().mockResolvedValue({}),
  mockToast: { error: vi.fn(), success: vi.fn() },
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: {
      clinics: [{ clinicId: "1", clinicName: "テスト動物病院", isMain: true }],
      clinic: { name: "テスト動物病院" },
    },
    currentClinicId: "1",
    hasPermission: (_resource: string, action: "view" | "create" | "edit" | "delete") => {
      switch (action) {
        case "view":
          return permissionState.canView;
        case "create":
          return permissionState.canCreate;
        case "edit":
          return permissionState.canEdit;
        case "delete":
          return permissionState.canDelete;
      }
    },
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
    previewNonce: 0,
    handleDateChange: vi.fn(),
    handlePeriodChange: vi.fn(),
    enablePreview: vi.fn(),
  }),
}));

vi.mock("../api/get-cash-register-preview", () => ({
  useGetCashRegisterPreview: () => mockPreviewState,
}));

vi.mock("../api/create-cash-register-close", () => ({
  useCreateCashRegisterClose: () => ({ mutateAsync: mockMutateAsync, isPending: false }),
}));

vi.mock("sonner", () => ({
  toast: mockToast,
}));

const PREVIEW: ClosePreviewResult = {
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

function renderPage() {
  return render(
    <MemoryRouter>
      <CashRegisterClosePage />
    </MemoryRouter>,
  );
}

async function confirmClose() {
  const user = userEvent.setup();
  await user.type(screen.getByLabelText("実際のレジ現金（円）"), "1000");
  await user.click(screen.getAllByRole("button", { name: "締める" })[0]);
  const dialog = await screen.findByRole("alertdialog");
  await user.click(within(dialog).getByRole("button", { name: "締める" }));
  return user;
}

describe("CashRegisterClosePage", () => {
  beforeEach(() => {
    mockPreviewState.data = undefined;
    mockPreviewState.isLoading = false;
    permissionState.canView = true;
    permissionState.canCreate = true;
    permissionState.canEdit = true;
    permissionState.canDelete = true;
    mockMutateAsync.mockReset();
    mockMutateAsync.mockResolvedValue({});
    mockToast.error.mockReset();
    mockToast.success.mockReset();
  });

  it("プレビュー Primary CTA は DESIGN.md の pill 形状を使う", () => {
    renderPage();

    const previewButton = screen.getByRole("button", { name: "プレビュー" });
    expect(previewButton).toHaveClass("rounded-full");
    expect(previewButton).not.toHaveClass("rounded-xs");
  });

  it("プレビュー後の印刷操作を44px以上に保つ", () => {
    mockPreviewState.data = PREVIEW;
    renderPage();

    expect(screen.getByTestId("close-print-button")).toHaveClass("min-h-11");
  });

  it("canCreate=true なら締め確定で mutateAsync を呼ぶ", async () => {
    mockPreviewState.data = PREVIEW;
    renderPage();

    await confirmClose();

    expect(mockMutateAsync).toHaveBeenCalledWith({
      date: "2026-07-21",
      period: "am",
      actual_cash: 1000,
      memo: undefined,
    });
    expect(mockToast.error).not.toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("canCreate=false なら締め確定で mutateAsync せず toast する", async () => {
    permissionState.canCreate = false;
    permissionState.canEdit = true;
    mockPreviewState.data = PREVIEW;
    renderPage();

    await confirmClose();

    expect(mockMutateAsync).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });
});
