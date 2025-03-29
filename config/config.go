package config

const (
	TarantoolAddress       = "127.0.0.1:3301"
	MattermostURL          = "http://localhost:8065"
	MettermostWebSocketURL = "ws://localhost:8065"
	MattermostToken        = "yzsqmq796b88mxw135fusjbnha"
	ChannelID              = "tbczkzdgy7b98xikgpbyh15yzr"
	UserID                 = "sj1tz4q9it86pdmac87u4x8hoo"
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
