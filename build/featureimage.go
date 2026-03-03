package build

import (
	"fmt"
	"os"
	"path/filepath"
)

// SelectFeatureImage returns the URL of the best available feature image for
// a page, based on its front-matter metadata and the images available in
// imageDir (the filesystem directory to check for file existence, typically
// the "images/" subdirectory of the content root).
//
// Returns empty string when no suitable image is found.
//
// Selection cascade (highest to lowest priority):
//  1. fm.Image is set → return it directly (person-specific photo from gen)
//  2. Layout/section-specific image: diary, search, list* pages each have a
//     dedicated image name
//  3. Place category: select by placetype (city, town, village, hamlet,
//     parish) with fallback to generic place image
//  4. Citation/source category: use category-sources image
//  5. Person category: progressive fallback from most-specific to least:
//     person-{gender}-{era}-{trade}-{maturity}.webp → person-{gender}.webp
//  6. Default: default-oak.webp
func SelectFeatureImage(fm FrontMatter, imageDir string) string {
	// 1. Direct image from front-matter (person-specific photo set by gen).
	if fm.Image != "" {
		return fm.Image
	}

	// try returns the URL "/images/<name>" if imageDir/<name> exists as a
	// file, otherwise returns "".
	try := func(name string) string {
		_, err := os.Stat(filepath.Join(imageDir, name))
		if err == nil {
			return "/images/" + name
		}
		return ""
	}

	// 2. Layout / section-specific images.
	switch fm.Layout {
	case "diary":
		if u := try("section-diary.webp"); u != "" {
			return u
		}
	case "search":
		if u := try("section-search.webp"); u != "" {
			return u
		}
	case "listtrees":
		if u := try("section-trees.webp"); u != "" {
			return u
		}
	case "listplaces":
		if u := try("category-place.webp"); u != "" {
			return u
		}
	case "listpeople":
		if u := try("category-people.webp"); u != "" {
			return u
		}
	case "listsurnames":
		if u := try("category-surnames.webp"); u != "" {
			return u
		}
	case "listtodo":
		if u := try("category-todo.webp"); u != "" {
			return u
		}
	case "listsources":
		if u := try("category-sources.webp"); u != "" {
			return u
		}
	}

	// 3–4. Category-specific images.
	switch fm.Category {
	case "place":
		switch fm.PlaceType {
		case "city":
			if u := try("place-city.webp"); u != "" {
				return u
			}
		case "town":
			if u := try("place-town.webp"); u != "" {
				return u
			}
		case "village":
			if u := try("place-village.webp"); u != "" {
				return u
			}
		case "hamlet":
			if u := try("place-hamlet.webp"); u != "" {
				return u
			}
		case "parish":
			// Parish falls back to the village image, matching the Hugo theme.
			if u := try("place-village.webp"); u != "" {
				return u
			}
		default:
			if u := try("category-place.webp"); u != "" {
				return u
			}
		}

	case "citation", "source":
		if u := try("category-sources.webp"); u != "" {
			return u
		}

	case "person":
		g, e, t, m := fm.Gender, fm.Era, fm.Trade, fm.Maturity
		for _, name := range []string{
			fmt.Sprintf("person-%s-%s-%s-%s.webp", g, e, t, m),
			fmt.Sprintf("person-%s-%s-%s.webp", g, e, t),
			fmt.Sprintf("person-%s-%s-%s.webp", g, e, m),
			fmt.Sprintf("person-%s-%s.webp", g, e),
			fmt.Sprintf("person-%s.webp", g),
		} {
			if u := try(name); u != "" {
				return u
			}
		}
	}

	// 6. Final fallback.
	return try("default-oak.webp")
}
