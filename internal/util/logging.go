package util




type LogLevel int

const (
    Debug LogLevel = iota
    Info
    Warning
    Error
    Fatal
    Panic
)


func (l LogLevel) String() string {
    return []string{"DEBUG:", "INFO:", "WARNING:", "ERROR:", "FATAL:", "PANIC:"}[l]
}


