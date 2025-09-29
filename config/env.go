package config

var (
	DebugEnabled      bool
	DebugSQLEnabled   bool
	DSN               string
	UpstreamBaseURL   string
	UpstreamAPIKey    string
	DailyRequestLimit int64
)

func ReloadEnv() {
	DebugEnabled = Bool("DEBUG", false)
	DebugSQLEnabled = Bool("DEBUG_SQL", false)
	DSN = String("DSN", "")
	UpstreamBaseURL = String("UPSTREAM_BASE_URL", "https://aiproxy.hzh.sealos.run")
	UpstreamAPIKey = String("UPSTREAM_API_KEY", "")
	DailyRequestLimit = Int64("DAILY_REQUEST_LIMIT", 30)
}

func init() {
	ReloadEnv()
}
