package main

import (
	"os"
	"pmm-transferer/pkg/clickhouse"
	"pmm-transferer/pkg/transfer"
	"pmm-transferer/pkg/victoriametrics"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TODO:
//  lint checker;
//  readme;
//  git version command;
//  end points ping;
//  vendor;
//  short versions of commands;
//  more logs;

func main() {
	var (
		transferer = kingpin.New("pmm-transferer", "Percona PMM Transferer")

		clickHouseURL      = transferer.Flag("click_house_url", "ClickHouse connection string").String()
		victoriaMetricsURL = transferer.Flag("victoria_metrics_url", "VictoriaMetrics connection string").String()
		enableVerboseMode  = transferer.Flag("verbose_mode", "Enable verbose mode").Short('v').Bool()

		exportCmd  = transferer.Command("export", "Export PMM Server metrics to dump file")
		outPath    = exportCmd.Flag("out", "Path to put out file").Short('o').String()
		tsSelector = exportCmd.Flag("ts_selector", "Time series selector to pass to VM").String()
		start      = exportCmd.Flag("start", "Start date-time to filter exported metrics, ex. "+time.RFC3339).String()
		end        = exportCmd.Flag("end", "End date-time to filter exported metrics, ex. "+time.RFC3339).String()

		importCmd = transferer.Command("import", "Import PMM Server metrics from dump file")
	)

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	if *enableVerboseMode {
		log.Logger = log.Logger.
			With().Caller().Logger(). // TODO: fix with caller log
			Level(zerolog.DebugLevel)
	}

	cmd, err := transferer.DefaultEnvars().Parse(os.Args[1:])
	if err != nil {
		log.Fatal().Msgf("Error parsing parameters: %s", err.Error())
	}

	if *clickHouseURL == "" && *victoriaMetricsURL == "" {
		log.Fatal().Msg("Please, specify at least one data source via connection string")
	}

	var (
		vmConfig *victoriametrics.Config
		chConfig *clickhouse.Config
	)

	if url := *victoriaMetricsURL; url != "" {
		vmConfig = &victoriametrics.Config{
			ConnectionURL:      url,
			TimeSeriesSelector: *tsSelector,
		}
		log.Info().Msgf("Got Victoria Metrics URL: %s", vmConfig.ConnectionURL)
	}

	if url := *clickHouseURL; url != "" {
		chConfig = &clickhouse.Config{
			ConnectionURL: url,
		}
		log.Info().Msgf("Got ClickHouse URL: %s", chConfig.ConnectionURL)
	}

	switch cmd {
	case exportCmd.FullCommand():
		p := exportParams{
			exporter: transfer.ExportConfig{
				OutPath: *outPath,
			},
			victoriaMetrics: vmConfig,
			clickHouse:      chConfig,
		}

		if *start != "" {
			start, err := time.Parse(time.RFC3339, *start)
			if err != nil {
				log.Fatal().Msgf("Error parsing start date-time: %v", err)
			}
			p.exporter.Start = &start
		}

		if *end != "" {
			end, err := time.Parse(time.RFC3339, *end)
			if err != nil {
				log.Fatal().Msgf("Error parsing end date-time: %v", err)
			}
			p.exporter.End = &end
		}

		if err = runExport(p); err != nil {
			log.Fatal().Msgf("Failed to export: %v", err)
		}
	case importCmd.FullCommand():
		log.Fatal().Msg("TO BE DONE") // TODO: import
	default:
		log.Fatal().Msgf("Undefined command found: %s", cmd)
	}
}
