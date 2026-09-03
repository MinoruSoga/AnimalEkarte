import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi, beforeEach } from "vitest";
import { toast } from "sonner";

import { createTestWrapper } from "@/testing/TestUtils";
import type { OwnerSearchItem, PetSearchItem } from "@/types/generated/identitylink-responses";

import { useIdentityLinksWorkbench } from "./use-identity-links-workbench";

const PERMISSION_DENIED_MESSAGE = "この操作を行う権限がありません";

const createOwnerIdentityGroup = vi.fn();
const unlinkOwnerIdentityMember = vi.fn();
const createPetIdentityGroup = vi.fn();
const unlinkPetIdentityMember = vi.fn();
const getLinkedTreatmentHistory = vi.fn();
const findOwnerIdentityGroupByMember = vi.fn();
const findPetIdentityGroupByMember = vi.fn();

vi.mock("sonner", () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

vi.mock("../api/identity-links-api", () => ({
  searchOwnersForLink: vi.fn().mockResolvedValue([]),
  searchPetsForLink: vi.fn().mockResolvedValue([]),
  createOwnerIdentityGroup: (...args: unknown[]) => createOwnerIdentityGroup(...args),
  unlinkOwnerIdentityMember: (...args: unknown[]) => unlinkOwnerIdentityMember(...args),
  createPetIdentityGroup: (...args: unknown[]) => createPetIdentityGroup(...args),
  unlinkPetIdentityMember: (...args: unknown[]) => unlinkPetIdentityMember(...args),
  getLinkedTreatmentHistory: (...args: unknown[]) => getLinkedTreatmentHistory(...args),
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

const petA: PetSearchItem = {
  clinic_id: 1,
  pet_id: 5,
  owner_id: 10,
  name: "ポチ",
  name_kana: "ポチ",
  pet_number: "P-001",
};

const petB: PetSearchItem = {
  clinic_id: 1,
  pet_id: 6,
  owner_id: 10,
  name: "タマ",
  name_kana: "タマ",
  pet_number: "P-002",
};

describe("useIdentityLinksWorkbench permission re-check", () => {
  beforeEach(() => {
    createOwnerIdentityGroup.mockReset().mockResolvedValue({ id: 42, members: [] });
    unlinkOwnerIdentityMember.mockReset().mockResolvedValue(undefined);
    createPetIdentityGroup.mockReset().mockResolvedValue({ id: 77, owner_group_id: 42 });
    unlinkPetIdentityMember.mockReset().mockResolvedValue(undefined);
    getLinkedTreatmentHistory.mockReset();
    findOwnerIdentityGroupByMember.mockReset().mockResolvedValue({
      id: 42,
      created_clinic_id: 1,
      version: 1,
      members: [{ clinic_id: 1, owner_id: 10 }],
    });
    findPetIdentityGroupByMember.mockReset().mockResolvedValue({
      id: 77,
      created_clinic_id: 1,
      owner_group_created_clinic_id: 1,
      owner_group_id: 42,
      version: 1,
      members: [{ clinic_id: 1, pet_id: 5 }],
    });
    vi.mocked(toast.error).mockClear();
  });

  it("canEdit=true なら飼主リンク mutateAsync を呼ぶ", async () => {
    const { result } = renderHook(() => useIdentityLinksWorkbench(true), {
      wrapper: createTestWrapper(),
    });

    act(() => {
      result.current.toggleOwner(ownerA);
      result.current.toggleOwner(ownerB);
    });

    act(() => {
      result.current.onLinkOwners();
    });

    await waitFor(() => {
      expect(createOwnerIdentityGroup).toHaveBeenCalledWith([
        { clinic_id: 1, owner_id: 10 },
        { clinic_id: 1, owner_id: 99 },
      ]);
    });
    expect(toast.error).not.toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
  });

  it("canEdit=false では飼主リンク mutateAsync を呼ばず toast.error する", async () => {
    const { result, rerender } = renderHook(
      ({ canEdit }: { canEdit: boolean }) => useIdentityLinksWorkbench(canEdit),
      { wrapper: createTestWrapper(), initialProps: { canEdit: true } },
    );

    act(() => {
      result.current.toggleOwner(ownerA);
      result.current.toggleOwner(ownerB);
    });

    rerender({ canEdit: false });

    act(() => {
      result.current.onLinkOwners();
    });

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(createOwnerIdentityGroup).not.toHaveBeenCalled();
    expect(result.current.errorMessage).toBe(PERMISSION_DENIED_MESSAGE);
  });

  it("canEdit=false では飼主 unlink の mutateAsync を呼ばず toast.error する", async () => {
    const { result, rerender } = renderHook(
      ({ canEdit }: { canEdit: boolean }) => useIdentityLinksWorkbench(canEdit),
      { wrapper: createTestWrapper(), initialProps: { canEdit: true } },
    );

    act(() => {
      result.current.toggleOwner(ownerA);
    });

    await waitFor(() => {
      expect(result.current.resolveOwnerGroupId(ownerA)).toBe(42);
    });

    rerender({ canEdit: false });

    act(() => {
      result.current.onUnlinkOwner(ownerA);
    });

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(unlinkOwnerIdentityMember).not.toHaveBeenCalled();
  });

  it("canEdit=false ではペットリンク mutateAsync を呼ばず toast.error する", async () => {
    const { result, rerender } = renderHook(
      ({ canEdit }: { canEdit: boolean }) => useIdentityLinksWorkbench(canEdit),
      { wrapper: createTestWrapper(), initialProps: { canEdit: true } },
    );

    act(() => {
      result.current.toggleOwner(ownerA);
      result.current.togglePet(petA);
      result.current.togglePet(petB);
    });

    await waitFor(() => {
      expect(result.current.canLinkPets).toBe(true);
    });

    rerender({ canEdit: false });

    act(() => {
      result.current.onLinkPets();
    });

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(createPetIdentityGroup).not.toHaveBeenCalled();
  });

  it("canEdit=false ではペット unlink の mutateAsync を呼ばず toast.error する", async () => {
    const { result, rerender } = renderHook(
      ({ canEdit }: { canEdit: boolean }) => useIdentityLinksWorkbench(canEdit),
      { wrapper: createTestWrapper(), initialProps: { canEdit: true } },
    );

    act(() => {
      result.current.togglePet(petA);
    });

    await waitFor(() => {
      expect(result.current.resolvePetGroupId(petA)).toBe(77);
    });

    rerender({ canEdit: false });

    act(() => {
      result.current.onUnlinkPet(petA);
    });

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith(PERMISSION_DENIED_MESSAGE);
    });
    expect(unlinkPetIdentityMember).not.toHaveBeenCalled();
  });
});
