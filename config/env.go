package config

var (
	DebugEnabled    bool
	DebugSQLEnabled bool
	DSN             string
)

func ReloadEnv() {
	DebugEnabled = Bool("DEBUG", false)
	DebugSQLEnabled = Bool("DEBUG_SQL", false)
	DSN = String("DSN", "")
}

func init() {
	ReloadEnv()
}
