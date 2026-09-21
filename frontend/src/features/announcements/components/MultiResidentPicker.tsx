import { useMemo, useState } from 'react'
import { useDebouncedValue } from '../../../lib/hooks/useDebouncedValue'
import { listResidents } from '../../residents/api'
import { useResidents } from '../../residents/hooks'

export type ResidentOption = {
  id: string
  name: string
  phone: string
}

type MultiResidentPickerProps = {
  selected: ResidentOption[]
  onChange: (residents: ResidentOption[]) => void
  allSelected: boolean
  onAllSelectedChange: (selected: boolean) => void
  disabled?: boolean
  error?: string
}

export function MultiResidentPicker({
  selected,
  onChange,
  allSelected,
  onAllSelectedChange,
  disabled = false,
  error,
}: MultiResidentPickerProps) {
  const [search, setSearch] = useState('')
  const [selectingAll, setSelectingAll] = useState(false)
  const [selectAllError, setSelectAllError] = useState<string | null>(null)
  const debouncedSearch = useDebouncedValue(search, 300)

  const filters = useMemo(
    () => ({
      page: 1,
      page_size: 100,
      search: debouncedSearch.trim(),
      gender: '' as const,
    }),
    [debouncedSearch],
  )

  const { data, isFetching, isError } = useResidents(filters)
  const options = data?.items ?? []

  function toggleResident(resident: ResidentOption) {
    if (selected.some((item) => item.id === resident.id)) {
      onChange(selected.filter((item) => item.id !== resident.id))
      onAllSelectedChange(false)
    } else {
      onChange([...selected, resident])
    }
  }

  async function handleSelectAll(checked: boolean) {
    onAllSelectedChange(checked)
    setSelectAllError(null)
    if (!checked) {
      onChange([])
      return
    }

    setSelectingAll(true)
    try {
      const first = await listResidents({
        page: 1,
        page_size: 100,
        search: '',
        gender: '',
      })
      const residents = [...first.items]
      for (let page = 2; page <= first.meta.total_pages; page += 1) {
        const next = await listResidents({
          page,
          page_size: 100,
          search: '',
          gender: '',
        })
        residents.push(...next.items)
      }
      onChange(residents.map(({ id, name, phone }) => ({ id, name, phone })))
    } catch {
      onAllSelectedChange(false)
      setSelectAllError('Gagal memilih semua warga. Silakan coba lagi.')
    } finally {
      setSelectingAll(false)
    }
  }

  return (
    <div className="space-y-1.5">
      <label className="text-sm font-semibold text-[var(--color-ink)]">Dikirim kepada</label>
      <details
        className={[
          'group overflow-hidden rounded-2xl border bg-[var(--color-surface)]',
          error ? 'border-[var(--color-danger)]' : 'border-[var(--color-line)]',
          disabled ? 'pointer-events-none opacity-60' : '',
        ].join(' ')}
      >
        <summary className="flex cursor-pointer list-none items-center justify-between gap-3 px-3.5 py-2.5 text-sm text-[var(--color-ink)]">
          <span>
            {selectingAll
              ? 'Memilih semua warga…'
              : allSelected
                ? `Semua warga (${selected.length})`
                : selected.length > 0
                  ? `${selected.length} warga dipilih`
                  : 'Pilih warga'}
          </span>
          <span
            className="material-symbols-outlined text-[20px] text-[var(--color-muted)] transition group-open:rotate-180"
            aria-hidden
          >
            expand_more
          </span>
        </summary>

        <div className="space-y-2 border-t border-[var(--color-line)] p-3">
          <input
            name="recipientSearch"
            placeholder="Cari nama atau nomor telepon…"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            disabled={disabled || selectingAll}
            className="w-full rounded-xl border border-[var(--color-line)] bg-[var(--color-panel)] px-3 py-2 text-sm text-[var(--color-ink)] outline-none placeholder:text-[var(--color-muted)]/60 focus:border-[var(--color-tertiary)] focus:ring-4 focus:ring-[var(--color-tertiary)]/15"
          />

          <label className="flex cursor-pointer items-center gap-3 rounded-lg px-2 py-2 text-sm font-semibold hover:bg-[var(--color-accent-soft)]">
            <input
              type="checkbox"
              checked={allSelected}
              onChange={(event) => {
                void handleSelectAll(event.target.checked)
              }}
              disabled={disabled || selectingAll}
              className="h-4 w-4 accent-[var(--color-accent)]"
            />
            Semua warga
          </label>

          {selectAllError ? (
            <p className="text-sm text-[var(--color-danger)]">{selectAllError}</p>
          ) : null}

          <ul className="max-h-48 overflow-y-auto" role="listbox" aria-label="Pilihan warga">
            {isFetching ? (
              <li className="px-2 py-2 text-sm text-[var(--color-muted)]">Mencari…</li>
            ) : null}
            {isError ? (
              <li className="px-2 py-2 text-sm text-[var(--color-danger)]">
                Gagal memuat warga.
              </li>
            ) : null}
            {!isFetching && !isError && options.length === 0 ? (
              <li className="px-2 py-2 text-sm text-[var(--color-muted)]">
                Tidak ada warga ditemukan.
              </li>
            ) : null}
            {options.map((resident) => (
              <li key={resident.id}>
                <label className="flex cursor-pointer items-center gap-3 rounded-lg px-2 py-2 text-sm hover:bg-[var(--color-accent-soft)]">
                  <input
                    type="checkbox"
                    checked={selected.some((item) => item.id === resident.id)}
                    onChange={() =>
                      toggleResident({
                        id: resident.id,
                        name: resident.name,
                        phone: resident.phone,
                      })
                    }
                    disabled={disabled || selectingAll}
                    className="h-4 w-4 shrink-0 accent-[var(--color-accent)]"
                  />
                  <span className="flex min-w-0 flex-col">
                    <span className="font-medium text-[var(--color-ink)]">{resident.name}</span>
                    <span className="text-[var(--color-muted)]">{resident.phone}</span>
                  </span>
                </label>
              </li>
            ))}
          </ul>
        </div>
      </details>
      {error ? <p className="text-sm text-[var(--color-danger)]">{error}</p> : null}
    </div>
  )
}
