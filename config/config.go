package config

const (
	TarantoolAddress       = "127.0.0.1:3301"
	MattermostURL          = "http://localhost:8065"
	MettermostWebSocketURL = "ws://localhost:8065"
	MattermostToken        = "YOUR_BOT_TOKEN"
	ChannelID              = "YOUR_CHANNEL_ID"
	UserID                 = "YOUR_USER_BOT_ID"
)

type MattermostConfig struct {
	URL          string
	WebSocketURL string
	Token        string
}

type TarantoolConfig struct {
	Address  string
	User     string
	Password string
}

func LoadConfig() (TarantoolConfig, MattermostConfig) {
	return TarantoolConfig{Address: TarantoolAddress}, MattermostConfig{URL: MattermostURL, Token: MattermostToken, WebSocketURL: MettermostWebSocketURL}
}
