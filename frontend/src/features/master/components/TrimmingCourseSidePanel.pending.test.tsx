import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import { createTestWrapper } from "@/testing/TestUtils";
import { TrimmingCourseSidePanel } from "./TrimmingCourseSidePanel";
vi.mock("@/hooks/use-current-clinic-name", () => ({ useCurrentClinicName: () => "clinic" }));
vi.mock("../api/trimming-course-type", () => ({ useGetTrimmingCourseTypes: () => ({ data: [] }) }));

it("disables Save during upload and forbids image operations during an unresolved save", async () => {
  let finishUpload: ((value: { key: string; signedUrl: string }) => void) | undefined;
  let finishSave: ((value: boolean) => void) | undefined;
  const upload = await import("../api/upload-trimming-course-image");
  vi.spyOn(upload, "uploadTrimmingCourseImage").mockImplementation(
    () =>
      new Promise((resolve) => {
        finishUpload = resolve;
      }),
  );
  const discard = vi.spyOn(upload, "discardTrimmingCourseImage").mockResolvedValue();
  const onSave = vi.fn(
    () =>
      new Promise<boolean>((resolve) => {
        finishSave = resolve;
      }),
  );
  const dirty = vi.fn();
  const { unmount } = render(
    <TrimmingCourseSidePanel item={null} onSave={onSave} onClose={vi.fn()} onDirtyChange={dirty} />,
    { wrapper: createTestWrapper() },
  );
  fireEvent.change(document.getElementById("master-title")!, { target: { value: "course" } });
  fireEvent.change(document.querySelector('input[type="file"]')!, {
    target: { files: [new File(["x"], "a.jpg", { type: "image/jpeg" })] },
  });
  await waitFor(() => expect(screen.getByRole("button", { name: "保存" })).toBeDisabled());
  expect(dirty).toHaveBeenLastCalledWith(true);
  await act(async () =>
    finishUpload?.({ key: "trimming-images/1/draft.jpg", signedUrl: "https://signed/draft" }),
  );
  fireEvent.click(screen.getByRole("button", { name: "保存" }));
  await waitFor(() => expect(onSave).toHaveBeenCalledTimes(1));
  expect(screen.queryByRole("button", { name: "画像を削除" })).not.toBeInTheDocument();
  unmount();
  expect(discard).not.toHaveBeenCalled();
  await act(async () => finishSave?.(false));
});
