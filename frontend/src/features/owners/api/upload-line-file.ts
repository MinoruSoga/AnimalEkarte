import { axios } from "@/lib/axios";

interface UploadedLineFile {
  id: number;
  file_name: string;
  file_type: string;
}

export async function uploadLineFile(
  clinicId: string,
  ownerId: string,
  file: File
): Promise<UploadedLineFile> {
  const fd = new FormData();
  fd.append("file", file);
  fd.append("purpose", "other");
  fd.append("owner_id", ownerId);

  const { data } = await axios.post<UploadedLineFile>(
    "/api/v1/shared-files",
    fd
  );
  return data;
}
