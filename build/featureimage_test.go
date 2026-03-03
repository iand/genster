package build

import (
	"os"
	"path/filepath"
	"testing"
)

// touch creates an empty file at path, creating parent dirs as needed.
func touch(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
}

// imageDir creates a temp directory pre-populated with the given image names.
func imageDir(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, n := range names {
		touch(t, filepath.Join(dir, n))
	}
	return dir
}

func TestSelectFeatureImage_DirectImage(t *testing.T) {
	dir := imageDir(t) // no files needed
	fm := FrontMatter{Image: "/trees/I1/media/photo.webp"}
	got := SelectFeatureImage(fm, dir)
	if got != "/trees/I1/media/photo.webp" {
		t.Errorf("got %q, want %q", got, "/trees/I1/media/photo.webp")
	}
}

func TestSelectFeatureImage_DirectImageSkipsExistenceCheck(t *testing.T) {
	// The image field is returned as-is even when the file doesn't exist in
	// imageDir; it's a real photo path set by gen, not a generic image.
	dir := imageDir(t) // empty dir
	fm := FrontMatter{Image: "/trees/I1/media/photo.webp"}
	got := SelectFeatureImage(fm, dir)
	if got != "/trees/I1/media/photo.webp" {
		t.Errorf("got %q, want %q", got, "/trees/I1/media/photo.webp")
	}
}

func TestSelectFeatureImage_LayoutDiary(t *testing.T) {
	dir := imageDir(t, "section-diary.webp")
	fm := FrontMatter{Layout: "diary"}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/section-diary.webp" {
		t.Errorf("got %q, want %q", got, "/images/section-diary.webp")
	}
}

func TestSelectFeatureImage_LayoutSearch(t *testing.T) {
	dir := imageDir(t, "section-search.webp")
	fm := FrontMatter{Layout: "search"}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/section-search.webp" {
		t.Errorf("got %q, want %q", got, "/images/section-search.webp")
	}
}

func TestSelectFeatureImage_LayoutListTrees(t *testing.T) {
	dir := imageDir(t, "section-trees.webp")
	fm := FrontMatter{Layout: "listtrees"}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/section-trees.webp" {
		t.Errorf("got %q, want %q", got, "/images/section-trees.webp")
	}
}

func TestSelectFeatureImage_LayoutListPlaces(t *testing.T) {
	dir := imageDir(t, "category-place.webp")
	fm := FrontMatter{Layout: "listplaces"}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/category-place.webp" {
		t.Errorf("got %q, want %q", got, "/images/category-place.webp")
	}
}

func TestSelectFeatureImage_LayoutListPeople(t *testing.T) {
	dir := imageDir(t, "category-people.webp")
	fm := FrontMatter{Layout: "listpeople"}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/category-people.webp" {
		t.Errorf("got %q, want %q", got, "/images/category-people.webp")
	}
}

func TestSelectFeatureImage_LayoutListSurnames(t *testing.T) {
	dir := imageDir(t, "category-surnames.webp")
	fm := FrontMatter{Layout: "listsurnames"}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/category-surnames.webp" {
		t.Errorf("got %q, want %q", got, "/images/category-surnames.webp")
	}
}

func TestSelectFeatureImage_LayoutListTodo(t *testing.T) {
	dir := imageDir(t, "category-todo.webp")
	fm := FrontMatter{Layout: "listtodo"}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/category-todo.webp" {
		t.Errorf("got %q, want %q", got, "/images/category-todo.webp")
	}
}

func TestSelectFeatureImage_LayoutListSources(t *testing.T) {
	dir := imageDir(t, "category-sources.webp")
	fm := FrontMatter{Layout: "listsources"}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/category-sources.webp" {
		t.Errorf("got %q, want %q", got, "/images/category-sources.webp")
	}
}

func TestSelectFeatureImage_PlaceCity(t *testing.T) {
	dir := imageDir(t, "place-city.webp")
	fm := FrontMatter{Category: "place", PlaceType: "city"}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/place-city.webp" {
		t.Errorf("got %q, want %q", got, "/images/place-city.webp")
	}
}

func TestSelectFeatureImage_PlaceTown(t *testing.T) {
	dir := imageDir(t, "place-town.webp")
	fm := FrontMatter{Category: "place", PlaceType: "town"}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/place-town.webp" {
		t.Errorf("got %q, want %q", got, "/images/place-town.webp")
	}
}

func TestSelectFeatureImage_PlaceVillage(t *testing.T) {
	dir := imageDir(t, "place-village.webp")
	fm := FrontMatter{Category: "place", PlaceType: "village"}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/place-village.webp" {
		t.Errorf("got %q, want %q", got, "/images/place-village.webp")
	}
}

