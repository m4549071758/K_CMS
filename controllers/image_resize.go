package controllers

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/chai2010/webp"
	"golang.org/x/image/draw"
)

// resizeWidths はフロントエンド(next/imageのdeviceSizes・imageSizesと記事本文のsrcset)が
// 要求しうる幅。任意の幅を受け付けるとキャッシュが無制限に増えるため、この集合へ丸める。
var resizeWidths = []int{16, 32, 48, 64, 96, 128, 256, 384, 640, 750, 828, 1080, 1200, 1920, 2048, 3840}

// resizedCacheDir は縮小済み画像の置き場所。画像ディレクトリ配下の隠しディレクトリにし、
// GetImageはディレクトリを配信しないことで外部から一覧されないようにする。
var resizedCacheDir = filepath.Join("images", ".resized")

const resizeQuality = 80

// resizeSlots は同時に実行する縮小処理の上限。
var resizeSlots = make(chan struct{}, 2)

// snapWidth は要求幅以上で最小の許可幅を返す。範囲外は最大の許可幅に丸める。
func snapWidth(requested int) int {
	i := sort.SearchInts(resizeWidths, requested)
	if i == len(resizeWidths) {
		return resizeWidths[len(resizeWidths)-1]
	}
	return resizeWidths[i]
}

// resizedImagePath は元画像を幅widthへ縮小したWebPのパスを返す。
// 元画像がwidth以下の幅なら縮小せず元画像のパスを返す。生成物はキャッシュして再利用する。
func resizedImagePath(originalPath, fileName string, width int) (string, error) {
	cachePath := filepath.Join(resizedCacheDir, strconv.Itoa(width), fileName)
	if info, err := os.Stat(cachePath); err == nil && !info.IsDir() {
		return cachePath, nil
	}

	// 大きな画像の縮小はメモリを食うため同時実行数を絞る。待機後に他リクエストが生成済みなら再利用する。
	resizeSlots <- struct{}{}
	defer func() { <-resizeSlots }()
	if info, err := os.Stat(cachePath); err == nil && !info.IsDir() {
		return cachePath, nil
	}

	data, err := os.ReadFile(originalPath)
	if err != nil {
		return "", err
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("decode config: %w", err)
	}
	if cfg.Width <= width {
		return originalPath, nil
	}
	height := max(cfg.Height*width/cfg.Width, 1)

	var dst image.Image
	if format == "webp" {
		// libwebpのスケーリングデコードで目的サイズへ直接展開し、元サイズのピクセルを確保しない。
		dst, err = webp.DecodeRGBAToSize(data, width, height)
		if err != nil {
			return "", fmt.Errorf("decode webp to size: %w", err)
		}
	} else {
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return "", fmt.Errorf("decode: %w", err)
		}
		scaled := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.CatmullRom.Scale(scaled, scaled.Bounds(), img, img.Bounds(), draw.Src, nil)
		dst = scaled
	}

	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return "", err
	}
	// 同時リクエストで書きかけのファイルを配信しないよう、一時ファイルに書いてからrenameする。
	tmp, err := os.CreateTemp(filepath.Dir(cachePath), fileName+".*.tmp")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	encodeErr := webp.Encode(tmp, dst, &webp.Options{Quality: resizeQuality})
	closeErr := tmp.Close()
	if encodeErr != nil || closeErr != nil {
		os.Remove(tmpName)
		if encodeErr != nil {
			return "", fmt.Errorf("encode: %w", encodeErr)
		}
		return "", closeErr
	}
	if err := os.Rename(tmpName, cachePath); err != nil {
		os.Remove(tmpName)
		return "", err
	}
	return cachePath, nil
}

// removeResizedImages は元画像の削除に合わせて縮小済みキャッシュを消す。
func removeResizedImages(fileName string) {
	for _, width := range resizeWidths {
		os.Remove(filepath.Join(resizedCacheDir, strconv.Itoa(width), fileName))
	}
}
