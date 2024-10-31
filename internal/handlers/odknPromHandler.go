package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/Emy/prom-opendata-kn-parking/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

var promReg = prometheus.NewRegistry()
var promHandler = promhttp.HandlerFor(
	promReg,
	promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	},
)

// The frame counter gets increased whenever an update has been polled from the open data platform.
// Its intended purpose is to indicate if there had been an update the last time data was requested
// from the prometheus instances.
//
// The counter gets reset to zero when the server is restarted.
var frameCounter = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name:      "frames_total",
		Namespace: "constance_parking",
		Help:      "Indicates if there had been a change in the values. (Increments when new data is pulled from the open data platform. Zeroes the counter on server restart)",
	},
)

var freeSpaces = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "constance_parking",
		Name:      "free_spaces",
		Help:      "Number of free parking spaces per parking lot in Constance",
	},
	[]string{"lot"}, // Labels for grouping
)

var occupancyRate = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: "constance_parking",
		Name:      "occupancy_rate",
		Help:      "Occupancy rate per parking lot in Constance (0 to 1 scale)",
	},
	[]string{"lot"}, // Labels for grouping
)

// This function is being called to initialize all needed handlers and event schedulers that are required for an automatic
// retrieval of parking data from the open data platform.
func InitializePrometheusHandling() {
	updatePrometheusData()
	promReg.MustRegister(frameCounter)
	promReg.MustRegister(freeSpaces)
	promReg.MustRegister(occupancyRate)
	http.Handle("/metrics", promHandler)
	logger.Info("odknPromHandler.go: Prometheus Handler initialized.")
	enableAutoScheduledFetch()
}

// This function is being called to fetch and populate the parking data from the open data platform and push them back
// into the prometheus handlers.
func updatePrometheusData() {
	logger.Debug("odknPromHandler.go: Updating API data...")

	fetchedData := fetchData()
	if fetchedData == nil {
		return
	}

	for _, feature := range types.ODKNParkingAPIResponse(*fetchedData).Features {
		lot := feature.Attributes.Name
		if feature.Attributes.RealFCap != nil {
			capacityFree := float64(*feature.Attributes.RealFCap)
			capacityAvailable := feature.Attributes.RealCapa
			logger.Debug(fmt.Sprintf("Feature ID: %d, Real Free Capacity (real_fcap): %d\n", feature.Attributes.ObjectID, &capacityFree))
			freeSpaces.With(prometheus.Labels{"lot": lot}).Set(capacityFree)
			if feature.Attributes.RealCapa > 0 {
				occupancy := float64(1 - (int(capacityFree) / capacityAvailable))
				occupancyRate.With(prometheus.Labels{"lot": lot}).Set(occupancy)
			} else {
				occupancyRate.With(prometheus.Labels{"lot": lot}).Set(-1)
			}
		} else {
			logger.Debug(fmt.Sprintf("Feature ID: %d, Real Free Capacity (real_fcap): null\n", feature.Attributes.ObjectID))
			freeSpaces.With(prometheus.Labels{"lot": lot}).Set(-1)
			occupancyRate.With(prometheus.Labels{"lot": lot}).Set(-1)
		}
	}

	// Finally increase frame counter by one to indicate an update has been made.
	frameCounter.Inc()

	logger.Debug("odknPromHandler.go: Updated API data successfully.")
}

// This function fetches the parking data from the open data platform with an HTTPS REST request.
//
// The function will return a ODKNParkingAPIResponse pointer. If that pointer is nil it means something went wrong
// on the side of the open data platform and there is no data being given out at the moment.
// In case this happens the data this function sent out should be disregarded.
func fetchData() *types.ODKNParkingAPIResponse {
	base, err := url.Parse("https://services.gis.konstanz.digital")
	if err != nil {
		logger.Error("odknPromHandler.go: Could not parse baseURL.")
		return nil
	}
	base.Path += "/geoportal/rest/services/Fachdaten/Parkplaetze_Parkleitsystem/MapServer/0/query"
	params := url.Values{}
	// Straight up jorking it. And by it I mean SQL into the query parameters (the docs told me to ;-;)
	params.Add("where", "1=1")
	params.Add("outFields", "*")
	params.Add("SR", "4326")
	params.Add("f", "json")
	base.RawQuery = params.Encode()
	res, err := http.Get(base.String())
	if err != nil {
		logger.Error("odknPromHandler.go: WebRequest > Could not make web request.", "Error", err)
		return nil
	}

	logger.Debug("odknPromHandler.go: WebRequest > Got Response with", "Statuscode", res.StatusCode)

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Error("odknPromHandler.go: WebRequest > Could not read response body.", "Error", err)
		return nil
	}

	var result types.ODKNParkingAPIResponse
	json.Unmarshal([]byte(resBody), &result)

	logger.Debug("odknPromHandler.go: Fetched and unmarshalled the data from the open data platform sucessfully.")

	return &result
}

// This function registers and calls the scheduled update job to fetch new data from the open data platform.
//
// Currently the interval is set that the job gets executed every 5 minutes on the clock with a 15 second delay to
// account for delays on the open data platform data base.
func enableAutoScheduledFetch() {
	c := cron.New()
	c.AddFunc("15 */5 * * *", func() { updatePrometheusData() }) // every 5 minutes with a 15 second delay.
	c.Start()
	logger.Info("odknPromHandler.go: Enabled scheduled fetching of API responses from the open data platform.")
}
