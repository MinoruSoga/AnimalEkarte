import { beforeEach, describe, expect, it, vi } from "vitest";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { toast } from "sonner";

import {
  CONFLICT_CODE_ANIMAL_SPECIES_NAME,
  CONFLICT_CODE_CAGE_NAME,
  CONFLICT_CODE_LSTEP_AUTO_MANAGED_PREFIX,
  CONFLICT_CODE_PERMISSION_GROUP_NAME,
  CONFLICT_CODE_SHIFT_TEMPLATE_NAME,
  handleApiError,
  localizeAlreadyExistsMessage,
  localizeConflictMessage,
} from "./handle-api-error";

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

function axiosError(
  status: number,
  data: Record<string, unknown>,
): AxiosError {
  const config = {
    headers: new AxiosHeaders(),
  } as InternalAxiosRequestConfig;
  return new AxiosError(
    "request failed",
    AxiosError.ERR_BAD_RESPONSE,
    config,
    undefined,
    {
      config,
      data,
      headers: new AxiosHeaders(),
      status,
      statusText: "Conflict",
    },
  );
}

describe("localizeConflictMessage", () => {
  it("maps permission_group_name_conflict with name", () => {
    expect(
      localizeConflictMessage(CONFLICT_CODE_PERMISSION_GROUP_NAME, {
        name: "執行",
      }),
    ).toBe("権限グループ名『執行』は既に使用されています");
  });

  it("maps animal_species_name_conflict with name", () => {
    expect(
      localizeConflictMessage(CONFLICT_CODE_ANIMAL_SPECIES_NAME, {
        name: "V04動物種類",
      }),
    ).toBe("動物種類『V04動物種類』は既に使用されています");
  });

  it("maps shift_template_name_conflict with name (BUG-026)", () => {
    expect(
      localizeConflictMessage(CONFLICT_CODE_SHIFT_TEMPLATE_NAME, {
        name: "早番",
      }),
    ).toBe("シフトテンプレート名『早番』は既に使用されています");
  });

  it("maps lstep_auto_managed_prefix_conflict with name (BUG-026)", () => {
    expect(
      localizeConflictMessage(CONFLICT_CODE_LSTEP_AUTO_MANAGED_PREFIX, {
        name: "checkup_",
      }),
    ).toBe("自動管理プレフィックス『checkup_』は既に使用されています");
  });

  it("returns null for unknown code (keep fallback)", () => {
    expect(
      localizeConflictMessage("some_other_conflict", { name: "X" }),
    ).toBeNull();
  });

  it("returns null when code is missing", () => {
    expect(localizeConflictMessage(undefined, { name: "X" })).toBeNull();
  });

  it("returns null when params name is empty (no empty 『』)", () => {
    expect(
      localizeConflictMessage(CONFLICT_CODE_PERMISSION_GROUP_NAME, {
        name: "   ",
      }),
    ).toBeNull();
    expect(
      localizeConflictMessage(CONFLICT_CODE_ANIMAL_SPECIES_NAME, {}),
    ).toBeNull();
  });
});

describe("handleApiError 409 localization", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows Japanese message for permission group name conflict and ignores raw English error", () => {
    handleApiError(
      axiosError(409, {
        error: "permission_group '' already exists",
        code: CONFLICT_CODE_PERMISSION_GROUP_NAME,
        params: { name: "執行" },
      }),
      "保存",
    );
    expect(toast.error).toHaveBeenCalledWith(
      "権限グループ名『執行』は既に使用されています",
    );
    expect(toast.error).not.toHaveBeenCalledWith(
      expect.stringContaining("already exists"),
    );
  });

  it("shows Japanese message for animal species name conflict", () => {
    handleApiError(
      axiosError(409, {
        error: "animal_species '' already exists",
        code: CONFLICT_CODE_ANIMAL_SPECIES_NAME,
        params: { name: "V04動物種類" },
      }),
      "保存",
    );
    expect(toast.error).toHaveBeenCalledWith(
      "動物種類『V04動物種類』は既に使用されています",
    );
  });

  it("shows Japanese message for shift template name conflict (BUG-026)", () => {
    handleApiError(
      axiosError(409, {
        error: "shift_template '' already exists",
        code: CONFLICT_CODE_SHIFT_TEMPLATE_NAME,
        params: { name: "早番" },
      }),
      "シフトテンプレートの作成",
    );
    expect(toast.error).toHaveBeenCalledWith(
      "シフトテンプレート名『早番』は既に使用されています",
    );
    expect(toast.error).not.toHaveBeenCalledWith(
      expect.stringContaining("already exists"),
    );
  });

  it("shows Japanese message for lstep auto managed prefix conflict (BUG-026)", () => {
    handleApiError(
      axiosError(409, {
        error: "lstep_auto_managed_prefix '' already exists",
        code: CONFLICT_CODE_LSTEP_AUTO_MANAGED_PREFIX,
        params: { name: "checkup_" },
      }),
      "自動管理プレフィックスの追加",
    );
    expect(toast.error).toHaveBeenCalledWith(
      "自動管理プレフィックス『checkup_』は既に使用されています",
    );
  });

  it("keeps serverMessage for unknown conflict code (non-regression)", () => {
    handleApiError(
      axiosError(409, {
        error: "この権限グループはスタッフに割り当てられているため削除できません",
        code: "CONFLICT",
      }),
      "削除",
    );
    expect(toast.error).toHaveBeenCalledWith(
      "この権限グループはスタッフに割り当てられているため削除できません",
    );
  });

  it("falls back when code present but params name empty", () => {
    handleApiError(
      axiosError(409, {
        error: "resource already exists",
        code: CONFLICT_CODE_PERMISSION_GROUP_NAME,
        params: { name: "" },
      }),
      "保存",
    );
    expect(toast.error).toHaveBeenCalledWith("既に登録されています");
  });

  it("localizes English already-exists messages without a domain code (BUG-022)", () => {
    handleApiError(
      axiosError(409, {
        error: "cage '' already exists",
      }),
      "保存",
    );
    expect(toast.error).toHaveBeenCalledWith("ケージは既に使用されています");
    expect(localizeAlreadyExistsMessage("occupation '' already exists")).toBe(
      "職種は既に使用されています",
    );
    expect(localizeAlreadyExistsMessage("cage 'ICU-1' already exists")).toBe(
      "ケージ『ICU-1』は既に使用されています",
    );
    expect(localizeAlreadyExistsMessage("lstep_condition_tag_mapping 'ckd' already exists")).toBe(
      "慢性疾患コード『ckd』は既に使用されています",
    );
  });

  it("shows Japanese message for cage name conflict code", () => {
    handleApiError(
      axiosError(409, {
        error: "resource already exists",
        code: CONFLICT_CODE_CAGE_NAME,
        params: { name: "ICU-1" },
      }),
      "保存",
    );
    expect(toast.error).toHaveBeenCalledWith("ケージ『ICU-1』は既に使用されています");
  });

  it("falls back to generic 409 message when code and serverMessage missing", () => {
    handleApiError(axiosError(409, {}), "保存");
    expect(toast.error).toHaveBeenCalledWith(
      "他のユーザーによって更新されています。一度リロードしてください。",
    );
  });
});
