-- Rebrand public portal defaults to Lebak Asri (RT name).
UPDATE site_settings
SET
    site_name = 'Lebak Asri',
    tagline = CASE
        WHEN tagline = '' OR tagline = 'Portal Administrasi RW' THEN 'Portal Digital RT / RW Lebak Asri'
        ELSE tagline
    END
WHERE site_name = '' OR site_name = 'RAISE UP';
