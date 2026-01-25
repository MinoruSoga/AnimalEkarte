import { Owner, CreateOwnerRequest, UpdateOwnerRequest } from "@/types";

const API_BASE_URL = "/api/v1";

export const getOwners = async (): Promise<Owner[]> => {
  const response = await fetch(`${API_BASE_URL}/owners`);
  if (!response.ok) {
    throw new Error("Failed to fetch owners");
  }
  return response.json();
};

export const getOwner = async (id: string): Promise<Owner> => {
  const response = await fetch(`${API_BASE_URL}/owners/${id}`);
  if (!response.ok) {
    throw new Error("Failed to fetch owner");
  }
  return response.json();
};

export const createOwner = async (data: CreateOwnerRequest): Promise<Owner> => {
  const response = await fetch(`${API_BASE_URL}/owners`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    throw new Error("Failed to create owner");
  }
  return response.json();
};

export const updateOwner = async (id: string, data: UpdateOwnerRequest): Promise<Owner> => {
  const response = await fetch(`${API_BASE_URL}/owners/${id}`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(data),
  });
  if (!response.ok) {
    throw new Error("Failed to update owner");
  }
  return response.json();
};

export const deleteOwner = async (id: string): Promise<void> => {
  const response = await fetch(`${API_BASE_URL}/owners/${id}`, {
    method: "DELETE",
  });
  if (!response.ok) {
    throw new Error("Failed to delete owner");
  }
};
