# `viam-ultrasonic`

This module implements the [`"rdk:component:sensor"` API](https://docs.viam.com/components/sensor/) and [`"rdk:component:camera"` API](https://docs.viam.com/components/camera/) to integrate the HC-S204 ultrasonic distance sensor into your machine.

Two models are provided:
* `viam:ultrasonic:sensor` - Configure as a sensor to access the sensor method GetReadings().
* `viam:ultrasonic:camera` - When configured as a camera, you can use the camera method GetPointCloud(), rather than GetReadings().

## Configure your ultrasonic component

Navigate to the **CONFIGURE** tab of your machine's page in [the Viam app](https://app.viam.com), searching for `ultrasonic` and selecting one of the above models.

Fill in the attributes as applicable to your sensor, according to the example below.

```json
{
  "components": [
    {
      "name": "<ultrasonic-sensor-name>",
      "model": "viam:ultrasonic:sensor",
      "type": "sensor",
      "namespace": "rdk",
      "attributes": {
        "trigger_pin": "<pin-number>",
        "echo_interrupt_pin": "<pin-number>",
        "board": "<your-board-name>",
        "timeout_ms": <int>
      },
      "depends_on": []
    }
  ]
}
```

Similarly for the ultrasonic camera:
```json
{
  "components": [
    {
      "name": "<ultrasonic-camera-name>",
      "model": "viam:ultrasonic:camera",
      "type": "camera",
      "namespace": "rdk",
      "attributes": {
        "trigger_pin": "<pin-number>",
        "echo_interrupt_pin": "<pin-number>",
        "board": "<your-board-name>",
        "timeout_ms": <int>
      },
      "depends_on": []
    }
  ]
}
```

The following attributes are available for `ultrasonic` sensors:

<!-- prettier-ignore -->
| Attribute | Type | Required? | Description |
| --------- | ---- | --------- | ----------- |
| `trigger_pin` | string | **Required** | The pin number of the [board's](https://docs.viam.com/components/board/) GPIO pin that you have wired [the ultrasonic's trigger pin](https://www.sparkfun.com/products/15569) to. |
| `echo_interrupt_pin` | string | **Required** | The pin number of the pin [the ultrasonic's echo pin](https://www.sparkfun.com/products/15569) is wired to on the board. If you have already created a [digital interrupt](https://docs.viam.com/components/board/#digital_interrupts) for this pin in the [board's configuration](https://docs.viam.com/components/board/), use that digital interrupt's `name` instead. |
| `board`  | string | **Required** | The `name` of the [board](https://docs.viam.com/components/board/) the ultrasonic is wired to. |
| `timeout_ms`  | int | Optional | Time to wait in milliseconds before timing out of requesting to get ultrasonic distance readings. <br> Default: `1000`. |
