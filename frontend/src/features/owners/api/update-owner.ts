import type { Owner, UpdateOwnerRequest } from "@/types/owner";
import { transformOwner } from "./transforms";
import type { BackendOwner } from "./types";

const API_URL = "/api/v1/owners";

export const updateOwner = async (
  id: string,
  data: UpdateOwnerRequest
): Promise<Owner> => {
  const response = await fetch(`${API_URL}/${id}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    throw new Error("Failed to update owner");
  }
  const responseData: BackendOwner = await response.json();
  return transformOwner(responseData);
};
