module marksafe

go 1.26.0

toolchain go1.26.10

require (
	github.com/alecthomas/chroma/v2 v2.23.1
	github.com/litao91/goldmark-mathjax v0.0.0-20210217064022-a43cf739a50f
	github.com/microcosm-cc/bluemonday v1.0.27
	github.com/spf13/cobra v1.10.2
	github.com/wailsapp/wails/v3 v3.1.0
	github.com/yuin/goldmark v1.7.17
	github.com/yuin/goldmark-emoji v1.0.6
	github.com/yuin/goldmark-highlighting/v2 v2.0.0-20230729083705-37449abec8cc
	github.com/yuin/goldmark-meta v1.1.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/ebitengine/purego v0.11.1 // indirect
	golang.org/x/term v0.46.0 // indirect
)

require (
	github.com/aymerick/douceur v0.2.0 // indirect
	github.com/coder/websocket v1.8.14 // indirect
	github.com/dlclark/regexp2 v1.11.5 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	github.com/gorilla/css v1.0.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mlechner911/mdviewer/mlcext v0.0.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	golang.org/x/net v0.60.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

// replace github.com/wailsapp/wails/v2 v2.11.0 => /home/mlc/go/pkg/mod

replace github.com/mlechner911/mdviewer/mlcext => ./mlcext
