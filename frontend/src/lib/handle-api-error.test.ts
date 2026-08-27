import { beforeEach, describe, expect, it, vi } from "vitest";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";
import { toast } from "sonner";

import {
  CONFLICT_CODE_ANIMAL_SPECIES_NAME,
  CONFLICT_CODE_CAGE_NAME,
  CONFLICT_CODE_CHECKUP_TYPE_NAME,
  CONFLICT_CODE_CONSULTATION_NAME,
  CONFLICT_CODE_EXAM_TYPE_NAME,
  CONFLICT_CODE_LSTEP_AUTO_MANAGED_PREFIX,
  CONFLICT_CODE_PERMISSION_GROUP_NAME,
  CONFLICT_CODE_PROCEDURE_NAME,
  CONFLICT_CODE_SHIFT_TEMPLATE_NAME,
  CONFLICT_CODE_VACCINE_NAME,
  CONFLICT_CODE_MEDICINE_NAME,
  extractApiErrorMessage,
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

  it("maps consultation_name_conflict with the actual item name (BUG-017)", () => {
    expect(
      localizeConflictMessage(CONFLICT_CODE_CONSULTATION_NAME, {
        name: "V04診察",
      }),
    ).toBe("診察『V04診察』は既に使用されています");
    expect(
      localizeConflictMessage(CONFLICT_CODE_EXAM_TYPE_NAME, { name: "V04検査" }),
    ).toBe("検査『V04検査』は既に使用されています");
    expect(
      localizeConflictMessage(CONFLICT_CODE_PROCEDURE_NAME, { name: "V04処置" }),
    ).toBe("処置『V04処置』は既に使用されています");
    expect(
      localizeConflictMessage(CONFLICT_CODE_VACCINE_NAME, { name: "V04予防接種" }),
    ).toBe("予防接種『V04予防接種』は既に使用されています");
    expect(
      localizeConflictMessage(CONFLICT_CODE_CHECKUP_TYPE_NAME, {
        name: "V04定期健診",
      }),
    ).toBe("定期健診『V04定期健診』は既に使用されています");
    expect(
      localizeConflictMessage(CONFLICT_CODE_MEDICINE_NAME, { name: "V04薬剤" }),
    ).toBe("薬剤『V04薬剤』は既に使用されています");
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
    expect(toast.error).toHaveBeenCalledWith("既に登録されています");
    expect(localizeAlreadyExistsMessage("occupation '' already exists")).toBe(
      "既に登録されています",
    );
    expect(localizeAlreadyExistsMessage("consultation '' already exists")).toBe(
      "既に登録されています",
    );
    expect(localizeAlreadyExistsMessage("consultation '' already exists")).not.toBe(
      "診察は既に使用されています",
    );
    expect(localizeAlreadyExistsMessage("cage 'ICU-1' already exists")).toBe(
      "ケージ『ICU-1』は既に使用されています",
    );
    expect(localizeAlreadyExistsMessage("lstep_condition_tag_mapping 'ckd' already exists")).toBe(
      "慢性疾患コード『ckd』は既に使用されています",
    );
  });

  it("shows the actual consultation name, not the tab label (BUG-017)", () => {
    handleApiError(
      axiosError(409, {
        error: "consultation '' already exists",
        code: CONFLICT_CODE_CONSULTATION_NAME,
        params: { name: "V04診察" },
      }),
      "保存",
    );
    expect(toast.error).toHaveBeenCalledWith(
      "診察『V04診察』は既に使用されています",
    );
    expect(toast.error).not.toHaveBeenCalledWith("診察は既に使用されています");
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

  it("英語メッセージ＋未知コードの 409 で日本語フォールバックを返す", () => {
    const message = extractApiErrorMessage(
      axiosError(409, {
        error: "pet owner is not in the specified owner identity group",
        code: "owner_identity_group_mismatch",
      }),
      "リンク",
    );
    expect(message).toBe(
      "他のユーザーによって更新されています。一度リロードしてください。",
    );
    expect(message).not.toMatch(/pet owner/i);
    handleApiError(
      axiosError(409, {
        error: "pet owner is not in the specified owner identity group",
        code: "owner_identity_group_mismatch",
      }),
      "リンク",
    );
    expect(toast.error).toHaveBeenCalledWith(
      "他のユーザーによって更新されています。一度リロードしてください。",
    );
    expect(toast.error).not.toHaveBeenCalledWith(
      expect.stringContaining("pet owner"),
    );
  });

  it("passes through Japanese reservation-style 409 server messages", () => {
    const reservationMessage =
      "この時間帯はすでに予約が入っています";
    expect(
      extractApiErrorMessage(
        axiosError(409, {
          error: reservationMessage,
          code: "reservation_slot_conflict",
        }),
        "予約",
      ),
    ).toBe(reservationMessage);
    expect(
      extractApiErrorMessage(
        axiosError(409, {
          error: "担当可能な医師がいません",
        }),
        "予約",
      ),
    ).toBe("担当可能な医師がいません");
  });
});

describe("extractApiErrorMessage 400/403/404 Japanese guard (BUG-006)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("400 English serverMessage falls back to Japanese and must not contain English", () => {
    const english =
      "appointment must use a general reservation type for a medical record";
    const message = extractApiErrorMessage(
      axiosError(400, { error: english }),
      "カルテ作成",
    );
    expect(message).toBe(
      "カルテ作成に失敗しました。入力内容を確認してください。",
    );
    expect(message).not.toMatch(/appointment|reservation|medical record/i);
    expect(message).not.toContain(english);
    handleApiError(axiosError(400, { error: english }), "カルテ作成");
    expect(toast.error).toHaveBeenCalledWith(
      "カルテ作成に失敗しました。入力内容を確認してください。",
    );
    expect(toast.error).not.toHaveBeenCalledWith(english);
  });

  it("400 Japanese serverMessage is passed through", () => {
    const ja = "入力内容が正しくありません";
    expect(
      extractApiErrorMessage(axiosError(400, { error: ja }), "カルテ作成"),
    ).toBe(ja);
  });

  it("403 English serverMessage falls back to Japanese", () => {
    const english = "forbidden: insufficient permissions";
    const message = extractApiErrorMessage(
      axiosError(403, { error: english }),
      "削除",
    );
    expect(message).toBe("削除の権限がありません。");
    expect(message).not.toMatch(/forbidden|insufficient|permissions/i);
  });

  it("403 Japanese serverMessage is passed through", () => {
    const ja = "この操作を行う権限がありません";
    expect(
      extractApiErrorMessage(axiosError(403, { error: ja }), "削除"),
    ).toBe(ja);
  });

  it("404 English serverMessage falls back to Japanese", () => {
    const english = "record not found";
    const message = extractApiErrorMessage(
      axiosError(404, { error: english }),
      "取得",
    );
    expect(message).toBe("取得対象が見つかりません。");
    expect(message).not.toMatch(/record not found/i);
  });

  it("404 Japanese serverMessage is passed through", () => {
    const ja = "指定されたカルテが見つかりません";
    expect(
      extractApiErrorMessage(axiosError(404, { error: ja }), "取得"),
    ).toBe(ja);
  });
});
