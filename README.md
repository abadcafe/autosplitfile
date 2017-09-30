# autosplitfile

auto split file for golang loggers

### Example

with logrus:
```go
func func main() {
    logFile, err := autosplitfile.New(&autosplitfile.FileOptions{
        PathPrefix:    "./the-log.log",
        BufferedLines: 4096,
        MaxSize:       1048576 * 1024, // 1GiB
        MaxTime:       "24h",
    }) // now got a io.WriteCloser
    
    if err != nil {
        logrus.WithError(err).Fatal("create log file failed")
    }
    
    defer logFile.Close() // close the file while normally exit.
    logrus.SetOutput(logFile) // set output to the file
    logrus.RegisterExitHandler(func() { logFile.Close() }) // close the file while fatally exit.
}
```

with built-in log:
```go
func func main() {
    logFile, err := autosplitfile.New(&autosplitfile.FileOptions{
        PathPrefix:    "./the-log.log",
        BufferedLines: 4096,
        MaxSize:       1048576 * 1024, // 1GiB
        MaxTime:       "24h",
    }) // now got a io.WriteCloser
    
    if err != nil {
        log.Fatal("create log file failed")
    }
    
    defer logFile.Close() // close the file while normally exit.
    log.SetOutput(logFile) // set output to the file
}
```

while the program is running, the log files looks like this:
```
the-log.log.20170930-00-00.0001 # the first file
the-log.log.20171001-00-00.0001 # will split every 24 hour
the-log.log.20171001-00-00.0002 # or split if file size exceed 1GiB
```

## NOTE

the autosplitfile cached writes, so close the file is necessary, otherwise writes maybe loss.