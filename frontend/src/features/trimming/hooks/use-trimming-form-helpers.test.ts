import { describe, it, expect, vi, beforeEach } from "vitest";
import { toast } from "sonner";
import { createTrimmingDeleteHandler } from "./use-trimming-form-helpers";

vi.mock("sonner", () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

vi.mock("@/lib/handle-api-error", () => ({ handleApiError: vi.fn() }));

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";
const DECEASED_SAVE_MESSAGE = "死亡したペットのトリミング記録は削除できません";

describe("createTrimmingDeleteHandler (FE-RC-101/102)", () => {
  const deleteTrimming = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("canDelete=false では delete せず toast.error する", () => {
    createTrimmingDeleteHandler({
      isEdit: true,
      id: "15",
      isMutationAllowed: (action) => action !== "canDelete",
      isEditPetReady: () => true,
      isPetDeceased: () => false,
      deleteTrimming,
    })();

    expect(deleteTrimming).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("死亡ペットでは delete せず toast.error する", () => {
    createTrimmingDeleteHandler({
      isEdit: true,
      id: "15",
      isMutationAllowed: () => true,
      isEditPetReady: () => true,
      isPetDeceased: () => true,
      deleteTrimming,
    })();

    expect(deleteTrimming).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith(DECEASED_SAVE_MESSAGE);
  });

  it("pet 未着では delete せず toast.error する", () => {
    createTrimmingDeleteHandler({
      isEdit: true,
      id: "15",
      isMutationAllowed: () => true,
      isEditPetReady: () => false,
      isPetDeceased: () => false,
      deleteTrimming,
    })();

    expect(deleteTrimming).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("ペット情報の読み込みが完了してから削除してください");
  });

  it("権限ありかつ生存ペットでは delete する", () => {
    createTrimmingDeleteHandler({
      isEdit: true,
      id: "15",
      isMutationAllowed: () => true,
      isEditPetReady: () => true,
      isPetDeceased: () => false,
      deleteTrimming,
    })();

    expect(deleteTrimming).toHaveBeenCalledWith(
      "15",
      expect.objectContaining({
        onSuccess: expect.any(Function),
        onError: expect.any(Function),
      }),
    );
    expect(toast.error).not.toHaveBeenCalled();
  });
});
