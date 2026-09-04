import { axios } from "@/lib/axios";
import type { Owner } from "@/types/owner";
import { transformOwner, type OwnerApiResponse } from "./transforms";

export const getOwner = async (id: string): Promise<Owner> => {
  const { data } = await axios.get<OwnerApiResponse>(`/v1/owners/${id}`);
  return transformOwner(data);
};
