package controllers

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/chai2010/webp"
)

func TestSnapWidth(t *testing.T) {
	cases := []struct{ requested, want int }{
		{1, 16},
		{640, 640},
		{641, 750},
		{1201, 1920},
		{99999, 3840},
	}
	for _, tc := range cases {
		if got := snapWidth(tc.requested); got != tc.want {
			t.Errorf("snapWidth(%d) = %d, want %d", tc.requested, got, tc.want)
		}
	}
}

func writeTestWebP(t *testing.T, path string, width, height int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 128, 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := webp.Encode(f, img, &webp.Options{Quality: 80}); err != nil {
		t.Fatal(err)
	}
}

func TestResizedImagePath(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.Mkdir("images", 0o755); err != nil {
		t.Fatal(err)
	}
	original := filepath.Join("images", "a.webp")
	writeTestWebP(t, original, 1000, 600)

	t.Run("縮小して縦横比を保つ", func(t *testing.T) {
		got, err := resizedImagePath(original, "a.webp", 640)
		if err != nil {
			t.Fatal(err)
		}
		if got == original {
			t.Fatal("縮小版ではなく元画像が返った")
		}
		f, err := os.Open(got)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		cfg, err := webp.DecodeConfig(f)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Width != 640 || cfg.Height != 384 {
			t.Fatalf("size = %dx%d, want 640x384", cfg.Width, cfg.Height)
		}
	})

	t.Run("元画像以上の幅では拡大しない", func(t *testing.T) {
		got, err := resizedImagePath(original, "a.webp", 1080)
		if err != nil {
			t.Fatal(err)
		}
		if got != original {
			t.Fatalf("got %s, want original", got)
		}
	})

	t.Run("削除でキャッシュも消える", func(t *testing.T) {
		cached, err := resizedImagePath(original, "a.webp", 256)
		if err != nil {
			t.Fatal(err)
		}
		removeResizedImages("a.webp")
		if _, err := os.Stat(cached); !os.IsNotExist(err) {
			t.Fatalf("cache still exists: %v", err)
		}
	})
}
