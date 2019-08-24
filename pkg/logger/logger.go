package logger


var (
	//Log is global logger
	Log *zap.Logger

	//timeFormat is custom Time Format
	customTimeFormat string

	//onceInit guarantee logger initializes only once
	onceInit sync.Once
)


//customTimeEncode encodes Time to our custom format
//This is an example of how we can customize zap default functionality
func customTimeEncoder(t time.Time,enc zapcore.PrimitiveArrayEncoder){
	enc.AppendString(t.Format(customTimeFormat))
}

//Init initializes log with input parameters
//level - global log level; Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
//timeFormat - custom time format for logger, incase of empty string use default
func Init(level int,timeFormat string)error{
	var err error

	onceInit.Do(func(){
		//First define our level-handling logic.
		globalLevel := zapcore.Level(level)

		//High-priority output should go to standard error.
		//It is useful for Kubernetes deployment.
		//Kubernetes interprets os.Stdout log items as INFO and os.Stderr log items
		//as ERROR by default
		highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool{
			return lvl >= zapcore.ErrorLevel
		})

		//low priority output should  go to standard output.
		lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level)bool{
			return lvl >= globalLevel && level < zapcore.ErrorLevel
		})
		consoleInfos := zapcore.Lock(os.Stdout)
		consoleErrors := zapcore.Lock(os.Stderr)

		//Configure console output
		var useCustomTimeFormat bool
		ecfg := zap.NewProductionEncoderConfig()
		if len(timeFormat) > 0{
			customTimeFormat = timeFormat
			ecfg.EncodeTime = customTimeEncoder
			useCustomTimeFormat = true
		}
		consoleEncoder := zapcore.NewJSONEncoder(ecfg)

		//Join the outputs, encoders, and level-handling functions into
		//zapcore
		core := zapcore.NewTee(
			zapcore.NewCore(consoleEncoder,consoleErrors,highPriority),
			zapcore.NewCore(consoleEncoder,consoleInfos,lowPriority),
		)

		//From a zapcore.Core, it's easy to construct a Logger.
		Log = zap.New(core)
		zap.RedirectStdLog(log)

		if !useCustomTimeFormat{
			Log.Warn("time format for logger is not provided - use zap default")
		}
	})

	return err
}