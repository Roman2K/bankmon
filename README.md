# bankmon

Bank account monitoring daemon

<img src="https://raw.githubusercontent.com/Roman2K/bankmon/master/screenshot.png"
	alt="Screenshot" width="500" height="435"/>

###### FEATURES

* Log activity to Slack
  * movements (balance deltas) throughout the day
  * daily summary
  * logs
* Supported banks:
  * BBVA Net Cash (enterprise)
  * (upcoming: Boursorama)
* Docker support

## Build

1. Clone to a `bankmon` directory in your GOPATH
2. Install dependencies with `glide install`
3. Install additional dependencies: see `Dockerfile`
4. Build with `go build`

## Run

```sh
$ ./bankmon
Usage: bankmon [<flag> ...] <config-path>
Flags:
  -cron
        cron mode
  -log string
        log path (default "-")
```

## Test

Run tests with `tools/test`
