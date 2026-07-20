import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { HospitalizationDetailActions } from "./HospitalizationDetailActions";
import { HOSPITALIZATION_STATUS } from "../constants";
import { HospitalizationStatusAdmitted } from "@/types/generated/models";
import type { Hospitalization } from "../api/transforms";

const mutateAsync = vi.fn();

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  }),
}));

vi.mock("../api/update-hospitalization", () => ({
  useUpdateHospitalization: () => ({
    mutateAsync,
    isPending: false,
  }),
}));

function makeHospitalization(
  status: Hospitalization["status"]
): Hospitalization {
  return {
    id: "42",
    hospitalizationNo: "42",
    ownerName: "山田",
    petName: "タロウ",
    species: "犬",
    hospitalizationType: "入院",
    startDate: "2026-07-20",
    endDate: "2026-07-25",
    status,
  };
}

function renderActions(status: Hospitalization["status"]) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  const onDischargeClick = vi.fn();
  render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <HospitalizationDetailActions
          hospitalization={makeHospitalization(status)}
          onDischargeClick={onDischargeClick}
        />
      </MemoryRouter>
    </QueryClientProvider>
  );
  return { onDischargeClick };
}

describe("HospitalizationDetailActions — チェックイン (FEAT-CHECKIN / DEC-2)", () => {
  beforeEach(() => {
    mutateAsync.mockReset();
    mutateAsync.mockResolvedValue(undefined);
  });

  it("status=予約 のときチェックインボタンを表示し、click で status=admitted を送信する", async () => {
    const user = userEvent.setup();
    renderActions(HOSPITALIZATION_STATUS.RESERVED);

    const checkIn = screen.getByRole("button", { name: "チェックイン" });
    expect(checkIn).toBeInTheDocument();

    await user.click(checkIn);

    expect(mutateAsync).toHaveBeenCalledWith({
      id: "42",
      req: { status: HospitalizationStatusAdmitted },
    });
  });

  it("status=入院中・退院済 ではチェックインボタンを表示しない", () => {
    const { unmount } = render(
      <QueryClientProvider
        client={
          new QueryClient({
            defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
          })
        }
      >
        <MemoryRouter>
          <HospitalizationDetailActions
            hospitalization={makeHospitalization(HOSPITALIZATION_STATUS.ACTIVE)}
            onDischargeClick={vi.fn()}
          />
        </MemoryRouter>
      </QueryClientProvider>
    );
    expect(screen.queryByRole("button", { name: "チェックイン" })).not.toBeInTheDocument();
    unmount();

    renderActions(HOSPITALIZATION_STATUS.DISCHARGED);
    expect(screen.queryByRole("button", { name: "チェックイン" })).not.toBeInTheDocument();
  });

  it("旧障害モード回帰: status=予約 では退院処理を非表示、status=入院中 では表示する", () => {
    const { unmount } = render(
      <QueryClientProvider
        client={
          new QueryClient({
            defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
          })
        }
      >
        <MemoryRouter>
          <HospitalizationDetailActions
            hospitalization={makeHospitalization(HOSPITALIZATION_STATUS.RESERVED)}
            onDischargeClick={vi.fn()}
          />
        </MemoryRouter>
      </QueryClientProvider>
    );
    expect(screen.queryByRole("button", { name: "退院処理" })).not.toBeInTheDocument();
    unmount();

    renderActions(HOSPITALIZATION_STATUS.ACTIVE);
    expect(screen.getByRole("button", { name: "退院処理" })).toBeInTheDocument();
  });
});
