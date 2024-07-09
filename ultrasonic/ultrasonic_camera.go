// Package ultrasonic provides an implementation for an ultrasonic sensor wrapped as a camera
package ultrasonic

import (
	"context"
	"errors"
	"image"

	"go.viam.com/rdk/components/camera"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	pointcloud "go.viam.com/rdk/pointcloud"
	"go.viam.com/rdk/resource"
)

var ModelCamera = resource.NewModel("viam", "ultrasonic", "camera")

type ultrasonicWrapper struct {
	usSensor sensor.Sensor
}

func init() {
	resource.RegisterComponent(
		camera.API,
		ModelCamera,
		resource.Registration[camera.Camera, *Config]{
			Constructor: newCamera,
		})
}

func newCamera(ctx context.Context, deps resource.Dependencies, conf resource.Config,
	logger logging.Logger,
) (camera.Camera, error) {
	usSensor, err := newSensor(ctx, deps, conf, logger)
	if err != nil {
		return nil, err
	}
	return cameraFromSensor(ctx, conf.ResourceName(), usSensor, logger)
}

func cameraFromSensor(ctx context.Context, name resource.Name, usSensor sensor.Sensor, logger logging.Logger) (camera.Camera, error) {
	usWrapper := ultrasonicWrapper{usSensor: usSensor}

	usVideoSource, err := camera.NewVideoSourceFromReader(ctx, &usWrapper, nil, camera.UnspecifiedStream)
	if err != nil {
		return nil, err
	}

	return camera.FromVideoSource(name, usVideoSource, logger), nil
}

// NextPointCloud queries the ultrasonic sensor then returns the result as a pointcloud,
// with a single point at (0, 0, distance).
func (usvs *ultrasonicWrapper) NextPointCloud(ctx context.Context) (pointcloud.PointCloud, error) {
	readings, err := usvs.usSensor.Readings(ctx, nil)
	if err != nil {
		return nil, err
	}
	pcToReturn := pointcloud.New()
	distFloat, ok := readings["distance"].(float64)
	if !ok {
		return nil, errors.New("unable to convert distance to float64")
	}
	basicData := pointcloud.NewBasicData()
	distVector := pointcloud.NewVector(0, 0, distFloat*1000)
	err = pcToReturn.Set(distVector, basicData)
	if err != nil {
		return nil, err
	}

	return pcToReturn, nil
}

// Properties returns the properties of the ultrasonic camera.
func (usvs *ultrasonicWrapper) Properties(ctx context.Context) (camera.Properties, error) {
	return camera.Properties{SupportsPCD: true, ImageType: camera.UnspecifiedStream}, nil
}

// Close closes the underlying ultrasonic sensor and the camera itself.
func (usvs *ultrasonicWrapper) Close(ctx context.Context) error {
	err := usvs.usSensor.Close(ctx)
	return err
}

// Read returns a not yet implemented error, as it is not needed for the ultrasonic camera.
func (usvs *ultrasonicWrapper) Read(ctx context.Context) (image.Image, func(), error) {
	return nil, nil, errors.New("not yet implemented")
}
