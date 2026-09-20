import { Badge } from '../../../components/ui/Badge'
import type { DuesPaymentStatusValue } from '../types'
import { formatDuesPaymentStatus } from '../types'

export function DuesPaymentStatusBadge({ status }: { status: DuesPaymentStatusValue }) {
  return (
    <Badge tone={status === 'PAID' ? 'success' : 'warning'} title={status}>
      {formatDuesPaymentStatus(status)}
      <span className="sr-only"> ({status})</span>
    </Badge>
  )
}
