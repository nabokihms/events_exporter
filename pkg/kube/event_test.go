package kube

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nabokihms/events_exporter/pkg/vault"
)

func TestEventToSample(t *testing.T) {
	tests := []struct {
		Name         string
		InputEvent   v1.Event
		OutputSample vault.Sample
	}{
		{
			Name:       "Empty",
			InputEvent: v1.Event{},
			OutputSample: vault.Sample{
				Value:     0,
				Labels:    []string{"", "", "", "", "", "", "", "", "", ""},
				Timestamp: metav1.Time{}.Local(),
			},
		},
		{
			Name: "Too long message",
			InputEvent: v1.Event{
				Message: strings.Repeat("toolong", 10000),
			},
			OutputSample: vault.Sample{
				Value:     0,
				Labels:    []string{"", "", "", "", "", "", "", "", "", strings.Repeat("toolong", 10000)[:200]},
				Timestamp: metav1.Time{}.Local(),
			},
		},
		{
			Name: "With count",
			InputEvent: v1.Event{
				Count: 5,
			},
			OutputSample: vault.Sample{
				Value:     5,
				Labels:    []string{"", "", "", "", "", "", "", "", "", ""},
				Timestamp: metav1.Time{}.Local(),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			sample := EventToSample(&tc.InputEvent)
			require.Equal(t, tc.OutputSample, sample)
		})
	}
}
