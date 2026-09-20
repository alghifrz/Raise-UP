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

type ResidentPickerProps = {
  selected: ResidentOption | null
  onSelect: (resident: ResidentOption | null) => void
  disabled?: boolean
}

export function ResidentPicker({ selected, onSelect, disabled = false }: ResidentPickerProps) {
  const [search, setSearch] = useState(selected?.name ?? '')
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

  const { data, isFetching } = useResidents(filters)
  const options = data?.items ?? []

  return (
    <div className="space-y-2">
      <Input
        name="residentSearch"
        label="Cari warga"
        placeholder="Ketik nama atau telepon"
        value={search}
        onChange={(event) => {
          setSearch(event.target.value)
          if (selected) {
            onSelect(null)
          }
        }}
        disabled={disabled}
      />

      {selected ? (
        <div className="rounded-md border border-[color-mix(in_srgb,var(--color-secondary)_40%,transparent)] bg-[color-mix(in_srgb,var(--color-secondary)_22%,white)] px-3 py-2 text-sm text-[var(--color-accent)]">
          Terpilih: <strong>{selected.name}</strong> · {selected.phone}
          <div className="mt-2">
            <Button
              type="button"
              variant="ghost"
              className="px-2 py-1"
              onClick={() => {
                onSelect(null)
                setSearch('')
              }}
              disabled={disabled}
            >
              Hapus pilihan
            </Button>
          </div>
        </div>
      ) : (
        <ul
          className="max-h-48 overflow-y-auto rounded-md border border-[var(--color-line)]"
          role="listbox"
          aria-label="Hasil pencarian warga"
        >
          {isFetching ? <li className="px-3 py-2 text-sm text-[var(--color-muted)]">Mencari…</li> : null}
          {!isFetching && options.length === 0 ? (
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
                onClick={() => {
                  onSelect({ id: resident.id, name: resident.name, phone: resident.phone })
                  setSearch(resident.name)
                }}
                disabled={disabled}
              >
                <span className="font-medium text-[var(--color-ink)]">{resident.name}</span>
                <span className="text-[var(--color-muted)]">{resident.phone}</span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
