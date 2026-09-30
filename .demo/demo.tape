# slackctl demo — rendered with VHS (https://github.com/charmbracelet/vhs).
#
# NO ACCOUNT, NO SIGNUP, NO KEYRING, NO NETWORK beyond loopback. All data is
# invented by tools/demoserver; tables, JSON, --jq and command execution are real.
# Configuration is isolated from the recorder's own profiles and credentials.
#
# Regenerate:
#   go build -o bin/slackctl ./cmd/slackctl && go build -o bin/demoserver ./tools/demoserver && vhs .demo/demo.tape

Output assets/demo.gif

Set Shell bash
Set FontSize 15
Set Width 1200
Set Height 760
Set Padding 18
Set Theme "Dracula"
Set TypingSpeed 28ms

Hide
Type "export PATH=$PWD/bin:$PATH XDG_CONFIG_HOME=$(mktemp -d) SLACKCTL_TOKEN=xoxb-demo-token-no-account"
Enter
Type "demoserver 127.0.0.1:8645 >/dev/null 2>&1 & demo_pid=$!; trap 'kill $demo_pid 2>/dev/null; rm -rf $XDG_CONFIG_HOME' EXIT; sleep 1"
Enter
Type "alias slackctl='slackctl --base-url http://127.0.0.1:8645 --workspace demo --no-store'; clear"
Enter
Show

Sleep 500ms

Type "slackctl conversations list --limit 2"
Sleep 250ms
Enter
Sleep 2400ms

Type "slackctl conversations history --channel CDEMO001 --limit 2"
Sleep 250ms
Enter
Sleep 2400ms

Type "clear"
Enter
Sleep 300ms

Type "slackctl msg post --channel CDEMO001 --text 'Demo build is green!' -o json"
Sleep 250ms
Enter
Sleep 2400ms

Type "clear"
Enter
Sleep 300ms

Type "slackctl conversations history --channel CDEMO001 --limit 2 -o json --jq '.[].text'"
Sleep 250ms
Enter
Sleep 2400ms
