import { todayJSTISO } from "@/lib/jst-date";

export interface AddCheckupFormState {
  checkup_type_id: string;
  date: string;
  next_schedule_type: string;
  next_date: string;
  doctor_id: string;
  result: string;
}

const EMPTY_ADD_FORM: AddCheckupFormState = {
  checkup_type_id: "",
  date: "",
  next_schedule_type: "1year",
  next_date: "",
  doctor_id: "",
  result: "",
};

function todayISODate(): string {
  return todayJSTISO();
}

export function makeDefaultCheckupAddForm(): AddCheckupFormState {
  return { ...EMPTY_ADD_FORM, date: todayISODate() };
}
