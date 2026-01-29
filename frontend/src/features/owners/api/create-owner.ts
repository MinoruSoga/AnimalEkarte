import type { Owner, CreateOwnerRequest } from "@/types/owner";
import { transformOwner } from "./transforms";
import type { BackendOwner } from "./types";

const API_URL = "/api/v1/owners";

export const createOwner = async (data: CreateOwnerRequest): Promise<Owner> => {
  const response = await fetch(API_URL, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    throw new Error("Failed to create owner");
  }
  const responseData: BackendOwner = await response.json();
  return transformOwner(responseData);
};
