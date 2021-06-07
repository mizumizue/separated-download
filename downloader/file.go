package downloader

type File struct {
	Data []byte
	Type string
}

func extByContentType(ct string) string {
	switch ct {
	case "application/json", "application/json; charset=UTF-8":
		return ".json"
	case "image/jpeg":
		return ".jpeg"
	default:
		return ""
	}
}
