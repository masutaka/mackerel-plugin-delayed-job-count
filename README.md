mackerel-plugin-delayed-job-count
=================================

[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]

[license]: https://masutaka.mit-license.org/

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
