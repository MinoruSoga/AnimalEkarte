import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { DailyRecordsTab } from "./DailyRecordsTab";

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ user: { id: "1" }, hasPermission: () => true }),
}));

const mocks = vi.hoisted(() => ({ canCreate: true }));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canView: true,
    canCreate: mocks.canCreate,
    canEdit: true,
    canDelete: true,
  }),
}));

vi.mock("../../api/daily-records", () => ({
  useGetDailyRecord: () => ({
    data: {
      id: "daily-1",
      hospitalization_id: "10",
      date: "2026-07-14T00:00:00+09:00",
      created_at: "",
      updated_at: "",
      vital_records: [],
      care_logs: [],
      staff_notes: [],
    },
    isLoading: false,
    isError: false,
  }),
  useCreateDailyRecord: () => ({ mutateAsync: vi.fn() }),
  useCreateDailyVital: () => ({ mutateAsync: vi.fn() }),
  useCreateCareLog: () => ({ mutateAsync: vi.fn() }),
  useCreateStaffNote: () => ({ mutateAsync: vi.fn() }),
}));

function renderTab(petIsDeceased: boolean) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <DailyRecordsTab
        hospitalizationId="10"
        admissionDate="2026-07-01"
        dischargeDate="2026-07-14"
        petIsDeceased={petIsDeceased}
      />
    </QueryClientProvider>,
  );
}

describe("FE-RC-002: DailyRecordsTab — 死亡ペットの render 側防壁", () => {
  beforeEach(() => {
    mocks.canCreate = true;
  });

  it("petIsDeceased=false では追加ボタンを3セクション分表示する", () => {
    renderTab(false);
    expect(screen.getAllByRole("button", { name: "追加" })).toHaveLength(3);
    expect(
      screen.queryByText("死亡したペットのため、デイリーカルテの記録・追加はできません"),
    ).not.toBeInTheDocument();
  });

  it("petIsDeceased=true では追加ボタンを一切表示せず理由を表示する", () => {
    renderTab(true);
    expect(
      screen.queryByRole("button", { name: "追加" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByText("死亡したペットのため、デイリーカルテの記録・追加はできません"),
    ).toBeInTheDocument();
  });
});
