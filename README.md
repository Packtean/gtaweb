# [gtaweb](https://zeozeozeo.github.io/gtaweb/)

## [Open in browser!](https://zeozeozeo.github.io/gtaweb/)

> [!NOTE]
> Your adblocker may block the ingame ads, make sure to disable it.

A conversion of Grand Theft Auto's internet to HTML.

For GTA IV, it works by parsing the .whm files and reconstructing the layout into HTML. WHM is quite similar to HTML, except stored in binary together with the DXT-compressed textures and zlib compressed. The DXT textures are decompressed and converted to PNG thanks to the library [mauserzjeh/dxt](https://github.com/mauserzjeh/dxt).

For GTA V, the game uses a completely different system for its UI and web pages called [Scaleform GFx](https://en.wikipedia.org/wiki/Scaleform_GFx), which is essentially a stripped down version of Adobe Flash Player by Autodesk. The .gfx files are converted to Flash .swf's and emulated with [Ruffle](https://ruffle.rs/) (this is mostly still WIP as there's no browser controls or textures).

## Usage

First, install [Go](https://go.dev/) if you haven't already.

**For GTA IV**: To run the HTML converter, you will need to copy the `<GTA IV installation directory>/pc/html` folder into `iv/html` and run `whm2html` (compile with `go build -o whm2html`). The `iv/american.txt` file can be obtained by decompiling the `common/text/american.gxt` file in OpenIV.

**For GTA V**: In OpenIV, locate the `scaleform_web.rpf` archives inside `update/update - Copy.rpf/x64/patch/data/cdimages/scaleform_web.rpf`, `update/update.rpf/x64/patch/data/cdimages/scaleform_web.rpf` and `x64b.rpf/data/cdimages/scaleform_web.rpf`, extract the contents of them into `v/update_copy_scaleform_web/`, `v/update_scaleform_web/` and `v/x64b_scaleform_web/` respectively. Afterwards, locate the `gfxfontlib.gfx` file inside `common.rpf` and extract it into `v/`

## FIXMEs

**GTA IV**:

- Where are the locale strings for www.eyefind.info, www.vipluxuryringtones.com, www.autoeroticar.com, www.littlelacysurprisepageant.com stored (e.g. `CAR_NAV_HOME`, `EYE_HOME_2`)? Can't seem to find them in `american.gxt` maybe i'm just missing something.
- SCO scripts? This is a long shot

**GTA V**:

- Support for textures. Currently they're extracted from the .ytd's, but i have no idea how to inform Ruffle of their existence.
- Browser controls (e.g. scrolling, input, mousedown)
- Fix corrupted textures
- Probably more..

## Resources

- [WHM page on GTAMods](https://web.archive.org/web/20251115012848/https://gtamods.com/wiki/WHM)
- [010 editor types](https://web.archive.org/web/20251115012848/http://public.sannybuilder.com/GTA4/iv_types_20090919.rar) by Stanislav "listener" Golovin
- [SparkIV](https://github.com/ahmed605/SparkIV), RageLib
