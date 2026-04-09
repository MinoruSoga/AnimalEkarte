/**
 * app/pages/LineReservationCustomersPage.tsx
 * Cross-feature合成: line-reservation + owners
 * owners feature を props 注入し、dependency inversion を実現する。
 */
import { LineReservationCustomers } from "@/features/line-reservation";
import { useGetOwners } from "@/features/owners";

export function LineReservationCustomersPage() {
  const { data: owners = [] } = useGetOwners();
  return <LineReservationCustomers owners={owners} />;
}
