# XEX Play Icon Assets

Place the following files in this directory before running the icon/splash generators:

## Required Files

### `app_icon.png`
- **Size:** 1024x1024 px
- **Format:** PNG, no transparency (iOS requirement)
- **Content:** Full app icon (used for iOS and Android legacy icons)

### `app_icon_foreground.png`
- **Size:** 1024x1024 px
- **Format:** PNG with transparency
- **Content:** Foreground layer only for Android adaptive icons
- **Note:** Keep the logo within the center 66% safe zone (approx 676x676 px centered).
  The outer area may be clipped depending on device mask shape.

### `splash_logo.png`
- **Size:** 1152x1152 px recommended (or at least 768x768)
- **Format:** PNG with transparency
- **Content:** Centered logo for the splash/launch screen
- **Background:** Transparent (the dark background #0A0E1A is set in config)

## Generation Commands

After placing the assets, run from the `app/` directory:

```bash
# Generate app icons
dart run flutter_launcher_icons -f flutter_launcher_icons.yaml

# Generate splash screens
dart run flutter_native_splash:create --path=flutter_native_splash.yaml
```
