package ultrasonic

import (
	"context"
	"runtime"
	"testing"
	"time"

	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/testutils/inject"
	"go.viam.com/test"
)

func TestConfigValidation(t *testing.T) {
	validConfig := Config{
		TriggerPin:    "10",
		EchoInterrupt: "8",
		Board:         "board",
		TimeoutMs:     1000,
	}

	invalidConfigNoBoard := Config{
		TriggerPin:    "10",
		EchoInterrupt: "8",
	}

	invalidConfigNoTriggerPin := Config{
		EchoInterrupt: "8",
		Board:         "board",
	}

	invalidConfigNoEchoInterrupt := Config{
		TriggerPin: "10",
		Board:      "board",
	}

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{"ValidConfig", validConfig, false},
		{"NoBoard", invalidConfigNoBoard, true},
		{"NoTriggerPin", invalidConfigNoTriggerPin, true},
		{"NoEchoInterrupt", invalidConfigNoEchoInterrupt, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.config.Validate("test")
			if tt.wantErr {
				test.That(t, err, test.ShouldNotBeNil)
				test.That(t, err.Error(), test.ShouldContainSubstring, "Error validating")
			} else {
				test.That(t, err, test.ShouldBeNil)
			}
		})
	}
}

func TestSensorLifecycle(t *testing.T) {
	deps := make(resource.Dependencies)
	mockGPIOPin := &inject.GPIOPin{
		SetFunc: func(ctx context.Context, high bool, _ map[string]interface{}) error {
			return nil
		},
	}
	depBoard := &inject.Board{
		GPIOPinByNameFunc: func(name string) (board.GPIOPin, error) {
			return mockGPIOPin, nil
		},
		DigitalInterruptByNameFunc: func(name string) (board.DigitalInterrupt, error) {
			return &inject.DigitalInterrupt{}, nil
		},
	}
	resourceName := resource.NewName(board.API, "board")
	deps[resourceName] = depBoard

	conf := resource.NewEmptyConfig(resource.NewName(sensor.API, "testSensor"), ModelSensor)
	conf.ConvertedAttributes = &Config{
		TriggerPin:    "10",
		EchoInterrupt: "8",
		Board:         "board",
		TimeoutMs:     1000,
	}

	logger := logging.NewTestLogger(t)

	// Test newSensor
	goRoutinesStart := runtime.NumGoroutine() // used to track start vs. ending num goroutines
	s, err := newSensor(context.Background(), deps, conf, logger)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, s, test.ShouldNotBeNil)

	us, ok := s.(*usSensor)
	test.That(t, ok, test.ShouldBeTrue)

	// Test Close
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = us.Close(ctx)
	test.That(t, err, test.ShouldBeNil)

	// Test goroutine leaks
	goRoutinesEnd := runtime.NumGoroutine()
	test.That(t, goRoutinesStart, test.ShouldEqual, goRoutinesEnd)
}
