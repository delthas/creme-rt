# creme-rt token

A trivially small tool to generate an access token for an app to access an account.

Usage:
- run `creme-token -key '<consumer_key>' -secret '<consumer_secret>' -callback '<callback_url>'`
- an URL will be printed, open it in your browser while connected as the desired account, and accept the dialog
- your browser will be redirected, enter the redirect URL in the program (stdin)
- `creme-token` will output the access token and access token secret

Building:
- `go install github.com/delthas/creme-rt/creme-token`

| OS | URL |
|---|---|
| Linux x64 | https://delthas.fr/creme-rt/linux/creme-token |
| Mac OS X x64 | https://delthas.fr/creme-rt/mac/creme-token |
| Windows x64 | https://delthas.fr/creme-rt/windows/creme-token.exe |
