package ecobee

import (
	"math/rand"
	"os"
	"testing"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func TestNoop(t *testing.T) {
	i := createContainer()

	i.Noop()
}

func TestWriteData(t *testing.T) {
	i := createContainer()

	err := i.StoreTemperature(rand.Intn(100), rand.Intn(100), rand.Intn(100), "ci-thermostat")

	if err != nil {
		t.FailNow()
	}
}

func TestWriteOutsideTemperature(t *testing.T) {
	i := createContainer()

	err := i.StoreCurrentOutsideTemperature(float64(rand.Intn(100))/10.0, "ci-thermostat")

	if err != nil {
		t.FailNow()
	}
}

func createContainer() InfluxContainer {

	c := influxdb2.NewClient(os.Getenv("INFLUXDB_URI"), os.Getenv("INFLUXDB_TOKEN"))
	defer c.Close()

	// Create InfluxContainer struct to hold influxdb config info
	// Also stores the client so we can write tests against it
	// This is optional
	ic := InfluxContainer{
		Client: c,
		Bucket: os.Getenv("INFLUXDB_BUCKET"),
		Org:    os.Getenv("INFLUXDB_ORG"),
	}

	return ic
}
