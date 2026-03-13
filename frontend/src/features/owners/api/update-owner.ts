import { axios } from "@/lib/axios";
import type { Owner, UpdateOwnerRequest } from "@/types/owner";
import { transformOwner } from "./transforms";
import type { Owner as BackendOwner } from "@/types/generated/models";

export const updateOwner = async (
  id: string,
  data: UpdateOwnerRequest
): Promise<Owner> => {
  const { data: responseData } = await axios.patch<BackendOwner>(`/v1/owners/${id}`, data);
  return transformOwner(responseData);
};
