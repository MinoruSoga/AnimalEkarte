import { memo } from "react";

import { RangePicker } from "./NotionDatePickerRange";
import { SinglePicker } from "./NotionDatePickerSingle";
import type { NotionDatePickerProps } from "./NotionDatePickerParts";

export const NotionDatePicker = memo(function NotionDatePicker(props: NotionDatePickerProps) {
  if (props.mode === "range") {
    return <RangePicker {...props} />;
  }
  return <SinglePicker {...props} />;
});
