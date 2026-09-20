import { useState, type FormEvent } from 'react'
import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import { Select } from '../../../components/ui/Select'
import { parsePositiveIntegerAmount } from '../../../lib/utils'
import { toDuesErrorMessage } from '../errors'
import type { CreateDuesPeriodRequest, DuesPeriod, UpdateDuesPeriodRequest } from '../types'
import { HALF_OPTIONS, MONTH_OPTIONS } from '../types'

type Mode = 'create' | 'edit'

type PeriodFormProps = {
  mode: Mode
  initial?: DuesPeriod
  submitLabel: string
  onSubmitCreate?: (payload: CreateDuesPeriodRequest) => Promise<void>
  onSubmitUpdate?: (payload: UpdateDuesPeriodRequest) => Promise<void>
  onCancel: () => void
}

type FieldErrors = {
  year?: string
  month?: string
  half?: string
  amount?: string
}

export function PeriodForm({
  mode,
  initial,
  submitLabel,
  onSubmitCreate,
  onSubmitUpdate,
  onCancel,
}: PeriodFormProps) {
  const [year, setYear] = useState(String(initial?.year ?? new Date().getFullYear()))
  const [month, setMonth] = useState(String(initial?.month ?? new Date().getMonth() + 1))
  const [half, setHalf] = useState(String(initial?.half ?? 1))
  const [amount, setAmount] = useState(initial ? String(initial.amount) : '')
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const errors: FieldErrors = {}
    const yearNum = Number.parseInt(year, 10)
    const monthNum = Number.parseInt(month, 10)
    const halfNum = Number.parseInt(half, 10)

    if (!/^\d+$/.test(year.trim()) || yearNum < 2000 || yearNum > 2100) {
      errors.year = 'Tahun harus antara 2000–2100.'
    }
    if (!month || monthNum < 1 || monthNum > 12) {
      errors.month = 'Bulan wajib dipilih.'
    }
    if (halfNum !== 1 && halfNum !== 2) {
      errors.half = 'Half wajib 1 atau 2.'
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
      const payload = {
        year: yearNum,
        month: monthNum,
        half: halfNum,
        amount: parsedAmount.value,
      }
      if (mode === 'create' && onSubmitCreate) {
        await onSubmitCreate(payload)
      }
      if (mode === 'edit' && onSubmitUpdate) {
        await onSubmitUpdate(payload)
      }
    } catch (error) {
      setFormError(toDuesErrorMessage(error))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form className="flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
      <Input
        name="year"
        label="Tahun"
        inputMode="numeric"
        value={year}
        onChange={(event) => setYear(event.target.value)}
        error={fieldErrors.year}
        disabled={submitting}
        required
      />
      <Select
        name="month"
        label="Bulan"
        value={month}
        onChange={(event) => setMonth(event.target.value)}
        options={MONTH_OPTIONS}
        error={fieldErrors.month}
        disabled={submitting}
        required
      />
      <Select
        name="half"
        label="Half"
        value={half}
        onChange={(event) => setHalf(event.target.value)}
        options={HALF_OPTIONS}
        error={fieldErrors.half}
        disabled={submitting}
        required
      />
      <Input
        name="amount"
        label="Nominal (Rp)"
        inputMode="numeric"
        placeholder="50000"
        value={amount}
        onChange={(event) => setAmount(event.target.value)}
        error={fieldErrors.amount}
        disabled={submitting}
        required
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
