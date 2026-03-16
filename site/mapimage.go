package site

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png" // register PNG decoder for OSM tile fetching
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/iand/genster/logging"
	"github.com/iand/genster/model"
)

// Map tile size in pixels (standard for XYZ tile services).
const mapTileSize = 256

// Output image dimensions for downloaded place maps.
const (
	mapOutputWidth  = 800
	mapOutputHeight = 600
)

// NLS Maps tile layer IDs hosted on MapTiler Cloud.
const (
	// mapLayerSixInch is the OS six-inch to the mile survey, 1888–1913.
	// Use for streets and buildings.
	mapLayerSixInch = "uk-osgb10k1888"

	// mapLayerOneInch is the OS one-inch Hills edition, 1885–1903.
	// Use for parishes, towns and counties.
	mapLayerOneInch = "uk-osgb63k1885"

	// mapLayerLondon is the OS five-foot to the mile London survey, 1893–1896.
	// Use for any place within Greater London.
	mapLayerLondon = "uk-oslondon1k1893"

	// mapLayerIreland is the Bartholomew quarter-inch Ireland map, 1940.
	// Use for places in the Republic of Ireland or Northern Ireland.
	mapLayerIreland = "uk-baire250k1940"

	// mapLayerOSM is the MapTiler OpenStreetMap raster style.
	// Used as a fallback for places outside NLS coverage.
	mapLayerOSM = "openstreetmap"
)

// nlsLayerID is the layer number used in the NLS geo/explore interface for
// each MapTiler tile layer.
var nlsLayerID = map[string]int{
	mapLayerSixInch: 6,
	mapLayerOneInch: 161,
	mapLayerLondon:  163,
	mapLayerIreland: 13,
}

// mapLayerName is the human-readable name for each tile layer.
var mapLayerName = map[string]string{
	mapLayerSixInch: "Ordnance Survey six-inch to the mile, 1888–1913",
	mapLayerOneInch: "Ordnance Survey one-inch Hills edition, 1885–1903",
	mapLayerLondon:  "Ordnance Survey five-foot to the mile (London), 1893–1896",
	mapLayerIreland: "Bartholomew quarter-inch to the mile (Ireland), 1940",
	mapLayerOSM:     "OpenStreetMap",
}

// PlaceMap holds the results of a downloaded place map image.
type PlaceMap struct {
	ImageLink   string // URL of the stitched map image file
	ExploreURL  string // link to an interactive map viewer for this location
	ExploreText string // anchor text for ExploreURL
	MapName     string // human-readable name of the map layer used
}

// placeMapConfig holds the tile layer and zoom level to use for a place.
type placeMapConfig struct {
	layer string
	zoom  int
}

// mapConfigForPlace returns the appropriate tile layer and zoom level for p.
// Returns false when the place has no geo-coordinates or falls outside the
// coverage of the available NLS map layers (UK and Ireland only).
func mapConfigForPlace(p *model.Place) (placeMapConfig, bool) {
	if p.GeoLocation == nil {
		return placeMapConfig{}, false
	}

	// Ireland (Republic or Northern Ireland) uses the Bartholomew map.
	if p.Country != nil && isIrelandCountry(p.Country.Name) {
		return placeMapConfig{layer: mapLayerIreland, zoom: irishZoom(p.PlaceType)}, true
	}

	// Remaining NLS layers cover Great Britain only; fall back to OSM elsewhere.
	if !isGBPlace(p) {
		return placeMapConfig{layer: mapLayerOSM, zoom: osmZoom(p.PlaceType)}, true
	}

	// London uses the OS five-foot map.
	if isLondonPlace(p) {
		return placeMapConfig{layer: mapLayerLondon, zoom: londonZoom(p.PlaceType)}, true
	}

	// Small-scale UK places use the six-inch map.
	switch p.PlaceType {
	case model.PlaceTypeBuilding, model.PlaceTypeAddress,
		model.PlaceTypeStreet, model.PlaceTypeBurialGround,
		model.PlaceTypeHamlet, model.PlaceTypeVillage,
		model.PlaceTypeParish, model.PlaceTypeRegistrationDistrict,
		model.PlaceTypeProbateOffice:
		return placeMapConfig{layer: mapLayerSixInch, zoom: sixInchZoom(p.PlaceType)}, true
	}

	// Larger UK places use the one-inch map.
	return placeMapConfig{layer: mapLayerOneInch, zoom: oneInchZoom(p.PlaceType)}, true
}

