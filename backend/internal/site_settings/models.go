package sitesettings

// UpdateRequest is the JSON body for PATCH /api/v1/site-settings.
type UpdateRequest struct {
	SiteName                 *string `json:"site_name"`
	Tagline                  *string `json:"tagline"`
	ChairmanName             *string `json:"chairman_name"`
	ChairmanRole             *string `json:"chairman_role"`
	ChairmanQuote            *string `json:"chairman_quote"`
	ChairmanPhotoURL         *string `json:"chairman_photo_url"`
	ChairmanPhotoStoragePath *string `json:"chairman_photo_storage_path"`
	MapTitle                 *string `json:"map_title"`
	MapDescription           *string `json:"map_description"`
	MapsURL                  *string `json:"maps_url"`
	EmbedURL                 *string `json:"embed_url"`
	Address                  *string `json:"address"`
	Phone                    *string `json:"phone"`
	WhatsappURL              *string `json:"whatsapp_url"`
	FooterBlurb              *string `json:"footer_blurb"`
}

// Settings is the API-safe site settings representation.
type Settings struct {
	ID                       string  `json:"id"`
	SiteName                 string  `json:"site_name"`
	Tagline                  string  `json:"tagline"`
	ChairmanName             string  `json:"chairman_name"`
	ChairmanRole             string  `json:"chairman_role"`
	ChairmanQuote            string  `json:"chairman_quote"`
	ChairmanPhotoURL         *string `json:"chairman_photo_url"`
	ChairmanPhotoStoragePath *string `json:"chairman_photo_storage_path"`
	MapTitle                 string  `json:"map_title"`
	MapDescription           string  `json:"map_description"`
	MapsURL                  *string `json:"maps_url"`
	EmbedURL                 *string `json:"embed_url"`
	Address                  string  `json:"address"`
	Phone                    string  `json:"phone"`
	WhatsappURL              *string `json:"whatsapp_url"`
	FooterBlurb              string  `json:"footer_blurb"`
	UpdatedAt                string  `json:"updated_at"`
}
