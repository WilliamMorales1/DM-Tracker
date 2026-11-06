// plot renders a GameMap's MapMarkers as numbered circles on top of the
// map's image, so their (x, y) coordinates can be sanity-checked against
// the actual art instead of guessed blind.
//
// Usage:
//
//	go run ./cmd/plot                     # plot every map in GameMaps
//	go run ./cmd/plot -map blingdenstone
//	go run ./cmd/plot -map blingdenstone -out check.png
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/webp"

	"tracker/internal/db"
)

type marker struct {
	i     int
	x, y  int
	title string
}

func main() {
	dbPath := flag.String("db", "tracker.db", "path to the sqlite database")
	mapID := flag.String("map", "", "map_id to plot (default: every map in GameMaps)")
	out := flag.String("out", "", "output PNG path (default <map_id>.png; ignored when plotting more than one map)")
	flag.Parse()

	conn, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	mapIDs := []string{*mapID}
	if *mapID == "" {
		mapIDs, err = allMapIDs(conn)
		if err != nil {
			log.Fatalf("list maps: %v", err)
		}
		if len(mapIDs) == 0 {
			log.Fatal("no maps found in GameMaps")
		}
	}

	var outputs []string
	for _, id := range mapIDs {
		outPath := *out
		if outPath == "" || len(mapIDs) > 1 {
			outPath = id + ".png"
		}
		if err := plotMap(conn, id, outPath); err != nil {
			log.Printf("plot %q: %v", id, err)
			continue
		}
		outputs = append(outputs, outPath)
	}

	if len(outputs) != len(mapIDs) {
		os.Exit(1)
	}
}

func allMapIDs(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT id FROM GameMaps ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func plotMap(db *sql.DB, mapID, out string) error {
	var imgPath string
	if err := db.QueryRow(`SELECT img FROM GameMaps WHERE id = ?`, mapID).Scan(&imgPath); err != nil {
		return fmt.Errorf("lookup map %q: %w", mapID, err)
	}

	rows, err := db.Query(`SELECT i, x, y, title FROM MapMarkers WHERE map_id = ? ORDER BY i`, mapID)
	if err != nil {
		return fmt.Errorf("query markers: %w", err)
	}
	defer rows.Close()

	var markers []marker
	for rows.Next() {
		var m marker
		if err := rows.Scan(&m.i, &m.x, &m.y, &m.title); err != nil {
			return fmt.Errorf("scan marker: %w", err)
		}
		markers = append(markers, m)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate markers: %w", err)
	}
	if len(markers) == 0 {
		return fmt.Errorf("no markers found for map %q", mapID)
	}

	resolved := imgPath
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join("images", filepath.Base(imgPath))
	}
	src, err := loadImage(resolved)
	if err != nil {
		return fmt.Errorf("load image %q: %w", resolved, err)
	}

	canvas := image.NewRGBA(src.Bounds())
	draw.Draw(canvas, canvas.Bounds(), src, src.Bounds().Min, draw.Src)

	for _, m := range markers {
		drawMarker(canvas, m)
	}

	final := image.Image(canvas)
	final = downscale(final, 2000)

	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("create %q: %w", out, err)
	}
	defer f.Close()
	if err := png.Encode(f, final); err != nil {
		return fmt.Errorf("encode png: %w", err)
	}
	fmt.Printf("plotted %d markers for %q -> %s\n", len(markers), mapID, out)
	return nil
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	switch strings.ToLower(filepath.Ext(path)) {
	case ".webp":
		return webp.Decode(f)
	case ".png":
		return png.Decode(f)
	default:
		img, _, err := image.Decode(f)
		return img, err
	}
}

// downscale returns img resized (nearest-neighbor) so its longest side is
// at most maxDim pixels. It's a diagnostic image, not a deliverable, so
// nearest-neighbor is fine.
func downscale(img image.Image, maxDim int) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	longest := max(h, w)
	if longest <= maxDim {
		return img
	}
	scale := float64(maxDim) / float64(longest)
	nw, nh := int(float64(w)*scale), int(float64(h)*scale)

	out := image.NewRGBA(image.Rect(0, 0, nw, nh))
	for y := range nh {
		sy := b.Min.Y + int(float64(y)/scale)
		for x := range nw {
			sx := b.Min.X + int(float64(x)/scale)
			out.Set(x, y, img.At(sx, sy))
		}
	}
	return out
}

var markerColor = color.RGBA{R: 255, G: 0, B: 0, A: 255}

// markerRadius is the marker circle radius in source-image pixels. Never
// needed tuning in practice, so it's a constant rather than a flag.
const markerRadius = 18

func drawMarker(canvas *image.RGBA, m marker) {
	drawCircle(canvas, m.x, m.y, markerRadius, markerColor)
	label := fmt.Sprintf("%d %s", m.i, m.title)
	drawLabel(canvas, m.x+markerRadius+6, m.y+markerRadius/2, label)
}

// drawCircle draws a ring (not filled) of the given radius, ~4px thick.
func drawCircle(canvas *image.RGBA, cx, cy, radius int, col color.Color) {
	const thickness = 4
	b := canvas.Bounds()
	for dy := -radius - thickness; dy <= radius+thickness; dy++ {
		for dx := -radius - thickness; dx <= radius+thickness; dx++ {
			d2 := dx*dx + dy*dy
			if d2 < (radius-thickness)*(radius-thickness) || d2 > (radius+thickness)*(radius+thickness) {
				continue
			}
			x, y := cx+dx, cy+dy
			if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
				continue
			}
			canvas.Set(x, y, col)
		}
	}
}

// drawLabel draws white-outlined black text so it reads on any background.
func drawLabel(canvas *image.RGBA, x, y int, text string) {
	face := basicfont.Face7x13
	pt := fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}
	for _, off := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		drawText(canvas, face, pt.X+fixed.I(off[0]), pt.Y+fixed.I(off[1]), text, color.White)
	}
	drawText(canvas, face, pt.X, pt.Y, text, color.Black)
}

func drawText(canvas *image.RGBA, face font.Face, x, y fixed.Int26_6, text string, col color.Color) {
	d := &font.Drawer{
		Dst:  canvas,
		Src:  image.NewUniform(col),
		Face: face,
		Dot:  fixed.Point26_6{X: x, Y: y},
	}
	d.DrawString(text)
}
