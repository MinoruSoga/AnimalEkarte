import { useState, useEffect, useLayoutEffect, useCallback, useRef, useMemo } from "react";
import type { ExamResult } from "../api/transforms";
import type { ExamItemRow } from "../components/ExamItemsTable";
import { useGetExamTypeFields } from "../api/get-exam-type-fields";
import { buildRowsFromTemplate, mapExamResultsToFormRows } from "./use-examination-form-model";

function useExaminationFormItemEffects(input: {
  id: string | undefined;
  isEdit: boolean;
  existingItems: ExamResult[] | undefined;
  existingItemsQuerySucceeded: boolean;
  currentTestTypeId: string;
  formItemsRef: { current: ExamItemRow[] };
  itemsReadyForIDRef: { current: string | undefined };
  setFormItems: (rows: ExamItemRow[]) => void;
  setFormItemsOwnerID: (id: string | undefined) => void;
}) {
  const {
    id,
    isEdit,
    existingItems,
    existingItemsQuerySucceeded,
    currentTestTypeId,
    formItemsRef,
    itemsReadyForIDRef,
    setFormItems,
    setFormItemsOwnerID,
  } = input;
  const { data: examTypeFields } = useGetExamTypeFields(currentTestTypeId);
  const itemsInitializedForIDRef = useRef<string | undefined>(undefined);
  const formItemsExamIDRef = useRef(id);
  const emptyItemsAwaitingTemplateForIDRef = useRef<string | undefined>(undefined);

  useLayoutEffect(() => {
    if (
      isEdit &&
      !existingItemsQuerySucceeded &&
      itemsReadyForIDRef.current === id
    ) {
      itemsReadyForIDRef.current = undefined;
    }
  }, [existingItemsQuerySucceeded, id, isEdit, itemsReadyForIDRef]);

  useEffect(() => {
    if (!isEdit || formItemsExamIDRef.current === id) return;
    formItemsExamIDRef.current = id;
    itemsInitializedForIDRef.current = undefined;
    itemsReadyForIDRef.current = undefined;
    emptyItemsAwaitingTemplateForIDRef.current = undefined;
    formItemsRef.current = [];
    setFormItemsOwnerID(id);
    setFormItems([]);
  }, [id, isEdit, formItemsRef, itemsReadyForIDRef, setFormItems, setFormItemsOwnerID]);

  useEffect(() => {
    if (!isEdit) return;
    if (!existingItemsQuerySucceeded) return;
    if (itemsInitializedForIDRef.current === id) {
      itemsReadyForIDRef.current = id;
      return;
    }
    if (!existingItems) return;
    if (existingItems.length > 0) {
      const rows = mapExamResultsToFormRows(existingItems);
      formItemsRef.current = rows;
      setFormItemsOwnerID(id);
      setFormItems(rows);
    } else {
      // 成功した空応答は権威ある初期状態。テンプレが遅れて届いた場合だけ、
      // ユーザーが行を追加していなければ後続 effect で補完する。
      emptyItemsAwaitingTemplateForIDRef.current =
        examTypeFields === undefined ? id : undefined;
      if (
        formItemsRef.current.length === 0 &&
        examTypeFields &&
        examTypeFields.length > 0
      ) {
        const rows = buildRowsFromTemplate(examTypeFields);
        formItemsRef.current = rows;
        setFormItemsOwnerID(id);
        setFormItems(rows);
      }
    }
    formItemsExamIDRef.current = id;
    itemsInitializedForIDRef.current = id;
    itemsReadyForIDRef.current = id;
  }, [id, isEdit, existingItems, existingItemsQuerySucceeded, examTypeFields, formItemsRef, itemsReadyForIDRef, setFormItems, setFormItemsOwnerID]);

  useEffect(() => {
    if (emptyItemsAwaitingTemplateForIDRef.current !== id) return;
    if (examTypeFields === undefined) return;
    emptyItemsAwaitingTemplateForIDRef.current = undefined;
    if (formItemsRef.current.length === 0 && examTypeFields.length > 0) {
      const rows = buildRowsFromTemplate(examTypeFields);
      formItemsRef.current = rows;
      setFormItemsOwnerID(id);
      setFormItems(rows);
    }
  }, [examTypeFields, id, formItemsRef, setFormItems, setFormItemsOwnerID]);

  const previousTestTypeRef = useRef<{
    examinationID: string | undefined;
    testTypeID: string | undefined;
  }>({ examinationID: id, testTypeID: undefined });
  useEffect(() => {
    const next = currentTestTypeId;
    if (previousTestTypeRef.current.examinationID !== id) {
      previousTestTypeRef.current = {
        examinationID: id,
        testTypeID: next || undefined,
      };
      return;
    }
    if (!next) return;
    if (previousTestTypeRef.current.testTypeID === undefined) {
      previousTestTypeRef.current = { examinationID: id, testTypeID: next };
      return;
    }
    if (previousTestTypeRef.current.testTypeID === next) return;
    previousTestTypeRef.current = { examinationID: id, testTypeID: next };
    setFormItemsOwnerID(id);
    if (examTypeFields) {
      setFormItems(buildRowsFromTemplate(examTypeFields));
    } else {
      setFormItems([]);
    }
  }, [currentTestTypeId, examTypeFields, id, setFormItems, setFormItemsOwnerID]);

  const newModeInitializedRef = useRef(false);
  useEffect(() => {
    if (isEdit) return;
    if (newModeInitializedRef.current) return;
    if (!currentTestTypeId) return;
    if (!examTypeFields) return;
    setFormItemsOwnerID(id);
    setFormItems(buildRowsFromTemplate(examTypeFields));
    newModeInitializedRef.current = true;
  }, [isEdit, currentTestTypeId, examTypeFields, id, setFormItems, setFormItemsOwnerID]);
}

