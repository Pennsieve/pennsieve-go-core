package pusher

type Config struct {
	AppId   string `json:"app_id"`
	Key     string `json:"key"`
	secret  string `json:"secret"`
	cluster string `json:"cluster"`
}
