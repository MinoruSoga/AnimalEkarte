import { beforeEach, describe, expect, it, vi } from "vitest";
import { axios } from "@/lib/axios";
import {
  HOSPITALIZATION_FILTER_STATUS,
  HOSPITALIZATION_LIST_DEFAULT_LIMIT,
  HOSPITALIZATION_LIST_DEFAULT_PAGE,
  toHospitalizationWireStatus,
} from "../constants";
import { getAllHospitalizations, getHospitalizations } from "./get-hospitalizations";

vi.mock("@/lib/axios", () => ({
  axios: {
    get: vi.fn(),
  },
}));

const mockedGet = vi.mocked(axios.get);

function toBackendRow(item: { id: number; status: string }) {
  return {
    id: item.id,
    clinic_id: 1,
    pet_id: 10,
    owner_id: 20,
    start_date: "2026-03-25T00:00:00Z",
    end_date: "",
    status: item.status,
    hospitalization_type: "hospitalization",
    memo: "",
    owner_request: "",
    staff_notes: "",
    created_at: "2026-03-25T00:00:00Z",
    updated_at: "2026-03-25T00:00:00Z",
  };
}

function mockPage(items: Array<{ id: number; status: string }>, total = items.length) {
  mockedGet.mockResolvedValueOnce({
    data: {
      data: items.map(toBackendRow),
      total,
      page: HOSPITALIZATION_LIST_DEFAULT_PAGE,
      limit: HOSPITALIZATION_LIST_DEFAULT_LIMIT,
    },
  });
}

describe("toHospitalizationWireStatus (BUG-009 tab → wire mapping)", () => {
  it("maps active → admitted", () => {
    expect(toHospitalizationWireStatus(HOSPITALIZATION_FILTER_STATUS.ACTIVE)).toBe("admitted");
  });
  it("maps reserved → reserved", () => {
    expect(toHospitalizationWireStatus(HOSPITALIZATION_FILTER_STATUS.RESERVED)).toBe("reserved");
  });
  it("maps discharged → discharged", () => {
    expect(toHospitalizationWireStatus(HOSPITALIZATION_FILTER_STATUS.DISCHARGED)).toBe(
      "discharged",
    );
  });
  it("maps all → undefined (no status param)", () => {
    expect(toHospitalizationWireStatus(HOSPITALIZATION_FILTER_STATUS.ALL)).toBeUndefined();
  });
});

describe("getHospitalizations request params (BUG-009)", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("sends status=admitted for active tab with default page/limit", async () => {
    mockPage([{ id: 1, status: "admitted" }], 5);
    await getHospitalizations({ statusFilter: HOSPITALIZATION_FILTER_STATUS.ACTIVE });
    expect(mockedGet).toHaveBeenCalledWith("/v1/hospitalizations", {
      params: {
        page: HOSPITALIZATION_LIST_DEFAULT_PAGE,
        limit: HOSPITALIZATION_LIST_DEFAULT_LIMIT,
        status: "admitted",
      },
    });
  });

  it("sends status=reserved for reserved tab", async () => {
    mockPage([{ id: 2, status: "reserved" }], 3);
    await getHospitalizations({ statusFilter: HOSPITALIZATION_FILTER_STATUS.RESERVED });
    expect(mockedGet).toHaveBeenCalledWith("/v1/hospitalizations", {
      params: expect.objectContaining({ status: "reserved" }),
    });
  });

  it("sends status=discharged for discharged tab", async () => {
    mockPage([{ id: 3, status: "discharged" }], 1);
    await getHospitalizations({ statusFilter: HOSPITALIZATION_FILTER_STATUS.DISCHARGED });
    expect(mockedGet).toHaveBeenCalledWith("/v1/hospitalizations", {
      params: expect.objectContaining({ status: "discharged" }),
    });
  });

  it("omits status for all tab", async () => {
    mockPage(
      [
        { id: 1, status: "admitted" },
        { id: 2, status: "reserved" },
      ],
      40,
    );
    await getHospitalizations({ statusFilter: HOSPITALIZATION_FILTER_STATUS.ALL });
    const params = mockedGet.mock.calls[0]?.[1]?.params as Record<string, unknown>;
    expect(params).toEqual({
      page: HOSPITALIZATION_LIST_DEFAULT_PAGE,
      limit: HOSPITALIZATION_LIST_DEFAULT_LIMIT,
    });
    expect(params).not.toHaveProperty("status");
  });

  it("preserves server total (page-window contract)", async () => {
    mockPage([{ id: 1, status: "reserved" }], 37);
    const result = await getHospitalizations({
      statusFilter: HOSPITALIZATION_FILTER_STATUS.RESERVED,
    });
    expect(result.total).toBe(37);
    expect(result.data).toHaveLength(1);
    expect(result.page).toBe(HOSPITALIZATION_LIST_DEFAULT_PAGE);
    expect(result.limit).toBe(HOSPITALIZATION_LIST_DEFAULT_LIMIT);
  });

  it("forwards pet_id so history pagination is scoped before the page is selected", async () => {
    mockPage([{ id: 1, status: "discharged" }], 1);
    await getHospitalizations({ petId: "42", statusFilter: HOSPITALIZATION_FILTER_STATUS.ALL });
    expect(mockedGet).toHaveBeenCalledWith("/v1/hospitalizations", {
      params: expect.objectContaining({ pet_id: "42" }),
    });
  });

  it("forwards start_date/end_date with status", async () => {
    mockPage([{ id: 1, status: "admitted" }], 1);
    await getHospitalizations({
      statusFilter: HOSPITALIZATION_FILTER_STATUS.ACTIVE,
      startDate: "2026-01-01",
      endDate: "2026-01-31",
      page: 2,
      limit: 20,
    });
    expect(mockedGet).toHaveBeenCalledWith("/v1/hospitalizations", {
      params: {
        page: 2,
        limit: 20,
        status: "admitted",
        start_date: "2026-01-01",
        end_date: "2026-01-31",
      },
    });
  });
});

