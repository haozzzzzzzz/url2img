package url2img

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/therecipe/qt/core"
	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/network"
	"github.com/therecipe/qt/webkit"
	"github.com/therecipe/qt/widgets"
)

//go:generate qtmoc

// Object struct
type Object struct {
	core.QObject

	_ func(data string)     `signal:"load"`
	_ func(id, data string) `signal:"loadFinished"`
}

// Loader represents image loader
type Loader struct {
	*Object
	*widgets.QWidget
	app *widgets.QApplication
	Map sync.Map
}

// NewLoader returns new loader
func NewLoader() *Loader {
	os.Setenv("QT_QPA_PLATFORM", "offscreen")

	app := widgets.NewQApplication(len(os.Args), os.Args)
	app.SetApplicationName(Name)
	app.SetApplicationVersion(Version)

	widget := widgets.NewQWidget(nil, 0)
	widget.SetAttribute(core.Qt__WA_DontShowOnScreen, true)
	widget.Show()

	var sm sync.Map

	l := &Loader{NewObject(nil), widget, app, sm}

	l.ConnectLoad(func(data string) {
		params := NewParams()
		err := params.Unmarshal(data)
		if err == nil {
			l.LoadPage(params)
		}
	})

	l.ConnectLoadFinished(func(id, data string) {
		l.Map.Store(id, data)
	})

	return l
}

// LoadPage loads page
func (l *Loader) LoadPage(p Params) {
	view := webkit.NewQWebView(l.QWidget_PTR())
	view.SetAttribute(core.Qt__WA_DontShowOnScreen, true)
	view.Resize2(p.Width, p.Width)

	page := webkit.NewQWebPage(view.QWidget_PTR())

	// ignore ssl certificate errors
	networkAccessManager := network.NewQNetworkAccessManager(page)
	networkAccessManager.ConnectSslErrors(func(reply *network.QNetworkReply, errors []*network.QSslError) {
		reply.IgnoreSslErrors()
	})

	// check assets reply status
	loadAssetErrorCount := uint64(0)
	headers := map[string]string{
		"Refer": "https://img9.doubanio.com/",
		"sec-ch-ua": `"Google Chrome";v="87", " Not;A Brand";v="99", "Chromium";v="87"`,
		"sec-ch-ua-mobile": "?0",
		"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36",
	}

	_ = headers
	networkAccessManager.ConnectCreateRequest(func(op network.QNetworkAccessManager__Operation, originalReq *network.QNetworkRequest, outgoingData *core.QIODevice) *network.QNetworkReply {
		fmt.Println(originalReq.Url().Url(core.QUrl__None))
		//for key, value := range headers {
		//	originalReq.SetRawHeader(core.NewQByteArray2(key, len(key)), core.NewQByteArray2(value, len(value)))
		//}
		return networkAccessManager.CreateRequest(op, originalReq, outgoingData)
	})

	networkAccessManager.ConnectFinished(func(reply *network.QNetworkReply) {
		err := reply.Error()
		if err != network.QNetworkReply__NoError {
			path := reply.Url().Path(core.QUrl__PrettyDecoded)
			if strings.HasSuffix(path, "/favicon.ico") {
				return
			}

			atomic.AddUint64(&loadAssetErrorCount, 1)
			fmt.Printf("asset reply error. reply: %#v, err: %#v\n", reply.Url().Url(core.QUrl__None), err)
		}
	})

	page.SetNetworkAccessManager(networkAccessManager)

	view.SetPage(page)

	page.MainFrame().SetZoomFactor(p.Zoom)
	page.MainFrame().SetScrollBarPolicy(core.Qt__Horizontal, core.Qt__ScrollBarAlwaysOff)
	page.MainFrame().SetScrollBarPolicy(core.Qt__Vertical, core.Qt__ScrollBarAlwaysOff)

	page.ConnectUserAgentForUrl(func(url *core.QUrl) string {
		if p.UA != "" {
			return p.UA
		}

		return page.UserAgentForUrlDefault(url)
	})

	l.setAttributes(page.Settings())
	l.setPath(page.Settings(), os.TempDir())

	page.ConnectLoadFinished(func(pageOk bool) {

		if !pageOk && p.AbortOnLoadError {
			fmt.Printf("abort_on_load_error. url: %s\n", p.Url)

			l.LoadFinished(p.Id, "ErrAbortOnLoadError")
			view.DeleteLater()
			return
		}

		errAssetCount := atomic.LoadUint64(&loadAssetErrorCount)
		if errAssetCount > 0 && p.AbortOnLoadAssetError {
			fmt.Printf("abort_on_load_asset_error. url: %s, error_asset_count: %d\n", p.Url, errAssetCount)

			l.LoadFinished(p.Id, "ErrAbortOnLoadAssetError")
			view.DeleteLater()
			return
		}

		if p.Delay > 0 && !p.Full {
			time.Sleep(time.Duration(p.Delay) * time.Millisecond)
		}

		if p.Full {
			js := `var d=document;
				Math.max(Math.max(d.body.scrollHeight, d.documentElement.scrollHeight),
				Math.max(d.body.offsetHeight, d.documentElement.offsetHeight),
				Math.max(d.body.clientHeight, d.documentElement.clientHeight));`
			tmp := true
			p.Height = page.MainFrame().EvaluateJavaScript(js).ToInt(&tmp)

			if p.Height == 0 {
				p.Height = DefHeight
			} else if p.Height > 32768 {
				p.Height = 32768
			}

			page.SetViewportSize(core.NewQSize2(p.Width, p.Height))
			view.Resize2(p.Width, p.Height)

			page.MainFrame().EvaluateJavaScript(`window.scrollTo(0, ` + strconv.Itoa(p.Height) + `);`)

			if p.Delay > 0 {
				time.Sleep(time.Duration(p.Delay) * time.Millisecond)
			}
		}

		image := gui.NewQImage3(p.Width, p.Height, gui.QImage__Format_RGB888)
		if image.IsNull() {
			l.LoadFinished(p.Id, "ErrIsNull")
			view.DeleteLater()
			return
		}

		painter := gui.NewQPainter()
		painter.Begin(gui.NewQPaintDeviceFromPointer(image.Pointer()))
		if !painter.IsActive() {
			l.LoadFinished(p.Id, "ErrIsActive")
			view.DeleteLater()
			return
		}

		painter.SetRenderHint(gui.QPainter__Antialiasing, true)
		painter.SetRenderHint(gui.QPainter__TextAntialiasing, true)
		painter.SetRenderHint(gui.QPainter__HighQualityAntialiasing, true)
		painter.SetRenderHint(gui.QPainter__SmoothPixmapTransform, true)
		page.MainFrame().Render(painter, gui.NewQRegion())
		painter.End()

		buff := core.NewQBuffer(view)
		buff.Open(core.QIODevice__ReadWrite)
		if !buff.IsWritable() {
			l.LoadFinished(p.Id, "ErrIsWritable")
			view.DeleteLater()
			return
		}

		ok := image.Save2(buff, strings.ToUpper(p.Format), p.Quality)
		data := []byte(buff.Data().ConstData())
		if !ok {
			data = []byte("ErrSave2")
		}

		l.LoadFinished(p.Id, hex.EncodeToString(data))

		image.DestroyQImage()

		buff.Close()
		buff.DeleteLater()

		view.DeleteLater()
	})

	view.Show()
	view.Load(core.NewQUrl3(p.Url, core.QUrl__TolerantMode))
}

