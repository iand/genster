# genster

[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/iand/genster)
[![Check Status](https://github.com/iand/genster/actions/workflows/check.yml/badge.svg)](https://github.com/iand/genster/actions/workflows/check.yml)
[![Test Status](https://github.com/iand/genster/actions/workflows/test.yml/badge.svg)](https://github.com/iand/genster/actions/workflows/test.yml)

`genster` generates a static family history website from a GEDCOM or Gramps database. It uses a two-step pipeline: `gen` writes markdown content files, then `build` renders them into a complete HTML site ready to serve or rsync to a server.

I created it for myself so it makes many assumptions about my particular style of using GEDCOM, especially around Ancestry exports.

## Installation

As of Go 1.19, install the latest genster executable using:

	go install github.com/iand/genster@latest

This will download and build a binary in `$GOBIN`.

## Workflow

```
genster gen   --gedcom family.ged --id mytree --site content/
genster build --content content/ --pub pub/
npx serve pub/           # or: rsync pub/ user@host:/var/www/html/
```

**Step 1 — `gen`** reads the genealogy data and writes markdown files with YAML front matter into the content directory. These files represent every person, place, source, family, citation, list page, and chart in the tree. Manual content (the research diary, stories, a home page) lives alongside generated files in the same content directory.

**Step 2 — `build`** walks the content directory, parses each markdown file, renders the body through goldmark, applies an HTML template based on the `layout` front-matter field, and writes complete HTML pages to the pub directory. Non-markdown files (images, media) are copied verbatim.

The pub directory is self-contained and can be deployed directly with `rsync`, `scp`, or any static hosting service.

---

## Commands

### `genster gen` — generate content from genealogy data

Reads a GEDCOM or Gramps file and writes markdown content files to a directory.
Exactly one of `--gedcom` or `--gramps` must be supplied.

| Flag | Short | Description |
|------|-------|-------------|
| `--gedcom <file>` | `-g` | GEDCOM file to read |
| `--gramps <file>` | | Gramps XML file to read |
| `--gramps-dbname <name>` | | Name of the Gramps database, used to keep IDs stable across exports |
| `--id <string>` | | Identifier for this tree (used in URL paths and to locate config) |
| `--site <dir>` | `-s` | Directory to write generated content files into |
| `--basepath <path>` | `-b` | URL path prefix for all links (default `/`) |
| `--key <id>` | `-k` | ID of the key individual; sets the anchor person for relation filtering and ancestor charts |
| `--relation <mode>` | | Filter which people get pages: `any` (default), `common` (must share a common ancestor with key person), or `direct` (must be a direct ancestor) |
| `--include-private` | | Include living people and those who died within the last 20 years (normally redacted) |
| `--config <dir>` | `-c` | Path to the config directory (default: OS-appropriate user config dir) |
| `--wikitree` | | Generate WikiTree markup on person pages for copy-and-paste |
| `--inspect <type/id>` | | Print the internal data structure for one object (e.g. `person/I123`) and exit |
| `--debug` | | Embed debug information as inline HTML comments |
| `--verbose` / `--veryverbose` | | Increase log verbosity |

### `genster build` — render content to HTML

Walks a content directory and renders every markdown file into a complete HTML page.

| Flag | Short | Description |
|------|-------|-------------|
| `--content <dir>` | `-c` | Content directory to read (required) |
| `--pub <dir>` | `-p` | Output directory for rendered HTML (required) |
| `--assets <dir>` | `-a` | Directory of static assets (CSS, JS) to copy into pub; embedded defaults used when not set |
| `--base-url <url>` | | Scheme and host for absolute URLs in `sitemap.xml` (e.g. `https://example.com`); sitemap is omitted when not set |
| `--include-drafts` | | Publish pages marked `draft: true` |
| `--verbose` / `--veryverbose` | | Increase log verbosity |

### `genster chart` — generate a standalone family tree chart

Produces an SVG family tree chart directly from a GEDCOM or Gramps file without generating a full site. Chart types: `descendant`, `ancestor`, `butterfly`, `fan`, `focus`.

### `genster report` — produce a text report

Outputs a plain-text `descendant` or `familyline` report to stdout.

### `genster annotate` — annotate diary markdown files

Walks a directory of hand-authored markdown files and replaces bare footnote references (`[^label]`) with fully-rendered citation links drawn from the genealogy database. Pass `--undo` to strip the generated citations and restore original syntax.

---

## Content directory layout

The content directory holds both generated and hand-authored files. `gen` writes into `trees/<tree-id>/`; manual content sits at the top level alongside it.

```
content/
├── index.md                          # site home page  (layout: home)
├── images/                           # generic feature images — see below
├── diary/                            # hand-authored research diary
│   ├── index.md                      # diary home      (layout: diary)
│   └── 2024/
│       ├── index.md                  # year index      (layout: diary)
│       └── 2024-03-15/
│           └── index.md              # daily entry     (layout: single)
├── stories/                          # hand-authored narrative articles
│   ├── index.md
│   └── my-story/
│       └── index.md                  #                 (layout: single)
└── trees/
    └── <tree-id>/                    # one subtree per --id value
        ├── index.md                  # tree overview   (layout: treeoverview)
        ├── person/<id>/index.md      #                 (layout: person)
        ├── place/<id>/index.md       #                 (layout: place)
        ├── source/<id>/index.md      #                 (layout: source)
        ├── citation/<id>/index.md    #                 (layout: citation)
        ├── family/<id>/index.md      #                 (layout: family)
        ├── chart/<id>/index.md       #                 (layout: chartancestors)
        └── list/
            ├── people/               #                 (layout: listpeople)
            ├── surnames/             #                 (layout: listsurnames)
            ├── places/               #                 (layout: listplaces)
            ├── sources/              #                 (layout: listsources)
            ├── todo/                 #                 (layout: listtodo)
            ├── changes/              #                 (layout: listchanges)
            ├── anomalies/            #                 (layout: listanomalies)
            ├── inferences/           #                 (layout: listinferences)
            ├── families/             #                 (layout: listfamilies)
            └── familylines/          #                 (layout: listfamilylines)
```

The `build` command also generates `/tags/` pages automatically from front-matter tags — do not create these manually.

### Generic feature images

`build` looks for generic images in `content/images/`. These are never embedded in the binary, so you can supply your own. The expected filename conventions are:

| Filename pattern | Used for |
|-----------------|----------|
| `person-{gender}-{era}-{trade}-{maturity}.webp` | Most specific person silhouette |
| `person-{gender}-{era}-{trade}.webp` | Person by gender, era, and trade |
| `person-{gender}-{era}-{maturity}.webp` | Person by gender, era, and maturity |
| `person-{gender}-{era}.webp` | Person by gender and era |
| `person-{gender}.webp` | Person by gender only |
| `place-{placetype}.webp` | Place by type (e.g. `place-village.webp`, `place-parish.webp`) |
| `place-building-{kind}.webp` | Building by kind (e.g. `place-building-church.webp`) |
| `place-building.webp` | Building fallback |
| `place.webp` | Place final fallback |
| `section-diary.webp` | Diary section |
| `section-stories.webp` | Stories section |
| `section-trees.webp` | Trees list page |
| `section-search.webp` | Search page |
| `category-people.webp` | People list |
| `category-surnames.webp` | Surnames list |
| `category-place.webp` | Places list |
| `category-sources.webp` | Sources and citations |
| `category-todo.webp` | To-do list |
| `default-oak.webp` | Final fallback for all other pages |

For person pages, `build` tries the five person patterns from most to least specific and uses the first file it finds. If nothing matches, no image is rendered.

Person-specific photos (actual photographs from Gramps) are set via the `image:` front-matter field by `gen` and take priority over all generic images.

---

## Front-matter fields

Every content file begins with a YAML front-matter block delimited by `---`. `gen` sets these automatically on generated pages; hand-authored files may set any field explicitly.

### Core fields

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Page title, shown in `<h1>` and `<title>` |
| `layout` | string | Template to use (see [Layouts](#layouts) below) |
| `id` | string | Stable identifier for the entity this page represents |
| `summary` | string | Short description used in meta tags and section listings |
| `draft` | bool | When `true`, excluded from build output unless `--include-drafts` is set |
| `lastmod` | string | Last-modified date `YYYY-MM-DD`, shown in the page footer |
| `aliases` | list | Additional URL paths that redirect to this page |
| `tags` | list | Tag names; `build` generates a `/tags/<slug>/` page for each |
| `image` | string | URL of a person-specific photo; takes priority over all generic silhouettes |

### Tree navigation

| Field | Type | Description |
|-------|------|-------------|
| `basepath` | string | Base URL of the tree this page belongs to (e.g. `/trees/mytree/`); drives the per-tree nav bar |

### Pagination (list pages only)

| Field | Type | Description |
|-------|------|-------------|
| `first` | string | Stem of the first page in the sequence |
| `last` | string | Stem of the last page |
| `next` | string | Stem of the next page |
| `prev` | string | Stem of the previous page |

### Person-specific fields

| Field | Type | Values | Description |
|-------|------|--------|-------------|
| `category` | string | `person` | Marks this as a person page |
| `gender` | string | `male`, `female`, `unknown` | Used to select silhouette |
| `era` | string | `1600s` `1700s` `1800s` `1900s` `modern` | Derived from birth/death year |
| `maturity` | string | `child` `young` `mature` `old` | Derived from age at death |
| `trade` | string | `labourer` `miner` `nautical` `crafts` `clerical` `commercial` `military` `service` | Derived from occupation group |
| `ancestor` | bool | | `true` if this person is a direct ancestor of the key person |
| `grampsid` | string | | Gramps handle |
| `slug` | string | | Short alias for diary links (e.g. `john-smith` → `/r/john-smith`) |
| `diarylinks` | list of `{title, link}` | | Research diary entries mentioning this person |
| `links` | list of `{title, link}` | | External links (Ancestry, FindMyPast, etc.) |
| `descendants` | list of `{name, link, detail}` | | Ancestor path entries in the sidebar |
| `wikitreeformat` | string | | WikiTree markup (set when `--wikitree` is used) |

### Place-specific fields

| Field | Type | Description |
|-------|------|-------------|
| `category` | string | `place` |
| `placetype` | string | Type of place: `city`, `town`, `village`, `hamlet`, `parish`, `county`, `country`, `building`, `street`, `address`, etc. |
| `buildingkind` | string | Kind of building when `placetype` is `building`: `church`, `workhouse`, `farm`, `hospital`, etc. |

### Source/citation fields

| Field | Type | Description |
|-------|------|-------------|
| `category` | string | `source` or `citation` |

### Story/diary fields (hand-authored)

| Field | Type | Description |
|-------|------|-------------|
| `author` | string | Author name |
| `started` | string | Date started |
| `updated` | string | Date last updated |
| `status` | string | Completion status (e.g. `draft`, `complete`) |

### Sitemap control

| Field | Type | Description |
|-------|------|-------------|
| `sitemap` | map | Set `{disable: "1"}` to suppress this page from `sitemap.xml` |

---

## Templating system

The `build` command uses Go's `html/template` package. All templates live in `genster/build/templates/` and are embedded into the binary at compile time.

### Template data

Every template receives a `PageData` value as its dot (`.`):

```go
type PageData struct {
    FrontMatter        // all front-matter fields promoted onto dot
    Body  template.HTML // rendered HTML from the markdown body
    Tree  TreeData      // tree-level metadata
}

type TreeData struct {
    Title    string // title of the tree's index page
    BasePath string // base URL path (e.g. /trees/mytree/)
}
```

Because `FrontMatter` is embedded, all front-matter fields are accessible directly — for example `{{.Title}}`, `{{.Category}}`, `{{.Gender}}`. The embedded struct itself is also accessible as `{{.FrontMatter}}` when you need to pass it to a template function.

### Layouts

The `layout` front-matter field selects the template. `gen` sets this on all generated pages; hand-authored files must set it explicitly (or leave it blank for a plain content page with no sidebar).

| Layout | Template | Used for |
|--------|----------|----------|
| `person` | `person.html` | Person pages |
| `place` | `place.html` | Place pages |
| `source` | `source.html` | Source pages |
| `citation` | `citation.html` | Citation pages |
| `family` | `family.html` | Family pages |
| `treeoverview` | `treeoverview.html` | Tree overview/index |
| `chartancestors` | `chartancestors.html` | Ancestor SVG chart |
| `calendar` | `calendar.html` | Monthly event calendar |
| `listpeople` | `listpeople.html` | Alphabetical people list |
| `listsurnames` | `listsurnames.html` | Surnames list |
| `listplaces` | `listplaces.html` | Places list |
| `listsources` | `listsources.html` | Sources list |
| `listtodo` | `listtodo.html` | Research to-do list |
| `listchanges` | `listchanges.html` | Recently updated pages |
| `listanomalies` | `listanomalies.html` | Data anomalies |
| `listinferences` | `listinferences.html` | Inferences |
| `listfamilies` | `listfamilies.html` | Families list |
| `listfamilylines` | `listfamilylines.html` | Family lines list |
| `listtrees` | `listtrees.html` | All trees |
| `home` | `home.html` | Site home page |
| `diary` | `diary.html` | Diary section index and year index pages |
| `single` | `single.html` | Simple content pages (stories, diary entries, tag pages) |
| `list` | `list.html` | Generic paginated list |
| `search` | `search.html` | Search page |
| *(empty)* | `plain.html` | Plain content, no sidebar |

### Shared partials (`base.html`)

Layout templates compose these named blocks via `{{template "name" .}}`:

| Partial | Description |
|---------|-------------|
| `head` | `<head>` element: charset, title, meta description, viewport, CSS links |
| `site-header` | Top nav: home, trees, diary, stories, tags |
| `tree-header` | Like `site-header` plus a per-tree nav row (people, surnames, places, to-do, recent updates) when `.Tree.BasePath` is set |
| `footer` | Page footer with last-modified date when `.LastMod` is set |
| `scripts` | JavaScript includes |
| `featureimage` | Sidebar image: real photo when `.Image` is set; otherwise best available generic image from `content/images/` via cascade; nothing when no match; shows "representative image" caption on person pages with a generic silhouette |
| `tags` | Sidebar tag list linking to `/tags/<slug>/` |
| `pagination` | Previous/next/first/last nav for paginated list pages |

### Template functions

| Function | Description |
|----------|-------------|
| `urlize s` | Lowercase and replace spaces with hyphens — used to build tag URL slugs |
| `ukdate s` | Format `YYYY-MM-DD` as `2 January 2006`; returns input unchanged if unparseable |
| `featureImageSrc fm` | Given a `FrontMatter` value, return the best available feature image URL by checking `content/images/`; returns `""` when nothing matches |

### Customising templates

Override the embedded templates by passing `--assets <dir>` to `build`, where `<dir>` contains a `templates/` subdirectory with replacement `.html` files. Only the files you provide are replaced; all others use the built-in versions.

To add new template functions, add them to `buildSiteTemplates` in `genster/build/template.go` and reference them from a template file.

---

## GEDCOM conventions

Genster understands several Ancestry-specific GEDCOM extensions:

- `_APID` — Ancestry source citation identifier, translated to an Ancestry URL
- `_TREE` — Ancestry tree reference in the GEDCOM header

Custom `EVEN` fact labels with specific handling:

- `Nickname` — preferred nickname for the person
- `OLB` — one-line biography: a short sentence summarising the person's life

Events whose values begin with certain phrases are included verbatim in the generated narrative:

- `He was recorded as`
- `She was recorded as`
- `It was recorded that`

---

## License

This is free and unencumbered software released into the public domain. For more
information, see <http://unlicense.org/> or the accompanying [`UNLICENSE`](UNLICENSE) file.
