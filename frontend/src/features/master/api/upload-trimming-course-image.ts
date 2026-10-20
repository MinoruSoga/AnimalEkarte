import { axios } from "@/lib/axios";

export interface TrimmingCourseImageUpload {
  key: string;
  signedUrl: string;
}

/**
 * EMR-457: master 専用経路は所有台帳の予約・upload 完了を記録してから key を返す。
 * 永続化するのは key だけで、signedUrl は表示専用。
 * feature 間 import 禁止のため trimming feature の同等関数をここで呼び直す。
 */
export async function uploadTrimmingCourseImage(file: File): Promise<TrimmingCourseImageUpload> {
  const body = new FormData();
  body.append("file", file);
  const { data } = await axios.post<unknown>("/v1/masters/trimming-courses/images/upload", body);
  if (typeof data !== "object" || data === null) {
    throw new Error("invalid trimming image upload response");
  }
  const record = data as Record<string, unknown>;
  if (
    typeof record.key !== "string" ||
    record.key === "" ||
    typeof record.signed_url !== "string"
  ) {
    throw new Error("invalid trimming image upload response");
  }
  return { key: record.key, signedUrl: record.signed_url };
}

/**
 * 未保存のまま置換・削除したコース画像の best-effort 掃除（DELETE /v1/masters/trimming-courses/images）。
 * ローカル候補を未保存・未送信の作業コピーに限定し、サーバーでも医院・所有者・完了・未送信・参照を再検査する。
 * 保存済み・送信済み・不明keyは呼ばない。
 */
export async function discardTrimmingCourseImage(key: string): Promise<void> {
  await axios.delete("/v1/masters/trimming-courses/images", { data: { key } });
}
