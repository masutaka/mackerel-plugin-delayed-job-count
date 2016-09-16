package main

import (
	"database/sql"
	"flag"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mackerelio/go-mackerel-plugin-helper"
)

var graphdef = map[string](mackerelplugin.Graphs){
	"delayed_job": {
		Label: "delayed_job",
		Unit:  "integer",
		Metrics: [](mackerelplugin.Metrics){
			{Name: "processed", Label: "Processed Job Count", Diff: true},
			{Name: "queued", Label: "Queued Job Count", Type: "uint64"},
			{Name: "processing", Label: "Processing Job Count", Type: "uint64"},
			{Name: "failed", Label: "Failed Job Count", Type: "uint64"},
		},
	},
}

type DelayedJobPlugin struct {
	driverName     string
	dataSourceName string
}

func (dj DelayedJobPlugin) FetchMetrics() (map[string]interface{}, error) {
	db, err := sql.Open(dj.driverName, dj.dataSourceName)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	totalProcessedCount, err := GetTotalProcessedCount(db)
	if err != nil {
		return nil, err
	}

	queuedCount, processingCount, failedCount, err := GetOtherCounts(db)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"processed":  totalProcessedCount,
		"queued":     queuedCount,
		"processing": processingCount,
		"failed":     failedCount,
	}, nil
}

func GetTotalProcessedCount(db *sql.DB) (uint64, error) {
	rows, err := db.Query("SHOW TABLE STATUS LIKE 'delayed_jobs'")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return 0, err
	}

	rows.Next()

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	err = rows.Scan(scanArgs...)
	if err != nil {
		return 0, err
	}

	autoIncrement, err := strconv.ParseUint(string(values[NthAutoIncrement(columns)]), 10, 64)
	if err != nil {
		return 0, err
	}

	return autoIncrement - 1, err
}

func NthAutoIncrement(columns []string) int {
	for key, value := range columns {
		if value == "Auto_increment" {
			return key
		}
	}

	return -1
}

func GetOtherCounts(db *sql.DB) (queued uint64, processing uint64, failed uint64, error error) {
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
		return 0, 0, 0, err
	}
	defer rows.Close()

	rows.Next()

	err = rows.Scan(&queued)
	if err != nil {
		return 0, 0, 0, err
	}

	rows.Next()

	err = rows.Scan(&processing)
	if err != nil {
		return 0, 0, 0, err
	}

	rows.Next()

	err = rows.Scan(&failed)
	if err != nil {
		return 0, 0, 0, err
	}

	err = rows.Err()
	if err != nil {
		return 0, 0, 0, err
	}

	return queued, processing, failed, err
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