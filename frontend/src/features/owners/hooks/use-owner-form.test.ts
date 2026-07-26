import { startTransition, useLayoutEffect, useRef } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { createTestWrapper } from "@/testing/utils";
import type { Owner } from "@/types/owner";
import { useOwnerForm } from "./use-owner-form";

const CREATE_PERMISSIONS = {
  canCreate: true,
  canEdit: false,
  canDelete: false,
} as const;

const EDIT_PERMISSIONS = {
  canCreate: false,
  canEdit: true,
  canDelete: false,
} as const;

const { mockCreateOwner, mockUpdateOwner } = vi.hoisted(() => ({
  mockCreateOwner: vi.fn(),
  mockUpdateOwner: vi.fn(),
}));

vi.mock("../api/create-owner", () => ({
  createOwner: mockCreateOwner,
}));

vi.mock("../api/update-owner", () => ({
  updateOwner: mockUpdateOwner,
}));

vi.mock("sonner", () => ({
  toast: {
    success: vi.fn(),
    warning: vi.fn(),
    error: vi.fn(),
  },
}));

function makeOwner(overrides: Partial<Owner> = {}): Owner {
  return {
    id: "123",
    ownerName: "山田太郎",
    ownerNameKana: "ヤマダタロウ",
    company: "",
    postalCode: "",
    address1: "",
    address2: "",
    homePostalCode: "",
    homeAddress1: "",
    homeAddress2: "",
    birthDate: "1980-01-02",
    phone: "090-1234-5678",
    companyPhone: "",
    email: "",
    remarks: "",
    isDangerous: false,
    discountRate: 0,
    membershipType: "non_member",
    deliveryExcluded: false,
    deliveryCaution: false,
    isTransferred: false,
    lstepOptOut: false,
    createdAt: "2026-01-01T00:00:00Z",
    updatedAt: "2026-01-01T00:00:00Z",
    pets: [],
    ...overrides,
  };
}

async function submitForm(formAction: ReturnType<typeof useOwnerForm>["formAction"]) {
  await act(async () => {
    startTransition(() => formAction(new FormData()));
  });
}

describe("useOwnerForm birth_date payload (BUG-432)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateOwner.mockResolvedValue(makeOwner({ id: "new-owner" }));
    mockUpdateOwner.mockResolvedValue(makeOwner());
  });

  it("新規登録時に入力した birth_date を create payload へ送る", async () => {
    const { result } = renderHook(
      () => useOwnerForm(undefined, undefined, undefined, CREATE_PERMISSIONS),
      {
      wrapper: createTestWrapper(),
      },
    );
    act(() => {
      result.current.setOwnerData((previous) => ({
        ...previous,
        ownerName: "山田太郎",
        ownerNameKana: "ヤマダタロウ",
        phone: "090-1234-5678",
        birthDate: "1990-04-01",
      }));
    });

    await submitForm(result.current.formAction);

    await waitFor(() => {
      expect(mockCreateOwner).toHaveBeenCalledWith(
        expect.objectContaining({ birth_date: "1990-04-01" }),
      );
    });
  });

  it("編集時に変更した birth_date を update payload へ送る", async () => {
    const { result } = renderHook(
      () => useOwnerForm("123", makeOwner(), undefined, EDIT_PERMISSIONS),
      {
        wrapper: createTestWrapper(),
      },
    );
    act(() => {
      result.current.setOwnerData((previous) => ({
        ...previous,
        birthDate: "1991-05-02",
      }));
    });

    await submitForm(result.current.formAction);

    await waitFor(() => {
      expect(mockUpdateOwner).toHaveBeenCalledWith(
        "123",
        expect.objectContaining({ birth_date: "1991-05-02" }),
      );
    });
  });

  it("編集時に空へ戻した birth_date は null として送る", async () => {
    const { result } = renderHook(
      () => useOwnerForm("123", makeOwner(), undefined, EDIT_PERMISSIONS),
      {
        wrapper: createTestWrapper(),
      },
    );
    act(() => {
      result.current.setOwnerData((previous) => ({
        ...previous,
        birthDate: "",
      }));
    });

    await submitForm(result.current.formAction);

    await waitFor(() => {
      expect(mockUpdateOwner).toHaveBeenCalledWith(
        "123",
        expect.objectContaining({ birth_date: null }),
      );
    });
  });
});

describe("useOwnerForm mutation permission boundary (FE12-02 C6a)", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateOwner.mockResolvedValue(makeOwner({ id: "new-owner" }));
    mockUpdateOwner.mockResolvedValue(makeOwner());
  });

  it("作成権限剥奪をcommitしたlayout phaseで取得済みformActionが発火してもcreateOwnerを呼ばない", async () => {
    const { result, rerender } = renderHook(
      ({ canCreate }: { canCreate: boolean }) => {
        const form = useOwnerForm(undefined, undefined, undefined, {
          canCreate,
          canEdit: false,
          canDelete: false,
        });
        const capturedActionRef = useRef(form.formAction);
        useLayoutEffect(() => {
          if (!canCreate) {
            startTransition(() => capturedActionRef.current(new FormData()));
          }
        }, [canCreate]);
        return form;
      },
      {
        initialProps: { canCreate: true },
        wrapper: createTestWrapper(),
      },
    );
    act(() => {
      result.current.setOwnerData((previous) => ({
        ...previous,
        ownerName: "山田太郎",
        ownerNameKana: "ヤマダタロウ",
        phone: "090-1234-5678",
      }));
    });
    const initialTimestamp = result.current.formState.timestamp;

    await act(async () => {
      rerender({ canCreate: false });
    });

    await waitFor(() => {
      expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
    });
    expect(mockCreateOwner).not.toHaveBeenCalled();
  });

  it("更新権限剥奪をcommitしたlayout phaseで取得済みformActionが発火してもupdateOwnerを呼ばない", async () => {
    const { result, rerender } = renderHook(
      ({ canEdit }: { canEdit: boolean }) => {
        const form = useOwnerForm("123", makeOwner(), undefined, {
          canCreate: false,
          canEdit,
          canDelete: false,
        });
        const capturedActionRef = useRef(form.formAction);
        useLayoutEffect(() => {
          if (!canEdit) {
            startTransition(() => capturedActionRef.current(new FormData()));
          }
        }, [canEdit]);
        return form;
      },
      {
        initialProps: { canEdit: true },
        wrapper: createTestWrapper(),
      },
    );
    const initialTimestamp = result.current.formState.timestamp;

    await act(async () => {
      rerender({ canEdit: false });
    });

    await waitFor(() => {
      expect(result.current.formState.timestamp).not.toBe(initialTimestamp);
    });
    expect(mockUpdateOwner).not.toHaveBeenCalled();
  });
});
