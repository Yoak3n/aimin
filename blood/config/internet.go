package config

type Internet struct {
	BilibiliCookie string `json:"bilibili_cookie"`
}

func DefaultInternet() *Internet {
	return &Internet{
		BilibiliCookie: "",
	}
}
