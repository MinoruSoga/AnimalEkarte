import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { LabDeviceBoard } from "./LabDeviceBoard";

const { permission, clinic, receivePanelStatus, receiveMutateAsync, calls } = vi.hoisted(() => ({
  permission: { canView: true, canCreate: true, canEdit: true, canDelete: true },
  clinic: { currentClinicId: "clinic-1" as string | null },
  receivePanelStatus: {
    connected: false,
    pending: 0,
    rejected: 0,
    overflow: 0,
    authPaused: false,
    error: null as string | null,
  },
  receiveMutateAsync: vi
    .fn()
    .mockResolvedValue([
      { duplicate: false, job: { deviceHint: "fuji_nx600", sourceType: "fuji_nx600" } },
    ]),
  calls: {
    connect: vi.fn(),
    // useGetLabDeviceAgentConsumer の呼出し検出用。
    // 通常受信は旧queryの状態や共有secretを参照しない。
    legacyConsumer: vi.fn(),
    onFrame: undefined as
      ((frame: { payloadBase64: string; deviceHint: "auto" }) => Promise<void>) | undefined,
  },
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ currentClinicId: clinic.currentClinicId }),
}));
vi.mock("@/hooks/use-permission", () => ({ usePermission: () => ({ ...permission }) }));
vi.mock("sonner", () => ({ toast: { error: vi.fn(), dismiss: vi.fn() } }));
vi.mock("../hooks/use-lab-device-browser-receive", () => ({
  useLabDeviceBrowserReceive: (input: {
    enabled: boolean;
    clinicId: string | null;
    onFrame: (frame: { payloadBase64: string; deviceHint: "auto" }) => Promise<void>;
  }) => ({
    ...(() => {
      calls.onFrame = input.onFrame;
      return {};
    })(),
    supported: true,
    connected: receivePanelStatus.connected,
    openPorts: 0,
    pending: receivePanelStatus.pending,
    rejected: receivePanelStatus.rejected,
    overflow: receivePanelStatus.overflow,
    authPaused: receivePanelStatus.authPaused,
    error: receivePanelStatus.error,
    connect: () => {
      calls.connect();
      return input.onFrame({ payloadBase64: "AQ==", deviceHint: "auto" });
    },
    retry: vi.fn(),
    stop: vi.fn(),
  }),
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
    useGetLabDeviceAgentConsumer: (...args: unknown[]) => {
      calls.legacyConsumer(...args);
      throw new Error("legacy consumer endpoint must not be called"); // 旧capabilityへフォールバックしない。
    },
    usePutLabDeviceWait: () => ({ mutate: vi.fn() }),
    useClearLabDeviceWait: () => ({ mutate: vi.fn() }),
    useReceiveLabDeviceFrames: () => ({ mutateAsync: receiveMutateAsync }),
    useAttachLabDeviceJob: () => ({ mutateAsync: vi.fn() }),
    useDetachLabDeviceJob: () => ({ mutate: vi.fn() }),
  };
});

function renderBoard() {
  return render(
    <QueryClientProvider client={new QueryClient()}>
      <MemoryRouter>
        <LabDeviceBoard />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("LabDeviceBoard browser receive", () => {
  beforeEach(() => {
    Object.assign(permission, { canView: true, canCreate: true, canEdit: true, canDelete: true });
    clinic.currentClinicId = "clinic-1";
    Object.assign(receivePanelStatus, {
      connected: false,
      pending: 0,
      rejected: 0,
      overflow: 0,
      authPaused: false,
      error: null,
    });
    calls.connect.mockReset();
    calls.legacyConsumer.mockReset();
    calls.onFrame = undefined;
    receiveMutateAsync.mockReset();
    receiveMutateAsync.mockResolvedValue([
      { duplicate: false, job: { deviceHint: "fuji_nx600", sourceType: "fuji_nx600" } },
    ]);
  });

  it("shows the direct receive controls and never requests a legacy consumer token", () => {
    renderBoard();
    // 未設定は旧APIの状態であり、この画面の接続条件にはしない。
    expect(screen.getByRole("heading", { name: "ブラウザ直接受信" })).toBeTruthy();
    expect(screen.getByRole("button", { name: "接続する" })).toBeTruthy();
    expect(calls.legacyConsumer).not.toHaveBeenCalled();
  });

  it("connect action dispatches through the authenticated receive mutation", async () => {
    renderBoard();
    fireEvent.click(screen.getByRole("button", { name: "接続する" }));
    await waitFor(() =>
      expect(receiveMutateAsync).toHaveBeenCalledWith({
        payloadBase64: "AQ==",
        deviceHint: "auto",
        clinicId: "clinic-1",
      }),
    );
    expect(calls.connect).toHaveBeenCalledTimes(1);
  });

  it("canCreate=false keeps receive disabled and permission check rejects a late frame", async () => {
    const { rerender } = renderBoard();
    const oldFrame = calls.onFrame;
    permission.canCreate = false;
    rerender(
      <QueryClientProvider client={new QueryClient()}>
        <MemoryRouter>
          <LabDeviceBoard />
        </MemoryRouter>
      </QueryClientProvider>,
    );
    expect(screen.getByRole("button", { name: "接続する" }).hasAttribute("disabled")).toBe(true);
    // 旧APIの失敗表示ではなく、通常の権限ガードを検証する。
    await oldFrame?.({ payloadBase64: "AQ==", deviceHint: "auto" });
    expect(receiveMutateAsync).not.toHaveBeenCalled();
    expect(calls.legacyConsumer).not.toHaveBeenCalled();
  });

  it("does not invoke the receive mutation without a selected clinic", async () => {
    clinic.currentClinicId = null;
    renderBoard();
    await calls.onFrame?.({ payloadBase64: "AQ==", deviceHint: "auto" });
    expect(receiveMutateAsync).not.toHaveBeenCalled();
  });

  it("keeps stop reachable when a queue remains after all ports disconnect", () => {
    receivePanelStatus.pending = 1;
    receivePanelStatus.error = "ポートから切断されました";
    renderBoard();
    expect(screen.getByRole("button", { name: "停止" })).toBeTruthy();
  });
});
