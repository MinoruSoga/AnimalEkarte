export interface AddCheckupFormState {
  checkup_type_id: string;
  date: string;
  next_date: string;
  doctor_id: string;
  result: string;
}

const EMPTY_ADD_FORM: AddCheckupFormState = {
  checkup_type_id: "",
  date: "",
  next_date: "",
  doctor_id: "",
  result: "",
};

function todayISODate(): string {
  return new Date().toISOString().slice(0, 10);
}

export function makeDefaultCheckupAddForm(): AddCheckupFormState {
  return { ...EMPTY_ADD_FORM, date: todayISODate() };
}
