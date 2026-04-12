import { axios } from "@/lib/axios";
import type { Owner, CreateOwnerRequest } from "@/types/owner";
import { transformOwner, type OwnerApiResponse } from "./transforms";

export const createOwner = async (data: CreateOwnerRequest): Promise<Owner> => {
  const { data: responseData } = await axios.post<OwnerApiResponse>("/v1/owners", data);
  return transformOwner(responseData);
};
