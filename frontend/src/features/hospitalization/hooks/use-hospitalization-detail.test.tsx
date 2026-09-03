import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { useLayoutEffect, useState } from "react";

import { useHospitalizationDetail } from "./use-hospitalization-detail";

const mocks = vi.hoisted(() => ({
  canEdit: true,
  petIsDeceased: false,
  updateHospitalization: vi.fn(),
  dischargeWithBilling: vi.fn(),
  handleApiError: vi.fn(),
}));

vi.mock("@/lib/handle-api-error", () => ({
  handleApiError: mocks.handleApiError,
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

vi.mock("../api/get-hospitalization", () => ({
  useGetHospitalization: () => ({
    data: { id: "42", petId: "7", petIsDeceased: mocks.petIsDeceased },
    isLoading: false,
    isError: false,
    error: null,
  }),
}));

vi.mock("../api/update-hospitalization", () => ({
  useUpdateHospitalization: () => ({
    mutateAsync: mocks.updateHospitalization,
  }),
}));

vi.mock("../api/discharge-with-billing", () => ({
  dischargeWithBilling: mocks.dischargeWithBilling,
}));

beforeEach(() => {
  mocks.canEdit = true;
  mocks.petIsDeceased = false;
  mocks.updateHospitalization.mockReset();
  mocks.updateHospitalization.mockResolvedValue(undefined);
  mocks.dischargeWithBilling.mockReset();
  mocks.dischargeWithBilling.mockResolvedValue({ accounting_id: 99 });
  mocks.handleApiError.mockReset();
});

describe("useHospitalizationDetail — discharge permission boundary", () => {
  it("petが死亡している場合は会計あり・なしの両退院mutationを拒否する", async () => {
    mocks.petIsDeceased = true;
    const { result } = renderHook(() => useHospitalizationDetail("42"));

    await act(async () => {
      expect(await result.current.dischargeHospitalization(false)).toEqual({
        success: false,
      });
      expect(await result.current.dischargeHospitalization(true)).toEqual({
        success: false,
      });
    });

    expect(mocks.updateHospitalization).not.toHaveBeenCalled();
    expect(mocks.dischargeWithBilling).not.toHaveBeenCalled();
  });

  it("同一commitで編集権限が失効した場合、captured discharge callbackは両mutation経路を拒否する", async () => {
    const { result } = renderHook(() => {
      const [revoked, setRevoked] = useState(false);
      const detail = useHospitalizationDetail("42");
      const [capturedDischarge] = useState(() => detail.dischargeHospitalization);

      useLayoutEffect(() => {
        if (revoked) {
          void capturedDischarge(false);
          void capturedDischarge(true);
        }
      }, [capturedDischarge, revoked]);

      return {
        revoke: () => {
          mocks.canEdit = false;
          setRevoked(true);
        },
      };
    });

    await act(async () => {
      result.current.revoke();
    });

    expect(mocks.updateHospitalization).not.toHaveBeenCalled();
    expect(mocks.dischargeWithBilling).not.toHaveBeenCalled();
  });

  // BUG-457: guard は canEdit。delete が true でも edit が false なら API 0 回。
  it("canEdit:false かつ createAccounting=false で退院は {success:false} かつ updateHospitalization を呼ばない", async () => {
    mocks.canEdit = false;
    const { result } = renderHook(() => useHospitalizationDetail("42"));

    await act(async () => {
      expect(await result.current.dischargeHospitalization(false)).toEqual({
        success: false,
      });
    });

    expect(mocks.updateHospitalization).toHaveBeenCalledTimes(0);
    expect(mocks.dischargeWithBilling).toHaveBeenCalledTimes(0);
  });

  it("canEdit:false かつ createAccounting=true で退院は {success:false} かつ dischargeWithBilling を呼ばない", async () => {
    mocks.canEdit = false;
    const { result } = renderHook(() => useHospitalizationDetail("42"));

    await act(async () => {
      expect(await result.current.dischargeHospitalization(true)).toEqual({
        success: false,
      });
    });

    expect(mocks.dischargeWithBilling).toHaveBeenCalledTimes(0);
    expect(mocks.updateHospitalization).toHaveBeenCalledTimes(0);
  });

  // FE-RC-005: useUpdateHospitalization.onError が既に handleApiError 済みのため、
  // dischargeHospitalization 側の catch で再度呼んではならない（二重トースト回避）。
  it("createAccounting=false でupdateHospitalizationが失敗してもhandleApiErrorを呼ばない（重複トースト防止）", async () => {
    mocks.updateHospitalization.mockRejectedValueOnce(new Error("network error"));
    const { result } = renderHook(() => useHospitalizationDetail("42"));

    await act(async () => {
      expect(await result.current.dischargeHospitalization(false)).toEqual({
        success: false,
      });
    });

    expect(mocks.updateHospitalization).toHaveBeenCalledTimes(1);
    expect(mocks.handleApiError).not.toHaveBeenCalled();
  });

  // dischargeWithBilling は mutation ではない生の非同期呼び出しのため、
  // ここでの handleApiError が唯一の通知経路として維持される。
  it("createAccounting=trueでdischargeWithBillingが失敗した場合はhandleApiErrorを呼ぶ（唯一の通知経路）", async () => {
    mocks.dischargeWithBilling.mockRejectedValueOnce(new Error("network error"));
    const { result } = renderHook(() => useHospitalizationDetail("42"));

    await act(async () => {
      expect(await result.current.dischargeHospitalization(true)).toEqual({
        success: false,
      });
    });

    expect(mocks.dischargeWithBilling).toHaveBeenCalledTimes(1);
    expect(mocks.handleApiError).toHaveBeenCalledTimes(1);
    expect(mocks.handleApiError).toHaveBeenCalledWith(expect.any(Error), "退院処理");
  });
});
