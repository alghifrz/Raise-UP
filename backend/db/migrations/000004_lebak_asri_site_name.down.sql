UPDATE site_settings
SET
    site_name = 'RAISE UP',
    tagline = CASE
        WHEN tagline = 'Portal Digital RT / RW Lebak Asri' THEN 'Portal Administrasi RW'
        ELSE tagline
    END
WHERE site_name = 'Lebak Asri';
