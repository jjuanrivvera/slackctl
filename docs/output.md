# Output & filtering

Every command shares one output layer, so these flags work everywhere.

## Formats — `-o` / `--output`

| Format | Use |
|---|---|
| `table` (default) | human-readable; colored on a TTY, honors `NO_COLOR` |
| `json` | full structured payload — pipe to `jq` or a script |
| `yaml` | same data, YAML |
| `csv` | spreadsheet-friendly (cells sanitized against formula injection) |
| `id` | one id per line — pipe straight to `xargs` |

```sh
slackctl conversations list -o json
slackctl conversations members --channel C0123456 -o id | xargs -n1 slackctl users info --user
```

## Pick columns — `--columns`

Table and CSV render a deterministic default set; override it:

```sh
slackctl users list --columns id,name,real_name,is_bot
```

Wide cells are truncated with `…` (a stderr hint suggests `-o json` for the full value).

## Filter with `--jq`

A built-in [gojq](https://github.com/itchyny/gojq) expression runs against the JSON payload
before rendering — no external `jq` needed:

```sh
slackctl conversations list --jq '.[] | select(.is_private) | .name'
slackctl users list --jq '[.[] | select(.is_bot)] | length'
```

## Pagination — `--all` / `--limit`

List commands page automatically over Slack's cursors:

```sh
slackctl conversations history --channel C0123456 --limit 500   # cap the total
slackctl users list --all -o csv > users.csv                    # every page
```

`--limit` caps the total items collected; `--all` fetches every page. (`search` uses Slack's
page/count pagination instead — see `slackctl search messages --help`.)

## Dry run — `--dry-run`

Print the exact `curl` slackctl *would* send (token redacted) and make no request. Great for
debugging or learning the API:

```sh
$ slackctl msg post --channel C0123456 --text "hi" --dry-run
curl -sS -X POST 'https://slack.com/api/chat.postMessage' \
  -H 'Authorization: Bearer xoxb-****' \
  --data-urlencode 'channel=C0123456' --data-urlencode 'text=hi'
```

## The raw escape hatch — `slackctl api`

Call any Web API method slackctl doesn't wrap, with the same auth, output, and `--dry-run`:

```sh
slackctl api conversations.info -q channel=C0123456 --idempotent
slackctl api chat.postMessage -d '{"channel":"C0123456","text":"json body"}'
```

Pass `--idempotent` for read methods so they're sent as `GET` and retried safely on transient
failures.

## Quiet & color

- `--quiet` suppresses the notes slackctl writes to **stderr** (stdout stays pipe-clean).
- `--no-color` (or `NO_COLOR=1`) disables table colors.
