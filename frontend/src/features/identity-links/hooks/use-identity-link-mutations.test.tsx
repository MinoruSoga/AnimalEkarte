import { act, renderHook, waitFor } from "@testing-library/react";
import { useQueryClient } from "@tanstack/react-query";
import { describe, expect, it, vi, beforeEach } from "vitest";

import { createTestWrapper } from "@/testing/utils";
import { queryKeys } from "@/lib/query-keys";

import {
  useCreateOwnerLink,
  useCreatePetLink,
  useLinkedTreatmentHistory,
  useUnlinkOwnerMember,
  useUnlinkPetMember,
} from "./use-identity-link-mutations";

const createOwnerIdentityGroup = vi.fn();
const unlinkOwnerIdentityMember = vi.fn();
const createPetIdentityGroup = vi.fn();
const unlinkPetIdentityMember = vi.fn();
const getLinkedTreatmentHistory = vi.fn();

vi.mock("../api/identity-links-api", () => ({
  createOwnerIdentityGroup: (...args: unknown[]) => createOwnerIdentityGroup(...args),
  unlinkOwnerIdentityMember: (...args: unknown[]) => unlinkOwnerIdentityMember(...args),
  createPetIdentityGroup: (...args: unknown[]) => createPetIdentityGroup(...args),
  unlinkPetIdentityMember: (...args: unknown[]) => unlinkPetIdentityMember(...args),
  getLinkedTreatmentHistory: (...args: unknown[]) => getLinkedTreatmentHistory(...args),
}));

beforeEach(() => {
  createOwnerIdentityGroup.mockReset();
  unlinkOwnerIdentityMember.mockReset();
  createPetIdentityGroup.mockReset();
  unlinkPetIdentityMember.mockReset();
  getLinkedTreatmentHistory.mockReset();
});

describe("useCreateOwnerLink", () => {
  it("成功時に各メンバーの逆引きキャッシュへ新グループを書き込む", async () => {
    const group = { id: 42, created_clinic_id: 1, version: 1, members: [] };
    createOwnerIdentityGroup.mockResolvedValue(group);
    const wrapper = createTestWrapper();

    const { result } = renderHook(
      () => ({ mutation: useCreateOwnerLink(), queryClient: useQueryClient() }),
      { wrapper },
    );

    await act(async () => {
      await result.current.mutation.mutateAsync([{ clinic_id: 1, owner_id: 10 }]);
    });

    expect(
      result.current.queryClient.getQueryData(queryKeys.identityLinks.ownerGroup(1, 10)),
    ).toEqual(group);
  });
});

describe("useUnlinkOwnerMember", () => {
  it("成功時に該当メンバーの逆引きキャッシュを null にする", async () => {
    unlinkOwnerIdentityMember.mockResolvedValue(undefined);
    const wrapper = createTestWrapper();

    const { result } = renderHook(
      () => ({ mutation: useUnlinkOwnerMember(), queryClient: useQueryClient() }),
      { wrapper },
    );

    result.current.queryClient.setQueryData(queryKeys.identityLinks.ownerGroup(1, 10), {
      id: 42,
    });

    await act(async () => {
      await result.current.mutation.mutateAsync({
        groupId: 42,
        member: { clinic_id: 1, owner_id: 10 },
      });
    });

    expect(unlinkOwnerIdentityMember).toHaveBeenCalledWith(42, { clinic_id: 1, owner_id: 10 });
    expect(
      result.current.queryClient.getQueryData(queryKeys.identityLinks.ownerGroup(1, 10)),
    ).toBeNull();
  });
});

describe("useCreatePetLink", () => {
  it("親飼主グループ id を渡して作成し、成功時にキャッシュへ書き込む", async () => {
    const group = { id: 77, owner_group_id: 42 };
    createPetIdentityGroup.mockResolvedValue(group);
    const wrapper = createTestWrapper();

    const { result } = renderHook(
      () => ({ mutation: useCreatePetLink(), queryClient: useQueryClient() }),
      { wrapper },
    );

    await act(async () => {
      await result.current.mutation.mutateAsync({
        ownerGroupId: 42,
        members: [{ clinic_id: 1, pet_id: 5 }],
      });
    });

    expect(createPetIdentityGroup).toHaveBeenCalledWith(42, [{ clinic_id: 1, pet_id: 5 }]);
    expect(
      result.current.queryClient.getQueryData(queryKeys.identityLinks.petGroup(1, 5)),
    ).toEqual(group);
  });
});

describe("useUnlinkPetMember", () => {
  it("成功時に該当メンバーの逆引きキャッシュを null にする", async () => {
    unlinkPetIdentityMember.mockResolvedValue(undefined);
    const wrapper = createTestWrapper();

    const { result } = renderHook(
      () => ({ mutation: useUnlinkPetMember(), queryClient: useQueryClient() }),
      { wrapper },
    );

    await act(async () => {
      await result.current.mutation.mutateAsync({
        groupId: 77,
        member: { clinic_id: 1, pet_id: 5 },
      });
    });

    expect(unlinkPetIdentityMember).toHaveBeenCalledWith(77, { clinic_id: 1, pet_id: 5 });
    expect(
      result.current.queryClient.getQueryData(queryKeys.identityLinks.petGroup(1, 5)),
    ).toBeNull();
  });
});

describe("useLinkedTreatmentHistory", () => {
  it("clinicId/petId を渡して履歴を取得する", async () => {
    getLinkedTreatmentHistory.mockResolvedValue({ items: [], total: 0, page: 1, limit: 20 });
    const wrapper = createTestWrapper();

    const { result } = renderHook(() => useLinkedTreatmentHistory(), { wrapper });

    await act(async () => {
      await result.current.mutateAsync({ clinicId: 1, petId: 5 });
    });

    await waitFor(() => {
      expect(getLinkedTreatmentHistory).toHaveBeenCalledWith(1, 5, true, 1, 20);
    });
  });
});
