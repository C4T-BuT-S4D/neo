package client

type Config struct {
	Host        string `mapstructure:"host"`
	ExploitDir  string `mapstructure:"exploit_dir"`
	GrpcAuthKey string `mapstructure:"grpc_auth_key"`
	UseTLS      bool   `mapstructure:"use_tls"`
}