describe("getAllHospitalizations (Codex PR606 #2/#10 snapshot)", () => {
  beforeEach(() => {
    mockedGet.mockReset();
    localStorage.clear();
  });

  const items = (from: number, count: number, status = "admitted") =>
    Array.from({ length: count }, (_, i) => ({ id: from + i, status }));

  function mockSnapshot(
    rows: Array<{ id: number; status: string }>,
    overrides?: Record<string, unknown>,
  ) {
    mockedGet.mockResolvedValueOnce({
      data: {
        data: rows.map(toBackendRow),
        total: rows.length,
        snapshot: true,
        ...overrides,
      },
    });
  }

  it("200 件超を snapshot=true の単一応答で全件取得する", async () => {
    mockSnapshot(items(1, 450));

    const result = await getAllHospitalizations({
      statusFilter: HOSPITALIZATION_FILTER_STATUS.ACTIVE,
    });

    expect(result).toHaveLength(450);
    expect(mockedGet).toHaveBeenCalledTimes(1);
    const params = mockedGet.mock.calls[0]?.[1]?.params as Record<string, unknown>;
    expect(params).toEqual({ snapshot: true, status: "admitted" });
    expect(params).not.toHaveProperty("page");
    expect(params).not.toHaveProperty("limit");
  });

  it("filter を forward しつつ page/limit は送らない", async () => {
    mockSnapshot(items(1, 2));

    await getAllHospitalizations({ petId: "42", startDate: "2026-01-01", endDate: "2026-01-31" });

    const params = mockedGet.mock.calls[0]?.[1]?.params as Record<string, unknown>;
    expect(params).toEqual({
      snapshot: true,
      pet_id: "42",
      start_date: "2026-01-01",
      end_date: "2026-01-31",
    });
  });

  it("X-Clinic-ID を発行時点の clinic でピン留めする", async () => {
    localStorage.setItem("auth_current_clinic:v1", "clinic-9");
    mockSnapshot(items(1, 3));

    await getAllHospitalizations({ statusFilter: HOSPITALIZATION_FILTER_STATUS.ALL });

    expect(mockedGet).toHaveBeenCalledTimes(1);
    expect(mockedGet.mock.calls[0]?.[1]?.headers).toEqual({ "X-Clinic-ID": "clinic-9" });
  });

  it("snapshot marker が無い応答は拒否する（offset fallback しない）", async () => {
    mockSnapshot(items(1, 2), { snapshot: false });

    await expect(
      getAllHospitalizations({ statusFilter: HOSPITALIZATION_FILTER_STATUS.ALL }),
    ).rejects.toThrow();
    expect(mockedGet).toHaveBeenCalledTimes(1);
  });

  it("total と行数が不一致なら不完全応答として拒否する", async () => {
    mockSnapshot(items(1, 2), { total: 5 });

    await expect(
      getAllHospitalizations({ statusFilter: HOSPITALIZATION_FILTER_STATUS.ALL }),
    ).rejects.toThrow();
  });

  it("重複 ID を含む応答は拒否する", async () => {
    mockSnapshot([
      { id: 1, status: "admitted" },
      { id: 1, status: "admitted" },
    ]);

    await expect(
      getAllHospitalizations({ statusFilter: HOSPITALIZATION_FILTER_STATUS.ALL }),
    ).rejects.toThrow();
  });

  it("行数上限を超える total は拒否する", async () => {
    mockSnapshot(items(1, 2), { total: 10001 });

    await expect(
      getAllHospitalizations({ statusFilter: HOSPITALIZATION_FILTER_STATUS.ALL }),
    ).rejects.toThrow();
  });

  it("data が配列でない応答は拒否する", async () => {
    mockedGet.mockResolvedValueOnce({
      data: { data: null, total: 0, snapshot: true },
    });

    await expect(
      getAllHospitalizations({ statusFilter: HOSPITALIZATION_FILTER_STATUS.ALL }),
    ).rejects.toThrow();
  });
});
