package doc

import _ "embed"

//go:embed data/prompts/chat.txt
var ChatPrompt string

//go:embed data/prompts/chart-report.txt
var ChartReportPrompt string

//go:embed data/prompts/bond-report.txt
var BondReportPrompt string

//go:embed data/prompts/naming-report.txt
var NamingReportPrompt string

