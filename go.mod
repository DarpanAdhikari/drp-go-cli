module github.com/yourorg/drp

go 1.22.2

require (
	github.com/go-sql-driver/mysql v1.7.1
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
	github.com/spf13/cobra v1.8.1
	github.com/inconshreveable/mousetrap v1.1.0
	github.com/spf13/pflag v1.0.5
	gopkg.in/yaml.v3 v3.0.1
)

replace gopkg.in/yaml.v3 => github.com/go-yaml/yaml v3.0.1+incompatible

replace gopkg.in/check.v1 => github.com/go-check/check v0.0.0-20161208181325-20d25e280405
