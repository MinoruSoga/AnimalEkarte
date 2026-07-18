import { StatusBadge } from '@/components/shared/StatusBadge/StatusBadge';
import { getEstimateStatusColor } from '@/lib/status-helpers';
import type { EstimateStatus } from '../../types';

const STATUS_LABELS: Record<EstimateStatus, string> = {
  draft: '下書き',
  sent: '送付済み',
  approved: '承認済み',
  rejected: '却下',
};

interface EstimateStatusBadgeProps {
  status: EstimateStatus;
}

export function EstimateStatusBadge({ status }: EstimateStatusBadgeProps) {
  return (
    <StatusBadge colorClass={getEstimateStatusColor(status)}>
      {STATUS_LABELS[status] ?? status}
    </StatusBadge>
  );
}
