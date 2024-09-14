package util

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
)

func CheckIfCommandExists(bin string) (path string, err error) {
	var fMsg string
	path, err = exec.LookPath(bin)
	if err != nil {
		fMsg = fmt.Sprintf("Command %s does not exist in PATH", bin)
		slog.Error(fMsg)
		return path, err
	}
	fMsg = fmt.Sprintf("'%s' exists, found at %s", bin, path)
	slog.Info(fMsg)
	return path, nil

}

func RunCommand(binPath string, args string) (exitCode int, e error) {
	fullCmd := fmt.Sprintf("%s %s", binPath, args)
	cmd := exec.Command("bash", "-c", fullCmd)
	var stdout, stderr bytes.Buffer
	// Write to stdout/err but also capture it in a variable
	prefixWriterStdOut := NewPrefixWriter(os.Stdout, "[CMD] ")
	prefixWriterStdErr := NewPrefixWriter(os.Stderr, "[CMD-ERR] ")
	cmd.Stdout = io.MultiWriter(prefixWriterStdOut, &stdout)
	cmd.Stderr = io.MultiWriter(prefixWriterStdErr, &stderr)
	err := cmd.Run()
	if err != nil {
		switch e := err.(type) {
		case *exec.Error:
			slog.Error("The command failed executing: " + err.Error())
			return 126, err
		case *exec.ExitError:
			errCodeMsg := fmt.Sprintf("Exit Code: %d", e.ExitCode())
			slog.Error("The command executed, but an error happened")
			slog.Error(errCodeMsg)
			return e.ExitCode(), nil
		default:
			log.Fatal("[FATAL] Unexpected error executing your command,", err)
		}
	}
	return 0, nil
}

type PrefixWriter struct {
	w      io.Writer
	prefix string
}

func NewPrefixWriter(w io.Writer, prefix string) *PrefixWriter {
	return &PrefixWriter{w, prefix}
}

func (e PrefixWriter) Write(p []byte) (int, error) {
	prefix := []byte(e.prefix)
	n, err := e.w.Write(append(prefix, p...))
	if err != nil {
		return n, err
	}
	if n != len(p) {
		return n, io.ErrShortWrite
	}
	return len(p), nil
}