func isIrelandCountry(name string) bool {
	switch name {
	case "Ireland", "Republic of Ireland", "Northern Ireland":
		return true
	}
	return false
}

// isGBPlace reports whether p is in Great Britain (England, Scotland or Wales).
func isGBPlace(p *model.Place) bool {
	// A UKNationName is set for places whose hierarchy includes an English,
	// Scottish or Welsh nation place.
	if p.UKNationName != nil {
		return true
	}
	if p.Country != nil {
		switch p.Country.Name {
		case "England", "Scotland", "Wales", "Great Britain", "United Kingdom":
			return true
		}
	}
	return false
}

// isLondonPlace reports whether p or any ancestor is within Greater London.
func isLondonPlace(p *model.Place) bool {
	for cur := p; cur != nil; cur = cur.Parent {
		n := strings.ToLower(cur.Name)
		if n == "london" || n == "greater london" || n == "city of london" {
			return true
		}
	}
	return false
}

// sixInchZoom returns a zoom level appropriate for the OS six-inch layer.
func sixInchZoom(pt model.PlaceType) int {
	switch pt {
	case model.PlaceTypeBuilding, model.PlaceTypeAddress:
		return 15
	case model.PlaceTypeStreet, model.PlaceTypeBurialGround,
		model.PlaceTypeHamlet, model.PlaceTypeProbateOffice:
		return 14
	default: // village, parish, registration district
		return 13
	}
}

// oneInchZoom returns a zoom level appropriate for the OS one-inch Hills layer.
func oneInchZoom(pt model.PlaceType) int {
	switch pt {
	case model.PlaceTypeTown:
		return 13
	case model.PlaceTypeCity:
		return 12
	case model.PlaceTypeCounty:
		return 10
	case model.PlaceTypeCountry, model.PlaceTypeNation:
		return 8
	default:
		return 11
	}
}

// londonZoom returns a zoom level appropriate for the OS London five-foot layer.
func londonZoom(pt model.PlaceType) int {
	switch pt {
	case model.PlaceTypeBuilding, model.PlaceTypeAddress:
		return 17
	case model.PlaceTypeStreet, model.PlaceTypeBurialGround:
		return 16
	default:
		return 15
	}
}

// osmZoom returns a zoom level appropriate for the OpenStreetMap fallback layer.
func osmZoom(pt model.PlaceType) int {
	switch pt {
	case model.PlaceTypeBuilding, model.PlaceTypeAddress:
		return 16
	case model.PlaceTypeStreet, model.PlaceTypeBurialGround:
		return 16
	case model.PlaceTypeHamlet, model.PlaceTypeVillage, model.PlaceTypeParish:
		return 13
	case model.PlaceTypeTown:
		return 11
	case model.PlaceTypeCity:
		return 10
	case model.PlaceTypeCounty:
		return 8
	case model.PlaceTypeCountry, model.PlaceTypeNation:
		return 6
	default:
		return 13
	}
}

// irishZoom returns a zoom level appropriate for the Bartholomew Ireland layer.
func irishZoom(pt model.PlaceType) int {
	switch pt {
	case model.PlaceTypeCounty:
		return 9
	case model.PlaceTypeTown, model.PlaceTypeCity:
		return 11
	default:
		return 10
	}
}

