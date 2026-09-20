import type { ReactNode } from 'react'
import { useState } from 'react'
import { cn } from '../../../lib/utils'

type IconProps = {
  name: string
  className?: string
  filled?: boolean
}

export function Icon({ name, className, filled = false }: IconProps) {
  return (
    <span
      className={cn('material-symbols-outlined', className)}
      style={filled ? { fontVariationSettings: "'FILL' 1, 'wght' 400, 'GRAD' 0, 'opsz' 24" } : undefined}
      aria-hidden
    >
      {name}
    </span>
  )
}

export function LandingContainer({
  children,
  className,
}: {
  children: ReactNode
  className?: string
}) {
  return (
    <div className={cn('mx-auto w-full max-w-[1280px] px-5 lg:px-10', className)}>{children}</div>
  )
}

/** Renders an image; if the path 404s, shows `fallback` instead. */
export function SafeImage({
  src,
  alt,
  className,
  fallback,
}: {
  src: string | null | undefined
  alt: string
  className?: string
  fallback?: ReactNode
}) {
  const [failed, setFailed] = useState(false)

  if (!src || failed) {
    return <>{fallback ?? null}</>
  }

  return <img src={src} alt={alt} className={className} onError={() => setFailed(true)} />
}
