package ecobee

import (
	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/prometheus/common/log"
)

// InfluxContainer holds our influxdb config
type InfluxContainer struct {
	Client influxdb2.Client
	Bucket string
	Org    string
}

// Noop is for development so the app will compile without calling real influx funcs
func (c InfluxContainer) Noop() {

}

// StoreTemperature accepts a desired heat (dh), desired cool (dc), an actual temperature (at), and the thermostat name (n)
// Then writes to influx
// This function converts an incoming temperature ints to float64 and divides them by 10.0,
// Make sure you pass the raw values received from ecobee
func (c InfluxContainer) StoreTemperature(dh, dc, at int, n string) error {

	w := c.Client.WriteAPIBlocking(c.Org, c.Bucket)

	p := influxdb2.NewPoint("stat",
		map[string]string{"unit": "temperature", "ecobee_thermostat_name": n},
		map[string]interface{}{"ecobee_desired_heat": float64(dh / 10.0), "ecobee_desired_cool": float64(dc / 10.0), "ecobee_actual_temperature": float64(at / 10.0)},
		time.Now())

	err := w.WritePoint(context.Background(), p)

	if err != nil {
		log.Error("Error writing point to influx ", err.Error())
		return err
	}

	log.Info("Wrote information to influx")

	return nil
}

// StoreCurrentOutsideTemperature writes current temperature information to InfluxDB
func (c InfluxContainer) StoreCurrentOutsideTemperature(t float64, n string) error {
	w := c.Client.WriteAPI(c.Org, c.Bucket)

	p := influxdb2.NewPoint("stat",
		map[string]string{"unit": "temperature", "ecobee_thermostat_name": n},
		map[string]interface{}{"current_outside_temperature": t},
		time.Now())

	w.WritePoint(p)
	w.Flush()

	log.Info("Wrote outside temperature information to influx")

	return nil
}
