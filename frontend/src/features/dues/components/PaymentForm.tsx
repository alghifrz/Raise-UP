import { useState, type FormEvent } from 'react'
import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import { ResidentPicker, type ResidentOption } from '../../complaints/components/ResidentPicker'
import {
  fromDatetimeLocalValue,
  formatIdr,
  parsePositiveIntegerAmount,
  toDatetimeLocalValue,
} from '../../../lib/utils'
import { toDuesErrorMessage } from '../errors'
import type { CreateDuesPaymentRequest, DuesPayment, DuesPeriod, UpdateDuesPaymentRequest } from '../types'

type Mode = 'create' | 'edit'

type PaymentFormProps = {
  mode: Mode
  period: DuesPeriod
  initial?: DuesPayment
  submitLabel: string
  onSubmitCreate?: (payload: CreateDuesPaymentRequest) => Promise<void>
  onSubmitUpdate?: (payload: UpdateDuesPaymentRequest) => Promise<void>
  onCancel: () => void
}

type FieldErrors = {
  resident?: string
  amount?: string
  paid_at?: string
}

export function PaymentForm({
  mode,
  period,
  initial,
  submitLabel,
  onSubmitCreate,
  onSubmitUpdate,
  onCancel,
}: PaymentFormProps) {
  const [resident, setResident] = useState<ResidentOption | null>(
    initial
      ? {
          id: initial.resident_id,
          name: initial.resident_name,
          phone: '',
        }
      : null,
  )
  const [amount, setAmount] = useState(String(initial?.amount ?? period.amount))
  const [paidAtLocal, setPaidAtLocal] = useState(
    initial?.paid_at ? toDatetimeLocalValue(initial.paid_at) : toDatetimeLocalValue(new Date().toISOString()),
  )
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const errors: FieldErrors = {}
    if (!resident) {
      errors.resident = 'Pilih warga pembayar.'
    }

    const parsedAmount = parsePositiveIntegerAmount(amount)
    if (!parsedAmount.ok) {
      errors.amount = parsedAmount.message
    } else if (parsedAmount.value !== period.amount) {
      errors.amount = `Nominal harus sama dengan nominal periode (${formatIdr(period.amount)}).`
    }

    let paidAt = ''
    if (!paidAtLocal) {
      errors.paid_at = 'Tanggal bayar wajib diisi.'
    } else {
      try {
        paidAt = fromDatetimeLocalValue(paidAtLocal)
      } catch {
        errors.paid_at = 'Tanggal bayar tidak valid.'
      }
    }

    setFieldErrors(errors)
    if (Object.keys(errors).length > 0 || !parsedAmount.ok || !resident) {
      return
    }

    setSubmitting(true)
    try {
      if (mode === 'create' && onSubmitCreate) {
        await onSubmitCreate({
          period_id: period.id,
          resident_id: resident.id,
          amount: parsedAmount.value,
          paid_at: paidAt,
        })
      }
      if (mode === 'edit' && onSubmitUpdate) {
        await onSubmitUpdate({
          period_id: period.id,
          resident_id: resident.id,
          amount: parsedAmount.value,
          paid_at: paidAt,
        })
      }
    } catch (error) {
      setFormError(toDuesErrorMessage(error))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form className="flex flex-col gap-4" onSubmit={handleSubmit} noValidate>
      <p className="rounded-2xl border border-[var(--color-line)] bg-[var(--color-accent-soft)] px-3 py-2 text-sm text-[var(--color-muted)]">
        Nominal periode: <strong className="text-[var(--color-ink)]">{formatIdr(period.amount)}</strong>
      </p>

      <div>
        <ResidentPicker selected={resident} onSelect={setResident} disabled={submitting} />
        {fieldErrors.resident ? (
          <p className="mt-1 text-sm text-[var(--color-danger)]">{fieldErrors.resident}</p>
        ) : null}
      </div>

      <Input
        name="amount"
        label="Nominal pembayaran (Rp)"
        inputMode="numeric"
        value={amount}
        onChange={(event) => setAmount(event.target.value)}
        error={fieldErrors.amount}
        disabled={submitting}
        required
      />
      <Input
        name="paid_at"
        label="Tanggal bayar"
        type="datetime-local"
        value={paidAtLocal}
        onChange={(event) => setPaidAtLocal(event.target.value)}
        error={fieldErrors.paid_at}
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
