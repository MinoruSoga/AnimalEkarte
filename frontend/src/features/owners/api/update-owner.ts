import { axios } from "@/lib/axios";
import type { Owner, UpdateOwnerRequest } from "@/types/owner";
import { transformOwner, type OwnerApiResponse } from "./transforms";

export const updateOwner = async (id: string, data: UpdateOwnerRequest): Promise<Owner> => {
  const { data: responseData } = await axios.patch<OwnerApiResponse>(`/v1/owners/${id}`, data);
  return transformOwner(responseData);
};
