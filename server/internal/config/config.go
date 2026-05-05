package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DB_URI                string
	ServerPort            string
	JWTSecret             string
	ENV                   string
	JWT_ACCESS_SECRET     string
	COOKIE_DOMAIN         string
	JWT_ISSUER            string
	AWS_ACCESS_KEY        string
	AWS_SECRET_ACCESS_KEY string
	AWSRegion             string
	S3Bucket              string
	FACE_API_KEY          string
	FACE_API_SECRET       string
	CORSAllowedOrigins    string
}

func LoadConfig() (Config, error) {
	// if err := godotenv.Load(); err != nil {
	// 	return Config{}, fmt.Errorf("Failed to load .env")
	// }
	_ = godotenv.Load()

	Db_uri, err := extractEnv("DB_URI")
	if err != nil {
		return Config{}, err
	}

	port, err := extractEnv("PORT")
	if err != nil {
		return Config{}, err
	}

	jwtsecret, err := extractEnv("JWT_SECRET")
	if err != nil {
		return Config{}, err
	}

	env, err := extractEnv("ENV")
	if err != nil {
		return Config{}, err
	}

	jwtaccesssecret, err := extractEnv("JWT_ACCESS_SECRET")
	if err != nil {
		return Config{}, err
	}

	cookiedomain, err := extractEnv("COOKIE_DOMAIN")
	if err != nil {
		return Config{}, err
	}

	jwtissuer, err := extractEnv("JWT_ISSUER")
	if err != nil {
		return Config{}, err
	}

	accesskey, err := extractEnv("AWS_ACCESS_KEY_ID")
	if err != nil {
		return Config{}, err
	}

	secretaccesskey, err := extractEnv("AWS_SECRET_ACCESS_KEY")
	if err != nil {
		return Config{}, err
	}

	awsRegion, err := extractEnv("AWS_REGION")
	if err != nil {
		return Config{}, err
	}
	s3Bucket, err := extractEnv("S3_BUCKET")
	if err != nil {
		return Config{}, err
	}
	faceapikey, err := extractEnv("FACE_API_KEY")
	if err != nil {
		return Config{}, err
	}

	faceapisecret, err := extractEnv("FACE_API_SECRET")
	if err != nil {
		return Config{}, err
	}

	CORSAllowedOrigins, err := extractEnv("CORS_ALLOWED_ORIGINS")
	if err != nil {
		return Config{}, err
	}

	return Config{
		DB_URI:                Db_uri,
		ServerPort:            port,
		JWTSecret:             jwtsecret,
		ENV:                   env,
		JWT_ACCESS_SECRET:     jwtaccesssecret,
		COOKIE_DOMAIN:         cookiedomain,
		JWT_ISSUER:            jwtissuer,
		AWS_ACCESS_KEY:        accesskey,
		AWS_SECRET_ACCESS_KEY: secretaccesskey,
		AWSRegion:             awsRegion,
		S3Bucket:              s3Bucket,
		FACE_API_KEY:          faceapikey,
		FACE_API_SECRET:       faceapisecret,
		CORSAllowedOrigins:    CORSAllowedOrigins,
	}, nil

}

func extractEnv(key string) (string, error) {
	val := os.Getenv(key)

	if val == "" {
		return "", fmt.Errorf("missing required env: %s", key)
	}

	return val, nil
}

func extractEnvOptional(key string) string {
	return os.Getenv(key)
}

var config *Config

// GetConfig returns the application configuration
// It loads the configuration if it hasn't been loaded yet
func GetConfig() *Config {
	if config == nil {
		cfg, err := LoadConfig()
		if err != nil {
			panic("Failed to load configuration: " + err.Error())
		}

		config = &cfg
	}
	return config
}
