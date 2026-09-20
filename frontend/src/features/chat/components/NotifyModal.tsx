import { useState } from 'react'
import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import { InlineAlert } from '../../../components/ui/InlineAlert'
import { Modal } from '../../../components/ui/Modal'
import { Textarea } from '../../../components/ui/Textarea'
import { cn } from '../../../lib/utils'
import { toChatErrorMessage } from '../errors'
import { useSendWhatsAppNotification } from '../hooks'

type NotifyModalProps = {
  open: boolean
  onClose: () => void
  enabled: boolean
}

type NotifyMode = 'text' | 'template'

export function NotifyModal({ open, onClose, enabled }: NotifyModalProps) {
  const [mode, setMode] = useState<NotifyMode>('text')
  const [to, setTo] = useState('')
  const [body, setBody] = useState('')
  const [templateName, setTemplateName] = useState('')
  const [language, setLanguage] = useState('id')
  const [feedback, setFeedback] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const sendMutation = useSendWhatsAppNotification()

  function reset() {
    setMode('text')
    setTo('')
    setBody('')
    setTemplateName('')
    setLanguage('id')
    setFeedback(null)
    setSuccess(null)
  }

  function handleClose() {
    reset()
    onClose()
  }

  async function handleSubmit() {
    setFeedback(null)
    setSuccess(null)
    if (!enabled) {
      setFeedback('WhatsApp belum dikonfigurasi di server.')
      return
    }
    if (!to.trim()) {
      setFeedback('Nomor tujuan wajib diisi.')
      return
    }

    try {
      const result = await sendMutation.mutateAsync(
        mode === 'text'
          ? { to: to.trim(), type: 'text', body: body.trim() }
          : {
              to: to.trim(),
              type: 'template',
              template_name: templateName.trim(),
              language: language.trim() || 'id',
            },
      )
      setSuccess(`Terkirim (${result.message_id})`)
    } catch (error) {
      setFeedback(toChatErrorMessage(error))
    }
  }

  return (
    <Modal
      open={open}
      title="Kirim notifikasi WA"
      description="Uji kirim text (dalam window 24 jam) atau template yang sudah disetujui Meta."
      onClose={handleClose}
    >
      {!enabled ? (
        <InlineAlert tone="danger">
          Set WHATSAPP_* di .env lalu restart API agar notifikasi aktif.
        </InlineAlert>
      ) : null}

      <div className="mt-3 mb-3 flex gap-1 rounded-2xl bg-[var(--color-surface)] p-1">
        {(['text', 'template'] as const).map((item) => (
          <button
            key={item}
            type="button"
            onClick={() => setMode(item)}
            className={cn(
              'flex-1 rounded-xl px-3 py-2 text-sm font-semibold transition',
              mode === item
                ? 'bg-[var(--color-panel)] text-[var(--color-ink)] shadow-sm'
                : 'text-[var(--color-muted)]',
            )}
          >
            {item === 'text' ? 'Text' : 'Template'}
          </button>
        ))}
      </div>

      <div className="space-y-3">
        <Input
          label="Nomor tujuan"
          name="notify_to"
          value={to}
          onChange={(event) => setTo(event.target.value)}
          placeholder="08… atau 628…"
          required
        />
        {mode === 'text' ? (
          <Textarea
            label="Pesan"
            name="notify_body"
            value={body}
            onChange={(event) => setBody(event.target.value)}
            placeholder="Isi pesan notifikasi…"
            rows={4}
            required
          />
        ) : (
          <>
            <Input
              label="Nama template"
              name="template_name"
              value={templateName}
              onChange={(event) => setTemplateName(event.target.value)}
              placeholder="hello_world"
              required
            />
            <Input
              label="Bahasa"
              name="language"
              value={language}
              onChange={(event) => setLanguage(event.target.value)}
              placeholder="id"
            />
          </>
        )}
      </div>

      {feedback ? (
        <div className="mt-3">
          <InlineAlert tone="danger">{feedback}</InlineAlert>
        </div>
      ) : null}
      {success ? (
        <div className="mt-3">
          <InlineAlert tone="success">{success}</InlineAlert>
        </div>
      ) : null}

      <div className="mt-4 flex justify-end gap-2">
        <Button type="button" variant="secondary" onClick={handleClose}>
          Tutup
        </Button>
        <Button
          type="button"
          loading={sendMutation.isPending}
          disabled={!enabled}
          onClick={() => void handleSubmit()}
        >
          Kirim
        </Button>
      </div>
    </Modal>
  )
}
