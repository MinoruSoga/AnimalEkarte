import { useState, useEffect, useCallback } from "react";
import type { ExaminationRecord } from "../api/transforms";
import { omitCorrectedExaminationFieldErrors } from "./use-examination-form-model";

export function useExaminationFormOverrides(id: string | undefined) {
  const [localOverrideScope, setLocalOverrideScope] = useState<{
    examinationID: string | undefined;
    values: Partial<ExaminationRecord>;
  }>({ examinationID: id, values: {} });
  const localOverrides = localOverrideScope.examinationID === id ? localOverrideScope.values : {};

  const [manualFieldErrors, setManualFieldErrors] = useState<Record<string, string>>({});

  // Direct hook consumers can change id without remounting. Scope overrides to
  // the active record immediately, then discard the previous record's values.
  /* eslint-disable react-hooks/set-state-in-effect -- defensive reset for non-route hook consumers */
  useEffect(() => {
    if (localOverrideScope.examinationID === id) return;
    setLocalOverrideScope((previous) =>
      previous.examinationID === id ? previous : { examinationID: id, values: {} },
    );
  }, [id, localOverrideScope.examinationID]);
  /* eslint-enable react-hooks/set-state-in-effect */

  const setFormData = useCallback(
    (next: Partial<ExaminationRecord>) => {
      setLocalOverrideScope((previous) => ({
        examinationID: id,
        values: previous.examinationID === id ? { ...previous.values, ...next } : { ...next },
      }));
      // Clear only errors for fields the user just corrected.
      setManualFieldErrors((previous) => omitCorrectedExaminationFieldErrors(previous, next));
    },
    [id],
  );

  return {
    localOverrides,
    setFormData,
    manualFieldErrors,
    setManualFieldErrors,
  };
}
