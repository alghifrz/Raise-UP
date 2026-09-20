import { useState } from 'react'
import { cn } from '../../lib/utils'

type ImageWithFallbackProps = {
  src: string
  alt: string
  className?: string
  fallbackLabel?: string
}

export function ImageWithFallback({
  src,
  alt,
  className,
  fallbackLabel = 'Gambar tidak tersedia',
}: ImageWithFallbackProps) {
  const [failedSrc, setFailedSrc] = useState<string | null>(null)
  const failed = !src || failedSrc === src

  if (failed) {
    return (
      <div
        className={cn(
          'flex items-center justify-center bg-[var(--color-accent-soft)] text-center text-xs text-[var(--color-muted)]',
          className,
        )}
        role="img"
        aria-label={fallbackLabel}
      >
        {fallbackLabel}
      </div>
    )
  }

  return (
    <img
      src={src}
      alt={alt}
      className={cn('object-cover', className)}
      onError={() => setFailedSrc(src)}
    />
  )
}
