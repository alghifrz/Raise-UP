import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { ErrorState } from '../../components/ui/ErrorState'
import { useAuth } from '../auth/useAuth'
import { landingAssets, landingBrand } from '../landing/data'
import { formatDuesPeriodLabel, formatIdr, cn } from '../../lib/utils'
import { DashboardSkeleton } from './components/DashboardSkeleton'
import { useDashboardSummary } from './hooks'
import type { DashboardSummary } from './types'

const fadeUp = {
  hidden: { opacity: 0, y: 16 },
  show: (i: number) => ({
    opacity: 1,
    y: 0,
    transition: { delay: 0.05 * i, duration: 0.4, ease: [0.22, 1, 0.36, 1] as const },
  }),
}

function greetingForHour(hour: number): string {
  if (hour < 11) return 'Selamat pagi'
  if (hour < 15) return 'Selamat siang'
  if (hour < 18) return 'Selamat sore'
  return 'Selamat malam'
}

function pct(part: number, total: number): number {
  if (total <= 0) return 0
  return Math.min(100, Math.round((part / total) * 100))
}

function IconBubble({
  name,
  className,
}: {
  name: string
  className?: string
}) {
  return (
    <span
      className={cn(
        'inline-flex h-11 w-11 items-center justify-center rounded-2xl',
        className,
      )}
      aria-hidden
    >
      <span className="material-symbols-outlined text-[22px]">{name}</span>
    </span>
  )
}

function ProgressBar({
  value,
  tone = 'accent',
}: {
  value: number
  tone?: 'accent' | 'lime' | 'warn' | 'danger' | 'muted'
}) {
  const toneClass = {
    accent: 'bg-[var(--color-tertiary)]',
    lime: 'bg-[var(--color-secondary)]',
    warn: 'bg-amber-400',
    danger: 'bg-[var(--color-danger)]',
    muted: 'bg-[var(--color-line)]',
  }[tone]

  return (
    <div className="h-2.5 overflow-hidden rounded-full bg-[var(--color-accent-soft)]">
      <motion.div
        className={cn('h-full rounded-full', toneClass)}
        initial={{ width: 0 }}
        animate={{ width: `${Math.max(0, Math.min(100, value))}%` }}
        transition={{ duration: 0.8, ease: [0.22, 1, 0.36, 1] }}
      />
    </div>
  )
}

function KpiCard({
  index,
  icon,
  label,
  value,
  hint,
  to,
  featured,
}: {
  index: number
  icon: string
  label: string
  value: string
  hint?: string
  to: string
  featured?: boolean
}) {
  return (
    <motion.div custom={index} variants={fadeUp} initial="hidden" animate="show">
      <Link
        to={to}
        className={cn(
          'group relative flex h-full flex-col overflow-hidden rounded-3xl border p-5 transition',
          featured
            ? 'border-transparent bg-[var(--color-accent)] text-[var(--color-surface)] shadow-[0_24px_48px_-28px_rgba(13,59,63,0.7)]'
            : 'border-[var(--color-line)] bg-[var(--color-panel)] shadow-[0_12px_32px_-24px_rgba(15,37,39,0.35)] hover:border-[var(--color-tertiary)]/40 hover:shadow-[0_18px_40px_-24px_rgba(15,37,39,0.4)]',
        )}
      >
        <div className="flex items-start justify-between gap-3">
          <IconBubble
            name={icon}
            className={
              featured
                ? 'bg-[var(--color-secondary)] text-[var(--color-secondary-ink)]'
                : 'bg-[var(--color-accent-soft)] text-[var(--color-accent)]'
            }
          />
          <span
            className={cn(
              'material-symbols-outlined text-[18px] transition group-hover:translate-x-0.5',
              featured ? 'text-white/50' : 'text-[var(--color-muted)]',
            )}
            aria-hidden
          >
            arrow_outward
          </span>
        </div>
        <p
          className={cn(
            'mt-4 text-[11px] font-bold uppercase tracking-[0.14em]',
            featured ? 'text-white/65' : 'text-[var(--color-muted)]',
          )}
        >
          {label}
        </p>
        <p
          className={cn(
            'mt-1 text-2xl font-extrabold tracking-tight tabular-nums sm:text-[1.75rem]',
            featured ? 'text-white' : 'text-[var(--color-ink)]',
          )}
        >
          {value}
        </p>
        {hint ? (
          <p className={cn('mt-1 text-xs', featured ? 'text-white/60' : 'text-[var(--color-muted)]')}>
            {hint}
          </p>
        ) : null}
        {featured ? (
          <div
            className="pointer-events-none absolute -right-8 -bottom-10 h-36 w-36 rounded-full bg-[var(--color-secondary)]/20 blur-2xl"
            aria-hidden
          />
        ) : null}
      </Link>
    </motion.div>
  )
}

