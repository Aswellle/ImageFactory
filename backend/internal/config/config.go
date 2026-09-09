package config

import (
	"github.com/spf13/viper"
)

// Config holds all ImageForge configuration. Values are read from environment
// variables or a config file via Viper. No secret is ever logged.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Email    EmailConfig    `mapstructure:"email"`
	Sub2API  Sub2APIConfig  `mapstructure:"sub2api"`
}

type ServerConfig struct {
	Host            string   `mapstructure:"host"`
	Port            int      `mapstructure:"port"`
	Mode            string   `mapstructure:"mode"` // debug, release, test
	Timeout         int      `mapstructure:"timeout_seconds"`
	AllowedOrigins  []string `mapstructure:"allowed_origins"`
	TrustedProxies  []string `mapstructure:"trusted_proxies"`
	// TLS 可选配置。设置 TLSCertFile + TLSKeyFile 后启用 HTTPS。
	// 生产环境通常在 Caddy 终止 TLS，后端保持 HTTP。
	TLSCertFile     string   `mapstructure:"tls_cert_file"`
	TLSKeyFile      string   `mapstructure:"tls_key_file"`
}



type DatabaseConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"` // secret
	Name        string `mapstructure:"name"`
	SSLMode     string `mapstructure:"sslmode"`
	AutoMigrate bool   `mapstructure:"auto_migrate"` // run Ent schema migration on startup
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"` // secret
	DB       int    `mapstructure:"db"`
}

type AuthConfig struct {
	JWTSecret          string `mapstructure:"jwt_secret"` // secret, never logged
	AccessTokenMinutes int    `mapstructure:"access_token_minutes"`
	BcryptCost         int    `mapstructure:"bcrypt_cost"`
	AdminPanelKey      string `mapstructure:"admin_panel_key"` // optional service-to-service admin panel key
}


// EmailConfig configures the email delivery service (SendGrid).
type EmailConfig struct {
	// Provider selects the email delivery backend: "sendgrid" | "smtp" | "console".
	Provider string `mapstructure:"provider"`
	// SendGridAPIKey is the SendGrid API key (required when Provider="sendgrid").
	SendGridAPIKey string `mapstructure:"sendgrid_api_key"` // secret
	// FromEmail is the sender address shown in sent emails.
	FromEmail string `mapstructure:"from_email"`
	// FromName is the sender name shown in sent emails.
	FromName string `mapstructure:"from_name"`
}


type StorageConfig struct {
	// S3-compatible object storage (Cloudflare R2, MinIO, etc.)
	Provider   string `mapstructure:"provider"` // "r2" | "minio" | "filesystem"
	Bucket     string `mapstructure:"bucket"`
	Region     string `mapstructure:"region"`
	Endpoint   string `mapstructure:"endpoint"` // custom endpoint for R2/MinIO
	AccessKey  string `mapstructure:"access_key"` // secret
	SecretKey  string `mapstructure:"secret_key"` // secret
	PublicHost string `mapstructure:"public_host"` // optional CDN/public host override
}

// Sub2APIConfig configures the upstream AI gateway connection.
type Sub2APIConfig struct {
	// BaseURL is the Sub2API proxy address ImageForge workers call.
	BaseURL string `mapstructure:"base_url"`
	// AdminAPIKey authenticates ImageForge backend → Sub2API. Secret, never exposed to browsers.
	AdminAPIKey string `mapstructure:"admin_api_key"`
	// AccountID is the Sub2API upstream account used for image generation.
	AccountID string `mapstructure:"account_id"`
	// GeminiAPIKey is the Google Gemini API key used for image generation.
	GeminiAPIKey string `mapstructure:"gemini_api_key"`
	TimeoutSeconds int `mapstructure:"timeout_seconds"`
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return "host=" + d.Host +
		" port=" + itoa(d.Port) +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.Name +
		" sslmode=" + d.SSLMode
}

func itoa(n int) string {
	if n == 0 {
		return ""
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// Load reads configuration from file + environment (IF_ prefix).
func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "release")
	v.SetDefault("server.timeout_seconds", 30)

	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.sslmode", "disable")
	v.SetDefault("database.auto_migrate", true)


	v.SetDefault("redis.host", "localhost")
	v.SetDefault("redis.port", 6379)
	v.SetDefault("redis.db", 0)

	v.SetDefault("auth.access_token_minutes", 60*24)
	v.SetDefault("auth.bcrypt_cost", 12)

	v.SetDefault("storage.provider", "filesystem")
	v.SetDefault("storage.region", "auto")
	v.SetDefault("storage.bucket", "imageforge")

	v.SetDefault("sub2api.timeout_seconds", 120)

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./configs")
	}

	v.SetEnvPrefix("IF")
	v.AutomaticEnv()

	// Bind explicit env overrides for secrets so they never need to live in a file.
	_ = v.BindEnv("auth.jwt_secret")
	_ = v.BindEnv("database.password")
	_ = v.BindEnv("redis.password")
	_ = v.BindEnv("storage.access_key")
	_ = v.BindEnv("storage.secret_key")
	_ = v.BindEnv("sub2api.admin_api_key")
	_ = v.BindEnv("sub2api.gemini_api_key")
	// Bind slice-typed env vars (comma-separated) for CORS + trusted proxies.
	_ = v.BindEnv("server.allowed_origins")
	_ = v.BindEnv("server.trusted_proxies")

	var cfg Config
	if err := v.ReadInConfig(); err != nil {
		// Config file is optional; env-only is valid.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
