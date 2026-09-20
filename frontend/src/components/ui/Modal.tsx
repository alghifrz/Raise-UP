import { useEffect, useId, useRef, type ReactNode } from 'react'
import { Button } from './Button'
import { cn } from '../../lib/utils'

type ModalProps = {
  open: boolean
  title: string
  description?: string
  onClose: () => void
  children: ReactNode
  className?: string
}

export function Modal({ open, title, description, onClose, children, className }: ModalProps) {
  const titleId = useId()
  const descriptionId = useId()
  const dialogRef = useRef<HTMLDialogElement>(null)

  useEffect(() => {
    const dialog = dialogRef.current
    if (!dialog) {
      return
    }

    if (open && !dialog.open) {
      dialog.showModal()
    } else if (!open && dialog.open) {
      dialog.close()
    }
  }, [open])

  return (
    <dialog
      ref={dialogRef}
      className={cn(
        'm-auto max-h-[min(90vh,40rem)] w-[min(100%-2rem,32rem)] flex-col overflow-hidden rounded-3xl border border-[var(--color-line)] bg-[var(--color-panel)] p-0 shadow-xl',
        'backdrop:bg-[var(--color-ink)]/40',
        '[&:not([open])]:hidden [&[open]]:flex',
        className,
      )}
      aria-labelledby={titleId}
      aria-describedby={description ? descriptionId : undefined}
      onClose={onClose}
      onClick={(event) => {
        if (event.target === dialogRef.current) {
          onClose()
        }
      }}
    >
      <div className="flex shrink-0 items-start justify-between gap-3 border-b border-[var(--color-line)] px-5 py-4">
        <div>
          <h2 id={titleId} className="text-lg font-bold tracking-tight text-[var(--color-ink)]">
            {title}
          </h2>
          {description ? (
            <p id={descriptionId} className="mt-1 text-sm text-[var(--color-muted)]">
              {description}
            </p>
          ) : null}
        </div>
        <Button type="button" variant="ghost" className="rounded-full px-3" onClick={onClose} aria-label="Tutup dialog">
          Tutup
        </Button>
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto overscroll-contain px-5 py-4">{children}</div>
    </dialog>
  )
}

type ConfirmDialogProps = {
  open: boolean
  title: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  loading?: boolean
  tone?: 'danger' | 'primary'
  onConfirm: () => void
  onCancel: () => void
}

export function ConfirmDialog({
  open,
  title,
  message,
  confirmLabel = 'Ya, lanjutkan',
  cancelLabel = 'Batal',
  loading = false,
  tone = 'danger',
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  return (
    <Modal open={open} title={title} onClose={onCancel}>
      <p className="text-sm text-[var(--color-muted)]">{message}</p>
      <div className="mt-5 flex justify-end gap-2">
        <Button type="button" variant="secondary" onClick={onCancel} disabled={loading}>
          {cancelLabel}
        </Button>
        <Button
          type="button"
          variant={tone === 'danger' ? 'danger' : 'primary'}
          loading={loading}
          onClick={onConfirm}
        >
          {confirmLabel}
        </Button>
      </div>
    </Modal>
  )
}
