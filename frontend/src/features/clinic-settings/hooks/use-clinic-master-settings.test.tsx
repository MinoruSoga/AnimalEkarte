import { act, renderHook, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { describe, expect, it, vi, beforeEach } from "vitest";

import { useClinicMasterSettings } from "./use-clinic-master-settings";

const mocks = vi.hoisted(() => ({
  permission: { canView: true, canCreate: true, canEdit: true, canDelete: true },
  createMutateAsync: vi.fn().mockResolvedValue({}),
  updateMutateAsync: vi.fn().mockResolvedValue({}),
  deleteMutate: vi.fn(),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => mocks.permission,
}));

vi.mock("sonner", () => ({ toast: { success: vi.fn(), error: vi.fn() } }));
vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));

vi.mock("../api/clinics", () => ({
  useGetClinics: () => ({
    data: [{ id: 1, name: "既存病院", phoneNumber: "", email: "", isActive: true }],
    isPending: false,
    isError: false,
  }),
  useCreateClinic: () => ({ mutateAsync: mocks.createMutateAsync, isPending: false }),
  useUpdateClinic: () => ({ mutateAsync: mocks.updateMutateAsync, isPending: false }),
  useDeleteClinic: () => ({ mutate: mocks.deleteMutate, isPending: false }),
}));

function wrapper({ children }: { children: React.ReactNode }) {
  return <MemoryRouter>{children}</MemoryRouter>;
}

describe("useClinicMasterSettings — 権限 ref (mutation 境界の再検査)", () => {
  beforeEach(() => {
    mocks.permission = { canView: true, canCreate: true, canEdit: true, canDelete: true };
    mocks.createMutateAsync.mockClear();
    mocks.updateMutateAsync.mockClear();
    mocks.deleteMutate.mockClear();
  });

  it("canCreate が false に変わった後の新規保存は create mutation を呼ばない", async () => {
    const { result, rerender } = renderHook(() => useClinicMasterSettings(), { wrapper });

    act(() => {
      result.current.handleCreate();
      result.current.setFormData({ ...result.current.formData, name: "新規病院" });
    });

    // UI が権限変化を検知する前に mutation 境界へ届くケースを再現する
    mocks.permission = { ...mocks.permission, canCreate: false };
    rerender();

    await act(async () => {
      result.current.formAction(new FormData());
    });

    await waitFor(() => {
      expect(mocks.createMutateAsync).not.toHaveBeenCalled();
    });
  });

  it("canEdit が false に変わった後の更新保存は update mutation を呼ばない", async () => {
    const { result, rerender } = renderHook(() => useClinicMasterSettings(), { wrapper });

    await waitFor(() => expect(result.current.filteredItems).toHaveLength(1));

    act(() => {
      result.current.handleEdit(result.current.filteredItems[0]);
    });

    mocks.permission = { ...mocks.permission, canEdit: false };
    rerender();

    await act(async () => {
      result.current.formAction(new FormData());
    });

    await waitFor(() => {
      expect(mocks.updateMutateAsync).not.toHaveBeenCalled();
    });
  });

  it("canDelete が false に変わった後の削除確定は delete mutation を呼ばない", async () => {
    const { result, rerender } = renderHook(() => useClinicMasterSettings(), { wrapper });

    await waitFor(() => expect(result.current.filteredItems).toHaveLength(1));

    act(() => {
      result.current.setPendingDelete(result.current.filteredItems[0]);
    });

    mocks.permission = { ...mocks.permission, canDelete: false };
    rerender();

    act(() => {
      result.current.handleDeleteConfirm();
    });

    expect(mocks.deleteMutate).not.toHaveBeenCalled();
  });

  it("権限が揃っている通常経路では新規作成 mutation を呼ぶ", async () => {
    const { result } = renderHook(() => useClinicMasterSettings(), { wrapper });

    act(() => {
      result.current.handleCreate();
      result.current.setFormData({ ...result.current.formData, name: "新規病院" });
    });

    await act(async () => {
      result.current.formAction(new FormData());
    });

    await waitFor(() => {
      expect(mocks.createMutateAsync).toHaveBeenCalledTimes(1);
    });
  });
});
