import { axios } from "@/lib/axios";

export interface TrimmingImageUploadResult {
  key: string;
  signed_url: string;
}

function parseTrimmingImageUpload(data: unknown): TrimmingImageUploadResult {
  if (typeof data !== "object" || data === null) {
    throw new Error("invalid trimming image upload response");
  }
  const record = data as Record<string, unknown>;
  const key = record.key;
  const signedUrl = record.signed_url;
  if (typeof key !== "string" || key === "" || typeof signedUrl !== "string") {
    throw new Error("invalid trimming image upload response");
  }
  return { key, signed_url: signedUrl };
}

/** POST /v1/trimmings/images/upload。永続化するのは key だけで、signed_url は表示用。 */
export async function uploadTrimmingImage(file: File): Promise<TrimmingImageUploadResult> {
  const body = new FormData();
  body.append("file", file);
  // Content-Type は axios が FormData 用に boundary 付きで設定するため手動指定しない
  const { data } = await axios.post<unknown>("/v1/trimmings/images/upload", body);
  return parseTrimmingImageUpload(data);
}
