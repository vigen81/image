package processor

import "errors"

type ContentType string

const (
	Jpg  ContentType = "image/jpg"
	Jpeg ContentType = "image/jpeg"
	Png  ContentType = "image/png"
	Gif  ContentType = "image/gif"
	Webp ContentType = "image/webp"
	Svg  ContentType = "image/svg+xml"
)

func ShouldConvertToWebp(contentType string) bool {
	switch ContentType(contentType) {
	case Jpeg, Jpg, Png, Webp:
		return true
	}
	return false
}

func ReleaseExtension(contentType string) (string, error) {
	switch ContentType(contentType) {
	case Jpeg, Jpg, Png, Webp:
		return "webp", nil
	case Gif:
		return "gif", nil
	case Svg:
		return "svg", nil
	}
	return "", errors.New("unsupported content type")
}
