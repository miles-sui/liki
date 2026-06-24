package doc

import _ "embed"


//go:embed openapi.json
var OpenAPIJSON []byte
//go:embed data/prompts/chat.txt
var ChatPrompt string

//go:embed data/prompts/report.txt
var ReportPrompt string

//go:embed web/skills/report-chart.md
var ChartReportPrompt string

//go:embed web/skills/report-bond.md
var BondReportPrompt string

//go:embed web/skills/report-naming.md
var NamingReportPrompt string

//go:embed web/skills/report-ziwei.md
var ZiweiReportPrompt string

//go:embed web/skills/report-bazhai.md
var BazhaiReportPrompt string

//go:embed web/skills/report-xuankong.md
var XuankongReportPrompt string

