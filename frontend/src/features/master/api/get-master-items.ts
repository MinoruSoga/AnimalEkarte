// Re-export from shared hooks layer.
// This file is kept for backward-compatibility within the master feature.
// All implementations live in @/hooks/use-master-items.
export {
  MASTER_CATEGORY_ENDPOINT,
  transformGenericMasterItem,
  useGetMasterItemsByCategory,
} from "@/hooks/use-master-items";
export type { GenericMasterBackendItem } from "@/hooks/use-master-items";
