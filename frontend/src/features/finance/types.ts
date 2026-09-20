import type { PaginationMeta } from '../../lib/api/types'

export type TransactionType = 'INCOME' | 'EXPENSE'

export type FinanceTransaction = {
  id: string
  type: TransactionType
  title: string
  amount: number
  category: string
  note: string
  created_at: string
  updated_at: string
}

export type FinanceSummary = {
  total_income: number
  total_expense: number
  balance: number
}

export type FinanceFilters = {
  page: number
  page_size: number
  search: string
  type: TransactionType | ''
  category: string
  from: string
  to: string
}

export type FinanceSummaryFilters = {
  from: string
  to: string
}

export type FinanceTransactionListResponse = {
  items: FinanceTransaction[]
  meta: PaginationMeta
}

export type CreateFinanceTransactionRequest = {
  type: TransactionType
  title: string
  amount: number
  category: string
  note: string
}

export type UpdateFinanceTransactionRequest = {
  type?: TransactionType
  title?: string
  amount?: number
  category?: string
  note?: string
}

export const TRANSACTION_TYPE_OPTIONS: Array<{ value: TransactionType; label: string }> = [
  { value: 'INCOME', label: 'Pemasukan' },
  { value: 'EXPENSE', label: 'Pengeluaran' },
]

export function formatTransactionType(type: TransactionType): string {
  return TRANSACTION_TYPE_OPTIONS.find((item) => item.value === type)?.label ?? type
}