// latLonToTileXY converts geographic coordinates to fractional tile XY
// coordinates at the given zoom level using the standard Web Mercator
// (Slippy Map) projection.
func latLonToTileXY(lat, lon float64, zoom int) (float64, float64) {
	n := math.Exp2(float64(zoom))
	x := (lon + 180.0) / 360.0 * n
	latRad := lat * math.Pi / 180.0
	y := (1.0 - math.Log(math.Tan(latRad)+1.0/math.Cos(latRad))/math.Pi) / 2.0 * n
	return x, y
}

// mapTileCacheDir returns the directory used to cache downloaded tiles.
// It follows the XDG Base Directory Specification, defaulting to
// ~/.cache/genster/maptiles when $XDG_CACHE_HOME is not set.
func mapTileCacheDir() string {
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		base = filepath.Join(home, ".cache")
	}
	return filepath.Join(base, "genster", "maptiles")
}

// fetchTile returns the image for a single XYZ tile, using a disk cache to
// avoid redundant network requests.  cacheDir is the root of the tile cache;
// if empty the tile is fetched directly without caching.
func fetchTile(cacheDir, apiKey, layer string, zoom, tx, ty int) (image.Image, error) {
	// Check the cache first.
	if cacheDir != "" {
		cachePath := filepath.Join(cacheDir, layer, fmt.Sprintf("%d", zoom),
			fmt.Sprintf("%d", tx), fmt.Sprintf("%d.jpg", ty))
		if f, err := os.Open(cachePath); err == nil {
			img, _, err := image.Decode(f)
			f.Close()
			if err == nil {
				return img, nil
			}
		}

		// Not cached: download, save to cache, then return.
		img, body, err := fetchTileFromNetwork(apiKey, layer, zoom, tx, ty)
		if err != nil {
			return nil, err
		}
		if err := saveTileToCache(cachePath, body); err != nil {
			logging.Warn("map tile cache write failed", "path", cachePath, "err", err)
		}
		return img, nil
	}

	img, _, err := fetchTileFromNetwork(apiKey, layer, zoom, tx, ty)
	return img, err
}

// fetchTileFromNetwork downloads a single tile and returns the decoded image
// along with the raw bytes for caching.  NLS layers use the /tiles/ endpoint
// (JPEG); the OSM fallback uses the /maps/ endpoint (PNG).
func fetchTileFromNetwork(apiKey, layer string, zoom, tx, ty int) (image.Image, []byte, error) {
	var url string
	if layer == mapLayerOSM {
		url = fmt.Sprintf("https://api.maptiler.com/maps/%s/%d/%d/%d.png?key=%s",
			layer, zoom, tx, ty, apiKey)
	} else {
		url = fmt.Sprintf("https://api.maptiler.com/tiles/%s/%d/%d/%d.jpg?key=%s",
			layer, zoom, tx, ty, apiKey)
	}

	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return nil, nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read body: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, nil, fmt.Errorf("decode image: %w", err)
	}
	return img, data, nil
}

// saveTileToCache writes raw tile bytes to cachePath, creating parent
// directories as needed.
func saveTileToCache(cachePath string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(cachePath, data, 0o644)
}

// fillArrow draws a downward-pointing filled arrow onto img with its tip at
// (cx, cy).  The outline is drawn first in white, then the fill in red, so
// the marker is visible against both light and dark map backgrounds.
func fillArrow(img *image.RGBA, cx, cy int) {
	const (
		halfBase = 10 // half-width at the widest point
		arrowH   = 22 // total height from tip to flat top
		outline  = 2  // white border thickness
	)

	// fillTri fills a triangle defined by a flat top edge and a bottom tip
	// using scanline rasterisation.
	fillTri := func(tipX, tipY, topLeft, topRight, topY int, c color.RGBA) {
		for y := topY; y <= tipY; y++ {
			t := float64(y-topY) / float64(tipY-topY)
			xL := int(math.Round(float64(topLeft) + t*float64(tipX-topLeft)))
			xR := int(math.Round(float64(topRight) + t*float64(tipX-topRight)))
			for x := xL; x <= xR; x++ {
				img.SetRGBA(x, y, c)
			}
		}
	}

	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	red := color.RGBA{R: 220, G: 50, B: 50, A: 255}

	topY := cy - arrowH
	fillTri(cx, cy, cx-(halfBase+outline), cx+(halfBase+outline), topY-outline, white)
	fillTri(cx, cy, cx-halfBase, cx+halfBase, topY, red)
}

