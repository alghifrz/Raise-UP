export type SiteSettings = {
  id: string
  site_name: string
  tagline: string
  chairman_name: string
  chairman_role: string
  chairman_quote: string
  chairman_photo_url: string | null
  chairman_photo_storage_path: string | null
  map_title: string
  map_description: string
  maps_url: string | null
  embed_url: string | null
  address: string
  phone: string
  whatsapp_url: string | null
  footer_blurb: string
  updated_at: string
}

export type UpdateSiteSettingsRequest = {
  site_name?: string
  tagline?: string
  chairman_name?: string
  chairman_role?: string
  chairman_quote?: string
  chairman_photo_url?: string
  chairman_photo_storage_path?: string
  map_title?: string
  map_description?: string
  maps_url?: string
  embed_url?: string
  address?: string
  phone?: string
  whatsapp_url?: string
  footer_blurb?: string
}
