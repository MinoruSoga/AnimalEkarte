import { act, renderHook } from "@testing-library/react";
import { beforeEach, expect, it, vi } from "vitest";
import { useCourseImageDrafts } from "./use-course-image-drafts";
const discard = vi.hoisted(() => vi.fn());
vi.mock("../api/upload-trimming-course-image", () => ({ discardTrimmingCourseImage: discard }));
beforeEach(() => {
  discard.mockReset();
  discard.mockResolvedValue(undefined);
});
it("discards only proven local drafts and protects submitted keys even after unknown save", () => {
  const { result, unmount } = renderHook(useCourseImageDrafts);
  act(() => {
    result.current.register("draft");
    result.current.register("sent");
    result.current.protectSubmitted("sent");
    result.current.discard("original");
    result.current.discard("sent");
  });
  expect(discard).not.toHaveBeenCalled();
  unmount();
  expect(discard).toHaveBeenCalledExactlyOnceWith("draft");
});
it("cleans a late upload after leaving without updating the departed form", () => {
  const { result, unmount } = renderHook(useCourseImageDrafts);
  const register = result.current.register;
  unmount();
  expect(register("late")).toBe(false);
  expect(discard).toHaveBeenCalledExactlyOnceWith("late");
});
