import { useCallback } from "react";

/** zipcloud 郵便番号検索 API のベース URL (FE5-6) */
const ZIPCLOUD_API_URL = "https://zipcloud.ibsnet.co.jp/api/search";

interface PostalCodeResult {
  prefecture: string;
  city: string;
  town: string;
}

interface ZipcloudResponse {
  status: number;
  message: string | null;
  results:
    | {
        address1: string;
        address2: string;
        address3: string;
        kana1: string;
        kana2: string;
        kana3: string;
        prefcode: string;
        zipcode: string;
      }[]
    | null;
}

/**
 * 郵便番号から住所を検索する hook。
 * zipcloud API（https://zipcloud.ibsnet.co.jp/）を利用。
 * 無効な郵便番号や通信エラー時は null を返す（静かに失敗）。
 */
export function usePostalCodeLookup() {
  const lookup = useCallback(async (postalCode: string): Promise<PostalCodeResult | null> => {
    const cleaned = postalCode.replace(/[-−ー\s]/g, "");
    if (cleaned.length !== 7 || !/^\d{7}$/.test(cleaned)) return null;

    try {
      const response = await fetch(`${ZIPCLOUD_API_URL}?zipcode=${cleaned}`);
      if (!response.ok) return null; // 外部 API 失敗時は住所自動入力をスキップ（既存の無音失敗方針を維持）
      const data: ZipcloudResponse = await response.json();
      const firstResult = data.results?.[0];
      if (!firstResult) return null;

      return {
        prefecture: firstResult.address1,
        city: firstResult.address2,
        town: firstResult.address3,
      };
    } catch {
      return null;
    }
  }, []);

  return { lookup };
}
