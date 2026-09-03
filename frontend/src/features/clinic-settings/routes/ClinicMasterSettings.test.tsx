import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { Clinic } from "../api/clinics";
import { filterClinics } from "../lib/clinic-master-settings-model";
import { ClinicMasterSettings } from "./ClinicMasterSettings";

const mocks = vi.hoisted(() => ({
  queryResult: {
    data: undefined as
      | Array<{
          id: number;
          name: string;
          phoneNumber: string;
          email: string;
          isActive: boolean;
        }>
      | undefined,
    isPending: false,
    isError: false,
  },
}));

vi.mock("@/hooks/use-auth", () => ({
  useAuth: () => ({ hasPermission: () => true }),
}));

vi.mock("@/hooks/use-permission", () => ({
  usePermission: () => ({
    canView: true,
    canCreate: true,
    canEdit: true,
    canDelete: true,
  }),
}));

vi.mock("../api/clinics", () => ({
  useGetClinics: () => mocks.queryResult,
  useCreateClinic: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateClinic: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useDeleteClinic: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock("../components/CompanyInvoiceSection", () => ({
  CompanyInvoiceSection: () => <div data-testid="company-invoice-section" />,
}));

vi.mock("@/components/shared/NavigationBlocker/NavigationBlocker", () => ({
  NavigationBlocker: () => null,
}));

describe("ClinicMasterSettings", () => {
  beforeEach(() => {
    mocks.queryResult.data = undefined;
    mocks.queryResult.isPending = false;
    mocks.queryResult.isError = false;
  });

  it("取得失敗時は空一覧ではなくエラーを表示する", () => {
    mocks.queryResult.isError = true;

    render(
      <MemoryRouter>
        <ClinicMasterSettings />
      </MemoryRouter>,
    );

    expect(screen.getByText("医院一覧の取得に失敗しました")).toBeInTheDocument();
    expect(screen.queryByText("医院が登録されていません")).not.toBeInTheDocument();
  });

  it("取得成功時は医院名を一覧表示する", () => {
    mocks.queryResult.data = [
      {
        id: 1,
        name: "八王子病院",
        phoneNumber: "042-000-0000",
        email: "hachioji@example.com",
        isActive: true,
      },
    ];

    render(
      <MemoryRouter>
        <ClinicMasterSettings />
      </MemoryRouter>,
    );

    expect(screen.getByText("八王子病院")).toBeInTheDocument();
    expect(screen.queryByText("医院一覧の取得に失敗しました")).not.toBeInTheDocument();
  });
});

describe("filterClinics", () => {
  it("院名とステータスで絞り込む", () => {
    const items = [
      {
        id: 1,
        name: "八王子病院",
        phoneNumber: "042-000-0000",
        email: "a@example.com",
        isActive: true,
      },
      { id: 2, name: "分院", phoneNumber: "", email: "branch@example.com", isActive: false },
    ] as Clinic[];

    expect(filterClinics(items, "八王子", []).map((item) => item.id)).toEqual([1]);
    expect(filterClinics(items, "branch", []).map((item) => item.id)).toEqual([2]);
    expect(
      filterClinics(items, "", [
        { key: "status", condition: "is", value: "inactive", displayValue: "無効" },
      ]).map((item) => item.id),
    ).toEqual([2]);
  });
});
