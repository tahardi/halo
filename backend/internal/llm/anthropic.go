package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	model     = "claude-sonnet-5-5"
	maxTokens = 16000
)

type Anthropic struct {
	client anthropic.Client
}

func NewAnthropic(apiKey string, opts ...option.RequestOption) *Anthropic {
	all := append([]option.RequestOption{option.WithAPIKey(apiKey)}, opts...)
	return &Anthropic{client: anthropic.NewClient(all...)}
}

func (a *Anthropic) Name() string {
	return model
}

func (a *Anthropic) Structured(
	ctx context.Context,
	system string,
	user string,
	schema map[string]any,
	out any,
) error {
	msg, err := a.client.Beta.Messages.New(
		ctx,
		anthropic.BetaMessageNewParams{
			Model:     model,
			MaxTokens: maxTokens,
			System:    []anthropic.BetaTextBlockParam{{Text: system}},
			Messages: []anthropic.BetaMessageParam{
				anthropic.NewBetaUserMessage(anthropic.NewBetaTextBlock(user)),
			},
			OutputConfig: anthropic.BetaOutputConfigParam{
				Effort: anthropic.BetaOutputConfigEffortLow,
				Format: anthropic.BetaJSONOutputFormatParam{Schema: schema},
			},
			Fallbacks: anthropic.BetaFallbacksParamOfDefault(),
			Betas:     []anthropic.AnthropicBeta{anthropic.AnthropicBetaServerSideFallback2026_07_01},
		},
	)
	if err != nil {
		return fmt.Errorf("calling claude: %w", err)
	}
	if msg.StopReason == anthropic.BetaStopReasonRefusal {
		return ErrRefused
	}
	var text strings.Builder
	for _, block := range msg.Content {
		if b, ok := block.AsAny().(anthropic.BetaTextBlock); ok {
			text.WriteString(b.Text)
		}
	}
	if err := json.Unmarshal([]byte(text.String()), out); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidResponse, err)
	}
	return nil
}
