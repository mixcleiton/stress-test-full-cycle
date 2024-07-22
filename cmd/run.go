package cmd

import (
	"github.com/mixcleiton/stress-test-full-cycle/internal/stresstest"
)

func Execute(config stresstest.Config) {
	stresstest.ExecutarTestes(config)
}
