// Cross-feature composition: accounting + master (company)
// AccountingDetail に company のインボイス番号を注入する

import { AccountingDetail } from "@/features/accounting";
import { useGetCompany } from "@/features/master";

export function AccountingDetailPage() {
  const { data: company } = useGetCompany();

  return (
    <AccountingDetail
      invoiceRegistrationNumber={company?.invoiceRegistrationNumber}
    />
  );
}
