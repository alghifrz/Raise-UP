import { useReducedMotion, type Variants } from 'framer-motion'

const ease = [0.22, 1, 0.36, 1] as const

export function useRevealVariants() {
  const reduce = useReducedMotion()

  const container: Variants = {
    hidden: {},
    visible: {
      transition: {
        staggerChildren: reduce ? 0 : 0.12,
        delayChildren: reduce ? 0 : 0.08,
      },
    },
  }

  const item: Variants = {
    hidden: reduce ? { opacity: 1, y: 0 } : { opacity: 0, y: 28 },
    visible: {
      opacity: 1,
      y: 0,
      transition: { duration: reduce ? 0 : 0.7, ease },
    },
  }

  return { container, item, reduce }
}
