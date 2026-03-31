import { useCallback } from "react";

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
  const lookup = useCallback(
    async (postalCode: string): Promise<PostalCodeResult | null> => {
      const cleaned = postalCode.replace(/[-−ー\s]/g, "");
      if (cleaned.length !== 7 || !/^\d{7}$/.test(cleaned)) return null;

      try {
        const response = await fetch(
          `https://zipcloud.ibsnet.co.jp/api/search?zipcode=${cleaned}`,
        );
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
    },
    [],
  );

  return { lookup };
}
