File: handler.test
Type: alloc_space
Time: 2026-08-03 05:59:53 MSK
Showing nodes accounting for -4236.96MB, 8.80% of 48121.09MB total
Dropped 80 nodes (cum <= 240.61MB)
      flat  flat%   sum%        cum   cum%
-4235.50MB  8.80%  8.80% -4235.50MB  8.80%  io.ReadAll
 1783.18MB  3.71%  5.10%  1783.18MB  3.71%  bufio.NewReaderSize (inline)
-1099.35MB  2.28%  7.38% -1099.35MB  2.28%  net/textproto.MIMEHeader.Set (inline)
-1084.35MB  2.25%  9.63% -1084.35MB  2.25%  net/http.Header.Clone (inline)
 -347.74MB  0.72% 10.36% -2458.51MB  5.11%  github.com/postman17/metrics/internal/handler.BenchmarkUpdatesMetric.func1.UpdatesMetric.1
  330.39MB  0.69%  9.67%   437.89MB  0.91%  github.com/postman17/metrics/internal/model.easyjson2220f231DecodeGithubComPostman17MetricsInternalModel
 -257.21MB  0.53% 10.20% -1205.65MB  2.51%  github.com/postman17/metrics/internal/handler.BenchmarkUpdatesMetricMixed.func1.UpdatesMetric.1
  161.55MB  0.34%  9.87%   161.55MB  0.34%  net/http.(*Request).WithContext (inline)
  125.54MB  0.26%  9.61%   258.05MB  0.54%  net/http.readRequest
  111.51MB  0.23%  9.38%   111.51MB  0.23%  net/http/httptest.NewRecorder (inline)
   86.01MB  0.18%  9.20%    86.01MB  0.18%  net/url.parse
      62MB  0.13%  9.07%       62MB  0.13%  github.com/mailru/easyjson/jlexer.(*Lexer).String
   45.50MB 0.095%  8.97%    45.50MB 0.095%  net/textproto.readMIMEHeader
      30MB 0.062%  8.91%   467.90MB  0.97%  github.com/mailru/easyjson.Unmarshal
      28MB 0.058%  8.85%       90MB  0.19%  github.com/postman17/metrics/internal/model.easyjson2220f231DecodeGithubComPostman17MetricsInternalModel1
      15MB 0.031%  8.82%  2217.28MB  4.61%  net/http/httptest.NewRequestWithContext
      12MB 0.025%  8.80%  -925.38MB  1.92%  github.com/postman17/metrics/internal/handler.BenchmarkUpdatesMetricInvalidJSON.UpdatesMetric.func1
   -3.50MB 0.0073%  8.80%  -355.17MB  0.74%  github.com/postman17/metrics/internal/handler.BenchmarkUpdatesMetricCounterAccumulation.UpdatesMetric.func1
         0     0%  8.80%  1783.18MB  3.71%  bufio.NewReader (inline)
         0     0%  8.80% -3130.79MB  6.51%  github.com/postman17/metrics/internal/handler.BenchmarkUpdatesMetric.func1
         0     0%  8.80%   -45.15MB 0.094%  github.com/postman17/metrics/internal/handler.BenchmarkUpdatesMetricCounterAccumulation
         0     0%  8.80% -1555.01MB  3.23%  github.com/postman17/metrics/internal/handler.BenchmarkUpdatesMetricMethodNotAllowed.UpdatesMetric.func1
         0     0%  8.80%  -993.04MB  2.06%  github.com/postman17/metrics/internal/handler.BenchmarkUpdatesMetricMixed.func1
         0     0%  8.80%       90MB  0.19%  github.com/postman17/metrics/internal/model.(*Metrics).UnmarshalEasyJSON (inline)
         0     0%  8.80%   437.89MB  0.91%  github.com/postman17/metrics/internal/model.(*MetricsList).UnmarshalEasyJSON
         0     0%  8.80% -1099.35MB  2.28%  net/http.Header.Set (partial-inline)
         0     0%  8.80%   257.55MB  0.54%  net/http.ReadRequest
         0     0%  8.80% -1084.35MB  2.25%  net/http/httptest.(*ResponseRecorder).WriteHeader
         0     0%  8.80%  2217.28MB  4.61%  net/http/httptest.NewRequest (inline)
         0     0%  8.80%    45.50MB 0.095%  net/textproto.(*Reader).ReadMIMEHeader (inline)
         0     0%  8.80%    86.01MB  0.18%  net/url.ParseRequestURI
         0     0%  8.80% -4177.44MB  8.68%  testing.(*B).launch
         0     0%  8.80% -4178.44MB  8.68%  testing.(*B).runN
