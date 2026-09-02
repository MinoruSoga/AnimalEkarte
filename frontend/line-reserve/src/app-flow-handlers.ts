import { useCallback, type Dispatch, type SetStateAction } from 'react';

import type { PageType } from './types/models';

export function useReservationAppHandlers(
  goTo: (page: PageType) => void,
  resetFlow: () => void,
  setNotice: Dispatch<SetStateAction<string | null>>,
  setCompletedReservationId: Dispatch<SetStateAction<number>>,
  setCompletedNotes: Dispatch<SetStateAction<string>>,
) {
  const handleNewReservation = useCallback(() => {
    setNotice(null);
    resetFlow();
    goTo('step1');
  }, [resetFlow, goTo, setNotice]);

  const handleConfirm = useCallback((reservationId: number, notes: string) => {
    setNotice(null);
    setCompletedReservationId(reservationId);
    setCompletedNotes(notes);
    goTo('step8');
  }, [goTo, setNotice, setCompletedReservationId, setCompletedNotes]);

  const handleSlotTaken = useCallback((message: string, redirectStep: number) => {
    setNotice(message);
    const stepMap: Record<number, PageType> = {
      1: 'step1', 2: 'step2', 3: 'step3', 4: 'step4',
      5: 'step5', 6: 'step6', 7: 'step7',
    };
    goTo(stepMap[redirectStep] ?? 'step4');
  }, [goTo, setNotice]);

  const handleCompleteToMyReservations = useCallback(() => {
    goTo('my-reservations');
  }, [goTo]);

  const handleCompleteNewReservation = useCallback(() => {
    setNotice(null);
    resetFlow();
    goTo('step1');
  }, [resetFlow, goTo, setNotice]);

  return {
    handleNewReservation,
    handleConfirm,
    handleSlotTaken,
    handleCompleteToMyReservations,
    handleCompleteNewReservation,
  };
}
