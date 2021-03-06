mackerel-plugin-delayed-job-count
=================================

[![License](https://img.shields.io/github/license/masutaka/mackerel-plugin-delayed-job-count.svg)][license]
[![GoDoc](https://godoc.org/github.com/masutaka/mackerel-plugin-delayed-job-count?status.svg)][godoc]

[license]: https://github.com/masutaka/mackerel-plugin-delayed-job-count/blob/master/LICENSE.txt
[godoc]: https://godoc.org/github.com/masutaka/mackerel-plugin-delayed-job-count

Description
-----------

[delayed_job](https://rubygems.org/gems/delayed_job) custom metrics plugin for mackerel.io agent.

Synopsis
--------

    mackerel-plugin-delayed-job-count -dsn=<dataSourceName>

See https://github.com/go-sql-driver/mysql/#dsn-data-source-name

SQL Drivers
-----------

* MySQL
* PostgreSQL (WIP)

Example of mackerel-agent.conf
------------------------------

    [plugin.metrics.delayed_job_count]
    command = "/path/to/mackerel-plugin-delayed-job-count -dsn=<dataSourceName>"