// setAttributes sets web page attributes
func (l *Loader) setAttributes(settings *webkit.QWebSettings) {
	settings.SetAttribute(webkit.QWebSettings__AutoLoadImages, true)
	settings.SetAttribute(webkit.QWebSettings__JavascriptEnabled, true)
	settings.SetAttribute(webkit.QWebSettings__JavascriptCanOpenWindows, false)
	settings.SetAttribute(webkit.QWebSettings__JavascriptCanCloseWindows, false)
	settings.SetAttribute(webkit.QWebSettings__JavascriptCanAccessClipboard, false)
	settings.SetAttribute(webkit.QWebSettings__LocalContentCanAccessFileUrls, true)
	settings.SetAttribute(webkit.QWebSettings__LocalContentCanAccessRemoteUrls, true)
	settings.SetAttribute(webkit.QWebSettings__SiteSpecificQuirksEnabled, true)
	settings.SetAttribute(webkit.QWebSettings__PrivateBrowsingEnabled, true)

	settings.SetAttribute(webkit.QWebSettings__PluginsEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__JavaEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__WebGLEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__WebAudioEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__NotificationsEnabled, false)

	settings.SetAttribute(webkit.QWebSettings__Accelerated2dCanvasEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__AcceleratedCompositingEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__TiledBackingStoreEnabled, false)

	settings.SetAttribute(webkit.QWebSettings__WebSecurityEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__LocalStorageEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__LocalStorageDatabaseEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__OfflineStorageDatabaseEnabled, false)
	settings.SetAttribute(webkit.QWebSettings__OfflineWebApplicationCacheEnabled, false)
}

// setPath sets storage path
func (l *Loader) setPath(settings *webkit.QWebSettings, path string) {
	settings.SetIconDatabasePath(path)
	settings.SetLocalStoragePath(path)
	settings.SetOfflineStoragePath(path)
	settings.SetOfflineWebApplicationCachePath(path)
}

// Exec starts Qt main loop
func (l *Loader) Exec() {
	l.app.Exec()
}

// Destroy destroys application
func (l *Loader) Destroy() {
	l.app.DestroyQApplication()
}
