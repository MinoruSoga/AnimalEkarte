import { axios } from "@/lib/axios";

/**
 * DELETE /v1/trimmings/images。フォーム放棄・置換・stale 解決でレコードへ
 * 永続化されなかったアップロード済みオブジェクトの best-effort 掃除。
 * 呼び出し側は失敗を致命的に扱わないこと。
 */
export async function deleteTrimmingImage(key: string): Promise<void> {
  await axios.delete("/v1/trimmings/images", { data: { key } });
}
