Set-ExecutionPolicy Bypass -Scope Process -Force; iex ((New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1'))

choco install -y git hg golang dep ruby msys2 mingw netcat

[string]$source  = 'C:\vagrant'
[string]$destination = 'C:\opt\stackstate-go\src\github.com\StackVista\stackstate-process-agent'
mkdir -Force $destination
Copy-Item -Path (Get-Item -Path "$source\*" -Exclude ('vendor', '.idea', '.circleci') -Force).FullName -Destination $destination -Recurse -Force

[string]$gopath = 'C:\opt\stackstate-go'
setx GOPATH $gopath
setx GO_PROCESS_AGENT $destination
setx path "%path%;$gopath\bin"
setx DEPNOLOCK 1

$profile

if (Get-Command rake -errorAction SilentlyContinue)
{
    rake --version
} else {
    C:\tools\ruby26\bin\gem.cmd install rake
}