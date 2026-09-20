import { useMemo, useState } from 'react'
import { Button } from '../../../components/ui/Button'
import { Input } from '../../../components/ui/Input'
import { useDebouncedValue } from '../../../lib/hooks/useDebouncedValue'
import { cn } from '../../../lib/utils'
import { useResidents } from '../../residents/hooks'

export type ResidentOption = {
  id: string
  name: string
  phone: string
}

type MultiResidentPickerProps = {
  selected: ResidentOption[]
  onChange: (residents: ResidentOption[]) => void
  disabled?: boolean
  error?: string
}

export function MultiResidentPicker({
  selected,
  onChange,
  disabled = false,
  error,
}: MultiResidentPickerProps) {
  const [search, setSearch] = useState('')
  const debouncedSearch = useDebouncedValue(search, 300)

  const filters = useMemo(
    () => ({
      page: 1,
      page_size: 8,
      search: debouncedSearch.trim(),
      gender: '' as const,
    }),
    [debouncedSearch],
  )

  const { data, isFetching, isError } = useResidents(filters)
  const options = (data?.items ?? []).filter(
    (resident) => !selected.some((item) => item.id === resident.id),
  )

  function addResident(resident: ResidentOption) {
    if (selected.some((item) => item.id === resident.id)) {
      return
    }
    onChange([...selected, resident])
    setSearch('')
  }

  function removeResident(id: string) {
    onChange(selected.filter((item) => item.id !== id))
  }

  return (
    <div className="space-y-3">
      <Input
        name="recipientSearch"
        label="Cari penerima warga"
        placeholder="Ketik nama atau telepon"
        value={search}
        onChange={(event) => setSearch(event.target.value)}
        disabled={disabled}
        error={error}
      />

      {selected.length > 0 ? (
        <ul className="flex flex-wrap gap-2" aria-label="Penerima terpilih">
          {selected.map((resident) => (
            <li
              key={resident.id}
              className="inline-flex items-center gap-2 rounded-md border border-[color-mix(in_srgb,var(--color-secondary)_40%,transparent)] bg-[color-mix(in_srgb,var(--color-secondary)_22%,white)] px-2 py-1 text-sm text-[var(--color-accent)]"
            >
              <span>
                {resident.name} · {resident.phone}
              </span>
              <Button
                type="button"
                variant="ghost"
                className="px-1 py-0 text-xs"
                onClick={() => removeResident(resident.id)}
                disabled={disabled}
                aria-label={`Hapus ${resident.name}`}
              >
                Hapus
              </Button>
            </li>
          ))}
        </ul>
      ) : (
        <p className="text-sm text-[var(--color-muted)]">Belum ada penerima dipilih.</p>
      )}

      <ul
        className="max-h-48 overflow-y-auto rounded-md border border-[var(--color-line)]"
        role="listbox"
        aria-label="Hasil pencarian warga"
      >
        {isFetching ? <li className="px-3 py-2 text-sm text-[var(--color-muted)]">Mencari…</li> : null}
        {isError ? (
          <li className="px-3 py-2 text-sm text-[var(--color-danger)]">Gagal memuat warga.</li>
        ) : null}
        {!isFetching && !isError && options.length === 0 ? (
          <li className="px-3 py-2 text-sm text-[var(--color-muted)]">Tidak ada warga ditemukan.</li>
        ) : null}
        {options.map((resident) => (
          <li key={resident.id}>
            <button
              type="button"
              role="option"
              aria-selected={false}
              className={cn(
                'flex w-full flex-col items-start px-3 py-2 text-left text-sm hover:bg-[var(--color-accent-soft)]',
                'focus-visible:bg-[var(--color-accent-soft)]',
              )}
              onClick={() =>
                addResident({ id: resident.id, name: resident.name, phone: resident.phone })
              }
              disabled={disabled}
            >
              <span className="font-medium text-[var(--color-ink)]">{resident.name}</span>
              <span className="text-[var(--color-muted)]">{resident.phone}</span>
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}
