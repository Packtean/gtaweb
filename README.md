# [gtaweb](https://zeozeozeo.github.io/gtaweb/)

## [Open in browser!](https://zeozeozeo.github.io/gtaweb/)

> [!NOTE]
> Your adblocker may block the ingame ads, make sure to disable it.

A conversion of GTA IV's internet into HTML. It works by parsing the .whm files and reconstructing the layout into HTML. WHM is quite similar to HTML, except stored in binary together with the DXT-compressed textures and zlib compressed. The DXT textures are decompressed and converted to PNG thanks to the library [mauserzjeh/dxt](https://github.com/mauserzjeh/dxt).

## Usage

First, install [Go](https://go.dev/) if you haven't already.

**For GTA IV**: To run the HTML converter, you will need to copy the `<GTA IV installation directory>/pc/html` folder into `iv/html` and run `whm2html` (compile with `go build -o whm2html`). The `iv/american.txt` file can be obtained by decompiling the `common/text/american.gxt` file in OpenIV.

## FIXMEs

- Where are the locale strings for www.eyefind.info, www.vipluxuryringtones.com, www.autoeroticar.com, www.littlelacysurprisepageant.com stored (e.g. `CAR_NAV_HOME`, `EYE_HOME_2`)? Can't seem to find them in `american.gxt` maybe i'm just missing something.
- SCO scripts? This is a long shot

## Resources

- [WHM page on GTAMods](https://web.archive.org/web/20251115012848/https://gtamods.com/wiki/WHM)
- [010 editor types](https://web.archive.org/web/20251115012848/http://public.sannybuilder.com/GTA4/iv_types_20090919.rar) by Stanislav "listener" Golovin
- [SparkIV](https://github.com/ahmed605/SparkIV), RageLib
