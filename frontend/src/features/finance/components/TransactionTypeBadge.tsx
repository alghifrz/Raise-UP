import { Badge } from '../../../components/ui/Badge'
import type { TransactionType } from '../types'
import { formatTransactionType } from '../types'

export function TransactionTypeBadge({ type }: { type: TransactionType }) {
  return (
    <Badge tone={type === 'INCOME' ? 'success' : 'danger'} title={type}>
      {formatTransactionType(type)}
      <span className="sr-only"> ({type})</span>
    </Badge>
  )
}
