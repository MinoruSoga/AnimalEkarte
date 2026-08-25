import { beforeEach, describe, expect, it, vi } from "vitest";
import { AxiosError, AxiosHeaders, type InternalAxiosRequestConfig } from "axios";

vi.mock("@/lib/axios", () => ({
  axios: {
    get: vi.fn(),
  },
}));

import { axios } from "@/lib/axios";
import {
  findOwnerIdentityGroupByMember,
  findPetIdentityGroupByMember,
} from "./identity-links-api";
import type {
  OwnerGroupResponse,
  PetGroupResponse,
} from "@/types/generated/identitylink-responses";

const mockedGet = vi.mocked(axios.get);

function axiosError(status: number): AxiosError {
  const config = { headers: new AxiosHeaders() } as InternalAxiosRequestConfig;
  return new AxiosError(
    "request failed",
    AxiosError.ERR_BAD_RESPONSE,
    config,
    undefined,
    {
      config,
      data: { error: "error" },
      headers: new AxiosHeaders(),
      status,
      statusText: "Error",
    },
  );
}

describe("findOwnerIdentityGroupByMember", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("returns OwnerGroupResponse on success", async () => {
    const group: OwnerGroupResponse = {
      id: 42,
      created_clinic_id: 1,
      version: 1,
      members: [{ clinic_id: 1, owner_id: 10 }],
    };
    mockedGet.mockResolvedValueOnce({ data: group });

    await expect(findOwnerIdentityGroupByMember(1, 10)).resolves.toEqual(group);
    expect(mockedGet).toHaveBeenCalledWith("/v1/identity-links/owners/1/10/group");
  });

  it("returns null on 404", async () => {
    mockedGet.mockRejectedValueOnce(axiosError(404));
    await expect(findOwnerIdentityGroupByMember(1, 99)).resolves.toBeNull();
  });

  it("rethrows non-404 errors", async () => {
    mockedGet.mockRejectedValueOnce(axiosError(403));
    await expect(findOwnerIdentityGroupByMember(1, 10)).rejects.toMatchObject({
      response: { status: 403 },
    });
  });
});

describe("findPetIdentityGroupByMember", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("returns PetGroupResponse on success", async () => {
    const group: PetGroupResponse = {
      id: 77,
      created_clinic_id: 1,
      owner_group_created_clinic_id: 1,
      owner_group_id: 42,
      version: 1,
      members: [{ clinic_id: 1, pet_id: 5 }],
    };
    mockedGet.mockResolvedValueOnce({ data: group });

    await expect(findPetIdentityGroupByMember(1, 5)).resolves.toEqual(group);
    expect(mockedGet).toHaveBeenCalledWith("/v1/identity-links/pets/1/5/group");
  });

  it("returns null on 404", async () => {
    mockedGet.mockRejectedValueOnce(axiosError(404));
    await expect(findPetIdentityGroupByMember(1, 99)).resolves.toBeNull();
  });

  it("rethrows non-404 errors", async () => {
    mockedGet.mockRejectedValueOnce(axiosError(500));
    await expect(findPetIdentityGroupByMember(1, 5)).rejects.toMatchObject({
      response: { status: 500 },
    });
  });
});
