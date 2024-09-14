package types

type LFTP struct {
	Threads  int `mapstructure:"threads"`
	Segments int `mapstructure:"segments"`
}

type ServerInfo struct {
	Host     string `mapstructure:"host"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}
type ClientRules struct {
	CodeDestinations map[string]string `mapstructure:"codeDestinations"`
	LFTP             LFTP              `mapstructure:"lftp"`
	ServerInfo       ServerInfo        `mapstructure:"serverInfo"`
}

type Rule struct {
	Value    string `mapstructure:"value"`
	Operator string `mapstructure:"operator"`
	Entity   string `mapstructure:"entity"`
	Code     string `mapstructure:"code"`
}

type ServerRules struct {
	DefaultCode    string `mapstructure:"defaultCode"`
	CodeConditions []Rule `mapstructure:"codeConditions"`
}

type TorrentRules struct {
	Client string `mapstructure:"client"`
	Source string `mapstructure:"source"`
	Label  string `mapstructure:"labelling"`
}

type MQTTRules struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"`
	ClientId string `mapstructure:"clientId"`
	Host     string `mapstructure:"host"`
}

type Config struct {
	MQTT   MQTTRules   `mapstructure:"mqtt"`
	Server ServerRules `mapstructure:"server"`
	Client ClientRules `mapstructure:"client"`
}
