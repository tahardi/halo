package llm_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tahardi/halo/backend/internal/llm"
)

type reply struct {
	Value string `json:"value"`
}

var testSchema = map[string]any{
	"type":                 "object",
	"additionalProperties": false,
	"required":             []string{"value"},
	"properties":           map[string]any{"value": map[string]any{"type": "string"}},
}

func messageJSON(t *testing.T, stopReason string, text string) string {
	t.Helper()
	return messageContentJSON(t, stopReason, []map[string]any{{"type": "text", "text": text}})
}

func messageContentJSON(t *testing.T, stopReason string, content []map[string]any) string {
	t.Helper()
	body := map[string]any{
		"id":            "msg_1",
		"type":          "message",
		"role":          "assistant",
		"model":         "claude-sonnet-5-5",
		"content":       content,
		"stop_reason":   stopReason,
		"stop_sequence": nil,
		"usage":         map[string]any{"input_tokens": 1, "output_tokens": 1},
	}
	return string(mustJSON(t, body))
}

func TestAnthropic_Structured(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		response string
		want     reply
		wantErr  error
		wantAny  bool
	}{
		{
			name:     "happy path - unmarshals reply",
			status:   http.StatusOK,
			response: messageJSON(t, "end_turn", `{"value":"Maggie"}`),
			want:     reply{Value: "Maggie"},
		},
		{
			name:   "happy path - skips thinking and joins text blocks",
			status: http.StatusOK,
			response: messageContentJSON(t, "end_turn", []map[string]any{
				{"type": "thinking", "thinking": "", "signature": "sig"},
				{"type": "text", "text": `{"value":`},
				{"type": "text", "text": `"Maggie"}`},
			}),
			want: reply{Value: "Maggie"},
		},
		{
			name:     "error - truncated at max tokens",
			status:   http.StatusOK,
			response: messageJSON(t, "max_tokens", `{"value":"Mag`),
			wantErr:  llm.ErrInvalidResponse,
		},
		{
			name:     "error - refusal",
			status:   http.StatusOK,
			response: messageJSON(t, "refusal", ""),
			wantErr:  llm.ErrRefused,
		},
		{
			name:     "error - reply not json",
			status:   http.StatusOK,
			response: messageJSON(t, "end_turn", "not json"),
			wantErr:  llm.ErrInvalidResponse,
		},
		{
			name:     "error - server error",
			status:   http.StatusInternalServerError,
			response: `{"type":"error","error":{"type":"api_error","message":"boom"}}`,
			wantAny:  true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// given
			var gotBody map[string]any
			var gotBeta string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				raw, _ := io.ReadAll(r.Body)
				_ = json.Unmarshal(raw, &gotBody)
				gotBeta = r.Header.Get("anthropic-beta")
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.response)
			}))
			defer srv.Close()
			client := llm.NewAnthropic("test", option.WithBaseURL(srv.URL), option.WithMaxRetries(0))
			var got reply

			// when
			err := client.Structured(t.Context(), "system text", "user text", testSchema, &got)

			// then
			assert.Equal(t, "claude-sonnet-5-5", gotBody["model"])
			assert.Equal(t, "default", gotBody["fallbacks"])
			assert.Contains(t, gotBeta, "server-side-fallback-2026-07-01")
			outputConfig, ok := gotBody["output_config"].(map[string]any)
			require.True(t, ok)
			assert.Equal(t, "low", outputConfig["effort"])
			format, ok := outputConfig["format"].(map[string]any)
			require.True(t, ok)
			assert.Equal(t, "json_schema", format["type"])
			assert.NotNil(t, format["schema"])
			assert.Contains(t, string(mustJSON(t, gotBody["messages"])), "user text")
			assert.Contains(t, string(mustJSON(t, gotBody["system"])), "system text")
			switch {
			case tc.wantErr != nil:
				require.ErrorIs(t, err, tc.wantErr)
			case tc.wantAny:
				require.Error(t, err)
				require.NotErrorIs(t, err, llm.ErrRefused)
			default:
				require.NoError(t, err)
				assert.Equal(t, tc.want, got)
			}
		})
	}
}

func TestAnthropic_Name(t *testing.T) {
	// given
	client := llm.NewAnthropic("test")

	// when
	got := client.Name()

	// then
	assert.Equal(t, "claude-sonnet-5-5", got)
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}
