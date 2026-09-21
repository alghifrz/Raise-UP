/**
 * Static content & asset paths for the public landing page (Lebak Asri).
 * Put images under `frontend/public/landing/` then point paths here.
 */

export const landingAssets = {
  logo: '/logo.png',
  /** Hero right-panel cover when gallery API has no photo yet */
  heroCover: '/landing.webp',
  /** Optional mosaic photos (order matches `landingMosaicTiles`) */
  mosaic: [
    '/1.webp',
    '/2.webp',
    '/3.webp',
    '/4.webp',
  ] as const,
  /** Optional gallery showcase when API gallery is empty */
  gallery: [
    '/landing/gallery-1.jpg',
    '/landing/gallery-2.jpg',
    '/landing/gallery-3.jpg',
  ] as const,
} as const

export const landingBrand = {
  siteName: 'Lebak Asri',
  tagline: 'Portal Digital RT Lebak Asri',
  rtLabel: 'RT 03 / RW 07',
  areaLabel: 'Cilandak Barat, Jakarta Selatan',
  fullAddress: 'RT 03 / RW 07 Lebak Asri, Cilandak Barat, Jakarta Selatan',
  mapQuery: 'Lebak Asri RT 03 RW 07 Cilandak Barat Jakarta Selatan',
  mapTitle: 'Peta Lokasi Lebak Asri',
  footerBlurb:
    'Wadah kolaborasi warga dan transparansi aksi nyata untuk lingkungan yang guyub, lestari, dan berdaya bersama.',
} as const

export const landingNavLinks = [
  { href: '#beranda', label: 'Beranda' },
  { href: '#tentang', label: 'Tentang' },
  { href: '#program', label: 'Pengumuman' },
  { href: '#agenda', label: 'Agenda' },
  { href: '#galeri', label: 'Galeri' },
  { href: '#kontak', label: 'Kontak' },
] as const

export const landingHero = {
  eyebrow: 'Gotong Royong & Portal Digital',
  headlineBefore: 'Membangun Lingkungan',
  headlineHighlight: 'Guyub, Transparan,',
  headlineAfter: 'dan Berdaya.',
  primaryCta: { href: '#program', label: 'Lihat Pengumuman' },
  secondaryCta: { href: '#tentang', label: 'Jelajahi Portal', icon: 'folder_open' },
  statPublic: {
    badge: 'Info Publik Terbuka',
    caption: 'Pengumuman publik aktif',
    icon: 'verified_user',
  },
  statTeam: {
    avatars: ['R', 'W', 'A'] as const,
    caption: 'Pengurus terhubung',
    badge: 'Tim',
  },
  coverBadge: {
    icon: 'park',
    subtitle: 'Komunitas digital & lingkungan guyub',
  },
} as const

export const landingPartners = {
  title: 'Didukung & Berkolaborasi Bersama Lembaga Lingkungan',
  items: [
    { icon: 'apartment', label: 'Balai Warga' },
    { icon: 'recycling', label: 'Bank Sampah' },
    { icon: 'groups', label: 'Karang Taruna' },
    { icon: 'volunteer_activism', label: 'Posyandu' },
    { icon: 'shield', label: 'Siskamling' },
  ],
} as const

export const landingAbout = {
  eyebrow: 'Tentang Portal',
  titleBefore: 'Empati Nyata dalam Aksi,',
  titleHighlight: 'Terhubung Transparan.',
  fallbackBlurb:
    'Memanfaatkan transparansi pencatatan digital dan musyawarah warga untuk menumbuhkan rasa kepemilikan serta kepedulian antar tetangga.',
  cta: { href: '#kontak', label: 'Pelajari Lebih Lanjut' },
} as const

export const landingMosaicTiles = [
  {
    label: 'Terbuka & Akuntabel',
    icon: 'account_balance',
    image: landingAssets.mosaic[0],
  },
  {
    label: 'Guyub Rukun Rembuk',
    icon: 'diversity_3',
    image: landingAssets.mosaic[1],
  },
  {
    label: 'Peduli Warga',
    icon: 'favorite',
    image: landingAssets.mosaic[2],
  },
  {
    label: 'Aman & Terjaga',
    icon: 'shield',
    image: landingAssets.mosaic[3],
  },
] as const

