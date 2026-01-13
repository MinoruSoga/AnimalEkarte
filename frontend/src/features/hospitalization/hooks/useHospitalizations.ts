import { useMemo, useState, useCallback, useEffect } from "react";
import { toast } from "sonner";
import { Hospitalization } from "../../../types";
import { getHospitalizations, updateHospitalization } from "../api";
import { HospitalizationFilterStatus, HOSPITALIZATION_STATUS } from "../constants";

export function useHospitalizations(searchTerm: string, statusFilter: HospitalizationFilterStatus = "active") {
  const [hospitalizations, setHospitalizations] = useState<Hospitalization[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  const loadData = useCallback(async () => {
    setIsLoading(true);
    try {
        const data = await getHospitalizations();
        setHospitalizations(data);
    } catch (e) {
        console.error("Failed to load hospitalizations", e);
        toast.error("データの取得に失敗しました");
    } finally {
        setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const filteredHospitalizations = useMemo(() => {
    let result = hospitalizations;

    // Status Filter
    if (statusFilter !== "all") {
      result = result.filter((h) => {
        if (statusFilter === "active") return h.status === HOSPITALIZATION_STATUS.ACTIVE || h.status === HOSPITALIZATION_STATUS.TEMP_DISCHARGE;
        if (statusFilter === "discharged") return h.status === HOSPITALIZATION_STATUS.DISCHARGED;
        if (statusFilter === "reserved") return h.status === HOSPITALIZATION_STATUS.RESERVED;
        return true;
      });
    }

    // Search Filter
    if (searchTerm) {
      const lowerTerm = searchTerm.toLowerCase();
      result = result.filter(
        (h) =>
          h.ownerName.toLowerCase().includes(lowerTerm) ||
          h.petName.toLowerCase().includes(lowerTerm) ||
          h.hospitalizationNo.toLowerCase().includes(lowerTerm)
      );
    }
    
    return result;
  }, [hospitalizations, searchTerm, statusFilter]);

  const handleUpdateHospitalization = useCallback(async (id: string, updates: Partial<Hospitalization>) => {
    // Optimistic Update
    const previous = [...hospitalizations];
    setHospitalizations((prev) => 
        prev.map((h) => (h.id === id ? { ...h, ...updates } : h))
    );

    try {
        const updated = await updateHospitalization(id, updates);
        // Ensure state matches server response
        setHospitalizations((prev) => 
            prev.map((h) => (h.id === id ? updated : h))
        );
        toast.success("更新しました");
    } catch (e) {
        console.error("Failed to update hospitalization", e);
        toast.error("更新に失敗しました");
        setHospitalizations(previous);
    }
  }, [hospitalizations]);

  const movePet = useCallback(async (hospitalizationId: string, targetCageId: string) => {
    const sourceHosp = hospitalizations.find(h => h.id === hospitalizationId);
    if (!sourceHosp) return;

    const previousHospitalizations = [...hospitalizations];

    // Check if target cage is occupied by another ACTIVE hospitalization
    const targetHosp = hospitalizations.find(
      h => h.cageId === targetCageId && h.status === HOSPITALIZATION_STATUS.ACTIVE && h.id !== hospitalizationId
    );

    // 1. Optimistic Update
    setHospitalizations(prev => {
        const next = [...prev];
        const sourceIndex = next.findIndex(h => h.id === hospitalizationId);
        
        if (targetHosp) {
            const targetIndex = next.findIndex(h => h.id === targetHosp.id);
            // Swap cageIds locally
            if (sourceIndex !== -1 && targetIndex !== -1) {
                next[sourceIndex] = { ...next[sourceIndex], cageId: targetCageId };
                next[targetIndex] = { ...next[targetIndex], cageId: sourceHosp.cageId || "" };
            }
        } else {
            // Move locally
            if (sourceIndex !== -1) {
                next[sourceIndex] = { ...next[sourceIndex], cageId: targetCageId };
            }
        }
        return next;
    });

    try {
        if (targetHosp) {
            // Swap: Update both sequentially
            await updateHospitalization(sourceHosp.id, { cageId: targetCageId });
            await updateHospitalization(targetHosp.id, { cageId: sourceHosp.cageId || "" });
        } else {
            // Move to empty
            await updateHospitalization(sourceHosp.id, { cageId: targetCageId });
        }
        
        // Optionally fetch fresh data to be absolutely sure
        // await loadData(); 
    } catch (e) {
        console.error("Failed to move pet", e);
        toast.error("部屋の移動に失敗しました");
        // Rollback
        setHospitalizations(previousHospitalizations);
    }
  }, [hospitalizations]);

  return { 
      data: filteredHospitalizations, 
      updateHospitalization: handleUpdateHospitalization, 
      movePet,
      isLoading 
  };
}
