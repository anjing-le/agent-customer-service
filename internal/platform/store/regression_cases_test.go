package store

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

type regressionCase struct {
	Name                 string   `json:"name"`
	Content              string   `json:"content"`
	Generator            string   `json:"generator"`
	ExpectEngine         string   `json:"expectEngine"`
	ExpectFallbackReason string   `json:"expectFallbackReason"`
	ExpectEvidenceIDs    []string `json:"expectEvidenceIds"`
	ExpectTraceEvidence  int      `json:"expectTraceEvidenceCount"`
	ExpectGap            bool     `json:"expectGap"`
	ExpectTraceStrategy  string   `json:"expectTraceStrategy"`
	ExpectModelAttempted bool     `json:"expectModelAttempted"`
	ExpectModelFallback  bool     `json:"expectModelFallback"`
	ExpectModel          string   `json:"expectModel"`
}

type regressionReplyGenerator struct {
	reply string
	model string
	err   error
}

func (r regressionReplyGenerator) GenerateReply(context.Context, ReplyRequest) (ReplyGeneration, error) {
	if r.err != nil {
		return ReplyGeneration{}, r.err
	}
	return ReplyGeneration{Content: r.reply, Model: r.model}, nil
}

func TestAgentRegressionCases(t *testing.T) {
	payload, err := os.ReadFile("testdata/agent_regression_cases.json")
	if err != nil {
		t.Fatalf("read regression cases: %v", err)
	}

	var cases []regressionCase
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatalf("decode regression cases: %v", err)
	}
	if len(cases) == 0 {
		t.Fatal("expected at least one regression case")
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			st := NewSeedStore(regressionOptions(tc)...)
			result, err := st.SendMessage("conv_demo_refund", tc.Content)
			if err != nil {
				t.Fatalf("send message: %v", err)
			}

			if result.AgentMessage.Engine != tc.ExpectEngine {
				t.Fatalf("engine mismatch: got %q want %q", result.AgentMessage.Engine, tc.ExpectEngine)
			}
			if result.AgentMessage.FallbackReason != tc.ExpectFallbackReason {
				t.Fatalf("fallback mismatch: got %q want %q", result.AgentMessage.FallbackReason, tc.ExpectFallbackReason)
			}
			if (result.Gap != nil) != tc.ExpectGap {
				t.Fatalf("gap mismatch: got %#v want gap=%v", result.Gap, tc.ExpectGap)
			}
			assertStringSet(t, "evidence ids", result.AgentMessage.EvidenceIDs, tc.ExpectEvidenceIDs)
			assertTrace(t, result.AgentMessage.Trace, tc)
		})
	}
}

func regressionOptions(tc regressionCase) []Option {
	switch tc.Generator {
	case "success":
		return []Option{WithReplyGenerator(regressionReplyGenerator{reply: "这是基于发票知识的回归模型回复。", model: "regression-model"})}
	case "failure":
		return []Option{WithReplyGenerator(regressionReplyGenerator{err: errors.New("regression model unavailable")})}
	default:
		return nil
	}
}

func assertStringSet(t *testing.T, label string, got, want []string) {
	t.Helper()
	got = append([]string(nil), got...)
	want = append([]string(nil), want...)
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("%s mismatch: got %#v want %#v", label, got, want)
	}
}

func assertTrace(t *testing.T, trace *AgentTrace, tc regressionCase) {
	t.Helper()
	if trace == nil {
		t.Fatal("expected agent trace")
	}
	if trace.Strategy != tc.ExpectTraceStrategy {
		t.Fatalf("trace strategy mismatch: got %q want %q", trace.Strategy, tc.ExpectTraceStrategy)
	}
	if trace.EvidenceCount != tc.ExpectTraceEvidence {
		t.Fatalf("trace evidence count mismatch: got %d want %d", trace.EvidenceCount, tc.ExpectTraceEvidence)
	}
	if trace.ModelAttempted != tc.ExpectModelAttempted {
		t.Fatalf("trace model attempted mismatch: got %v want %v", trace.ModelAttempted, tc.ExpectModelAttempted)
	}
	if trace.ModelFallback != tc.ExpectModelFallback {
		t.Fatalf("trace model fallback mismatch: got %v want %v", trace.ModelFallback, tc.ExpectModelFallback)
	}
	if tc.ExpectModel != "" && trace.Model != tc.ExpectModel {
		t.Fatalf("trace model mismatch: got %q want %q", trace.Model, tc.ExpectModel)
	}
	if tc.ExpectModelFallback && trace.ModelFallbackReason == "" {
		t.Fatal("expected model fallback reason")
	}
}