function SectionCard({
  index,
  icon,
  title,
  description,
  action,
  children,
  className,
}: {
  index: number
  icon: string
  title: string
  description?: string
  action?: { to: string; label: string }
  children: ReactNode
  className?: string
}) {
  return (
    <motion.section
      custom={index}
      variants={fadeUp}
      initial="hidden"
      animate="show"
      className={cn(
        'rounded-3xl border border-[var(--color-line)] bg-[var(--color-panel)] p-5 shadow-[0_12px_32px_-24px_rgba(15,37,39,0.3)] sm:p-6',
        className,
      )}
    >
      <div className="mb-5 flex items-start justify-between gap-3">
        <div className="flex items-start gap-3">
          <IconBubble
            name={icon}
            className="bg-[var(--color-accent-soft)] text-[var(--color-accent)]"
          />
          <div>
            <h2 className="text-base font-bold tracking-tight text-[var(--color-ink)]">{title}</h2>
            {description ? (
              <p className="mt-0.5 text-sm text-[var(--color-muted)]">{description}</p>
            ) : null}
          </div>
        </div>
        {action ? (
          <Link
            to={action.to}
            className="inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-xs font-bold text-[var(--color-tertiary)] transition hover:bg-[var(--color-accent-soft)]"
          >
            {action.label}
            <span className="material-symbols-outlined text-[14px]" aria-hidden>
              chevron_right
            </span>
          </Link>
        ) : null}
      </div>
      {children}
    </motion.section>
  )
}

