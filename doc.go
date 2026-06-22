package doc

import _ "embed"


//go:embed openapi.json
var OpenAPIJSON []byte
//go:embed data/prompts/chat.txt
var ChatPrompt string

//go:embed web/skills/report-chart.md
var ChartReportPrompt string

//go:embed web/skills/report-bond.md
var BondReportPrompt string

//go:embed web/skills/report-naming.md
var NamingReportPrompt string

