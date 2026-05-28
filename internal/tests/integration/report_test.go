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

order1 := models.Order{
   TotalAmount:   1000,
   CurrentStatus: models.ORDER_STATUS_DELIVERED,
   UserInfo:      mustMarshalUserInfo("user1", "0900000000", "Address"),
   CreatedAt:     time.Date(2026, time.May, 3, 10, 0, 0, 0, time.UTC),
   UpdatedAt:     time.Date(2026, time.May, 3, 10, 0, 0, 0, time.UTC),
}
require.NoError(t, db.Create(&order1).Error)


event1 := models.OrderEvent{
   OrderID:        order1.ID,
   PreviousStatus: models.ORDER_STATUS_PACKED,
   NewStatus:      models.ORDER_STATUS_DELIVERED,
   EventAt:        time.Date(2026, time.May, 3, 12, 0, 0, 0, time.UTC),
}
require.NoError(t, db.Create(&event1).Error)


order2 := models.Order{
   TotalAmount:   2000,
   CurrentStatus: models.ORDER_STATUS_DELIVERED,
   UserInfo:      mustMarshalUserInfo("user2", "0900000001", "Address"),
   CreatedAt:     time.Date(2026, time.May, 2, 10, 0, 0, 0, time.UTC),
}
require.NoError(t, db.Create(&order2).Error)


event2 := models.OrderEvent{
   OrderID:        order2.ID,
   PreviousStatus: models.ORDER_STATUS_PACKED,
   NewStatus:      models.ORDER_STATUS_DELIVERED,
   EventAt:        time.Date(2026, time.May, 3, 14, 0, 0, 0, time.UTC),
}
require.NoError(t, db.Create(&event2).Error)


order3 := models.Order{
   TotalAmount:   3000,
   CurrentStatus: models.ORDER_STATUS_DELIVERED,
   UserInfo:      mustMarshalUserInfo("user3", "0900000002", "Address"),
   CreatedAt:     time.Date(2026, time.May, 3, 15, 0, 0, 0, time.UTC),
}
require.NoError(t, db.Create(&order3).Error)


event3 := models.OrderEvent{
   OrderID:        order3.ID,
   PreviousStatus: models.ORDER_STATUS_PACKED,
   NewStatus:      models.ORDER_STATUS_DELIVERED,
   EventAt:        time.Date(2026, time.May, 4, 10, 0, 0, 0, time.UTC),
}
require.NoError(t, db.Create(&event3).Error)


cases := []struct {
   name           string
   method         string
   path           string
   body           []byte
   apiKey         string
   expectedStatus int
   validate       func(t *testing.T, respBody []byte)
}{
   {
       name:   "create daily report - full coverage",
       method: "POST",
       path:   "/api/v1/reports/daily",
       body: func() []byte {
           b, _ := json.Marshal(dto.GetDailyReportRequest{Date: "2026-05-03"})
           return b
       }(),
       expectedStatus: 201,
       validate: func(t *testing.T, respBody []byte) {
           var postBody struct {
               Status int           `json:"status"`
               Data   models.Report `json:"data"`
           }
           require.NoError(t, json.Unmarshal(respBody, &postBody))


           assert.Equal(t, int64(2), postBody.Data.TotalDelivered)
           assert.Equal(t, int64(3000), postBody.Data.TotalIncome)
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
               Status int           `json:"status"`
               Data   models.Report `json:"data"`
           }
           require.NoError(t, json.Unmarshal(respBody, &getBody))


           assert.Equal(t, int64(2), getBody.Data.TotalDelivered)
           assert.Equal(t, int64(3000), getBody.Data.TotalIncome)
       },
   },
   {
       name:           "get daily report - missing date param",
       method:         "GET",
       path:           "/api/v1/reports/daily",
       expectedStatus: 400,
   },
   {
       name:           "get daily report - invalid date format",
       method:         "GET",
       path:           "/api/v1/reports/daily?date=04-05-2026",
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
       expectedStatus: 404,
   },
   {
       name:           "get daily report - unauthenticated",
       method:         "GET",
       path:           "/api/v1/reports/daily?date=2026-05-03",
       apiKey:         "",
       expectedStatus: 401,
   },
   {
       name:           "get daily report - wrong role customer",
       method:         "GET",
       path:           "/api/v1/reports/daily?date=2026-05-03",
       apiKey:         customerAPIKey,
       expectedStatus: 403,
   },
   {
       name:           "get daily report - wrong role driver",
       method:         "GET",
       path:           "/api/v1/reports/daily?date=2026-05-03",
       apiKey:         driverAPIKey,
       expectedStatus: 403,
   },
   {
       name:           "create daily report - invalid json body",
       method:         "POST",
       path:           "/api/v1/reports/daily",
       body:           []byte(`{invalid json`),
       apiKey:         adminAPIKey,
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


       if tc.apiKey != "" {
           req.Header.Set("X-API-KEY", tc.apiKey)
       } else if tc.name != "get daily report - unauthenticated" {
           req.Header.Set("X-API-KEY", adminAPIKey)
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
