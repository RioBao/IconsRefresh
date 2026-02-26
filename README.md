## Update 2026: Trying to revive this project

This project stopped being maintained many years ago. With the help of AI I want to review it, since its really helpful.

## About

Icons Refresh is a program written in [Go](https://golang.org/) to refresh Desktop, Start Menu and Taskbar icons
without restart Explorer on Windows.

## Download

TBD

## Usage

```text
IconsRefresh.exe [--dry-run] [--json] <quick|soft|standard|deep>
IconsRefresh-tray.exe [--preset quick|soft|standard|deep] [--dry-run] [--json]
```

Modes:

| Mode | ie4uinit | Shell notify | IconCache.db | Explorer iconcache_*.db | AppIconCache |
|------|----------|--------------|--------------|-------------------------|--------------|
| `quick` | ✓ | ✓ | ✓ | | |
| `soft` | | ✓ | ✓ | | |
| `standard` | ✓ | ✓ | ✓ | ✓ | |
| `deep` | ✓ | ✓ | ✓ | ✓ | ✓ |

Windows 11 cache targets:
- `%LocalAppData%\IconCache.db`
- `%LocalAppData%\Microsoft\Windows\Explorer\iconcache_*.db`
- `%LocalAppData%\Packages\Microsoft.Windows.Search_*\LocalState\AppIconCache`

Flags:
- `--dry-run`: print planned actions without deleting cache files
- `--json`: emit a machine-readable execution report

## Build

> Go 1.22 or higher required

```
go mod download
mage build
```



## License

MIT. See `LICENSE` for more details.<br />
Icon credit to Oliver Scholtz
