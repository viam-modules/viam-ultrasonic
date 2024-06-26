// package main is a module with correction-station component
package main

import (
	"context"

	"viamultrasonic/ultrasonic"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/module"
	"go.viam.com/utils"
)

func main() {
	utils.ContextualMain(mainWithArgs, logging.NewLogger("ultrasonic"))
}

func mainWithArgs(ctx context.Context, args []string, logger logging.Logger) error {
	module, err := module.NewModuleFromArgs(ctx, logger)

	if err != nil {
		return err
	}
	module.AddModelFromRegistry(ctx, sensor.API, ultrasonic.Model)

	err = module.Start(ctx)
	defer module.Close(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