func TestSelectFeatureImage_PlaceHamlet(t *testing.T) {
	dir := imageDir(t, "place-hamlet.webp")
	fm := FrontMatter{Category: "place", PlaceType: "hamlet"}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/place-hamlet.webp" {
		t.Errorf("got %q, want %q", got, "/images/place-hamlet.webp")
	}
}

func TestSelectFeatureImage_PlaceParishUsesVillage(t *testing.T) {
	// Parish falls back to the village image, matching the Hugo theme.
	dir := imageDir(t, "place-village.webp")
	fm := FrontMatter{Category: "place", PlaceType: "parish"}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/place-village.webp" {
		t.Errorf("got %q, want %q", got, "/images/place-village.webp")
	}
}

func TestSelectFeatureImage_PlaceGenericFallback(t *testing.T) {
	// Unknown placetype falls back to category-place.webp.
	dir := imageDir(t, "category-place.webp")
	fm := FrontMatter{Category: "place", PlaceType: "ocean"}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/category-place.webp" {
		t.Errorf("got %q, want %q", got, "/images/category-place.webp")
	}
}

func TestSelectFeatureImage_Citation(t *testing.T) {
	dir := imageDir(t, "category-sources.webp")
	fm := FrontMatter{Category: "citation"}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/category-sources.webp" {
		t.Errorf("got %q, want %q", got, "/images/category-sources.webp")
	}
}

func TestSelectFeatureImage_Source(t *testing.T) {
	dir := imageDir(t, "category-sources.webp")
	fm := FrontMatter{Category: "source"}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/category-sources.webp" {
		t.Errorf("got %q, want %q", got, "/images/category-sources.webp")
	}
}

func TestSelectFeatureImage_PersonMostSpecific(t *testing.T) {
	dir := imageDir(t,
		"person-male-victorian-farmer-mature.webp",
		"person-male-victorian-farmer.webp",
		"person-male-victorian.webp",
		"person-male.webp",
	)
	fm := FrontMatter{
		Category: "person",
		Gender:   "male",
		Era:      "victorian",
		Trade:    "farmer",
		Maturity: "mature",
	}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/person-male-victorian-farmer-mature.webp" {
		t.Errorf("got %q, want %q", got, "/images/person-male-victorian-farmer-mature.webp")
	}
}

func TestSelectFeatureImage_PersonFallsBackWhenMostSpecificMissing(t *testing.T) {
	// Only the era+trade variant is present; the most-specific (with maturity)
	// and the era+maturity variant are absent.
	dir := imageDir(t, "person-male-victorian-farmer.webp")
	fm := FrontMatter{
		Category: "person",
		Gender:   "male",
		Era:      "victorian",
		Trade:    "farmer",
		Maturity: "mature",
	}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/person-male-victorian-farmer.webp" {
		t.Errorf("got %q, want %q", got, "/images/person-male-victorian-farmer.webp")
	}
}

func TestSelectFeatureImage_PersonFallsBackToGender(t *testing.T) {
	// Only the gender-only variant is present.
	dir := imageDir(t, "person-female.webp")
	fm := FrontMatter{
		Category: "person",
		Gender:   "female",
		Era:      "victorian",
		Trade:    "farmer",
		Maturity: "mature",
	}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/person-female.webp" {
		t.Errorf("got %q, want %q", got, "/images/person-female.webp")
	}
}

func TestSelectFeatureImage_PersonNoMatch(t *testing.T) {
	// No person images present; should fall through to default.
	dir := imageDir(t, "default-oak.webp")
	fm := FrontMatter{
		Category: "person",
		Gender:   "male",
		Era:      "victorian",
	}
	got := SelectFeatureImage(fm, dir)
	if got != "/images/default-oak.webp" {
		t.Errorf("got %q, want %q", got, "/images/default-oak.webp")
	}
}

func TestSelectFeatureImage_Default(t *testing.T) {
	dir := imageDir(t, "default-oak.webp")
	fm := FrontMatter{} // no layout, category, or image
	got := SelectFeatureImage(fm, dir)
	if got != "/images/default-oak.webp" {
		t.Errorf("got %q, want %q", got, "/images/default-oak.webp")
	}
}

func TestSelectFeatureImage_NothingAvailable(t *testing.T) {
	dir := imageDir(t) // empty dir
	fm := FrontMatter{}
	got := SelectFeatureImage(fm, dir)
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestSelectFeatureImage_LayoutMissingFile(t *testing.T) {
	// Layout matches but the image file doesn't exist → empty string (no fallback
	// to default since the layout branch consumed the match attempt).
	dir := imageDir(t) // empty dir
	fm := FrontMatter{Layout: "diary"}
	got := SelectFeatureImage(fm, dir)
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}
