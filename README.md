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
