# tml-gen

Template Library Generator for TerrainBuilder (DayZ/Arma 3).  
`*.tml` is a Template Library file used by TerrainBuilder to import map objects.

This tool crawls your `P:/` (or any game root) and
generates libraries automatically so you can update them fast and consistently.
It was tested on DayZ.

## Features

<!-- markdownlint-disable-next-line MD033 -->
<img src="library.png" alt="ImageSet Packer" align="right" width="50%">

* Scans one or multiple `--path` roots
  (absolute or relative to `--game-root`)
* Groups directories by object count threshold
* Skips whole subtrees via `--skip`
  (supports prefix match)
* Preserves original model paths and casing in `<File>`
* Makes `<Name>` unique across all libraries
  (case-insensitive)
* Auto-assigns colors by library type;
  unknown groups use a hash color
* Library `shape` is set by group type
  (nature = ellipse, structures/roads = rectangle)

<!-- markdownlint-disable-next-line MD033 -->
## Usage <br clear="right"/>

```powershell
tml-gen.exe -g P:\ -p dz -o out -n 75 -f
```

```shell
./tml-gen -g /home/user/p_drive/ -p dz -o out -n 75 -f
```

## Options

Execute `tml-gen --help` to show all available options.

* `-g, --game-root` (required):
  absolute path to base game root, e.g. `P:\`
* `-p, --path` (repeatable, required):
  path(s) to scan inside game-root or absolute
* `-s, --skip` (repeatable): skip prefixes after normalization
  * Defaults to: `characters`, `vehicles`, `weapons`, `animals`, `gear`, `data`
  * Matches by prefix: `animals` will also skip `animals_bliss`, `animals/...`
* `-n, --threshold`: minimum objects per library (default `75`)
* `-o, --out`: output dir (default `out`)
* `-f, --force`: delete output directory before writing

## Grouping rules (Threshold)

Files are grouped by directory nodes.
If a node has fewer than `--threshold` objects,
the generator climbs up _until_ the second level (e.g. `dz/worlds`),
but never to the top-level `dz` group.
Files directly in `GameRoot` are placed in the root group only.

## Name uniqueness

`<Name>` must be unique across **all** libraries. The generator:

1. Keeps the base name if unique.
2. Otherwise tries to add a logical suffix:
   * If the first path segment is not `dz`, append `_<root>`.
     For example, the file `my_world/Ruin_Wall.p3d` has the model name
     `Ruin_Wall.p3d` as in the original game, then it will be registered as
     `Ruin_Wall_my_world`
   * Else if the path contains:
     `wrecks`, `ruins`, `bliss`, `sakhal`, `proxy`, `military`, `furniture`,
     `residential`, `industrial` append that tag.
3. If still not unique, appends `_N` (starting from `1`).

Matching is case-insensitive, but the emitted name keeps original casing.

## Colors and shapes

Library header (`<Library ...>`) gets:

* `default_fill` / `default_outline` based on library name
* `shape` based on type:
  * `plants`, `rocks` → `ellipse`
  * everything else → `rectangle`

Type colors are hardcoded (water = blue, industrial = yellow/brown, etc.).
Unknown types use a stable hash color.