// downloadMapImage fetches and stitches the tiles needed to produce a static
// map image of dimensions width×height centred on (lat, lon) at zoom.
func downloadMapImage(cacheDir, apiKey, layer string, lat, lon float64, zoom, width, height int) (image.Image, error) {
	// Fractional tile coordinates of the centre point.
	cx, cy := latLonToTileXY(lat, lon, zoom)

	// Use the standard 256 px tile size to estimate the tile range to fetch.
	// Some tile endpoints (e.g. MapTiler's /maps/ OSM layer) return 512 px
	// tiles, so we may over-request by one tile on each edge, but we correct
	// all pixel geometry once we know the real size.
	const assumed = mapTileSize
	tlpx0 := cx*assumed - float64(width)/2
	tlpy0 := cy*assumed - float64(height)/2
	txMin0 := int(math.Floor(tlpx0 / assumed))
	tyMin0 := int(math.Floor(tlpy0 / assumed))
	txMax0 := int(math.Floor((tlpx0 + float64(width) - 1) / assumed))
	tyMax0 := int(math.Floor((tlpy0 + float64(height) - 1) / assumed))

	// Fetch all tiles, detecting the actual pixel size from the first one.
	type tileEntry struct {
		tx, ty int
		img    image.Image
	}
	var fetched []tileEntry
	tileSize := 0
	for tx := txMin0; tx <= txMax0; tx++ {
		for ty := tyMin0; ty <= tyMax0; ty++ {
			tile, err := fetchTile(cacheDir, apiKey, layer, zoom, tx, ty)
			if err != nil {
				logging.Warn("map tile unavailable", "layer", layer, "z", zoom, "x", tx, "y", ty, "err", err)
				continue
			}
			if tileSize == 0 {
				tileSize = tile.Bounds().Dx()
			}
			fetched = append(fetched, tileEntry{tx, ty, tile})
		}
	}
	if tileSize == 0 {
		tileSize = assumed // no tiles fetched; use fallback
	}

	// Recalculate geometry using the actual tile pixel size.
	tlpx := cx*float64(tileSize) - float64(width)/2
	tlpy := cy*float64(tileSize) - float64(height)/2
	txMin := int(math.Floor(tlpx / float64(tileSize)))
	tyMin := int(math.Floor(tlpy / float64(tileSize)))
	txMax := int(math.Floor((tlpx + float64(width) - 1) / float64(tileSize)))
	tyMax := int(math.Floor((tlpy + float64(height) - 1) / float64(tileSize)))

	// Create a canvas large enough to hold all required tiles.
	canvasW := (txMax - txMin + 1) * tileSize
	canvasH := (tyMax - tyMin + 1) * tileSize
	canvas := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))

	// Draw fetched tiles onto the canvas.
	for _, te := range fetched {
		if te.tx < txMin || te.tx > txMax || te.ty < tyMin || te.ty > tyMax {
			continue // tile outside the area needed at actual tile size
		}
		dstX := (te.tx - txMin) * tileSize
		dstY := (te.ty - tyMin) * tileSize
		draw.Draw(canvas,
			image.Rect(dstX, dstY, dstX+tileSize, dstY+tileSize),
			te.img, image.Point{}, draw.Src)
	}

	// Crop the canvas to the desired output size, centred on the target
	// location.
	offX := int(math.Round(tlpx)) - txMin*tileSize
	offY := int(math.Round(tlpy)) - tyMin*tileSize

	// Draw a marker at the centre point (the place's coordinates).
	fillArrow(canvas, offX+width/2, offY+height/2)

	return canvas.SubImage(image.Rect(offX, offY, offX+width, offY+height)), nil
}

