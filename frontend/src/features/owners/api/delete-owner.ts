const API_URL = "/api/v1/owners";

export const deleteOwner = async (id: string): Promise<void> => {
  const response = await fetch(`${API_URL}/${id}`, {
    method: "DELETE",
  });
  if (!response.ok) {
    throw new Error("Failed to delete owner");
  }
};
