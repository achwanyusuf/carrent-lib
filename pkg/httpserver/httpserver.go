package httpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/achwanyusuf/carrent-lib/pkg/common"
	"github.com/achwanyusuf/carrent-lib/pkg/errormsg"
	"github.com/achwanyusuf/carrent-lib/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
)

type HTTPServer struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

var ginEngine *gin.Engine

type HTTPSetting struct {
	Env     string
	Conf    HTTPServer
	Swagger Swagger
	Log     logger.Logger
}

type Swagger struct {
	Title                  string
	Description            string
	Version                string
	Host                   string
	Port                   int
	OAuth2PasswordTokenUrl string
}

func (h *HTTPSetting) NewHTTPServer() *gin.Engine {
	if h.Conf.Port == 0 {
		h.Conf.Port = 8080
	}

	if h.Env == "Production" {
		gin.SetMode(gin.ReleaseMode)
	}
	e := gin.New()

	ginEngine = e
	h.middleware()
	return e
}

func (h *HTTPSetting) Run() {
	readSignal := make(chan os.Signal, 1)
	signal.Notify(
		readSignal,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	go func() {
		s := &http.Server{
			Addr:           fmt.Sprintf("%s:%v", h.Conf.Host, h.Conf.Port),
			Handler:        ginEngine,
			ReadTimeout:    h.Conf.ReadTimeout,
			WriteTimeout:   h.Conf.WriteTimeout,
			MaxHeaderBytes: 1 << 20,
		}
		if err := s.ListenAndServe(); err != nil {
			panic(err)
		}
	}()
	<-readSignal
}

func (h *HTTPSetting) middleware() {
	ginEngine.Use(h.requestLogger())
	ginEngine.Use(gin.Recovery())
	ginEngine.Use(h.generateRequestID)
	h.pprofRouter()
	h.healthRouter()
}

func (h *HTTPSetting) generateRequestID(c *gin.Context) {
	requestID := c.GetHeader("X-Request-ID")

	if requestID == "" {
		requestID = uuid.New().String()
		c.Request.Header.Set("X-Request-ID", requestID)
	}

	c.Header("X-Request-ID", requestID)
}

func (h *HTTPSetting) pprofRouter() {
	pprofRouter := ginEngine.Group("pprof")
	pprofRouter.GET("/", gin.WrapF(pprof.Index))
	pprofRouter.GET("/cmdline", gin.WrapF(pprof.Cmdline))
	pprofRouter.GET("/profile", gin.WrapF(pprof.Profile))
	pprofRouter.POST("/symbol", gin.WrapF(pprof.Symbol))
	pprofRouter.GET("/symbol", gin.WrapF(pprof.Symbol))
	pprofRouter.GET("/trace", gin.WrapF(pprof.Trace))
	pprofRouter.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
	pprofRouter.GET("/block", gin.WrapH(pprof.Handler("block")))
	pprofRouter.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
	pprofRouter.GET("/heap", gin.WrapH(pprof.Handler("heap")))
	pprofRouter.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
	pprofRouter.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
}

func (h *HTTPSetting) SetSwaggo(docs *swag.Spec) {
	docs.Title = h.Swagger.Title
	docs.Description = h.Swagger.Description
	docs.Version = h.Swagger.Version
	docs.Host = fmt.Sprintf("%s:%v", h.Swagger.Host, h.Swagger.Port)
	docs.BasePath = "/api"
	docs.Schemes = []string{"http"}
	docs.SwaggerTemplate = fmt.Sprintf(docs.SwaggerTemplate, h.Swagger.OAuth2PasswordTokenUrl)
	ginEngine.GET("/swagger", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})
	ginEngine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (h *HTTPSetting) healthRouter() {
	ginEngine.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, map[string]interface{}{
			"message": "health",
		})
	})

	ginEngine.GET("/live", func(ctx *gin.Context) {
		ctx.JSON(200, map[string]interface{}{
			"message": "live",
		})
	})

	ginEngine.GET("/ready", func(ctx *gin.Context) {
		ctx.JSON(200, map[string]interface{}{
			"message": "ready",
		})
	})
}

type reqData struct {
	Header string
	Method string
	Body   string
	Query  string
}

type respData struct {
	Latency    string
	StatusCode int
	Data       string
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (h *HTTPSetting) requestLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		requestDump, _ := httputil.DumpRequest(ctx.Request, true)
		var buf bytes.Buffer
		tee := io.TeeReader(ctx.Request.Body, &buf)
		body, _ := io.ReadAll(tee)
		ctx.Request.Body = io.NopCloser(&buf)
		reqStr, _ := json.Marshal(reqData{
			Header: string(requestDump),
			Method: ctx.Request.Method,
			Body:   string(body),
			Query:  ctx.Request.URL.Query().Encode(),
		})
		h.Log.Info(ctx, string(reqStr))
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = blw
		ctx.Next()

		latency := time.Since(start)

		statusCode := ctx.Writer.Status()
		resStr, _ := json.Marshal(respData{
			Latency:    latency.String(),
			StatusCode: statusCode,
			Data:       blw.body.String(),
		})
		if statusCode >= 400 {
			h.Log.Error(ctx, string(resStr))
		} else {
			h.Log.Info(ctx, string(resStr))
		}
	}
}

type transactionInfo struct {
	RequestURI    string    `json:"request_uri"`
	RequestMethod string    `json:"request_method"`
	RequestID     string    `json:"request_id"`
	Timestamp     time.Time `json:"timestamp"`
	ErrorCode     int64     `json:"error_code,omitempty"`
	Cause         string    `json:"cause,omitempty"`
}

type response struct {
	TransactionInfo transactionInfo `json:"transaction_info"`
	Code            int64           `json:"status_code"`
	Message         string          `json:"message,omitempty"`
	Translation     *translation    `json:"translation,omitempty"`
}

type translation struct {
	EN string `json:"en"`
}

func ValidateScope(log logger.Logger, scopes []string) gin.HandlerFunc {
	resp := response{}
	return func(ctx *gin.Context) {
		scope := ctx.Value("scope").(string)
		if found := common.FindStrInSlice(scope, scopes); !found {
			resp = response{
				TransactionInfo: transactionInfo{
					RequestURI:    ctx.Request.RequestURI,
					RequestMethod: ctx.Request.Method,
					RequestID:     ctx.GetHeader("x-request-id"),
					Timestamp:     time.Now(),
					ErrorCode:     errormsg.Error401.Code,
				},
				Code:    errormsg.Error401.StatusCode,
				Message: errormsg.Error401.Message,
				Translation: &translation{
					EN: errormsg.Error401.Translation.EN,
				},
			}
			ctx.JSON(int(errormsg.Error401.StatusCode), resp)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
