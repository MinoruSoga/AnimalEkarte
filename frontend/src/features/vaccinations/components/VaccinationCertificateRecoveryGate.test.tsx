import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { createTestWrapper } from "@/testing/TestUtils";
import type { CertificateApi } from "../api/certificate-api";
import { createMockCertificateApi, type MockCertificateControls } from "../mock/certificate-mock-api";
import { VaccinationCertificateScreen } from "./VaccinationCertificateScreen";

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({ canView: true, canCreate: true, canEdit: true, canDelete: false }),
}));

/**
 * EMR-487 Slice A: fence が無い間（CERTIFICATE_RECOVERY_FENCE_AVAILABLE の実際の既定 = false。ここでは mock しない）は、
 * 発見が 0 件・complete・成功のみでも新しい発行・再印刷を API へ送らない。プレビューは使える。
 */
function renderScreen(api: CertificateApi) {
  return render(<VaccinationCertificateScreen vaccinationId={101} api={api} />, { wrapper: createTestWrapper() });
}

function callsOf(controls: MockCertificateControls, operation: string) {
  return controls.calls.filter((c) => c.operation === operation);
}

async function seedNormalEdition(api: CertificateApi) {
  const request = {
    kind: "normal" as const,
    layout: "portrait" as const,
    inputs: {},
    preconditions: { source_version: 7, latest_edition_id: null },
  };
  const preview = await api.previewForIssue(101, request);
  await api.issue(101, {
    ...request,
    operation_key: "99999999-9999-4999-8999-999999999999",
    preview_fingerprint: preview.preview_fingerprint,
    preview_display_issue_date: preview.snapshot.display_issue_date,
  });
}

const PAUSED = /回復の安全確認（fence）が利用可能になるまで/;

beforeEach(() => {
  URL.createObjectURL = vi.fn(() => "blob:mock");
  URL.revokeObjectURL = vi.fn();
});

describe("fence が無い間は新しい発行・再印刷を始めない（既定）", () => {
  it("発見が空でも、プレビューは作れるが発行は無効で、押しても API を呼ばない", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    renderScreen(api);
    await user.click(await screen.findByRole("button", { name: "プレビューを作成" }));
    await screen.findByTitle("証明書の確認プレビュー");
    await waitFor(() => expect(callsOf(controls, "queryOperationRecovery").length).toBeGreaterThan(0));
    expect(await screen.findByText(PAUSED)).toBeInTheDocument();
    const issue = screen.getByRole("button", { name: "この内容で発行する" });
    expect(issue).toBeDisabled();
    await user.click(issue);
    expect(callsOf(controls, "issue")).toHaveLength(0);
  });

  it("成功済みの操作だけが見つかった場合も、新しい再印刷・発行は始められない", async () => {
    const user = userEvent.setup();
    const { api, controls } = createMockCertificateApi({ rabiesEligible: false });
    await seedNormalEdition(api);
    renderScreen(api);
    const row = (await screen.findByText("第1版")).closest("li") as HTMLElement;
    await within(row).findByText(PAUSED);
    const reprint = within(row).getByRole("button", { name: "再印刷の準備" });
    expect(reprint).toBeDisabled();
    await user.click(reprint);
    expect(callsOf(controls, "reprint")).toHaveLength(0);
    await user.click(await screen.findByRole("button", { name: "プレビューを作成" }));
    await screen.findByTitle("証明書の確認プレビュー");
    expect(screen.getByRole("button", { name: "この内容で発行する" })).toBeDisabled();
    expect(callsOf(controls, "issue")).toHaveLength(1); // seed の 1 回のみ
  });
});
