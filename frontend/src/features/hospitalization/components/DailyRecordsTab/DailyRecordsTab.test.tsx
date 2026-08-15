import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { DailyRecordsTab } from "./DailyRecordsTab";

const mutateAsync = vi.fn();

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: { id: "1" },
    hasPermission: () => true,
  }),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: false,
  }),
}));

vi.mock("../../api/daily-records", () => ({
  useGetDailyRecord: () => ({
    data: undefined,
    isLoading: false,
    isError: true,
  }),
  useCreateDailyRecord: () => ({
    mutateAsync,
  }),
  useCreateDailyVital: () => ({ mutateAsync: vi.fn() }),
  useCreateCareLog: () => ({ mutateAsync: vi.fn() }),
  useCreateStaffNote: () => ({ mutateAsync: vi.fn() }),
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

function renderTab() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={client}>
      <DailyRecordsTab
        hospitalizationId="10"
        admissionDate="2026-07-01"
        dischargeDate="2026-07-14"
        petIsDeceased={false}
      />
    </QueryClientProvider>
  );
}

describe("DailyRecordsTab — GET 404 → create CTA (AUD-003)", () => {
  beforeEach(() => {
    mutateAsync.mockReset();
    mutateAsync.mockResolvedValue({
      id: "1",
      date: "2026-07-14T00:00:00Z",
    });
  });

  it("shows create CTA when GET is missing (isError) and create posts via useCreateDailyRecord", async () => {
    const user = userEvent.setup();
    renderTab();

    expect(screen.getByText("この日の記録はまだありません")).toBeInTheDocument();
    const button = screen.getByRole("button", { name: /この日の記録を作成/ });
    await user.click(button);

    expect(mutateAsync).toHaveBeenCalled();
  });
});
