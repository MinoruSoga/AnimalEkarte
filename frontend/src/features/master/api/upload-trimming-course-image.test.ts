import { beforeEach, expect, it, vi } from "vitest";
import { axios } from "@/lib/axios";
import {
  discardTrimmingCourseImage,
  uploadTrimmingCourseImage,
} from "./upload-trimming-course-image";
vi.mock("@/lib/axios", () => ({ axios: { post: vi.fn(), delete: vi.fn() } }));
beforeEach(() => vi.clearAllMocks());
it("uses the provenance-backed master upload route and keeps the returned key", async () => {
  const key = "trimming-images/7/123_0123456789abcdef0123456789abcdef.png";
  vi.mocked(axios.post).mockResolvedValue({
    data: { key, signed_url: "https://signed.example/image" },
  });
  const file = new File(["image"], "image.png", { type: "image/png" });
  expect(await uploadTrimmingCourseImage(file)).toEqual({
    key,
    signedUrl: "https://signed.example/image",
  });
  expect(axios.post).toHaveBeenCalledWith(
    "/v1/masters/trimming-courses/images/upload",
    expect.any(FormData),
  );
});
it("uses only the master draft discard route", async () => {
  vi.mocked(axios.delete).mockResolvedValue({});
  await discardTrimmingCourseImage("draft-key");
  expect(axios.delete).toHaveBeenCalledExactlyOnceWith("/v1/masters/trimming-courses/images", {
    data: { key: "draft-key" },
  });
});
it("rejects an incomplete upload response before it can become a local draft", async () => {
  vi.mocked(axios.post).mockResolvedValue({ data: { signed_url: "https://signed.example/image" } });
  await expect(uploadTrimmingCourseImage(new File(["image"], "image.png"))).rejects.toThrow(
    "invalid trimming image upload response",
  );
});
