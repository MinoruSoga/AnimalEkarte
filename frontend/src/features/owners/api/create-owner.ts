import { axios } from "@/lib/axios";
import type { Owner, CreateOwnerRequest } from "@/types/owner";
import { transformOwner } from "./transforms";
import type { Owner as BackendOwner } from "@/types/generated/models";

export const createOwner = async (data: CreateOwnerRequest): Promise<Owner> => {
  const { data: responseData } = await axios.post<BackendOwner>("/v1/owners", data);
  return transformOwner(responseData);
};