export function useExaminationFormItems(input: {
  id: string | undefined;
  isEdit: boolean;
  existingItems: ExamResult[] | undefined;
  existingItemsQuerySucceeded: boolean;
  currentTestTypeId: string;
}) {
  const { id, isEdit, existingItems, existingItemsQuerySucceeded, currentTestTypeId } =
    input;
  const [formItems, setFormItems] = useState<ExamItemRow[]>([]);
  const [formItemsOwnerID, setFormItemsOwnerID] = useState(id);
  const visibleFormItems = useMemo(
    () => (isEdit && formItemsOwnerID !== id ? [] : formItems),
    [formItems, formItemsOwnerID, id, isEdit],
  );
  const formItemsRef = useRef(visibleFormItems);
  useLayoutEffect(() => {
    formItemsRef.current = visibleFormItems;
  }, [visibleFormItems]);

  const itemsReadyForIDRef = useRef<string | undefined>(undefined);
  useExaminationFormItemEffects({
    id,
    isEdit,
    existingItems,
    existingItemsQuerySucceeded,
    currentTestTypeId,
    formItemsRef,
    itemsReadyForIDRef,
    setFormItems,
    setFormItemsOwnerID,
  });

  const setInspectionValue = useCallback((key: string, value: string) => {
    setFormItems((prev) =>
      prev.map((row) =>
        row.key === key ? { ...row, inspectionValue: value } : row,
      ),
    );
  }, []);

  const manualItemSequenceRef = useRef(0);
  const addManualItem = useCallback(() => {
    manualItemSequenceRef.current += 1;
    const key = `manual-${manualItemSequenceRef.current}`;
    setFormItems((previous) => {
      const nextSortOrder =
        previous.reduce(
          (maximum, row) => Math.max(maximum, row.sortOrder),
          -1,
        ) + 1;
      return [
        ...previous,
        {
          key,
          name: "",
          inspectionValue: "",
          unit: "",
          normalValue: "",
          referenceValue: "",
          sortOrder: nextSortOrder,
        },
      ];
    });
  }, []);

  const removeItem = useCallback((key: string) => {
    setFormItems((previous) => previous.filter((row) => row.key !== key));
  }, []);

  const setItemName = useCallback((key: string, value: string) => {
    setFormItems((previous) =>
      previous.map((row) => (row.key === key ? { ...row, name: value } : row)),
    );
  }, []);

  return {
    visibleFormItems,
    formItemsRef,
    itemsReadyForIDRef,
    setInspectionValue,
    addManualItem,
    removeItem,
    setItemName,
  };
}
