import { memo } from "react";

import { RangePicker } from "./DatePickerRange";
import { SinglePicker } from "./DatePickerSingle";
import type { DatePickerProps } from "./DatePickerParts";

export const DatePicker = memo(function DatePicker(props: DatePickerProps) {
  if (props.mode === "range") {
    return <RangePicker {...props} />;
  }
  return <SinglePicker {...props} />;
});
