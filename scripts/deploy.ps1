param(
    [Parameter(Position = 0)]
    [string]$Cmd = "uname -a",
    [string]$Host_ = "ubuntu@13.49.49.51"
)
# The SSH private key is read from an env var; set once per shell:
#   $env:AWS_SSH_KEY = "C:\Users\AMBE\.ssh\deploy-sv.pem"
$key = if ($env:AWS_SSH_KEY) { $env:AWS_SSH_KEY } else { "C:\Users\AMBE\.ssh\deploy-sv.pem" }

if (-not (Test-Path -LiteralPath $key)) {
    Write-Error "SSH key not found at '$key'. Set `$env:AWS_SSH_KEY to the correct .pem path."
    exit 1
}

ssh -i $key -o ConnectTimeout=15 $Host_ $Cmd
exit $LASTEXITCODE