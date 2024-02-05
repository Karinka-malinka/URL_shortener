package logger

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger = logrus.New()

func Initialize() {

	Log.SetLevel(logrus.InfoLevel)

	Log.SetReportCaller(true)

	jsonFormatter := &logrus.JSONFormatter{
		PrettyPrint:      true,
		DisableTimestamp: false,
		TimestampFormat:  "2006-01-02 15:04:05",
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "time",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
			logrus.FieldKeyFunc:  "caller"},
	}

	Log.SetFormatter(jsonFormatter)

	/*
		log.SetFormatter(&logrus.JSONFormatter{})

		file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.SetOutput(file)
		} else {
			log.Info("Не удалось открыть файл логов, используется стандартный stderr")
		}
	*/
}

func RequestLogger(c echo.Context, values middleware.RequestLoggerValues) error {

	Log.WithFields(logrus.Fields{
		"URI":           values.URI,
		"metod":         values.Method,
		"latency":       values.Latency.Seconds(),
		"status":        values.Status,
		"response_size": values.ResponseSize,
		"error":         values.Error,
	}).Info("Request")

	return nil
}
