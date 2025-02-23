// Package ultrasonic implements an ultrasonic sensor based of the yahboom ultrasonic sensor
package ultrasonic

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"sync"
	"time"

	"go.viam.com/rdk/components/board"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	rdkutils "go.viam.com/utils"
)

var ModelSensor = resource.NewModel("viam", "ultrasonic", "sensor")

// Config is used for converting config attributes.
type Config struct {
	TriggerPin    string `json:"trigger_pin"`
	EchoInterrupt string `json:"echo_interrupt_pin"`
	Board         string `json:"board"`
	TimeoutMs     uint   `json:"timeout_ms,omitempty"`
}

// Validate ensures all parts of the config are valid.
func (conf *Config) Validate(path string) ([]string, error) {
	var deps []string
	if len(conf.Board) == 0 {
		return nil, resource.NewConfigValidationFieldRequiredError(path, "board")
	}
	deps = append(deps, conf.Board)
	if len(conf.TriggerPin) == 0 {
		return nil, resource.NewConfigValidationFieldRequiredError(path, "trigger pin")
	}
	if len(conf.EchoInterrupt) == 0 {
		return nil, resource.NewConfigValidationFieldRequiredError(path, "echo interrupt pin")
	}
	return deps, nil
}

func init() {
	resource.RegisterComponent(
		sensor.API,
		ModelSensor,
		resource.Registration[sensor.Sensor, *Config]{
			Constructor: newSensor,
		})
}

// NewSensor creates and configures a new ultrasonic sensor.
func newSensor(
	ctx context.Context,
	deps resource.Dependencies,
	conf resource.Config,
	logger logging.Logger,
) (sensor.Sensor, error) {
	nativeConf, err := resource.NativeConfig[*Config](conf)
	if err != nil {
		return nil, err
	}

	s := &usSensor{
		Named:  conf.ResourceName().AsNamed(),
		logger: logger,
		config: nativeConf,
	}
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	s.cancelCtx = cancelCtx
	s.cancelFunc = cancelFunc

	s.board, err = board.FromDependencies(deps, nativeConf.Board)
	if err != nil {
		return nil, fmt.Errorf("ultrasonic: cannot find board %q", nativeConf.Board)
	}

	s.timeoutMs = 1000 // default to 1 sec
	if nativeConf.TimeoutMs > 0 {
		s.timeoutMs = nativeConf.TimeoutMs
	}

	s.ticksChan = make(chan board.Tick, 2)

	// Set the trigger pin to low, so it's ready for later.
	triggerPin, err := s.board.GPIOPinByName(nativeConf.TriggerPin)
	if err != nil {
		return nil, fmt.Errorf("%w. ultrasonic: cannot grab gpio %q", err, nativeConf.TriggerPin)
	}
	if err := triggerPin.Set(ctx, false, nil); err != nil {
		return nil, fmt.Errorf("%w. ultrasonic: cannot set trigger pin to low", err)
	}

	s.f, _ = os.Create("/tmp/ultrasonicdata.txt")

	return s, nil
}

// usSensor ultrasonic sensor.
type usSensor struct {
	resource.Named
	resource.AlwaysRebuild
	mu         sync.Mutex
	config     *Config
	board      board.Board
	ticksChan  chan board.Tick
	timeoutMs  uint
	cancelCtx  context.Context
	cancelFunc func()
	logger     logging.Logger
	f          *os.File
}

func (s *usSensor) namedError(err error) error {
	return fmt.Errorf(
		"%w. Error in ultrasonic sensor with name %s: ", err, s.Name(),
	)
}

// Readings returns the calculated distance.
func (s *usSensor) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Grab the 2 pins from the board. We don't just get these once during setup, in case the board
	// reconfigures itself because someone decided to rewire things.
	echoInterrupt, err := s.board.DigitalInterruptByName(s.config.EchoInterrupt)
	if err != nil {
		return nil, fmt.Errorf("ultrasonic: cannot grab digital interrupt %q", s.config.EchoInterrupt)
	}
	triggerPin, err := s.board.GPIOPinByName(s.config.TriggerPin)
	if err != nil {
		return nil, fmt.Errorf("%w. ultrasonic: cannot grab gpio %q", err, s.config.TriggerPin)
	}

	s.board.StreamTicks(ctx, []board.DigitalInterrupt{echoInterrupt}, s.ticksChan, nil)
	s.logger.Debug("returning stream ticks")

	// we send a high and a low to the trigger pin 10 microseconds
	// apart to signal the sensor to begin sending the sonic pulse
	if err := triggerPin.Set(ctx, true, nil); err != nil {
		return nil, s.namedError(fmt.Errorf("%w. ultrasonic cannot set trigger pin to high", err))
	}
	rdkutils.SelectContextOrWait(ctx, time.Microsecond*10)
	if err := triggerPin.Set(ctx, false, nil); err != nil {
		return nil, s.namedError(fmt.Errorf("%w. ultrasonic cannot set trigger pin to low", err))
	}

	// the first signal from the interrupt indicates that the sonic
	// pulse has been sent and the second indicates that the echo has been received
	var tick board.Tick
	ticks := make([]board.Tick, 2)

	for i := 0; i < 2; i++ {
		var signalStr string
		if i == 0 {
			signalStr = "sound pulse was emitted"
		} else {
			signalStr = "echo was received"
		}
		select {
		case tick = <-s.ticksChan:
			ticks[i] = tick
		case <-s.cancelCtx.Done():
			return nil, s.namedError(errors.New("ultrasonic: context canceled"))
		case <-time.After(time.Millisecond * time.Duration(s.timeoutMs)):
			return nil, s.namedError(fmt.Errorf("timed out waiting for signal that %s", signalStr))
		}
	}
	timeB := ticks[0].TimestampNanosec
	timeA := ticks[1].TimestampNanosec
	// we calculate the distance to the nearest object based
	// on the time interval between the sound and its echo
	// and the speed of sound (343 m/s)
	secondsElapsed := float64(timeA-timeB) / math.Pow10(9)
	distMeters := secondsElapsed * 343.0 / 2.0
	return map[string]interface{}{"distance": distMeters}, nil
}

// Close remove interrupt callback of ultrasonic sensor.
func (s *usSensor) Close(ctx context.Context) error {
	fmt.Println("closing modular ultrasonic sensor")
	s.cancelFunc()
	return nil
}
