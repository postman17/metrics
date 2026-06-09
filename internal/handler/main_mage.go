package handler

import (
	"fmt"
	"html/template"
	"net/http"

	mem "github.com/postman17/metrics/internal/repository"
)

const htmlPage = `
<!DOCTYPE html>
<html>
<head>
    <title>Metrics List</title>
</head>
<body>
    <h1>Current Metrics</h1>
    <ul>
        {{range $name, $value := .}}
            <li><strong>{{$name}}</strong>: {{$value}}</li>
        {{end}}
    </ul>
</body>
</html>`

func GetMainPage(storage mem.MetricsRepository) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		rw.Header().Set("Content-Type", "text/html")
		rw.WriteHeader(http.StatusOK)
		template := template.Must(template.New("metrics").Parse(htmlPage))
		err := template.Execute(rw, storage.GetAll())
		if err != nil {
			fmt.Println("Ошибка шаблона:", err)
		}
	}
}
