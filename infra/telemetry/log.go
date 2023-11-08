package telemetry

/**
基于 TelemetrySDK 2.0.0 实现

Created by Dustin.zhu on 2023/08/11.
*/

import (
	"os"

	dhive "devops.aishu.cn/AISHUDevOps/AnyShareFamily/_git/Hive"

	"devops.aishu.cn/AISHUDevOps/ONE-Architecture/_git/TelemetrySDK-Go.git/span/encoder"
	"devops.aishu.cn/AISHUDevOps/ONE-Architecture/_git/TelemetrySDK-Go.git/span/field"
	"devops.aishu.cn/AISHUDevOps/ONE-Architecture/_git/TelemetrySDK-Go.git/span/log"
	"devops.aishu.cn/AISHUDevOps/ONE-Architecture/_git/TelemetrySDK-Go.git/span/open_standard"
	"devops.aishu.cn/AISHUDevOps/ONE-Architecture/_git/TelemetrySDK-Go.git/span/runtime"
)

//go:generate mockgen -package mock_infra -source log.go -destination ./mock/telemetry_log_mock.go

func init() {
	dhive.Prepare(func(initiator dhive.Initiator) {
		initiator.BindInfra(false, initiator.IsPrivate(), func() *TLogImpl {
			var rt *runtime.Runtime
			// init baseLog
			logger := log.NewDefaultSamplerLogger()
			output := os.Stdout
			writer := open_standard.NewOpenTelemetry(encoder.NewJsonEncoder(output), nil)
			rt = runtime.NewRuntime(&writer, field.NewSpanFromPool)
			logger.SetRuntime(rt)
			logger.SetLevel(log.AllLevel)

			// start runtime
			go rt.Run()
			return &TLogImpl{logger: logger}
		})
	})
}

// TLog 提供可观测性数据的生产和上报能力
type TLog interface {
	Trace(typ string, message interface{}, options ...field.LogOptionFunc)
	Debug(typ string, message interface{}, options ...field.LogOptionFunc)
	Info(typ string, message interface{}, options ...field.LogOptionFunc)
	Warn(typ string, message interface{}, options ...field.LogOptionFunc)
	Error(typ string, message interface{}, options ...field.LogOptionFunc)
	Fatal(typ string, message interface{}, options ...field.LogOptionFunc)
}

type TLogImpl struct {
	dhive.Infra
	logger *log.SamplerLogger
}

// BeginRequest .
func (l *TLogImpl) BeginRequest(worker dhive.Worker) {
	l.Infra.BeginRequest(worker)
}

// Trace do a Info log a object into LogSpan,
// if LogSpan is not nil, this interface will log the info,
// if LogSpan is nil, this interface will create a LogSpan
// to log the info and signal the LogSpan.
func (l *TLogImpl) Trace(typ string, message interface{}, options ...field.LogOptionFunc) {
	// structured log
	l.logger.TraceField(field.MallocJsonField(message), typ, options...)
	l.logger.Close()
}

// Debug do a Info log a object into LogSpan,
// if LogSpan is not nil, this interface will log the info,
// if LogSpan is nil, this interface will create a LogSpan
// to log the info and signal the LogSpan.
func (l *TLogImpl) Debug(typ string, message interface{}, options ...field.LogOptionFunc) {
	// structured log
	l.logger.DebugField(field.MallocJsonField(message), typ, options...)
	l.logger.Close()
}

// Info do a Info log a object into LogSpan,
// if LogSpan is not nil, this interface will log the info,
// if LogSpan is nil, this interface will create a LogSpan
// to log the info and signal the LogSpan.
func (l *TLogImpl) Info(typ string, message interface{}, options ...field.LogOptionFunc) {
	// structured log
	l.logger.InfoField(field.MallocJsonField(message), typ, options...)
	l.logger.Close()
}

// Warn do a Info log a object into LogSpan,
// if LogSpan is not nil, this interface will log the info,
// if LogSpan is nil, this interface will create a LogSpan
// to log the info and signal the LogSpan.
func (l *TLogImpl) Warn(typ string, message interface{}, options ...field.LogOptionFunc) {
	// structured log
	l.logger.WarnField(field.MallocJsonField(message), typ, options...)
	l.logger.Close()
}

// Error do a Info log a object into LogSpan,
// if LogSpan is not nil, this interface will log the info,
// if LogSpan is nil, this interface will create a LogSpan
// to log the info and signal the LogSpan.
func (l *TLogImpl) Error(typ string, message interface{}, options ...field.LogOptionFunc) {
	// structured log
	l.logger.ErrorField(field.MallocJsonField(message), typ, options...)
	l.logger.Close()
}

// Fatal do a Info log a object into LogSpan,
// if LogSpan is not nil, this interface will log the info,
// if LogSpan is nil, this interface will create a LogSpan
// to log the info and signal the LogSpan.
func (l *TLogImpl) Fatal(typ string, message interface{}, options ...field.LogOptionFunc) {
	// structured log
	l.logger.FatalField(field.MallocJsonField(message), typ, options...)
	l.logger.Close()
}
