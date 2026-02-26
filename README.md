<p align="center"><img width="100" src="https://raw.githubusercontent.com/crazy-max/IconsRefresh/master/.github/logo.png"></p>

<p align="center">
  <a href="https://github.com/crazy-max/IconsRefresh/releases/latest"><img src="https://img.shields.io/github/release/crazy-max/IconsRefresh.svg?style=flat-square" alt="GitHub release"></a>
  <a href="https://github.com/crazy-max/IconsRefresh/releases/latest"><img src="https://img.shields.io/github/downloads/crazy-max/IconsRefresh/total.svg?style=flat-square" alt="Total downloads"></a>
  <a href="https://github.com/crazy-max/IconsRefresh/actions"><img src="https://github.com/crazy-max/IconsRefresh/workflows/build/badge.svg" alt="Build Status"></a>
  <a href="https://goreportcard.com/report/github.com/crazy-max/IconsRefresh"><img src="https://goreportcard.com/badge/github.com/crazy-max/IconsRefresh?style=flat-square" alt="Go Report"></a>
  <a href="https://www.codacy.com/app/crazy-max/IconsRefresh"><img src="https://img.shields.io/codacy/grade/834f62e0849c4c008dd8df69b816d2a0/master.svg?style=flat-square" alt="Code Quality"></a>
  <br /><a href="https://github.com/sponsors/crazy-max"><img src="https://img.shields.io/badge/sponsor-crazy--max-181717.svg?logo=github&style=flat-square" alt="Become a sponsor"></a>
  <a href="https://www.paypal.me/crazyws"><img src="https://img.shields.io/badge/donate-paypal-00457c.svg?logo=paypal&style=flat-square" alt="Donate Paypal"></a>
</p>
## Update 2026: Trying to revive this project

This project stopped being maintained many years ago. With the help of AI I want to review it, since its really helpful.

## About

Icons Refresh is a program written in [Go](https://golang.org/) to refresh Desktop, Start Menu and Taskbar icons
without restart Explorer on Windows.

## Download

TBD

## Usage

```text
IconsRefresh.exe [--dry-run] [--json] <quick|standard|deep>
IconsRefresh-tray.exe [--preset quick|standard|deep] [--dry-run] [--json]
```

- CLI presets map to the shared `internal/engine` orchestration API.
- Tray presets call the same engine API with `tray-*` trigger metadata.
- Event hooks are available for future monitor automation (resolution changes, shell restarts).
- `--dry-run`: print planned actions without deleting cache files
- `--json`: emit a machine-readable execution report

## Build
TBD 


## License

MIT. See `LICENSE` for more details.<br />
Icon credit to Oliver Scholtz
