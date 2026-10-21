import { useCallback, type Dispatch, type SetStateAction } from "react";

import type { PageType } from "./types/models";
import type { RevisitIntakeReservationSummary } from "./pages/RevisitIntakePage";
import { LIFF_REVISIT_INTAKE_ACTIVE } from "./lib/revisit-intake-activation";

/** EMR-474: 再診問診画面へ渡す state（予約 ID は URL へ載せない）。 */
export interface RevisitIntakeRoute {
  appointmentId: string;
  returnTo: PageType;
  summary?: RevisitIntakeReservationSummary;
}

export function useReservationAppHandlers(
  goTo: (page: PageType) => void,
  resetFlow: () => void,
  setNotice: Dispatch<SetStateAction<string | null>>,
  setCompletedReservationId: Dispatch<SetStateAction<string | null>>,
  setCompletedNotes: Dispatch<SetStateAction<string>>,
  setRevisitIntake?: Dispatch<SetStateAction<RevisitIntakeRoute | null>>,
  setCreateResultUnknown?: Dispatch<SetStateAction<boolean>>,
) {
  const handleNewReservation = useCallback(() => {
    setNotice(null);
    setCreateResultUnknown?.(false);
    resetFlow();
    goTo("step1");
  }, [resetFlow, goTo, setNotice, setCreateResultUnknown]);

  const handleConfirm = useCallback(
    (reservationId: string, notes: string) => {
      setNotice(null);
      setCreateResultUnknown?.(false);
      setCompletedReservationId(reservationId);
      setCompletedNotes(notes);
      goTo("step8");
    },
    [goTo, setNotice, setCompletedReservationId, setCompletedNotes, setCreateResultUnknown],
  );

  // EMR-519: 201 契約違反を報告され、画面遷移後も「結果不明」を保持する。
  const handleResultUnknown = useCallback(() => {
    setCreateResultUnknown?.(true);
  }, [setCreateResultUnknown]);

  const handleSlotTaken = useCallback(
    (message: string, redirectStep: number) => {
      setNotice(message);
      const stepMap: Record<number, PageType> = {
        1: "step1",
        2: "step2",
        3: "step3",
        4: "step4",
        5: "step5",
        6: "step6",
        7: "step7",
      };
      goTo(stepMap[redirectStep] ?? "step4");
    },
    [goTo, setNotice],
  );

  const handleCompleteToMyReservations = useCallback(() => {
    setCreateResultUnknown?.(false);
    goTo("my-reservations");
  }, [goTo, setCreateResultUnknown]);

  const handleCompleteNewReservation = useCallback(() => {
    setNotice(null);
    setCreateResultUnknown?.(false);
    resetFlow();
    goTo("step1");
  }, [resetFlow, goTo, setNotice, setCreateResultUnknown]);

  const handleOpenRevisitIntake = useCallback(
    (route: RevisitIntakeRoute) => {
      // EMR-474: 利用開始前は問診ページへ遷移しない。
      if (!LIFF_REVISIT_INTAKE_ACTIVE) return;
      setNotice(null);
      setRevisitIntake?.(route);
      goTo("revisit-intake");
    },
    [goTo, setNotice, setRevisitIntake],
  );

  return {
    handleOpenRevisitIntake,
    handleNewReservation,
    handleConfirm,
    handleResultUnknown,
    handleSlotTaken,
    handleCompleteToMyReservations,
    handleCompleteNewReservation,
  };
}
