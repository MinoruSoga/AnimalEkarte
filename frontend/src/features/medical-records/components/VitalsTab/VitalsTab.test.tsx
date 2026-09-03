import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { act, render, screen, waitFor, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { toast } from "sonner";

import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { VitalsTab } from "./VitalsTab";

const permissionStore = vi.hoisted(() => {
  let snapshot = { canView: true, canCreate: true, canEdit: true, canDelete: true };
  const listeners = new Set<() => void>();
  return {
    get: () => snapshot,
    subscribe: (listener: () => void) => {
      listeners.add(listener);
      return () => listeners.delete(listener);
    },
    setCreate: (canCreate: boolean) => {
      snapshot = { ...snapshot, canCreate };
      listeners.forEach((listener) => listener());
    },
    reset: () => {
      snapshot = { canView: true, canCreate: true, canEdit: true, canDelete: true };
    },
  };
});

vi.mock("@/hooks/use-permission", async () => {
  const { useSyncExternalStore } = await import("react");
  return {
    usePermission: () =>
      useSyncExternalStore(permissionStore.subscribe, permissionStore.get, permissionStore.get),
  };
});

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

const MEDICAL_RECORD_ID = "10";

beforeEach(() => {
  vi.mocked(toast.success).mockClear();
  vi.mocked(toast.error).mockClear();
  permissionStore.reset();
});

afterEach(() => {
  server.resetHandlers();
});

// FE-RC-005 系: useCreateVital は hook 側で onError → handleApiError（toast.error）を
// 持つ。呼び出し元がさらに ".mutate(input, { onError })" で handleApiError を渡すと、
// react-query は hook 側 + 呼び出し側の両方のコールバックを実行するため失敗時に
// toast.error が二重発火する（billing-confirmation と同型の回帰）。
describe("VitalsTab FE-RC-005 系 二重トースト回帰", () => {
  it("バイタル追加の失敗時、エラートーストは1回だけ表示する", async () => {
    server.use(
      http.get("*/v1/medical-records/:id/vitals", () => HttpResponse.json([])),
      http.post("*/v1/medical-records/:id/vitals", () =>
        HttpResponse.json({ error: "internal error" }, { status: 500 }),
      ),
    );

    render(<VitalsTab medicalRecordId={MEDICAL_RECORD_ID} />, {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByText("記録を追加")).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText("記録を追加"));

    fireEvent.change(screen.getByLabelText("記録日時"), {
      target: { value: "2026-07-20T10:00" },
    });
    fireEvent.change(screen.getByLabelText("体温"), { target: { value: "38.5" } });

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledTimes(1);
    });
    expect(toast.success).not.toHaveBeenCalled();
  });
});

describe("VitalsTab FE-RC-114 add form action", () => {
  it("追加 UI は form action で送信し、バリデーション失敗は fieldErrors を出す", async () => {
    server.use(http.get("*/v1/medical-records/:id/vitals", () => HttpResponse.json([])));

    render(<VitalsTab medicalRecordId={MEDICAL_RECORD_ID} />, {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByText("記録を追加")).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText("記録を追加"));

    expect(screen.getByRole("form", { name: "バイタル追加" })).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("記録日時"), {
      target: { value: "2026-07-20T10:00" },
    });

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() => {
      expect(
        screen.getByText("体温・心拍数・呼吸数・体重のいずれかを入力してください"),
      ).toBeInTheDocument();
    });
    expect(toast.success).not.toHaveBeenCalled();
  });

  it("追加権限を失ったあとでも送信すると ActionState.error を表示する", async () => {
    server.use(http.get("*/v1/medical-records/:id/vitals", () => HttpResponse.json([])));

    render(<VitalsTab medicalRecordId={MEDICAL_RECORD_ID} />, {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByText("記録を追加")).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText("記録を追加"));

    fireEvent.change(screen.getByLabelText("記録日時"), {
      target: { value: "2026-07-20T10:00" },
    });
    fireEvent.change(screen.getByLabelText("体温"), { target: { value: "38.5" } });

    act(() => {
      permissionStore.setCreate(false);
    });

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() => {
      expect(screen.getByRole("alert")).toHaveTextContent("この操作を行う権限がありません");
    });
    expect(toast.success).not.toHaveBeenCalled();
  });
});

describe("VitalsTab deceased pet dual-gate", () => {
  it("死亡ペットでは記録追加ボタンを出さない", async () => {
    server.use(http.get("*/v1/medical-records/:id/vitals", () => HttpResponse.json([])));

    render(<VitalsTab medicalRecordId={MEDICAL_RECORD_ID} isPetDeceased />, {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      expect(screen.queryByText("読み込み中...")).not.toBeInTheDocument();
    });
    expect(screen.queryByText("記録を追加")).not.toBeInTheDocument();
  });

  it("追加中に死亡へ変わると create せず ActionState.error を出す", async () => {
    server.use(
      http.get("*/v1/medical-records/:id/vitals", () => HttpResponse.json([])),
      http.post("*/v1/medical-records/:id/vitals", () => {
        throw new Error("create must not run for deceased pet");
      }),
    );

    function Harness({ deceased }: { deceased: boolean }) {
      return <VitalsTab medicalRecordId={MEDICAL_RECORD_ID} isPetDeceased={deceased} />;
    }

    const { rerender } = render(<Harness deceased={false} />, {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      expect(screen.getByText("記録を追加")).toBeInTheDocument();
    });
    fireEvent.click(screen.getByText("記録を追加"));
    fireEvent.change(screen.getByLabelText("記録日時"), {
      target: { value: "2026-07-20T10:00" },
    });
    fireEvent.change(screen.getByLabelText("体温"), { target: { value: "38.5" } });

    rerender(<Harness deceased={true} />);

    const user = userEvent.setup();
    await user.click(screen.getByRole("button", { name: "追加" }));

    await waitFor(() => {
      expect(screen.getByRole("alert")).toHaveTextContent(
        "死亡したペットのバイタルは保存できません",
      );
    });
    expect(toast.success).not.toHaveBeenCalled();
  });
});
