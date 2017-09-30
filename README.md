# autosplitfile
auto split file for golang loggers

Example

with logrus:
<pre>
	logFile, err := autosplitfile.New(&autosplitfile.FileOptions{
		PathPrefix:    "./the-log.log",
		BufferedLines: 4096,
		MaxSize:       1048576 * 1024,
		MaxTime:       "24h",
	})
	if err != nil {
		logrus.WithError(err).Fatal("create log file failed")
	}
	defer logFile.Close()

	logrus.SetOutput(logFile)
	logrus.RegisterExitHandler(func() { logFile.Close() })
</pre>
