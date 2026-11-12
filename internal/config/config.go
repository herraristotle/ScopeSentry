package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/utils"

	"github.com/Autumn-27/ScopeSentry/internal/utils/random"
	"github.com/spf13/viper"
)

// Config global configuration struct
type Config struct {
	System  SystemConfig  `mapstructure:"system"`
	Server  ServerConfig  `mapstructure:"server"`
	MongoDB MongoDBConfig `mapstructure:"mongodb"`
	Redis   RedisConfig   `mapstructure:"redis"`
	JWT     JWTConfig     `mapstructure:"jwt"`
	Log     LogConfig     `mapstructure:"logs"`
}

// SystemConfig system configuration
type SystemConfig struct {
	Timezone      string `mapstructure:"timezone"`
	PluginKey     string `mapstructure:"plugin_key"`
	SecretKey     string `mapstructure:"secret_key"`
	Debug         bool   `mapstructure:"debug"`
	ExeDir        string `yaml:"-"`
	IconDir       string `yaml:"-"`
	ScreenshotDir string `yaml:"-"`
	ImgDir        string `yaml:"-"`
}

// ServerConfig server configuration
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// MongoDBConfig MongoDB configuration
type MongoDBConfig struct {
	IP       string `mapstructure:"ip"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// RedisConfig Redis configuration
type RedisConfig struct {
	IP       string `mapstructure:"ip"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
}

// JWTConfig JWT configuration
type JWTConfig struct {
	Secret string        `mapstructure:"secret"`
	Expire time.Duration `mapstructure:"expire"`
}

// LogConfig logging configuration
type LogConfig struct {
	TotalLogs int    `mapstructure:"total_logs"`
	Filename  string `mapstructure:"file_name"`
	Level     string `mapstructure:"leval"`
}

// GlobalConfig global configuration instance
var GlobalConfig Config

// init loads configuration
func init() {
	// Set default values
	setDefaults()

	// Set config file path
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println(err)
		return
	}
	exeDir := filepath.Dir(exePath)
	GlobalConfig.System.ExeDir = exeDir

	err = utils.EnsureDir(filepath.Join(exeDir, "files", "export"))
	if err != nil {
		fmt.Printf("Error creating directory: %s\n", err)
		return
	}

	// Create uploads directory
	err = utils.EnsureDir(filepath.Join(exeDir, "uploads"))
	if err != nil {
		fmt.Printf("Error creating uploads directory: %s\n", err)
		return
	}
	// Create images directory
	GlobalConfig.System.ImgDir = filepath.Join(exeDir, "images")
	err = utils.EnsureDir(GlobalConfig.System.ImgDir)
	if err != nil {
		fmt.Printf("Error creating images directory: %s\n", err)
		return
	}
	GlobalConfig.System.IconDir = filepath.Join(GlobalConfig.System.ImgDir, "icon")
	err = utils.EnsureDir(GlobalConfig.System.IconDir)
	if err != nil {
		fmt.Printf("Error creating icon directory: %s\n", err)
		return
	}
	GlobalConfig.System.ScreenshotDir = filepath.Join(GlobalConfig.System.ImgDir, "screenshots")
	err = utils.EnsureDir(GlobalConfig.System.ScreenshotDir)
	if err != nil {
		fmt.Printf("Error creating screenshots directory: %s\n", err)
		return
	}

	viper.AddConfigPath(exeDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Set environment variable prefix and separator
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		} else {
			fmt.Printf("error reading config file: %v\n", err)
			return
		}
	}
	// Load configuration from environment variables
	loadFromEnv()

	// Unmarshal configuration into struct
	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		fmt.Printf("error unmarshaling config: %v\n", err)
		return
	}

	// Handle plugin key
	if err := handlePluginKey(); err != nil {
		fmt.Println(err)
		return
	}
	if err := createConfigFile(); err != nil {
		fmt.Printf("error creating config file: %v\n", err)
		return
	}

	return
}

