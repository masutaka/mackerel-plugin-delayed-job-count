package main

import (
	"database/sql"
	"flag"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	mp "github.com/mackerelio/go-mackerel-plugin-helper"
)

func main() {
	optName := flag.String("name", "mysql", "driverName")
	optDSN := flag.String("dsn", "", "dataSourceName")
	optMetricKeyPrefix := flag.String("metric-key-prefix", "delayed_job", "Metric Key Prefix")
	flag.Parse()

	var delayedJobCount DelayedJobCountPlugin

	delayedJobCount.driverName = *optName
	delayedJobCount.dataSourceName = *optDSN
	delayedJobCount.prefix = *optMetricKeyPrefix

	helper := mp.NewMackerelPlugin(delayedJobCount)
	helper.Run()
}

// DelayedJobCountPlugin mackerel plugin for delayed_job
type DelayedJobCountPlugin struct {
	driverName     string
	dataSourceName string
	prefix         string
}

// FetchMetrics interface for PluginWithPrefix
func (p DelayedJobCountPlugin) FetchMetrics() (map[string]interface{}, error) {
	db, err := sql.Open(p.driverName, p.dataSourceName)
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

// GetTotalProcessedCount is total processed count
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

// NthAutoIncrement is position in columns
func NthAutoIncrement(columns []string) int {
	for key, value := range columns {
		if value == "Auto_increment" {
			return key
		}
	}

	return -1
}

// GetOtherCounts is some counts except the total processed count
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

// GraphDefinition interface for PluginWithPrefix
func (p DelayedJobCountPlugin) GraphDefinition() map[string](mp.Graphs) {
	labelPrefix := strings.Title(p.prefix)

	// metric value structure
	var graphdef = map[string](mp.Graphs){
		"count": {
			Label: (labelPrefix + " Count"),
			Unit:  "integer",
			Metrics: [](mp.Metrics){
				{Name: "processed", Label: "Processed Job Count", Diff: true},
				{Name: "queued", Label: "Queued Job Count", Type: "uint64"},
				{Name: "processing", Label: "Processing Job Count", Type: "uint64"},
				{Name: "failed", Label: "Failed Job Count", Type: "uint64"},
			},
		},
	}

	return graphdef
}

// MetricKeyPrefix interface for PluginWithPrefix
func (p DelayedJobCountPlugin) MetricKeyPrefix() string {
	if p.prefix == "" {
		p.prefix = "delayed_job"
	}
	return p.prefix
}
