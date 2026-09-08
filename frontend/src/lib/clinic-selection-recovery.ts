import { toast } from "sonner";
import { setStoredClinicId } from "@/lib/current-clinic";
import { queryClient } from "@/lib/react-query";

export const CLINIC_SELECTION_UNAVAILABLE = "clinic_selection_unavailable";

export type ClinicSelectionBlockReason = "none" | "no-clinic" | "recovery-failed";

let writesPaused = false;
let recoveryPromise: Promise<void> | null = null;
let blockReason: ClinicSelectionBlockReason = "none";
const listeners = new Set<() => void>();

function notifyBlockListeners(): void {
  for (const listener of listeners) {
    listener();
  }
}

function setBlockReason(next: ClinicSelectionBlockReason): void {
  if (blockReason === next) {
    return;
  }
  blockReason = next;
  notifyBlockListeners();
}

export function areClinicWritesPaused(): boolean {
  return writesPaused;
}

export function pauseClinicWrites(): void {
  writesPaused = true;
}

export function getClinicSelectionBlockReason(): ClinicSelectionBlockReason {
  return blockReason;
}

export function subscribeClinicSelectionBlock(listener: () => void): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

export function clearClinicSelectionRecovery(): void {
  writesPaused = false;
  recoveryPromise = null;
  setBlockReason("none");
}

export function resetClinicSelectionRecoveryForTests(): void {
  clearClinicSelectionRecovery();
}

export function isClinicSelectionUnavailable(data: unknown): boolean {
  if (data === null || typeof data !== "object") {
    return false;
  }
  return "error_code" in data && data.error_code === CLINIC_SELECTION_UNAVAILABLE;
}

export function recoverClinicSelectionOnce(): Promise<void> {
  if (recoveryPromise !== null) {
    return recoveryPromise;
  }
  pauseClinicWrites();
  recoveryPromise = recoverClinicSelection().finally(() => {
    recoveryPromise = null;
  });
  return recoveryPromise;
}

async function recoverClinicSelection(): Promise<void> {
  const apiBase = import.meta.env.VITE_API_URL || "/api";
  try {
    // Omit X-Clinic-ID so Auth can use the live main clinic instead of the stale JWT default.
    const response = await fetch(`${apiBase}/v1/me`, {
      method: "GET",
      credentials: "include",
      headers: {
        Accept: "application/json",
        "X-Requested-With": "XMLHttpRequest",
      },
    });
    if (!response.ok) {
      setBlockReason("recovery-failed");
      return;
    }
    const data: {
      main_clinic_id?: string;
      clinics?: Array<{ clinic_id?: string; clinic_name?: string; is_main?: boolean }>;
    } = await response.json();
    const clinics = Array.isArray(data.clinics) ? data.clinics : [];
    if (clinics.length === 0) {
      setBlockReason("no-clinic");
      return;
    }
    const nextId =
      (typeof data.main_clinic_id === "string" && data.main_clinic_id) ||
      clinics.find((clinic) => clinic.is_main)?.clinic_id ||
      clinics[0]?.clinic_id;
    if (typeof nextId !== "string" || nextId === "") {
      setBlockReason("recovery-failed");
      return;
    }
    if (!setStoredClinicId(nextId)) {
      setBlockReason("recovery-failed");
      return;
    }
    const clinicName = clinics.find((clinic) => clinic.clinic_id === nextId)?.clinic_name ?? "";
    toast.warning(
      clinicName === ""
        ? "選択していた医院は利用できなくなったため、別の医院に切り替えます。"
        : `選択していた医院は利用できなくなったため、${clinicName}に切り替えました`,
    );
    queryClient.clear();
    window.location.reload();
  } catch {
    setBlockReason("recovery-failed");
  }
}
