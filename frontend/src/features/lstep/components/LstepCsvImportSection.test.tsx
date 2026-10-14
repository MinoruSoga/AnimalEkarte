import { render, screen, fireEvent, act, waitFor } from "@testing-library/react";
import { uploadFriendAttributesCsv } from "../api/upload-friend-attributes-csv";
import userEvent from "@testing-library/user-event";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { describe, expect, it, vi } from "vitest";
import { LstepCsvImportSection } from "./LstepCsvImportSection";

vi.mock("../api/get-lstep-csv-imports", () => ({
  useGetLstepCsvImports: () => ({ data: [], isLoading: false }),
}));
vi.mock("../api/upload-friend-attributes-csv", () => ({ uploadFriendAttributesCsv: vi.fn() }));

vi.mock("@/lib/current-clinic", () => ({ requireStoredClinicId: () => "synthetic-clinic" }));

describe("CSV upload selection", () => {
  it("disables while pending and respects an explicit clear after upload", async () => {
    const user = userEvent.setup();
    let finish = () => {};
    vi.mocked(uploadFriendAttributesCsv).mockImplementationOnce(
      () =>
        new Promise<void>((resolve) => {
          finish = resolve;
        }),
    );
    render(
      <QueryClientProvider client={new QueryClient()}>
        <LstepCsvImportSection />
      </QueryClientProvider>,
    );
    await user.upload(
      screen.getByLabelText("友だち属性CSVファイルを選択"),
      new File(["name\nsynthetic"], "valid.csv", { type: "text/csv" }),
    );
    await user.click(screen.getByRole("button", { name: "アップロード" }));
    expect(screen.getByRole("button", { name: "保存中..." })).toBeDisabled();
    await act(async () => finish());
    await waitFor(() => expect(screen.getByText("アップロードが完了しました")).toBeInTheDocument());
    fireEvent.change(screen.getByLabelText("友だち属性CSVファイルを選択"), {
      target: { files: [] },
    });
    expect(screen.getByRole("button", { name: "アップロード" })).toBeDisabled();
    expect(screen.getByText("未選択")).toBeInTheDocument();
  });
  it("requires a nonempty actual file and disables again after clearing", async () => {
    const user = userEvent.setup();
    render(
      <QueryClientProvider client={new QueryClient()}>
        <LstepCsvImportSection />
      </QueryClientProvider>,
    );
    const input = screen.getByLabelText("友だち属性CSVファイルを選択");
    const submit = screen.getByRole("button", { name: "アップロード" });
    expect(submit).toBeDisabled();
    await user.upload(input, new File([], "empty.csv", { type: "text/csv" }));
    expect(submit).toBeDisabled();
    await user.upload(input, new File(["name\nsynthetic"], "valid.csv", { type: "text/csv" }));
    expect(submit).not.toBeDisabled();
    fireEvent.change(input, { target: { files: [] } });
    expect(submit).toBeDisabled();
    expect(screen.getByText("未選択")).toBeInTheDocument();
  });
});
