package metrics

import (
	"time"
	"fmt"

	"github.com/2Bot/2Bot-Discord-Bot/config"
	"github.com/influxdata/influxdb/client/v2"
)

var InfluxClient client.Client

func init() {
	var err error
	InfluxClient, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     config.Conf.Influx.URL,
		Username: config.Conf.Influx.Username,
		Password: config.Conf.Influx.Password,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
}

func NewMetric(database, name string, tags map[string]string, fields map[string]interface{}) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database: database,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	point, err := client.NewPoint(name, tags, fields, time.Now())
	if err != nil {
		return err
	}

	bp.AddPoint(point)

	return InfluxClient.Write(bp)
}