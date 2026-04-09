import { useState, useCallback } from 'react';
import type { ReservationFlow, CustomerInfo } from '../types/models';

const initialFlow: ReservationFlow = {
  customerInfo: {
    name: '',
    phone: '',
    ownerName: '',
    pets: [],
  },
  courseId: null,
  courseName: '',
  staffId: 0,
  staffName: '',
  date: '',
  startTime: '',
  endTime: '',
  requestText: '',
};

interface UseReservationFlowReturn {
  flow: ReservationFlow;
  setCustomerInfo: (info: CustomerInfo) => void;
  setCourse: (id: number, name: string) => void;
  setStaff: (id: number, name: string) => void;
  setDate: (date: string) => void;
  setTime: (startTime: string, endTime: string) => void;
  setRequestText: (text: string) => void;
  resetFlow: () => void;
}

export function useReservationFlow(): UseReservationFlowReturn {
  const [flow, setFlow] = useState<ReservationFlow>(() => initialFlow);

  const setCustomerInfo = useCallback((info: CustomerInfo) => {
    setFlow(prev => ({ ...prev, customerInfo: info }));
  }, []);

  const setCourse = useCallback((id: number, name: string) => {
    setFlow(prev => ({ ...prev, courseId: id, courseName: name }));
  }, []);

  const setStaff = useCallback((id: number, name: string) => {
    setFlow(prev => ({ ...prev, staffId: id, staffName: name }));
  }, []);

  const setDate = useCallback((date: string) => {
    setFlow(prev => ({ ...prev, date }));
  }, []);

  const setTime = useCallback((startTime: string, endTime: string) => {
    setFlow(prev => ({ ...prev, startTime, endTime }));
  }, []);

  const setRequestText = useCallback((text: string) => {
    setFlow(prev => ({ ...prev, requestText: text }));
  }, []);

  const resetFlow = useCallback(() => {
    setFlow(initialFlow);
  }, []);

  return {
    flow,
    setCustomerInfo,
    setCourse,
    setStaff,
    setDate,
    setTime,
    setRequestText,
    resetFlow,
  };
}
