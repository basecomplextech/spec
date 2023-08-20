package rpc

import (
	"time"
)

type ClientConfig struct {
	DialTimeout   time.Duration `json:"dial_timeout"`
	DialKeepAlive time.Duration `json:"dial_keep_alive"`
	ProxyFromEnv  bool          `json:"proxy_from_env"`

	TLSRootCert     string `json:"tls_root_cert"`
	TLSInsecureSkip bool   `json:"tls_insecure_skip"`

	MaxIdleConns          int           `json:"max_idle_conns"`
	IdleConnTimeout       time.Duration `json:"idle_conn_timeout"`
	ExpectContinueTimeout time.Duration `json:"expect_continue_timeout"`
	TLSHandshakeTimeout   time.Duration `json:"tls_handshake_timeout"`
}

func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		DialTimeout:   3 * time.Second,
		DialKeepAlive: 30 * time.Second,
		ProxyFromEnv:  false,

		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

type ServerConfig struct {
	Listen          string        `json:"listen"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`

	CertPath string `json:"cert_path"`
	KeyPath  string `json:"key_path"`
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Listen:          ":0",
		ShutdownTimeout: 5 * time.Second,

		CertPath: "certs/localhost.crt",
		KeyPath:  "certs/localhost.key",
	}
}
