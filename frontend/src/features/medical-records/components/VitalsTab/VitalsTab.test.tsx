import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { toast } from "sonner";

import { server } from "@/testing/mocks/node";
import { createTestWrapper } from "@/testing/TestUtils";
import { VitalsTab } from "./VitalsTab";

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: true, canCreate: true, canEdit: true, canDelete: true }),
}));

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
