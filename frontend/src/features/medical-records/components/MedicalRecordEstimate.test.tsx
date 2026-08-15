import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createElement, type ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { toast } from "sonner";

import { MedicalRecordEstimate } from "./MedicalRecordEstimate";

const mockCreateMutateAsync = vi.fn();
const mockUpdateMutateAsync = vi.fn();
let mockExisting: { id: number; title: string; comment?: string; notes?: string; discount_amount?: number; items?: unknown[] } | null = null;

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
    info: vi.fn(),
  },
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canEdit: true, canCreate: true, canDelete: true }),
}));

vi.mock("../api/save-estimate", () => ({
  useGetEstimateByRecord: () => ({ data: mockExisting }),
  useCreateEstimateRecord: () => ({ mutateAsync: mockCreateMutateAsync }),
  useUpdateEstimateRecord: () => ({ mutateAsync: mockUpdateMutateAsync }),
}));

function wrapper({ children }: { children: ReactNode }) {
  const qc = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return createElement(QueryClientProvider, { client: qc }, children);
}

describe("MedicalRecordEstimate BUG-016 resave", () => {
  beforeEach(() => {
    mockExisting = null;
    mockCreateMutateAsync.mockReset();
    mockUpdateMutateAsync.mockReset();
    vi.mocked(toast.success).mockClear();
    vi.mocked(toast.error).mockClear();
  });

  it("初回保存は POST (create)、2回目以降は PATCH (update) を呼び成功トーストは実成功時のみ", async () => {
    const user = userEvent.setup();
    const registered: Array<() => Promise<void>> = [];

    mockCreateMutateAsync.mockResolvedValue({
      id: 42,
      title: "初回件名",
      comment: "",
      notes: "",
      discount_amount: 0,
      items: [],
    });
    mockUpdateMutateAsync.mockResolvedValue({
      id: 42,
      title: "更新件名",
      comment: "",
      notes: "",
      discount_amount: 0,
      items: [],
    });

    render(
      createElement(MedicalRecordEstimate, {
        medicalRecordId: "99",
        onRegisterSave: (fn) => {
          registered.push(fn);
        },
      }),
      { wrapper },
    );

    const subject = await screen.findByRole("textbox", { name: /見積書件名/i }).catch(async () => {
      // Label が htmlFor 未接続の場合は placeholder 無しの textbox を拾う
      const inputs = screen.getAllByRole("textbox");
      return inputs[0];
    });

    await user.clear(subject);
    await user.type(subject, "初回件名");

    // 最新の登録ハンドラを使用
    const firstSave = registered[registered.length - 1];
    expect(firstSave).toEqual(expect.any(Function));

    await act(async () => {
      await firstSave();
    });

    expect(mockCreateMutateAsync).toHaveBeenCalledOnce();
    expect(mockUpdateMutateAsync).not.toHaveBeenCalled();
    expect(toast.success).toHaveBeenCalledWith("見積書を保存しました");

    // 件名変更 → 2回目は PATCH
    await user.clear(subject);
    await user.type(subject, "更新件名");

    const secondSave = registered[registered.length - 1];
    await act(async () => {
      await secondSave();
    });

    await waitFor(() => expect(mockUpdateMutateAsync).toHaveBeenCalledOnce());
    expect(mockCreateMutateAsync).toHaveBeenCalledOnce(); // 増えない
    expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
      id: 42,
      payload: expect.objectContaining({
        title: "更新件名",
        medical_record_id: 99,
      }),
    });
    expect(toast.success).toHaveBeenCalledTimes(2);
  });

  it("existing がある状態からの保存は最初から PATCH", async () => {
    mockExisting = {
      id: 7,
      title: "既存",
      comment: "",
      notes: "",
      discount_amount: 0,
      items: [],
    };
    mockUpdateMutateAsync.mockResolvedValue({ ...mockExisting, title: "既存更新" });

    const registered: Array<() => Promise<void>> = [];
    const user = userEvent.setup();

    render(
      createElement(MedicalRecordEstimate, {
        medicalRecordId: "5",
        onRegisterSave: (fn) => {
          registered.push(fn);
        },
      }),
      { wrapper },
    );

    const inputs = await screen.findAllByRole("textbox");
    const subject = inputs[0];
    await user.clear(subject);
    await user.type(subject, "既存更新");

    await act(async () => {
      await registered[registered.length - 1]();
    });

    expect(mockCreateMutateAsync).not.toHaveBeenCalled();
    expect(mockUpdateMutateAsync).toHaveBeenCalledWith({
      id: 7,
      payload: expect.objectContaining({ title: "既存更新" }),
    });
  });

  it("API 失敗時は成功トーストを出さない", async () => {
    mockCreateMutateAsync.mockRejectedValue(new Error("network"));
    const registered: Array<() => Promise<void>> = [];
    const user = userEvent.setup();

    render(
      createElement(MedicalRecordEstimate, {
        medicalRecordId: "1",
        onRegisterSave: (fn) => {
          registered.push(fn);
        },
      }),
      { wrapper },
    );

    const subject = (await screen.findAllByRole("textbox"))[0];
    await user.type(subject, "x");

    await expect(
      act(async () => {
        await registered[registered.length - 1]();
      }),
    ).rejects.toThrow();

    expect(toast.success).not.toHaveBeenCalled();
  });
});
