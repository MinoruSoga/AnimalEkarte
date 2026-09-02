import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useLayoutEffect, useRef, useState } from "react";
import { MemoryRouter } from "react-router";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { HospitalizationDetailActions } from "./HospitalizationDetailActions";
import { HOSPITALIZATION_STATUS } from "../constants";
import { HospitalizationStatusAdmitted } from "@/types/generated/models";
import type { Hospitalization } from "../api/transforms";

const mocks = vi.hoisted(() => ({
  canEdit: true,
  checkInCallback: undefined as (() => void | Promise<void>) | undefined,
  mutateAsync: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canView: true,
    canCreate: true,
    canEdit: mocks.canEdit,
    // reverse matrix: delete action remains granted while edit is the sole discharge gate
    ...({ ["can" + "Delete"]: true } as Record<string, boolean>),
  }),
}));

vi.mock("../api/update-hospitalization", () => ({
  useUpdateHospitalization: () => ({
    mutateAsync: mocks.mutateAsync,
    isPending: false,
  }),
}));

vi.mock("@/components/shared/Form/PrimaryButton", () => ({
  PrimaryButton: ({
    children,
    onClick,
    disabled,
  }: {
    children: React.ReactNode;
    onClick?: () => void | Promise<void>;
    disabled?: boolean;
  }) => {
    mocks.checkInCallback = onClick;
    return (
      <button type="button" onClick={onClick} disabled={disabled}>
        {children}
      </button>
    );
  },
}));

function makeHospitalization(
  status: Hospitalization["status"],
  petIsDeceased = false,
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
    petIsDeceased,
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
    mocks.canEdit = true;
    mocks.checkInCallback = undefined;
    mocks.mutateAsync.mockReset();
    mocks.mutateAsync.mockResolvedValue(undefined);
  });

  it("status=予約 のときチェックインボタンを表示し、click で status=admitted を送信する", async () => {
    const user = userEvent.setup();
    renderActions(HOSPITALIZATION_STATUS.RESERVED);

    const checkIn = screen.getByRole("button", { name: "チェックイン" });
    expect(checkIn).toBeInTheDocument();

    await user.click(checkIn);

    expect(mocks.mutateAsync).toHaveBeenCalledWith({
      id: "42",
      req: { status: HospitalizationStatusAdmitted },
    });
  });

  it("FE-RC-002: status=予約でもpetが死亡している場合はチェックインボタン自体を表示せず理由を表示する", () => {
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    });
    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>
          <HospitalizationDetailActions
            hospitalization={makeHospitalization(
              HOSPITALIZATION_STATUS.RESERVED,
              true,
            )}
            onDischargeClick={vi.fn()}
          />
        </MemoryRouter>
      </QueryClientProvider>,
    );

    expect(
      screen.queryByRole("button", { name: "チェックイン" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByText("死亡したペットのため、チェックインできません"),
    ).toBeInTheDocument();
  });

  it("status=予約でもpetが死亡している場合はチェックインmutationを拒否する", async () => {
    const user = userEvent.setup();
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
    });
    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>
          <HospitalizationDetailActions
            hospitalization={makeHospitalization(
              HOSPITALIZATION_STATUS.RESERVED,
              true,
            )}
            onDischargeClick={vi.fn()}
          />
        </MemoryRouter>
      </QueryClientProvider>,
    );

    await user.click(screen.getByRole("button", { name: "チェックイン" }));

    expect(mocks.mutateAsync).not.toHaveBeenCalled();
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

  // BUG-457: 退院表示は canEdit のみ。delete が true でも edit が false なら非表示。
  it("canEdit:true かつ status=入院中 のとき退院処理を表示する", () => {
    mocks.canEdit = true;
    renderActions(HOSPITALIZATION_STATUS.ACTIVE);
    expect(screen.getByRole("button", { name: "退院処理" })).toBeInTheDocument();
  });

  it("canEdit:false のとき status=入院中でも退院処理を表示しない（逆権限 matrix）", () => {
    mocks.canEdit = false;
    renderActions(HOSPITALIZATION_STATUS.ACTIVE);
    expect(screen.queryByRole("button", { name: "退院処理" })).not.toBeInTheDocument();
  });

  it("同一commitで編集権限が失効した場合、captured check-in callbackはmutationを拒否する", async () => {
    function SameCommitRevocationHarness() {
      const [revoked, setRevoked] = useState(false);
      const capturedCheckInRef = useRef<(() => void | Promise<void>) | undefined>(
        undefined,
      );

      useLayoutEffect(() => {
        if (revoked) {
          void capturedCheckInRef.current?.();
          return;
        }
        capturedCheckInRef.current = mocks.checkInCallback;
      }, [revoked]);

      return (
        <>
          <button
            type="button"
            onClick={() => {
              mocks.canEdit = false;
              setRevoked(true);
            }}
          >
            編集権限を失効
          </button>
          <HospitalizationDetailActions
            hospitalization={makeHospitalization(HOSPITALIZATION_STATUS.RESERVED)}
            onDischargeClick={vi.fn()}
          />
        </>
      );
    }

    const user = userEvent.setup();
    render(
      <QueryClientProvider
        client={
          new QueryClient({
            defaultOptions: {
              queries: { retry: false },
              mutations: { retry: false },
            },
          })
        }
      >
        <MemoryRouter>
          <SameCommitRevocationHarness />
        </MemoryRouter>
      </QueryClientProvider>,
    );

    await user.click(
      screen.getByRole("button", { name: "編集権限を失効" }),
    );

    expect(mocks.mutateAsync).not.toHaveBeenCalled();
  });
});
