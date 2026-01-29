import type { Owner } from "@/types/owner";
import { transformOwner } from "./transforms";
import type { BackendOwner } from "./types";

const API_URL = "/api/v1/owners";

export const getOwners = async (): Promise<Owner[]> => {
  const response = await fetch(API_URL);
  if (!response.ok) {
    throw new Error("Failed to fetch owners");
  }
  const data: BackendOwner[] = await response.json();
  return data.map(transformOwner);
};
