export type NavItem = {
  to: string
  label: string
  icon: string
}

export const adminNavItems: NavItem[] = [
  { to: '/dashboard', label: 'Dashboard', icon: 'dashboard' },
  { to: '/chat', label: 'Chat', icon: 'chat' },
  { to: '/residents', label: 'Warga', icon: 'groups' },
  { to: '/complaints', label: 'Pengaduan', icon: 'report' },
  { to: '/announcements', label: 'Pengumuman', icon: 'campaign' },
  { to: '/activities', label: 'Kegiatan', icon: 'event' },
  { to: '/finance', label: 'Keuangan', icon: 'payments' },
  { to: '/dues', label: 'Iuran', icon: 'account_balance_wallet' },
  { to: '/gallery', label: 'Galeri', icon: 'photo_library' },
  { to: '/village', label: 'Profil Wilayah', icon: 'home_pin' },
  { to: '/site-settings', label: 'Pengaturan Situs', icon: 'settings' },
]
