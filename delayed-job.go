package main

import (
	"database/sql"
	"flag"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mackerelio/go-mackerel-plugin-helper"
)

var graphdef = map[string](mackerelplugin.Graphs){
	"delayed_job": {
		Label: "delayed_job",
		Unit:  "integer",
		Metrics: [](mackerelplugin.Metrics){
			{Name: "queued", Label: "Queued Job Count"},
			{Name: "processing", Label: "Processing Job Count"},
			{Name: "failed", Label: "Failed Job Count"},
		},
	},
}

type DelayedJobPlugin struct {
	driverName     string
	dataSourceName string
}

func (dj DelayedJobPlugin) FetchMetrics() (map[string]interface{}, error) {

	fmt.Printf("driverName (%s), dataSourceName (%s)\n", dj.driverName, dj.dataSourceName)

	db, err := sql.Open(dj.driverName, dj.dataSourceName)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	const query string = `
SELECT count FROM (
  -- queued job
  SELECT 1 AS id, COUNT(*) AS count FROM delayed_jobs WHERE failed_at IS NULL AND locked_by IS NULL
  UNION ALL
  -- processing job
  SELECT 2 AS id, COUNT(*) AS count FROM delayed_jobs WHERE failed_at IS NULL AND locked_by IS NOT NULL
  UNION ALL
  -- failed job
  SELECT 3 AS id, COUNT(*) AS count FROM delayed_jobs WHERE failed_at IS NOT NULL
) AS t ORDER BY t.id;
`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rows.Next()

	var queuedCount, processingCount, failedCount float64

	err = rows.Scan(&queuedCount)
	if err != nil {
		return nil, err
	}

	rows.Next()

	err = rows.Scan(&processingCount)
	if err != nil {
		return nil, err
	}

	rows.Next()

	err = rows.Scan(&failedCount)
	if err != nil {
		return nil, err
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"queued":     queuedCount,
		"processing": processingCount,
		"failed":     failedCount,
	}, nil
}

func (dj DelayedJobPlugin) GraphDefinition() map[string](mackerelplugin.Graphs) {
	return graphdef
}

func main() {
	optName := flag.String("name", "mysql", "driverName")
	optDSN := flag.String("dsn", "", "dataSourceName")
	flag.Parse()

	var delayed_job DelayedJobPlugin

	delayed_job.driverName = *optName
	delayed_job.dataSourceName = *optDSN

	helper := mackerelplugin.NewMackerelPlugin(delayed_job)
	helper.Run()
}
