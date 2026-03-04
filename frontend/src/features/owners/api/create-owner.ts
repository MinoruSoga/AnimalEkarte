import { axios } from "@/lib/axios";
import type { Owner, CreateOwnerRequest } from "@/types/owner";
import { transformOwner } from "./transforms";
import type { BackendOwner } from "./types";

export const createOwner = async (data: CreateOwnerRequest): Promise<Owner> => {
  const { data: responseData } = await axios.post<BackendOwner>("/v1/owners", data);
  return transformOwner(responseData);
};
