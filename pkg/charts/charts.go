package charts

import (
	"embed"
)

//go:embed *
//go:embed workspace/templates/_helpers.tpl
var EmbeddedCharts embed.FS
