package config

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrMissingKey          = errors.New("HALO_KEY is required")
	ErrMissingDBPath       = errors.New("HALO_DB_PATH is required")
	ErrMissingMediaDir     = errors.New("HALO_MEDIA_DIR is required")
	ErrUnknownProvider     = errors.New("HALO_LLM_PROVIDER must be anthropic")
	ErrMissingAnthropicKey = errors.New("ANTHROPIC_API_KEY is required")
	ErrMissingWhisperModel = errors.New("WHISPER_MODEL is required")
	ErrInvalidHome         = errors.New("HALO_HOME_LAT and HALO_HOME_LON must both be valid coordinates or both be empty")
)

const ProviderAnthropic = "anthropic"

type Config struct {
	ListenAddr      string
	Key             string
	DBPath          string
	MediaDir        string
	LLMProvider     string
	AnthropicAPIKey string
	WhisperBin      string
	WhisperModel    string
	FFmpegBin       string
	HomeLat         *float64
	HomeLon         *float64
}

func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		ListenAddr:      withDefault(getenv("LISTEN_ADDR"), "127.0.0.1:8082"),
		Key:             getenv("HALO_KEY"),
		DBPath:          getenv("HALO_DB_PATH"),
		MediaDir:        getenv("HALO_MEDIA_DIR"),
		LLMProvider:     withDefault(getenv("HALO_LLM_PROVIDER"), ProviderAnthropic),
		AnthropicAPIKey: getenv("ANTHROPIC_API_KEY"),
		WhisperBin:      withDefault(getenv("WHISPER_BIN"), "whisper-cli"),
		WhisperModel:    getenv("WHISPER_MODEL"),
		FFmpegBin:       withDefault(getenv("FFMPEG_BIN"), "ffmpeg"),
	}
	required := []struct {
		value string
		err   error
	}{
		{cfg.Key, ErrMissingKey},
		{cfg.DBPath, ErrMissingDBPath},
		{cfg.MediaDir, ErrMissingMediaDir},
	}
	for _, r := range required {
		if strings.TrimSpace(r.value) == "" {
			return Config{}, r.err
		}
	}
	if cfg.LLMProvider != ProviderAnthropic {
		return Config{}, ErrUnknownProvider
	}
	if strings.TrimSpace(cfg.AnthropicAPIKey) == "" {
		return Config{}, ErrMissingAnthropicKey
	}
	if strings.TrimSpace(cfg.WhisperModel) == "" {
		return Config{}, ErrMissingWhisperModel
	}
	lat, lon, err := home(getenv("HALO_HOME_LAT"), getenv("HALO_HOME_LON"))
	if err != nil {
		return Config{}, err
	}
	cfg.HomeLat = lat
	cfg.HomeLon = lon
	return cfg, nil
}

func home(latText string, lonText string) (*float64, *float64, error) {
	if latText == "" && lonText == "" {
		return nil, nil, nil
	}
	lat, latErr := strconv.ParseFloat(latText, 64)
	lon, lonErr := strconv.ParseFloat(lonText, 64)
	if latErr != nil || lonErr != nil || lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return nil, nil, ErrInvalidHome
	}
	return &lat, &lon, nil
}

func withDefault(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
