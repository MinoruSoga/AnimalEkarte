import { render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { LabDeviceBoard } from "./LabDeviceBoard";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

const { permission, captured, receiveMutateAsync, mockToast, agentConsumer } = vi.hoisted(() => ({
  permission: {
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  },
  captured: {
    onFrame: undefined as
      undefined | ((frame: { payloadBase64: string; deviceHint: "auto" }) => Promise<void>),
    consumerToken: undefined as string | undefined,
    enabled: undefined as boolean | undefined,
  },
  receiveMutateAsync: vi.fn().mockResolvedValue([]),
  mockToast: { error: vi.fn(), success: vi.fn(), dismiss: vi.fn() },
  // useGetLabDeviceAgentConsumer の返り値差し替え用。
  // data: string=発行済み / null=503 未設定 / undefined=未取得、status で query 状態を制御。
  agentConsumer: {
    data: undefined as string | null | undefined,
    error: null as unknown,
    status: "success" as "pending" | "error" | "success",
  },
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({
    user: {
      clinics: [{ clinicId: "1", clinicName: "テスト動物病院", isMain: true }],
      clinic: { name: "テスト動物病院" },
    },
    currentClinicId: "clinic-1",
    hasPermission: () => true,
  }),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ ...permission }),
}));

vi.mock("sonner", () => ({
  toast: mockToast,
}));

vi.mock("../hooks/use-lab-device-agent-listen", () => ({
  useLabDeviceAgentListen: (input: {
    enabled: boolean;
    clinicId: string | null;
    consumerToken?: string;
    onFrame: (frame: { payloadBase64: string; deviceHint: "auto" }) => Promise<void>;
    onUnauthorized?: () => void;
  }) => {
    captured.onFrame = input.onFrame;
    captured.consumerToken = input.consumerToken;
    captured.enabled = input.enabled;
    return {
      connected: false,
      openPorts: 0,
      configuredPorts: 0,
      pending: 0,
      rejected: 0,
      overflow: 0,
      inputOverflow: 0,
      portDiscoveryFailures: 0,
      portOpenFailures: 0,
      queueFailures: 0,
      portCloseFailures: 0,
      responseFailures: 0,
      lastErrorCategory: "none",
      degraded: false,
    };
  },
}));

vi.mock("../api/lab-device", async (importOriginal) => {
  const actual = await importOriginal<typeof import("../api/lab-device")>();
  return {
    ...actual,
    useGetLabDeviceBoard: () => ({
      data: {
        wait: null,
        unlinked: [],
        saved: [],
        received: [],
        todayVisits: [],
        station: { waitTtlSeconds: 60, slotsJson: "[]" },
      },
      isLoading: false,
    }),
    useGetLabDeviceAgentConsumer: () => ({
      data: agentConsumer.data,
      error: agentConsumer.error,
      status: agentConsumer.status,
      isPending: agentConsumer.status === "pending",
      isError: agentConsumer.status === "error",
      isSuccess: agentConsumer.status === "success",
    }),
    usePutLabDeviceWait: () => ({ mutate: vi.fn() }),
    useClearLabDeviceWait: () => ({ mutate: vi.fn() }),
    useReceiveLabDeviceFrames: () => ({ mutateAsync: receiveMutateAsync }),
    useAttachLabDeviceJob: () => ({ mutateAsync: vi.fn() }),
    useDetachLabDeviceJob: () => ({ mutate: vi.fn() }),
  };
});

describe("LabDeviceBoard permissions", () => {
  beforeEach(() => {
    permission.canView = true;
    permission.canCreate = true;
    permission.canEdit = true;
    permission.canDelete = true;
    captured.onFrame = undefined;
    receiveMutateAsync.mockReset();
    receiveMutateAsync.mockResolvedValue([]);
    mockToast.error.mockReset();
  });

  it("canCreate=false になったあと onFrame は受信 mutation せず toast する", async () => {
    const queryClient = new QueryClient();
    const { rerender } = render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>
          <LabDeviceBoard />
        </MemoryRouter>
      </QueryClientProvider>,
    );

    await waitFor(() => {
      expect(captured.onFrame).toBeTypeOf("function");
    });

    permission.canCreate = false;
    rerender(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>
          <LabDeviceBoard />
        </MemoryRouter>
      </QueryClientProvider>,
    );

    await captured.onFrame?.({ payloadBase64: "AA==", deviceHint: "auto" });

    expect(receiveMutateAsync).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });
});

describe("LabDeviceBoard agent-consumer error surface (EMR-335)", () => {
  const AGENT_UNCONFIGURED_MESSAGE = "検査機器エージェントが未設定です";

  function renderBoard() {
    const queryClient = new QueryClient();
    return render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>
          <LabDeviceBoard />
        </MemoryRouter>
      </QueryClientProvider>,
    );
  }

  beforeEach(() => {
    agentConsumer.data = undefined;
    agentConsumer.error = null;
    agentConsumer.status = "success";
    captured.consumerToken = undefined;
    captured.enabled = undefined;
    mockToast.error.mockReset();
    mockToast.dismiss.mockReset();
  });

  it("data=null（秘密未設定の 503）は「検査機器エージェントが未設定です」を出し、リスナーへトークンを渡さない", async () => {
    agentConsumer.data = null;

    renderBoard();

    await waitFor(() => {
      expect(screen.getByText(AGENT_UNCONFIGURED_MESSAGE)).toBeTruthy();
    });
    // 未設定は期待状態。汎用サーバーエラー等のトーストは出さない。
    expect(mockToast.error).not.toHaveBeenCalled();
    expect(captured.consumerToken).toBeUndefined();
    expect(screen.getByRole("status").textContent).toBe(AGENT_UNCONFIGURED_MESSAGE);
  });

  it("query エラー時はリスンを無効化し、キャッシュに残った旧 capability を使わない", async () => {
    agentConsumer.status = "error";
    agentConsumer.error = new Error("server error");
    agentConsumer.data = "stale-capability-dummy"; // refetch 失敗後も data に残る旧トークン

    renderBoard();

    await waitFor(() => {
      expect(captured.enabled).toBe(false);
    });
    expect(captured.consumerToken).toBeUndefined();
    // 非 503 系の失敗は未設定メッセージではなく通常のエラー経路で可視化される。
    expect(screen.queryByText(AGENT_UNCONFIGURED_MESSAGE)).toBeNull();
  });

  it("成功時は発行済みトークンをリスナーへ渡し、未設定メッセージは出ない", async () => {
    agentConsumer.data = "cap-token-dummy";

    renderBoard();

    await waitFor(() => {
      expect(captured.consumerToken).toBe("cap-token-dummy");
    });
    expect(screen.queryByText(AGENT_UNCONFIGURED_MESSAGE)).toBeNull();
    expect(screen.getByRole("status").textContent).toBe("");
  });

  it("未設定のキャッシュが残る再取得エラーは未設定表示と混同しない", () => {
    agentConsumer.data = null;
    agentConsumer.status = "error";
    agentConsumer.error = new Error("server error");

    renderBoard();

    expect(screen.queryByText(AGENT_UNCONFIGURED_MESSAGE)).toBeNull();
    expect(captured.enabled).toBe(false);
    expect(captured.consumerToken).toBeUndefined();
  });

  it("pending 中は未設定メッセージを出さない（読み込み中の誤表示防止）", () => {
    agentConsumer.status = "pending";
    agentConsumer.data = undefined;

    renderBoard();

    expect(screen.queryByText(AGENT_UNCONFIGURED_MESSAGE)).toBeNull();
    expect(captured.consumerToken).toBeUndefined();
  });
});
