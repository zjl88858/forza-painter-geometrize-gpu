$ErrorActionPreference = 'Continue'

$root = Split-Path -Parent $MyInvocation.MyCommand.Path

$benchBin      = Join-Path $root 'benchmark\bin'
$benchSettings = Join-Path $root 'benchmark\settings'
$benchImages   = Join-Path $root 'benchmark\images'
$benchRuntime  = Join-Path $root 'benchmark\runtime'

# -------------------------------------------------------------------
# Validate inputs
# -------------------------------------------------------------------
if (-not (Test-Path $benchBin)) {
    Write-Host "ERROR: benchmark\bin folder not found at $benchBin" -ForegroundColor Red
    exit 1
}
if (-not (Test-Path $benchSettings)) {
    Write-Host "ERROR: benchmark\settings folder not found at $benchSettings" -ForegroundColor Red
    exit 1
}
if (-not (Test-Path $benchImages)) {
    Write-Host "ERROR: benchmark\images folder not found at $benchImages" -ForegroundColor Red
    exit 1
}

$exes      = @(Get-ChildItem -Path $benchBin -Filter '*.exe' | Select-Object -ExpandProperty Name)
$settings  = @(Get-ChildItem -Path $benchSettings -Filter '*.ini' | Select-Object -ExpandProperty Name)
$images    = @(Get-ChildItem -Path "$benchImages\*" -Include '*.png','*.jpg','*.jpeg','*.bmp','*.webp','*.tiff','*.tif' | Select-Object -ExpandProperty Name)

if ($exes.Count -eq 0) {
    Write-Host "ERROR: No .exe files found in $benchBin" -ForegroundColor Red
    exit 1
}
if ($settings.Count -eq 0) {
    Write-Host "ERROR: No .ini settings files found in $benchSettings" -ForegroundColor Red
    exit 1
}
if ($images.Count -eq 0) {
    Write-Host "ERROR: No image files found in $benchImages" -ForegroundColor Red
    exit 1
}

# -------------------------------------------------------------------
# Timestamped run folder
# -------------------------------------------------------------------
$ts = Get-Date -Format 'yyyyMMdd-HHmmss'
$runDir = Join-Path $benchRuntime $ts
New-Item -ItemType Directory -Force -Path $runDir | Out-Null

Write-Host "==============================================" -ForegroundColor Cyan
Write-Host " Benchmark Run: $ts" -ForegroundColor Cyan
Write-Host " EXEs:      $($exes -join ', ')" -ForegroundColor White
Write-Host " Settings:  $($settings -join ', ')" -ForegroundColor White
Write-Host " Images:    $($images -join ', ')" -ForegroundColor White
Write-Host " Output:    $runDir" -ForegroundColor White
Write-Host "==============================================" -ForegroundColor Cyan
Write-Host ""

$totalCombinations = $exes.Count * $settings.Count * $images.Count
$current = 0

# -------------------------------------------------------------------
# Summary accumulator (for final CSV)
# -------------------------------------------------------------------
$summaryRows = @()

# -------------------------------------------------------------------
# Helper: parse GPU name from log
# -------------------------------------------------------------------
function Get-GpuNameFromLog {
    param([string]$logPath)
    if (-not (Test-Path $logPath)) { return "unknown-gpu" }
    $log = Get-Content $logPath -ErrorAction SilentlyContinue
    foreach ($line in $log) {
        if ($line -match 'OpenCL: Selected device "(.+?)"') {
            return $Matches[1]
        }
        if ($line -match 'Vulkan: Selected device "(.+?)"') {
            return $Matches[1]
        }
    }
    return "unknown-gpu"
}

# -------------------------------------------------------------------
# Helper: parse step timings from log
# -------------------------------------------------------------------
function Get-TimingsFromLog {
    param([string]$logPath)
    if (-not (Test-Path $logPath)) { return $null }
    $log = Get-Content $logPath -ErrorAction SilentlyContinue
    $timings = @()
    foreach ($line in $log) {
        if ($line -match 'Step completed in (.+)$') {
            $timeStr = $Matches[1]
            if ($timeStr -match '([\d.]+)ms') {
                $timings += [double]$Matches[1]
            }
            elseif ($timeStr -match '([\d.]+)s') {
                $timings += [double]$Matches[1] * 1000
            }
        }
    }
    if ($timings.Count -eq 0) { return $null }
    return $timings
}