// setDefaults sets the default configuration
func setDefaults() {
	// System configuration
	viper.SetDefault("system.timezone", "Asia/Shanghai")
	viper.SetDefault("system.debug", false)

	viper.SetDefault("jwt.secret", random.GenerateString(32))
	viper.SetDefault("jwt.expire", time.Duration(24)*time.Hour)
	// Server configuration
	viper.SetDefault("server.port", 8082)
	viper.SetDefault("server.mode", "info")
	viper.SetDefault("server.read_timeout", 60*time.Second)
	viper.SetDefault("server.write_timeout", 60*time.Second)

	// MongoDB configuration
	viper.SetDefault("mongodb.ip", "127.0.0.1")
	viper.SetDefault("mongodb.port", 27017)
	viper.SetDefault("mongodb.database", "ScopeSentry")
	viper.SetDefault("mongodb.username", "")
	viper.SetDefault("mongodb.password", "")

	// Redis configuration
	viper.SetDefault("redis.ip", "127.0.0.1")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.password", "")

	// Logging configuration
	viper.SetDefault("logs.total_logs", 1000)
	viper.SetDefault("logs.level", "info")
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv() {
	// Timezone configuration
	if timezone := os.Getenv("TIMEZONE"); timezone != "" {
		viper.Set("system.timezone", timezone)
	}

	// MongoDB configuration
	if ip := os.Getenv("MONGODB_IP"); ip != "" {
		viper.Set("mongodb.ip", ip)
	}
	if port := os.Getenv("MONGODB_PORT"); port != "" {
		viper.Set("mongodb.port", port)
	}
	if database := os.Getenv("MONGODB_DATABASE"); database != "" {
		viper.Set("mongodb.database", database)
	}
	if username := os.Getenv("MONGODB_USER"); username != "" {
		viper.Set("mongodb.username", username)
	}
	if password := os.Getenv("MONGODB_PASSWORD"); password != "" {
		viper.Set("mongodb.password", password)
	}

	// Redis configuration
	if ip := os.Getenv("REDIS_IP"); ip != "" {
		viper.Set("redis.ip", ip)
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		viper.Set("redis.port", port)
	}
	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		viper.Set("redis.password", password)
	}
}

// createConfigFile creates the config file
func createConfigFile() error {
	// Get the directory containing the executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)

	// Set config file path
	configPath := filepath.Join(exeDir, "config.yaml")

	// Ensure the config directory exists
	if err := os.MkdirAll(exeDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %v", err)
	}

	// Write the config file
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}

// handlePluginKey handles the plugin key
func handlePluginKey() error {
	pluginKeyFile := "PLUGINKEY"
	if _, err := os.Stat(pluginKeyFile); err == nil {
		// File exists, read the key
		data, err := os.ReadFile(pluginKeyFile)
		if err != nil {
			return err
		}
		GlobalConfig.System.PluginKey = string(data)
	} else {
		// File does not exist, generate a new key
		key := random.GeneratePluginKey()
		if err := os.WriteFile(pluginKeyFile, []byte(key), 0644); err != nil {
			return err
		}
		GlobalConfig.System.PluginKey = key
	}
	return nil
}

// GetMongoURI returns the MongoDB connection URI
func GetMongoURI() string {
	uri := GlobalConfig.MongoDB.IP + ":" + string(GlobalConfig.MongoDB.Port)
	if GlobalConfig.MongoDB.Username != "" && GlobalConfig.MongoDB.Password != "" {
		uri = "mongodb://" + GlobalConfig.MongoDB.Username + ":" + GlobalConfig.MongoDB.Password + "@" + uri
	}
	return uri
}

// GetRedisAddr returns the Redis connection address
func GetRedisAddr() string {
	return GlobalConfig.Redis.IP + ":" + GlobalConfig.Redis.Port
}

// GetPluginKey returns the plugin key
func GetPluginKey() string {
	return GlobalConfig.System.PluginKey
}

// GetSecretKey returns the secret key
func GetSecretKey() string {
	return GlobalConfig.System.SecretKey
}

// GetTimezone returns the timezone
func GetTimezone() string {
	return GlobalConfig.System.Timezone
}

// GetTotalLogs returns the total log limit
func GetTotalLogs() int {
	return GlobalConfig.Log.TotalLogs
}
