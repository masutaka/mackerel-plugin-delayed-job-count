mackerel-plugin-delayed-job
===========================

delayed\_job custom metrics plugin for mackerel.io agent.

## Synopsis

    mackerel-plugin-delayed-job -dsn=<dataSourceName>

## Requirements

- [delayed_job](https://rubygems.org/gems/delayed_job)

## Example of mackerel-agent.conf


    [plugin.metrics.delayed_job_count]
    command = "/path/to/mackerel-plugin-delayed-job -dsn=<dataSourceName>"
