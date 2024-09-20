package service

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"test/internal/utils"
)

const (
	inDataPath  = "/tmp/indata.txt"
	outDataPath = "/tmp/outdata.txt"
	bufferSize  = 1024
)

type App struct {
	outFile *os.File
	lessCmd *exec.Cmd
	wg      sync.WaitGroup
	done    chan struct{}
}

func NewApp() *App {
	return &App{
		done: make(chan struct{}),
	}
}

func (a *App) Run() error {
	utils.LogMessage("Application started")

	var err error
	a.outFile, err = os.Create(outDataPath)
	if err != nil {
		utils.LogMessage("Failed to create output file: %v", err)
		return err
	}

	a.lessCmd = exec.Command("less", "-F", "-R", outDataPath)
	a.lessCmd.Stdout = os.Stdout
	a.lessCmd.Stderr = os.Stderr
	if err := a.lessCmd.Start(); err != nil {
		utils.LogMessage("Failed to start less: %v", err)
		return err
	}

	if err := a.readAllInputData(); err != nil {
		utils.LogMessage("Failed to read input data: %v", err)
		return err
	}

	a.wg.Add(1)
	go a.watchInputFile()

	return nil
}

func (a *App) readAllInputData() error {
	inFile, err := os.Open(inDataPath)
	if err != nil {
		return err
	}
	defer inFile.Close()

	reader := bufio.NewReader(inFile)
	buffer := make([]byte, bufferSize)

	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			line := string(buffer[:n])
			_, err := a.outFile.WriteString(line)
			if err != nil {
				return err
			}
			utils.LogMessage("Read and wrote: %s", line)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	a.outFile.Sync()
	return nil
}

func (a *App) watchInputFile() {
	defer a.wg.Done()

	var lastModTime time.Time
	var lastSize int64
	var lastReadPos int64

	for {
		select {
		case <-a.done:
			return
		default:
			func() {
				timeout := time.After(5 * time.Second)
				done := make(chan struct{})

				go func() {
					defer close(done)

					inFileInfo, err := os.Stat(inDataPath)
					if err != nil {
						utils.LogMessage("Failed to stat input file: %v", err)
						return
					}

					if inFileInfo.ModTime().After(lastModTime) || inFileInfo.Size() != lastSize {
						lastModTime = inFileInfo.ModTime()
						lastSize = inFileInfo.Size()

						inFile, err := os.Open(inDataPath)
						if err != nil {
							utils.LogMessage("Failed to open input file: %v", err)
							return
						}
						defer inFile.Close()

						_, err = inFile.Seek(lastReadPos, 0)
						if err != nil {
							utils.LogMessage("Failed to seek input file: %v", err)
							return
						}

						reader := bufio.NewReader(inFile)
						buffer := make([]byte, bufferSize)
						for {
							n, err := reader.Read(buffer)
							if n > 0 {
								line := string(buffer[:n])
								_, err := a.outFile.WriteString(line)
								if err != nil {
									utils.LogMessage("Failed to write to output file: %v", err)
									return
								}
								utils.LogMessage("Read and wrote line: %s", line)
							}
							if err != nil {
								if err == io.EOF {
									break
								}
								utils.LogMessage("Error reading input file: %v", err)
								return
							}
						}

						lastReadPos, _ = inFile.Seek(0, io.SeekCurrent)
						a.outFile.Sync()
					}
				}()

				select {
				case <-done:
				case <-timeout:
					utils.LogMessage("Watch goroutine timed out")
					return
				}
			}()

			time.Sleep(1 * time.Second)
		}
	}
}

func (a *App) Wait() {
	if err := a.lessCmd.Wait(); err != nil {
		utils.LogMessage("less exited with error: %v", err)
	}
	a.wg.Wait()
}

func (a *App) Cleanup() {
	close(a.done)
	if a.outFile != nil {
		a.outFile.Close()
	}
	if err := os.Remove(outDataPath); err != nil {
		utils.LogMessage("Failed to remove file: %v", err)
	}
	utils.LogMessage("Application stopped")
}
