package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"main/internal/dto"
	"main/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegrationReport(t *testing.T) {
	cleanOrders()

	order := models.Order{
		TotalAmount:   1000,
		CurrentStatus: models.ORDER_STATUS_DELIVERED,
		UserInfo:      mustMarshalUserInfo("report_user", "0900000000", "Report Address"),
		CreatedAt:     time.Date(2026, time.May, 3, 10, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2026, time.May, 3, 10, 0, 0, 0, time.UTC),
	}
	require.NoError(t, db.Create(&order).Error)

	reportEvent := models.OrderEvent{
		OrderID:        order.ID,
		PreviousStatus: models.ORDER_STATUS_PACKED,
		NewStatus:      models.ORDER_STATUS_DELIVERED,
		EventAt:        time.Date(2026, time.May, 3, 12, 30, 0, 0, time.UTC),
		CreatedAt:      time.Date(2026, time.May, 3, 12, 30, 0, 0, time.UTC),
		UpdatedAt:      time.Date(2026, time.May, 3, 12, 30, 0, 0, time.UTC),
	}
	require.NoError(t, db.Create(&reportEvent).Error)

	cases := []struct {
		name           string
		method         string
		path           string
		body           []byte
		token          string
		expectedStatus int
		validate       func(t *testing.T, respBody []byte)
	}{
		{
			name:   "create daily report",
			method: "POST",
			path:   "/api/v1/reports/daily",
			body: func() []byte {
				b, _ := json.Marshal(dto.GetDailyReportRequest{Date: "2026-05-03"})
				return b
			}(),
			expectedStatus: 201,
			validate: func(t *testing.T, respBody []byte) {
				var postBody struct {
					Status string        `json:"status"`
					Data   models.Report `json:"data"`
				}
				require.NoError(t, json.Unmarshal(respBody, &postBody))
				// assert.Equal(t, 201, postBody.Status)
				assert.Equal(t, "SUCCESS", postBody.Status)
				assert.Equal(t, int64(1), postBody.Data.TotalOrders)
				assert.Equal(t, int64(1), postBody.Data.TotalDelivered)
				assert.Equal(t, int64(1000), postBody.Data.TotalIncome)
			},
		},
		{
			name:           "get daily report",
			method:         "GET",
			path:           "/api/v1/reports/daily?date=2026-05-03",
			body:           nil,
			expectedStatus: 200,
			validate: func(t *testing.T, respBody []byte) {
				var getBody struct {
					Status string        `json:"status"`
					Data   models.Report `json:"data"`
				}
				require.NoError(t, json.Unmarshal(respBody, &getBody))
				// assert.Equal(t, 200, getBody.Status)
				assert.Equal(t, "SUCCESS", getBody.Status)
				assert.Equal(t, int64(1), getBody.Data.TotalOrders)
				assert.Equal(t, int64(1), getBody.Data.TotalDelivered)
			},
		},
		{
			name:           "get daily report - missing date param",
			method:         "GET",
			path:           "/api/v1/reports/daily",
			body:           nil,
			expectedStatus: 400,
		},
		{
			name:           "get daily report - invalid date format",
			method:         "GET",
			path:           "/api/v1/reports/daily?date=04-05-2026",
			body:           nil,
			expectedStatus: 400,
		},
		{
			name:           "create daily report - missing date field",
			method:         "POST",
			path:           "/api/v1/reports/daily",
			body:           []byte(`{}`),
			expectedStatus: 400,
		},
		{
			name:           "create daily report - invalid date format",
			method:         "POST",
			path:           "/api/v1/reports/daily",
			body:           []byte(`{"date": "04/05/2026"}`),
			expectedStatus: 400,
		},
		{
			name:           "get report - date has no report",
			method:         "GET",
			path:           "/api/v1/reports/daily?date=2099-01-01",
			body:           nil,
			expectedStatus: 404,
		},
		{
			name:           "get daily report - unauthenticated",
			method:         "GET",
			path:           "/api/v1/reports/daily?date=2026-05-03",
			body:           nil,
			token:          "",
			expectedStatus: 401,
		},
		{
			name:           "get daily report - wrong role driver",
			method:         "GET",
			path:           "/api/v1/reports/daily?date=2026-05-03",
			body:           nil,
			token:          driverToken,
			expectedStatus: 403,
		},
		{
			name:           "create daily report - invalid json body",
			method:         "POST",
			path:           "/api/v1/reports/daily",
			body:           []byte(`{invalid json`),
			token:          adminToken,
			expectedStatus: 400,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.body != nil {
				req = httptest.NewRequest(tc.method, tc.path, bytes.NewBuffer(tc.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}

			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			} else if tc.name != "get daily report - unauthenticated" {
				req.Header.Set("Authorization", "Bearer "+adminToken) // fallback to admin for older test cases
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			resp.Body.Close()

			if tc.validate != nil {
				tc.validate(t, respBody)
			}
		})
	}
}
