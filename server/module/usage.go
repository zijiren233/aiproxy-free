package module

// UsageResponse API key使用情况查询响应
type UsageResponse struct {
	TotalLimit     int64 `json:"total_limit"`     // 总共可以使用多少次
	UsedToday      int64 `json:"used_today"`      // 过去一天内使用了多少次
	RemainingToday int64 `json:"remaining_today"` // 今天还能使用多少次
	NextResetTime  int64 `json:"next_reset_time"` // 下一次重置时间
}
