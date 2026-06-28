package agent

import _ "embed"

//go:embed data/tools.json
var ToolsJSON []byte

//go:embed data/naming.txt
var NamingPrompt string
