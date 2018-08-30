package metrics

import (
	"time"
	"fmt"

	"github.com/2Bot/2Bot-Discord-Bot/config"
	"github.com/influxdata/influxdb/client/v2"
)

var InfluxClient client.Client

func New() error {
	var err error
	InfluxClient, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     config.Conf.Influx.URL,
		Username: config.Conf.Influx.Username,
		Password: config.Conf.Influx.Password,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	q := client.Query{
		Command:  "CREATE DATABASE \"2Bot\";",
		Database: "2Bot",
	}
	if response, err := InfluxClient.Query(q); err == nil {
		if response.Error() != nil {
			fmt.Println(response.Error())
			return err
		}
	} else {
		fmt.Println(err)
		return err
	}
	return nil
}

func NewMetric(name string, tags map[string]string, fields map[string]interface{}) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database: "2Bot",
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

	err = InfluxClient.Write(bp)
	if err != nil {
		fmt.Println(err)
	}
	return err
}