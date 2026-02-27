## Update 2026: Project revival

This project was unmaintained for several years and is now being cleaned up and modernized.

## About

Icons Refresh is a program written in [Go](https://golang.org/) to refresh Desktop, Start Menu and Taskbar icons
without restart Explorer on Windows.

## Download

TBD

### Binary compatibility

Current release binaries target `windows/amd64` only.

- Supported: Windows 11 on x64 (Intel/AMD) machines
- Not natively supported: Windows 11 ARM64 (unless x64 emulation is available)

## Usage

```text
IconsRefresh.exe [--dry-run] [--json] <quick|soft|standard|deep>
IconsRefreshUI.exe
```

`IconsRefreshUI.exe` is an interactive UI app; it does not take command-line mode flags.

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

CLI flags (`IconsRefresh.exe`):
- `--dry-run`: print planned actions without deleting cache files
- `--json`: emit a machine-readable execution report

## Build

> Go 1.23.8 or higher required

```
go mod download
mage build
```



## License

MIT. See `LICENSE` for more details.<br />
Icon credit to Oliver Scholtz