export const landingPrograms = {
  eyebrow: 'Informasi Publik',
  title: 'Pengumuman & Kabar Warga',
  empty: 'Belum ada pengumuman publik yang diterbitkan.',
} as const

export const landingAgenda = {
  eyebrow: 'Agenda',
  title: 'Kegiatan Mendatang',
  empty: 'Belum ada kegiatan terjadwal.',
} as const

export const landingTransparency = {
  eyebrow: 'Transparansi Nyata',
  title: 'Amanah Terjaga. Transparansi Tanpa Syarat.',
  body: 'Setiap iuran warga, pengumuman resmi, dan dokumentasi kegiatan tercatat digital melalui portal Lebak Asri — terbuka untuk publik dan warga.',
  cta: { href: '#kontak', label: 'Hubungi Pengurus' },
  mockupGreeting: 'Selamat datang,',
  mockupUser: 'Warga Lebak Asri',
  mockupBadge: 'RW',
  features: [
    {
      icon: 'done_all',
      title: 'Laporan Kas Digital',
      body: 'Pencatatan pemasukan iuran, saldo kas, dan pengeluaran tercatat rapi serta dapat diaudit oleh pengurus.',
    },
    {
      icon: 'campaign',
      title: 'Pengumuman Terarah',
      body: 'Informasi publik dan privat tersampaikan jelas, termasuk agenda dan pengingat kegiatan warga.',
    },
    {
      icon: 'photo_camera',
      title: 'Dokumentasi Terukur',
      body: 'Galeri kegiatan dan bukti visual progres lapangan tersedia untuk warga.',
    },
  ],
} as const

export const landingGallery = {
  eyebrow: 'Dokumentasi',
  title: 'Galeri Kegiatan',
  description: 'Cuplikan momen gotong royong, agenda warga, dan suasana lingkungan Lebak Asri.',
  cta: { href: '#kontak', label: 'Kirim dokumentasi' },
  emptyCaption: 'Belum ada foto diunggah',
  placeholders: [
    { icon: 'photo_camera', label: 'Gotong Royong', image: landingAssets.gallery[0] },
    { icon: 'groups', label: 'Silaturahmi Warga', image: landingAssets.gallery[1] },
    { icon: 'park', label: 'Lingkungan Asri', image: landingAssets.gallery[2] },
  ],
} as const

export const landingLocation = {
  eyebrowFallback: landingBrand.mapTitle,
  headline: `Melayani Warga Lebak Asri`,
  description:
    'Terbuka bagi setiap inisiatif kebaikan. Mari bertegur sapa langsung di Balai Warga atau koordinasikan kegiatan sosial bersama pengurus.',
  mapsCta: 'Buka di Google Maps',
  contactCta: 'Hubungi Pengurus',
  whatsappCta: 'Grup / WhatsApp',
  badges: [
    { icon: 'location_on', label: 'Balai Warga Utama' },
    { icon: 'shield', label: 'Pos Kamling 24 Jam' },
    { icon: 'yard', label: 'Taman Kolaborasi Hijau' },
  ],
} as const

export const landingFooter = {
  quickLinks: [
    { href: '#beranda', label: 'Beranda' },
    { href: '#tentang', label: 'Tentang' },
    { href: '#agenda', label: 'Agenda' },
    { href: '#galeri', label: 'Galeri' },
  ],
  serviceLinks: [
    { href: '#program', label: 'Pengumuman' },
    { href: '#transparansi', label: 'Transparansi' },
    { href: '#kontak', label: 'Kontak' },
    { href: '/login', label: 'Masuk Admin' },
  ],
  social: [
    { icon: 'public', href: '#beranda' },
    { icon: 'share', href: '#program' },
    { icon: 'chat', hrefKey: 'whatsapp' as const },
    { icon: 'mail', hrefKey: 'phone' as const },
  ] as Array<{ icon: string; href?: string; hrefKey?: 'whatsapp' | 'phone' }>,
  newsletterTitle: 'Kabar Komunitas',
  newsletterBody: 'Pantau pengumuman publik dan agenda kegiatan langsung dari portal ini.',
  newsletterCta: { href: '#program', label: 'Lihat Kabar Terbaru' },
  values: ['Transparan', 'Inklusif', 'Berkelanjutan'] as const,
  copyrightSuffix: 'Bersama menjaga harmoni dan keberlanjutan.',
}
