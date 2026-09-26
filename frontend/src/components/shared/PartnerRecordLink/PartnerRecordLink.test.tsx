import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { PartnerRecordLink } from "./PartnerRecordLink";

const mocks = vi.hoisted(() => ({
  navigate: vi.fn(),
  canView: true,
  canCreate: true,
  target: null as {
    mode: "open" | "create";
    href: string;
    appointmentId?: string;
  } | null,
  usePartnerRecordLink: vi.fn(),
}));

vi.mock("react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("react-router")>();
  return {
    ...actual,
    useNavigate: () => mocks.navigate,
    useLocation: () => ({ pathname: "/trimming/new", search: "?petId=10" }),
  };
});

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canView: mocks.canView,
    canCreate: mocks.canCreate,
    canEdit: false,
    canDelete: false,
  }),
}));

vi.mock("@/hooks/use-partner-record-link", () => ({
  usePartnerRecordLink: mocks.usePartnerRecordLink,
}));

const BASE_PROPS = {
  petId: "10",
  visitDate: "2026-09-26",
} as const;

beforeEach(() => {
  mocks.navigate.mockReset();
  mocks.canView = true;
  mocks.canCreate = true;
  mocks.target = null;
  mocks.usePartnerRecordLink.mockReset();
  mocks.usePartnerRecordLink.mockImplementation(() => ({
    isLoading: false,
    target: mocks.target,
  }));
});

describe("PartnerRecordLink — 相方あり（開く）", () => {
  it("診察カルテが既存なら「診察カルテを開く」を表示し詳細へ遷移する", async () => {
    const user = userEvent.setup();
    mocks.target = { mode: "open", href: "/medical-records/42" };

    render(<PartnerRecordLink kind="medical-record" {...BASE_PROPS} />);

    await user.click(screen.getByRole("button", { name: "診察カルテを開く" }));

    expect(mocks.navigate).toHaveBeenCalledWith("/medical-records/42", {
      state: {
        from: "/trimming/new?petId=10",
        appointmentId: undefined,
        visitDate: "2026-09-26",
      },
    });
  });

  it("trimming appointment が既存なら「トリミング記録を開く」で appointmentId 付きフォームへ遷移する", async () => {
    const user = userEvent.setup();
    mocks.target = {
      mode: "open",
      href: "/trimming/new?visitDate=2026-09-26&petId=10&appointmentId=55",
      appointmentId: "55",
    };

    render(<PartnerRecordLink kind="trimming" {...BASE_PROPS} />);

    await user.click(screen.getByRole("button", { name: "トリミング記録を開く" }));

    expect(mocks.navigate).toHaveBeenCalledWith(
      "/trimming/new?visitDate=2026-09-26&petId=10&appointmentId=55",
      {
        state: {
          from: "/trimming/new?petId=10",
          appointmentId: "55",
          visitDate: "2026-09-26",
        },
      },
    );
  });
});

describe("PartnerRecordLink — 相方なし（作成）", () => {
  it("カルテ未存在かつ create 権限ありなら「診察カルテを作成」を表示し new 経路へ遷移する", async () => {
    const user = userEvent.setup();
    mocks.target = {
      mode: "create",
      href: "/medical-records/new?visitDate=2026-09-26&petId=10",
    };

    render(<PartnerRecordLink kind="medical-record" {...BASE_PROPS} />);

    await user.click(screen.getByRole("button", { name: "診察カルテを作成" }));

    expect(mocks.navigate).toHaveBeenCalledWith(
      "/medical-records/new?visitDate=2026-09-26&petId=10",
      expect.objectContaining({
        state: expect.objectContaining({ visitDate: "2026-09-26" }),
      }),
    );
  });

  it("トリミング未存在かつ create 権限ありなら「トリミング記録を作成」を表示する", () => {
    mocks.target = {
      mode: "create",
      href: "/trimming/new?visitDate=2026-09-26&petId=10",
    };

    render(<PartnerRecordLink kind="trimming" {...BASE_PROPS} />);

    expect(screen.getByRole("button", { name: "トリミング記録を作成" })).toBeInTheDocument();
  });
});

describe("PartnerRecordLink — 権限による表示制御", () => {
  it("遷移先リソースの権限が一切なければ描画しない", () => {
    mocks.canView = false;
    mocks.canCreate = false;
    mocks.target = { mode: "open", href: "/medical-records/42" };

    const { container } = render(<PartnerRecordLink kind="medical-record" {...BASE_PROPS} />);

    expect(container).toBeEmptyDOMElement();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it("open 時に view 権限がなければ描画しない（create 権限のみでも不可）", () => {
    mocks.canView = false;
    mocks.canCreate = true;
    mocks.target = { mode: "open", href: "/medical-records/42" };

    const { container } = render(<PartnerRecordLink kind="medical-record" {...BASE_PROPS} />);

    expect(container).toBeEmptyDOMElement();
  });

  it("create 時に create 権限がなければ描画しない（view 権限のみでも不可）", () => {
    mocks.canView = true;
    mocks.canCreate = false;
    mocks.target = { mode: "create", href: "/trimming/new?visitDate=2026-09-26&petId=10" };

    const { container } = render(<PartnerRecordLink kind="trimming" {...BASE_PROPS} />);

    expect(container).toBeEmptyDOMElement();
  });
});

describe("PartnerRecordLink — 非表示条件", () => {
  it("死亡ペットでは描画せず相方解決 query も無効化する", () => {
    mocks.target = { mode: "open", href: "/medical-records/42" };

    const { container } = render(
      <PartnerRecordLink kind="medical-record" {...BASE_PROPS} isPetDeceased />,
    );

    expect(container).toBeEmptyDOMElement();
    expect(mocks.usePartnerRecordLink).toHaveBeenCalledWith(
      expect.objectContaining({ enabled: false }),
    );
  });

  it("相方解決中（target=null）では描画しない", () => {
    mocks.target = null;

    const { container } = render(<PartnerRecordLink kind="trimming" {...BASE_PROPS} />);

    expect(container).toBeEmptyDOMElement();
  });
});
