package pusher

type Config struct {
	AppId   string `json:"app_id"`
	Key     string `json:"key"`
	Secret  string `json:"secret"`
	Cluster string `json:"cluster"`
}