# -------------------------------------------------------------------
# Main benchmark loop:
#   For each exe → for each settings file → for each image
#   Output organized as: runDir/<image_name>/<exe>_<settings>/
# -------------------------------------------------------------------
foreach ($exeName in $exes) {
    $exePath = Join-Path $benchBin $exeName
    $exeBase = [System.IO.Path]::GetFileNameWithoutExtension($exeName)

    foreach ($iniName in $settings) {
        $iniPath = Join-Path $benchSettings $iniName
        $iniBase = [System.IO.Path]::GetFileNameWithoutExtension($iniName)

        foreach ($img in $images) {
            $current++
            $imgBase = $img
            $imgPath = Join-Path $benchImages $img

            $imgDir = Join-Path $runDir $imgBase
            New-Item -ItemType Directory -Force -Path $imgDir | Out-Null

            $label = "${exeBase}__${iniBase}"
            $comboDir = Join-Path $imgDir $label
            New-Item -ItemType Directory -Force -Path $comboDir | Out-Null

            $logFile   = Join-Path $comboDir 'run.log'
            $previewPath = Join-Path $comboDir 'preview'
            $outputPath  = Join-Path $comboDir 'output'

            $progressMsg = "[$current/$totalCombinations] $img  |  $exeName  |  $iniName"
            Write-Host $progressMsg -ForegroundColor Cyan

            # --- Run the tool ---
            $proc = Start-Process -FilePath $exePath `
                -ArgumentList "`"$imgPath`"", "-settings", "`"$iniPath`"", "-preview", "`"$previewPath`"", "-output", "`"$outputPath`"" `
                -NoNewWindow -RedirectStandardOutput $logFile -PassThru -Wait

            if ($proc.ExitCode -ne 0) {
                Write-Host "  FAILED (exit code $($proc.ExitCode))" -ForegroundColor Red
                $summaryRows += [PSCustomObject]@{
                    Image       = $imgBase
                    Exe         = $exeName
                    Settings    = $iniName
                    Status      = "FAILED($($proc.ExitCode))"
                    GPU         = "-"
                    Shapes      = "-"
                    AvgMs       = "-"
                    MinMs       = "-"
                    MaxMs       = "-"
                    TotalS      = "-"
                }
                continue
            }

            # --- Parse results ---
            $gpuName = Get-GpuNameFromLog -logPath $logFile
            $timings = Get-TimingsFromLog -logPath $logFile

            if ($null -eq $timings -or $timings.Count -eq 0) {
                Write-Host "  WARNING: No timing data found" -ForegroundColor Yellow
                $summaryRows += [PSCustomObject]@{
                    Image       = $imgBase
                    Exe         = $exeName
                    Settings    = $iniName
                    Status      = "NO_TIMING"
                    GPU         = $gpuName
                    Shapes      = "-"
                    AvgMs       = "-"
                    MinMs       = "-"
                    MaxMs       = "-"
                    TotalS      = "-"
                }
                continue
            }

            $count = $timings.Count
            $avg   = [math]::Round(($timings | Measure-Object -Average).Average)
            $min   = [math]::Round(($timings | Measure-Object -Minimum).Minimum)
            $max   = [math]::Round(($timings | Measure-Object -Maximum).Maximum)
            $total = [math]::Round(($timings | Measure-Object -Sum).Sum / 1000, 1)

            Write-Host "  OK | shapes=$count | avg=${avg}ms | min=${min}ms | max=${max}ms | total=${total}s | gpu=$gpuName" -ForegroundColor Green

            $summaryRows += [PSCustomObject]@{
                Image       = $imgBase
                Exe         = $exeName
                Settings    = $iniName
                Status      = "OK"
                GPU         = $gpuName
                Shapes      = $count
                AvgMs       = $avg
                MinMs       = $min
                MaxMs       = $max
                TotalS      = $total
            }
        }
    }
}

# -------------------------------------------------------------------
# Write summary CSV
# -------------------------------------------------------------------
$csvPath = Join-Path $runDir 'summary.csv'
$summaryRows | Export-Csv -Path $csvPath -NoTypeInformation -Encoding UTF8

# -------------------------------------------------------------------
# Write human-readable summary grouped by image
# -------------------------------------------------------------------
$txtPath = Join-Path $runDir 'summary.txt'
$txtLines = @()
$txtLines += "Benchmark Run: $ts"
$txtLines += "Generated: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
$txtLines += ""

# Detect all unique GPU names
$gpuNames = ($summaryRows | Where-Object { $_.GPU -ne '-' } | Select-Object -ExpandProperty GPU -Unique)
if ($gpuNames.Count -gt 0) {
    $txtLines += "GPU(s): $($gpuNames -join ', ')"
    $txtLines += ""
}

$grouped = $summaryRows | Group-Object -Property Image
foreach ($group in $grouped) {
    $txtLines += "=" * 70
    $txtLines += "  Image: $($group.Name)"
    $txtLines += "=" * 70
    $txtLines += ""

    # Table header
    $header = "{0,-35} {1,-30} {2,8} {3,8} {4,8} {5,8} {6,8} {7}" -f `
        "Exe / Settings", "GPU", "Shapes", "AvgMs", "MinMs", "MaxMs", "TotalS", "Status"
    $txtLines += $header
    $txtLines += "-" * ($header.Length)

    foreach ($row in $group.Group) {
        $exeSettings = "$($row.Exe) | $($row.Settings)"
        $line = "{0,-35} {1,-30} {2,8} {3,8} {4,8} {5,8} {6,8} {7}" -f `
            $exeSettings, $row.GPU, $row.Shapes, $row.AvgMs, $row.MinMs, $row.MaxMs, $row.TotalS, $row.Status
        $txtLines += $line
    }
    $txtLines += ""
}

$txtLines | Out-File -FilePath $txtPath -Encoding UTF8

# -------------------------------------------------------------------
# Print final summary to console
# -------------------------------------------------------------------
Write-Host ""
Write-Host "==============================================" -ForegroundColor Cyan
Write-Host " BENCHMARK COMPLETE" -ForegroundColor Cyan
Write-Host "==============================================" -ForegroundColor Cyan
Write-Host ""

# Print grouped summary
foreach ($group in $grouped) {
    Write-Host "--- $($group.Name) ---" -ForegroundColor Yellow
    foreach ($row in $group.Group) {
        if ($row.Status -eq 'OK') {
            Write-Host "  $($row.Exe) | $($row.Settings) : avg=$($row.AvgMs)ms shapes=$($row.Shapes) total=$($row.TotalS)s gpu=$($row.GPU)" -ForegroundColor Green
        }
        else {
            Write-Host "  $($row.Exe) | $($row.Settings) : $($row.Status)" -ForegroundColor Red
        }
    }
    Write-Host ""
}

Write-Host "Results: $runDir" -ForegroundColor Green
Write-Host "Summary CSV: $csvPath" -ForegroundColor Green
Write-Host "Summary TXT: $txtPath" -ForegroundColor Green
