import { afterEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { http, HttpResponse } from "msw";
import { server } from "@/testing/mocks/node";
import { transformBackendPetToFrontend } from "@/lib/transforms/pet";
import type { Reservation } from "@/types";
import { ReservationFormModal } from "./ReservationFormModal";
import { createWrapper, silentApiHandlers, noop } from "./ReservationFormModal.test-helpers";

// Keep the real selection hook, panels, radios and Action save path. Search rendering is
// independently exercised by PatientSelectionTable.test.tsx.
vi.mock("./PatientSelectionTable", () => ({
  PatientSelectionTable: ({
    onSelect,
  }: {
    onSelect: (pet: ReturnType<typeof transformBackendPetToFrontend>) => void;
  }) => (
    <>
      {["1", "2", "3"].map((id) => (
        <button
          key={id}
          type="button"
          onClick={() =>
            onSelect(
              transformBackendPetToFrontend({
                id: Number(id),
                owner_id: id === "3" ? 20 : 10,
                name: `pet-${id}`,
                status: "alive",
              }),
            )
          }
        >
          select-{id}
        </button>
      ))}
    </>
  ),
}));

const initialData: Partial<Reservation> = {
  start: new Date(2999, 0, 1, 10),
  end: new Date(2999, 0, 1, 11),
  type: "1",
  visitType: "first",
  notes: "既存の予約内容",
};

afterEach(() => server.resetHandlers());

function setup(edit = false) {
  server.use(
    ...silentApiHandlers,
    http.get("/api/v1/reservations/available-times", () => HttpResponse.json([])),
    http.get("/api/v1/reservations/123", () => HttpResponse.json({ id: 123, visit_type: "first" })),
  );
  const onSave = vi.fn();
  render(
    <ReservationFormModal
      isOpen
      onClose={noop}
      onSave={onSave}
      initialData={edit ? { ...initialData, id: "123" } : initialData}
      canCreate
      canEdit
    />,
    { wrapper: createWrapper() },
  );
  return onSave;
}

function expectVisit(type: "first" | "revisit", locked: boolean) {
  expect(screen.getByRole("radio", { name: type === "first" ? "初診" : "再診" })).toBeChecked();
  const other = screen.getByRole("radio", { name: type === "first" ? "再診" : "初診" });
  if (locked) expect(other).toBeDisabled();
  else expect(other).toBeEnabled();
}

describe("EMR-274 visit-type input assistance", () => {
  it("locks existing selections to revisit across multiple pets and owner changes, with matching save value", async () => {
    const onSave = setup();
    fireEvent.change(screen.getByRole("textbox", { name: "予約内容" }), {
      target: { value: "変更した予約内容" },
    });
    fireEvent.click(screen.getByRole("button", { name: "select-1" }));
    expectVisit("revisit", true);
    fireEvent.click(screen.getByRole("button", { name: "select-2" }));
    expectVisit("revisit", true);
    fireEvent.click(screen.getByRole("button", { name: "予約を確定" }));
    await waitFor(() => expect(onSave).toHaveBeenCalledOnce());
    expect(onSave.mock.calls[0][0]).toMatchObject({
      visitType: "revisit",
      notes: "変更した予約内容",
    });
    expect(onSave.mock.calls[0][1]).toHaveLength(2);
    fireEvent.click(screen.getByRole("button", { name: "select-3" }));
    expectVisit("revisit", true);
    fireEvent.click(screen.getByRole("button", { name: "予約を確定" }));
    await waitFor(() => expect(onSave).toHaveBeenCalledTimes(2));
    expect(onSave.mock.calls[1][1]).toHaveLength(1);
  });

  it("new-owner mode locks first even after a prior revisit selection and switches back consistently", () => {
    setup();
    fireEvent.click(screen.getByRole("button", { name: "select-1" }));
    expectVisit("revisit", true);
    fireEvent.click(screen.getByTestId("mode-new"));
    expectVisit("first", true);
    fireEvent.click(screen.getByRole("radio", { name: "再診" }));
    expectVisit("first", true);
    fireEvent.click(screen.getByTestId("mode-existing"));
    expectVisit("revisit", true);
    fireEvent.click(screen.getByRole("button", { name: "select-1" }));
    expect(screen.getByRole("radio", { name: "初診" })).toBeEnabled();
    expect(screen.getByRole("radio", { name: "再診" })).toBeEnabled();
  });

  it("editing preserves first and allows changing to revisit with the matching save value", async () => {
    const onSave = setup(true);
    fireEvent.click(screen.getByRole("button", { name: "select-1" }));
    expectVisit("first", false);
    fireEvent.click(screen.getByRole("radio", { name: "再診" }));
    expectVisit("revisit", false);
    fireEvent.click(screen.getByRole("button", { name: "更新する" }));
    await waitFor(() => expect(onSave).toHaveBeenCalledOnce());
    expect(onSave.mock.calls[0][0]).toMatchObject({ visitType: "revisit" });
  });

  it("labels the existing notes as reservation content without changing its value", () => {
    setup();
    expect(screen.getByText("予約内容")).toBeInTheDocument();
    expect(screen.queryByText("メモ")).not.toBeInTheDocument();
    expect(screen.getByDisplayValue("既存の予約内容")).toBeInTheDocument();
  });
});
