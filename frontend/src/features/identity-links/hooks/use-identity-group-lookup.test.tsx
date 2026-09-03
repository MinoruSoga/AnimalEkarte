import { renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";

import { createTestWrapper } from "@/testing/TestUtils";
import type { OwnerSearchItem, PetSearchItem } from "@/types/generated/identitylink-responses";

import {
  ownerMemberKey,
  petMemberKey,
  useOwnerGroupLookup,
  usePetGroupLookup,
} from "./use-identity-group-lookup";

const findOwnerIdentityGroupByMember = vi.fn();
const findPetIdentityGroupByMember = vi.fn();

vi.mock("../api/identity-links-api", () => ({
  findOwnerIdentityGroupByMember: (...args: unknown[]) => findOwnerIdentityGroupByMember(...args),
  findPetIdentityGroupByMember: (...args: unknown[]) => findPetIdentityGroupByMember(...args),
}));

const ownerA: OwnerSearchItem = {
  clinic_id: 1,
  owner_id: 10,
  name: "佐藤太郎",
  name_kana: "サトウタロウ",
  phone: "09011112222",
};
const ownerB: OwnerSearchItem = {
  clinic_id: 1,
  owner_id: 99,
  name: "未所属花子",
  name_kana: "ミショゾクハナコ",
  phone: "09099998888",
};

const pet: PetSearchItem = {
  clinic_id: 1,
  pet_id: 5,
  owner_id: 10,
  name: "ポチ",
  name_kana: "ポチ",
  pet_number: "P-001",
};

describe("useOwnerGroupLookup", () => {
  beforeEach(() => {
    findOwnerIdentityGroupByMember.mockReset();
    findPetIdentityGroupByMember.mockReset();
  });

  it("選択が無いときは何も解決しない", () => {
    const { result } = renderHook(() => useOwnerGroupLookup([]), {
      wrapper: createTestWrapper(),
    });

    expect(result.current.groupIdsByMember).toEqual({});
    expect(result.current.sessionGroupId).toBeNull();
    expect(findOwnerIdentityGroupByMember).not.toHaveBeenCalled();
  });

  it("既存グループに所属するメンバーだけを解決する（未所属は含めない）", async () => {
    findOwnerIdentityGroupByMember.mockImplementation(async (_clinicId: number, ownerId: number) => {
      if (ownerId === 10) {
        return { id: 42, created_clinic_id: 1, version: 1, members: [{ clinic_id: 1, owner_id: 10 }] };
      }
      return null;
    });

    const { result } = renderHook(() => useOwnerGroupLookup([ownerA, ownerB]), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      expect(result.current.groupIdsByMember[ownerMemberKey(1, 10)]).toBe(42);
    });
    expect(result.current.groupIdsByMember[ownerMemberKey(1, 99)]).toBeUndefined();
    expect(result.current.sessionGroupId).toBe(42);
  });

  it("選択解除するとそのメンバーの結果はマップから消える", async () => {
    findOwnerIdentityGroupByMember.mockResolvedValue({
      id: 42,
      created_clinic_id: 1,
      version: 1,
      members: [{ clinic_id: 1, owner_id: 10 }],
    });

    const { result, rerender } = renderHook(
      ({ owners }: { owners: OwnerSearchItem[] }) => useOwnerGroupLookup(owners),
      { wrapper: createTestWrapper(), initialProps: { owners: [ownerA] } },
    );

    await waitFor(() => {
      expect(result.current.groupIdsByMember[ownerMemberKey(1, 10)]).toBe(42);
    });

    rerender({ owners: [] });

    expect(result.current.groupIdsByMember).toEqual({});
  });

  it("逆引きが失敗した選択中メンバーがいればエラーメッセージを返す", async () => {
    findOwnerIdentityGroupByMember.mockRejectedValue(new Error("network"));

    const { result } = renderHook(() => useOwnerGroupLookup([ownerA]), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      expect(result.current.errorMessage).toBe("飼主グループ取得に失敗しました");
    });
  });
});

describe("usePetGroupLookup", () => {
  beforeEach(() => {
    findOwnerIdentityGroupByMember.mockReset();
    findPetIdentityGroupByMember.mockReset();
  });

  it("解決したペットグループの親飼主グループ id を sessionOwnerGroupId として返す", async () => {
    findPetIdentityGroupByMember.mockResolvedValue({
      id: 77,
      created_clinic_id: 1,
      owner_group_created_clinic_id: 1,
      owner_group_id: 42,
      version: 1,
      members: [{ clinic_id: 1, pet_id: 5 }],
    });

    const { result } = renderHook(() => usePetGroupLookup([pet]), {
      wrapper: createTestWrapper(),
    });

    await waitFor(() => {
      expect(result.current.groupIdsByMember[petMemberKey(1, 5)]).toBe(77);
    });
    expect(result.current.sessionOwnerGroupId).toBe(42);
  });
});
