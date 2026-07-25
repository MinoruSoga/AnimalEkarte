import { startTransition } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { createTestWrapper } from "@/testing/utils";
import type { Owner } from "@/types/owner";
import { useOwnerForm } from "./use-owner-form";

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
    const { result } = renderHook(() => useOwnerForm(), {
      wrapper: createTestWrapper(),
    });
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
    const { result } = renderHook(() => useOwnerForm("123", makeOwner()), {
      wrapper: createTestWrapper(),
    });
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
    const { result } = renderHook(() => useOwnerForm("123", makeOwner()), {
      wrapper: createTestWrapper(),
    });
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
