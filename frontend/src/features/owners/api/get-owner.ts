import type { Owner } from "@/types/owner";
import { transformOwner } from "./transforms";
import type { BackendOwner } from "./types";

const API_URL = "/api/v1/owners";

export const getOwner = async (id: string): Promise<Owner> => {
  const response = await fetch(`${API_URL}/${id}`);
  if (!response.ok) {
    throw new Error("Failed to fetch owner");
  }
  const data: BackendOwner = await response.json();
  return transformOwner(data);
};
