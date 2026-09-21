param(
    [Parameter(Mandatory = $true)]
    [string]$SupabaseUrl,

    [Parameter(Mandatory = $true)]
    [string]$ServerKey,

    [string]$Bucket = "gallery",
    [int64]$MaxUploadBytes = 10485760
)

$ErrorActionPreference = "Stop"
$baseUrl = $SupabaseUrl.TrimEnd("/")
$headers = @{
    apikey = $ServerKey
}
if (-not $ServerKey.StartsWith("sb_secret_")) {
    $headers.Authorization = "Bearer $ServerKey"
}
$body = @{
    id                 = $Bucket
    name               = $Bucket
    public             = $true
    file_size_limit    = $MaxUploadBytes
    allowed_mime_types = @("image/jpeg", "image/png", "image/webp")
} | ConvertTo-Json

try {
    Invoke-RestMethod `
        -Method Post `
        -Uri "$baseUrl/storage/v1/bucket" `
        -Headers $headers `
        -ContentType "application/json" `
        -Body $body | Out-Null
    Write-Host "Created public bucket '$Bucket'."
}
catch {
    $statusCode = 0
    if ($_.Exception.Response -and $_.Exception.Response.StatusCode) {
        $statusCode = [int]$_.Exception.Response.StatusCode
    }
    if ($statusCode -eq 409 -or $_.ErrorDetails.Message -match '"statusCode":"409"') {
        Invoke-RestMethod `
            -Method Put `
            -Uri "$baseUrl/storage/v1/bucket/$Bucket" `
            -Headers $headers `
            -ContentType "application/json" `
            -Body $body | Out-Null
        Write-Host "Updated existing public bucket '$Bucket'."
    }
    else {
        throw
    }
}
