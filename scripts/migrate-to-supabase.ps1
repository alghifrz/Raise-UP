param(
    [Parameter(Mandatory = $true)]
    [string]$SourceDatabaseUrl,

    [Parameter(Mandatory = $true)]
    [string]$TargetDatabaseUrl,

    [string]$WorkDirectory = "tmp/supabase-migration"
)

$ErrorActionPreference = "Stop"

function Convert-ToDockerHostUrl([string]$DatabaseUrl) {
    $converted = $DatabaseUrl.Replace("@127.0.0.1:", "@host.docker.internal:").Replace("@localhost:", "@host.docker.internal:")
    if ($converted.StartsWith("postgresql://")) {
        $converted = "postgres://" + $converted.Substring("postgresql://".Length)
    }
    return $converted
}

$sourceUrl = Convert-ToDockerHostUrl $SourceDatabaseUrl
$targetUrl = Convert-ToDockerHostUrl $TargetDatabaseUrl
$root = Split-Path -Parent $PSScriptRoot
$workPath = Join-Path $root $WorkDirectory
$migrationPath = Join-Path $root "backend/db/migrations"
New-Item -ItemType Directory -Force -Path $workPath | Out-Null

$workMount = "$($workPath.Replace('\', '/')):/work"
$migrationMount = "$($migrationPath.Replace('\', '/')):/migrations:ro"

Write-Host "Creating source backups..."
docker run --rm -v $workMount postgres:16-alpine `
    pg_dump --dbname=$sourceUrl --format=custom --file=/work/source-full.dump
if ($LASTEXITCODE -ne 0) { throw "Full backup failed" }

docker run --rm -v $workMount postgres:16-alpine `
    pg_dump --dbname=$sourceUrl --data-only --schema=public `
    --exclude-table=public.schema_migrations --no-owner --no-privileges `
    --file=/work/source-data.sql
if ($LASTEXITCODE -ne 0) { throw "Data export failed" }

Write-Host "Applying repository migrations to Supabase..."
$previousErrorActionPreference = $ErrorActionPreference
$ErrorActionPreference = "Continue"
$migrateOutput = docker run --rm -v $migrationMount migrate/migrate:v4.18.1 `
    -path /migrations -database $targetUrl up 2>&1
$migrateExitCode = $LASTEXITCODE
$ErrorActionPreference = $previousErrorActionPreference
if ($migrateExitCode -ne 0) {
    $safeOutput = ($migrateOutput -join [Environment]::NewLine) `
        -replace 'postgres(?:ql)?://[^\s",]+', '<database-url-redacted>'
    Write-Error $safeOutput
    throw "Target migration failed"
}
if ($migrateOutput) {
    $migrateOutput | Write-Host
}

$tables = @(
    "announcement_recipients",
    "messages",
    "conversation_participants",
    "dues_finance_links",
    "whatsapp_bot_sessions",
    "dues_payments",
    "village_officials",
    "announcements",
    "conversations",
    "activities",
    "complaints",
    "cash_transactions",
    "dues_periods",
    "cash_reminders",
    "gallery_items",
    "site_settings",
    "village_profile",
    "residents",
    "users"
)
$qualifiedTables = ($tables | ForEach-Object { "public.$_" }) -join ", "
$truncateSql = "TRUNCATE TABLE $qualifiedTables CASCADE;"

Write-Host "Replacing seeded target rows with source data..."
docker run --rm postgres:16-alpine `
    psql --dbname=$targetUrl --set=ON_ERROR_STOP=1 --command=$truncateSql
if ($LASTEXITCODE -ne 0) { throw "Target truncate failed" }

docker run --rm -v $workMount postgres:16-alpine `
    psql --dbname=$targetUrl --set=ON_ERROR_STOP=1 `
    --command="SET session_replication_role = replica;" `
    --file=/work/source-data.sql `
    --command="SET session_replication_role = origin;"
if ($LASTEXITCODE -ne 0) { throw "Target restore failed" }

$countSql = ($tables | ForEach-Object {
    "SELECT '$_' AS table_name, COUNT(*) AS row_count FROM public.$_"
}) -join " UNION ALL "

Write-Host "Source row counts:"
docker run --rm postgres:16-alpine `
    psql --dbname=$sourceUrl --tuples-only --command=$countSql
if ($LASTEXITCODE -ne 0) { throw "Source verification failed" }

Write-Host "Target row counts:"
docker run --rm postgres:16-alpine `
    psql --dbname=$targetUrl --tuples-only --command=$countSql
if ($LASTEXITCODE -ne 0) { throw "Target verification failed" }

Write-Host "Migration finished. Keep source-full.dump until cutover is verified."
