package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var (
	// Version is the current version of rcsm
	Version string = "0.0.2"
	// EnvFile is the path to the .env file config
	EnvFile string = ".env"

	// InstanceName is used for event reporting on Redis and Webhooks, useful if you have multiple rcsm instances
	InstanceName string = "server"

	// RedisEnabled specifies if Redis communication should be enabled
	RedisEnabled bool = false
	// RedisHost specifies the Redis server to use
	RedisHost string = "localhost:6379"
	// RedisPassword is the plaintext password of the server
	RedisPassword string = ""
	// RedisDatabase is the database ID used for Redis
	RedisDatabase int64 = 0
	// RedisPubSubChannel is the channel used for Redis pub/sub notifications
	RedisPubSubChannel string = "rcsm"

	// S3Enabled specifies wether or not S3 is enabled to update the server from templates
	S3Enabled bool = false
	// S3Endpoint specifies the S3 endpoint if you use something else than AWS
	S3Endpoint string = ""
	// S3Region specifies the region to use for the S3 bucket
	S3Region string = ""
	// S3Bucket specifies the bucket name for server templates
	S3Bucket string = ""
	// AWSAccessKeyID is the key ID for S3 authentication
	AWSAccessKeyID string = ""
	// AWSSecretAccessKey is the secret key for S3 authentication
	AWSSecretAccessKey string = ""

	// MinecraftServersDirectory is the directory where server directories are stored
	MinecraftServersDirectory string = "/opt/minecraft"
	// MinecraftServersToCreate is the servers you want to deploy if a template exists on S3
	MinecraftServersToCreate string = ""
	// MinecraftTmuxSessionPrefix is the prefix to use for tmux session names
	MinecraftTmuxSessionPrefix string = "rcsm_"

	// AutoStartOnBoot specifies if Minecraft servers should start when rcsm starts
	AutoStartOnBoot bool = true
	// AutoStopOnClose specifies if Minecraft servers should stopped when rcsm closes
	AutoStopOnClose bool = false
	// AutoRestartCrashEnabled specifies if rcsm should attempt to restart servers on crash
	AutoRestartCrashEnabled bool = true
	// AutoRestartCrashMaxTries specifies how many tries rcsm should attempt to get a server running for more than 5 minutes
	AutoRestartCrashMaxTries int64 = 3
	// AutoRestartCrashTimeoutSec specifies for how long rcsm will wait to kill the server if not responding
	AutoRestartCrashTimeoutSec int64 = 60

	// WebhooksEnabled specifies if Webhooks (using Discord format) are enabled for alerts
	WebhooksEnabled bool = false
	// WebhooksEndpoint is the endpoint to use to send notifications to
	WebhooksEndpoint string = ""

	// AutoUpdateEnabled specifies if the auto update system should check for new versions of rcsm and install them
	AutoUpdateEnabled bool = true
	// AutoUpdateIntervalMinutes specifies how often updates should be checked
	AutoUpdateIntervalMinutes int64 = 60
	// AutoUpdateRepo specifies where to download updates for the last rcsm release
	AutoUpdateRepo string = "redcraft-org/redcraft_server_management"
	// ExitOnAutoUpdate specifies if rcsm should quit itself once updated. This is very useful when wrapped with systemd
	ExitOnAutoUpdate bool = false
)

// ReadConfig reads the config from the env file
func ReadConfig() {
	EnvFile = ReadEnvString("RCSM_ENV_FILE", ".env")

	godotenv.Load(EnvFile)

	InstanceName = ReadEnvString("INSTANCE_NAME", InstanceName)

	RedisEnabled = ReadEnvBool("REDIS_ENABLED", RedisEnabled)
	RedisHost = ReadEnvString("REDIS_HOST", RedisHost)
	RedisPassword = ReadEnvString("REDIS_PASSWORD", RedisPassword)
	RedisDatabase = ReadEnvInt("REDIS_DATABASE", RedisDatabase)
	RedisPubSubChannel = ReadEnvString("REDIS_PUB_SUB_CHANNEL", RedisPubSubChannel)

	S3Enabled = ReadEnvBool("S3_ENABLED", S3Enabled)
	S3Endpoint = ReadEnvString("S3_ENDPOINT", S3Endpoint)
	S3Region = ReadEnvString("S3_REGION", S3Region)
	S3Bucket = ReadEnvString("S3_BUCKET", S3Bucket)
	AWSAccessKeyID = ReadEnvString("AWS_ACCESS_KEY_ID", AWSAccessKeyID)
	AWSSecretAccessKey = ReadEnvString("AWS_SECRET_ACCESS_KEY", AWSSecretAccessKey)

	MinecraftServersDirectory = ReadEnvString("MINECRAFT_SERVERS_DIRECTORY", MinecraftServersDirectory)
	MinecraftServersToCreate = ReadEnvString("MINECRAFT_SERVERS_TO_CREATE", MinecraftServersToCreate)
	MinecraftTmuxSessionPrefix = ReadEnvString("MINECRAFT_TMUX_SESSION_PREFIX", MinecraftTmuxSessionPrefix)

	AutoStartOnBoot = ReadEnvBool("AUTO_START_ON_BOOT", AutoStartOnBoot)
	AutoStopOnClose = ReadEnvBool("AUTO_STOP_ON_CLOSE", AutoStopOnClose)
	AutoRestartCrashEnabled = ReadEnvBool("AUTO_RESTART_CRASH_ENABLED", AutoRestartCrashEnabled)
	AutoRestartCrashMaxTries = ReadEnvInt("AUTO_RESTART_CRASH_MAX_TRIES", AutoRestartCrashMaxTries)
	AutoRestartCrashTimeoutSec = ReadEnvInt("AUTO_RESTART_CRASH_TIMEOUT_SEC", AutoRestartCrashTimeoutSec)

	WebhooksEnabled = ReadEnvBool("WEBHOOKS_ENABLED", WebhooksEnabled)
	WebhooksEndpoint = ReadEnvString("WEBHOOKS_ENDPOINT", WebhooksEndpoint)

	AutoUpdateEnabled = ReadEnvBool("AUTO_UPDATE_ENABLED", AutoUpdateEnabled)
	AutoUpdateIntervalMinutes = ReadEnvInt("AUTO_UPDATE_INTERVAL_MINUTES", AutoUpdateIntervalMinutes)
	AutoUpdateRepo = ReadEnvString("AUTO_UPDATE_REPO", AutoUpdateRepo)
	ExitOnAutoUpdate = ReadEnvBool("EXIT_ON_AUTO_UPDATE", ExitOnAutoUpdate)
}

// ReadEnvString reads a string from the env variables
func ReadEnvString(envName string, defaultValue string) string {
	envVar := os.Getenv(envName)
	if envVar == "" {
		return defaultValue
	}
	return envVar
}

// ReadEnvInt reads an integer from the env variables
func ReadEnvInt(envName string, defaultValue int64) int64 {
	envVarRaw := os.Getenv(envName)
	if envVarRaw == "" {
		return defaultValue
	}
	envVar, err := strconv.ParseInt(envVarRaw, 10, 64)
	if err != nil {
		return defaultValue
	}
	return envVar
}

// ReadEnvBool reads a boolean from the env variables
func ReadEnvBool(envName string, defaultValue bool) bool {
	envVarRaw := os.Getenv(envName)
	if envVarRaw == "" {
		return defaultValue
	}
	envVar := strings.ToLower(envVarRaw)

	switch envVar {
	case
		"true",
		"yes",
		"on",
		"1":
		return true
	}
	return false
}
