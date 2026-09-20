import { LandingNav } from '../components/LandingNav'
import { HeroSection } from '../components/Hero'
import {
  AboutMosaicSection,
  AgendaSection,
  GalleryStripSection,
  LandingFooter,
  LocationBanner,
  PartnersSection,
  ProgramsSection,
  TransparencySection,
} from '../components/Sections'
import { landingBrand } from '../data'
import {
  usePublicActivities,
  usePublicAnnouncements,
  usePublicGallery,
  usePublicSiteSettings,
  usePublicVillageOfficials,
  usePublicVillageProfile,
} from '../hooks'

export function LandingPage() {
  const siteQuery = usePublicSiteSettings()
  const villageQuery = usePublicVillageProfile()
  const officialsQuery = usePublicVillageOfficials()
  const announcementsQuery = usePublicAnnouncements()
  const activitiesQuery = usePublicActivities()
  const galleryQuery = usePublicGallery()

  const settings = siteQuery.data
  const siteName = settings?.site_name?.trim() || landingBrand.siteName
  const tagline = settings?.tagline?.trim() || landingBrand.tagline
  const contactHref = settings?.whatsapp_url?.trim() || '#kontak'
  const galleryItems = galleryQuery.data?.items
  const announcements = announcementsQuery.data?.items

  return (
    <div className="landing-root min-h-screen antialiased">
      <LandingNav siteName={siteName} contactHref={contactHref} />
      <main className="w-full bg-[var(--lp-surface)] pt-20">
        <HeroSection
          siteName={siteName}
          tagline={tagline}
          announcementCount={announcementsQuery.data?.meta.total ?? announcements?.length ?? 0}
          officialCount={officialsQuery.data?.length ?? 0}
          galleryCover={galleryItems?.[0]?.image_url}
        />
        <PartnersSection />
        <AboutMosaicSection profile={villageQuery.data} gallery={galleryItems} />
        <ProgramsSection items={announcements} />
        <AgendaSection items={activitiesQuery.data?.items} />
        <TransparencySection officials={officialsQuery.data} />
        <GalleryStripSection items={galleryItems} />
        <LocationBanner settings={settings} />
      </main>
      <LandingFooter settings={settings} />
    </div>
  )
}
