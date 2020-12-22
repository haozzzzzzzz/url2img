package url2img

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Params represent parameters
type Params struct {
	Id      string  `json:"id"`
	Url     string  `json:"url"`
	Output  string  `json:"output"`
	Format  string  `json:"format"`
	UA      string  `json:"ua"`
	Quality int     `json:"quality"`
	Delay   int     `json:"delay"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
	Zoom    float64 `json:"zoom"`
	Full    bool    `json:"full"`
	AbortOnLoadError bool `json:"abort_on_load_error"`
	AbortOnLoadAssetError bool `json:"abort_on_load_asset_error"`
	Headers map[string]string `json:"headers"`
}

// Default and maximum values
const (
	DefOutput = "raw"
	DefFormat = "jpg"

	DefQuality = 85
	DefDelay   = 0
	DefWidth   = 1600
	DefHeight  = 1200
	DefZoom    = 1.0
	DefFull    = false
	DefAbortOnLoadError = false
	DefAbortOnLoadAssetError = false

	maxQuality = 100
	maxDelay   = 10000
	maxWidth   = 4096
	maxHeight  = 4096
	maxZoom    = 5.0
)

// NewParams returns new params
func NewParams() Params {
	return Params{}
}

// FormValues gets params values from form
func (p *Params) FormValues(r *http.Request) (err error) {
	p.Url = strings.TrimSpace(r.FormValue("url"))
	if p.Url == "" {
		err = fmt.Errorf("empty url")
		return
	}

	if !strings.HasPrefix(p.Url, "http://") && !strings.HasPrefix(p.Url, "https://") {
		p.Url = "http://" + p.Url
	}

	err = p.genId()
	if err != nil {
		return
	}

	p.Output = DefOutput
	if r.FormValue("output") != "" {
		p.Output = r.FormValue("output")
		if !p.validOutput(p.Output) {
			err = fmt.Errorf("invalid output %s", p.Output)
			return
		}
	}

	p.Format = DefFormat
	if r.FormValue("format") != "" {
		p.Format = r.FormValue("format")
		if !p.validFormat(p.Format) {
			err = fmt.Errorf("invalid format %s", p.Format)
			return
		}
	}

	if r.FormValue("ua") != "" {
		p.UA = r.FormValue("ua")
	}

	p.Quality = DefQuality
	if r.FormValue("quality") != "" {
		p.Delay, err = strconv.Atoi(r.FormValue("quality"))
		if err != nil {
			return
		}

		if p.Quality > maxQuality {
			err = fmt.Errorf("quality maximum is %d", maxQuality)
			return
		}
	}

	p.Delay = DefDelay
	if r.FormValue("delay") != "" {
		p.Delay, err = strconv.Atoi(r.FormValue("delay"))
		if err != nil {
			return
		}

		if p.Delay > maxDelay {
			err = fmt.Errorf("delay maximum is %d", maxDelay)
			return
		}
	}

	p.Width = DefWidth
	if r.FormValue("width") != "" {
		p.Width, err = strconv.Atoi(r.FormValue("width"))
		if err != nil {
			return
		}

		if p.Width > maxWidth {
			err = fmt.Errorf("width maximum is %d", maxWidth)
			return
		}
	}

	p.Height = DefHeight
	if r.FormValue("height") != "" {
		p.Height, err = strconv.Atoi(r.FormValue("height"))
		if err != nil {
			return
		}

		if p.Height > maxHeight {
			err = fmt.Errorf("height maximum is %d", maxHeight)
			return
		}
	}

	p.Zoom = DefZoom
	if r.FormValue("zoom") != "" {
		p.Zoom, err = strconv.ParseFloat(r.FormValue("zoom"), 64)
		if err != nil {
			return
		}

		if p.Zoom > maxZoom {
			err = fmt.Errorf("zoom maximum is %f", maxZoom)
			return
		}
	}

	p.Full = DefFull
	if r.FormValue("full") != "" {
		p.Full = (r.FormValue("full") == "true" || r.FormValue("full") == "1")
	}

	p.AbortOnLoadError = DefAbortOnLoadError
	strAbortOnLoadError := r.FormValue("abort_on_load_error")
	if strAbortOnLoadError != "" {
		p.AbortOnLoadError = ( strAbortOnLoadError == "true" || strAbortOnLoadError == "1" )
	}

	p.AbortOnLoadAssetError = DefAbortOnLoadAssetError
	strAbortOnLoadAssetError := r.FormValue("abort_on_load_asset_error")
	if strAbortOnLoadAssetError != "" {
		p.AbortOnLoadAssetError = ( strAbortOnLoadAssetError == "true" || strAbortOnLoadAssetError == "1" )
	}

	p.Headers = make(map[string]string)

	return
}

// BodyValues gets params values from json body
func (p *Params) BodyValues(r *http.Request) (err error) {
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(p)
	if err != nil {
		return
	}

	p.Url = strings.TrimSpace(p.Url)
	if p.Url == "" {
		err = fmt.Errorf("empty url")
		return
	}

	if !strings.HasPrefix(p.Url, "http://") && !strings.HasPrefix(p.Url, "https://") {
		p.Url = "http://" + p.Url
	}

	err = p.genId()
	if err != nil {
		return
	}

	if p.Output == "" {
		p.Output = DefOutput
	} else {
		if !p.validOutput(p.Output) {
			err = fmt.Errorf("invalid output %s", p.Output)
			return
		}
	}

	if p.Format == "" {
		p.Format = DefFormat
	} else {
		if !p.validFormat(p.Format) {
			err = fmt.Errorf("invalid format %s", p.Format)
			return
		}
	}

	if p.Quality == 0 {
		p.Quality = DefQuality
	} else {
		if p.Quality > maxQuality {
			err = fmt.Errorf("quality maximum is %d", maxQuality)
			return
		}
	}

	if p.Delay != 0 {
		if p.Delay > maxDelay {
			err = fmt.Errorf("delay maximum is %d", maxDelay)
			return
		}
	}

	if p.Width == 0 {
		p.Width = DefWidth
	} else {
		if p.Width > maxWidth {
			err = fmt.Errorf("width maximum is %d", maxWidth)
			return
		}
	}

	if p.Height == 0 {
		p.Height = DefHeight
	} else {
		if p.Height > maxHeight {
			err = fmt.Errorf("height maximum is %d", maxHeight)
			return
		}
	}

	if p.Zoom == 0 {
		p.Zoom = DefZoom
	} else {
		if p.Zoom > maxZoom {
			err = fmt.Errorf("zoom maximum is %f", maxZoom)
			return
		}
	}

	return
}

// Marshal marshals params to string
func (p *Params) Marshal() (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Unmarshal unmarshals params from string
func (p *Params) Unmarshal(data string) error {
	err := json.Unmarshal([]byte(data), p)
	if err != nil {
		return err
	}

	return nil
}

// genId generates random id
func (p *Params) genId() (err error) {
	id := make([]byte, 16)
	_, err = rand.Read(id)
	if err != nil {
		return
	}

	p.Id = hex.EncodeToString(id)
	return
}

// validFormat checks if image format is valid
func (p *Params) validFormat(format string) bool {
	for _, f := range []string{"jpg", "jpeg", "png"} {
		if f == format {
			return true
		}
	}
	return false
}

// validOutput checks if output is valid
func (p *Params) validOutput(out string) bool {
	for _, o := range []string{"raw", "base64", "html"} {
		if o == out {
			return true
		}
	}
	return false
}
