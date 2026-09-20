import { useState, type FormEvent } from 'react'
import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import { Select } from '../../../components/ui/Select'
import { Textarea } from '../../../components/ui/Textarea'
import { parsePositiveIntegerAmount } from '../../../lib/utils'
import { toFinanceErrorMessage } from '../errors'
import type {
  CreateFinanceTransactionRequest,
  FinanceTransaction,
  TransactionType,
  UpdateFinanceTransactionRequest,
} from '../types'
import { TRANSACTION_TYPE_OPTIONS } from '../types'

type Mode = 'create' | 'edit'

type TransactionFormProps = {
  mode: Mode
  initial?: FinanceTransaction
  submitLabel: string
  onSubmitCreate?: (payload: CreateFinanceTransactionRequest) => Promise<void>
  onSubmitUpdate?: (payload: UpdateFinanceTransactionRequest) => Promise<void>
  onCancel: () => void
}

type FieldErrors = {
  type?: string
  title?: string
  amount?: string
  category?: string
}

export function TransactionForm({
  mode,
  initial,
  submitLabel,
  onSubmitCreate,
  onSubmitUpdate,
  onCancel,
}: TransactionFormProps) {
  const [type, setType] = useState<TransactionType | ''>(initial?.type ?? 'INCOME')
  const [title, setTitle] = useState(initial?.title ?? '')
  const [amount, setAmount] = useState(initial ? String(initial.amount) : '')
  const [category, setCategory] = useState(initial?.category ?? '')
  const [note, setNote] = useState(initial?.note ?? '')
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const errors: FieldErrors = {}
    if (!type) {
      errors.type = 'Jenis transaksi wajib dipilih.'
    }
    if (!title.trim()) {
      errors.title = 'Judul wajib diisi.'
    }
    if (!category.trim()) {
      errors.category = 'Kategori wajib diisi.'
    }

    const parsedAmount = parsePositiveIntegerAmount(amount)
    if (!parsedAmount.ok) {
      errors.amount = parsedAmount.message
    }

    setFieldErrors(errors)
    if (Object.keys(errors).length > 0 || !parsedAmount.ok) {
      return
    }

    setSubmitting(true)
    try {
      if (mode === 'create' && onSubmitCreate) {
        await onSubmitCreate({
          type: type as TransactionType,
          title: title.trim(),
          amount: parsedAmount.value,
          category: category.trim(),
          note: note.trim(),
        })
      }
      if (mode === 'edit' && onSubmitUpdate) {
        await onSubmitUpdate({
          type: type as TransactionType,
          title: title.trim(),
          amount: parsedAmount.value,
          category: category.trim(),
          note: note.trim(),
        })
      }
    } catch (error) {
      setFormError(toFinanceErrorMessage(error))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form className="flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
      <Select
        name="type"
        label="Jenis"
        value={type}
        onChange={(event) => setType(event.target.value as TransactionType | '')}
        options={TRANSACTION_TYPE_OPTIONS}
        error={fieldErrors.type}
        disabled={submitting}
        required
      />
      <Input
        name="title"
        label="Judul"
        value={title}
        onChange={(event) => setTitle(event.target.value)}
        error={fieldErrors.title}
        disabled={submitting}
        required
      />
      <Input
        name="amount"
        label="Nominal (Rp)"
        inputMode="numeric"
        placeholder="1500000"
        value={amount}
        onChange={(event) => setAmount(event.target.value)}
        error={fieldErrors.amount}
        disabled={submitting}
        required
      />
      <Input
        name="category"
        label="Kategori"
        value={category}
        onChange={(event) => setCategory(event.target.value)}
        error={fieldErrors.category}
        disabled={submitting}
        required
      />
      <Textarea
        name="note"
        label="Catatan"
        value={note}
        onChange={(event) => setNote(event.target.value)}
        disabled={submitting}
      />

      {formError ? (
        <p
          className="rounded-2xl border border-red-100 bg-[var(--color-danger-soft)] px-3 py-2 text-sm text-[var(--color-danger)]"
          role="alert"
        >
          {formError}
        </p>
      ) : null}

      <div className="flex justify-end gap-2">
        <Button type="button" variant="secondary" onClick={onCancel} disabled={submitting}>
          Batal
        </Button>
        <Button type="submit" loading={submitting}>
          {submitLabel}
        </Button>
      </div>
    </form>
  )
}
