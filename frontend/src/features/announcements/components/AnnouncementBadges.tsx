import { Badge } from '../../../components/ui/Badge'
import type { AnnouncementStatus, AnnouncementVisibility } from '../types'
import { formatAnnouncementStatus, formatAnnouncementVisibility } from '../types'

export function AnnouncementVisibilityBadge({
  visibility,
}: {
  visibility: AnnouncementVisibility
}) {
  return (
    <Badge tone={visibility === 'PUBLIC' ? 'info' : 'warning'} title={visibility}>
      {formatAnnouncementVisibility(visibility)}
      <span className="sr-only"> ({visibility})</span>
    </Badge>
  )
}

export function AnnouncementStatusBadge({ status }: { status: AnnouncementStatus }) {
  return (
    <Badge tone={status === 'PUBLISHED' ? 'success' : 'neutral'} title={status}>
      {formatAnnouncementStatus(status)}
      <span className="sr-only"> ({status})</span>
    </Badge>
  )
}
