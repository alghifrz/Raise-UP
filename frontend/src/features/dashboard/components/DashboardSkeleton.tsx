import { Skeleton } from '../../../components/ui/LoadingState'

export function DashboardSkeleton() {
  return (
    <div className="space-y-5" aria-busy="true" aria-label="Memuat dashboard">
      <Skeleton className="h-44 w-full rounded-[1.75rem]" />

      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-36 rounded-3xl" />
        ))}
      </div>

      <div className="grid gap-4 lg:grid-cols-5">
        <Skeleton className="h-64 rounded-3xl lg:col-span-2" />
        <Skeleton className="h-64 rounded-3xl lg:col-span-3" />
      </div>

      <div className="grid gap-4 lg:grid-cols-3">
        <Skeleton className="h-56 rounded-3xl" />
        <Skeleton className="h-56 rounded-3xl" />
        <Skeleton className="h-56 rounded-3xl" />
      </div>

      <Skeleton className="h-40 rounded-3xl" />
    </div>
  )
}
