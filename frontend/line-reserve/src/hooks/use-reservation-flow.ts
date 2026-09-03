import { useState, useCallback } from "react";
import type { ReservationFlow, CustomerInfo } from "../types/models";

const initialFlow: ReservationFlow = {
  customerInfo: {
    name: "",
    phone: "",
    ownerName: "",
    pets: [],
  },
  courseId: null,
  courseName: "",
  courseCategory: "general",
  staffId: 0,
  staffName: "",
  date: "",
  startTime: "",
  endTime: "",
  requestText: "",
  trimmingCourseId: null,
  trimmingCourseName: "",
  trimmingOptionIds: [],
};

interface UseReservationFlowReturn {
  flow: ReservationFlow;
  setCustomerInfo: (info: CustomerInfo) => void;
  setCourse: (id: number, name: string, category?: "general" | "trimming") => void;
  setStaff: (id: number, name: string) => void;
  setDate: (date: string) => void;
  setTime: (startTime: string, endTime: string) => void;
  setRequestText: (text: string) => void;
  setTrimmingCourse: (id: number, name: string) => void;
  setTrimmingOptions: (ids: number[]) => void;
  resetFlow: () => void;
}

export function useReservationFlow(): UseReservationFlowReturn {
  const [flow, setFlow] = useState<ReservationFlow>(() => initialFlow);

  const setCustomerInfo = useCallback((info: CustomerInfo) => {
    setFlow((prev) => ({ ...prev, customerInfo: info }));
  }, []);

  const setCourse = useCallback(
    (id: number, name: string, category: "general" | "trimming" = "general") => {
      setFlow((prev) => ({ ...prev, courseId: id, courseName: name, courseCategory: category }));
    },
    [],
  );

  const setStaff = useCallback((id: number, name: string) => {
    setFlow((prev) => ({ ...prev, staffId: id, staffName: name }));
  }, []);

  const setDate = useCallback((date: string) => {
    setFlow((prev) => ({ ...prev, date }));
  }, []);

  const setTime = useCallback((startTime: string, endTime: string) => {
    setFlow((prev) => ({ ...prev, startTime, endTime }));
  }, []);

  const setRequestText = useCallback((text: string) => {
    setFlow((prev) => ({ ...prev, requestText: text }));
  }, []);

  const setTrimmingCourse = useCallback((id: number, name: string) => {
    setFlow((prev) => ({ ...prev, trimmingCourseId: id, trimmingCourseName: name }));
  }, []);

  const setTrimmingOptions = useCallback((ids: number[]) => {
    setFlow((prev) => ({ ...prev, trimmingOptionIds: ids }));
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
    setTrimmingCourse,
    setTrimmingOptions,
    resetFlow,
  };
}