function DashboardBody({ data, isFetching }: { data: DashboardSummary; isFetching: boolean }) {
  const { user } = useAuth()
  const now = new Date()
  const greeting = greetingForHour(now.getHours())
  const dateLabel = new Intl.DateTimeFormat('id-ID', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    timeZone: 'Asia/Jakarta',
  }).format(now)

  const periodLabel = data.dues.period
    ? formatDuesPeriodLabel(data.dues.period.year, data.dues.period.month, data.dues.period.half)
    : 'Belum ada periode iuran'

  const duesPaidPct = pct(data.dues.paid_count, data.dues.resident_count)
  const duesCollectPct = pct(data.dues.collected_total, data.dues.expected_total)
  const openComplaints = data.complaints.baru + data.complaints.diproses
  const malePct = pct(data.residents.male, data.residents.total)
  const femalePct = pct(data.residents.female, data.residents.total)

  const complaintRows = [
    { label: 'Baru', value: data.complaints.baru, tone: 'warn' as const },
    { label: 'Diproses', value: data.complaints.diproses, tone: 'accent' as const },
    { label: 'Selesai', value: data.complaints.selesai, tone: 'lime' as const },
    { label: 'Ditolak', value: data.complaints.ditolak, tone: 'danger' as const },
  ]

  const quickLinks = [
    { to: '/residents', icon: 'person_add', label: 'Tambah warga' },
    { to: '/complaints', icon: 'report', label: 'Pengaduan' },
    { to: '/announcements', icon: 'campaign', label: 'Pengumuman' },
    { to: '/finance', icon: 'payments', label: 'Keuangan' },
    { to: '/dues', icon: 'account_balance_wallet', label: 'Iuran' },
    { to: '/activities', icon: 'event', label: 'Kegiatan' },
  ]

  return (
    <div className="space-y-5">
      <motion.header
        custom={0}
        variants={fadeUp}
        initial="hidden"
        animate="show"
        className="relative overflow-hidden rounded-[1.75rem] border border-[var(--color-line)] bg-[var(--color-accent)] px-5 py-6 text-[var(--color-surface)] shadow-[0_28px_60px_-32px_rgba(13,59,63,0.75)] sm:px-7 sm:py-7"
      >
        <div
          className="pointer-events-none absolute inset-0 opacity-90"
          style={{
            background:
              'radial-gradient(ellipse 60% 80% at 100% 0%, rgba(210,248,67,0.28), transparent 55%), radial-gradient(ellipse 50% 70% at 0% 100%, rgba(28,91,98,0.55), transparent 50%)',
          }}
          aria-hidden
        />
        <div className="relative flex flex-wrap items-end justify-between gap-5">
          <div className="min-w-0">
            <div className="mb-4 flex items-center gap-3">
              <span className="flex h-12 w-12 items-center justify-center overflow-hidden rounded-2xl bg-white/10 ring-1 ring-white/15">
                <img src={landingAssets.logo} alt="" className="h-9 w-9 object-contain" />
              </span>
              <div>
                <p className="text-[11px] font-bold uppercase tracking-[0.16em] text-[var(--color-secondary)]">
                  {landingBrand.siteName}
                </p>
                <p className="text-xs text-white/60">{landingBrand.rtLabel}</p>
              </div>
            </div>
            <h1 className="text-[clamp(1.5rem,3vw,2rem)] font-extrabold tracking-tight text-white">
              {greeting}
              {user?.name ? `, ${user.name.split(' ')[0]}` : ''}
            </h1>
            <p className="mt-1.5 max-w-xl text-sm text-white/70">
              Ringkasan operasional {landingBrand.siteName} — pantau warga, keuangan, iuran, dan
              layanan warga dalam satu tampilan.
            </p>
            <p className="mt-3 text-xs font-medium text-white/50">{dateLabel}</p>
          </div>

          <div className="flex flex-col items-end gap-2">
            {isFetching ? (
              <span className="inline-flex items-center gap-1.5 rounded-full bg-white/10 px-3 py-1 text-[11px] font-semibold text-white/80">
                <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-[var(--color-secondary)]" />
                Memperbarui…
              </span>
            ) : (
              <span className="inline-flex items-center gap-1.5 rounded-full bg-[var(--color-secondary)]/20 px-3 py-1 text-[11px] font-bold text-[var(--color-secondary)]">
                <span className="h-1.5 w-1.5 rounded-full bg-[var(--color-secondary)]" />
                Data terkini
              </span>
            )}
            <Link
              to="/"
              className="inline-flex items-center gap-1.5 rounded-full border border-white/15 bg-white/8 px-3.5 py-2 text-xs font-semibold text-white/85 transition hover:bg-white/14"
            >
              <span className="material-symbols-outlined text-[16px]" aria-hidden>
                public
              </span>
              Lihat portal publik
            </Link>
          </div>
        </div>
      </motion.header>

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <KpiCard
          index={1}
          icon="groups"
          label="Total warga"
          value={String(data.residents.total)}
          hint={`${data.residents.male} L · ${data.residents.female} P`}
          to="/residents"
          featured
        />
        <KpiCard
          index={2}
          icon="account_balance"
          label="Saldo kas"
          value={formatIdr(data.finance.balance)}
          hint={`Masuk ${formatIdr(data.finance.income)}`}
          to="/finance"
        />
        <KpiCard
          index={3}
          icon="report"
          label="Pengaduan aktif"
          value={String(openComplaints)}
          hint={`${data.complaints.total} total`}
          to="/complaints"
        />
        <KpiCard
          index={4}
          icon="savings"
          label="Iuran terkumpul"
          value={`${duesCollectPct}%`}
          hint={periodLabel}
          to="/dues"
        />
      </div>

      <div className="grid gap-4 lg:grid-cols-5">
        <SectionCard
          index={5}
          icon="payments"
          title="Keuangan"
          description="Ringkasan arus kas"
          action={{ to: '/finance', label: 'Detail' }}
          className="lg:col-span-2"
        >
          <div className="space-y-4">
            <div className="rounded-2xl bg-[var(--color-accent-soft)]/80 px-4 py-4">
              <p className="text-[11px] font-bold uppercase tracking-[0.14em] text-[var(--color-muted)]">
                Saldo saat ini
              </p>
              <p className="mt-1 text-3xl font-extrabold tracking-tight text-[var(--color-accent)] tabular-nums">
                {formatIdr(data.finance.balance)}
              </p>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="rounded-2xl border border-[var(--color-line)] px-3.5 py-3">
                <p className="text-[11px] font-bold uppercase tracking-wide text-[var(--color-muted)]">
                  Pemasukan
                </p>
                <p className="mt-1 text-lg font-bold text-[var(--color-tertiary)] tabular-nums">
                  {formatIdr(data.finance.income)}
                </p>
              </div>
              <div className="rounded-2xl border border-[var(--color-line)] px-3.5 py-3">
                <p className="text-[11px] font-bold uppercase tracking-wide text-[var(--color-muted)]">
                  Pengeluaran
                </p>
                <p className="mt-1 text-lg font-bold text-[var(--color-danger)] tabular-nums">
                  {formatIdr(data.finance.expense)}
                </p>
              </div>
            </div>
          </div>
        </SectionCard>

        <SectionCard
          index={6}
          icon="account_balance_wallet"
          title="Iuran periode"
          description={periodLabel}
          action={{ to: '/dues', label: 'Kelola' }}
          className="lg:col-span-3"
        >
          <div className="grid gap-5 sm:grid-cols-[auto_1fr] sm:items-center">
            <div className="relative mx-auto flex h-36 w-36 items-center justify-center">
              <svg viewBox="0 0 120 120" className="h-full w-full -rotate-90" aria-hidden>
                <circle
                  cx="60"
                  cy="60"
                  r="48"
                  fill="none"
                  stroke="var(--color-accent-soft)"
                  strokeWidth="12"
                />
                <motion.circle
                  cx="60"
                  cy="60"
                  r="48"
                  fill="none"
                  stroke="var(--color-secondary)"
                  strokeWidth="12"
                  strokeLinecap="round"
                  strokeDasharray={2 * Math.PI * 48}
                  initial={{ strokeDashoffset: 2 * Math.PI * 48 }}
                  animate={{
                    strokeDashoffset: 2 * Math.PI * 48 * (1 - duesPaidPct / 100),
                  }}
                  transition={{ duration: 1, ease: [0.22, 1, 0.36, 1] }}
                />
              </svg>
              <div className="absolute inset-0 flex flex-col items-center justify-center text-center">
                <span className="text-2xl font-extrabold tabular-nums text-[var(--color-accent)]">
                  {duesPaidPct}%
                </span>
                <span className="text-[10px] font-bold uppercase tracking-wide text-[var(--color-muted)]">
                  Lunas
                </span>
              </div>
            </div>

            <div className="space-y-3">
              <div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
                {[
                  { label: 'Warga', value: String(data.dues.resident_count) },
                  { label: 'Lunas', value: String(data.dues.paid_count) },
                  { label: 'Belum', value: String(data.dues.unpaid_count) },
                  { label: 'Target', value: formatIdr(data.dues.expected_total) },
                  { label: 'Terkumpul', value: formatIdr(data.dues.collected_total) },
                  { label: 'Tunggakan', value: formatIdr(data.dues.outstanding_total) },
                ].map((item) => (
                  <div
                    key={item.label}
                    className="rounded-2xl border border-[var(--color-line)] bg-[var(--color-accent-soft)]/50 px-3 py-2.5"
                  >
                    <p className="text-[10px] font-bold uppercase tracking-wide text-[var(--color-muted)]">
                      {item.label}
                    </p>
                    <p className="mt-0.5 truncate text-sm font-bold tabular-nums text-[var(--color-ink)]">
                      {item.value}
                    </p>
                  </div>
                ))}
              </div>
              <div>
                <div className="mb-1.5 flex justify-between text-xs font-semibold text-[var(--color-muted)]">
                  <span>Progress pengumpulan</span>
                  <span className="tabular-nums text-[var(--color-accent)]">{duesCollectPct}%</span>
                </div>
                <ProgressBar value={duesCollectPct} tone="lime" />
              </div>
            </div>
          </div>
        </SectionCard>
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <SectionCard
          index={7}
          icon="report"
          title="Pengaduan"
          description={`${data.complaints.total} total aduan`}
          action={{ to: '/complaints', label: 'Lihat' }}
        >
          <div className="space-y-3.5">
            {complaintRows.map((row) => {
              const share = pct(row.value, data.complaints.total)
              return (
                <div key={row.label}>
                  <div className="mb-1.5 flex items-center justify-between gap-2 text-sm">
                    <span className="font-semibold text-[var(--color-ink)]">{row.label}</span>
                    <span className="tabular-nums text-[var(--color-muted)]">
                      {row.value} · {share}%
                    </span>
                  </div>
                  <ProgressBar value={share} tone={row.tone} />
                </div>
              )
            })}
          </div>
        </SectionCard>

        <SectionCard
          index={8}
          icon="groups"
          title="Demografi warga"
          description="Sebaran jenis kelamin"
          action={{ to: '/residents', label: 'Kelola' }}
        >
          <div className="space-y-4">
            <div className="flex h-4 overflow-hidden rounded-full bg-[var(--color-accent-soft)]">
              <motion.div
                className="bg-[var(--color-tertiary)]"
                initial={{ width: 0 }}
                animate={{ width: `${malePct}%` }}
                transition={{ duration: 0.8, ease: [0.22, 1, 0.36, 1] }}
              />
              <motion.div
                className="bg-[var(--color-secondary)]"
                initial={{ width: 0 }}
                animate={{ width: `${femalePct}%` }}
                transition={{ duration: 0.8, delay: 0.1, ease: [0.22, 1, 0.36, 1] }}
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="rounded-2xl border border-[var(--color-line)] px-3.5 py-3">
                <div className="flex items-center gap-2">
                  <span className="h-2.5 w-2.5 rounded-full bg-[var(--color-tertiary)]" />
                  <span className="text-xs font-bold text-[var(--color-muted)]">Laki-laki</span>
                </div>
                <p className="mt-2 text-2xl font-extrabold tabular-nums text-[var(--color-ink)]">
                  {data.residents.male}
                </p>
                <p className="text-xs text-[var(--color-muted)]">{malePct}%</p>
              </div>
              <div className="rounded-2xl border border-[var(--color-line)] px-3.5 py-3">
                <div className="flex items-center gap-2">
                  <span className="h-2.5 w-2.5 rounded-full bg-[var(--color-secondary)]" />
                  <span className="text-xs font-bold text-[var(--color-muted)]">Perempuan</span>
                </div>
                <p className="mt-2 text-2xl font-extrabold tabular-nums text-[var(--color-ink)]">
                  {data.residents.female}
                </p>
                <p className="text-xs text-[var(--color-muted)]">{femalePct}%</p>
              </div>
            </div>
          </div>
        </SectionCard>

        <SectionCard
          index={9}
          icon="campaign"
          title="Konten & kegiatan"
          description="Status publikasi"
          action={{ to: '/announcements', label: 'Kelola' }}
        >
          <div className="grid grid-cols-2 gap-3">
            {[
              {
                icon: 'draft',
                label: 'Draft',
                value: data.announcements.draft,
                to: '/announcements',
              },
              {
                icon: 'published_with_changes',
                label: 'Terbit',
                value: data.announcements.published,
                to: '/announcements',
              },
              {
                icon: 'event_upcoming',
                label: 'Mendatang',
                value: data.activities.upcoming,
                to: '/activities',
              },
              {
                icon: 'event',
                label: 'Total kegiatan',
                value: data.activities.total,
                to: '/activities',
              },
            ].map((item) => (
              <Link
                key={item.label}
                to={item.to}
                className="rounded-2xl border border-[var(--color-line)] bg-[var(--color-accent-soft)]/40 px-3.5 py-3 transition hover:border-[var(--color-tertiary)]/35 hover:bg-[var(--color-accent-soft)]"
              >
                <span
                  className="material-symbols-outlined text-[20px] text-[var(--color-tertiary)]"
                  aria-hidden
                >
                  {item.icon}
                </span>
                <p className="mt-2 text-[10px] font-bold uppercase tracking-wide text-[var(--color-muted)]">
                  {item.label}
                </p>
                <p className="text-xl font-extrabold tabular-nums text-[var(--color-ink)]">
                  {item.value}
                </p>
              </Link>
            ))}
          </div>
        </SectionCard>
      </div>

      <motion.section
        custom={10}
        variants={fadeUp}
        initial="hidden"
        animate="show"
        className="rounded-3xl border border-[var(--color-line)] bg-[var(--color-panel)] p-5 shadow-[0_12px_32px_-24px_rgba(15,37,39,0.3)] sm:p-6"
      >
        <div className="mb-4 flex items-center gap-3">
          <IconBubble
            name="bolt"
            className="bg-[color-mix(in_srgb,var(--color-secondary)_35%,white)] text-[var(--color-accent)]"
          />
          <div>
            <h2 className="text-base font-bold tracking-tight text-[var(--color-ink)]">
              Aksi cepat
            </h2>
            <p className="text-sm text-[var(--color-muted)]">Loncat ke modul yang sering dipakai</p>
          </div>
        </div>
        <div className="grid grid-cols-2 gap-2.5 sm:grid-cols-3 lg:grid-cols-6">
          {quickLinks.map((link) => (
            <Link
              key={link.to}
              to={link.to}
              className="group flex flex-col items-center gap-2 rounded-2xl border border-[var(--color-line)] px-3 py-4 text-center transition hover:border-[var(--color-secondary)] hover:bg-[color-mix(in_srgb,var(--color-secondary)_18%,white)]"
            >
              <span className="flex h-10 w-10 items-center justify-center rounded-xl bg-[var(--color-accent-soft)] text-[var(--color-accent)] transition group-hover:bg-[var(--color-accent)] group-hover:text-[var(--color-secondary)]">
                <span className="material-symbols-outlined text-[22px]" aria-hidden>
                  {link.icon}
                </span>
              </span>
              <span className="text-xs font-bold text-[var(--color-ink)]">{link.label}</span>
            </Link>
          ))}
        </div>
      </motion.section>
    </div>
  )
}

export function DashboardPage() {
  const { data, isLoading, isError, refetch, isFetching } = useDashboardSummary()

  if (isLoading) {
    return <DashboardSkeleton />
  }

  if (isError || !data) {
    return (
      <div className="space-y-4">
        <header>
          <h1 className="text-2xl font-bold text-[var(--color-ink)]">Dashboard</h1>
          <p className="mt-1 text-sm text-[var(--color-muted)]">Ringkasan operasional RW</p>
        </header>
        <ErrorState
          title="Gagal memuat dashboard."
          message="Data ringkasan tidak dapat diambil. Periksa koneksi atau coba lagi."
          onRetry={() => {
            void refetch()
          }}
        />
      </div>
    )
  }

  return <DashboardBody data={data} isFetching={isFetching} />
}
