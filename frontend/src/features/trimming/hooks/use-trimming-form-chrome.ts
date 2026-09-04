import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router";

import { useGetMasterItems } from "@/hooks/use-master-items";
import { useGetTrimmingCourseTypes } from "@/hooks/use-trimming-course-types";
import { useUnsavedChanges } from "@/hooks/use-unsaved-changes";
import { paths } from "@/config/paths";
import { useTrimmingHistory } from "./use-trimming-history";
import {
  decorateTrimmingCourses,
  decorateTrimmingOptions,
  TRIMMING_PRIORITY_FIELDS,
} from "../routes/trimming-form-model";
import type { TrimmingFormData } from "@/types/trimming";

export function useTrimmingFormChrome(input: {
  formData: TrimmingFormData;
  setFormData: (updates: Partial<TrimmingFormData>) => void;
  formState: { success?: boolean; timestamp?: number; fieldErrors?: Record<string, string> };
  selectedPetId: string | undefined;
  redirectPath: string;
  fromPath: string | undefined;
  handleDelete: (onSuccess: () => void) => void;
}) {
  const navigate = useNavigate();
  const { data: coursesRaw = [] } = useGetMasterItems("trimmingCourse");
  const { data: optionsRaw = [] } = useGetMasterItems("trimmingOption");
  const { data: staffItems = [] } = useGetMasterItems("staff");
  const { data: courseTypes = [] } = useGetTrimmingCourseTypes();
  const courseTypeNameById = useMemo(
    () => new Map(courseTypes.map((type) => [type.id, type.name])),
    [courseTypes],
  );
  const courses = useMemo(
    () => decorateTrimmingCourses(coursesRaw, courseTypeNameById, input.formData.courseId),
    [coursesRaw, courseTypeNameById, input.formData.courseId],
  );
  const options = useMemo(
    () => decorateTrimmingOptions(optionsRaw, input.formData.optionIds),
    [optionsRaw, input.formData.optionIds],
  );
  const { isDirty, markDirty, markClean } = useUnsavedChanges();

  useEffect(() => {
    const errorFields = Object.keys(input.formState.fieldErrors || {});
    if (errorFields.length === 0) return;
    const firstError =
      TRIMMING_PRIORITY_FIELDS.find((field) => errorFields.includes(field)) || errorFields[0];
    const element = document.getElementById(firstError);
    if (element) {
      element.focus();
      element.scrollIntoView({ behavior: "smooth", block: "center" });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- fieldErrors は timestamp と同期。timestamp のみで十分
  }, [input.formState.timestamp]);

  useEffect(() => {
    if (input.formState.success) {
      markClean();
      navigate(input.redirectPath);
    }
  }, [input.formState.success, input.formState.timestamp, navigate, markClean, input.redirectPath]);

  const [courseModalOpen, setCourseModalOpen] = useState(false);
  const [staffModalOpen, setStaffModalOpen] = useState(false);
  const [deleteConfirmOpen, setDeleteConfirmOpen] = useState(false);
  const history = useTrimmingHistory(input.selectedPetId ?? "");
  const setFormData = input.setFormData;
  const handleFormChange = useCallback(
    (updates: Partial<TrimmingFormData>) => {
      markDirty();
      setFormData(updates);
    },
    [markDirty, setFormData],
  );
  const handleDelete = input.handleDelete;
  const handleDeleteClick = useCallback(() => {
    handleDelete(() => {
      markClean();
      navigate(paths.trimming.getHref());
    });
  }, [handleDelete, markClean, navigate]);
  const handleBack = useCallback(() => {
    navigate(input.fromPath ?? paths.trimming.getHref());
  }, [input.fromPath, navigate]);
  const handleHistoryClick = useCallback(
    (updates: Partial<TrimmingFormData>) => {
      handleFormChange(history.handleHistoryClick(updates));
    },
    [history, handleFormChange],
  );
  const activeStaffItems = useMemo(
    () => staffItems.filter((staff) => staff.status === "active"),
    [staffItems],
  );

  return {
    courses,
    options,
    isDirty,
    courseModalOpen,
    setCourseModalOpen,
    staffModalOpen,
    setStaffModalOpen,
    deleteConfirmOpen,
    setDeleteConfirmOpen,
    history,
    handleFormChange,
    handleDeleteClick,
    handleBack,
    handleHistoryClick,
    activeStaffItems,
  };
}
