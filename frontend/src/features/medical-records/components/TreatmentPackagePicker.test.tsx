import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";
import { axios } from "@/lib/axios";
import { TreatmentPackagePicker } from "./TreatmentPackagePicker";
import { TREATMENT_PACKAGE_RUNTIME_AVAILABLE } from "../lib/treatment-package-operation-state";

vi.mock("@/lib/axios", () => ({ axios: { get: vi.fn(), post: vi.fn() } }));

const wrapper = ({ children }: { children: ReactNode }) => (
  <QueryClientProvider client={new QueryClient()}>{children}</QueryClientProvider>
);

describe("TreatmentPackagePicker（runtime 未提供の既定 gate）", () => {
  it("既定は無効で「実行系は未提供」と表示し、API を呼ばない", () => {
    expect(TREATMENT_PACKAGE_RUNTIME_AVAILABLE).toBe(false);
    render(
      <TreatmentPackagePicker
        scope={{ recordId: "91" }}
        editContext={null}
        canCreate={true}
        onClose={vi.fn()}
      />,
      { wrapper },
    );
    expect(screen.getByRole("note").textContent).toContain("実行系は未提供");
    expect(axios.get).not.toHaveBeenCalled();
    expect(axios.post).not.toHaveBeenCalled();
  });
});
