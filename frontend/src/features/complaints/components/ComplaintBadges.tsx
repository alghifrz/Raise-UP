import { Badge } from '../../../components/ui/Badge'
import type { ComplaintStatus, ComplaintUrgency } from '../types'
import { formatComplaintStatus, formatComplaintUrgency } from '../types'

export function ComplaintStatusBadge({ status }: { status: ComplaintStatus }) {
  const tone =
    status === 'BARU'
      ? 'info'
      : status === 'DIPROSES'
        ? 'warning'
        : status === 'SELESAI'
          ? 'success'
          : 'danger'

  return (
    <Badge tone={tone} title={status}>
      {formatComplaintStatus(status)}
      <span className="sr-only"> ({status})</span>
    </Badge>
  )
}

export function ComplaintUrgencyBadge({ urgency }: { urgency: ComplaintUrgency }) {
  const tone = urgency === 'PRIORITY' ? 'danger' : urgency === 'MEDIUM' ? 'warning' : 'neutral'

  return (
    <Badge tone={tone} title={urgency}>
      {formatComplaintUrgency(urgency)}
      <span className="sr-only"> ({urgency})</span>
    </Badge>
  )
}
