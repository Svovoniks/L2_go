package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"syscall"
)

type Session struct {
	curDir string
}

func tryCd(session *Session, newDir string) bool {
	if info, err := os.Stat(newDir); err == nil && info.IsDir() {
		session.curDir = newDir
		return true
	}

	return false
}

func handleCd(session *Session, args []string) string {
	if len(args) != 1 {
		return "Usage: cd <dir>"
	}

	if tryCd(session, args[0]) {
		return ""
	}

	if tryCd(session, path.Join(session.curDir, args[0])) {
		return ""
	}

	return args[0] + " :no such directory"
}

func handlePwd(session *Session, args []string) string {
	return session.curDir
}

func handleEcho(session *Session, args []string) string {
	for i := range args {
		if len(args[i]) > 0 && args[i][0] == '$' {
			args[i] = os.Getenv(args[i][1:])
		}
	}

	return strings.Join(args, " ")
}

func handleKill(session *Session, args []string) string {
	for i := range args {
		if pid, err := strconv.Atoi(args[i]); err == nil {
			err = syscall.Kill(pid, syscall.SIGKILL)
			if err != nil {
				fmt.Println(err)
			}
			continue
		}

		return "Invalid pid"
	}

	return ""
}

type procInfo struct {
	pid   int
	name  string
	state string
	mem   string
}

func tryGetprocInfo(fileName string) (*procInfo, error) {
	pidInt, err := strconv.Atoi(fileName)

	if err != nil {
		return nil, err
	}

	infoFile, fileErr := os.Open(path.Join("/proc", fileName, "status"))
	if fileErr != nil {
		return nil, fileErr
	}

	defer infoFile.Close()

	info := procInfo{pid: pidInt}

	scanner := bufio.NewScanner(infoFile)

	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")

		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "Name":
			info.name = val
		case "State":
			info.state = val
		case "VmRSS":
			info.mem = val
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &info, nil

}

func handlePs(session *Session, args []string) string {
	dir, errDir := os.Open("/proc")
	if errDir != nil {
		return "Failed to read /proc directory"
	}

	files, errFile := dir.ReadDir(0)
	if errFile != nil {
		return "Failed to read /proc directory"
	}

	procs := []*procInfo{}
	mxPID := len("PID")
	mxName := len("Name")
	mxState := len("State")
	mxMem := len("Mem")

	for i := range files {
		if info, err := tryGetprocInfo(files[i].Name()); err == nil {
			if len(files[i].Name()) > mxPID {
				mxPID = len(files[i].Name())
			}

			if len(info.state) > mxState {
				mxState = len(info.state)
			}

			if len(info.mem) > mxMem {
				mxMem = len(info.mem)
			}

			if len(info.name) > mxName {
				mxName = len(info.name)
			}

			procs = append(procs, info)
		}
	}

	buff := bytes.Buffer{}
	buff.WriteString("%-" + strconv.Itoa(mxPID) + "v   ")
	buff.WriteString("%-" + strconv.Itoa(mxName) + "v   ")
	buff.WriteString("%-" + strconv.Itoa(mxMem) + "v   ")
	buff.WriteString("%-" + strconv.Itoa(mxState) + "v\n")

	formatStr := buff.String()

	resp := bytes.Buffer{}
	resp.WriteString(fmt.Sprintf(formatStr, "PID", "Name", "Mem", "State"))

	for i := range procs {
		resp.WriteString(fmt.Sprintf(formatStr, procs[i].pid, procs[i].name, procs[i].mem, procs[i].state))
	}

	return resp.String()

}

func tryExec(session *Session, args []string) (*os.Process, error) {
	proc, err := os.StartProcess(args[0], args[1:], &os.ProcAttr{})
	if err == nil {
		return proc, nil
	}

	paths := strings.Split(os.Getenv("PATH"), ":")
	paths = append(paths, session.curDir)

	for i := range paths {
		proc, err := os.StartProcess(path.Join(paths[i], args[0]), args[1:], &os.ProcAttr{})
		if err == nil {
			return proc, nil
		}
	}

	return nil, err
}

func handleExec(session *Session, args []string) string {
	if len(args) < 1 {
		return ""
	}
	proc, err := tryExec(session, args)
	if err != nil {
		return err.Error()
	}

	proc.Wait()

	return "Finished"
}

func handleFork(session *Session, args []string) string {
	if len(args) < 1 {
		return ""
	}

	proc, err := tryExec(session, args)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	return "PID: " + strconv.Itoa(proc.Pid)
}

func getInput(session *Session) string {
	fmt.Printf("> %v > ", session.curDir)

	scanner := bufio.NewScanner(os.Stdin)

	scanner.Scan()

	line := scanner.Text()

	if err := scanner.Err(); err != nil {
		fmt.Println("Fatal error")
		os.Exit(1)
	}

	return line
}

func runStage(session *Session, str string) string {
	parts := strings.Split(strings.TrimSpace(str), " ")

	if len(parts) < 1 {
		return ""
	}

	var task func(*Session, []string) string

	switch parts[0] {
	case "cd":
		task = handleCd
	case "pwd":
		task = handlePwd
	case "echo":
		task = handleEcho
	case "kill":
		task = handleKill
	case "ps":
		task = handlePs
	case "exec":
		task = handleExec
	case "fork":
		task = handleFork
	case "exit":
		os.Exit(0)
	default:
		fmt.Println("Unknown command", parts[0])
		return ""
	}

	return task(session, parts[1:])
}

func execute(str string, session *Session) {
	stages := strings.Split(str, "|")

	input := ""

	for i := range stages {
		stageStr := strings.TrimSpace(stages[i]) + " " + input
		input = runStage(session, stageStr)
	}

	fmt.Println(input)

}

func run(session *Session) {
	for {
		execute(getInput(session), session)
	}
}

func main() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}

	sessh := &Session{curDir: dir}
	run(sessh)

}
