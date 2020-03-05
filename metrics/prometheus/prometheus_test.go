package metrics

import (
	"testing"
	"time"

	"github.com/prebid/prebid-cache/config"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func createPrometheusMetricsForTesting() *PrometheusMetrics {
	return CreatePrometheusMetrics(config.PrometheusMetrics{
		Port:      8080,
		Namespace: "prebid",
		Subsystem: "cache",
	})
}

func assertCounterVecValue(t *testing.T, description string, counterVec *prometheus.CounterVec, expected float64, labels prometheus.Labels) {
	counter := counterVec.With(labels)
	assertCounterValue(t, description, counter, expected)
}

func assertCounterValue(t *testing.T, description string, counter prometheus.Counter, expected float64) {
	m := dto.Metric{}
	counter.Write(&m)
	actual := *m.GetCounter().Value

	assert.Equal(t, expected, actual, description)
}

func assertGaugeValue(t *testing.T, description string, gauge prometheus.Gauge, expected float64) {
	m := dto.Metric{}
	gauge.Write(&m)
	actual := *m.GetGauge().Value

	assert.Equal(t, expected, actual, description)
}

func assertHistogram(t *testing.T, name string, histogram prometheus.Histogram, expectedCount uint64, expectedSum float64) {
	m := dto.Metric{}
	histogram.Write(&m)
	actual := *m.GetHistogram()

	assert.Equal(t, expectedCount, actual.GetSampleCount(), name+":count")
	assert.Equal(t, expectedSum, actual.GetSampleSum(), name+":sum")
}

func TestPrometheusRequestStatusMetric(t *testing.T) {
	m := createPrometheusMetricsForTesting()

	type testCaseObject struct {
		description      string
		expDuration      float64
		expRequestTotals float64
		expRequestErrors float64
		expBadRequests   float64
		testCase         func(pm *PrometheusMetrics)
	}

	testGroups := map[*PrometheusRequestStatusMetric][]testCaseObject{
		m.Puts: {
			{
				description: "Log put request duration",
				testCase: func(pm *PrometheusMetrics) {
					timeZero := time.Time{}
					tenSeconds := time.Second * 10

					now := timeZero.Add(tenSeconds)

					pm.RecordPutDuration(&now)
				},
				expDuration:      9.223372036854776e+09,
				expRequestTotals: 0, expRequestErrors: 0, expBadRequests: 0,
			},
			{
				description:      "Count put request total",
				testCase:         func(pm *PrometheusMetrics) { pm.RecordPutTotal() },
				expDuration:      9.223372036854776e+09,
				expRequestTotals: 1, expRequestErrors: 0, expBadRequests: 0,
			},
			{
				description:      "Count put request error",
				testCase:         func(pm *PrometheusMetrics) { pm.RecordPutError() },
				expDuration:      9.223372036854776e+09,
				expRequestTotals: 1, expRequestErrors: 1, expBadRequests: 0,
			},
			{
				description:      "Count put request bad request",
				testCase:         func(pm *PrometheusMetrics) { pm.RecordPutBadRequest() },
				expDuration:      9.223372036854776e+09,
				expRequestTotals: 1, expRequestErrors: 1, expBadRequests: 1,
			},
		},
		m.Gets: {
			{
				description: "Log get request duration",
				testCase: func(pm *PrometheusMetrics) {
					timeZero := time.Time{}
					tenSeconds := time.Second * 10

					now := timeZero.Add(tenSeconds)

					pm.RecordGetDuration(&now)
				},
				expDuration:      9.223372036854776e+09,
				expRequestTotals: 0, expRequestErrors: 0, expBadRequests: 0,
			},
			{
				description:      "Count get request total",
				testCase:         func(pm *PrometheusMetrics) { pm.RecordGetTotal() },
				expDuration:      9.223372036854776e+09,
				expRequestTotals: 1, expRequestErrors: 0, expBadRequests: 0,
			},
			{
				description:      "Count get request error",
				testCase:         func(pm *PrometheusMetrics) { pm.RecordGetError() },
				expDuration:      9.223372036854776e+09,
				expRequestTotals: 1, expRequestErrors: 1, expBadRequests: 0,
			},
			{
				description:      "Count get request bad request",
				testCase:         func(pm *PrometheusMetrics) { pm.RecordGetBadRequest() },
				expDuration:      9.223372036854776e+09,
				expRequestTotals: 1, expRequestErrors: 1, expBadRequests: 1,
			},
		},
		m.GetsBackend: {
			{
				description: "Log get backend request duration",
				testCase: func(pm *PrometheusMetrics) {
					timeZero := time.Time{}
					tenSeconds := time.Second * 10

					now := timeZero.Add(tenSeconds)

					pm.RecordGetBackendDuration(&now)
				},
				expDuration:      9.223372036854776e+09,
				expRequestTotals: 0, expRequestErrors: 0, expBadRequests: 0,
			},
			{
				description:      "Count get backend request total",
				testCase:         func(pm *PrometheusMetrics) { pm.RecordGetBackendTotal() },
				expDuration:      9.223372036854776e+09,
				expRequestTotals: 1, expRequestErrors: 0, expBadRequests: 0,
			},
			{
				description:      "Count get backend request error",
				testCase:         func(pm *PrometheusMetrics) { pm.RecordGetBackendError() },
				expDuration:      9.223372036854776e+09,
				expRequestTotals: 1, expRequestErrors: 1, expBadRequests: 0,
			},
			{
				description:      "Count get backend request bad request",
				testCase:         func(pm *PrometheusMetrics) { pm.RecordGetBackendBadRequest() },
				expDuration:      9.223372036854776e+09,
				expRequestTotals: 1, expRequestErrors: 1, expBadRequests: 1,
			},
		},
	}

	for prometheusMetric, testCaseArray := range testGroups {
		for _, test := range testCaseArray {
			test.testCase(m)

			assertHistogram(t, test.description, prometheusMetric.Duration, 1, test.expDuration)
			assertCounterVecValue(t, test.description, prometheusMetric.RequestStatus, test.expRequestTotals, prometheus.Labels{StatusKey: TotalsVal})
			assertCounterVecValue(t, test.description, prometheusMetric.RequestStatus, test.expRequestErrors, prometheus.Labels{StatusKey: ErrorVal})
			assertCounterVecValue(t, test.description, prometheusMetric.RequestStatus, test.expBadRequests, prometheus.Labels{StatusKey: BadRequestVal})
		}
	}
}

func TestPutBackendMetrics(t *testing.T) {
	m := createPrometheusMetricsForTesting()

	type testCaseObject struct {
		description string
		testCase    func(pm *PrometheusMetrics)

		//counters
		expXmlCount     float64
		expJsonCount    float64
		expInvalidCount float64
		expDefTTLCount  float64
		expErrorCount   float64

		//Duration and sixe in bytes
		expDuration      float64
		expSizeHistSum   float64
		expSizeHistCount uint64
	}

	testCases := []testCaseObject{
		{
			description: "Log put backend request duration",
			testCase: func(pm *PrometheusMetrics) {
				timeZero := time.Time{}
				tenSeconds := time.Second * 10

				now := timeZero.Add(tenSeconds)

				pm.RecordPutBackendDuration(&now)
			},
			expDuration: 9.223372036854776e+09,
		},
		{
			description: "Count put backend xml request",
			testCase:    func(pm *PrometheusMetrics) { pm.RecordPutBackendXml() },
			expDuration: 9.223372036854776e+09,
			expXmlCount: 1,
		},
		{
			description:  "Count put backend json request",
			testCase:     func(pm *PrometheusMetrics) { pm.RecordPutBackendJson() },
			expDuration:  9.223372036854776e+09,
			expXmlCount:  1,
			expJsonCount: 1,
		},
		{
			description:     "Count put backend invalid request",
			testCase:        func(pm *PrometheusMetrics) { pm.RecordPutBackendInvalid() },
			expDuration:     9.223372036854776e+09,
			expXmlCount:     1,
			expJsonCount:    1,
			expInvalidCount: 1,
		},
		{
			description:     "Count put backend of requests that define TTL",
			testCase:        func(pm *PrometheusMetrics) { pm.RecordPutBackendDefTTL() },
			expDuration:     9.223372036854776e+09,
			expXmlCount:     1,
			expJsonCount:    1,
			expInvalidCount: 1,
			expDefTTLCount:  1,
		},
		{
			description:     "Count put backend request errors",
			testCase:        func(pm *PrometheusMetrics) { pm.RecordPutBackendError() },
			expDuration:     9.223372036854776e+09,
			expXmlCount:     1,
			expJsonCount:    1,
			expInvalidCount: 1,
			expDefTTLCount:  1,
			expErrorCount:   1,
		},
		{
			description: "Log put backend request duration",
			testCase: func(pm *PrometheusMetrics) {
				pm.RecordPutBackendSize(16)
			},
			expDuration:      9.223372036854776e+09,
			expXmlCount:      1,
			expJsonCount:     1,
			expInvalidCount:  1,
			expDefTTLCount:   1,
			expErrorCount:    1,
			expSizeHistSum:   16,
			expSizeHistCount: 1,
		},
	}

	for _, test := range testCases {
		test.testCase(m)

		assertHistogram(t, test.description, m.PutsBackend.Duration, 1, test.expDuration)
		assertCounterVecValue(t, test.description, m.PutsBackend.PutBackendRequests, test.expXmlCount, prometheus.Labels{FormatKey: XmlVal})
		assertCounterVecValue(t, test.description, m.PutsBackend.PutBackendRequests, test.expJsonCount, prometheus.Labels{FormatKey: JsonVal})
		assertCounterVecValue(t, test.description, m.PutsBackend.PutBackendRequests, test.expInvalidCount, prometheus.Labels{FormatKey: InvFormatVal})
		assertCounterVecValue(t, test.description, m.PutsBackend.PutBackendRequests, test.expDefTTLCount, prometheus.Labels{FormatKey: DefinesTTLVal})
		assertCounterVecValue(t, test.description, m.PutsBackend.PutBackendRequests, test.expErrorCount, prometheus.Labels{FormatKey: ErrorVal})
		assertHistogram(t, test.description, m.PutsBackend.RequestLength, test.expSizeHistCount, test.expSizeHistSum)
	}
}

func TestConnectionMetrics(t *testing.T) {
	testCases := []struct {
		description                    string
		testCase                       func(pm *PrometheusMetrics)
		expectedOpenConnectionCount    float64
		expectedAcceptConnectionErrors float64
		expectedCloseConnectionErrors  float64
	}{
		{
			description: "Add a connection to the open connection count",
			testCase: func(pm *PrometheusMetrics) {
				pm.IncreaseOpenConnections()
			},
			expectedOpenConnectionCount:    1,
			expectedAcceptConnectionErrors: 0,
			expectedCloseConnectionErrors:  0,
		},
		{
			description: "Remove a connection from the open connection count",
			testCase: func(pm *PrometheusMetrics) {
				pm.DecreaseOpenConnections()
			},
			expectedOpenConnectionCount:    0,
			expectedAcceptConnectionErrors: 0,
			expectedCloseConnectionErrors:  0,
		},
		{
			description: "Count a close connection error",
			testCase: func(pm *PrometheusMetrics) {
				pm.RecordCloseConnectionErrors()
			},
			expectedOpenConnectionCount:    0,
			expectedAcceptConnectionErrors: 0,
			expectedCloseConnectionErrors:  1,
		},
		{
			description: "Count an accept connection error",
			testCase: func(pm *PrometheusMetrics) {
				pm.RecordCloseConnectionErrors()
				pm.RecordAcceptConnectionErrors()
			},
			expectedOpenConnectionCount:    0,
			expectedAcceptConnectionErrors: 1,
			expectedCloseConnectionErrors:  2,
		},
	}

	m := createPrometheusMetricsForTesting()

	for _, test := range testCases {
		test.testCase(m)

		assertGaugeValue(t, test.description, m.Connections.ConnectionsOpened, test.expectedOpenConnectionCount)
		assertCounterVecValue(t, test.description, m.Connections.ConnectionsErrors, test.expectedAcceptConnectionErrors, prometheus.Labels{ConnErrorKey: AcceptVal})
		assertCounterVecValue(t, test.description, m.Connections.ConnectionsErrors, test.expectedCloseConnectionErrors, prometheus.Labels{ConnErrorKey: CloseVal})
	}
}

func TestExtraTTLMetrics(t *testing.T) {
	m := createPrometheusMetricsForTesting()

	assertHistogram(t, "Assert the extra time to live in seconds allocated when metrics were created", m.ExtraTTL.ExtraTTLSeconds, 1, 5000.00)

	m.RecordExtraTTLSeconds(5)
	assertHistogram(t, "Assert the extra time to live in seconds was logged", m.ExtraTTL.ExtraTTLSeconds, 2, 5005.00)
}