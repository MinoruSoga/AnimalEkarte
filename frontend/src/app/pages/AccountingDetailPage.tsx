// Cross-feature composition: accounting + company
// AccountingDetail に company のインボイス番号を注入する

import { AccountingDetail } from "@/features/accounting";
import { useGetCompany } from "@/hooks/use-company";

export function AccountingDetailPage() {
  const { data: company } = useGetCompany();

  return <AccountingDetail invoiceRegistrationNumber={company?.invoiceRegistrationNumber} />;
}
