/** Resolve admin nav title for the current pathname (supports nested routes). */
export function resolveNavLabel(pathname: string, items: Array<{ to: string; label: string }>): string {
  const exact = items.find((item) => item.to === pathname)
  if (exact) {
    return exact.label
  }

  const nested = items
    .filter((item) => item.to !== '/' && pathname.startsWith(`${item.to}/`))
    .sort((a, b) => b.to.length - a.to.length)[0]

  return nested?.label ?? 'Admin'
}

/** Whether a nav item should appear active for the current pathname. */
export function isNavItemActive(pathname: string, to: string): boolean {
  return pathname === to || pathname.startsWith(`${to}/`)
}
