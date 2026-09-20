import { useRef, useState } from 'react'
import { Button } from '../../../components/ui/Button'
import { InlineAlert } from '../../../components/ui/InlineAlert'
import { createResident } from '../api'
import { toResidentErrorMessage } from '../errors'
import {
  downloadResidentImportTemplate,
  MAX_RESIDENT_IMPORT_ROWS,
  parseResidentImportFile,
  toCreateResidentRequest,
  type ParsedResidentRow,
  type ResidentImportResult,
} from '../sheet'

type ImportResidentsModalProps = {
  onCancel: () => void
  onImported: (summary: { success: number; failed: number }) => void
}

export function ImportResidentsModal({ onCancel, onImported }: ImportResidentsModalProps) {
  const inputRef = useRef<HTMLInputElement>(null)
  const [fileName, setFileName] = useState<string | null>(null)
  const [rows, setRows] = useState<ParsedResidentRow[]>([])
  const [parseError, setParseError] = useState<string | null>(null)
  const [importing, setImporting] = useState(false)
  const [progress, setProgress] = useState({ done: 0, total: 0 })
  const [results, setResults] = useState<ResidentImportResult[] | null>(null)

  const validRows = rows.filter((row) => !row.error)
  const invalidRows = rows.filter((row) => row.error)

  async function handleFileChange(file: File | null) {
    setParseError(null)
    setResults(null)
    setRows([])
    setFileName(null)

    if (!file) {
      return
    }

    const lower = file.name.toLowerCase()
    if (!lower.endsWith('.xlsx') && !lower.endsWith('.xls') && !lower.endsWith('.csv')) {
      setParseError('Format file harus .xlsx, .xls, atau .csv.')
      return
    }

    try {
      const parsed = await parseResidentImportFile(file)
      setFileName(file.name)
      setRows(parsed)
      if (parsed.length === 0) {
        setParseError('Tidak ada baris data di file.')
      }
    } catch (error) {
      setParseError(error instanceof Error ? error.message : 'Gagal membaca file.')
    }
  }

  async function handleImport() {
    if (validRows.length === 0 || importing) {
      return
    }

    setImporting(true)
    setResults(null)
    setProgress({ done: 0, total: validRows.length })

    const nextResults: ResidentImportResult[] = []
    let success = 0
    let failed = 0

    for (const [index, row] of validRows.entries()) {
      const payload = toCreateResidentRequest(row)
      if (!payload) {
        failed += 1
        nextResults.push({
          rowNumber: row.rowNumber,
          name: row.name,
          phone: row.phone,
          ok: false,
          message: row.error ?? 'Data tidak valid.',
        })
        setProgress({ done: index + 1, total: validRows.length })
        continue
      }

      try {
        await createResident(payload)
        success += 1
        nextResults.push({
          rowNumber: row.rowNumber,
          name: row.name,
          phone: row.phone,
          ok: true,
          message: 'Berhasil',
        })
      } catch (error) {
        failed += 1
        nextResults.push({
          rowNumber: row.rowNumber,
          name: row.name,
          phone: row.phone,
          ok: false,
          message: toResidentErrorMessage(error),
        })
      }
      setProgress({ done: index + 1, total: validRows.length })
    }

    setResults(nextResults)
    setImporting(false)
    onImported({ success, failed })
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="rounded-2xl border border-[var(--color-line)] bg-[var(--color-surface)] p-4 text-sm text-[var(--color-muted)]">
        <ol className="list-decimal space-y-1 pl-4">
          <li>Unduh template Excel.</li>
          <li>Isi kolom <strong className="text-[var(--color-ink)]">nama</strong>,{' '}
            <strong className="text-[var(--color-ink)]">telepon</strong>, dan{' '}
            <strong className="text-[var(--color-ink)]">jenis_kelamin</strong> (Laki-laki / Perempuan).
          </li>
          <li>Unggah file (.xlsx / .xls / .csv), lalu impor. Maks. {MAX_RESIDENT_IMPORT_ROWS} baris.</li>
        </ol>
      </div>

      <div className="flex flex-wrap gap-2">
        <Button type="button" variant="secondary" onClick={downloadResidentImportTemplate} disabled={importing}>
          <span className="material-symbols-outlined text-[18px]" aria-hidden>
            download
          </span>
          Unduh template
        </Button>
        <Button
          type="button"
          variant="secondary"
          onClick={() => inputRef.current?.click()}
          disabled={importing}
        >
          <span className="material-symbols-outlined text-[18px]" aria-hidden>
            upload_file
          </span>
          Pilih file
        </Button>
        <input
          ref={inputRef}
          type="file"
          accept=".xlsx,.xls,.csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/vnd.ms-excel,text/csv"
          className="sr-only"
          onChange={(event) => {
            void handleFileChange(event.target.files?.[0] ?? null)
            event.target.value = ''
          }}
        />
      </div>

      {fileName ? (
        <p className="text-sm text-[var(--color-ink)]">
          File: <span className="font-semibold">{fileName}</span>
        </p>
      ) : null}

      {parseError ? <InlineAlert tone="danger">{parseError}</InlineAlert> : null}

      {rows.length > 0 ? (
        <div className="space-y-2">
          <p className="text-sm text-[var(--color-muted)]">
            {validRows.length} baris siap impor
            {invalidRows.length > 0 ? ` · ${invalidRows.length} baris bermasalah` : ''}
          </p>
          <div className="max-h-48 overflow-auto rounded-xl border border-[var(--color-line)]">
            <table className="min-w-full text-left text-xs">
              <thead className="sticky top-0 bg-[var(--color-surface)] text-[var(--color-muted)]">
                <tr>
                  <th className="px-3 py-2 font-semibold">Baris</th>
                  <th className="px-3 py-2 font-semibold">Nama</th>
                  <th className="px-3 py-2 font-semibold">Telepon</th>
                  <th className="px-3 py-2 font-semibold">JK</th>
                  <th className="px-3 py-2 font-semibold">Status</th>
                </tr>
              </thead>
              <tbody>
                {rows.slice(0, 50).map((row) => (
                  <tr key={row.rowNumber} className="border-t border-[var(--color-line)]">
                    <td className="px-3 py-2 tabular-nums">{row.rowNumber}</td>
                    <td className="px-3 py-2">{row.name || '—'}</td>
                    <td className="px-3 py-2">{row.phone || '—'}</td>
                    <td className="px-3 py-2">{row.genderRaw || '—'}</td>
                    <td
                      className={
                        row.error
                          ? 'px-3 py-2 text-[var(--color-danger)]'
                          : 'px-3 py-2 text-emerald-700'
                      }
                    >
                      {row.error ?? 'OK'}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {rows.length > 50 ? (
            <p className="text-xs text-[var(--color-muted)]">
              Menampilkan 50 baris pertama dari {rows.length} baris.
            </p>
          ) : null}
        </div>
      ) : null}

      {importing ? (
        <p className="text-sm text-[var(--color-muted)]" aria-live="polite">
          Mengimpor {progress.done}/{progress.total}…
        </p>
      ) : null}

      {results ? (
        <InlineAlert tone={results.some((item) => !item.ok) ? 'danger' : 'success'}>
          Selesai: {results.filter((item) => item.ok).length} berhasil,{' '}
          {results.filter((item) => !item.ok).length} gagal.
        </InlineAlert>
      ) : null}

      {results && results.some((item) => !item.ok) ? (
        <div className="max-h-32 overflow-auto rounded-xl border border-red-100 bg-[var(--color-danger-soft)] p-3 text-xs text-[var(--color-danger)]">
          {results
            .filter((item) => !item.ok)
            .slice(0, 20)
            .map((item) => (
              <p key={`${item.rowNumber}-${item.phone}`}>
                Baris {item.rowNumber} ({item.name || item.phone}): {item.message}
              </p>
            ))}
        </div>
      ) : null}

      <div className="flex justify-end gap-2">
        <Button type="button" variant="secondary" onClick={onCancel} disabled={importing}>
          {results ? 'Tutup' : 'Batal'}
        </Button>
        <Button
          type="button"
          loading={importing}
          disabled={validRows.length === 0 || Boolean(results)}
          onClick={() => {
            void handleImport()
          }}
        >
          Impor {validRows.length > 0 ? `${validRows.length} warga` : ''}
        </Button>
      </div>
    </div>
  )
}
