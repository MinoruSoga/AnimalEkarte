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

function mockPage(items: Array<{ id: number; status: string }>, total = items.length) {
  mockedGet.mockResolvedValueOnce({
    data: {
      data: items.map((item) => ({
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
      })),
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

describe("getAllHospitalizations (Codex PR606 #2/#10)", () => {
  beforeEach(() => {
    mockedGet.mockReset();
    localStorage.clear();
  });

  const items = (from: number, count: number, status = "admitted") =>
    Array.from({ length: count }, (_, i) => ({ id: from + i, status }));

  it("total を超える行を 200/page の複数ページで全件取得する", async () => {
    mockPage(items(1, 200), 250);
    mockPage(items(201, 50), 250);

    const result = await getAllHospitalizations({
      statusFilter: HOSPITALIZATION_FILTER_STATUS.ACTIVE,
    });

    expect(result).toHaveLength(250);
    expect(mockedGet).toHaveBeenCalledTimes(2);
    expect(mockedGet.mock.calls[0]?.[1]?.params).toMatchObject({ page: 1, limit: 200 });
    expect(mockedGet.mock.calls[1]?.[1]?.params).toMatchObject({ page: 2, limit: 200 });
  });

  it("total が 1 ページに収まるなら 1 回だけ発行する", async () => {
    mockPage(items(1, 3), 3);
    const result = await getAllHospitalizations({
      statusFilter: HOSPITALIZATION_FILTER_STATUS.ALL,
    });
    expect(result).toHaveLength(3);
    expect(mockedGet).toHaveBeenCalledTimes(1);
  });

  it("空ページが返れば total 未到達でも早期打切りする", async () => {
    mockPage(items(1, 200), 999);
    mockPage([], 999);

    const result = await getAllHospitalizations({
      statusFilter: HOSPITALIZATION_FILTER_STATUS.ALL,
    });

    expect(result).toHaveLength(200);
    expect(mockedGet).toHaveBeenCalledTimes(2);
  });

  it("全ページに X-Clinic-ID をピン留めする（発行中のクリニック切替混入防止）", async () => {
    localStorage.setItem("auth_current_clinic:v1", "clinic-9");
    mockPage(items(1, 200), 201);
    mockPage(items(201, 1), 201);

    await getAllHospitalizations({ statusFilter: HOSPITALIZATION_FILTER_STATUS.ALL });

    for (const call of mockedGet.mock.calls) {
      expect(call[1]?.headers).toEqual({ "X-Clinic-ID": "clinic-9" });
    }
  });
});
