package store

import (
	"encoding/json"
	"os"
	"testing"
)

type qualityEvalCase struct {
	Name                  string `json:"name"`
	Content               string `json:"content"`
	ExpectMinScore        int    `json:"expectMinScore"`
	ExpectEvidenceAnswers int    `json:"expectEvidenceAnswers"`
	ExpectSafeFallbacks   int    `json:"expectSafeFallbacks"`
	ExpectHumanTransfers  int    `json:"expectHumanTransfers"`
}

func TestQualityEvaluationCases(t *testing.T) {
	payload, err := os.ReadFile("testdata/quality_eval_cases.json")
	if err != nil {
		t.Fatalf("read quality cases: %v", err)
	}
	var cases []qualityEvalCase
	if err := json.Unmarshal(payload, &cases); err != nil {
		t.Fatalf("decode quality cases: %v", err)
	}
	if len(cases) == 0 {
		t.Fatal("expected quality evaluation cases")
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			st := NewSeedStore()
			result, err := st.SendMessage("conv_demo_refund", tc.Content)
			if err != nil {
				t.Fatalf("send message: %v", err)
			}
			summary := qualitySummary([]Message{result.AgentMessage}, nil, nil, nil)
			if summary.Score < tc.ExpectMinScore {
				t.Fatalf("quality score too low: got %d want >= %d", summary.Score, tc.ExpectMinScore)
			}
			if summary.EvidenceAnswers != tc.ExpectEvidenceAnswers {
				t.Fatalf("evidence answer mismatch: got %d want %d", summary.EvidenceAnswers, tc.ExpectEvidenceAnswers)
			}
			if summary.SafeFallbacks != tc.ExpectSafeFallbacks {
				t.Fatalf("safe fallback mismatch: got %d want %d", summary.SafeFallbacks, tc.ExpectSafeFallbacks)
			}
			if summary.HumanTransfers != tc.ExpectHumanTransfers {
				t.Fatalf("human transfer mismatch: got %d want %d", summary.HumanTransfers, tc.ExpectHumanTransfers)
			}
			if len(summary.Notes) == 0 {
				t.Fatal("expected quality notes for human review")
			}
		})
	}
}
