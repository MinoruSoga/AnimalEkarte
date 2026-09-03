import { axios } from "@/lib/axios";

// POST /v1/clinics/:clinic_id/lstep/csv-imports/friend-attributes
export async function uploadFriendAttributesCsv(clinicId: string, file: File): Promise<void> {
  const fd = new FormData();
  fd.append("file", file);
  await axios.post(`/v1/clinics/${clinicId}/lstep/csv-imports/friend-attributes`, fd, {
    headers: { "Content-Type": "multipart/form-data" },
  });
}
