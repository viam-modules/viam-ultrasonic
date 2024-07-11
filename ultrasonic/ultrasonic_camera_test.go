package ultrasonic

import (
	"context"
	"testing"
	"time"

	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/testutils/inject"
	"go.viam.com/test"
)

func TestUltrasonicCameraLifecycle(t *testing.T) {
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

	conf := resource.NewEmptyConfig(resource.NewName(camera.API, "testUltrasonicCamera"), ModelCamera)
	conf.ConvertedAttributes = &Config{
		TriggerPin:    "10",
		EchoInterrupt: "8",
		Board:         "board",
		TimeoutMs:     1000,
	}

	logger := logging.NewTestLogger(t)

	// Test newCamera
	cam, err := newCamera(context.Background(), deps, conf, logger)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, cam, test.ShouldNotBeNil)

	// Test Close
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = cam.Close(ctx)
	test.That(t, err, test.ShouldBeNil)
}
