import { describe, it, expect, vi, beforeEach } from "vitest";

import { axios } from "@/lib/axios";

import { uploadTrimmingImage } from "./upload-trimming-image";

vi.mock("@/lib/axios", () => ({
  axios: {
    post: vi.fn(),
  },
}));

const STYLE_KEY = "trimming-images/1/style.jpg";
const SIGNED_URL = "https://signed.example/style.jpg";

describe("uploadTrimmingImage", () => {
  beforeEach(() => {
    vi.mocked(axios.post).mockReset();
  });

  it("file フィールドでアップロードし、key と signed_url を返す", async () => {
    vi.mocked(axios.post).mockResolvedValue({
      data: { key: STYLE_KEY, signed_url: SIGNED_URL },
    });
    const file = new File(["img"], "style.jpg", { type: "image/jpeg" });

    const result = await uploadTrimmingImage(file);

    // Content-Type は axios が FormData 用に boundary 付きで設定するため、手動指定は渡さない
    expect(axios.post).toHaveBeenCalledWith("/v1/trimmings/images/upload", expect.any(FormData));
    expect(vi.mocked(axios.post).mock.calls[0]?.[2]).toBeUndefined();
    const body = vi.mocked(axios.post).mock.calls[0]?.[1] as FormData;
    expect(body.get("file")).toBe(file);
    expect(result).toEqual({ key: STYLE_KEY, signed_url: SIGNED_URL });
  });

  it("key が無い応答は保存できないので拒否する", async () => {
    vi.mocked(axios.post).mockResolvedValue({
      data: { signed_url: SIGNED_URL },
    });

    await expect(uploadTrimmingImage(new File(["img"], "style.jpg"))).rejects.toThrow(
      /invalid trimming image upload response/,
    );
  });
});
