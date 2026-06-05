# Require to make a bash script

$shimsDir = "$env:USERPROFILE\.inusdk\shims"

Write-Host "Building InuSDK. . ."
go build -o "build/inusdk.exe"

if ($LASTEXITCODE -ne 0)
{
   Write-Host "Build failed" -ForegroundColor Red
   exit 1
}

Write-Host "Copying to shim"
Copy-Item "build/inusdk.exe" "$shimsDir\inusdk.exe" -Force

Write-Host "Done - Run `inusdk` in a new terminal" -ForegroundColor Green
