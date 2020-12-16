package ecobee

import (
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

// StoreTemperature accepts a desired heat (dh), desired cool (dc), and an actual temperature (at)
// Then writes to influx
func (c InfluxContainer) StoreTemperature(dh, dc, at int) error {

	w := c.Client.WriteAPI(c.Org, c.Bucket)

	p := influxdb2.NewPoint("stat", map[string]string{"unit": "temperature"}, map[string]interface{}{"desired_heat": float64(dh / 10.0), "desired_cool": float64(dc / 10.0), "actual_temperature": float64(at / 10.0)}, time.Now())

	w.WritePoint(p)
	w.Flush()

	log.Info("Wrote information to influx")

	return nil
}
