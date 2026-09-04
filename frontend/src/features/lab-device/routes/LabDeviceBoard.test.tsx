import { render, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { LabDeviceBoard } from "./LabDeviceBoard";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

const { permission, captured, receiveMutateAsync, mockToast } = vi.hoisted(() => ({
  permission: {
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  },
  captured: {
    onFrame: undefined as
      undefined | ((frame: { payloadBase64: string; deviceHint: "auto" }) => Promise<void>),
  },
  receiveMutateAsync: vi.fn().mockResolvedValue([]),
  mockToast: { error: vi.fn(), success: vi.fn(), dismiss: vi.fn() },
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
    onFrame: (frame: { payloadBase64: string; deviceHint: "auto" }) => Promise<void>;
  }) => {
    captured.onFrame = input.onFrame;
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
    useGetLabDeviceAgentConsumer: () => ({ data: undefined }),
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
    const { rerender } = render(
      <MemoryRouter>
        <LabDeviceBoard />
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(captured.onFrame).toBeTypeOf("function");
    });

    permission.canCreate = false;
    rerender(
      <MemoryRouter>
        <LabDeviceBoard />
      </MemoryRouter>,
    );

    await captured.onFrame?.({ payloadBase64: "AA==", deviceHint: "auto" });

    expect(receiveMutateAsync).not.toHaveBeenCalled();
    expect(mockToast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });
});