// exploreLink returns the URL and anchor text for the interactive map explorer
// appropriate for the given layer and location.
func exploreLink(layer string, lat, lon float64, zoom int) (url, text string) {
	if layer == mapLayerOSM {
		return fmt.Sprintf("https://www.openstreetmap.org/#map=%d/%f/%f", zoom, lat, lon),
			"explore on OpenStreetMap"
	}
	return fmt.Sprintf("https://maps.nls.uk/geo/explore/#zoom=%d&lat=%f&lon=%f&layers=%d&b=osm&o=100&marker=%f,%f",
			zoom, lat, lon, nlsLayerID[layer], lat, lon),
		"explore at the National Library of Scotland"
}

// mapImageCacheDir returns the directory used to cache stitched place map
// images, following the XDG Base Directory Specification.
func mapImageCacheDir() string {
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		base = filepath.Join(home, ".cache")
	}
	return filepath.Join(base, "genster", "maps")
}

// WritePlaceMapImage produces a historic map image for p, saves it as a JPEG
// in the media subdirectory of contentDir, and returns a PlaceMap with the
// image URL, NLS explore link, and map name.  Returns (nil, nil) when the
// place has no coordinates or the site has no MapTiler API key configured.
//
// The stitched image is cached in $XDG_CACHE_HOME/genster/maps/ so subsequent
// runs copy from cache rather than re-downloading and re-stitching tiles.
func (s *Site) WritePlaceMapImage(p *model.Place, contentDir string) (*PlaceMap, error) {
	if s.MapTilerAPIKey == "" || p.GeoLocation == nil {
		return nil, nil
	}

	cfg, ok := mapConfigForPlace(p)
	if !ok {
		return nil, nil
	}

	fname := "place-map-" + p.ID + ".jpg"
	destPath := filepath.Join(contentDir, s.MediaDir, fname)
	cachePath := filepath.Join(mapImageCacheDir(), fname)

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return nil, fmt.Errorf("create media dir: %w", err)
	}

	// Use the cached image if available, otherwise stitch and cache it.
	if _, err := os.Stat(cachePath); err != nil {
		if err := s.buildAndCachePlaceMapImage(cfg, p, cachePath); err != nil {
			return nil, err
		}
	}

	if err := CopyFile(destPath, cachePath); err != nil {
		return nil, fmt.Errorf("copy map image to media: %w", err)
	}

	eURL, eText := exploreLink(cfg.layer, p.GeoLocation.Latitude, p.GeoLocation.Longitude, cfg.zoom)
	return &PlaceMap{
		ImageLink:   fmt.Sprintf(s.MediaLinkPattern, fname),
		ExploreURL:  eURL,
		ExploreText: eText,
		MapName:     mapLayerName[cfg.layer],
	}, nil
}

// buildAndCachePlaceMapImage stitches tiles into a map image for p and saves
// it to cachePath.
func (s *Site) buildAndCachePlaceMapImage(cfg placeMapConfig, p *model.Place, cachePath string) error {
	tileCache := mapTileCacheDir()
	img, err := downloadMapImage(tileCache, s.MapTilerAPIKey, cfg.layer,
		p.GeoLocation.Latitude, p.GeoLocation.Longitude,
		cfg.zoom, mapOutputWidth, mapOutputHeight)
	if err != nil {
		return fmt.Errorf("download map image: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return fmt.Errorf("create map cache dir: %w", err)
	}

	f, err := os.Create(cachePath)
	if err != nil {
		return fmt.Errorf("create cached map image: %w", err)
	}
	defer f.Close()

	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 85}); err != nil {
		return fmt.Errorf("encode map image: %w", err)
	}
	return nil
}
